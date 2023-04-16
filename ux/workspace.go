/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/printing"
	"golang.org/x/exp/slices"
)

const dockGroupClientDataKey = "dock.group"

// WS holds our Workspace singleton.
var WS *Workspace

// GroupedCloser defines the methods required of a tab that wishes to be closed when another tab is closed.
type GroupedCloser interface {
	unison.TabCloser
	CloseWithGroup(other unison.Paneler) bool
}

// Workspace holds the data necessary to track the Workspace.
type Workspace struct {
	Window       *unison.Window
	TopDock      *unison.Dock
	Navigator    *Navigator
	DocumentDock *DocumentDock
	PrintMgr     *printing.PrintManager
}

// InitWorkspace initializes the Workspace singleton, WS.
func InitWorkspace(wnd *unison.Window) {
	WS = &Workspace{
		Window:       wnd,
		TopDock:      unison.NewDock(),
		Navigator:    newNavigator(),
		DocumentDock: NewDocumentDock(),
		PrintMgr:     &printing.PrintManager{},
	}
	wnd.SetContent(WS.TopDock)
	WS.TopDock.DockTo(WS.Navigator, nil, unison.LeftSide)
	dc := unison.Ancestor[*unison.DockContainer](WS.Navigator)
	WS.TopDock.DockTo(WS.DocumentDock, dc, unison.RightSide)
	dc.SetCurrentDockable(WS.Navigator)
	wnd.AllowCloseCallback = WS.allowClose
	wnd.WillCloseCallback = WS.willClose
	global := gurps.GlobalSettings()
	if global.WorkspaceFrame != nil {
		r := *global.WorkspaceFrame
		if r.Width < 10 {
			r.Width = 10
		}
		if r.Height < 10 {
			r.Height = 10
		}
		r = unison.BestDisplayForRect(r).FitRectOnto(r)
		*global.WorkspaceFrame = r
		wnd.SetFrameRect(r)
	} else {
		wnd.SetFrameRect(unison.PrimaryDisplay().Usable)
	}
	// On some platforms, this needs to be done after a delay... but we do it without the delay, too, so that
	// well-behaved platforms don't flash
	WS.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	unison.InvokeTaskAfter(func() {
		WS.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	}, time.Millisecond)
	wnd.ToFront()
	WS.Navigator.InitialFocus()
}

// Activate attempts to locate an existing dockable that 'matcher' returns true for. If found, it will have been
// activated and focused.
func Activate(matcher func(d unison.Dockable) bool) (dc *unison.DockContainer, found bool) {
	dc = WS.CurrentlyFocusedDockContainer()
	WS.DocumentDock.RootDockLayout().ForEachDockContainer(func(container *unison.DockContainer) bool {
		for _, one := range container.Dockables() {
			if matcher(one) {
				found = true
				container.SetCurrentDockable(one)
				container.AcquireFocus()
				return true
			}
			if dc == nil {
				dc = container
			}
		}
		return false
	})
	return dc, found
}

// ActiveDockable returns the currently active dockable in the active window.
func ActiveDockable() unison.Dockable {
	if WS.Window != unison.ActiveWindow() {
		return nil
	}
	dc := WS.CurrentlyFocusedDockContainer()
	if dc == nil {
		return nil
	}
	return dc.CurrentDockable()
}

func (w *Workspace) allowClose() bool {
	may := true
	w.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, other := range dc.Dockables() {
			if tc, ok2 := other.(unison.TabCloser); ok2 {
				if _, ok := other.(GroupedCloser); !ok {
					if !tc.MayAttemptClose() {
						may = false
						return true
					}
					if !tc.AttemptClose() {
						may = false
						return true
					}
				}
			}
		}
		return false
	})
	return may
}

func (w *Workspace) willClose() {
	global := gurps.GlobalSettings()
	global.LibraryExplorer.OpenRowKeys = w.Navigator.DisclosedPaths()
	global.LibraryExplorer.DividerPosition = w.TopDock.RootDockLayout().DividerPosition()
	frame := w.Window.FrameRect()
	global.WorkspaceFrame = &frame
	if err := global.Save(); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to save global settings"), err)
	}
}

// CurrentlyFocusedDockContainer returns the currently focused DockContainer, if any.
func (w *Workspace) CurrentlyFocusedDockContainer() *unison.DockContainer {
	if focus := w.Window.Focus(); focus != nil {
		if dc := unison.Ancestor[*unison.DockContainer](focus); dc != nil && dc.Dock == w.DocumentDock.Dock {
			return dc
		}
	}
	return nil
}

// LocateFileBackedDockable searches for a FileBackedDockable with the given path.
func (w *Workspace) LocateFileBackedDockable(filePath string) FileBackedDockable {
	var dockable FileBackedDockable
	w.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, one := range dc.Dockables() {
			if fbd, ok := one.(FileBackedDockable); ok {
				if filePath == fbd.BackingFilePath() {
					dockable = fbd
					return true
				}
			}
		}
		return false
	})
	return dockable
}

// LocateDockContainerForExtension searches for the first FileBackedDockable with the given extension and returns its
// DockContainer.
func (w *Workspace) LocateDockContainerForExtension(ext ...string) *unison.DockContainer {
	var extDC *unison.DockContainer
	w.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		if DockContainerHoldsExtension(dc, ext...) {
			extDC = dc
			return true
		}
		return false
	})
	return extDC
}

// PlaceInDock places the Dockable into the workspace document dock, grouped with the provided group, if that group is
// present. 'dc' is the preferred DockContainer to place the Dockable within, which is usually the current one.
func PlaceInDock(dc *unison.DockContainer, dockable unison.Dockable, group string) {
	dockable.AsPanel().ClientData()[dockGroupClientDataKey] = group
	if DockContainerHasGroup(dc, group) {
		dc.Stack(dockable, -1)
		return
	}
	if dc = DockContainerForGroup(WS.DocumentDock.Dock, group); dc != nil {
		dc.Stack(dockable, -1)
		return
	}
	side := unison.RightSide
	if group == subEditorGroup {
		if dc = DockContainerForGroup(WS.DocumentDock.Dock, EditorGroup); dc != nil {
			side = unison.BottomSide
		}
	}
	WS.DocumentDock.DockTo(dockable, dc, side)
}

// DockContainerHasGroup returns true if the DockContainer contains at least one Dockable associated with the given
// group. May pass nil for the dc.
func DockContainerHasGroup(dc *unison.DockContainer, group string) bool {
	if dc == nil {
		return false
	}
	for _, dockable := range dc.Dockables() {
		if dockable.AsPanel().ClientData()[dockGroupClientDataKey] == group {
			return true
		}
	}
	return false
}

// DockContainerForGroup returns the first DockContainer which has a Dockable with the given group, if any.
func DockContainerForGroup(dock *unison.Dock, group string) *unison.DockContainer {
	var found *unison.DockContainer
	dock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		if DockContainerHasGroup(dc, group) {
			found = dc
			return true
		}
		return false
	})
	return found
}

// DockContainerHoldsExtension returns true if an immediate child of the given DockContainer has a FileBackedDockable
// with the given extension.
func DockContainerHoldsExtension(dc *unison.DockContainer, ext ...string) bool {
	for _, one := range dc.Dockables() {
		if fbd, ok := one.(FileBackedDockable); ok {
			fbdExt := path.Ext(fbd.BackingFilePath())
			for _, e := range ext {
				if strings.EqualFold(fbdExt, e) {
					return true
				}
			}
		}
	}
	return false
}

// AssociatedUUIDKey is the key used with CloseUUID().
const AssociatedUUIDKey = "associated_uuid"

// CloseUUID attempts to close any Dockables associated with the given UUIDs. Returns false if a dockable refused to
// close.
func CloseUUID(ids map[uuid.UUID]bool) bool {
	allow := true
	WS.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, other := range dc.Dockables() {
			if tc, ok := other.(unison.TabCloser); ok {
				if otherValue, ok2 := other.AsPanel().ClientData()[AssociatedUUIDKey]; ok2 {
					if otherID, ok3 := otherValue.(uuid.UUID); ok3 && ids[otherID] {
						if !tc.MayAttemptClose() {
							allow = false
							return true
						}
						if !tc.AttemptClose() {
							allow = false
							return true
						}
					}
				}
			}
		}
		return false
	})
	return allow
}

// MayAttemptCloseOfGroup returns true if the grouped Dockables associated with the given dockable may be closed.
func MayAttemptCloseOfGroup(d unison.Dockable) bool {
	allow := true
	traverseGroup(d, func(target GroupedCloser) bool {
		if !target.MayAttemptClose() {
			allow = false
			return true
		}
		return false
	})
	return allow
}

// CloseGroup attempts to close any grouped Dockables associated with the given Dockable. Returns false if a dockable
// refused to close.
func CloseGroup(d unison.Dockable) bool {
	allow := true
	traverseGroup(d, func(target GroupedCloser) bool {
		if !target.MayAttemptClose() {
			allow = false
			return true
		}
		if !target.AttemptClose() {
			allow = false
			return true
		}
		return false
	})
	return allow
}

func traverseGroup(d unison.Dockable, f func(target GroupedCloser) bool) {
	WS.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, other := range dc.Dockables() {
			if fe, ok := other.(GroupedCloser); ok && fe.CloseWithGroup(d) {
				if f(fe) {
					return true
				}
			}
		}
		return false
	})
}

// SaveDockable attempts to save the contents of the dockable using its existing path.
func SaveDockable(d FileBackedDockable, saver func(filePath string) error, setUnmodified func()) bool {
	filePath := d.BackingFilePath()
	if err := saver(filePath); err != nil {
		unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to save %s"), fs.BaseName(filePath)), err)
		return false
	}
	setUnmodified()
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	return true
}

// SaveDockableAs attempts to save the contents of the dockable, prompting for a new path.
func SaveDockableAs(d FileBackedDockable, extension string, saver func(filePath string) error, setUnmodifiedAndNewPath func(filePath string)) bool {
	dialog := unison.NewSaveDialog()
	existingPath := d.BackingFilePath()
	if !strings.HasPrefix(existingPath, markdownContentOnlyPrefix) && fs.FileExists(existingPath) {
		dialog.SetInitialDirectory(filepath.Dir(existingPath))
	} else {
		dialog.SetInitialDirectory(gurps.GlobalSettings().LastDir(gurps.DefaultLastDirKey))
	}
	dialog.SetAllowedExtensions(extension)
	if dialog.RunModal() {
		filePath, ok := unison.ValidateSaveFilePath(dialog.Path(), extension, false)
		if !ok {
			return false
		}
		gurps.GlobalSettings().SetLastDir(gurps.DefaultLastDirKey, filepath.Dir(filePath))
		if err := saver(filePath); err != nil {
			unison.ErrorDialogWithError(i18n.Text("Unable to save as ")+fs.BaseName(filePath), err)
			return false
		}
		setUnmodifiedAndNewPath(filePath)
		gurps.GlobalSettings().AddRecentFile(filePath)
		if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
			dc.UpdateTitle(d)
		}
		return true
	}
	return false
}

// PromptForDestination puts up a modal dialog to choose one or more destinations if choices contains more than one
// choice. Return an empty list if canceled or there are no selections made.
func PromptForDestination[T FileBackedDockable](choices []T) []T {
	if len(choices) < 2 {
		return choices
	}
	slices.SortFunc(choices, func(a, b T) bool {
		ta := a.Title()
		tb := b.Title()
		if ta == tb {
			return txt.NaturalLess(a.BackingFilePath(), b.BackingFilePath(), true)
		}
		return txt.NaturalLess(ta, tb, true)
	})
	list := unison.NewList[T]()
	list.SetAllowMultipleSelection(true)
	list.DoubleClickCallback = func() {
		if dialog, ok := list.Window().ClientData()[unison.DialogClientDataKey].(*unison.Dialog); ok {
			dialog.Button(unison.ModalResponseOK).Click()
		}
	}
	list.Append(choices...)
	scroll := unison.NewScrollPanel()
	scroll.SetBorder(unison.NewLineBorder(unison.DividerColor, 0, unison.NewUniformInsets(1), false))
	scroll.SetContent(list, unison.FillBehavior, unison.FillBehavior)
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   unison.FillAlignment,
		VAlign:   unison.FillAlignment,
	})
	label := unison.NewLabel()
	label.Text = i18n.Text("Choose one or more destinations:")
	panel.AddChild(label)
	panel.AddChild(scroll)
	if unison.QuestionDialogWithPanel(panel) != unison.ModalResponseOK || list.Selection.Count() == 0 {
		return nil
	}
	result := make([]T, 0, list.Selection.Count())
	i := list.Selection.FirstSet()
	for i != -1 {
		result = append(result, choices[i])
		i = list.Selection.NextSet(i + 1)
	}
	return result
}
