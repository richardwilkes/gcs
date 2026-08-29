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
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
	"github.com/rjeczalik/notify"
)

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
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	entity.Profile.PlayerName = "Robert"
	entity.Profile.Organization = "The Black Dragons"
	fileName := "sheet1" + gurps.SheetExt
	c.NoError(entity.Save(filepath.Join(dir, fileName)))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SheetExt: true},
	}
	node := NewFileNode(lib, fileName, nil)
	for _, one := range []struct {
		name string
		text string
	}{
		{name: "character name", text: "conan"},
		{name: "partial character name", text: "cona"},
		{name: "player name", text: "robert"},
		{name: "organization", text: "the black dragons"},
	} {
		n.searchResult = nil
		n.search(one.text, []*NavigatorNode{node})
		c.Equal(1, len(n.searchResult), one.name)
	}

	// Something that isn't present must still not match.
	n.searchResult = nil
	n.search("belit", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult))
}

// TestDeepSearchCachesMarkdownContent verifies what a search stores in the content cache for a markdown file: the raw
// text, lowercased and with surrounding whitespace trimmed at load time, keyed by the file's full path on disk, and
// that subsequent searches consult that entry rather than the file.
func TestDeepSearchCachesMarkdownContent(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes() // The deep search resolves the file's extension through the file type registry.
	ext := uti.Markdown.Extensions[0]
	dir := t.TempDir()
	fileName := "notes" + ext
	p := filepath.Join(dir, fileName)
	c.NoError(os.WriteFile(p, []byte("# Conan The Barbarian\n"), 0o600))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{ext: true},
	}
	node := NewFileNode(lib, fileName, nil)
	n.search("barbarian", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))

	// The lowercased content must now be cached under the file's full path.
	cached, ok := n.contentCache[p]
	c.True(ok, "expected the markdown content to be cached for %q", p)
	c.Equal("# conan the barbarian", cached.content)

	// Removing the file must not change the outcome of a subsequent search, proving the cache is what gets consulted.
	c.NoError(os.Remove(p))
	n.searchResult = nil
	n.search("barbarian", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))
}

// TestDeepSearchServesCacheHitsUntilInvalidated verifies that a search serves a cache hit without checking the file on
// disk — revalidating every hit would be a stat per deep-searchable file on each keystroke — and that dropping the
// entry, as the filesystem watches do when they report the file changed, makes the next search re-read the file rather
// than match on its previous contents.
func TestDeepSearchServesCacheHitsUntilInvalidated(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes() // The deep search resolves the file's extension through the file type registry.
	ext := uti.Markdown.Extensions[0]
	dir := t.TempDir()
	fileName := "notes" + ext
	p := filepath.Join(dir, fileName)
	c.NoError(os.WriteFile(p, []byte("alpha\n"), 0o600))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{ext: true},
	}
	node := NewFileNode(lib, fileName, nil)
	n.search("alpha", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult), "the original content must match and populate the cache")

	// Rewrite the file with different content and a clearly different modification time, so that a search which did
	// check the file against the entry would notice the change.
	c.NoError(os.WriteFile(p, []byte("bravo bravo\n"), 0o600))
	c.NoError(os.Chtimes(p, time.Time{}, time.Now().Add(time.Hour)))

	n.searchResult = nil
	n.search("alpha", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult), "a cache hit must be served without checking the file on disk")
	n.searchResult = nil
	n.search("bravo", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult), "a cache hit must be served without checking the file on disk")

	// Once the watch reports the change, the next search must re-read the file.
	n.watchCallback(lib, p, notify.Write)
	_, ok := n.contentCache[p]
	c.False(ok, "the reported change must drop the entry")
	n.searchResult = nil
	n.search("bravo", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult), "the new content must be found after the change is reported")
	n.searchResult = nil
	n.search("alpha", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult), "the previous content must no longer match")
	c.True(n.contentCache[p].isCurrent(p), "the re-read must capture the file's current metadata")
}

// TestInvalidationReachesInFlightCacheBuild verifies that a file change reported while a background cache build is
// running is applied to that build's result, since the build may have read the file before it changed: the entry is
// dropped from the result when it is installed, a superseded build leaves the invalidation for the newer build to
// apply, and starting a build consumes the invalidations made before it, which its starting snapshot already lacks.
func TestInvalidationReachesInFlightCacheBuild(t *testing.T) {
	c := check.New(t)
	n := &Navigator{contentCache: map[string]*contentCacheEntry{"a": {}, "b": {}}}
	gen := n.cacheGeneration.Add(1) // A build is in flight from here

	n.invalidateContentCacheEntry("a")
	_, ok := n.contentCache["a"]
	c.False(ok, "the live cache must drop the entry at once")
	c.NotNil(n.contentCache["b"], "other entries must be untouched")

	fresh := map[string]*contentCacheEntry{"a": {}, "b": {}}
	c.True(n.applyPrewarmedContentCache(gen, fresh))
	_, ok = n.contentCache["a"]
	c.False(ok, "the installed build must lack the entry invalidated while it ran")
	c.True(fresh["b"] == n.contentCache["b"], "the installed build must keep the rest")
	c.Equal(0, len(n.invalidatedPaths), "an installed build must consume the invalidations")

	// A superseded build must leave the invalidation for the newer build's result.
	n.invalidateContentCacheEntry("b")
	n.cacheGeneration.Add(1)
	c.False(n.applyPrewarmedContentCache(gen, map[string]*contentCacheEntry{"b": {}}))
	c.True(n.invalidatedPaths["b"], "a discarded build must not consume the invalidations")
	c.True(n.applyPrewarmedContentCache(n.cacheGeneration.Load(), map[string]*contentCacheEntry{"b": {}}))
	_, ok = n.contentCache["b"]
	c.False(ok, "the newer build's result must lack the entry invalidated while the older one ran")

	// Starting a build consumes the invalidations made before it, since it starts from a cache that already lacks
	// those entries. The table is real but empty, so the prewarm runs without spawning a background build.
	n.table = unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{})
	n.invalidateContentCacheEntry("c")
	n.prewarmContentCache()
	c.Equal(0, len(n.invalidatedPaths), "a build's start must consume the invalidations made before it")
}

// TestRerunSearchPreservesPosition verifies that re-running an active search keeps the user's place in the match list.
// A completed background cache prewarm re-runs the search to correct incomplete results, and must not yank a user who
// has walked to match 5 of 12 back to the start; the position only resets when the row they were on no longer matches.
func TestRerunSearchPreservesPosition(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	nodes := make([]*NavigatorNode, 3)
	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	for i := range nodes {
		fileName := "sheet" + string(rune('1'+i)) + gurps.SheetExt
		c.NoError(entity.Save(filepath.Join(dir, fileName)))
		nodes[i] = NewFileNode(lib, fileName, nil)
	}
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SheetExt: true},
	}
	n.search("conan", nodes)
	c.Equal(3, len(n.searchResult))

	// With the user on the last match, a re-run over the same rows must keep them there.
	n.searchIndex = 2
	current := n.searchResult[2]
	n.rerunSearch("conan", nodes)
	c.Equal(3, len(n.searchResult))
	c.Equal(2, n.searchIndex)
	c.True(n.searchResult[2] == current, "the position must still refer to the same row")

	// When earlier matches disappear, the position must follow the row to its new index.
	n.rerunSearch("conan", nodes[1:])
	c.Equal(2, len(n.searchResult))
	c.Equal(1, n.searchIndex)
	c.True(n.searchResult[1] == current, "the position must follow the row the user was on")

	// When the row the user was on no longer matches, the position must reset to "no position".
	n.rerunSearch("conan", nodes[:2])
	c.Equal(2, len(n.searchResult))
	c.Equal(-1, n.searchIndex)

	// A re-run without an established position must leave it unestablished.
	n.rerunSearch("conan", nodes)
	c.Equal(-1, n.searchIndex)

	// A Reload rebuilds the tree with fresh node objects, so pointer identity always fails; the position must follow
	// the file to its node in the new tree by path.
	n.searchIndex = 1
	current = n.searchResult[1]
	fresh := make([]*NavigatorNode, len(nodes))
	for i := range fresh {
		fresh[i] = NewFileNode(lib, "sheet"+string(rune('1'+i))+gurps.SheetExt, nil)
	}
	n.rerunSearch("conan", fresh)
	c.Equal(3, len(n.searchResult))
	c.Equal(1, n.searchIndex, "the position must survive a tree rebuild by following the file's path")
	c.True(n.searchResult[1] != current, "sanity: the rebuilt tree must consist of fresh node objects")
	c.Equal(current.Path(), n.searchResult[1].Path())
}

// TestUpdateMatchControlsLeavesSelectionAlone verifies the split between the two search toolbar refreshes: the
// toolbar-only refresh used when a completed background cache prewarm re-runs the active search must not touch the
// table selection, since the prewarm finishes asynchronously and re-selecting the match would yank the tree away from
// whatever the user has clicked on since; the user-driven adjustForMatch must still select the current match.
func TestUpdateMatchControlsLeavesSelectionAlone(t *testing.T) {
	c := check.New(t)
	lib := gurps.NewLibrary("Test", "test", "", "test", t.TempDir())
	rows := []*NavigatorNode{
		NewFileNode(lib, "one"+gurps.NotesExt, nil),
		NewFileNode(lib, "two"+gurps.NotesExt, nil),
		NewFileNode(lib, "three"+gurps.NotesExt, nil),
	}
	n := &Navigator{
		table:         unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{}),
		backButton:    unison.NewButton(),
		forwardButton: unison.NewButton(),
		matchesLabel:  unison.NewLabel(),
		searchIndex:   1,
		searchResult:  rows,
	}
	unison.NewPanel().AddChild(n.matchesLabel) // The label's refresh marks its parent for layout.
	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.table.SetRootRows(rows)

	// The user has clicked away from the match to the first row.
	n.table.SelectByIndex(0)

	// The asynchronous refresh must update the toolbar without moving the selection.
	n.updateMatchControls()
	c.Equal(0, n.table.FirstSelectedRowIndex(), "the toolbar-only refresh must not move the selection")
	c.Equal("2 of 3", n.matchesLabel.String())
	c.True(n.backButton.Enabled())
	c.True(n.forwardButton.Enabled())

	// The user-driven refresh must select the current match.
	n.adjustForMatch()
	c.Equal(1, n.table.FirstSelectedRowIndex(), "adjustForMatch must select the current match")
	c.Equal(1, len(n.table.SelectedRows(false)), "adjustForMatch must replace the selection, not add to it")

	// With no position established, neither refresh may enable stepping back.
	n.searchIndex = -1
	n.updateMatchControls()
	c.Equal("- of 3", n.matchesLabel.String())
	c.False(n.backButton.Enabled())
	c.True(n.forwardButton.Enabled())

	// With no matches at all, the label must reset and both buttons must disable.
	n.searchResult = nil
	n.updateMatchControls()
	c.Equal("-", n.matchesLabel.String())
	c.False(n.backButton.Enabled())
	c.False(n.forwardButton.Enabled())
}

// TestDeepSearchDisabledTypeIgnoresCache verifies that disabling a file type for deep search takes effect immediately,
// even while the content cache still holds entries for that type. The cache is rebuilt asynchronously after a settings
// change, so the search must gate on the deep search setting before consulting the cache, not after.
func TestDeepSearchDisabledTypeIgnoresCache(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	fileName := "sheet1" + gurps.SheetExt
	p := filepath.Join(dir, fileName)
	c.NoError(entity.Save(p))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SheetExt: true},
	}
	node := NewFileNode(lib, fileName, nil)
	n.search("conan", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))
	_, ok := n.contentCache[p]
	c.True(ok, "expected the sheet's content to be cached for %q", p)

	// Disabling the type must stop matches at once, despite the stale cache entry that remains until the rebuild.
	n.deepSearch = make(map[string]bool)
	n.searchResult = nil
	n.search("conan", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult), "a disabled type must not match, even with its entry still cached")

	// Re-enabling must match again, served from the still-present cache entry.
	n.deepSearch = map[string]bool{gurps.SheetExt: true}
	n.searchResult = nil
	n.search("conan", []*NavigatorNode{node})
	c.Equal(1, len(n.searchResult))
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
	RegisterKnownFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	fileName := "panics" + gurps.SkillsExt
	p := filepath.Join(dir, fileName)
	// A null row is valid JSON, so this gets past parsing and panics when the nil skill is traversed.
	c.NoError(os.WriteFile(p, []byte(`{"version":5,"rows":[null]}`), 0o600))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		searchIndex: -1,
		deepSearch:  map[string]bool{gurps.SkillsExt: true},
	}
	node := NewFileNode(lib, fileName, nil)
	n.search("anything", []*NavigatorNode{node})
	c.Equal(0, len(n.searchResult))
	entry, ok := n.contentCache[p]
	c.True(ok, "expected the panicking file to be cached for %q", p)
	c.Equal("", entry.content)
}

// TestApplyPrewarmedContentCache verifies the generation handshake that protects the background cache builds: a build
// that is no longer the newest must be discarded rather than installed, since a newer prewarm or a suspension has made
// its contents suspect, while the newest build replaces the cache wholesale.
func TestApplyPrewarmedContentCache(t *testing.T) {
	c := check.New(t)
	n := &Navigator{}
	gen := n.cacheGeneration.Add(1)
	original := map[string]*contentCacheEntry{"original": {}}
	n.contentCache = original

	// A newer prewarm (or a suspension) started after this build began, so its result must be discarded.
	n.cacheGeneration.Add(1)
	c.False(n.applyPrewarmedContentCache(gen, map[string]*contentCacheEntry{"stale": {}}),
		"a superseded build must not be installed")
	c.True(original["original"] == n.contentCache["original"], "a superseded build must not clobber the cache")

	// The newest build must be installed.
	fresh := map[string]*contentCacheEntry{"fresh": {}}
	c.True(n.applyPrewarmedContentCache(n.cacheGeneration.Load(), fresh))
	c.True(fresh["fresh"] == n.contentCache["fresh"], "the newest build must replace the cache")
}

// TestSupersededCacheBuildHandsEntriesToItsSuccessor verifies that a background cache build inherits the entries of the
// build it supersedes rather than starting over from the live cache alone, which is what keeps the back-to-back
// reloads at startup from restarting the first full parse of a library from zero. The hand-off rules are checked
// against a synthetic prior build, whose contents are known: its entries fill in what the live cache lacks, the live
// cache's own entry wins where both hold one, and its entry for a file whose change was reported after it began is not
// inherited, since it may have read the old contents. A real pair of builds then confirms the wiring: the second
// waits for the first to stop and reuses whatever it completed. Neither build's completion task runs, since nothing
// here pumps unison's task queue, so the results are read from the build records directly.
func TestSupersededCacheBuildHandsEntriesToItsSuccessor(t *testing.T) {
	c := check.New(t)
	RegisterKnownFileTypes() // The path collection resolves each file's extension through the file type registry.
	ext := uti.Markdown.Extensions[0]
	dir := t.TempDir()
	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	names := []string{"a" + ext, "b" + ext, "c" + ext}
	paths := make([]string, len(names))
	nodes := make([]*NavigatorNode, len(names))
	for i, name := range names {
		paths[i] = filepath.Join(dir, name)
		c.NoError(os.WriteFile(paths[i], []byte("content "+name+"\n"), 0o600))
		nodes[i] = NewFileNode(lib, name, nil)
	}
	n := &Navigator{
		table:       unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{}),
		searchIndex: -1,
		deepSearch:  map[string]bool{ext: true},
	}
	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.table.SetRootRows(nodes)

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
	RegisterKnownFileTypes() // The path collection resolves each file's extension through the file type registry.
	ext := uti.Markdown.Extensions[0]
	dir := t.TempDir()
	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{deepSearch: map[string]bool{ext: true}}
	shared := "notes" + ext
	nested := filepath.Join("sub", "extra"+ext)
	// The container nodes are built by hand rather than through the constructors, which read the filesystem and the
	// global settings, and their IDs are minted directly rather than through gurps.IDForNavNode, which records every
	// path it is asked about in the global settings; only the shape of the tree matters here.
	favorites := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorFavorites)}
	favorites.children = []*NavigatorNode{NewFileNode(lib, shared, favorites)}
	library := &NavigatorNode{id: tid.MustNewTID(kinds.NavigatorLibrary), library: lib}
	sub := &NavigatorNode{
		id:      tid.MustNewTID(kinds.NavigatorDirectory),
		path:    "sub",
		library: lib,
		parent:  library,
	}
	sub.children = []*NavigatorNode{
		NewFileNode(lib, nested, sub),
		NewFileNode(lib, filepath.Join("sub", "excluded.txt"), sub),
	}
	library.children = []*NavigatorNode{
		sub,
		NewFileNode(lib, shared, library), // The same file the favorites node holds
	}
	paths := make(map[string]bool)
	n.collectDeepSearchPaths([]*NavigatorNode{favorites, library}, paths)
	c.Equal(2, len(paths))
	c.True(paths[filepath.Join(dir, shared)], "the file favorites repeats must be collected exactly once")
	c.True(paths[filepath.Join(dir, nested)], "the file nested in a subdirectory must be reached by recursion")
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
	n := &Navigator{table: unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{})}
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

// TestPrewarmWithNoDeepSearchPathsRerunsSearch verifies the completion path taken when the collected path set is
// empty, as happens when the user unchecks the last deep-search type: no background build runs, so the prewarm itself
// must drop the cache and re-run the active search, clearing the now-invalid deep-search matches immediately rather
// than leaving them — and the match position and label — pointing at stale hits until the next keystroke.
func TestPrewarmWithNoDeepSearchPathsRerunsSearch(t *testing.T) {
	c := check.New(t)
	RegisterGCSFileTypes() // The deep search resolves the file's extension through the file type registry.
	dir := t.TempDir()
	entity := gurps.NewEntity()
	entity.Profile.Name = "Conan"
	fileName := "sheet1" + gurps.SheetExt
	c.NoError(entity.Save(filepath.Join(dir, fileName)))

	lib := gurps.NewLibrary("Test", "test", "", "test", dir)
	n := &Navigator{
		table:         unison.NewTable(&unison.SimpleTableModel[*NavigatorNode]{}),
		searchField:   unison.NewField(),
		backButton:    unison.NewButton(),
		forwardButton: unison.NewButton(),
		matchesLabel:  unison.NewLabel(),
		searchIndex:   -1,
		deepSearch:    map[string]bool{gurps.SheetExt: true},
	}
	unison.NewPanel().AddChild(n.matchesLabel) // The label's refresh marks its parent for layout.
	n.table.Columns = make([]unison.ColumnInfo, 1)
	n.table.SetRootRows([]*NavigatorNode{NewFileNode(lib, fileName, nil)})
	n.searchField.SetText("conan")

	n.search("conan", n.table.RootRows())
	c.Equal(1, len(n.searchResult), "the sheet must match by content while its type is deep-searched")
	n.searchIndex = 0

	// Unchecking the last deep-search type leaves nothing to build; the prewarm itself must clear the stale results.
	n.deepSearch = make(map[string]bool)
	n.prewarmContentCache()
	c.True(n.contentCache == nil, "the cache must be dropped")
	c.Equal(0, len(n.searchResult), "the deep-search matches must be gone immediately")
	c.Equal(-1, n.searchIndex, "the match position must no longer point at a stale hit")
	c.Equal("-", n.matchesLabel.String())
	c.False(n.backButton.Enabled())
	c.False(n.forwardButton.Enabled())
}

// TestToggleFavoritesKeysOnFullPath verifies that the de-duplication done while toggling favorites is keyed on the full
// path on disk rather than the library-relative path. Two libraries can each hold the same relative path, and favorites
// are tracked per library, so both must be toggled.
func TestToggleFavoritesKeysOnFullPath(t *testing.T) {
	c := check.New(t)
	relPath := filepath.Join("Notes", "foo.not")
	lib1 := gurps.NewLibrary("One", "one", "", "one", t.TempDir())
	lib2 := gurps.NewLibrary("Two", "two", "", "two", t.TempDir())
	rows := []*NavigatorNode{
		NewFileNode(lib1, relPath, nil),
		NewFileNode(lib2, relPath, nil),
	}
	c.True(toggleFavorites(rows))
	c.Equal([]string{relPath}, lib1.Favorites())
	c.Equal([]string{relPath}, lib2.Favorites())

	// A repeated reference to the same file within one call must only toggle once, leaving it a favorite rather than
	// toggling it back off.
	lib3 := gurps.NewLibrary("Three", "three", "", "three", t.TempDir())
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

// TestNewFolderRejectsCollisionWithWhitespacePaddedName verifies that the existence check the new-folder dialog makes
// and the directory it would create refer to the same path, so padding a name with whitespace can't slip a duplicate
// past validation.
func TestNewFolderRejectsCollisionWithWhitespacePaddedName(t *testing.T) {
	c := check.New(t)
	dir := t.TempDir()
	c.NoError(os.Mkdir(filepath.Join(dir, "Notes"), 0o750))

	// The name is padded, but the collision check must still find the existing directory.
	p := newFolderPath(dir, "  Notes  ")
	_, err := os.Stat(p)
	c.NoError(err, "the padded name must resolve to the existing directory")

	// And creating it must be what fails, rather than quietly producing a second, differently-named directory.
	c.HasError(os.Mkdir(p, 0o750))
	entries, err := os.ReadDir(dir)
	c.NoError(err)
	c.Equal(1, len(entries))
	c.Equal("Notes", entries[0].Name())
}
