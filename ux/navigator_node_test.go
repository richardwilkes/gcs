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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// TestNavigatorNodeSortTextIsTheTextShown verifies that a navigator row's sort text, which a screen reader is given as
// the row's name, is the text the row shows with no ordering rank ahead of it.
func TestNavigatorNodeSortTextIsTheTextShown(t *testing.T) {
	c := check.New(t)
	n := newDeepSearchNavigator()
	lib := newTestLibrary(t)
	libNode := NewLibraryNode(n, lib, nil)
	c.Equal("Test", libNode.CellDataForSort(0), "a library row is named by its title alone")
	c.Equal("", libNode.CellDataForSort(1), "the navigator has only one column")
	fileNode := writeTestLibraryFile(c, lib, "Conan.gcs", []byte("{}"))
	c.Equal("Conan", fileNode.CellDataForSort(0), "a file row is named by its file name without the extension")
}

// flattenEntries returns the paths of the entries and, after each folder's, of those beneath it, folders marked with a
// trailing slash.
func flattenEntries(entries []*navigatorEntry) []string {
	var paths []string
	for _, entry := range entries {
		if entry.isDir {
			paths = append(paths, entry.path+"/")
			paths = append(paths, flattenEntries(entry.entries)...)
		} else {
			paths = append(paths, entry.path)
		}
	}
	return paths
}

// ignoreSymlinkedFolder is the symlinkedFolder callback for a listing whose symlinks are of no interest.
func ignoreSymlinkedFolder(_ *library.Library, _ string) {}

// TestListFolderListsWhatTheTreeShows verifies the walk the rows are built from: the entries come in the order the tree
// shows them, folders ahead of files when they are grouped and otherwise mixed in, in natural order of the name without
// its extension and then of the whole name; entries named with a leading dot, the library's own settings and output
// template folders, though not folders so named deeper down, the Master Library's copy of the User Guide, though not
// another library's folder at that path, and the files of special types are left out; a symlink to a folder that holds
// something is listed as a folder and reported, so that the library's watch can follow it, while one to an empty folder
// or to a file is listed as a file of the type its own name says, and one that leads back to a folder it lies within is
// listed as a folder with nothing beneath it and not reported.
func TestListFolderListsWhatTheTreeShows(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes()
	lib := newTestLibrary(t)
	root := lib.Path(false)
	for _, dir := range []string{
		"sub/inner", "zeta", ".hidden", library.SettingsDirName,
		library.OutputTemplatesDirName, "sub/" + library.SettingsDirName, "empty",
	} {
		c.NoError(os.MkdirAll(filepath.Join(root, dir), 0o750))
	}
	for _, name := range []string{
		"b.gcs", "a.skl", "10.gcs", "9.gcs", "alpha.gcs", "same.gcs", "same.skl", ".dotfile",
		"weird" + gurps.ClosedFolder, "sub/inner/deep.gcs", "sub/x.gcs",
	} {
		c.NoError(os.WriteFile(filepath.Join(root, name), []byte("{}"), 0o600))
	}
	c.Equal([]string{
		"empty/", "sub/", "sub/inner/", "sub/inner/deep.gcs", "sub/" + library.SettingsDirName + "/", "sub/x.gcs",
		"zeta/", "9.gcs", "10.gcs", "a.skl", "alpha.gcs", "b.gcs", "same.gcs", "same.skl",
	}, flattenEntries(listFolder(lib, ".", true, ignoreSymlinkedFolder)), "grouped")
	mixed := listFolder(lib, ".", false, ignoreSymlinkedFolder)
	top := make([]string, 0, len(mixed))
	for _, entry := range mixed {
		top = append(top, entry.path)
	}
	c.Equal([]string{"9.gcs", "10.gcs", "a.skl", "alpha.gcs", "b.gcs", "empty", "same.gcs", "same.skl", "sub", "zeta"},
		top, "mixed in")

	master := library.NewLibraries().Master()
	c.NoError(master.SetPath(t.TempDir()))
	other := newTestLibrary(t)
	for _, one := range []*library.Library{master, other} {
		for _, dir := range []string{"Markdown/User Guide", "Markdown/Mine"} {
			c.NoError(os.MkdirAll(filepath.Join(one.Path(false), dir), 0o750))
			c.NoError(os.WriteFile(filepath.Join(one.Path(false), dir, "Home.md"), []byte("# Home\n"), 0o600))
		}
	}
	c.Equal([]string{"Markdown/", "Markdown/Mine/", "Markdown/Mine/Home.md"},
		flattenEntries(listFolder(master, ".", true, ignoreSymlinkedFolder)),
		"the Master Library's copy of the User Guide must be left out")
	c.Equal([]string{
		"Markdown/", "Markdown/Mine/", "Markdown/Mine/Home.md", "Markdown/User Guide/", "Markdown/User Guide/Home.md",
	}, flattenEntries(listFolder(other, ".", true, ignoreSymlinkedFolder)), "while another library's folder of that name is the user's own")

	if err := os.Symlink(filepath.Join(root, "sub"), filepath.Join(root, "link-sub")); err != nil {
		t.Skipf("unable to create a symlink here: %v", err) // Windows requires a privilege for this by default
	}
	c.NoError(os.Symlink(filepath.Join(root, "empty"), filepath.Join(root, "link-empty.gcs")))
	c.NoError(os.Symlink(filepath.Join(root, "b.gcs"), filepath.Join(root, "link-file.gcs")))
	var reported []string
	byPath := make(map[string]*navigatorEntry)
	for _, entry := range listFolder(lib, ".", true, func(reportedLib *library.Library, dirPath string) {
		c.True(reportedLib == lib)
		reported = append(reported, dirPath)
	}) {
		byPath[entry.path] = entry
	}
	c.Equal([]string{"link-sub"}, reported, "only a symlink to a folder that holds something is reported")
	c.True(byPath["link-sub"] != nil && byPath["link-sub"].isDir, "and listed as a folder")
	c.Equal([]string{
		"link-sub/inner/", "link-sub/inner/deep.gcs", "link-sub/" + library.SettingsDirName + "/",
		"link-sub/x.gcs",
	}, flattenEntries(byPath["link-sub"].entries), "with what the folder holds beneath it")
	c.True(byPath["link-empty.gcs"] != nil && !byPath["link-empty.gcs"].isDir, "a symlink to an empty folder is a file")
	c.True(byPath["link-file.gcs"] != nil && !byPath["link-file.gcs"].isDir, "as is one to a file")

	loopLib := newTestLibrary(t)
	loopRoot := loopLib.Path(false)
	c.NoError(os.MkdirAll(filepath.Join(loopRoot, "a"), 0o750))
	c.NoError(os.WriteFile(filepath.Join(loopRoot, "a", "b.gcs"), []byte("{}"), 0o600))
	c.NoError(os.Symlink(loopRoot, filepath.Join(loopRoot, "a", "up")))
	reported = nil
	c.Equal([]string{"a/", "a/b.gcs", "a/up/"}, flattenEntries(listFolder(loopLib, ".", true,
		func(_ *library.Library, dirPath string) { reported = append(reported, dirPath) })),
		"a symlink that leads back to a folder it lies within must not have that folder listed within itself")
	c.Equal(0, len(reported), "nor be reported, since the folder it leads to is watched already")
}

// TestListFavoritesDropsWhatIsGone verifies that listing the favorites drops from each library the favorites that can
// no longer be found, lists the rest in the order the Favorites row shows them, folders ahead of files when they are
// grouped, and lists what a favorite folder holds, leaving out a favorite that is the Master Library's copy of the User
// Guide or lies within it, though not another library's folder at that path.
func TestListFavoritesDropsWhatIsGone(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes()
	lib := newTestLibrary(t)
	root := lib.Path(false)
	c.NoError(os.MkdirAll(filepath.Join(root, "sub"), 0o750))
	for _, name := range []string{"alpha.gcs", "beta.gcs", "sub/inner.gcs"} {
		c.NoError(os.WriteFile(filepath.Join(root, name), []byte("{}"), 0o600))
	}
	for _, p := range []string{"sub", "alpha.gcs", "gone.gcs", "beta.gcs"} {
		lib.ToggleFavorite(p)
	}
	entries := func(lib *library.Library, grouped bool) []string {
		favs := listFavorites([]*library.Library{lib}, grouped, ignoreSymlinkedFolder)
		paths := make([]string, 0, len(favs))
		for _, one := range favs {
			c.True(one.library == lib)
			paths = append(paths, flattenEntries([]*navigatorEntry{one.entry})...)
		}
		return paths
	}
	c.Equal([]string{"sub/", "sub/inner.gcs", "alpha.gcs", "beta.gcs"}, entries(lib, true), "grouped")
	c.Equal([]string{"alpha.gcs", "beta.gcs", "sub"}, lib.Favorites(), "the favorite that is gone must be dropped")
	c.Equal([]string{"alpha.gcs", "beta.gcs", "sub/", "sub/inner.gcs"}, entries(lib, false), "mixed in")

	master := library.NewLibraries().Master()
	c.NoError(master.SetPath(t.TempDir()))
	other := newTestLibrary(t)
	for _, one := range []*library.Library{master, other} {
		c.NoError(os.MkdirAll(filepath.Join(one.Path(false), "Markdown", "User Guide"), 0o750))
		c.NoError(os.WriteFile(filepath.Join(one.Path(false), "Markdown", "User Guide", "Home.md"), []byte("# Home\n"),
			0o600))
		for _, p := range []string{"Markdown", "Markdown/User Guide", "Markdown/User Guide/Home.md"} {
			one.ToggleFavorite(p)
		}
	}
	c.Equal([]string{"Markdown/"}, entries(master, true),
		"a favorite that is the Master Library's copy of the User Guide, or lies within it, must be left out")
	c.Equal([]string{
		"Markdown/", "Markdown/User Guide/", "Markdown/User Guide/Home.md",
		"Markdown/User Guide/", "Markdown/User Guide/Home.md",
		"Markdown/User Guide/Home.md",
	}, entries(other, true), "while another library's folder of that name is the user's own")
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
