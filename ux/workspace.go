/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/side"
	"github.com/richardwilkes/unison/printing"
)

const (
	dockGroupClientDataKey = "dock.group"
	dockableClientDataKey  = "dockable"
)

// Workspace holds the data necessary to track the Workspace.
var Workspace struct {
	Window       *unison.Window
	TopDock      *unison.Dock
	Navigator    *Navigator
	DocumentDock *DocumentDock
	PrintMgr     printing.PrintManager
}

// GroupedCloser defines the methods required of a tab that wishes to be closed when another tab is closed.
type GroupedCloser interface {
	unison.TabCloser
	CloseWithGroup(other unison.Paneler) bool
}

// InitWorkspace initializes the Workspace singleton.
func InitWorkspace(wnd *unison.Window) {
	Workspace.Window = wnd
	Workspace.TopDock = unison.NewDock()
	Workspace.Navigator = newNavigator()
	Workspace.DocumentDock = NewDocumentDock()
	wnd.SetContent(Workspace.TopDock)
	Workspace.TopDock.DockTo(Workspace.Navigator, nil, side.Left)
	dc := unison.Ancestor[*unison.DockContainer](Workspace.Navigator)
	Workspace.TopDock.DockTo(Workspace.DocumentDock, dc, side.Right)
	dc.SetCurrentDockable(Workspace.Navigator)
	wnd.AllowCloseCallback = isWorkspaceAllowedToClose
	wnd.WillCloseCallback = workspaceWillClose
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
	Workspace.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	unison.InvokeTaskAfter(func() {
		Workspace.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	}, time.Millisecond)
	wnd.ToFront()
	Workspace.Navigator.InitialFocus()
}

// Activate attempts to locate an existing dockable that 'matcher' returns true for. Will return true if a suitable
// match was found. If found, it will have been activated and focused.
func Activate(matcher func(d unison.Dockable) bool) bool {
	for _, d := range AllDockables() {
		if matcher(d) {
			ActivateDockable(d)
			return true
		}
	}
	return false
}

// ActivateDockable activates the dockable, giving it focus.
func ActivateDockable(d unison.Dockable) {
	if dc := unison.Ancestor[*unison.DockContainer](d.AsPanel()); dc != nil {
		dc.SetCurrentDockable(d)
		dc.AcquireFocus()
		return
	}
	d.AsPanel().Window().ToFront()
}

// ActiveDockable returns the currently active dockable in the active window.
func ActiveDockable() unison.Dockable {
	wnd := unison.ActiveWindow()
	if wnd == nil {
		return nil
	}
	if Workspace.Window == wnd {
		dc := CurrentlyFocusedDockContainer()
		if dc == nil {
			return nil
		}
		return dc.CurrentDockable()
	}
	return dockableFromWindow(wnd)
}

func isWorkspaceAllowedToClose() bool {
	for _, d := range AllDockables() {
		if !mayDockableClose(d) {
			return false
		}
	}
	return true
}

func workspaceWillClose() {
	global := gurps.GlobalSettings()
	global.LibraryExplorer.OpenRowKeys = Workspace.Navigator.DisclosedPaths()
	global.LibraryExplorer.DividerPosition = Workspace.TopDock.RootDockLayout().DividerPosition()
	frame := Workspace.Window.FrameRect()
	global.WorkspaceFrame = &frame
	if err := global.Save(); err != nil {
		unison.ErrorDialogWithError(i18n.Text("Unable to save global settings"), err)
	}
}

func mayDockableClose(d unison.Dockable) bool {
	if tc, ok := d.(unison.TabCloser); ok {
		if _, ok = d.(GroupedCloser); !ok {
			if !tc.MayAttemptClose() {
				return false
			}
			if !tc.AttemptClose() {
				return false
			}
		}
	}
	return true
}

// AllDockables returns all Dockables, whether in the workspace or in a separate window.
func AllDockables() []unison.Dockable {
	var all []unison.Dockable
	Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		all = append(all, dc.Dockables()...)
		return false
	})
	for _, wnd := range unison.Windows() {
		if wnd != Workspace.Window {
			if d := dockableFromWindow(wnd); d != nil {
				all = append(all, d)
			}
		}
	}
	return all
}

// AllMatchingDockables returns all Dockables that 'matcher' returns true for.
func AllMatchingDockables(matcher func(d unison.Dockable) bool) []unison.Dockable {
	var result []unison.Dockable
	for _, d := range AllDockables() {
		if matcher(d) {
			result = append(result, d)
		}
	}
	return result
}

func dockableFromWindow(wnd *unison.Window) unison.Dockable {
	if clientData, ok := wnd.ClientData()[dockableClientDataKey]; ok {
		if d, ok2 := clientData.(unison.Dockable); ok2 {
			return d
		}
	}
	return nil
}

// IsDockableInWorkspace returns true if the Dockable is inside the Workspace as opposed to an external window.
func IsDockableInWorkspace(d unison.Dockable) bool {
	return d.AsPanel().Window() == Workspace.Window
}

// CurrentlyFocusedDockContainer returns the currently focused DockContainer, if any.
func CurrentlyFocusedDockContainer() *unison.DockContainer {
	if focus := Workspace.Window.Focus(); focus != nil {
		if dc := unison.Ancestor[*unison.DockContainer](focus); dc != nil && dc.Dock == Workspace.DocumentDock.Dock {
			return dc
		}
	}
	return nil
}

// DefaultDockContainer returns the currently focused DockContainer, if possible. If not, returns the first
// DockContainer that can be found.
func DefaultDockContainer() *unison.DockContainer {
	dc := CurrentlyFocusedDockContainer()
	if dc == nil {
		Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(container *unison.DockContainer) bool {
			dc = container
			return true
		})
	}
	return dc
}

// LocateFileBackedDockable searches for a FileBackedDockable with the given path.
func LocateFileBackedDockable(filePath string) FileBackedDockable {
	for _, d := range AllDockables() {
		if fbd, ok := d.(FileBackedDockable); ok && filePath == fbd.BackingFilePath() {
			return fbd
		}
	}
	return nil
}

// LocateDockContainerForExtension searches for the first FileBackedDockable with the given extension and returns its
// DockContainer.
func LocateDockContainerForExtension(ext ...string) *unison.DockContainer {
	var extDC *unison.DockContainer
	Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		if DockContainerHoldsExtension(dc, ext...) {
			extDC = dc
			return true
		}
		return false
	})
	return extDC
}

// PlaceInDock places the Dockable into the workspace document dock, grouped with the provided group, if that group is
// present.
func PlaceInDock(dockable unison.Dockable, group dgroup.Group, forceIntoDock bool) {
	InstallDockUndockCmd(dockable)
	if !forceIntoDock && slices.Contains(gurps.GlobalSettings().OpenInWindow, group) {
		if _, err := NewWindowForDockable(dockable, group); err != nil {
			errs.Log(err)
		}
		return
	}
	dockable.AsPanel().ClientData()[dockGroupClientDataKey] = group
	dc := DefaultDockContainer()
	if dc != nil {
		if DockContainerHasGroup(dc, group) {
			dc.Stack(dockable, -1)
			return
		}
	}
	if dc = DockContainerForGroup(Workspace.DocumentDock.Dock, group); dc != nil {
		dc.Stack(dockable, -1)
		return
	}
	s := side.Right
	if group == dgroup.SubEditors {
		if dc = DockContainerForGroup(Workspace.DocumentDock.Dock, dgroup.Editors); dc != nil {
			s = side.Bottom
		}
	}
	Workspace.DocumentDock.DockTo(dockable, dc, s)
}

// MoveDockableToWorkspace closes the window a dockable is in and places it within the workspace. If already in the
// workspace, does nothing.
func MoveDockableToWorkspace(dockable unison.Dockable) {
	panel := dockable.AsPanel()
	wnd := panel.Window()
	if wnd == Workspace.Window {
		return
	}
	if wnd != nil {
		wnd.WillCloseCallback = nil
		wnd.Dispose()
	}
	panel.RemoveFromParent()
	group, ok := panel.ClientData()[dockGroupClientDataKey].(dgroup.Group)
	if !ok {
		group = dgroup.Editors // Arbitrary
	}
	PlaceInDock(dockable, group, true)
}

// MoveDockableToWindow closes the tab a dockable is in within the workspace and opens a windows for it instead. If
// already in its own window, does nothing.
func MoveDockableToWindow(dockable unison.Dockable) (*unison.Window, error) {
	panel := dockable.AsPanel()
	wnd := panel.Window()
	if wnd != Workspace.Window {
		return wnd, nil
	}
	if dc := unison.Ancestor[*unison.DockContainer](dockable); dc != nil {
		dc.Close(dockable)
	} else {
		panel.RemoveFromParent()
	}
	panel.Hidden = false
	group, ok := panel.ClientData()[dockGroupClientDataKey].(dgroup.Group)
	if !ok {
		group = dgroup.Editors // Arbitrary
	}
	return NewWindowForDockable(dockable, group)
}

// InstallDockUndockCmd installs the dock or undock command handler.
func InstallDockUndockCmd(dockable unison.Dockable) {
	panel := dockable.AsPanel()
	panel.InstallCmdHandlers(DockUnDockItemID,
		func(_ any) bool {
			if panel.Window() == Workspace.Window {
				dockUnDockAction.Title = i18n.Text("Undock From Workspace")
			} else {
				dockUnDockAction.Title = i18n.Text("Dock Into Workspace")
			}
			return true
		},
		func(_ any) {
			if panel.Window() == Workspace.Window {
				if _, err := MoveDockableToWindow(dockable); err != nil {
					errs.Log(err)
				}
			} else {
				MoveDockableToWorkspace(dockable)
			}
		})
}

// NewWindowForDockable creates a new window and places a Dockable inside it.
func NewWindowForDockable(dockable unison.Dockable, group dgroup.Group) (*unison.Window, error) {
	var frame unison.Rect
	if focused := unison.ActiveWindow(); focused != nil {
		frame = focused.FrameRect()
	} else {
		frame = unison.PrimaryDisplay().Usable
	}
	wnd, err := unison.NewWindow(dockable.Title())
	if err != nil {
		return nil, err
	}
	SetupMenuBar(wnd)
	content := wnd.Content()
	content.SetLayout(&unison.FlexLayout{Columns: 1})
	panel := dockable.AsPanel()
	panel.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  1,
		VSpan:  1,
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	content.AddChild(panel)
	wnd.ClientData()[dockableClientDataKey] = dockable
	if tc, ok := dockable.(unison.TabCloser); ok {
		wnd.AllowCloseCallback = tc.MayAttemptClose
		pendingClose := false
		wnd.WillCloseCallback = func() {
			if !pendingClose {
				pendingClose = true
				tc.AttemptClose()
				pendingClose = false
			}
		}
	}
	panel.ClientData()[dockGroupClientDataKey] = group
	InstallDockUndockCmd(dockable)
	wnd.Pack()
	wndFrame := wnd.FrameRect()
	frame.Y += (frame.Height - wndFrame.Height) / 3
	frame.Height = wndFrame.Height
	frame.X += (frame.Width - wndFrame.Width) / 2
	frame.Width = wndFrame.Width
	frame = frame.Align()
	wnd.SetFrameRect(unison.BestDisplayForRect(frame).FitRectOnto(frame))
	wnd.ToFront()
	return wnd, nil
}

// DockContainerHasGroup returns true if the DockContainer contains at least one Dockable associated with the given
// group. May pass nil for the dc.
func DockContainerHasGroup(dc *unison.DockContainer, group dgroup.Group) bool {
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
func DockContainerForGroup(dock *unison.Dock, group dgroup.Group) *unison.DockContainer {
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
	for _, d := range AllDockables() {
		if tc, ok := d.(unison.TabCloser); ok {
			if otherValue, ok2 := d.AsPanel().ClientData()[AssociatedUUIDKey]; ok2 {
				if otherID, ok3 := otherValue.(uuid.UUID); ok3 && ids[otherID] {
					if !tc.MayAttemptClose() {
						return false
					}
					if !tc.AttemptClose() {
						return false
					}
				}
			}
		}
	}
	return true
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
	for _, other := range AllDockables() {
		if fe, ok := other.(GroupedCloser); ok && fe.CloseWithGroup(d) {
			if f(fe) {
				return
			}
		}
	}
}

// SaveDockable attempts to save the contents of the dockable using its existing path.
func SaveDockable(d FileBackedDockable, saver func(filePath string) error, setUnmodified func()) bool {
	filePath := d.BackingFilePath()
	if err := saver(filePath); err != nil {
		unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to save %s"), fs.BaseName(filePath)), err)
		return false
	}
	setUnmodified()
	UpdateTitleForDockable(d)
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
	dialog.SetInitialFileName(fs.SanitizeName(fs.BaseName(existingPath)))
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
		UpdateTitleForDockable(d)
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
	slices.SortFunc(choices, func(a, b T) int {
		ta := a.Title()
		tb := b.Title()
		if ta == tb {
			return txt.NaturalCmp(a.BackingFilePath(), b.BackingFilePath(), true)
		}
		return txt.NaturalCmp(ta, tb, true)
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
	scroll.SetContent(list, behavior.Fill, behavior.Fill)
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	panel := unison.NewPanel()
	panel.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
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

// MarkRootAncestorForLayoutRecursively looks for a parent DockContainer (and, failing to find one of those, a parent
// Dockable) and marks it and all of its descendents as needing to be laid out.
func MarkRootAncestorForLayoutRecursively(p unison.Paneler) {
	if dc := unison.Ancestor[*unison.DockContainer](p); dc != nil {
		dc.MarkForLayoutRecursively()
	} else if d := unison.Ancestor[unison.Dockable](p); d != nil {
		d.AsPanel().MarkForLayoutRecursively()
	}
}

// UpdateTitleForDockable updates the title for the given Dockable, whether it is within the workspace or a separate
// window.
func UpdateTitleForDockable(d unison.Dockable) {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	} else {
		var buffer strings.Builder
		if d.Modified() {
			buffer.WriteByte('*')
		}
		buffer.WriteString(d.Title())
		d.AsPanel().Window().SetTitle(buffer.String())
	}
}

// AttemptCloseForDockable attempts to close a dockable.
func AttemptCloseForDockable(d unison.Dockable) bool {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
		return true
	}
	return d.AsPanel().Window().AttemptClose()
}
