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

package workspace

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/richardwilkes/gcs/v5/model/settings"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
)

const workspaceClientDataKey = "workspace"

// Workspace holds the data necessary to track the Workspace.
type Workspace struct {
	Window       *unison.Window
	TopDock      *unison.Dock
	Navigator    *Navigator
	DocumentDock *DocumentDock
}

// ShowUnableToLocateWorkspaceError displays an error dialog.
func ShowUnableToLocateWorkspaceError() {
	unison.ErrorDialogWithMessage(i18n.Text("Unable to locate workspace"), "")
}

// Activate attempts to locate an existing dockable that 'matcher' returns true for. If found, it will have been
// activated and focused.
func Activate(matcher func(d unison.Dockable) bool) (ws *Workspace, dc *unison.DockContainer, found bool) {
	if ws = Any(); ws == nil {
		jot.Error("no workspace available")
		return nil, nil, false
	}
	dc = ws.CurrentlyFocusedDockContainer()
	ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(container *unison.DockContainer) bool {
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
	return ws, dc, found
}

// ActiveDockable returns the currently active dockable in the active window.
func ActiveDockable() unison.Dockable {
	ws := FromWindow(unison.ActiveWindow())
	if ws == nil {
		return nil
	}
	dc := ws.CurrentlyFocusedDockContainer()
	if dc == nil {
		return nil
	}
	return dc.CurrentDockable()
}

// FromWindowOrAny first calls FromWindow(wnd) and if that fails to find a Workspace, then calls Any().
func FromWindowOrAny(wnd *unison.Window) *Workspace {
	ws := FromWindow(wnd)
	if ws == nil {
		ws = Any()
	}
	return ws
}

// FromWindow returns the Workspace associated with the given Window, or nil.
func FromWindow(wnd *unison.Window) *Workspace {
	if wnd != nil {
		if data, ok := wnd.ClientData()[workspaceClientDataKey]; ok {
			if w, ok2 := data.(*Workspace); ok2 {
				return w
			}
		}
	}
	return nil
}

// Any first tries to return the workspace for the active window. If that fails, then it looks for any available
// workspace and returns that.
func Any() *Workspace {
	if ws := FromWindow(unison.ActiveWindow()); ws != nil {
		return ws
	}
	for _, wnd := range unison.Windows() {
		if ws := FromWindow(wnd); ws != nil {
			return ws
		}
	}
	return nil
}

// NewWorkspace creates a new Workspace for the given Window.
func NewWorkspace(wnd *unison.Window) *Workspace {
	w := &Workspace{
		Window:       wnd,
		TopDock:      unison.NewDock(),
		Navigator:    newNavigator(),
		DocumentDock: NewDocumentDock(),
	}
	wnd.SetContent(w.TopDock)
	w.TopDock.DockTo(w.Navigator, nil, unison.LeftSide)
	dc := unison.Ancestor[*unison.DockContainer](w.Navigator)
	w.TopDock.DockTo(w.DocumentDock, dc, unison.RightSide)
	dc.SetCurrentDockable(w.Navigator)
	wnd.ClientData()[workspaceClientDataKey] = w
	wnd.AllowCloseCallback = w.allowClose
	wnd.WillCloseCallback = w.willClose
	global := settings.Global()
	if global.WorkspaceFrame != nil {
		wnd.SetFrameRect(*global.WorkspaceFrame)
	} else {
		wnd.SetFrameRect(unison.PrimaryDisplay().Usable)
	}
	// On some platforms, this needs to be done after a delay... but we do it without the delay, too, so that
	// well-behaved platforms don't flash
	w.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	unison.InvokeTaskAfter(func() {
		w.TopDock.RootDockLayout().SetDividerPosition(global.LibraryExplorer.DividerPosition)
	}, time.Millisecond)
	wnd.ToFront()
	return w
}

func (w *Workspace) allowClose() bool {
	may := true
	w.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
		for _, other := range dc.Dockables() {
			if tc, ok2 := other.(unison.TabCloser); ok2 {
				if _, ok := other.(widget.GroupedCloser); !ok {
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
	global := settings.Global()
	global.LibraryExplorer.OpenRowKeys = w.Navigator.DisclosedPaths()
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
	for _, wnd := range unison.Windows() {
		if ws := FromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
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
		}
	}
	return allow
}

// MayAttemptCloseOfGroup returns true if the grouped Dockables associated with the given dockable may be closed.
func MayAttemptCloseOfGroup(d unison.Dockable) bool {
	allow := true
	traverseGroup(d, func(target widget.GroupedCloser) bool {
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
	traverseGroup(d, func(target widget.GroupedCloser) bool {
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

func traverseGroup(d unison.Dockable, f func(target widget.GroupedCloser) bool) {
	for _, wnd := range unison.Windows() {
		if ws := FromWindow(wnd); ws != nil {
			ws.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
				for _, other := range dc.Dockables() {
					if fe, ok := other.(widget.GroupedCloser); ok && fe.CloseWithGroup(d) {
						if f(fe) {
							return true
						}
					}
				}
				return false
			})
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
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.UpdateTitle(d)
	}
	return true
}

// SaveDockableAs attempts to save the contents of the dockable, prompting for a new path.
func SaveDockableAs(d FileBackedDockable, extension string, saver func(filePath string) error, setUnmodifiedAndNewPath func(filePath string)) bool {
	dialog := unison.NewSaveDialog()
	existingPath := d.BackingFilePath()
	dir := filepath.Dir(existingPath)
	if existingPath != dir {
		dialog.SetInitialDirectory(dir)
	}
	dialog.SetAllowedExtensions(extension)
	if dialog.RunModal() {
		filePath := dialog.Path()
		if err := saver(filePath); err != nil {
			unison.ErrorDialogWithError(fmt.Sprintf(i18n.Text("Unable to save as %s"), fs.BaseName(filePath)), err)
			return false
		}
		setUnmodifiedAndNewPath(filePath)
		settings.Global().AddRecentFile(filePath)
		if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
			dc.UpdateTitle(d)
		}
		return true
	}
	return false
}
