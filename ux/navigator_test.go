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
	"path"
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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/rjeczalik/notify"
)

// newDeepSearchNavigator returns a navigator with deep search enabled for the given extensions. It registers the known
// file types, which the deep search resolves each file's extension through. Only the search state is populated; a test
// that needs the table or the search field adds them with installTestTable and installTestSearch.
func newDeepSearchNavigator(exts ...string) *Navigator {
	RegisterKnownFileTypes()
	n := &Navigator{deepSearch: make(map[string]bool, len(exts))}
	for _, ext := range exts {
		n.deepSearch[ext] = true
	}
	return n
}

// installTestTable gives the navigator a real table holding the given rows.
func installTestTable(n *Navigator, rows ...*NavigatorNode) {
	n.table = unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{})
	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.table.SetRootRows(rows)
}

// installTestSearch gives the navigator its search field and dropdown, as setupToolBar does. Without a window the
// field can't hold the focus, so the rows are never shown and typing in the field finds nothing; a test that wants rows
// on offer sets them on the dropdown directly.
func installTestSearch(n *Navigator) {
	n.search = newSearchDropdown(n.findMatches, func(match *searchMatch) { match.show() })
	n.search.status = n.searchStatus
	unison.NewPanel().AddChild(n.search.field)
}

// settleContentLoad waits for the background read a search started for the files it found no usable content entry
// for, if there is one, and installs its result as its completion task would, since nothing in these tests pumps
// unison's task queue. Returns true if there was one.
func settleContentLoad(n *Navigator) bool {
	load := n.contentLoad
	if load == nil {
		return false
	}
	<-load.done
	n.applyContentLoad(load)
	return true
}

// matchCount returns how many of the rows, and those beneath them, match the lower-cased text, as the dropdown ends up
// showing: the files the walk found no usable content entry for are read as the background load a search starts reads
// them, and the rows are walked again.
func matchCount(n *Navigator, text string, rows ...*NavigatorNode) int {
	var found []*searchMatch
	n.missingContent = nil
	n.collectMatches(&found, text, rows)
	n.loadMissingContent()
	if !settleContentLoad(n) {
		return len(found)
	}
	found = nil
	n.missingContent = nil
	n.collectMatches(&found, text, rows)
	return len(found)
}

// newTestLibrary returns a library rooted in a fresh temporary directory.
func newTestLibrary(t *testing.T) *library.Library {
	return library.NewLibrary("Test", "test", "", "test", t.TempDir())
}

// writeTestLibraryFile writes the data to the named file in the library's directory and returns the file's node.
func writeTestLibraryFile(c check.Checker, lib *library.Library, fileName string, data []byte) *NavigatorNode {
	c.NoError(os.WriteFile(filepath.Join(lib.Path(false), fileName), data, 0o600))
	return NewFileNode(lib, fileName, nil)
}

// newTestCharacter returns a character with the given name.
func newTestCharacter(name string) *gurps.Entity {
	entity := gurps.NewEntity()
	entity.Profile.Name = name
	return entity
}

// saveTestSheet saves the character as a sheet in the named file in the library's directory and returns the file's
// node.
func saveTestSheet(c check.Checker, lib *library.Library, fileName string, entity *gurps.Entity) *NavigatorNode {
	c.NoError(entity.Save(filepath.Join(lib.Path(false), fileName)))
	return NewFileNode(lib, fileName, nil)
}

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

// TestPrepareProfileForContentCache verifies that every profile field placed into the deep search content cache is
// lowercased, since the search text is lowercased before the comparison is made.
func TestPrepareProfileForContentCache(t *testing.T) {
	c := check.New(t)
	content := prepareProfileForContentCache(&gurps.Profile{
		Name:         "Conan",
		Age:          "Thirty",
		Birthday:     "January 1",
		Eyes:         "Blue",
		Hair:         "Black",
		Skin:         "Tan",
		Handedness:   "Right",
		Gender:       "Male",
		PlayerName:   "Robert",
		Title:        "King Of Aquilonia",
		Organization: "The Black Dragons",
		Religion:     "Crom",
	})
	for _, one := range []string{
		"conan",
		"thirty",
		"january 1",
		"blue",
		"black",
		"tan",
		"right",
		"male",
		"robert",
		"king of aquilonia",
		"the black dragons",
		"crom",
	} {
		c.True(strings.Contains(content, one), "expected %q in the cached content, got %q", one, content)
	}
}

// TestDeepSearchOfSheetProfileIsCaseInsensitive verifies that a deep search of a character sheet matches profile
// fields regardless of case. The search text is lowercased before comparison, so a character named "Conan" must be
// found when searching for "conan", even though the file name doesn't match.
func TestDeepSearchOfSheetProfileIsCaseInsensitive(t *testing.T) {
	c := check.New(t)
	entity := newTestCharacter("Conan")
	entity.Profile.PlayerName = "Robert"
	entity.Profile.Organization = "The Black Dragons"
	node := saveTestSheet(c, newTestLibrary(t), "sheet1"+gurps.SheetExt, entity)
	n := newDeepSearchNavigator(gurps.SheetExt)
	for _, one := range []struct {
		name string
		text string
	}{
		{name: "character name", text: "conan"},
		{name: "partial character name", text: "cona"},
		{name: "player name", text: "robert"},
		{name: "organization", text: "the black dragons"},
	} {
		c.Equal(1, matchCount(n, one.text, node), one.name)
	}

	// Something that isn't present must still not match.
	c.Equal(0, matchCount(n, "belit", node))
}

// TestDeepSearchServesCacheHitsUntilInvalidated verifies that a search caches a markdown file's text, lowercased and
// with surrounding whitespace trimmed, under the file's full path on disk, that it serves a cache hit without checking
// the file on disk — revalidating every hit would be a stat per deep-searchable file on each keystroke — and that
// dropping the entry, as the filesystem watches do when they report the file changed, makes the next search re-read
// the file rather than match on its previous contents. A symlinked file is the exception, since the watches report a
// change to the file it points at under that file's path, and not at all when it lies outside the library: a hit on it
// is checked against that file.
func TestDeepSearchServesCacheHitsUntilInvalidated(t *testing.T) {
	c := check.New(t)
	ext := uti.Markdown.Extensions[0]
	lib := newTestLibrary(t)
	fileName := "notes" + ext
	p := filepath.Join(lib.Path(false), fileName)
	node := writeTestLibraryFile(c, lib, fileName, []byte("# Conan The Barbarian\n"))
	n := newDeepSearchNavigator(ext)
	c.Equal(1, matchCount(n, "barbarian", node), "the original content must match and populate the cache")
	cached, ok := n.contentCache[p]
	c.True(ok, "expected the markdown content to be cached for %q", p)
	c.Equal("# conan the barbarian", cached.content, "the cached content must be lowercased and trimmed")

	// Rewrite the file with different content and a clearly different modification time, so that a search which did
	// check the file against the entry would notice the change.
	c.NoError(os.WriteFile(p, []byte("bravo bravo\n"), 0o600))
	c.NoError(os.Chtimes(p, time.Time{}, time.Now().Add(time.Hour)))

	c.Equal(1, matchCount(n, "barbarian", node), "a cache hit must be served without checking the file on disk")
	c.Equal(0, matchCount(n, "bravo", node), "a cache hit must be served without checking the file on disk")

	// Once the watch reports the change, the next search must re-read the file.
	n.watchCallback(lib, p, notify.Write)
	_, ok = n.contentCache[p]
	c.False(ok, "the reported change must drop the entry")
	c.Equal(1, matchCount(n, "bravo", node), "the new content must be found after the change is reported")
	c.Equal(0, matchCount(n, "barbarian", node), "the previous content must no longer match")
	c.True(n.contentCache[p].isCurrent(p), "the re-read must capture the file's current metadata")

	target := filepath.Join(t.TempDir(), "elsewhere"+ext)
	c.NoError(os.WriteFile(target, []byte("charlie\n"), 0o600))
	if err := os.Symlink(target, filepath.Join(lib.Path(false), "shortcut"+ext)); err != nil {
		t.Skipf("unable to create a symlink here: %v", err) // Windows requires a privilege for this by default
	}
	c.NoError(os.Symlink(p, filepath.Join(lib.Path(false), "same"+ext)))
	outside := NewFileNode(lib, "shortcut"+ext, nil)
	inside := NewFileNode(lib, "same"+ext, nil)
	c.Equal(1, matchCount(n, "charlie", outside), "the content of the file a symlink points at must match")
	c.Equal(1, matchCount(n, "bravo", inside), "the content of the file a symlink points at must match")
	c.NoError(os.WriteFile(target, []byte("delta\n"), 0o600))
	c.NoError(os.Chtimes(target, time.Time{}, time.Now().Add(2*time.Hour)))
	c.Equal(1, matchCount(n, "delta", outside), "a change to a file outside the library a symlink points at, "+
		"which no watch reports, must be found")
	c.NoError(os.WriteFile(p, []byte("echo\n"), 0o600))
	c.NoError(os.Chtimes(p, time.Time{}, time.Now().Add(2*time.Hour)))
	n.watchCallback(lib, p, notify.Write)
	c.Equal(1, matchCount(n, "echo", inside), "a change reported under the path of the file a symlink points at "+
		"must be found under the symlink's")

	// A symlink whose target is gone is re-read once, and the empty entry that produces is then served while the
	// target stays gone, rather than being re-read on every search because it can never be found current.
	c.NoError(os.Remove(target))
	c.Equal(0, matchCount(n, "delta", outside), "the content of a target that is gone must no longer match")
	shortcut := filepath.Join(lib.Path(false), "shortcut"+ext)
	entry := n.contentCache[shortcut]
	c.NotNil(entry, "the symlink whose target is gone must still be cached")
	c.Equal(0, matchCount(n, "delta", outside))
	c.True(n.contentCache[shortcut] == entry, "a symlink whose target is gone must not be re-read on every search")
	c.NoError(os.WriteFile(target, []byte("delta\n"), 0o600))
	c.NoError(os.Chtimes(target, time.Time{}, time.Now().Add(3*time.Hour)))
	c.Equal(1, matchCount(n, "delta", outside), "a target that is back must be read again")
}

// TestInvalidationReachesInFlightCacheBuild verifies that a file change reported while a background cache build is
// running is applied to that build's result, since the build may have read the file before it changed: the entry is
// dropped from the result when it is installed, a superseded build is discarded, leaving the cache as it was and the
// invalidation for the newer build to apply, and starting a build consumes the invalidations made before it, which its
// starting snapshot already lacks.
func TestInvalidationReachesInFlightCacheBuild(t *testing.T) {
	c := check.New(t)
	n := &Navigator{contentCache: map[string]*contentCacheEntry{"a": {}, "b": {}}}
	gen := n.cacheGeneration.Add(1) // A build is in flight from here

	n.invalidateContentCacheEntry("a")
	_, ok := n.contentCache["a"]
	c.False(ok, "the live cache must drop the entry at once")
	c.NotNil(n.contentCache["b"], "other entries must be untouched")

	fresh := map[string]*contentCacheEntry{"a": {}, "b": {}, "c": {}}
	c.True(n.applyPrewarmedContentCache(gen, fresh))
	_, ok = n.contentCache["a"]
	c.False(ok, "the installed build must lack the entry invalidated while it ran")
	c.True(fresh["b"] == n.contentCache["b"], "the installed build must keep the rest")
	c.Equal(0, len(n.invalidatedPaths), "an installed build must consume the invalidations")

	// A superseded build must be discarded, leaving the cache as it was and the invalidation for the newer build's
	// result.
	n.invalidateContentCacheEntry("b")
	n.cacheGeneration.Add(1)
	c.False(n.applyPrewarmedContentCache(gen, map[string]*contentCacheEntry{"b": {}}),
		"a superseded build must not be installed")
	c.True(fresh["c"] == n.contentCache["c"], "a superseded build must not clobber the cache")
	c.True(n.invalidatedPaths["b"], "a discarded build must not consume the invalidations")
	c.True(n.applyPrewarmedContentCache(n.cacheGeneration.Load(), map[string]*contentCacheEntry{"b": {}}))
	_, ok = n.contentCache["b"]
	c.False(ok, "the newer build's result must lack the entry invalidated while the older one ran")

	// Starting a build consumes the invalidations made before it, since it starts from a cache that already lacks
	// those entries. The table is real but empty, so the prewarm runs without spawning a background build.
	installTestTable(n)
	n.invalidateContentCacheEntry("c")
	n.prewarmContentCache()
	c.Equal(0, len(n.invalidatedPaths), "a build's start must consume the invalidations made before it")
}

// TestNavigatorSearchMatches verifies what a match carries: the name the tree shows for the row, the names of its
// ancestors from the root down as its title, the icon of its kind, and a key that follows the row's place in the tree,
// so that the dropdown can keep its highlight across a Reload, which rebuilds the tree from fresh node objects. A file
// can appear in several places: under its library and under Favorites, both as a favorite and within a favorited
// folder, and under a library rooted within another's folder. Each copy must have a key of its own.
func TestNavigatorSearchMatches(t *testing.T) {
	c := check.New(t)
	lib := newTestLibrary(t)
	n := newDeepSearchNavigator()
	// The container nodes are built by hand rather than through the constructors, which read the filesystem and the
	// global settings; only the shape of the tree matters here.
	favorites := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorFavorites)}
	libraryNode := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorLibrary), library: lib}
	sub := &NavigatorNode{
		id:      tid.MustNewTID(kinds.NavigatorDirectory),
		path:    "sub",
		library: lib,
		parent:  libraryNode,
	}
	name := path.Join("sub", "Conan"+gurps.SheetExt)
	underFavorites := NewFileNode(lib, name, favorites)
	underLibrary := NewFileNode(lib, name, sub)
	favoriteSub := &NavigatorNode{
		id:      tid.MustNewTID(kinds.NavigatorDirectory),
		path:    "sub",
		library: lib,
		parent:  favorites,
	}
	inFavoriteSub := NewFileNode(lib, name, favoriteSub)
	nested := library.NewLibrary("Nested", "nested", "", "nested", filepath.Join(lib.Path(false), "sub"))
	inNested := NewFileNode(nested, "Conan"+gurps.SheetExt,
		&NavigatorNode{id: tid.MustNewTID(kinds.NavigatorLibrary), library: nested})
	c.Equal(underLibrary.Path(), inNested.Path(), "sanity: the nested library's copy is the same file")
	fresh := NewFileNode(lib, name, sub)
	c.True(underLibrary.ID() != fresh.ID(), "sanity: a rebuilt tree's nodes have fresh IDs")

	match := n.newSearchMatch(underLibrary)
	c.Equal("Conan", match.name, "a file is named without its extension, as in the tree")
	c.Equal("Test / sub", match.title, "the title names the ancestors from the root down")
	c.Equal(match.key, n.newSearchMatch(fresh).key, "the key follows the path to a fresh node")
	c.NotEqual(match.key, n.newSearchMatch(underFavorites).key, "the favorites copy is told apart")
	c.NotEqual(n.newSearchMatch(underFavorites).key, n.newSearchMatch(inFavoriteSub).key,
		"a favorite is told apart from its copy within a favorited folder")
	c.NotEqual(match.key, n.newSearchMatch(inNested).key, "the nested library's copy is told apart")
	c.NotEqual("", n.newSearchMatch(favorites).key, "the Favorites row has a key, so its highlight can be kept")
	c.Equal("Favorites", n.newSearchMatch(underFavorites).title)
	c.Equal("", n.newSearchMatch(libraryNode).title, "a root row has no ancestors")
	icon, ok := match.icon.(*unison.DrawableSVG)
	c.True(ok && icon.SVG == gurps.FileInfoFor(gurps.SheetExt).SVG, "a file shows the icon of its type")
	icon, ok = n.newSearchMatch(sub).icon.(*unison.DrawableSVG)
	c.True(ok && icon.SVG == gurps.FileInfoFor(gurps.ClosedFolder).SVG, "a container shows a folder")
}

// TestNavigatorRefreshSearchWhileHiddenLeavesTableAlone verifies what refreshing the search does while the matches are
// hidden, as they are when the asynchronous cache prewarm completes after the user has clicked away: the rows on offer
// are dropped without walking the tree, so that they are found afresh when next asked for, and the table's selection is
// left where the user put it. Choosing a match is what selects it, alone. A navigator without a search, as some tests
// build, has nothing to refresh.
func TestNavigatorRefreshSearchWhileHiddenLeavesTableAlone(t *testing.T) {
	c := check.New(t)
	lib := newTestLibrary(t)
	rows := []*NavigatorNode{
		NewFileNode(lib, "one"+gurps.NotesExt, nil),
		NewFileNode(lib, "two"+gurps.NotesExt, nil),
		NewFileNode(lib, "three"+gurps.NotesExt, nil),
	}
	n := newDeepSearchNavigator()
	installTestTable(n, rows...)
	n.refreshSearch() // Nothing to refresh yet
	installTestSearch(n)
	walks := 0
	findMatches := n.search.findMatches
	n.search.findMatches = func(text string) []*searchMatch {
		walks++
		return findMatches(text)
	}
	n.search.field.SetText("t")
	c.Equal(0, walks, "text set while the field lacks the focus must not walk the tree")
	n.search.suggestions.SetRows(findMatches("t")) // As typing with the focus would
	c.Equal(2, n.search.suggestions.List().Count())
	c.False(n.search.suggestions.Showing(), "without a window, the rows can't show")

	// The user has clicked on the first row.
	n.table.SelectByIndex(0)
	n.refreshSearch()
	c.Equal(0, walks, "a refresh while the matches are hidden must not walk the tree")
	c.Equal(0, n.search.suggestions.List().Count(), "and drops the rows on offer")
	c.Equal(0, n.table.FirstSelectedRowIndex(), "the selection is left alone")
	c.Equal(1, n.table.SelectionCount())

	n.search.choose(findMatches("t")[1])
	c.Equal(2, n.table.FirstSelectedRowIndex(), "choosing a match selects it")
	c.Equal(1, n.table.SelectionCount(), "alone")
}

// TestDeepSearchDisabledTypeIgnoresCache verifies that disabling a file type for deep search takes effect immediately,
// even while the content cache still holds entries for that type. The cache is rebuilt asynchronously after a settings
// change, so the search must gate on the deep search setting before consulting the cache, not after.
func TestDeepSearchDisabledTypeIgnoresCache(t *testing.T) {
	c := check.New(t)
	lib := newTestLibrary(t)
	fileName := "sheet1" + gurps.SheetExt
	p := filepath.Join(lib.Path(false), fileName)
	node := saveTestSheet(c, lib, fileName, newTestCharacter("Conan"))
	n := newDeepSearchNavigator(gurps.SheetExt)
	c.Equal(1, matchCount(n, "conan", node))
	_, ok := n.contentCache[p]
	c.True(ok, "expected the sheet's content to be cached for %q", p)

	// Disabling the type must stop matches at once, despite the stale cache entry that remains until the rebuild.
	n.deepSearch = make(map[string]bool)
	c.Equal(0, matchCount(n, "conan", node), "a disabled type must not match, even with its entry still cached")

	// Re-enabling must match again, served from the still-present cache entry.
	n.deepSearch = map[string]bool{gurps.SheetExt: true}
	c.Equal(1, matchCount(n, "conan", node))
}

// TestBuildContentCacheReuseAndFailureCaching verifies the behavior the background cache prewarm relies on: entries
// for unchanged files are reused rather than re-parsed, a changed file gets a fresh entry, a file that fails to parse
// — whether by error or by panic — is cached with empty content so it isn't re-read on every keystroke, and a
// canceled build stops early.
func TestBuildContentCacheReuseAndFailureCaching(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes() // The content loader resolves each file's extension through the file type registry.
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "guide"+uti.Markdown.Extensions[0])
	c.NoError(os.WriteFile(mdPath, []byte("# Alpha\n"), 0o600))
	badPath := filepath.Join(dir, "broken"+gurps.SkillsExt)
	c.NoError(os.WriteFile(badPath, []byte("not json"), 0o600))
	// A null row is valid JSON, so this gets past parsing and panics when the nil skill is traversed.
	panicPath := filepath.Join(dir, "panics"+gurps.SkillsExt)
	c.NoError(os.WriteFile(panicPath, []byte(`{"version":5,"rows":[null]}`), 0o600))
	paths := map[string]bool{mdPath: true, badPath: true, panicPath: true}

	cache := buildContentCache(paths, nil, nil)
	c.Equal(3, len(cache))
	c.Equal("# alpha", cache[mdPath].content)
	c.True(cache[badPath] != nil, "a file that fails to parse must still be cached")
	c.Equal("", cache[badPath].content)
	c.True(cache[panicPath] != nil, "a file whose parse panics must still be cached")
	c.Equal("", cache[panicPath].content)

	// With nothing changed on disk, the entries must be reused rather than rebuilt.
	again := buildContentCache(paths, cache, nil)
	c.True(again[mdPath] == cache[mdPath], "the unchanged markdown entry must be reused")
	c.True(again[badPath] == cache[badPath], "the unchanged failure entry must be reused")
	c.True(again[panicPath] == cache[panicPath], "the unchanged panic entry must be reused")

	// A changed file must get a fresh entry. The modification time is set explicitly so the test doesn't depend on
	// the filesystem's timestamp granularity.
	c.NoError(os.WriteFile(mdPath, []byte("# Beta Content\n"), 0o600))
	c.NoError(os.Chtimes(mdPath, time.Time{}, cache[mdPath].modTime.Add(time.Second)))
	rebuilt := buildContentCache(paths, cache, nil)
	c.True(rebuilt[mdPath] != cache[mdPath], "the changed markdown entry must be rebuilt")
	c.Equal("# beta content", rebuilt[mdPath].content)
	c.True(rebuilt[badPath] == cache[badPath], "the unchanged failure entry must still be reused")

	// A build whose cancellation function reports true must stop without producing entries.
	c.Equal(0, len(buildContentCache(paths, nil, func() bool { return true })))
}

// TestDeepSearchCachesFileWhoseParsePanics verifies that the background read a keystroke starts for a file the cache
// lacks survives a file whose parse panics, and that the panic is cached like a parse error, so the file isn't re-read
// — and doesn't re-panic — on every subsequent keystroke.
func TestDeepSearchCachesFileWhoseParsePanics(t *testing.T) {
	c := check.New(t)
	lib := newTestLibrary(t)
	fileName := "panics" + gurps.SkillsExt
	p := filepath.Join(lib.Path(false), fileName)
	// A null row is valid JSON, so this gets past parsing and panics when the nil skill is traversed.
	node := writeTestLibraryFile(c, lib, fileName, []byte(`{"version":5,"rows":[null]}`))
	n := newDeepSearchNavigator(gurps.SkillsExt)
	c.Equal(0, matchCount(n, "anything", node))
	entry, ok := n.contentCache[p]
	c.True(ok, "expected the panicking file to be cached for %q", p)
	c.Equal("", entry.content)
}

// TestSupersededCacheBuildHandsEntriesToItsSuccessor verifies that a background cache build inherits the entries of the
// build it supersedes rather than starting over from the live cache alone, which is what keeps the back-to-back reloads
// at startup from restarting the first full parse of a library from zero. The hand-off rules are checked against a
// synthetic prior build, whose contents are known: its entries fill in what the live cache lacks, the live cache's own
// entry wins where both hold one, and its entry for a file whose change was reported after it began is not inherited,
// since it may have read the old contents. A real pair of builds then confirms the wiring: the second waits for the
// first to stop and reuses whatever it completed. Neither build's completion task runs, since nothing here pumps
// unison's task queue, so the results are read from the build records directly.
func TestSupersededCacheBuildHandsEntriesToItsSuccessor(t *testing.T) {
	c := check.New(t)
	ext := uti.Markdown.Extensions[0]
	lib := newTestLibrary(t)
	names := []string{"a" + ext, "b" + ext, "c" + ext}
	paths := make([]string, len(names))
	nodes := make([]*NavigatorNode, len(names))
	for i, name := range names {
		paths[i] = filepath.Join(lib.Path(false), name)
		nodes[i] = writeTestLibraryFile(c, lib, name, []byte("content "+name+"\n"))
	}
	n := newDeepSearchNavigator(ext)
	installTestTable(n, nodes...)

	// The synthetic prior build holds a current entry for each file. The live cache holds its own entry for the
	// second, and the third's change is reported after the prior build began.
	prior := &contentCacheBuild{done: make(chan struct{}), entries: make(map[string]*contentCacheEntry)}
	for _, p := range paths {
		prior.entries[p] = loadContentCacheEntry(p)
	}
	close(prior.done)
	live := loadContentCacheEntry(paths[1])
	n.contentCache = map[string]*contentCacheEntry{paths[1]: live}
	n.lastBuild = prior
	n.invalidateContentCacheEntry(paths[2])
	n.prewarmContentCache()
	build := n.lastBuild
	c.True(build != nil && build != prior, "a prewarm must record its build for the next one to inherit from")
	<-build.done
	c.Equal(len(paths), len(build.entries))
	c.True(build.entries[paths[0]] == prior.entries[paths[0]], "an entry the live cache lacks must be inherited")
	c.True(build.entries[paths[1]] == live, "the live cache's entry must win over the prior build's")
	c.True(build.entries[paths[2]] != prior.entries[paths[2]], "an entry invalidated since the prior build began must not be inherited")
	c.Equal("content "+names[2], build.entries[paths[2]].content, "the invalidated file must be re-read instead")

	// A real pair: the second build cancels the first and must wait for it, then reuse everything it completed. How
	// much that is depends on timing, so the check is that whatever the first recorded, the second holds the same entry
	// for, rather than a fresh one.
	n.contentCache = nil
	n.lastBuild = nil
	n.prewarmContentCache()
	first := n.lastBuild
	n.prewarmContentCache()
	second := n.lastBuild
	c.True(first != nil && second != nil && first != second)
	<-second.done
	select {
	case <-first.done:
	default:
		c.Fatal("the second build must not finish before the first has stopped")
	}
	c.Equal(len(paths), len(second.entries), "the second build was not canceled, so it must be complete")
	for p, entry := range first.entries {
		c.True(second.entries[p] == entry, "the second build must reuse the entry the first completed for %q", p)
	}
}

// TestCollectDeepSearchPaths verifies that collecting the paths for the cache prewarm recurses through container
// nodes, that a file reachable both under the favorites node and under its library collapses to a single path, and
// that files whose types aren't enabled for deep search are excluded.
func TestCollectDeepSearchPaths(t *testing.T) {
	c := check.New(t)
	ext := uti.Markdown.Extensions[0]
	lib := newTestLibrary(t)
	n := newDeepSearchNavigator(ext)
	shared := "notes" + ext
	nested := path.Join("sub", "extra"+ext)
	// The container nodes are built by hand rather than through the constructors, which read the filesystem and the
	// global settings, and their IDs are minted directly rather than through gurps.IDForNavNode, which records every
	// path it is asked about in the global settings; only the shape of the tree matters here.
	favorites := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorFavorites)}
	favorites.children = []*NavigatorNode{NewFileNode(lib, shared, favorites)}
	libraryNode := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorLibrary), library: lib}
	sub := &NavigatorNode{
		id:      tid.MustNewTID(kinds.NavigatorDirectory),
		path:    "sub",
		library: lib,
		parent:  libraryNode,
	}
	sub.children = []*NavigatorNode{
		NewFileNode(lib, nested, sub),
		NewFileNode(lib, path.Join("sub", "excluded.txt"), sub),
	}
	libraryNode.children = []*NavigatorNode{
		sub,
		NewFileNode(lib, shared, libraryNode), // The same file the favorites node holds
	}
	paths := make(map[string]bool)
	n.collectDeepSearchPaths([]*NavigatorNode{favorites, libraryNode}, paths)
	c.Equal(2, len(paths))
	c.True(paths[filepath.Join(lib.Path(false), shared)], "the file favorites repeats must be collected exactly once")
	c.True(paths[filepath.Join(lib.Path(false), nested)], "the file nested in a subdirectory must be reached by recursion")
}

// TestContentCachePrewarmSuspension verifies the hold a library update places on the background cache rebuilds:
// suspending abandons any build in flight (by advancing the cache generation) and makes prewarm requests latch a
// pending flag instead of starting builds, suspensions nest, and only the last lift takes ownership of the single
// rebuild the held-off requests collapsed into. The bookkeeping is exercised directly rather than through
// resumeContentCachePrewarm, whose only addition is to hand that rebuild to EventuallyReload -- a task on unison's
// process-global queue, which nothing in these tests pumps.
func TestContentCachePrewarmSuspension(t *testing.T) {
	c := check.New(t)
	// The table is real but empty, so a prewarm request that gets past the suspension guard would advance the
	// generation (and then find nothing to build) rather than bail out at the nil-table guard, leaving the suspension
	// as the only thing that keeps the generation still.
	n := &Navigator{}
	installTestTable(n)
	gen := n.cacheGeneration.Load()

	c.False(n.liftContentCachePrewarmSuspension(), "a lift with no suspension in force must be harmless")
	c.Equal(0, n.prewarmSuspensions)
	c.False(n.prewarmPending, "a lift with no suspension in force must not invent a rebuild")

	n.suspendContentCachePrewarm()
	c.Equal(gen+1, n.cacheGeneration.Load(), "the first suspension must abandon any build in flight")
	c.True(n.prewarmPending, "a rebuild must be owed, since a build may just have been abandoned")

	n.suspendContentCachePrewarm() // Suspensions nest, and only the first abandons a build.
	c.Equal(gen+1, n.cacheGeneration.Load(), "a nested suspension must not abandon anything")

	n.prewarmPending = false // Cleared so the request below is what latches it, not the suspension above.
	n.prewarmContentCache()
	c.Equal(gen+1, n.cacheGeneration.Load(), "a prewarm request while suspended must not start a build")
	c.True(n.prewarmPending, "a prewarm request while suspended must latch the pending rebuild")

	c.False(n.liftContentCachePrewarmSuspension(), "an outer suspension is still in force, so no rebuild is due yet")
	c.True(n.prewarmPending, "an inner lift must leave the rebuild pending")
	c.Equal(1, n.prewarmSuspensions)

	c.True(n.liftContentCachePrewarmSuspension(), "the last lift must report the single collapsed rebuild as due")
	c.False(n.prewarmPending, "the last lift must take ownership of the pending rebuild")
	c.Equal(0, n.prewarmSuspensions)

	c.False(n.liftContentCachePrewarmSuspension(), "an unmatched lift must be harmless")
	c.Equal(0, n.prewarmSuspensions)
	c.False(n.prewarmPending, "an unmatched lift must not report the rebuild as due twice")
}

// TestPrewarmWithNoDeepSearchPathsRefreshesSearch verifies the completion path taken when the collected path set is
// empty, as happens when the user unchecks the last deep-search type: no background build runs, so the prewarm itself
// must drop the cache and refresh the matches on offer, so that the now-invalid deep-search matches don't linger until
// the next keystroke.
func TestPrewarmWithNoDeepSearchPathsRefreshesSearch(t *testing.T) {
	c := check.New(t)
	n := newDeepSearchNavigator(gurps.SheetExt)
	installTestTable(n, saveTestSheet(c, newTestLibrary(t), "sheet1"+gurps.SheetExt, newTestCharacter("Conan")))
	installTestSearch(n)

	n.search.field.SetText("conan")
	c.Equal(0, len(n.search.findMatches("conan")), "the first search must start reading the sheet in the background")
	c.True(settleContentLoad(n), "and have a read to wait for")
	n.search.suggestions.SetRows(n.search.findMatches("conan")) // As typing with the focus would
	c.Equal(1, n.search.suggestions.List().Count(), "the sheet must match by content while its type is deep-searched")

	// Unchecking the last deep-search type leaves nothing to build; the prewarm itself must clear the stale matches.
	n.deepSearch = make(map[string]bool)
	n.prewarmContentCache()
	c.True(n.contentCache == nil, "the cache must be dropped")
	c.Equal(0, n.search.suggestions.List().Count(), "the deep-search matches must be gone immediately")
	c.Equal(0, len(n.search.findMatches("conan")), "and not be found again")
}

// TestSearchReadsMissingContentInTheBackground verifies that a search reads no file on the UI thread: a file the cache
// has no entry for is left out of the matches, the dropdown offers a row saying the file contents are still being
// indexed, a background read of the file is started, and once that is installed the search finds the file and the note
// is gone. While a build is in flight, as the first one is at startup, or a walk of the libraries, which ends in one,
// no read is started, since the build reads every file and its installation finds the matches again.
func TestSearchReadsMissingContentInTheBackground(t *testing.T) {
	c := check.New(t)
	ext := uti.Markdown.Extensions[0]
	lib := newTestLibrary(t)
	fileName := "notes" + ext
	p := filepath.Join(lib.Path(false), fileName)
	node := writeTestLibraryFile(c, lib, fileName, []byte("# Alpha\n"))
	n := newDeepSearchNavigator(ext)
	installTestTable(n, node)
	installTestSearch(n)

	found := n.findMatches("alpha")
	c.Equal(0, len(found), "a file whose contents are not cached must be left out")
	c.Equal(0, len(n.contentCache), "a search must not read the file on the UI thread")
	c.NotNil(n.contentLoad, "a search must start a background read of the files it could not check")
	rows := n.search.rowsFor(found)
	if len(rows) != 1 {
		t.Fatalf("expected the note alone, found %d rows", len(rows))
	}
	c.Equal(n.searchStatus(), rows[0].name, "the dropdown must say the contents are still being indexed")
	c.NotEqual("", rows[0].name)
	c.Nil(rows[0].show, "the note must choose nothing")
	load := n.contentLoad
	c.Equal(0, len(n.findMatches("alpha")))
	c.True(n.contentLoad == load, "a search made meanwhile must not start another read")

	c.True(settleContentLoad(n))
	c.Nil(n.contentLoad, "an installed read must make way for the next")
	c.NotNil(n.contentCache[p], "the read must be installed")
	found = n.findMatches("alpha")
	c.Equal(1, len(found), "the file must be found once its contents are cached")
	c.Equal("", n.searchStatus(), "with nothing left to index, there is no note")
	c.Equal(1, len(n.search.rowsFor(found)))

	// While a build is in flight, a search leaves the reading to it.
	n.contentCache = nil
	n.lastBuild = &contentCacheBuild{done: make(chan struct{})}
	c.Equal(0, len(n.findMatches("alpha")))
	c.Nil(n.contentLoad, "no read must be started while a build is in flight")
	c.NotEqual("", n.searchStatus(), "the note must say why the file is missing")
	n.lastBuild = nil
	n.walking = true
	c.Equal(0, len(n.findMatches("alpha")))
	c.Nil(n.contentLoad, "no read must be started while a walk of the libraries is in flight")
	n.walking = false
	built := buildContentCache(map[string]bool{p: true}, nil, nil)
	c.True(n.applyPrewarmedContentCache(n.cacheGeneration.Load(), built))
	c.Equal(1, len(n.findMatches("alpha")), "the installed build must serve the search")
	c.Equal("", n.searchStatus())
}

// TestContentLoadYieldsToBuildsAndInvalidations verifies the bookkeeping a background read a search starts shares with
// the builds: a read begun before a build started is discarded when it completes, since the build reads the file
// itself, a file whose change was reported while the read ran is left out of its result, since it may have been read
// before the change, and the next search has it read again. It also verifies that the read works from its own copy of
// the files it was handed, which the UI thread goes on noting files in while it runs.
func TestContentLoadYieldsToBuildsAndInvalidations(t *testing.T) {
	c := check.New(t)
	ext := uti.Markdown.Extensions[0]
	lib := newTestLibrary(t)
	fileName := "notes" + ext
	p := filepath.Join(lib.Path(false), fileName)
	node := writeTestLibraryFile(c, lib, fileName, []byte("# Alpha\n"))
	n := newDeepSearchNavigator(ext)
	installTestTable(n, node)

	c.Equal(0, len(n.findMatches("alpha")))
	load := n.contentLoad
	c.NotNil(load)
	n.cacheGeneration.Add(1) // A build starts from here
	<-load.done
	c.False(n.applyContentLoad(load), "a read begun before a build started must be discarded")
	c.Equal(0, len(n.contentCache), "and must not touch the cache")
	c.Nil(n.contentLoad, "a discarded read must still make way for the next")

	c.Equal(0, len(n.findMatches("alpha")))
	load = n.contentLoad
	c.NotNil(load)
	n.invalidateContentCacheEntry(p)
	<-load.done
	c.True(n.applyContentLoad(load), "a read that no build superseded must be installed")
	_, ok := n.contentCache[p]
	c.False(ok, "but without the entry for a file whose change was reported while it ran")
	c.Equal(1, matchCount(n, "alpha", node), "the next search must have the file read again")

	other := filepath.Join(lib.Path(false), "other"+ext)
	writeTestLibraryFile(c, lib, "other"+ext, []byte("# Beta\n"))
	n.contentCache = nil
	c.Equal(0, len(n.findMatches("alpha")))
	load = n.contentLoad
	c.NotNil(load)
	n.missingContent[other] = true // As a check of a file's contents made on the UI thread meanwhile notes it
	<-load.done
	_, ok = load.entries[other]
	c.False(ok, "a read must not take in a file noted as missing after it started")
	c.True(n.applyContentLoad(load))
}

// TestToggleFavoritesKeysOnFullPath verifies that the de-duplication done while toggling favorites is keyed on the full
// path on disk rather than the library-relative path. Two libraries can each hold the same relative path, and favorites
// are tracked per library, so both must be toggled.
func TestToggleFavoritesKeysOnFullPath(t *testing.T) {
	c := check.New(t)
	relPath := path.Join("Notes", "foo.not")
	lib1 := library.NewLibrary("One", "one", "", "one", t.TempDir())
	lib2 := library.NewLibrary("Two", "two", "", "two", t.TempDir())
	rows := []*NavigatorNode{
		NewFileNode(lib1, relPath, nil),
		NewFileNode(lib2, relPath, nil),
	}
	c.True(toggleFavorites(rows))
	c.Equal([]string{relPath}, lib1.Favorites())
	c.Equal([]string{relPath}, lib2.Favorites())

	// A repeated reference to the same file within one call must only toggle once, leaving it a favorite rather than
	// toggling it back off.
	lib3 := library.NewLibrary("Three", "three", "", "three", t.TempDir())
	c.True(toggleFavorites([]*NavigatorNode{
		NewFileNode(lib3, relPath, nil),
		NewFileNode(lib3, relPath, nil),
	}))
	c.Equal([]string{relPath}, lib3.Favorites())

	// A selection holding nothing but library nodes toggles nothing.
	c.False(toggleFavorites([]*NavigatorNode{NewLibraryNode(nil, lib1, nil)}))
	c.Equal([]string{relPath}, lib1.Favorites())
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
// template folders, though not folders so named deeper down, and the files of special types are left out; a symlink to
// a folder that holds something is listed as a folder and reported, so that the library's watch can follow it, while
// one to an empty folder or to a file is listed as a file of the type its own name says, and one that leads back to a
// folder it lies within is listed as a folder with nothing beneath it and not reported.
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
// grouped, and lists what a favorite folder holds.
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
	entries := func(grouped bool) []string {
		favs := listFavorites([]*library.Library{lib}, grouped, ignoreSymlinkedFolder)
		paths := make([]string, 0, len(favs))
		for _, one := range favs {
			c.True(one.library == lib)
			paths = append(paths, flattenEntries([]*navigatorEntry{one.entry})...)
		}
		return paths
	}
	c.Equal([]string{"sub/", "sub/inner.gcs", "alpha.gcs", "beta.gcs"}, entries(true), "grouped")
	c.Equal([]string{"alpha.gcs", "beta.gcs", "sub"}, lib.Favorites(), "the favorite that is gone must be dropped")
	c.Equal([]string{"alpha.gcs", "beta.gcs", "sub/", "sub/inner.gcs"}, entries(false), "mixed in")
}

// TestRowsForLeavesOutARemovedLibrary verifies that the rows built from a walk leave out a library, and its favorites,
// that was removed from the settings while the walk ran, since the reload its removal asked for is still to come and
// the library would otherwise show again until then.
func TestRowsForLeavesOutARemovedLibrary(t *testing.T) {
	c := check.New(t)
	libs := gurps.GlobalSettings().Libraries
	kept := library.NewLibrary("Kept", "kept", "", "kept", t.TempDir())
	libs.Store(kept.Key(), kept)
	t.Cleanup(func() { libs.Remove(kept.Key()) })
	removed := library.NewLibrary("Removed", "removed", "", "removed", t.TempDir())
	n := &Navigator{}
	listing := bareListing([]*library.Library{kept, removed})
	listing.favorites = []*favoriteEntry{
		{library: kept, entry: &navigatorEntry{path: "a.gcs"}},
		{library: removed, entry: &navigatorEntry{path: "b.gcs"}},
	}
	rows := n.rowsFor(listing)
	if len(rows) != 2 {
		t.Fatalf("expected the Favorites row and the kept library alone, got %d rows", len(rows))
	}
	c.True(rows[0].IsFavorites())
	c.Equal(1, len(rows[0].Children()), "the removed library's favorite must be left out")
	c.True(rows[0].Children()[0].library == kept)
	c.True(rows[1].library == kept)
}

// TestDoubleClickCountsTheFilesItOpens verifies that the count a double-click on the selection prompts with, when it
// is high, is of the distinct files it opens, leaving out the libraries and folders it opens or closes instead.
func TestDoubleClickCountsTheFilesItOpens(t *testing.T) {
	c := check.New(t)
	lib := newTestLibrary(t)
	libNode := NewLibraryNode(nil, lib, nil)
	folder := func(name string) *NavigatorNode { return NewDirectoryNode(nil, lib, name, libNode, nil) }
	folders := []*NavigatorNode{folder("A"), folder("B"), folder("C"), folder("D"), folder("E")}
	file := func(name string) *NavigatorNode { return NewFileNode(lib, name, libNode) }
	files := []*NavigatorNode{file("1.gcs"), file("2.gcs"), file("3.gcs"), file("4.gcs")}
	for _, one := range []struct {
		name string
		rows []*NavigatorNode
		want int
	}{
		{"folders alone", folders, 0},
		{"a library", []*NavigatorNode{libNode}, 0},
		{"a folder and files", append([]*NavigatorNode{folders[0]}, files...), 4},
		{"one file reached twice", []*NavigatorNode{files[0], NewFileNode(lib, "1.gcs", folders[0])}, 1},
	} {
		c.Equal(one.want, filesToOpen(one.rows), one.name)
	}
}

// TestTrimmedNameIsValid verifies that the name validation shared by the rename and new-folder dialogs both trims the
// name before judging it and reports back the trimmed name, so that callers build the path that was actually validated.
func TestTrimmedNameIsValid(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name    string
		in      string
		trimmed string
		valid   bool
	}{
		{name: "plain", in: "Notes", trimmed: "Notes", valid: true},
		{name: "surrounding whitespace is trimmed off", in: "  Notes  ", trimmed: "Notes", valid: true},
		{name: "empty", in: "", trimmed: "", valid: false},
		{name: "whitespace only", in: "   ", trimmed: "", valid: false},
		{name: "leading dot", in: ".hidden", trimmed: ".hidden", valid: false},
		{name: "leading dot after trimming", in: "  .hidden", trimmed: ".hidden", valid: false},
		{name: "path separator", in: "a/b", trimmed: "a/b", valid: false},
		{name: "path separator revealed by trimming", in: " a/b ", trimmed: "a/b", valid: false},
		{name: "reserved windows name", in: "con", trimmed: "con", valid: false},
		{name: "reserved windows name with whitespace", in: " CON ", trimmed: "CON", valid: false},
	} {
		trimmed, valid := trimmedNameIsValid(one.in)
		c.Equal(one.trimmed, trimmed, one.name)
		c.Equal(one.valid, valid, one.name)
	}
}

// TestPathsBuiltFromTheValidatedName verifies that the paths the rename and new-folder dialogs build are derived from
// the same trimmed name that trimmedNameIsValid judges, so a name entered with surrounding whitespace can't validate as
// one path and then create or rename to another.
func TestPathsBuiltFromTheValidatedName(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	for _, one := range []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "Notes", want: "Notes"},
		{name: "leading whitespace", in: "   Notes", want: "Notes"},
		{name: "trailing whitespace", in: "Notes   ", want: "Notes"},
		{name: "surrounding whitespace", in: "  Notes  ", want: "Notes"},
		{name: "interior whitespace is preserved", in: "  My Notes  ", want: "My Notes"},
	} {
		trimmed, valid := trimmedNameIsValid(one.in)
		c.True(valid, one.name)
		c.Equal(filepath.Join(dir, trimmed), newFolderPath(dir, one.in), one.name)
		c.Equal(filepath.Join(dir, one.want), newFolderPath(dir, one.in), one.name)

		oldPath := filepath.Join(dir, "old"+gurps.NotesExt)
		c.Equal(filepath.Join(dir, trimmed+gurps.NotesExt), renamedPath(oldPath, one.in), one.name)
		c.Equal(filepath.Join(dir, one.want+gurps.NotesExt), renamedPath(oldPath, one.in), one.name)
	}
}

// TestDeepSearchIndexesFromSavedCalc verifies that the content of a sheet, a note list, a template and a loot sheet is
// indexed from what the file recorded rather than by deriving it: each fixture's "calc" objects hold a resolved note
// text, and for the sheet and template a trait level, that deriving would not produce (see the fixtures' tests in the
// gurps package), and those are what must be searchable.
func TestDeepSearchIndexesFromSavedCalc(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes()
	dir := filepath.Join("..", "model", "gurps", "testdata")
	for _, ext := range []string{gurps.SheetExt, gurps.NotesExt, gurps.TemplatesExt, gurps.LootExt} {
		content := extractContentForCache(filepath.Join(dir, "saved_calc"+ext))
		c.Contains(content, "st 42", ext)
		c.NotContains(content, "<script>", ext)
		c.Contains(content, "plain note", ext)
		if ext == gurps.SheetExt || ext == gurps.TemplatesExt {
			c.Contains(content, "magery 5", ext)
			c.NotContains(content, "magery 3", ext)
		}
	}
	content := extractContentForCache(filepath.Join(dir, "saved_calc"+gurps.SheetExt))
	c.Contains(content, "saved calc")
	c.Contains(content, "combat reflexes")
	c.Contains(content, "dormant 2")
}

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
	var result string
	var ok, returned bool
	c.True(screen.Post(func() {
		result, ok = promptForFileSystemName("taken", "New Name", "taken", targetPath)
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
	c.Equal(filepath.Join(dir, "fresh"), result, "the result handed back is the trimmed name's target")

	// A prompt with no current name has only the entry, and canceling it hands back nothing.
	returned = false
	c.True(screen.Post(func() {
		result, ok = promptForFileSystemName("", "Folder Name", "", targetPath)
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
	c.Equal("", result)
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

// TestNavigatorContextMenuHeadless verifies that shift+F10 opens the library navigator's menu for the selected file,
// and that it opens nothing while nothing is selected.
func TestNavigatorContextMenuHeadless(t *testing.T) {
	c := check.New(t)
	// Every library row starts out disclosed, so that the file put in the user library below is a row of its own.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	user := gurps.GlobalSettings().Libraries.User()
	c.NoError(os.MkdirAll(user.Path(false), 0o750))
	sheetPath := filepath.Join(user.Path(false), "test"+gurps.SheetExt)
	c.NoError(gurps.NewEntity().Save(sheetPath))

	var n *Navigator
	var selected []string
	var focused bool
	screen.Do(func() { n = Workspace.Navigator })
	reloadNavigator(t, screen, n)
	screen.Do(func() {
		n.ApplySelectedPaths([]string{sheetPath})
		n.InitialFocus()
		selected = n.SelectedPaths()
		focused = wnd.Focus() == n.table.AsPanel()
	})
	c.Equal([]string{sheetPath}, selected, "the sheet in the user library must be selected")
	c.True(focused, "the navigator's list must have the focus")

	screen.KeyPress(unison.KeyF10, mod.Shift)
	titles := contextMenuTitles(t, screen, wnd)
	showOnDisk := i18n.Text("Show on Disk")
	c.True(slices.Contains(titles, showOnDisk), "shift+F10 must open the menu for the selected file; it offers %v",
		titles)
	c.True(slices.Contains(titles, cloneSheetAction.Title), "the menu for a sheet must offer %q; it offers %v",
		cloneSheetAction.Title, titles)
	closeContextMenu(c, screen, wnd)

	screen.Do(func() {
		n.table.ClearSelection()
		focused = wnd.Focus() == n.table.AsPanel()
	})
	c.True(focused, "the navigator's list must still have the focus for the chord to be asking it")
	screen.KeyPress(unison.KeyF10, mod.Shift)
	var open bool
	screen.Do(func() { open = uxtest.OpenMenuPopup(wnd) != nil })
	c.False(open, "shift+F10 must open nothing while nothing is selected")
}
