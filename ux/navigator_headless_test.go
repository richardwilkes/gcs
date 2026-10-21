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
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
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
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
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
// which shows where it matched.
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
	screen.Do(func() {
		n.deepSearch[gurps.SheetExt] = true
		n.Reload()
		for _, row := range n.table.RootRows() {
			if row.IsLibrary() && row.library.IsMaster() {
				libraryTitle = row.primaryColumnText()
			}
		}
	})
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
	var disclosed []string
	screen.Do(func() { disclosed = n.DisclosedPaths() })
	c.True(slices.Contains(disclosed, folder), "the folder holding the file was opened")
	c.True(fieldFocused(), "the focus stays in the field")

	// A reload of the tree while the matches show, as a change in a library brings about, rebuilds the tree from fresh
	// nodes; the matches are found again, with the highlight kept on the same row.
	screen.KeyPress(unison.KeyDown, mod.None)
	c.NotNil(offered(), "Down with the matches hidden offers them again")
	c.Equal(0, highlighted(), "with the first highlighted")
	screen.Do(func() { n.Reload() })
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
	screen.Do(func() { n.handleSelectionDoubleClick() })
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
