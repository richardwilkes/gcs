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
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestPromptForFileSystemNameGatesOKOnTarget drives the name prompt the rename and new-folder commands share inside a
// headless workspace: OK is disabled while the entry holds an invalid name or one whose target already exists and
// enabled once the target is free, the path handed back is the target of the trimmed name, so that the path validated
// is the one the caller acts on, a current name is shown in a disabled row of its own, and canceling returns nothing.
func TestPromptForFileSystemNameGatesOKOnTarget(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "taken"), 0o750))
	targetPath := func(name string) string { return newFolderPath(dir, name) }

	// Rename "taken" to something else. The dialog runs a nested modal loop, so the call is posted rather than run
	// through Do and its results are read once the dialog has been dismissed.
	var path string
	var ok, returned bool
	c.True(screen.Post(func() {
		path, ok = promptForFileSystemName("taken", "New Name", "taken", targetPath)
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var fields []*StringField
	var okButton *unison.Button
	screen.Do(func() {
		fields = panelsOfType[*StringField](dialogWnd.Content())
		okButton = dialog.Button(unison.ModalResponseOK)
	})
	if len(fields) != 2 {
		t.Fatalf("expected a current name field and an entry, found %d fields", len(fields))
	}
	current, entry := fields[0], fields[1]
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}
	screen.Do(func() {
		c.False(current.Enabled(), "the current name is shown, not edited")
		c.Equal("taken", current.Text())
		c.Equal("taken", entry.Text(), "the entry starts out holding the current name")
	})
	c.False(okEnabled(), "the current name's target exists, so OK starts out disabled")

	screen.Click(screen.PanelCenter(entry))
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("  ")
	c.False(okEnabled(), "a blank name is invalid")
	screen.Type("fresh  ")
	c.True(okEnabled(), "a valid name whose target is free enables OK")
	screen.Click(screen.PanelCenter(okButton))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the prompt has returned")
	c.True(ok, "an accepted name is reported as accepted")
	c.Equal(filepath.Join(dir, "fresh"), path, "the path handed back is the trimmed name's target")

	// A prompt with no current name has only the entry, and canceling it hands back nothing.
	returned = false
	c.True(screen.Post(func() {
		path, ok = promptForFileSystemName("", "Folder Name", "", targetPath)
		returned = true
	}))
	screen.Sync()
	dialogWnd, dialog = modalDialog(t, screen, wnd)
	screen.Do(func() {
		fields = panelsOfType[*StringField](dialogWnd.Content())
		okButton = dialog.Button(unison.ModalResponseOK)
	})
	c.Equal(1, len(fields), "with no current name, only the entry is shown")
	c.False(okEnabled(), "an empty entry leaves OK disabled")
	screen.Type("new folder")
	c.True(okEnabled())
	screen.KeyPress(unison.KeyEscape, mod.None)
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(returned, "the prompt has returned")
	c.False(ok, "a canceled prompt is reported as such")
	c.Equal("", path)
	_, err := os.Stat(filepath.Join(dir, "new folder"))
	c.True(os.IsNotExist(err), "the prompt itself creates nothing")
}

// TestReloadOfADiscardedNavigatorLeavesTheLiveOneAlone verifies that a navigator left behind by a workspace that has
// since been torn down (each headless test starts its own) does nothing when the reload it had scheduled finally runs
// during a later workspace: it must not close the live navigator's library rows, whose disclosure keys it shares, and
// it must have let go of its library watches.
func TestReloadOfADiscardedNavigatorLeavesTheLiveOneAlone(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	var discarded *Navigator
	t.Run("earlier workspace", func(t *testing.T) {
		screen, _ := startHeadlessWorkspace(t, check.New(t))
		screen.Do(func() { discarded = Workspace.Navigator })
	})
	if discarded == nil {
		t.Fatal("the earlier workspace must have had a navigator")
	}
	screen, _ := startHeadlessWorkspace(t, c)
	var discardedInWindow bool
	var watches int
	var closedLibraries []string
	screen.Do(func() {
		discardedInWindow = discarded.Window() != nil
		discarded.Reload()
		watches = len(discarded.tokens)
		for _, row := range Workspace.Navigator.table.RootRows() {
			if row.IsLibrary() && !row.IsOpen() {
				closedLibraries = append(closedLibraries, row.Path())
			}
		}
	})
	c.False(discardedInWindow, "the navigator of a torn-down workspace must no longer be in a window")
	c.Equal(0, watches, "the discarded navigator must have stopped watching the libraries")
	c.Equal(0, len(closedLibraries), "the live navigator's library rows must stay open; closed: %v", closedLibraries)
}

// stopNavigatorWatches stops the navigator's library watches and returns once nothing is left that would bring them
// back. A change the watches have already reported is still on its way to a reload, as is the reload a new navigator
// asks for of itself, and a reload watches every library afresh, so any that is pending is waited out first and the
// watches are stopped again behind it. Until the navigator is next reloaded, changes on disk go unnoticed: its rows go
// stale rather than being rebuilt, and a library folder that is removed is not put back.
func stopNavigatorWatches(t *testing.T, screen *unison.HeadlessScreen, n *Navigator) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var stopped bool
		screen.Do(func() {
			if n.needReload {
				return
			}
			stopped = len(n.tokens) == 0
			for _, token := range n.tokens {
				token.Stop()
			}
			n.tokens = nil
		})
		if stopped {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the navigator never stopped reloading")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestReloadServesInPlaceOfAPendingOne verifies that a reload made while one asked for with EventuallyReload is still
// pending serves in its place: once the delay is up, the navigator is not reloaded a second time, which would replace
// its rows and watch every library afresh behind the back of whoever made the reload, while a request made after the
// reload is still honored.
func TestReloadServesInPlaceOfAPendingOne(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// Waits out the reload a new navigator asks for of itself.
	stopNavigatorWatches(t, screen, n)

	// rowsAfterDelay returns the navigator's root rows once any reload that was pending has had time to be made.
	rowsAfterDelay := func() []*NavigatorNode {
		time.Sleep(2 * eventualReloadDelay)
		var rows []*NavigatorNode
		screen.Do(func() { rows = slices.Clone(n.table.RootRows()) })
		return rows
	}
	screen.Do(n.EventuallyReload)
	var pending bool
	var rows []*NavigatorNode
	screen.Do(func() {
		pending = n.needReload
		n.Reload()
		rows = slices.Clone(n.table.RootRows())
	})
	c.True(pending, "precondition: a reload is pending")
	c.True(len(rows) != 0, "precondition: the navigator has rows")
	c.True(slices.Equal(rows, rowsAfterDelay()), "a reload made in the meantime leaves nothing for the pending one "+
		"to do")

	screen.Do(n.EventuallyReload)
	c.False(slices.Equal(rows, rowsAfterDelay()), "a reload asked for after that is still made")
}

// TestNewFolderIsUnavailableForFavorites verifies that New Folder is disabled while the Favorites row is the selection
// and does nothing should it be asked for anyway. That row has no path of its own, so a folder made "inside" it would
// land in the process's working directory.
func TestNewFolderIsUnavailableForFavorites(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	var n *Navigator
	var favoritesSelected, enabledForLibrary, enabledForFavorites bool
	screen.Do(func() {
		n = Workspace.Navigator
		n.ApplySelectedPaths([]string{gurps.GlobalSettings().Libraries.User().Path(false)})
		enabledForLibrary = n.newFolderButton.Enabled()
		n.table.ClearSelection()
		n.table.SelectByIndex(0)
		sel := n.table.SelectedRows(false)
		favoritesSelected = len(sel) == 1 && sel[0].IsFavorites()
		enabledForFavorites = n.newFolderButton.Enabled()
	})
	c.True(enabledForLibrary, "New Folder must be offered for a library")
	c.True(favoritesSelected, "the Favorites row must be the selection")
	c.False(enabledForFavorites, "New Folder must not be offered for the Favorites row")

	c.True(screen.Post(n.newFolder))
	screen.Sync()
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "asking for a new folder in the Favorites row must not prompt for its name")
}

// TestNewFolderCreatesWhatIsMissingAboveIt drives New Folder for rows whose place on disk has gone: a library whose
// folder is missing, and a folder and a file left showing after the library holding them was deleted. Each time, the
// folder is made along with everything missing above it, the row it was made in is disclosed and the new folder ends
// up selected. The navigator's watches are stopped before the library is deleted, since a navigator that hears of the
// deletion reloads, which drops the stale rows and, by watching the library afresh, puts its folder back.
func TestNewFolderCreatesWhatIsMissingAboveIt(t *testing.T) {
	c := check.New(t)
	// Every row starts out disclosed, so that what is put in the user library below gets rows of its own.
	swapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, wnd := startHeadlessWorkspace(t, c)
	libPath := gurps.GlobalSettings().Libraries.User().Path(false)
	staleDir := filepath.Join(libPath, "stale")
	staleFile := filepath.Join(libPath, "stale"+gurps.SheetExt)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })

	for _, one := range []struct {
		name      string
		rowPath   string
		parentDir string
	}{
		{name: "library", rowPath: libPath, parentDir: libPath},
		{name: "folder", rowPath: staleDir, parentDir: staleDir},
		{name: "file", rowPath: staleFile, parentDir: libPath},
	} {
		c.NoError(os.MkdirAll(staleDir, 0o750), one.name)
		c.NoError(gurps.NewEntity().Save(staleFile), one.name)
		screen.Do(func() {
			n.Reload()
			n.ApplySelectedPaths([]string{one.rowPath})
		})
		stopNavigatorWatches(t, screen, n)
		var selected []string
		screen.Do(func() {
			selected = n.SelectedPaths()
			// Closed, so that the new folder can only end up selected if its parent's row gets disclosed.
			if row := n.table.SelectedRows(false); len(row) == 1 && row[0].Container() {
				row[0].SetOpen(false)
				n.table.SyncToModel()
			}
		})
		c.Equal([]string{one.rowPath}, selected, "the %s row must be the selection", one.name)
		c.NoError(os.RemoveAll(libPath), one.name)

		// The dialog runs a nested modal loop, so the call is posted rather than run through Do.
		c.True(screen.Post(n.newFolder), one.name)
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Type("  made in " + one.name + "  ")
		var okButton *unison.Button
		screen.Do(func() { okButton = dialog.Button(unison.ModalResponseOK) })
		screen.Click(screen.PanelCenter(okButton))

		created := filepath.Join(one.parentDir, "made in "+one.name)
		info, err := os.Stat(created)
		c.NoError(err, one.name)
		c.True(err == nil && info.IsDir(), "a folder must have been made in the %s row's place on disk", one.name)
		var windows int
		screen.Do(func() {
			windows = len(unison.Windows())
			selected = n.SelectedPaths()
		})
		c.Equal(1, windows, "the prompt for the %s row has been dismissed", one.name)
		c.Equal([]string{created}, selected, "the folder made in the %s row must end up selected", one.name)
	}
}
