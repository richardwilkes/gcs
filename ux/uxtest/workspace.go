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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// Workspace is what a test package tells uxtest about the GCS workspace, which uxtest cannot reach for itself (see the
// package comment). ux's own tests fill it in from ux's unexported pieces; a package built on ux fills it in from what
// ux exports.
type Workspace struct {
	// Setup stands the workspace up in the window StartHeadlessWorkspace has just created, as ux.Start does: it
	// registers the file types and the window's drag types, sets up the menu bar and initializes the workspace, and
	// points the workspace's error handler at the test, since the modal dialog it would otherwise put up has nobody
	// to dismiss it. It runs on the UI thread, before the session is handed to the test, and should put back any
	// process-wide state it replaces when the test ends (see SwapForTest).
	Setup func(t *testing.T, wnd *unison.Window)
	// AllDockables lists every open dockable, whether in the workspace or in a window of its own.
	AllDockables func() []unison.Dockable
}

// workspace is the Workspace Main was given.
var workspace Workspace

// ciScriptExecTimeLimit is the per-script execution time limit, in seconds, the tests run with under CI. It matches
// the one in model/gurps/main_test.go, which explains the choice.
var ciScriptExecTimeLimit = fxp.FromInteger(30)

// Main is what the TestMain of every package using this one runs, with the Workspace the package's tests drive. It
// raises the per-script execution time limit for the duration of the tests, exactly as model/gurps/main_test.go does
// and for the same reason: the sheets and templates the tests load resolve scripts as they are recalculated, and the
// production default is small enough that some CI runners cannot always finish even a trivial script within it.
//
// It also pins the platform-neutral modifier convention a headless session uses on every host, whose menu command key
// is Control rather than macOS's Command. ux's registerActions bakes mod.OSMenuCommand() into the key bindings once
// per process, in whichever test gets there first; under the host's convention, a plain test doing so on macOS would
// bind every menu shortcut to Command, and the headless tests that follow, which press Control, would never reach the
// menu items. Pinning it keeps the bindings independent of test order and makes the tests that never start a session
// behave the same on every host.
func Main(m *testing.M, ws Workspace) {
	workspace = ws
	limit := gurps.PermittedScriptExecTimeMax
	if os.Getenv("CI") != "" {
		limit = ciScriptExecTimeLimit
	}
	gurps.SetScriptExecTimeLimitForTesting(limit)
	mod.SetPlatformNeutral(true)
	os.Exit(m.Run())
}

// StartHeadlessWorkspace starts a headless unison session running the GCS workspace -- menu bar, navigator and document
// dock in one window -- as ux.Start does, minus the update checks and the handoff service. It returns the screen
// driving the session and the workspace window. The session is stopped when the test ends and the process-wide state
// the workspace touches is put back: the settings path, the libraries (see UseTestLibraries), the recent files and
// last-used directories and the workspace-restoration setting, which is turned off so that the dock does not try to
// restore whatever the settings hold. The workspace itself is stood up, and put back, by the Workspace.Setup that Main
// was given.
//
// Stop clears every window's close callbacks and stops every modal loop before it quits, so a dockable left open with
// unsaved changes cannot hang the shutdown with a save prompt, and the workspace's own close handler, which saves the
// global settings, never runs. Even so, a test should leave the dockables it opened closed or unmodified.
//
// Anything the session recorded through Errors() fails the test when it ends. Sessions run one at a time and own most
// of unison's mutable globals while they run, so a test using this must not call t.Parallel.
func StartHeadlessWorkspace(t *testing.T, c check.Checker) (*unison.HeadlessScreen, *unison.Window) {
	t.Helper()
	if workspace.Setup == nil || workspace.AllDockables == nil {
		t.Fatal("the package's TestMain must run uxtest.Main with its Workspace filled in")
	}
	SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	SwapForTest(t, &gurps.GlobalSettings().General.RestoreWorkspaceOnStart, false)
	UseTestLibraries(t, c)
	preserveRecentFilesAndLastDirs(t)

	var wnd *unison.Window
	screen, err := unison.StartHeadless(unison.HeadlessConfig{Width: 1400, Height: 900},
		unison.StartupFinishedCallback(func() {
			w, wndErr := unison.NewWindow("GCS")
			if wndErr != nil {
				t.Errorf("unable to create the workspace window: %v", wndErr)
				return
			}
			workspace.Setup(t, w)
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
// original set when the test finishes. It returns the master and user libraries, both of which start out with no
// "Output Templates" directory at all.
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
