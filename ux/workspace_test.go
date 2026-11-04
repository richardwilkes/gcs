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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// TestResolveDockable verifies that a Dockable is resolved to its outermost implementation, since the placement code is
// handed the embedded SettingsDockable rather than the settings view itself, and that this makes the optional
// interfaces only the settings view implements visible again. A sheet's settings view given a window of its own was
// otherwise no GroupedCloser, so traverseGroup skipped it and left it open when the sheet that owns it was closed.
func TestResolveDockable(t *testing.T) {
	c := check.New(t)
	d := &pageRefMappingsDockable{}
	d.Self = d
	c.Equal(d, resolveDockable(&d.SettingsDockable), "an inner layer resolves to the Dockable that contains it")
	c.Equal(d, resolveDockable(d), "an already-resolved Dockable is returned unchanged")
	c.Nil(resolveDockable(nil), "a nil Dockable is returned unchanged")

	s := &sheetSettingsDockable{}
	s.Self = s
	var placed unison.Dockable = &s.SettingsDockable
	_, ok := placed.(GroupedCloser)
	c.False(ok, "the embedded SettingsDockable is not a GroupedCloser")
	_, ok = resolveDockable(placed).(GroupedCloser)
	c.True(ok, "the resolved settings view is a GroupedCloser")
}

// TestWorkspaceRestoresTheFocusedTab verifies that closing the workspace records which tab has the keyboard focus along
// with the dock state, and that restoring the dock state hands the focus back to that tab when a file had it. When the
// navigator had it, or nothing was recorded, the focus goes to the front tab of the first document container instead,
// and only a workspace that comes back with no documents at all starts in the navigator. The files are spread over two
// dock containers, since each container already remembers which of its own tabs is in front; what has to be remembered
// is which container the focus was in.
func TestWorkspaceRestoresTheFocusedTab(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.TopDockState, nil)
	uxtest.SwapForTest(t, &global.DocDockState, nil)
	uxtest.SwapForTest(t, &global.FocusedDockKey, "")
	uxtest.SwapForTest(t, &global.OpenInWindow, nil) // The files must land in the dock for the dock state to record them.

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

	// The navigator is recorded, but since documents come back, the focus goes to the front tab of the first document
	// container -- the first sheet, in the container to the left of the library's -- rather than to the navigator.
	screen.Do(func() {
		Workspace.Navigator.InitialFocus()
		allowed = isWorkspaceAllowedToClose()
		key = global.FocusedDockKey
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(NavigatorDockKey, key, "closing with the navigator focused must record the navigator")
	screen.Do(func() {
		restored = restoreDockState()
		focusedPath = focusedFilePath(wnd)
		open = len(AllDockables())
	})
	c.True(restored, "restoring with the navigator recorded must report that a document was given the focus")
	c.Equal(first, focusedPath, "restoring with the navigator recorded must focus the front tab of the first container")
	c.Equal(3, open, "restoring must reopen every file")

	// With nothing recorded, the files still come back and the focus goes to the same document.
	screen.Do(func() {
		allowed = isWorkspaceAllowedToClose()
		global.FocusedDockKey = ""
		restored = restoreDockState()
		focusedPath = focusedFilePath(wnd)
		open = len(AllDockables())
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.True(restored, "restoring with no focus recorded must report that a document was given the focus")
	c.Equal(first, focusedPath, "restoring with no focus recorded must focus the front tab of the first container")
	c.Equal(3, open, "restoring must reopen every file")

	// Only a workspace that comes back with no documents starts in the navigator, with its table focused as at a fresh
	// start. Closing once leaves nothing open, so closing again records a dock state without files.
	var navigatorTableFocused bool
	screen.Do(func() {
		isWorkspaceAllowedToClose()
		Workspace.Navigator.InitialFocus()
		allowed = isWorkspaceAllowedToClose()
		key = global.FocusedDockKey
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(NavigatorDockKey, key, "closing with the navigator focused and nothing open must record the navigator")
	screen.Do(func() {
		restored = restoreDockState()
		navigatorTableFocused = wnd.CurrentFocus() == Workspace.Navigator.table.AsPanel()
		open = len(AllDockables())
	})
	c.True(restored, "restoring with no documents must report that the navigator was given the focus")
	c.True(navigatorTableFocused, "restoring with no documents must hand the focus to the navigator's table")
	c.Equal(0, open, "restoring a dock state without files must open nothing")
}

// TestClosingTheLastTabFocusesTheNavigator verifies that closing the tab that holds the keyboard focus hands the focus
// to a neighboring tab while there is one, and to the navigator's table once there is none, rather than leaving the
// window with nothing focused, which left the keyboard with nothing to act on until the mouse put the focus somewhere.
func TestClosingTheLastTabFocusesTheNavigator(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().OpenInWindow, nil) // The sheets must land in the dock to be tabs.

	dir := t.TempDir()
	first := filepath.Join(dir, "first"+gurps.SheetExt)
	second := filepath.Join(dir, "second"+gurps.SheetExt)
	c.NoError(gurps.NewEntity().Save(first))
	c.NoError(gurps.NewEntity().Save(second))

	var firstSheet, secondSheet FileBackedDockable
	var focusedPath string
	screen.Do(func() {
		OpenFiles([]string{first, second})
		firstSheet = LocateFileBackedDockable(first)
		secondSheet = LocateFileBackedDockable(second)
		ActivateDockable(firstSheet)
		focusedPath = focusedFilePath(wnd)
	})
	if xreflect.IsNil(firstSheet) || xreflect.IsNil(secondSheet) {
		t.Fatal("the sheets were not opened")
	}
	c.Equal(first, focusedPath, "the first sheet starts out with the focus")

	// With another tab beside it, the focus passes to that tab.
	uxtest.CloseEditorWithoutPrompt(t, screen, firstSheet)
	screen.Do(func() { focusedPath = focusedFilePath(wnd) })
	c.Equal(second, focusedPath, "closing the focused tab hands the focus to the remaining tab")

	// With no tab left, the focus goes to the navigator's table, as at startup.
	var navigatorTableFocused bool
	uxtest.CloseEditorWithoutPrompt(t, screen, secondSheet)
	screen.Do(func() { navigatorTableFocused = wnd.CurrentFocus() == Workspace.Navigator.table.AsPanel() })
	c.True(navigatorTableFocused, "closing the last tab hands the focus to the navigator's table")
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

// TestWidestTooltipWidth verifies that the widest tooltip anywhere in a panel's subtree is what sets the width, and
// that a subtree with no tooltips asks for nothing.
func TestWidestTooltipWidth(t *testing.T) {
	c := check.New(t)
	root := unison.NewPanel()
	c.Equal(float32(0), widestTooltipWidth(root), "no tooltips, no width")

	child := unison.NewPanel()
	child.Tooltip = newWrappedTooltip("short")
	root.AddChild(child)
	_, short, _ := child.Tooltip.Sizes(geom.Size{})
	c.Equal(short.Width, widestTooltipWidth(root), "a child's tooltip counts")

	grandchild := unison.NewPanel()
	grandchild.Tooltip = newWrappedTooltip("a considerably longer tooltip than the short one")
	child.AddChild(grandchild)
	_, long, _ := grandchild.Tooltip.Sizes(geom.Size{})
	c.True(long.Width > short.Width, "the fixture's long tooltip must be the wider one")
	c.Equal(long.Width, widestTooltipWidth(root), "the widest tooltip in the subtree wins, however deep it is")
}

// TestEditorWindowsFitTheirTooltips opens the ancestry and name generator editors in windows of their own and checks
// that each window is at least as wide as the widest tooltip the editor holds. Packed around their fields alone, those
// windows were narrower than their tooltips, which a window squeezes into its own width.
func TestEditorWindowsFitTheirTooltips(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().OpenInWindow, []dgroup.Group{dgroup.Settings})

	uxtest.ChooseMenuBarItem(t, screen, wnd, "File", "New", "New Ancestry")
	ancestry := uxtest.SoleEditor[*ancestryEditorDockable](t, screen, isAncestryEditor)
	checkWindowFitsTooltips(t, c, screen, wnd, ancestry)
	uxtest.CloseEditorWithoutPrompt(t, screen, ancestry)

	uxtest.ChooseMenuBarItem(t, screen, wnd, "File", "New", "New Name Generator")
	names := uxtest.SoleEditor[*nameGeneratorEditorDockable](t, screen, isNameGeneratorEditor)
	checkWindowFitsTooltips(t, c, screen, wnd, names)
	uxtest.CloseEditorWithoutPrompt(t, screen, names)
}

// checkWindowFitsTooltips checks that the dockable was given a window of its own, other than the workspace window, and
// that the window's content is at least as wide as the dockable's widest tooltip. It also checks that the dockable
// alone would have packed narrower than that, since otherwise the window's width would prove nothing.
func checkWindowFitsTooltips(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, workspace *unison.Window, d unison.Dockable) {
	t.Helper()
	var editorWnd *unison.Window
	var contentWidth, tooltipWidth, packedWidth float32
	screen.Do(func() {
		editorWnd = d.AsPanel().Window()
		if editorWnd == nil {
			return
		}
		contentWidth = editorWnd.ContentRect().Width
		tooltipWidth = widestTooltipWidth(d.AsPanel())
		_, pref, _ := d.AsPanel().Sizes(geom.Size{})
		packedWidth = pref.Width
	})
	if editorWnd == nil {
		t.Fatalf("%s has no window", d.Title())
	}
	c.True(editorWnd != workspace, "%s opens in a window of its own", d.Title())
	c.True(tooltipWidth > 0, "%s has tooltips", d.Title())
	c.True(packedWidth < tooltipWidth, "%s packs narrower (%v) than its widest tooltip (%v); the check below would prove nothing otherwise",
		d.Title(), packedWidth, tooltipWidth)
	c.True(contentWidth >= tooltipWidth, "%s's window content (%v wide) fits its widest tooltip (%v wide)",
		d.Title(), contentWidth, tooltipWidth)
}
