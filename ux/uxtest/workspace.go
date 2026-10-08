// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.
package uxtest

import (
	"image/png"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// StartHeadlessWorkspace starts a headless unison session running the GCS workspace -- menu bar, navigator and document
// dock in one window -- as ux.Start does, minus the update checks and the handoff service. It returns the screen
// driving the session and the workspace window. The session is stopped when the test ends and the process-wide state
// the workspace touches is put back: the ux.Workspace global, the settings path, the libraries (see UseTestLibraries),
// the recent files and last-used directories and the workspace-restoration setting, which is turned off so that the
// dock does not try to restore whatever the settings hold.
//
// Stop clears every window's close callbacks and stops every modal loop before it quits, so a dockable left open with
// unsaved changes cannot hang the shutdown with a save prompt, and the workspace's own close handler, which saves the
// global settings, never runs. Even so, a test should leave the dockables it opened closed or unmodified.
//
// Anything the session recorded through Errors() fails the test when it ends. Sessions run one at a time and own most
// of unison's mutable globals while they run, so a test using this must not call t.Parallel.
func StartHeadlessWorkspace(t *testing.T, c check.Checker) (*unison.HeadlessScreen, *unison.Window) {
	t.Helper()
	SwapForTest(t, &ux.Workspace, ux.Workspace) // The session replaces most of the workspace; put all of it back.
	SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	SwapForTest(t, &gurps.GlobalSettings().General.RestoreWorkspaceOnStart, false)
	UseTestLibraries(t, c)
	preserveRecentFilesAndLastDirs(t)
	// main registers the file types before starting the UI; the navigator looks its folder icons up in that registry.
	ux.RegisterKnownFileTypes()

	var wnd *unison.Window
	screen, err := unison.StartHeadless(unison.HeadlessConfig{Width: 1400, Height: 900},
		unison.StartupFinishedCallback(func() {
			w, wndErr := unison.NewWindow("GCS")
			if wndErr != nil {
				t.Errorf("unable to create the workspace window: %v", wndErr)
				return
			}
			ux.RegisterWindowDragTypes(w)
			ux.SetupMenuBar(w)
			ux.InitWorkspace(w)
			wnd = w
		}))
	if err != nil {
		t.Fatalf("unable to start the headless session: %v", err)
	}
	// Registered ahead of Stop so that it runs after the session has ended, by which time everything it will record
	// has been recorded.
	t.Cleanup(func() {
		for _, one := range screen.Errors() {
			t.Errorf("the headless session recorded an error: %v", one)
		}
	})
	t.Cleanup(screen.Stop)
	if wnd == nil {
		t.Fatal("the workspace window was not created")
	}
	// The workspace reports errors with a modal dialog once it has finished initializing. Nobody would dismiss it, so
	// report them as test failures instead.
	screen.Do(func() {
		ux.Workspace.ErrorHandler = func(msg string, err error) { t.Errorf("unexpected error: %s: %v", msg, err) }
	})
	return screen, wnd
}

// SwapForTest assigns value to the variable target points at for the duration of the test, putting back whatever it
// held when the test finishes. Package-level state and fields of the global settings are process-wide, so a test that
// alters one has to restore it or every test that runs afterwards sees the change.
func SwapForTest[T any](t *testing.T, target *T, value T) {
	t.Helper()
	saved := *target
	t.Cleanup(func() { *target = saved })
	*target = value
}

// UseTestLibraries points the global settings at a fresh library set rooted in a temporary directory, restoring the
// original set when the test finishes. It returns the master and user libraries.
func UseTestLibraries(t *testing.T, c check.Checker) (master, user *library.Library) {
	t.Helper()
	global := gurps.GlobalSettings()
	SwapForTest(t, &global.Libraries, library.NewLibraries())
	dir := t.TempDir()
	master = global.Libraries.Master()
	c.NoError(master.SetPath(filepath.Join(dir, "master")))
	user = global.Libraries.User()
	c.NoError(user.SetPath(filepath.Join(dir, "user")))
	return master, user
}

// preserveRecentFilesAndLastDirs puts the global settings' recent files list and last-used directories back when the
// test ends. Saving or opening a file records the file in the one and its directory in the other, and the global
// settings are process-wide, so a temporary path left in either would be offered to every test that runs afterwards.
func preserveRecentFilesAndLastDirs(t *testing.T) {
	t.Helper()
	global := gurps.GlobalSettings()
	savedRecentFiles := slices.Clone(global.RecentFiles)
	savedLastDirs := maps.Clone(global.LastDirs)
	t.Cleanup(func() {
		global.RecentFiles = savedRecentFiles
		global.LastDirs = savedLastDirs
	})
}

// CaptureScreen writes what the screen shows to name.png in the directory named by GCS_HEADLESS_CAPTURE_DIR, for the
// person running the test to look at. When nothing is named there, nothing is captured.
func CaptureScreen(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, name string) {
	t.Helper()
	dir := os.Getenv("GCS_HEADLESS_CAPTURE_DIR")
	if dir == "" {
		return
	}
	img := screen.Capture()
	if img == nil {
		t.Fatal("the screen could not be captured")
	}
	path := filepath.Join(dir, name+".png")
	f, err := os.Create(path) //nolint:gosec // G703: writing into the directory named in the environment is the point
	if err != nil {
		t.Fatalf("unable to create %s: %v", path, err)
	}
	c.NoError(png.Encode(f, img), "encoding %s", path)
	c.NoError(f.Close(), "closing %s", path)
}

// FocusForReadingSetter returns a function that sets the FocusForReading general setting, passes it to unison and
// describes the window, which also activates accessibility. Both are restored when the test ends.
func FocusForReadingSetter(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window) func(enabled bool) {
	t.Helper()
	gs := gurps.GlobalSettings().General
	SwapForTest(t, &gs.FocusForReading, false)
	// Restored on the UI thread: from elsewhere, with accessibility active, unison queues a task to describe the
	// windows again, and a task still queued at Stop runs in the next test.
	saved := unison.FocusForReading()
	t.Cleanup(func() { screen.Do(func() { unison.SetFocusForReading(saved) }) })
	return func(enabled bool) {
		t.Helper()
		screen.Do(func() {
			gs.FocusForReading = enabled
			gs.UpdateFocusForReading()
		})
		if screen.AccessibilityTree(wnd) == nil {
			t.Fatal("the window must be described")
		}
	}
}
