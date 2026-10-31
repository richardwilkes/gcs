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
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestPromptForFileSystemNameGatesOKOnTarget drives the name prompt the rename and new-folder commands share inside a
// headless workspace: OK is disabled while the entry holds an invalid name or one whose target already exists and
// enabled once the target is free, the path handed back is the target of the trimmed name, so that the path validated
// is the one the caller acts on, a current name is shown in a disabled row of its own, and canceling returns nothing.
func TestPromptForFileSystemNameGatesOKOnTarget(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
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
		fields = uxtest.PanelsOfType[*StringField](dialogWnd.Content())
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
		fields = uxtest.PanelsOfType[*StringField](dialogWnd.Content())
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
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	var discarded *Navigator
	t.Run("earlier workspace", func(t *testing.T) {
		screen, _ := uxtest.StartHeadlessWorkspace(t, check.New(t))
		screen.Do(func() { discarded = Workspace.Navigator })
	})
	if discarded == nil {
		t.Fatal("the earlier workspace must have had a navigator")
	}
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
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
// asks for of itself, and a reload watches every library afresh, so any that is pending, or whose walk of the
// libraries is still in flight, is waited out first and the watches are stopped again behind it. Until the navigator
// is next reloaded, changes on disk go unnoticed: its rows go stale rather than being rebuilt, and a library folder
// that is removed is not put back.
func stopNavigatorWatches(t *testing.T, screen *unison.HeadlessScreen, n *Navigator) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var stopped bool
		screen.Do(func() {
			if n.needReload || n.walking {
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

// awaitContentCache returns once the navigator's deep search content cache holds an entry for every deep-searchable
// file in the tree and nothing is in flight that would change it: a pending reload is waited out, as are the
// background build a reload starts and the read a search starts, and an entry a filesystem watch dropped with no
// rebuild behind it is rebuilt here, as a reload would. The searches a test makes afterwards then find every file's
// contents in the cache rather than leaving the file out while it is read in the background.
func awaitContentCache(t *testing.T, screen *unison.HeadlessScreen, n *Navigator) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var complete bool
		screen.Do(func() {
			if n.needReload || n.walking || n.lastBuild != nil || n.contentLoad != nil {
				return
			}
			paths := make(map[string]bool)
			n.collectDeepSearchPaths(n.table.RootRows(), paths)
			for p := range paths {
				if n.contentCache[p] == nil {
					n.prewarmContentCache()
					return
				}
			}
			complete = true
		})
		if complete {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the content cache never finished building")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// awaitNavigatorWalk returns once no walk of the libraries is in flight and none is owed, so that the rows are those
// the last reload built; see Navigator.Reload.
func awaitNavigatorWalk(t *testing.T, screen *unison.HeadlessScreen, n *Navigator) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		var done bool
		screen.Do(func() { done = !n.walking && !n.reloadAgain })
		if done {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the navigator never finished walking the libraries")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// reloadNavigator reloads the navigator and returns once the rows are rebuilt from the walk of the libraries that
// starts, so that the test can read them.
func reloadNavigator(t *testing.T, screen *unison.HeadlessScreen, n *Navigator) {
	t.Helper()
	screen.Do(n.Reload)
	awaitNavigatorWalk(t, screen, n)
}

// TestReloadServesInPlaceOfAPendingOne verifies that a reload made while one asked for with EventuallyReload is still
// pending serves in its place: once the delay is up, the navigator is not reloaded a second time, which would replace
// its rows and watch every library afresh behind the back of whoever made the reload, while a request made after the
// reload is still honored.
func TestReloadServesInPlaceOfAPendingOne(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// Waits out the reload a new navigator asks for of itself.
	stopNavigatorWatches(t, screen, n)

	// rowsNow returns the navigator's root rows.
	rowsNow := func() []*NavigatorNode {
		var rows []*NavigatorNode
		screen.Do(func() { rows = slices.Clone(n.table.RootRows()) })
		return rows
	}
	// rowsAfterDelay returns the navigator's root rows once any reload that was pending has had time to be made and
	// the walk it starts has returned.
	rowsAfterDelay := func() []*NavigatorNode {
		time.Sleep(2 * eventualReloadDelay)
		awaitNavigatorWalk(t, screen, n)
		return rowsNow()
	}
	screen.Do(n.EventuallyReload)
	var pending bool
	screen.Do(func() {
		pending = n.needReload
		n.Reload()
	})
	c.True(pending, "precondition: a reload is pending")
	awaitNavigatorWalk(t, screen, n)
	rows := rowsNow()
	c.True(len(rows) != 0, "precondition: the navigator has rows")
	c.True(slices.Equal(rows, rowsAfterDelay()), "a reload made in the meantime leaves nothing for the pending one "+
		"to do")

	screen.Do(n.EventuallyReload)
	c.False(slices.Equal(rows, rowsAfterDelay()), "a reload asked for after that is still made")
}

// TestReloadWalksTheLibrariesOffTheUIThread verifies that a reload watches and lists the libraries on a background
// goroutine and leaves the rows as they are until it returns, so that the UI stays responsive while a library on a slow
// disk or a network share is watched and listed; that a reload asked for while a walk is in flight is made once that
// walk's rows are in, rather than being dropped or canceling it; and that what is to follow a reload (see
// reloadAndSelect) runs after the walk that starts once it was asked for, since an earlier walk may have listed the
// folder it concerns before the change it waits on was made, and that a reload it asks for while one is owed starts no
// walk beside the owed one. The walks are held back by the test, which is what lets it look in on the UI thread while
// one runs.
func TestReloadWalksTheLibrariesOffTheUIThread(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	stopNavigatorWatches(t, screen, n)

	release := make(chan struct{})
	t.Cleanup(func() { close(release) }) // Lets a walk still held back return, rather than leaving its goroutine behind
	var walks atomic.Int32
	screen.Do(func() {
		n.listLibraries = func(libs []*library.Library, groupContainers bool, symlinkedFolder func(*library.Library, string)) *navigatorListing {
			walks.Add(1)
			<-release
			return listLibraries(libs, groupContainers, symlinkedFolder)
		}
	})
	// releaseWalk lets the walk that is held back return.
	releaseWalk := func() {
		select {
		case release <- struct{}{}:
		case <-time.After(10 * time.Second):
			t.Fatal("no walk is waiting to be released")
		}
	}
	// awaitWalks returns once that many walks have started.
	awaitWalks := func(count int32) {
		deadline := time.Now().Add(10 * time.Second)
		for walks.Load() < count {
			screen.Do(func() {}) // Runs the tasks a walk that returned queued
			if time.Now().After(deadline) {
				t.Fatalf("expected %d walks to have started, got %d", count, walks.Load())
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	rowsNow := func() []*NavigatorNode {
		var rows []*NavigatorNode
		screen.Do(func() { rows = slices.Clone(n.table.RootRows()) })
		return rows
	}
	var followed []string
	followedNow := func() []string {
		var result []string
		screen.Do(func() { result = slices.Clone(followed) })
		return result
	}

	before := rowsNow()
	var walking, again bool
	var watchedOnUIThread int
	screen.Do(func() {
		n.afterReload = append(n.afterReload, func() {
			followed = append(followed, "first")
			n.Reload()
		})
		n.Reload()
		walking = n.walking
		watchedOnUIThread = len(n.tokens)
		n.afterReload = append(n.afterReload, func() { followed = append(followed, "second") })
		n.Reload()
		again = n.reloadAgain
	})
	c.True(walking, "a reload must start a walk and return")
	c.Equal(0, watchedOnUIThread, "the libraries must be watched afresh by the walk, not on the UI thread")
	c.True(again, "a reload asked for while a walk is in flight must be owed")
	awaitWalks(1)
	c.True(slices.Equal(before, rowsNow()), "the rows must stay as they are until the walk returns")
	c.Equal(int32(1), walks.Load(), "an owed reload must not start a walk beside the one in flight")
	c.Equal(0, len(followedNow()), "nothing follows a reload before its walk returns")

	releaseWalk()
	awaitWalks(2)
	c.False(slices.Equal(before, rowsNow()), "the rows must be rebuilt once the walk returns")
	c.Equal([]string{"first"}, followedNow(), "what was to follow the first reload must run once its walk returns, "+
		"while what was asked for during that walk waits for the next")
	var stillWalking, stillOwed bool
	screen.Do(func() {
		stillWalking = n.walking
		stillOwed = n.reloadAgain
	})
	c.True(stillWalking, "the owed reload must have started a walk of its own")
	c.False(stillOwed, "and no longer be owed")

	releaseWalk()
	awaitNavigatorWalk(t, screen, n)
	c.Equal([]string{"first", "second"}, followedNow(), "what was to follow the second reload must run once its "+
		"walk returns")
	c.Equal(int32(2), walks.Load(), "a reload asked for by what follows a walk, while one is owed, must not start a "+
		"walk beside the owed one")
	var watches, libraries int
	screen.Do(func() {
		watches = len(n.tokens)
		libraries = len(gurps.GlobalSettings().Libraries.List())
	})
	c.Equal(libraries, watches, "each library must be watched once the walk returns")
}

// TestNewFolderIsUnavailableForFavorites verifies that New Folder is disabled while the Favorites row is the selection
// and does nothing should it be asked for anyway. That row has no path of its own, so a folder made "inside" it would
// land in the process's working directory.
func TestNewFolderIsUnavailableForFavorites(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
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

// TestEachCopyOfAFolderUnderFavoritesKeepsItsOwnState verifies that a folder shown more than once under Favorites, as a
// favorite and within a favorited folder, has an ID of its own in each place, as does each folder within it, so that
// selecting or disclosing one copy leaves the other alone, across a Reload too, and that each copy keeps its ID, on
// which the settings key its open state, in a tree built afresh, as on the next launch. A folder shown once keeps the
// ID it was given before copies were told apart, so that the open state saved for it is kept.
func TestEachCopyOfAFolderUnderFavoritesKeepsItsOwnState(t *testing.T) {
	c := check.New(t)
	// Keeps the open states set here out of the settings the other tests see.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	c.NoError(os.MkdirAll(filepath.Join(dir, "sub", "inner", "deeper"), 0o750))
	c.NoError(os.MkdirAll(filepath.Join(dir, "sub", "other"), 0o750))
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	lib.ToggleFavorite("sub")
	lib.ToggleFavorite("sub/inner")
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The library joins once the rows the navigator starts with are in, so that the reload below is what adds its rows.
	stopNavigatorWatches(t, screen, n)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })

	// The copies of inner and deeper as the favorite inner shows them and as the favorite sub shows them, and other,
	// which only sub shows.
	type copies struct {
		inner, innerInSub, deeper, deeperInSub, otherInSub *NavigatorNode
	}
	child := func(row *NavigatorNode, p string) *NavigatorNode {
		for _, one := range row.Children() {
			if one.library == lib && one.path == p {
				return one
			}
		}
		return nil
	}
	// find returns the copies beneath the navigator's Favorites row or, if fresh, beneath one built afresh.
	find := func(fresh bool) copies {
		var found copies
		screen.Do(func() {
			favorites := n.table.RootRows()[0]
			if fresh {
				settings := gurps.GlobalSettings()
				favorites = NewFavoritesNode(n, listFavorites(settings.Libraries.List(),
					settings.General.GroupContainersOnSort, ignoreSymlinkedFolder))
			}
			if found.inner = child(favorites, "sub/inner"); found.inner != nil {
				found.deeper = child(found.inner, "sub/inner/deeper")
			}
			if sub := child(favorites, "sub"); sub != nil {
				if found.innerInSub = child(sub, "sub/inner"); found.innerInSub != nil {
					found.deeperInSub = child(found.innerInSub, "sub/inner/deeper")
				}
				found.otherInSub = child(sub, "sub/other")
			}
		})
		if found.deeper == nil || found.deeperInSub == nil || found.otherInSub == nil {
			t.Fatal("expected inner and deeper under both favorites, and other under sub")
		}
		return found
	}
	reloadNavigator(t, screen, n)
	rows := find(false)
	var otherID tid.TID
	screen.Do(func() {
		otherID = gurps.IDForNavNode("F@"+filepath.Join(dir, "sub", "other"), kinds.NavigatorDirectory)
	})
	c.NotEqual(rows.inner.ID(), rows.innerInSub.ID(), "each copy of a favorite within a favorite has its own ID")
	c.NotEqual(rows.deeper.ID(), rows.deeperInSub.ID(), "as has each copy of a folder within it")
	c.Equal(otherID, rows.otherInSub.ID(), "a folder shown once keeps the ID it was given before copies were "+
		"told apart")

	// state returns whether each copy is open, and where the selected rows are.
	state := func(rows copies) (open []bool, selected []string) {
		screen.Do(func() {
			open = []bool{rows.inner.IsOpen(), rows.innerInSub.IsOpen(), rows.deeper.IsOpen(), rows.deeperInSub.IsOpen()}
			for _, row := range n.table.SelectedRows(false) {
				selected = append(selected, row.path+" in "+row.Parent().primaryColumnText())
			}
		})
		return open, selected
	}
	// A Reload closes the rows it adds, so each copy starts out closed.
	screen.Do(func() {
		rows.innerInSub.Parent().SetOpen(true) // So that both copies of inner show
		rows.inner.SetOpen(true)
		rows.deeperInSub.SetOpen(true)
		n.table.SyncToModel()
		n.table.SetSelectionMap(map[tid.TID]bool{rows.inner.ID(): true})
	})
	want := []bool{true, false, false, true}
	open, selected := state(rows)
	c.Equal(want, open, "opening a copy leaves the other closed")
	c.Equal([]string{"sub/inner in Favorites"}, selected, "selecting a copy selects it alone")

	reloadNavigator(t, screen, n)
	open, selected = state(find(false))
	c.Equal(want, open, "each copy keeps its own open state across a Reload")
	c.Equal([]string{"sub/inner in Favorites"}, selected, "and the selection stays on the copy selected")

	fresh := find(true)
	c.Equal([]tid.TID{rows.inner.ID(), rows.innerInSub.ID(), rows.deeper.ID(), rows.deeperInSub.ID()},
		[]tid.TID{fresh.inner.ID(), fresh.innerInSub.ID(), fresh.deeper.ID(), fresh.deeperInSub.ID()},
		"each copy keeps its ID in a tree built afresh")
}

// TestNewFolderSelectsOnlyTheFolderItMade verifies that the folder New Folder makes ends up the only row selected: the
// copy within the row it was asked for in, or beside it for a file, though a folder shown both under its library and
// under Favorites shows the new one in both, or the copy under its library when that place doesn't show it, as
// Favorites doesn't beside a favorite file.
func TestNewFolderSelectsOnlyTheFolderItMade(t *testing.T) {
	c := check.New(t)
	// Keeps the open states set here out of the settings the other tests see.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "sub"), 0o750))
	c.NoError(gurps.NewEntity().Save(filepath.Join(dir, "top"+gurps.SheetExt)))
	lib := library.NewLibrary("Test", "acct", "", "repo", dir)
	lib.ToggleFavorite("sub")
	lib.ToggleFavorite("top" + gurps.SheetExt)
	libraries := gurps.GlobalSettings().Libraries
	libraries.Store(lib.Key(), lib)
	t.Cleanup(func() { libraries.Remove(lib.Key()) })
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })

	// where returns the rows from the root down to the row: Favorites and the library by kind, the others by path.
	where := func(row *NavigatorNode) string {
		var parts []string
		for p := row; p != nil; p = p.parent {
			switch {
			case p.IsFavorites():
				parts = append(parts, "Favorites")
			case p.IsLibrary():
				parts = append(parts, "library")
			default:
				parts = append(parts, p.path)
			}
		}
		slices.Reverse(parts)
		return strings.Join(parts, " > ")
	}
	for _, one := range []struct {
		name string
		row  string // Where the row New Folder is asked for in is
		want string // Where the new folder must be the only row selected
	}{
		{name: "made in folder", row: "Favorites > sub", want: "Favorites > sub > sub/made in folder"},
		{name: "made beside file", row: "Favorites > top" + gurps.SheetExt, want: "library > made beside file"},
	} {
		stopNavigatorWatches(t, screen, n)
		reloadNavigator(t, screen, n)
		var found bool
		screen.Do(func() {
			// Discloses sub under both Favorites and the library, so that a folder made in it shows in both.
			for _, root := range n.table.RootRows() {
				if root.IsFavorites() || root.library == lib {
					root.SetOpen(true)
					for _, row := range root.Children() {
						if row.library == lib && row.Container() {
							row.SetOpen(true)
						}
					}
				}
			}
			n.table.SyncToModel()
			for _, row := range n.table.RootRows()[0].Children() {
				if where(row) == one.row {
					n.table.SetSelectionMap(map[tid.TID]bool{row.ID(): true})
					found = true
				}
			}
		})
		if !found {
			t.Fatalf("expected a row at %s", one.row)
		}

		// The dialog runs a nested modal loop, so the call is posted rather than run through Do.
		c.True(screen.Post(n.newFolder), one.name)
		screen.Sync()
		_, dialog := modalDialog(t, screen, wnd)
		screen.Type(one.name)
		var okButton *unison.Button
		screen.Do(func() { okButton = dialog.Button(unison.ModalResponseOK) })
		screen.Click(screen.PanelCenter(okButton))
		awaitNavigatorWalk(t, screen, n) // The folder is selected once the reload New Folder asks for has its rows in
		var selected []string
		screen.Do(func() {
			for _, row := range n.table.SelectedRows(false) {
				selected = append(selected, where(row))
			}
		})
		c.Equal([]string{one.want}, selected, "the folder %s must end up the only row selected", one.name)
	}
}

// TestNewFolderCreatesWhatIsMissingAboveIt drives New Folder for rows whose place on disk has gone: a library whose
// folder is missing, and a folder and a file left showing after the library holding them was deleted. Each time, the
// folder is made along with everything missing above it, the row it was made in is disclosed and the new folder ends
// up selected. The navigator's watches are stopped before the library is deleted, since a navigator that hears of the
// deletion reloads, which drops the stale rows and, by watching the library afresh, puts its folder back.
func TestNewFolderCreatesWhatIsMissingAboveIt(t *testing.T) {
	c := check.New(t)
	// Every row starts out disclosed, so that what is put in the user library below gets rows of its own.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
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
		reloadNavigator(t, screen, n)
		screen.Do(func() { n.ApplySelectedPaths([]string{one.rowPath}) })
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
		awaitNavigatorWalk(t, screen, n) // The folder is selected once the reload New Folder asks for has its rows in

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

// TestNavigatorToolbarWrapsToFitTheSearchField verifies that the Library Explorer's toolbar keeps its buttons and its
// search field on one row while the navigator is wide enough, giving the field what is left of the row, and otherwise
// wraps, never making the field narrower than its minimum.
func TestNavigatorToolbarWrapsToFitTheSearchField(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// frames returns the frames of the toolbar's first button, its search field and the toolbar's content with the
	// navigator made the given width.
	frames := func(width float32) (first, field, content geom.Rect) {
		screen.Do(func() {
			Workspace.TopDock.RootDockLayout().SetDividerPosition(width)
			wnd.ValidateLayout()
			first = n.toolbar.Children()[0].FrameRect()
			field = n.search.field.FrameRect()
			content = n.toolbar.ContentRect(false)
		})
		return first, field, content
	}
	var fieldMin geom.Size
	screen.Do(func() { fieldMin, _, _ = n.search.field.Sizes(geom.Size{}) })

	first, field, content := frames(800)
	c.True(math.Abs(float64(first.CenterY()-field.CenterY())) < 1, "with room, the field shares the buttons' row")
	c.True(math.Abs(float64(field.Right()-content.Right())) < 1, "and takes what is left of it")
	c.True(field.Width > fieldMin.Width, "which is more than its minimum")

	first, field, content = frames(250)
	c.True(field.Y > first.Bottom(), "without room, the field is on a row below the buttons")
	c.True(math.Abs(float64(field.Right()-content.Right())) < 1, "taking the whole row")
	c.True(field.Width >= fieldMin.Width, "and never less than its minimum")
	c.Equal(content.X, field.X)

	// A character sheet's toolbar wraps the same way as the space beside the navigator shrinks.
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	sheetFrames := func(dividerAt float32) (first, field, content geom.Rect) {
		screen.Do(func() {
			Workspace.TopDock.RootDockLayout().SetDividerPosition(dividerAt)
			wnd.ValidateLayout()
			first = sheet.toolbar.Children()[0].FrameRect()
			field = sheet.search.field.FrameRect()
			content = sheet.toolbar.ContentRect(false)
		})
		return first, field, content
	}
	first, field, content = sheetFrames(250)
	c.True(math.Abs(float64(first.CenterY()-field.CenterY())) < 1, "with room, the sheet's field shares the buttons' row")
	c.True(math.Abs(float64(field.Right()-content.Right())) < 1, "and takes what is left of it")
	first, field, content = sheetFrames(1000)
	c.True(field.Y > first.Bottom(), "without room, the sheet's field is on a row below the buttons")
	c.True(math.Abs(float64(field.Right()-content.Right())) < 1, "taking the whole row")
	uxtest.CaptureScreen(t, c, screen, "sheet_toolbar_wrapped")
	screen.Do(func() { sheet.markUnmodified() })
}

// TestNavigatorSearchOffersMatches drives the Library Explorer's search in a headless workspace: typing offers the
// matching rows in a dropdown, each titled with where it is, choosing one with RETURN discloses its folder and selects
// it alone while the focus stays in the field, a reload of the tree while the matches show finds them again with the
// highlight kept, and a file that matched on its contents opens with the search text handed to its own search field,
// which shows where it matched, even when the file's cache entry was dropped just before.
func TestNavigatorSearchOffersMatches(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// Waits out the reload a new navigator asks for of itself, so that the files written below are picked up by the
	// reload made here and by nothing else.
	stopNavigatorWatches(t, screen, n)
	folder := filepath.Join(gurps.GlobalSettings().Libraries.Master().Path(false), "Folder")
	c.NoError(os.MkdirAll(folder, 0o750))
	notesPath := filepath.Join(folder, "Findable"+gurps.NotesExt)
	c.NoError(os.WriteFile(notesPath, []byte("{}"), 0o600))
	entity := newTestCharacter("Hero")
	entity.Traits = []*gurps.Trait{newFindableTrait(entity, "Barbarian Rage")}
	sheetPath := filepath.Join(folder, "Hero"+gurps.SheetExt)
	c.NoError(entity.Save(sheetPath))
	var libraryTitle string
	screen.Do(func() { n.deepSearch[gurps.SheetExt] = true })
	reloadNavigator(t, screen, n)
	screen.Do(func() {
		for _, row := range n.table.RootRows() {
			if row.IsLibrary() && row.library.IsMaster() {
				libraryTitle = row.primaryColumnText()
			}
		}
	})
	// The writes above reach the watches the reload set up afresh late, dropping the sheet's cache entry and asking for
	// another reload, so the watches are stopped again behind that and the cache then waited for, so that the searches
	// below find the sheet's contents in the cache rather than leaving it out while they are read in the background.
	stopNavigatorWatches(t, screen, n)
	awaitContentCache(t, screen, n)
	offered := func() *unison.Window { return screen.SuggestionsWindow(wnd) }
	rows := func() []*searchMatch {
		var found []*searchMatch
		screen.Do(func() {
			list := n.search.suggestions.List()
			for i := range list.Count() {
				found = append(found, list.DataAtIndex(i))
			}
		})
		return found
	}
	highlighted := func() int {
		index := -2
		screen.Do(func() { index = n.search.suggestions.Selected() })
		return index
	}
	selectedPaths := func() []string {
		var paths []string
		screen.Do(func() { paths = n.SelectedPaths() })
		return paths
	}
	fieldFocused := func() bool {
		var focused bool
		screen.Do(func() { focused = wnd.Focus() == n.search.field.AsPanel() })
		return focused
	}

	c.Nil(offered(), "nothing is on offer before anything is typed")
	screen.Click(screen.PanelCenter(n.search.field))
	screen.Type("find")
	c.NotNil(offered(), "typing offers the matches")
	found := rows()
	if len(found) != 1 {
		t.Fatalf("expected the notes file alone to match, found %d rows", len(found))
	}
	c.Equal("Findable", found[0].name)
	c.Equal(libraryTitle+" / Folder", found[0].title, "the title says where the file is")
	c.Equal(-1, highlighted(), "none is highlighted to start with")
	c.Equal(0, len(selectedPaths()), "typing selects nothing")

	screen.KeyPress(unison.KeyDown, mod.None)
	c.Equal(0, highlighted(), "Down highlights the first row")
	uxtest.CaptureScreen(t, c, screen, "navigator_search")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Nil(offered(), "choosing hides the matches")
	c.Equal([]string{notesPath}, selectedPaths(), "the chosen row alone is selected")
	c.True(fieldFocused(), "the focus stays in the field")

	// A reload of the tree while the matches show, as a change in a library brings about, rebuilds the tree from fresh
	// nodes; the matches are found again, with the highlight kept on the same row.
	screen.KeyPress(unison.KeyDown, mod.None)
	c.NotNil(offered(), "Down with the matches hidden offers them again")
	c.Equal(0, highlighted(), "with the first highlighted")
	reloadNavigator(t, screen, n)
	c.NotNil(offered(), "the matches stay on offer across a reload")
	c.Equal(1, len(rows()))
	c.Equal(0, highlighted(), "with the highlight kept")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{notesPath}, selectedPaths(), "and choosing selects the row in the rebuilt tree")

	// The sheet matches "barbarian" on its contents, not its name. Opening it hands the text to its own search field,
	// which shows the trait that matched.
	screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
	screen.Type("barbarian")
	found = rows()
	if len(found) != 1 {
		t.Fatalf("expected the sheet alone to match on its contents, found %d rows", len(found))
	}
	c.Equal("Hero", found[0].name)
	screen.KeyPress(unison.KeyDown, mod.None)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{sheetPath}, selectedPaths())
	var noted bool
	screen.Do(func() {
		n.invalidateContentCacheEntry(sheetPath) // As a change to the sheet reported by a watch does
		n.handleSelectionDoubleClick()
		noted = n.missingContent[sheetPath]
	})
	c.False(noted, "opening a file must not leave it noted as missing for a search")
	sheet := uxtest.SoleEditor[*Sheet](t, screen, func(d unison.Dockable) bool { _, ok := d.(*Sheet); return ok })
	var text string
	var sheetSearchFocused bool
	var sheetRows []string
	screen.Do(func() {
		text = sheet.search.field.Text()
		sheetSearchFocused = wnd.Focus() == sheet.search.field.AsPanel()
		list := sheet.search.suggestions.List()
		for i := range list.Count() {
			sheetRows = append(sheetRows, list.DataAtIndex(i).name)
		}
	})
	c.Equal("barbarian", text, "the text the file matched on is handed to the sheet")
	c.True(sheetSearchFocused, "and its search field is focused")
	c.NotNil(offered(), "which shows the sheet's own matches")
	c.Equal([]string{"Barbarian Rage"}, sheetRows)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}
