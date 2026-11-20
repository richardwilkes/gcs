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
	"strings"
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

// TestClosingTheWorkspaceSettlesDocumentsNeverSaved verifies that closing the workspace asks whether to save a modified
// document that has never been saved, and stays open with the document when that is canceled. The documents never
// saved are closed rather than recorded in the dock state, since there is no file to reopen them from.
func TestClosingTheWorkspaceSettlesDocumentsNeverSaved(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.TopDockState, nil)
	uxtest.SwapForTest(t, &global.DocDockState, nil)
	uxtest.SwapForTest(t, &global.OpenInWindow, nil) // The documents must land in the dock for the dock state to see them.

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var modified, allowed bool
	screen.Do(func() {
		sheet.Entity().Profile.Name = "Changed"
		modified = sheet.Modified()
	})
	c.True(modified, "the sheet must have unsaved changes for the workspace to ask about")
	returned := false
	c.True(screen.Post(func() {
		allowed = isWorkspaceAllowedToClose()
		returned = true
	}))
	screen.Sync()
	_, dialog := modalDialog(t, screen, wnd)
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	var open int
	var recorded bool
	screen.Do(func() {
		open = len(AllDockables())
		recorded = global.DocDockState != nil
	})
	c.True(returned, "the close has been decided")
	c.False(allowed, "canceling the save must keep the workspace open")
	c.Equal(1, open, "canceling the save must leave the sheet open")
	c.False(recorded, "a canceled close must record nothing")

	screen.Do(sheet.markUnmodified)
	uxtest.OpenedByAction(t, screen, newTraitsLibraryAction)
	var keys []string
	screen.Do(func() {
		allowed = isWorkspaceAllowedToClose()
		open = len(AllDockables())
		keys = dockStateKeys(global.DocDockState)
	})
	c.True(allowed, "the workspace must be allowed to close")
	c.Equal(0, open, "closing the workspace must close the documents never saved")
	for _, key := range keys {
		c.False(strings.HasPrefix(key, filePrefix), "a document never saved must not be recorded: %s", key)
	}
}

// TestClosingAModifiedDocumentFollowsTheSavePrompt verifies that closing a document with unsaved changes asks whether
// to save them: saving writes them to the document's file and closes it, while canceling leaves it open with its
// changes unwritten.
func TestClosingAModifiedDocumentFollowsTheSavePrompt(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().OpenInWindow, nil)
	for _, one := range []struct {
		name     string
		response int
		closed   bool
	}{
		{name: "save", response: unison.ModalResponseOK, closed: true},
		{name: "cancel", response: unison.ModalResponseCancel},
	} {
		path := filepath.Join(t.TempDir(), "sheet"+gurps.SheetExt)
		c.NoError(gurps.NewEntity().Save(path))
		var sheet *Sheet
		var modified bool
		screen.Do(func() {
			OpenFiles([]string{path})
			if s, ok := LocateFileBackedDockable(path).(*Sheet); ok {
				sheet = s
				s.Entity().Profile.Name = "Changed"
				modified = s.Modified()
			}
		})
		if sheet == nil {
			t.Fatalf("%s: the sheet was not opened", one.name)
		}
		c.True(modified, "%s: the sheet must have unsaved changes for closing it to ask about", one.name)
		var closed, returned bool
		c.True(screen.Post(func() {
			closed = sheet.AttemptClose()
			returned = true
		}))
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, one.response)))
		var open bool
		screen.Do(func() {
			open = !xreflect.IsNil(LocateFileBackedDockable(path))
			modified = sheet.Modified()
		})
		c.True(returned, "%s: the close has been decided", one.name)
		c.Equal(one.closed, closed, "%s: whether the close went ahead", one.name)
		c.Equal(!one.closed, open, "%s: whether the sheet is still open", one.name)
		c.Equal(one.response != unison.ModalResponseOK, modified, "%s: whether the changes are still unsaved", one.name)
		saved := loadSavedFile(t, c, path, gurps.NewEntityFromFile).Profile.Name
		c.Equal(one.response == unison.ModalResponseOK, saved == "Changed", "%s: whether the changes were written",
			one.name)
		if open {
			screen.Do(sheet.markUnmodified)
			uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
		}
	}
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

// TestBringingADockedDockableForwardShowsIt verifies that placing a dockable in the workspace's dock, by PlaceInDock or
// DisplayNewDockable, or activating one already there, raises the workspace window over a window of its own that was in
// front of it, as a sheet's is when its Help button opens the User Guide, rather than leaving the dockable to open or
// come forward behind that window. While another container is maximized, in the document dock or in the dock that
// holds it, the dock is restored, rather than the dockable being left hidden behind that container, but a dockable in
// the maximized container leaves it maximized. The document dock is brought in front of a tab its own is stacked with.
func TestBringingADockedDockableForwardShowsIt(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().OpenInWindow, nil) // The dockables must land in the dock.

	var other *unison.Window
	screen.Do(func() {
		var err error
		if other, err = unison.NewWindow("Other"); err != nil {
			t.Errorf("unable to create the other window: %v", err)
		}
	})
	if other == nil {
		t.FailNow()
	}
	// otherInFront brings the other window to the front and reports whether it has the keyboard focus.
	otherInFront := func() (inFront bool) {
		screen.Do(other.ToFront)
		screen.Sync()
		screen.Do(func() { inFront = unison.ActiveWindow() == other })
		return inFront
	}
	// workspaceInFront reports whether the workspace window has the keyboard focus once bringForward has been called.
	workspaceInFront := func(bringForward func()) (inFront bool) {
		screen.Do(bringForward)
		screen.Sync()
		screen.Do(func() { inFront = unison.ActiveWindow() == wnd })
		return inFront
	}

	var d *MarkdownDockable
	screen.Do(func() { d = newMarkdownDockable(markdownContentOnlyPrefix+"Docked", "# Docked\n", false, false) })
	c.True(otherInFront(), "the other window must start out in front")
	c.True(workspaceInFront(func() { PlaceInDock(d, dgroup.Markdown, false) }),
		"placing a dockable in the dock must raise the workspace window")
	c.True(otherInFront(), "the other window must be back in front")
	c.True(workspaceInFront(func() { ActivateDockable(d) }), "activating a docked dockable must raise the workspace window")
	var displayed *MarkdownDockable
	screen.Do(func() {
		displayed = newMarkdownDockable(markdownContentOnlyPrefix+"Displayed", "# Displayed\n", false, false)
	})
	c.True(otherInFront(), "the other window must be back in front again")
	c.True(workspaceInFront(func() { DisplayNewDockable(displayed) }),
		"displaying a new dockable in the dock must raise the workspace window")
	screen.Do(other.Dispose)
	uxtest.CloseEditorWithoutPrompt(t, screen, displayed)

	// shownAfter reports whether target is in front of its container, and neither that container nor any panel
	// holding it is hidden or left without room, once bringForward has been called.
	shownAfter := func(target unison.Dockable, bringForward func()) (shown bool) {
		screen.Do(bringForward)
		screen.Sync()
		screen.Do(func() {
			dc := unison.Ancestor[*unison.DockContainer](target)
			shown = dc != nil && dc.CurrentDockable() == target && target.AsPanel().FrameRect().Height > 0
			for p := target.AsPanel(); shown && p != nil; p = p.Parent() {
				shown = !p.Hidden
			}
		})
		return shown
	}
	var big, late *MarkdownDockable
	screen.Do(func() {
		big = newMarkdownDockable(markdownContentOnlyPrefix+"Maximized", "# Maximized\n", false, false)
		PlaceInDock(big, dgroup.Editors, false)
	})
	for _, tc := range []struct {
		name      string
		container func() *unison.DockContainer
	}{
		{
			name:      "another document container",
			container: func() *unison.DockContainer { return unison.Ancestor[*unison.DockContainer](big) },
		},
		{
			name:      "the navigator's container",
			container: func() *unison.DockContainer { return unison.Ancestor[*unison.DockContainer](Workspace.Navigator) },
		},
	} {
		maximize := func() {
			dc := tc.container()
			dc.Dock.Maximize(dc)
		}
		c.True(shownAfter(d, func() {
			maximize()
			ActivateDockable(d)
		}), "%s maximized: activating a docked dockable must show it", tc.name)
		screen.Do(func() { late = newMarkdownDockable(markdownContentOnlyPrefix+"Late", "# Late\n", false, false) })
		c.True(shownAfter(late, func() {
			maximize()
			PlaceInDock(late, dgroup.Settings, false)
		}), "%s maximized: placing a dockable in a container of its own must show it", tc.name)
		uxtest.CloseEditorWithoutPrompt(t, screen, late)
		screen.Do(func() { late = newMarkdownDockable(markdownContentOnlyPrefix+"Late", "# Late\n", false, false) })
		c.True(shownAfter(late, func() {
			maximize()
			DisplayNewDockable(late)
		}), "%s maximized: displaying a new dockable must show it", tc.name)
		uxtest.CloseEditorWithoutPrompt(t, screen, late)
	}
	c.True(shownAfter(d, func() {
		dc := unison.Ancestor[*unison.DockContainer](Workspace.Navigator)
		dc.Stack(Workspace.DocumentDock, -1)
		dc.SetCurrentDockable(Workspace.Navigator)
		ActivateDockable(d)
	}), "the navigator's tab in front of the document dock's: activating a docked dockable must show it")
	var stillMaximized bool
	screen.Do(func() {
		dc := unison.Ancestor[*unison.DockContainer](big)
		dc.Dock.Maximize(dc)
		ActivateDockable(big)
		stillMaximized = dc.Dock.MaximizedContainer == dc
	})
	c.True(stillMaximized, "activating a dockable in the maximized container must leave it maximized")

	uxtest.CloseEditorWithoutPrompt(t, screen, big)
	uxtest.CloseEditorWithoutPrompt(t, screen, d)
}

// TestDockUnDockMovesADockableBetweenTheWorkspaceAndAWindow verifies that the Undock From Workspace command gives the
// docked dockable that has the focus a window of its own and that, from that window, the same command, retitled Dock
// Into Workspace, puts the dockable back in the dock under the group it was placed with and closes the window.
func TestDockUnDockMovesADockableBetweenTheWorkspaceAndAWindow(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().OpenInWindow, nil) // The dockable must start out in the dock.

	type state struct {
		title   string
		enabled bool
		window  *unison.Window
		docked  bool
		group   any
		windows int
	}
	var d *MarkdownDockable
	current := func(do func()) state {
		var s state
		screen.Do(do)
		screen.Sync()
		screen.Do(func() {
			s.enabled = dockUnDockAction.Enabled(nil)
			s.title = dockUnDockAction.Title
			s.window = d.AsPanel().Window()
			s.docked = unison.Ancestor[*unison.DockContainer](d) != nil
			s.group = d.AsPanel().ClientData()[dockGroupClientDataKey]
			s.windows = len(unison.Windows())
		})
		return s
	}

	s := current(func() {
		d = newMarkdownDockable(markdownContentOnlyPrefix+"Moving", "# Moving\n", false, false)
		PlaceInDock(d, dgroup.Markdown, false)
		ActivateDockable(d)
	})
	c.True(s.enabled, "the command must be enabled for a docked dockable that has the focus")
	c.Equal("Undock From Workspace", s.title)
	c.True(s.docked && s.window == wnd, "the dockable must start out in the workspace's dock")

	s = current(func() { dockUnDockAction.Execute(nil) })
	c.False(s.docked, "undocking must take the dockable out of the dock")
	c.True(s.window != nil && s.window != wnd, "undocking must give the dockable a window of its own")
	c.Equal(2, s.windows)

	own := s.window
	s = current(own.ToFront)
	c.True(s.enabled, "the command must be enabled in the dockable's own window")
	c.Equal("Dock Into Workspace", s.title)

	s = current(func() { dockUnDockAction.Execute(nil) })
	c.True(s.docked && s.window == wnd, "docking must put the dockable back in the workspace's dock")
	c.Equal(1, s.windows, "docking must close the window the dockable had")
	c.Equal(any(dgroup.Markdown), s.group, "docking must keep the group the dockable was placed with")
	uxtest.CloseEditorWithoutPrompt(t, screen, d)
}

// TestDockableAwaitingItsWindowIsOpen verifies that a dockable waiting to be given a window of its own, as a PDF's view
// does until its document has loaded or failed to, is among the open dockables until it is closed, and once only when
// it has its window: opening its file again finds it rather than opening a second view.
func TestDockableAwaitingItsWindowIsOpen(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.OpenInWindow, []dgroup.Group{dgroup.PDFs})
	uxtest.SwapForTest(t, &global.General.ExternalPDFCmdLine, "")
	// There is no such file, so each view's wait ends with its failure to load, which spares the test any rendering.
	path := filepath.Join(t.TempDir(), "Missing.pdf")
	occurrences := func(d unison.Dockable) int {
		count := 0
		for _, one := range AllDockables() {
			if one == d {
				count++
			}
		}
		return count
	}
	var next unison.Dockable
	screen.Do(func() {
		first, _ := OpenFile(path, gurps.PageInfo{})
		c.True(first.AsPanel().Window() == nil, "the view waits for its document before taking a window")
		again, wasOpen := OpenFile(path, gurps.PageInfo{})
		c.True(wasOpen && again == first, "opening the file again finds the view awaiting its window")
		closer, ok := first.(unison.TabCloser)
		c.True(ok && closer.AttemptClose(), "the view closes while awaiting its window")
		c.Equal(0, occurrences(first), "a view closed while awaiting its window is no longer open")
		var reopened bool
		next, reopened = OpenFile(path, gurps.PageInfo{})
		c.False(reopened, "opening the file once its view has closed opens a new one")
	})
	view, ok := next.(*PDFDockable)
	if !ok {
		t.Fatal("the file must open as a PDF")
	}
	awaitLibraryFilter(t, screen, "the view's window", func() bool { return view.Window() != nil })
	screen.Do(func() { c.Equal(1, occurrences(view), "a view given its window is open once") })
	uxtest.CloseEditorWithoutPrompt(t, screen, view)
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

// TestPromptForSingleDestinationSkipsTheDialogForOneChoice verifies that promptForSingleDestination and
// promptForDestinations hand a lone choice back without a dialog and report no choice at all as nothing chosen. Neither
// can put up a dialog here, since there is no window.
func TestPromptForSingleDestinationSkipsTheDialogForOneChoice(t *testing.T) {
	c := check.New(t)
	sheet := newApplyModifierTestSheet(t)
	dest, ok := promptForSingleDestination(promptOperation{}, []FileBackedDockable{sheet})
	c.True(ok, "a lone choice must be taken without asking")
	c.Equal(any(sheet), any(dest), "the lone choice must be the one handed back")
	_, ok = promptForSingleDestination[FileBackedDockable](promptOperation{}, nil)
	c.False(ok, "no choice at all must be reported as nothing chosen")
	c.Equal([]FileBackedDockable{sheet}, promptForDestinations(promptOperation{}, []FileBackedDockable{sheet}),
		"a lone choice must come back as it is")
	c.Equal(0, len(promptForDestinations[FileBackedDockable](promptOperation{}, nil)), "no choice at all must come back empty")
}
