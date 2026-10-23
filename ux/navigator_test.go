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
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
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
	unison.NewPanel().AddChild(n.search.field)
}

// matchCount returns how many of the rows, and those beneath them, match the lower-cased text.
func matchCount(n *Navigator, text string, rows ...*NavigatorNode) int {
	var found []*searchMatch
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
	libNode := NewLibraryNode(n, lib)
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
// the file rather than match on its previous contents.
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
// ancestors from the root down as its title, the icon of its kind, and a key that follows the row's path, so that the
// dropdown can keep its highlight across a Reload, which rebuilds the tree from fresh node objects. A favorited file
// appears both under Favorites and under its library, and the two copies must have different keys.
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
	fresh := NewFileNode(lib, name, sub)
	c.True(underLibrary.ID() != fresh.ID(), "sanity: a rebuilt tree's nodes have fresh IDs")

	match := n.newSearchMatch(underLibrary)
	c.Equal("Conan", match.name, "a file is named without its extension, as in the tree")
	c.Equal("Test / sub", match.title, "the title names the ancestors from the root down")
	c.Equal(match.key, n.newSearchMatch(fresh).key, "the key follows the path to a fresh node")
	c.NotEqual(match.key, n.newSearchMatch(underFavorites).key, "the favorites copy is told apart")
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

// TestDeepSearchCachesFileWhoseParsePanics verifies that the inline cache-miss path a keystroke takes survives a file
// whose parse panics, and that the panic is cached like a parse error, so the file isn't re-read — and doesn't
// re-panic — on every subsequent keystroke.
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
	n.search.suggestions.SetRows(n.search.findMatches("conan")) // As typing with the focus would
	c.Equal(1, n.search.suggestions.List().Count(), "the sheet must match by content while its type is deep-searched")

	// Unchecking the last deep-search type leaves nothing to build; the prewarm itself must clear the stale matches.
	n.deepSearch = make(map[string]bool)
	n.prewarmContentCache()
	c.True(n.contentCache == nil, "the cache must be dropped")
	c.Equal(0, n.search.suggestions.List().Count(), "the deep-search matches must be gone immediately")
	c.Equal(0, len(n.search.findMatches("conan")), "and not be found again")
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
	c.False(toggleFavorites([]*NavigatorNode{NewLibraryNode(nil, lib1)}))
	c.Equal([]string{relPath}, lib1.Favorites())
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
