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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// TestWorkspaceRestoresTheFocusedTab verifies that closing the workspace records which tab has the keyboard focus along
// with the dock state, and that restoring the dock state hands the focus back to that tab: a sheet when a sheet had it,
// the navigator's table when the navigator had it, and nothing when nothing was recorded, so that startup can fall
// back to the navigator as it always has. The files are spread over two dock containers, since each container already
// remembers which of its own tabs is in front; what has to be remembered is which container the focus was in.
func TestWorkspaceRestoresTheFocusedTab(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	swapForTest(t, &global.TopDockState, nil)
	swapForTest(t, &global.DocDockState, nil)
	swapForTest(t, &global.FocusedDockKey, "")
	swapForTest(t, &global.OpenInWindow, nil) // The files must land in the dock for the dock state to record them.

	// Two sheets share a container, and the trait library is put in a container of its own to their right, so the
	// library is the last tab reopened whenever the dock state is restored.
	dir := t.TempDir()
	first := filepath.Join(dir, "first"+gurps.SheetExt)
	second := filepath.Join(dir, "second"+gurps.SheetExt)
	library := filepath.Join(dir, "library"+gurps.TraitsExt)
	c.NoError(gurps.NewEntity().Save(first))
	c.NoError(gurps.NewEntity().Save(second))
	c.NoError(gurps.SaveTraits([]*gurps.Trait{gurps.NewTrait(nil, nil, false)}, library))

	// Closing the workspace with the first sheet focused records that sheet.
	var allowed bool
	var key string
	var open, containers int
	screen.Do(func() {
		OpenFiles([]string{first, second, library})
		containers = dockContainerCount()
		ActivateDockable(LocateFileBackedDockable(first))
		allowed = isWorkspaceAllowedToClose()
		key = global.FocusedDockKey
		open = len(AllDockables())
	})
	c.Equal(2, containers, "the sheets and the library must be in separate containers")
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(filePrefix+first, key, "closing the workspace must record the tab that has the focus")
	c.Equal(0, open, "closing the workspace must close its tabs")

	// Restoring reopens everything and hands the focus back to the first sheet, not to the library that was reopened
	// last.
	var restored bool
	var focusedPath string
	screen.Do(func() {
		restored = restoreDockState()
		focusedPath = focusedFilePath(wnd)
		open = len(AllDockables())
		containers = dockContainerCount()
	})
	c.True(restored, "restoring must report that the recorded tab was given the focus")
	c.Equal(3, open, "restoring must reopen every file")
	c.Equal(2, containers, "restoring must put the containers back")
	c.Equal(first, focusedPath, "restoring must hand the focus back to the sheet that had it")

	// The library in the other container is recorded and restored the same way.
	screen.Do(func() {
		ActivateDockable(LocateFileBackedDockable(library))
		allowed = isWorkspaceAllowedToClose()
		key = global.FocusedDockKey
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(filePrefix+library, key, "closing with the library focused must record the library")
	screen.Do(func() {
		restored = restoreDockState()
		focusedPath = focusedFilePath(wnd)
	})
	c.True(restored, "restoring must report that the library was given the focus")
	c.Equal(library, focusedPath, "restoring must hand the focus back to the library")

	// The navigator is recorded and restored the same way, ending up with its table focused, as at a fresh start.
	var navigatorTableFocused bool
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		allowed = isWorkspaceAllowedToClose()
		key = global.FocusedDockKey
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(NavigatorDockKey, key, "closing with the navigator focused must record the navigator")
	screen.Do(func() {
		restored = restoreDockState()
		navigatorTableFocused = wnd.CurrentFocus() == Workspace.Navigator.table.AsPanel()
		open = len(AllDockables())
	})
	c.True(restored, "restoring must report that the navigator was given the focus")
	c.True(navigatorTableFocused, "restoring must hand the focus back to the navigator's table")
	c.Equal(3, open, "restoring must reopen every file")

	// With nothing recorded, the files still come back, but the focus is left for the caller to place.
	screen.Do(func() {
		allowed = isWorkspaceAllowedToClose()
		global.FocusedDockKey = ""
		restored = restoreDockState()
		open = len(AllDockables())
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.False(restored, "restoring with no focus recorded must not claim to have placed the focus")
	c.Equal(3, open, "restoring must reopen every file")

	screen.Do(func() { isWorkspaceAllowedToClose() }) // Leave nothing open for the session to stop with.
}

// dockContainerCount returns the number of containers in the document dock.
func dockContainerCount() int {
	count := 0
	Workspace.DocumentDock.RootDockLayout().ForEachDockContainer(func(_ *unison.DockContainer) bool {
		count++
		return false
	})
	return count
}

// focusedFilePath returns the path of the file-backed dockable that holds wnd's keyboard focus, or "" if the focus is
// not within one.
func focusedFilePath(wnd *unison.Window) string {
	if focus := wnd.CurrentFocus(); focus != nil {
		if fbd := unison.Ancestor[FileBackedDockable](focus); !xreflect.IsNil(fbd) {
			return fbd.BackingFilePath()
		}
	}
	return ""
}
