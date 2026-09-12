// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"fmt"
	"log/slog"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xfilepath"
	"github.com/richardwilkes/toolbox/v2/xos"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/toolbox/v2/xstrings"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/side"
)

const (
	filePrefix             = "file:"
	dockGroupClientDataKey = "dock.group"
	dockableClientDataKey  = "dockable"
)

// Workspace holds the main window, its docks and the handler that errors are reported through.
var Workspace struct {
	Window       *unison.Window
	TopDock      *unison.Dock
	Navigator    *Navigator
	DocumentDock *DocumentDock
	ErrorHandler func(msg string, err error)
}

// KeyedDockable extends unison.Dockable to require a DockKey() method, used for restoring the dock layout.
type KeyedDockable interface {
	unison.Dockable
	DockKey() string
}

// GroupedCloser defines the methods required of a tab that wishes to be closed when another tab is closed.
type GroupedCloser interface {
	unison.TabCloser
	CloseWithGroup(other unison.Paneler) bool
}

// boundsDeferredDockable is implemented by a Dockable that can't report a meaningful size when it is created, because
// what it will display is still being prepared on a background goroutine. A window is sized to fit its content, so
// giving such a Dockable a window immediately would pack that window around a placeholder and leave the user with a
// uselessly small frame that nothing would ever correct. Window creation waits on the bounds instead; see
// placeInWindow.
type boundsDeferredDockable interface {
	// BoundsKnown returns true when the Dockable's preferred size reflects what it will actually display.
	BoundsKnown() bool
	// WhenBoundsKnown registers a callback to be invoked on the UI thread once the bounds become known, or once
	// waiting longer would be worse for the user than proceeding without them, whichever comes first. Each callback
	// is invoked exactly once, or not at all if the Dockable is closed before either occurs.
	WhenBoundsKnown(callback func())
}

// InitWorkspace initializes the Workspace singleton.
func InitWorkspace(wnd *unison.Window) {
	Workspace.ErrorHandler = func(msg string, err error) { errs.Log(errs.NewWithCause(msg, err)) }
	Workspace.Window = wnd
	Workspace.TopDock = unison.NewDock()
	Workspace.Navigator = newNavigator()
	Workspace.DocumentDock = NewDocumentDock()
	wnd.SetContent(Workspace.TopDock)
	Workspace.TopDock.DockTo(Workspace.Navigator, nil, side.Left)
	dc := Workspace.Navigator.Ancestor[*unison.DockContainer]()
	Workspace.TopDock.DockTo(Workspace.DocumentDock, dc, side.Right)
	dc.SetCurrentDockable(Workspace.Navigator)
	wnd.AllowCloseCallback = isWorkspaceAllowedToClose
	wnd.WillCloseCallback = workspaceWillClose
	wnd.ResizedCallback = finishInit // Doing this to get around platforms that don't immediately resize windows
	global := gurps.GlobalSettings()
	if global.WorkspaceFrame != nil {
		r := *global.WorkspaceFrame
		if r.Width < 100 {
			r.Width = 100
		}
		if r.Height < 100 {
			r.Height = 100
		}
		if d := unison.BestDisplayForRect(r); d != nil {
			r = d.FitRectOnto(r)
		}
		*global.WorkspaceFrame = r
		wnd.SetFrameRect(r)
	} else {
		wnd.SetFrameRect(primaryDisplayUsableRect())
	}
	wnd.ToFront()
}

func finishInit() {
	Workspace.Window.ResizedCallback = nil
	focused := false
	if gurps.GlobalSettings().General.RestoreWorkspaceOnStart {
		xos.SafeCall(func() { focused = restoreDockState() }, func(err error) {
			slog.Warn("Unable to restore workspace state", "error", err)
		})
	} else {
		Workspace.TopDock.RootDockLayout().SetDividerPosition(gurps.DefaultNavigatorDividerPosition)
	}
	if !focused {
		Workspace.Navigator.InitialFocus()
	}
	Workspace.ErrorHandler = func(msg string, err error) { unison.ErrorDialogWithError(msg, err) }
}

// restoreDockState puts the docks back the way they were when the workspace was last closed, reopening the files that
// were open, and returns the focus to the tab that had it then. Reports whether the focus was given to something, so
// that the caller can fall back to the navigator when it was not.
func restoreDockState() bool {
	global := gurps.GlobalSettings()
	keys := slices.Concat(dockStateKeys(global.TopDockState), dockStateKeys(global.DocDockState))
	if len(keys) == 0 {
		return false
	}
	m := make(map[string]unison.Dockable, len(keys))
	files := make([]string, 0, len(keys))
	for _, k := range keys {
		m[k] = nil
		if strings.HasPrefix(k, filePrefix) {
			files = append(files, k[len(filePrefix):])
		}
	}
	if global.TopDockState != nil {
		global.TopDockState.Apply(Workspace.TopDock, func(key string) unison.Dockable {
			switch key {
			case NavigatorDockKey:
				return Workspace.Navigator
			case DocumentsDockKey:
				return Workspace.DocumentDock
			}
			return nil
		})
	}
	OpenFiles(files)
	for _, k := range files {
		if d := LocateFileBackedDockable(k); !xreflect.IsNil(d) {
			m[filePrefix+k] = d
		} else {
			m[filePrefix+k] = newNotFoundDockable(k)
		}
	}
	if global.DocDockState != nil {
		global.DocDockState.Apply(Workspace.DocumentDock.Dock, func(key string) unison.Dockable {
			if d, ok := m[key]; ok {
				return d
			}
			slog.Warn("unable to locate dockable", "key", key)
			return nil
		})
	}
	return restoreFocusedDockable(global.FocusedDockKey, m)
}

// focusedDockKey returns the dock key of the dockable that holds the workspace window's keyboard focus, or "" when the
// focus is not within a keyed dockable, so that the tab the user was working in can be given the focus again when the
// workspace is restored. The nearest keyed dockable is the one recorded, so the focus in a file's tab names that file
// rather than the document dock that holds it, while the focus in the navigator names the navigator.
func focusedDockKey() string {
	focus := Workspace.Window.CurrentFocus()
	if focus == nil {
		return ""
	}
	if kd := unison.Ancestor[KeyedDockable](focus); !xreflect.IsNil(kd) {
		return kd.DockKey()
	}
	return ""
}

// restoreFocusedDockable gives the keyboard focus to the dockable with the given key, found in m when it is one of the
// file-backed dockables the dock state reopened, and reports whether the focus ended up within it. A dockable that
// follows the toolbar and content convention has its content focused, just as it does when it is first opened, rather
// than whatever focusable widget happens to come first in its toolbar. The navigator is focused the way it is when
// nothing is restored.
func restoreFocusedDockable(key string, m map[string]unison.Dockable) bool {
	var d unison.Dockable
	switch key {
	case "":
		return false
	case NavigatorDockKey:
		d = Workspace.Navigator
		Workspace.Navigator.InitialFocus()
	case DocumentsDockKey:
		d = Workspace.DocumentDock
		ActivateDockable(d)
	default:
		if d = m[key]; xreflect.IsNil(d) {
			return false
		}
		ActivateDockable(d)
		if children := d.AsPanel().Children(); len(children) > 1 {
			FocusFirstContent(children[0], children[1])
		}
	}
	return unison.DockableHasFocus(d)
}

// dockStateKeys returns the keys of the dockables the dock state records, in the order it records them, which is the
// order their tabs are laid out in. Reopening the files in that order, rather than in whatever order a map hands them
// back, keeps the recent files list and the order the tabs are created in the same from one start to the next.
func dockStateKeys(dockState *unison.DockState) []string {
	if dockState == nil {
		return nil
	}
	var keys []string
	if dockState.Type == unison.DockableType && dockState.Key != "" {
		keys = append(keys, dockState.Key)
	}
	for _, child := range dockState.Children {
		keys = append(keys, dockStateKeys(child)...)
	}
	return keys
}

// Activate activates and focuses the first dockable that 'matcher' returns true for, reporting whether one was found.
func Activate(matcher func(d unison.Dockable) bool) bool {
	for _, d := range AllDockables() {
		if matcher(d) {
			ActivateDockable(d)
			return true
		}
	}
	return false
}

// activateDockable is Activate for the common case of looking for a dockable of a particular type: it activates the
// first open dockable whose panel's Self is a T and, when match is not nil, that match accepts. The dockable is reached
// through its panel's Self, since what the dock hands out may be an inner layer rather than the dockable itself (see
// resolveDockable), which a direct type assertion would not see. Pass nil to match when the type alone identifies the
// dockable, as it does for the global settings views; pass a predicate when something else is part of its identity,
// such as the sheet a per-sheet settings view belongs to.
func activateDockable[T unison.Paneler](match func(T) bool) bool {
	return Activate(func(d unison.Dockable) bool {
		t, ok := d.AsPanel().Self.(T)
		return ok && (match == nil || match(t))
	})
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
	// First, close all of the non-file-backed dockables.
	for _, d := range AllDockables() {
		if tc, ok := d.(unison.TabCloser); ok {
			if _, ok = tc.(FileBackedDockable); !ok {
				if !tc.MayAttemptClose() {
					return false
				}
				if !tc.AttemptClose() {
					return false
				}
			}
		}
	}

	// Next, save any file-backed dockables that don't yet have a file on disk and are considered modified.
	for _, d := range AllDockables() {
		if tc, ok := d.(unison.TabCloser); ok {
			var fbd FileBackedDockable
			if fbd, ok = tc.(FileBackedDockable); ok {
				if !xos.FileExists(fbd.BackingFilePath()) {
					if !fbd.MayAttemptClose() {
						return false
					}
					if fbd.Modified() {
						if !AttemptSaveForDockable(fbd) {
							return false
						}
					}
					if !xos.FileExists(fbd.BackingFilePath()) {
						if !AttemptCloseForDockable(fbd) {
							return false
						}
					}
				}
			}
		}
	}

	// Then, record the current dock state, along with which of the tabs it records has the focus. The focus is looked
	// at only now, once the dockables that won't be restored have been closed, since closing one that had the focus
	// moves the focus to a neighbor, and it is that neighbor the user will find themselves in when the workspace comes
	// back.
	global := gurps.GlobalSettings()
	global.TopDockState = unison.NewDockState(Workspace.TopDock, collectDockKeys)
	global.DocDockState = unison.NewDockState(Workspace.DocumentDock.Dock, collectDockKeys)
	global.FocusedDockKey = focusedDockKey()

	// Finally, close the remaining dockables; grouped ones are closed by the dockable they are grouped with.
	for _, d := range AllDockables() {
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
	}
	return true
}

func collectDockKeys(dockable unison.Dockable) string {
	if kd, ok := dockable.(KeyedDockable); ok {
		return kd.DockKey()
	}
	return ""
}

func workspaceWillClose() {
	frame := Workspace.Window.FrameRect()
	gurps.GlobalSettings().WorkspaceFrame = &frame
	saveGlobalSettings()
}

// AllDockables returns all Dockables, whether in the workspace or in a separate window.
func AllDockables() []unison.Dockable {
	var all []unison.Dockable
	// There is no dock until the workspace has been set up.
	if Workspace.DocumentDock != nil {
		Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(dc *unison.DockContainer) bool {
			all = append(all, dc.Dockables()...)
			return false
		})
	}
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
		if dc := focus.Ancestor[*unison.DockContainer](); dc != nil && dc.Dock == Workspace.DocumentDock.Dock {
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
// present. If the group is one that is configured to open in its own window instead, and forceIntoDock is false, the
// Dockable is given a window of its own, the creation of which may be deferred; see placeInWindow.
func PlaceInDock(dockable unison.Dockable, group dgroup.Group, forceIntoDock bool) {
	InstallDockUndockCmd(dockable)
	if !forceIntoDock && slices.Contains(gurps.GlobalSettings().OpenInWindow, group) {
		if _, err := placeInWindow(dockable, group); err != nil {
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

// MoveDockableToWindow closes the tab a dockable is in within the workspace and opens a window for it instead. If
// already in its own window, does nothing. The returned window is nil if its creation had to be deferred until the
// dockable's bounds are known; see placeInWindow.
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
	return placeInWindow(dockable, group)
}

// resolveDockable returns the outermost implementation of the Dockable, i.e. its panel's Self. The code that places a
// Dockable is frequently handed an inner layer rather than the Dockable itself -- SettingsDockable.Setup, for example,
// passes its own embedded SettingsDockable rather than the settings view that contains it -- so the value has to be
// resolved before it is retained or examined, exactly as unison's dock containers do when a Dockable is docked. Without
// it, anything that type-asserts a Dockable to a concrete type or probes it for an optional interface would see a
// different value depending on whether the Dockable ended up in the workspace or in a window of its own. The Dockable
// is returned unchanged if its Self isn't a Dockable.
func resolveDockable(dockable unison.Dockable) unison.Dockable {
	if xreflect.IsNil(dockable) {
		return dockable
	}
	if resolved, ok := dockable.AsPanel().Self.(unison.Dockable); ok {
		return resolved
	}
	return dockable
}

// placeInWindow gives the Dockable a window of its own. A Dockable whose bounds aren't known yet has that window
// created later, once they are, since packing a window around content that can't yet say how big it is would produce a
// frame the user would have to fix by hand. Nothing is shown for the Dockable in the meantime, which is why the wait a
// boundsDeferredDockable permits is a short one. A nil window is returned when the creation was deferred, since there
// is no window to return yet.
func placeInWindow(dockable unison.Dockable, group dgroup.Group) (*unison.Window, error) {
	dockable = resolveDockable(dockable)
	if deferred, ok := dockable.(boundsDeferredDockable); ok && !deferred.BoundsKnown() {
		deferred.WhenBoundsKnown(func() {
			if dockable.AsPanel().Window() != nil {
				// It was given a window by some other means while the wait was on, so there is nothing left to do.
				return
			}
			if _, err := NewWindowForDockable(dockable, group); err != nil {
				errs.Log(err)
				return
			}
			// There was no window when the Dockable was placed, so any request to focus its content made then went
			// nowhere. Make it now, so a deferred window ends up with the focus where an immediate one would have it.
			if children := dockable.AsPanel().Children(); len(children) > 1 {
				FocusFirstContent(children[0], children[1])
			}
		})
		return nil, nil
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
	// Resolve before the Dockable is recorded on the window and before its closing behavior is hooked up, since both
	// the value dockableFromWindow hands back and the TabCloser the window closes through must be the Dockable itself
	// rather than one of its inner layers.
	dockable = resolveDockable(dockable)
	frame := windowPlacementFrame()
	wnd, err := unison.NewWindow(dockable.Title())
	if err != nil {
		return nil, err
	}
	registerWindowDragTypes(wnd)
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
		pendingClose := false
		wnd.AllowCloseCallback = func() bool {
			if !tc.MayAttemptClose() {
				return false
			}
			if !pendingClose {
				pendingClose = true
				defer func() { pendingClose = false }()
				return tc.AttemptClose()
			}
			return true
		}
	}
	panel.ClientData()[dockGroupClientDataKey] = group
	InstallDockUndockCmd(dockable)
	// The window is packed at the location it is going to occupy, not where it happens to have been created: packing
	// clamps the window to the display it is on at the time, and a freshly created window may be sitting on a smaller
	// display than the one it is about to be moved to. Packing first and moving after would carry that smaller
	// display's clamp along, leaving the window shorter than its content asked for even though its eventual display
	// had the room.
	wnd.PackWithLocation(frame.Point)
	widenToFitTooltips(wnd, panel)
	placeWindowOver(wnd, frame)
	return wnd, nil
}

// widenToFitTooltips widens the window, which must already be packed, so that its content is at least as wide as the
// widest tooltip the panel or its descendants carry. A tooltip wider than the window it is shown in is squeezed into
// that window's width, and an editor's tooltips are typically wider than the fields they explain, so a window packed
// around the fields alone, as the ancestry and name generator editors' windows were, would show none of its longer
// tooltips as written. The window is still clamped onto its display afterwards by placeWindowOver.
func widenToFitTooltips(wnd *unison.Window, panel *unison.Panel) {
	r := wnd.ContentRect()
	if width := widestTooltipWidth(panel); r.Width < width {
		r.Width = width
		wnd.SetContentRect(r)
	}
}

// widestTooltipWidth returns the preferred width of the widest tooltip the panel or any of its descendants carry, or 0
// when none of them has one.
func widestTooltipWidth(panel *unison.Panel) float32 {
	var width float32
	panel.HasInSelfOrDescendants(func(p *unison.Panel) bool {
		if p.Tooltip != nil {
			_, pref, _ := p.Tooltip.Sizes(geom.Size{})
			width = max(width, pref.Width)
		}
		return false
	})
	return width
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

// AssociatedIDKey is the key used with CloseID().
const AssociatedIDKey = "associated_id"

// CloseID attempts to close any Dockables associated with the given TIDs. Returns false if a dockable refused to close.
func CloseID(ids map[tid.TID]bool) bool {
	for _, d := range AllDockables() {
		if tc, ok := d.(unison.TabCloser); ok {
			if otherValue, ok2 := d.AsPanel().ClientData()[AssociatedIDKey]; ok2 {
				if otherID, ok3 := otherValue.(tid.TID); ok3 && ids[otherID] {
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
		Workspace.ErrorHandler(fmt.Sprintf(i18n.Text("Unable to save %s"), xfilepath.BaseName(filePath)), err)
		return false
	}
	setUnmodified()
	UpdateTitleForDockable(d)
	return true
}

// SaveDockableAs attempts to save the contents of the dockable, prompting for a new path.
func SaveDockableAs(d FileBackedDockable, extension string, saver func(filePath string) error, setUnmodifiedAndNewPath func(filePath string)) bool {
	return saveDockableAs(d, extension, func() string { return gurps.GlobalSettings().LastDir(gurps.DefaultLastDirKey) },
		gurps.DefaultLastDirKey, saver, setUnmodifiedAndNewPath)
}

// saveDockableAs is SaveDockableAs with the directory the dialog starts in when the dockable has no file on disk yet,
// and the last-directory key that records where the user saved, chosen by the caller. The directory is asked for only
// when it is needed, since finding it may have side effects, such as creating it.
func saveDockableAs(d FileBackedDockable, extension string, fallbackDir func() string, lastDirKey string, saver func(filePath string) error, setUnmodifiedAndNewPath func(filePath string)) bool {
	existingPath := d.BackingFilePath()
	var initialDir string
	if !strings.HasPrefix(existingPath, markdownContentOnlyPrefix) && xos.FileExists(existingPath) {
		initialDir = filepath.Dir(existingPath)
	} else {
		initialDir = fallbackDir()
	}
	filePath, ok := chooseFileToSave(initialDir, xfilepath.BaseName(existingPath), extension, lastDirKey)
	if !ok {
		return false
	}
	if err := saver(filePath); err != nil {
		Workspace.ErrorHandler(i18n.Text("Unable to save as ")+xfilepath.BaseName(filePath), err)
		return false
	}
	setUnmodifiedAndNewPath(filePath)
	gurps.GlobalSettings().AddRecentFile(filePath)
	UpdateTitleForDockable(d)
	return true
}

// PromptForDestination puts up a modal dialog to choose one or more destinations when choices holds more than one, and
// returns choices unchanged otherwise. Returns nil if the dialog was canceled or nothing was selected.
func PromptForDestination[T FileBackedDockable](choices []T) []T {
	if len(choices) < 2 {
		return choices
	}
	slices.SortFunc(choices, func(a, b T) int {
		ta := a.Title()
		tb := b.Title()
		if ta == tb {
			return xstrings.NaturalCmp(a.BackingFilePath(), b.BackingFilePath(), true)
		}
		return xstrings.NaturalCmp(ta, tb, true)
	})
	list := unison.NewList[T]()
	list.SetAllowMultipleSelection(true)
	list.DoubleClickCallback = func() {
		if dialog, ok := list.Window().ClientData()[unison.DialogClientDataKey].(*unison.Dialog); ok {
			dialog.Button(unison.ModalResponseOK).Click()
		}
	}
	list.Append(choices...)
	if !showListQuestionDialog(i18n.Text("Choose one or more destinations:"), list) || list.Selection.Count() == 0 {
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

// MarkRootAncestorForLayoutRecursively marks the nearest ancestor DockContainer (or, failing that, the nearest ancestor
// Dockable) and all of its descendants as needing to be laid out.
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
	} else if wnd := d.AsPanel().Window(); wnd != nil {
		var buffer strings.Builder
		if d.Modified() {
			buffer.WriteByte('*')
		}
		buffer.WriteString(d.Title())
		wnd.SetTitle(buffer.String())
	}
}

// AttemptCloseForDockable attempts to close a dockable.
func AttemptCloseForDockable(d unison.Dockable) bool {
	if dc := unison.Ancestor[*unison.DockContainer](d); dc != nil {
		dc.Close(d)
		return true
	}
	if wnd := d.AsPanel().Window(); wnd != nil {
		return wnd.AttemptClose()
	}
	return true
}

type saveable interface {
	save(bool) bool
}

// AttemptSaveForDockable closes the dockables grouped with the dockable, then, if it is modified, asks whether to save
// it. Returns false if something refused to close, the save failed, or the user canceled.
func AttemptSaveForDockable(d unison.Dockable) bool {
	if !CloseGroup(d) {
		return false
	}
	if !d.Modified() {
		return true
	}
	s, ok := d.(saveable)
	if !ok {
		return true
	}
	switch unison.YesNoCancelDialog(fmt.Sprintf(i18n.Text("Save changes made to\n%s?"), d.Title()), "") {
	case unison.ModalResponseDiscard:
	case unison.ModalResponseOK:
		if !s.save(false) {
			return false
		}
	case unison.ModalResponseCancel:
		return false
	}
	return true
}
