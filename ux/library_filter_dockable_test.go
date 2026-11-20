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
	"bytes"
	"os"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// openTestLibraryFilter has the Library Explorer list the libraries afresh, then opens a library filter view on the set
// of skills, which is unsaved unless it is nil, and returns it once its first rows are shown. The view is closed when
// the test ends, once its loads have finished.
func openTestLibraryFilter(t *testing.T, screen *unison.HeadlessScreen, set *gurps.LibrarySourceSet) *LibraryFilterDockable {
	t.Helper()
	isolateLibraryFilterSettings(t)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	reloadNavigator(t, screen, n)
	var d *LibraryFilterDockable
	screen.Do(func() { d = openLibraryFilter(librarySkillsKey, set, set != nil) })
	t.Cleanup(func() {
		screen.Do(func() { d.AttemptClose() })
		d.waitForLoads()
	})
	awaitLibraryFilter(t, screen, "the view's first rows", func() bool { return d.content.hasInstalled() })
	return d
}

// awaitLibraryFilter returns once done, which is asked on the UI thread, reports true, running what the loads left for
// the UI thread in between, and fails the test with what it was waiting for when that takes more than ten seconds.
func awaitLibraryFilter(t *testing.T, screen *unison.HeadlessScreen, what string, done func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for {
		screen.Sync()
		var ok bool
		screen.Do(func() { ok = done() })
		if ok {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("gave up waiting for %s", what)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// TestLibraryFilterIsReadOnly verifies that the view refuses every change to its rows and says so in a notice beneath
// its toolbar that names the command for changing an item, that an editor opened from it only shows its row, and that
// sorting by a column, which only reorders the view's own copies of the rows, is left on and leaves the file on disk
// alone.
func TestLibraryFilterIsReadOnly(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	file := writeLibraryFilterTestSkills(c, gurps.GlobalSettings().Libraries.User(), "Skills/Read Only.skl",
		newTestSkill("Bravo", fxp.One, nil), newTestSkill("Alpha", fxp.One, nil))
	before, err := os.ReadFile(file.fullPath)
	c.NoError(err)
	d := openTestLibraryFilter(t, screen, nil)
	content := skillLibraryFilterContent(t, d)

	screen.Do(func() {
		notice, ok := content.Children()[1].Self.(*unison.Label)
		c.True(ok, "a notice sits between the toolbar and the table")
		if ok {
			c.True(strings.Contains(notice.String(), showInSourceFileAction.Title),
				"the notice names the command for changing an item: %q", notice.String())
		}
		content.table.SelectByIndex(0)
		for _, cmd := range []struct {
			name string
			id   int
		}{
			{"Delete", unison.DeleteItemID},
			{"Duplicate", DuplicateItemID},
			{"New Skill", NewSkillItemID},
			{"New Skill Container", NewSkillContainerItemID},
			{"Sync with Source", SyncWithSourceItemID},
		} {
			c.False(content.table.CanPerformCmd(nil, cmd.id), "%s must not be offered", cmd.name)
		}
		content.table.PerformCmd(nil, OpenEditorItemID)
	})
	e := uxtest.SoleEditor[*editor[*gurps.Skill, *gurps.SkillEditData]](t, screen,
		func(unison.Dockable) bool { return true })
	screen.Do(func() { c.True(e.readOnly, "an editor opened from the view only shows its row") })

	screen.Do(func() {
		c.Equal([]string{"Bravo", "Alpha"}, shownSkillNames(content.table), "the rows start in file order")
		content.header.SortOn(content.header.ColumnHeaders[content.table.ColumnIndexForID(gurps.SkillDescriptionColumn)])
		content.header.ApplySort()
		c.Equal([]string{"Alpha", "Bravo"}, shownSkillNames(content.table), "sorting by name reorders the view")
	})
	after, err := os.ReadFile(file.fullPath)
	c.NoError(err)
	c.True(bytes.Equal(before, after), "sorting leaves the file on disk as it was")
}

// TestLibraryFilterReloadsChangedFilesOnly verifies that a view told one of its files changed reads that file again,
// but not the others, which it still has, and applies the quick filter to the new rows; that a file added to a folder
// its set covers joins the view once the Library Explorer has listed it; that a file changed while nothing watched it,
// as a library update changes its files, is read again, alone, once the Library Explorer has listed the same files
// anew; and that a listing that changes nothing leaves the rows shown as they are.
func TestLibraryFilterReloadsChangedFilesOnly(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &libraryFilterReloadDelay, time.Millisecond)
	user := gurps.GlobalSettings().Libraries.User()
	first := writeLibraryFilterTestSkills(c, user, "Skills/A.skl", newTestSkill("Apple", fxp.One, nil))
	writeLibraryFilterTestSkills(c, user, "Skills/B.skl", newTestSkill("Banana", fxp.One, nil))
	d := openTestLibraryFilter(t, screen, &gurps.LibrarySourceSet{
		Entries: []gurps.LibraryFile{{Library: user.Key(), Path: "Skills"}},
	})
	content := skillLibraryFilterContent(t, d)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The test reports the changes itself, so that the watches don't report them a second time.
	stopNavigatorWatches(t, screen, n)

	var lock sync.Mutex
	var read []string
	screen.Do(func() {
		parse := content.parse
		content.parse = func(file libraryFilterFile) ([]*gurps.Skill, error) {
			lock.Lock()
			read = append(read, file.Path)
			lock.Unlock()
			return parse(file)
		}
		content.filterField.SetText("an")
	})
	readNow := func() []string {
		lock.Lock()
		defer lock.Unlock()
		return slices.Clone(read)
	}
	screen.Do(func() { c.Equal([]string{"Banana"}, shownSkillNames(content.table), "the quick filter is in force") })

	writeLibraryFilterTestSkills(c, user, first.Path, newTestSkill("Mango Tree", fxp.One, nil),
		newTestSkill("Cherry", fxp.One, nil))
	screen.Do(func() { libraryFilterPathChanged(first.fullPath) })
	awaitLibraryFilter(t, screen, "the changed file's rows", func() bool {
		return slices.Equal(shownSkillNames(content.table), []string{"Mango Tree", "Banana"})
	})
	c.Equal([]string{first.Path}, readNow(), "only the changed file is read again")
	screen.Do(func() { c.True(content.table.IsFiltered(), "the quick filter stays in force") })

	third := writeLibraryFilterTestSkills(c, user, "Skills/C.skl", newTestSkill("Cantaloupe", fxp.One, nil))
	reloadNavigator(t, screen, n)
	awaitLibraryFilter(t, screen, "the added file's rows", func() bool {
		return slices.Equal(shownSkillNames(content.table), []string{"Mango Tree", "Banana", "Cantaloupe"})
	})

	// Listing the libraries afresh would watch them again, and the watch may yet report a write made before it began, so
	// the views are told of a new listing of the same files directly.
	stopNavigatorWatches(t, screen, n)
	second := writeLibraryFilterTestSkills(c, user, "Skills/B.skl", newTestSkill("Plantain", fxp.One, nil))
	screen.Do(libraryFilterListingChanged)
	awaitLibraryFilter(t, screen, "the rows of the file changed while unwatched", func() bool {
		return slices.Equal(shownSkillNames(content.table), []string{"Mango Tree", "Plantain", "Cantaloupe"})
	})
	c.Equal([]string{first.Path, third.Path, second.Path}, readNow(), "only the changed file is read again")

	var before []*Node[*gurps.Skill]
	screen.Do(func() { before = slices.Clone(content.table.RootRows()) })
	screen.Do(libraryFilterListingChanged)
	d.waitForLoads()
	screen.Sync()
	screen.Do(func() {
		c.True(slices.Equal(before, content.table.RootRows()), "a listing that changes nothing leaves the rows shown")
	})
}

// TestLibraryFilterDiscardsSupersededLoads verifies that a load that finishes after another set was chosen is
// discarded, rather than showing its rows until those of the set chosen since are read.
func TestLibraryFilterDiscardsSupersededLoads(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()
	first := writeLibraryFilterTestSkills(c, user, "Skills/First.skl", newTestSkill("First Skill", fxp.One, nil))
	second := writeLibraryFilterTestSkills(c, user, "Skills/Second.skl", newTestSkill("Second Skill", fxp.One, nil))
	d := openTestLibraryFilter(t, screen, &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{first.LibraryFile}})
	content := skillLibraryFilterContent(t, d)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The test has the file read again itself, so that the watches don't have it read a second time.
	stopNavigatorWatches(t, screen, n)
	// What the superseded load reads is told apart from what the view shows.
	writeLibraryFilterTestSkills(c, user, first.Path, newTestSkill("First Skill Changed", fxp.One, nil))

	// The first read of each file is reported to started, then held back until the file is released.
	hold := map[gurps.LibraryFile]chan struct{}{
		first.LibraryFile:  make(chan struct{}),
		second.LibraryFile: make(chan struct{}),
	}
	released := make(map[gurps.LibraryFile]bool)
	release := func(f gurps.LibraryFile) {
		if !released[f] {
			released[f] = true
			close(hold[f])
		}
	}
	t.Cleanup(func() { // Runs ahead of the view's cleanup, which waits for the loads.
		release(first.LibraryFile)
		release(second.LibraryFile)
	})
	started := make(chan gurps.LibraryFile, len(hold))
	awaitStart := func(want gurps.LibraryFile, what string) {
		select {
		case got := <-started:
			c.Equal(want, got, what)
		case <-time.After(10 * time.Second):
			t.Fatalf("gave up waiting for %s", what)
		}
	}
	var read sync.Map
	screen.Do(func() {
		parse := content.parse
		content.parse = func(file libraryFilterFile) ([]*gurps.Skill, error) {
			if _, again := read.LoadOrStore(file.LibraryFile, true); !again {
				started <- file.LibraryFile
				<-hold[file.LibraryFile]
			}
			return parse(file)
		}
		d.reload()
	})
	awaitStart(first.LibraryFile, "the reload to read the first set's file")
	screen.Do(func() {
		d.chooseSet(librarySkillsKey, &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{second.LibraryFile}}, true)
	})
	release(first.LibraryFile)
	// The superseded load has posted what it read to the UI thread by the time the next one reads its file.
	awaitStart(second.LibraryFile, "the second set's file to be read")
	screen.Sync()
	screen.Do(func() {
		c.Equal([]string{"First Skill"}, shownSkillNames(content.table), "the superseded load is discarded")
	})

	release(second.LibraryFile)
	awaitLibraryFilter(t, screen, "the second set's rows", func() bool {
		return slices.Equal(shownSkillNames(content.table), []string{"Second Skill"})
	})
}

// TestLastLibraryFilterSet verifies that a new view starts on the list type last used and on the saved set last used
// for it, found by name without regard to case, or on every file of the type when no set was recorded for it or the one
// recorded no longer exists.
func TestLastLibraryFilterSet(t *testing.T) {
	c := check.New(t)
	isolateLibraryFilterSettings(t)
	settings := gurps.GlobalSettings()
	spellsKey := gurps.ListFilterKeyForExtension(gurps.SpellsExt)
	mine := &gurps.LibrarySourceSet{Name: "Mine"}
	settings.AddLibrarySourceSet(spellsKey, mine)
	settings.AddLibrarySourceSet(librarySkillsKey, &gurps.LibrarySourceSet{Name: "Theirs"})
	for _, one := range []struct {
		name     string
		lastSets map[string]string
		want     *gurps.LibrarySourceSet
	}{
		{name: "the set last used", lastSets: map[string]string{spellsKey: "mine"}, want: mine},
		{name: "no set recorded"},
		{name: "a set that no longer exists", lastSets: map[string]string{spellsKey: "Gone"}},
		{name: "a set recorded for another type", lastSets: map[string]string{librarySkillsKey: "Theirs"}},
	} {
		settings.LibraryFilter.LastKey = spellsKey
		settings.LibraryFilter.LastSets = one.lastSets
		key, set := lastLibraryFilterSet()
		c.Equal(spellsKey, key, "%s: the type last used", one.name)
		c.True(set == one.want, "%s: the set", one.name)
	}
}

// TestLibraryFilterTypeSwitchKeepsOnlyQuickFilter verifies that a view switched to a set of another list type carries
// the quick filter's text over, while the chosen tags and saved filter, which belong to the type it showed, are
// dropped.
func TestLibraryFilterTypeSwitchKeepsOnlyQuickFilter(t *testing.T) {
	c := check.New(t)
	d, skills := newTestLibraryFilter(t)
	skills.setQuickFilterText("fire")
	skills.chooseFilter(gurps.NewListFilter("Mine"))
	skills.tagFilter.choose([]string{"combat"})
	d.chooseSet(gurps.ListFilterKeyForExtension(gurps.SpellsExt), nil, false)
	spells, ok := d.content.(*libraryFilterContent[*gurps.Spell])
	if !ok {
		t.Fatalf("expected the content of a view on spells, not a %T", d.content)
	}
	c.Equal("fire", spells.quickFilterText(), "the quick filter's text is carried over")
	c.Nil(spells.selectedFilter, "the saved filter is dropped")
	c.Equal(0, len(spells.tagFilter.selected), "the chosen tags are dropped")
}

// TestLibraryFilterRekeysUnsavedSet verifies that a view's unsaved set follows a library to its new key, with its
// entries normalized afterward, since the new key may have been there already.
func TestLibraryFilterRekeysUnsavedSet(t *testing.T) {
	c := check.New(t)
	d, _ := newTestLibraryFilter(t)
	d.chooseSet(librarySkillsKey, &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{
		{Library: "new", Path: "Skills/A.skl"},
		{Library: "old", Path: "Skills"},
		{Library: "other"},
	}}, true)
	d.rekeyUnsavedSet("old", "new")
	c.Equal([]gurps.LibraryFile{{Library: "new", Path: "Skills"}, {Library: "other"}}, d.set.Entries)
}
