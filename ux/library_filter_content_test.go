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
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/library"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// librarySkillsKey is the key of the skill list type.
var librarySkillsKey = gurps.ListFilterKeyForExtension(gurps.SkillsExt)

// isolateLibraryFilterSettings points the global settings at a file in the test's directory and puts back the library
// filter settings and the column widths, which a view records, when the test ends.
func isolateLibraryFilterSettings(t *testing.T) {
	t.Helper()
	uxtest.SwapForTest(t, &gurps.SettingsPath, filepath.Join(t.TempDir(), "settings.json"))
	global := gurps.GlobalSettings()
	uxtest.SwapForTest(t, &global.LibraryFilter, gurps.LibraryFilterSettings{LastSets: make(map[string]string)})
	widths := maps.Clone(global.ColumnSizing)
	if widths == nil {
		widths = make(map[string]map[int]float32)
	}
	uxtest.SwapForTest(t, &global.ColumnSizing, widths)
}

// newTestLibraryFilter returns an unplaced library filter view on every skill. It has no listing of the libraries to
// resolve against, so nothing is loaded until the test installs files itself (see installLibraryFilterFiles).
func newTestLibraryFilter(t *testing.T) (*LibraryFilterDockable, *libraryFilterContent[*gurps.Skill]) {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	isolateLibraryFilterSettings(t)
	uxtest.SwapForTest(t, &libraryFilterListing, func() *navigatorListing { return nil })
	d := newLibraryFilterDockable(librarySkillsKey, nil, false)
	return d, skillLibraryFilterContent(t, d)
}

// skillLibraryFilterContent returns the content of a library filter view on skills.
func skillLibraryFilterContent(t *testing.T, d *LibraryFilterDockable) *libraryFilterContent[*gurps.Skill] {
	t.Helper()
	content, ok := d.content.(*libraryFilterContent[*gurps.Skill])
	if !ok {
		t.Fatalf("expected the content of a view on skills, not a %T", d.content)
	}
	return content
}

// writeLibraryFilterTestSkills saves the skills as the list file at the slash-separated path within the library, and
// returns the file as a library filter view knows it.
func writeLibraryFilterTestSkills(c check.Checker, lib *library.Library, relPath string,
	skills ...*gurps.Skill,
) libraryFilterFile {
	fullPath := filepath.Join(lib.Path(false), filepath.FromSlash(relPath))
	c.NoError(os.MkdirAll(filepath.Dir(fullPath), 0o750))
	c.NoError(gurps.SaveSkills(skills, fullPath))
	return libraryFilterFile{Library: lib.Key(), Path: relPath, libTitle: lib.Data().Title, fullPath: fullPath}
}

// installLibraryFilterFiles has the content read the files and show their rows, as a load does, but before returning.
func installLibraryFilterFiles[T gurps.Node[T]](content *libraryFilterContent[T], files ...libraryFilterFile) {
	nameLibraryFilterFiles(files)
	l := content.startLoad()
	parsed := l.read(files, content.parse, func() bool { return false })
	close(l.done)
	content.install(l, files, parsed)
}

// newTestSkillContainer returns a skill container with the name, holding the children.
func newTestSkillContainer(name string, children ...*gurps.Skill) *gurps.Skill {
	container := gurps.NewSkill(nil, nil, true)
	container.Name = name
	container.SetChildren(children)
	return container
}

// libraryFilterRowsByName returns every row of the table, at any depth, disclosed or not, by the name of its skill.
func libraryFilterRowsByName(table *unison.Table[*Node[*gurps.Skill]]) map[string]*Node[*gurps.Skill] {
	rows := make(map[string]*Node[*gurps.Skill])
	var collect func(list []*Node[*gurps.Skill])
	collect = func(list []*Node[*gurps.Skill]) {
		for _, row := range list {
			rows[row.Data().Name] = row
			collect(row.Children())
		}
	}
	collect(table.RootRows())
	return rows
}

// shownSkillNames returns the names of the skills the table shows, in order.
func shownSkillNames(table *unison.Table[*Node[*gurps.Skill]]) []string {
	names := make([]string, 0, table.LastRowIndex()+1)
	for i := range table.LastRowIndex() + 1 {
		names = append(names, table.RowFromIndex(i).Data().Name)
	}
	return names
}

// TestLibraryFilterMergesFilesWithFileColumn verifies that the rows of every file are shown, even where a row of one
// file has the ID of a row of another, which then gets an ID of its own while still knowing the ID it has in its file,
// as it does when a later load gives it yet another; and that the File column, which comes last, names each row's file,
// for a row within a container too, with a tooltip that gives the file's paths as they are, not as Markdown.
func TestLibraryFilterMergesFilesWithFileColumn(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	alpha := newTestSkill("Alpha Skill", fxp.One, nil)
	beta := newTestSkill("Beta Skill", fxp.One, nil)
	beta.TID = alpha.TID
	alphaFile := writeLibraryFilterTestSkills(c, master, "Skills/Alpha.skl", alpha)
	betaFile := writeLibraryFilterTestSkills(c, user, "Skills/Beta.skl", newTestSkillContainer("Beta Group", beta))
	installLibraryFilterFiles(content, alphaFile, betaFile)

	col := content.table.ColumnIndexForID(libraryFileColumnID)
	c.Equal(len(content.table.Columns)-1, col, "the File column comes last")
	rows := libraryFilterRowsByName(content.table)
	for name, file := range map[string]string{"Alpha Skill": "Alpha", "Beta Group": "Beta", "Beta Skill": "Beta"} {
		row, ok := rows[name]
		if !ok {
			t.Errorf("%s must be shown", name)
			continue
		}
		c.Equal(file, row.cellData(col, true).Primary, "the File column of %s", name)
	}
	if t.Failed() {
		return
	}
	alphaRow, betaRow := rows["Alpha Skill"], rows["Beta Skill"]
	c.NotEqual(alphaRow.ID(), betaRow.ID(), "rows of different files never share an ID")
	c.Equal(alpha.TID, content.inFileID(alphaRow.Data()))
	c.Equal(alpha.TID, content.inFileID(betaRow.Data()), "a row given an ID of its own knows the ID it has in its file")
	file, ok := content.fileOf(betaRow.Data())
	c.True(ok, "the row's file is known")
	c.Equal(betaFile.LibraryFile, file.LibraryFile, "the row knows its file")
	var tip *unison.Panel
	if i := content.table.RowToIndex(alphaRow); i != -1 {
		tip = alphaRow.ColumnCell(i, col, unison.Black, unison.White, false, false, false).AsPanel().Tooltip
	}
	c.True(tip != nil && tip.Accessibility.Name == alphaFile.libTitle+": "+alphaFile.Path+"\n"+alphaFile.fullPath,
		"the File column's tooltip gives the file's library, its path there and its path on disk, as written")

	interloper := newTestSkill("Interloper", fxp.One, nil)
	interloper.TID = betaRow.ID()
	installLibraryFilterFiles(content, alphaFile,
		writeLibraryFilterTestSkills(c, master, "Skills/Between.skl", interloper), betaFile)
	betaRow = libraryFilterRowsByName(content.table)["Beta Skill"]
	c.NotEqual(interloper.TID, betaRow.ID(), "the row read before is given another ID")
	c.Equal(alpha.TID, content.inFileID(betaRow.Data()), "and still knows the ID it has in its file")
}

// TestLibraryFilterKeepsRowsUntilLoad verifies that the view hands the table the same rows, at every depth, each time
// it is filtered or sorted, so that what they built for their cells is kept, and new ones once a load replaces what is
// shown, since what the File column shows for a row turns on the other files shown.
func TestLibraryFilterKeepsRowsUntilLoad(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	file := writeLibraryFilterTestSkills(c, user, "Skills/Same.skl",
		newTestSkillContainer("Group", newTestSkill("Bravo", fxp.One, nil), newTestSkill("Alpha", fxp.One, nil)))
	installLibraryFilterFiles(content, file)
	fileCol := content.table.ColumnIndexForID(libraryFileColumnID)
	before := libraryFilterRowsByName(content.table)
	c.Equal("Same", before["Alpha"].cellData(fileCol, true).Primary)

	content.filterField.SetText("alpha")
	c.Equal([]string{"Group", "Alpha"}, shownSkillNames(content.table))
	c.True(content.table.RowFromIndex(0) == before["Group"] && content.table.RowFromIndex(1) == before["Alpha"],
		"the rows the filter shows are those built before it")
	content.filterField.SetText("")
	content.header.SortOn(content.header.ColumnHeaders[content.table.ColumnIndexForID(gurps.SkillDescriptionColumn)])
	content.header.ApplySort()
	c.Equal([]string{"Group", "Alpha", "Bravo"}, shownSkillNames(content.table))
	for name, row := range libraryFilterRowsByName(content.table) {
		c.True(row == before[name], "%s keeps its row once the filter is cleared and the rows are sorted", name)
	}

	installLibraryFilterFiles(content, file,
		writeLibraryFilterTestSkills(c, master, "Skills/Same.skl", newTestSkill("Other", fxp.One, nil)))
	c.Equal(file.libTitle+": Skills/Same", libraryFilterRowsByName(content.table)["Alpha"].cellData(fileCol, true).Primary,
		"a row read before names its file apart from the one of the same name that joined the view")
}

// TestLibraryFilterMeasuresRowsReadAgain verifies that a row of a file read again while the quick filter is in force is
// as tall as what it now holds makes it, though it has the ID the table kept its height by.
func TestLibraryFilterMeasuresRowsReadAgain(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	skill := newTestSkill("Alpha", fxp.One, nil)
	installLibraryFilterFiles(content, writeLibraryFilterTestSkills(c, user, "Skills/Same.skl", skill))
	content.filterField.SetText("alpha")
	c.Equal([]string{"Alpha"}, shownSkillNames(content.table))
	height := content.table.RowHeights()[0]

	skill.LocalNotes = "A note\nof three\nlines"
	installLibraryFilterFiles(content, writeLibraryFilterTestSkills(c, user, "Skills/Same.skl", skill))
	c.Equal([]string{"Alpha"}, shownSkillNames(content.table))
	c.True(content.table.RowHeights()[0] > height, "the row must be as tall as the note it was given makes it")
}

// TestLibraryFilterQuickFilterSkipsFileColumn verifies that the quick filter passes over the File column, which would
// otherwise let a word of a file's name match every row of the file, while still keeping the container of a match, as
// a dimmed row of context.
func TestLibraryFilterQuickFilterSkipsFileColumn(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	installLibraryFilterFiles(content,
		writeLibraryFilterTestSkills(c, user, "Skills/Fire.skl", newTestSkill("Swimming", fxp.One, nil)),
		writeLibraryFilterTestSkills(c, user, "Skills/Circus.skl",
			newTestSkillContainer("Circus Acts", newTestSkill("Fire Eating", fxp.One, nil))))

	content.filterField.SetText("fire")
	c.Equal([]string{"Circus Acts", "Fire Eating"}, shownSkillNames(content.table),
		"only the skill named for fire is shown, beneath its container")
	if container := content.table.RowFromIndex(0); container != nil {
		c.True(container.cellData(0, true).Dim, "the container is shown dimmed, only for context")
	}
}

// TestLibraryFilterReportsUnreadableFiles verifies that a file that can't be read leaves the rows of the others shown,
// and is counted in the footer and named in its tooltip, while a file that is gone by the time it is read, which the
// next listing of the libraries drops, is neither.
func TestLibraryFilterReportsUnreadableFiles(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	good := writeLibraryFilterTestSkills(c, user, "Skills/Good.skl", newTestSkill("Good Skill", fxp.One, nil))
	bad := writeLibraryFilterTestSkills(c, user, "Skills/Bad.skl")
	c.NoError(os.WriteFile(bad.fullPath, []byte("this is not a list"), 0o600))
	gone := writeLibraryFilterTestSkills(c, user, "Skills/Gone.skl")
	c.NoError(os.Remove(gone.fullPath))
	installLibraryFilterFiles(content, bad, gone, good)

	c.Equal([]string{"Good Skill"}, shownSkillNames(content.table), "the readable file's rows are shown")
	c.True(strings.Contains(content.footer.String(), "1 file couldn't be read"),
		"the footer counts the unreadable file alone: %q", content.footer.String())
	var tip []string
	if content.footer.Tooltip != nil {
		tip = uxtest.LabelTexts(content.footer.Tooltip)
	}
	c.True(slices.ContainsFunc(tip, func(text string) bool { return strings.Contains(text, bad.Path) }),
		"the footer's tooltip names the unreadable file: %q", tip)
	c.False(slices.ContainsFunc(tip, func(text string) bool { return strings.Contains(text, gone.Path) }),
		"the footer's tooltip doesn't name the file that is gone: %q", tip)
}

// TestLibraryFilterSelectionStaysWithReIDedRow verifies that a selected row given a new ID, because a file joining the
// view ahead of its own has a row with its ID, stays selected, rather than the selection passing to that other row.
func TestLibraryFilterSelectionStaysWithReIDedRow(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	kept := newTestSkill("Kept", fxp.One, nil)
	keptFile := writeLibraryFilterTestSkills(c, user, "Skills/Kept.skl", kept)
	installLibraryFilterFiles(content, keptFile)
	content.table.SelectAll()

	joined := newTestSkill("Joined", fxp.One, nil)
	joined.TID = kept.TID
	installLibraryFilterFiles(content, writeLibraryFilterTestSkills(c, master, "Skills/Joined.skl", joined), keptFile)
	selected := make([]string, 0, 1)
	for _, row := range content.table.SelectedRows(false) {
		selected = append(selected, row.Data().Name)
	}
	c.Equal([]string{"Kept"}, selected, "the row given a new ID stays selected")
}

// TestLibraryFilterSizesColumnsToEveryRow verifies that the first rows a view shows size its columns, when none were
// sized for its type before, to fit every row, even when a quick filter carried over from another type hides them all.
func TestLibraryFilterSizesColumnsToEveryRow(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	file := writeLibraryFilterTestSkills(c, user, "Skills/Wide.skl",
		newTestSkill("A Skill Whose Name Is Far Wider Than Its Column Header", fxp.One, nil))
	widths := func(filter string) []float32 {
		_, content := newTestLibraryFilter(t)
		c.False(content.savedWidths, "no column widths were saved for the type")
		content.setQuickFilterText(filter)
		installLibraryFilterFiles(content, file)
		list := make([]float32, len(content.table.Columns))
		for i, col := range content.table.Columns {
			list[i] = col.Current
		}
		return list
	}
	c.Equal(widths(""), widths("matches nothing"), "the columns fit the rows the quick filter hides")
}

// TestLibraryFilterReusesWhatSupersededLoadsRead verifies that a load superseded while files change hands the files
// it had read to the load that supersedes it, which doesn't read them again.
func TestLibraryFilterReusesWhatSupersededLoadsRead(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &libraryFilterReloadDelay, time.Millisecond)
	user := gurps.GlobalSettings().Libraries.User()
	first := writeLibraryFilterTestSkills(c, user, "Skills/A.skl", newTestSkill("Apple", fxp.One, nil))
	second := writeLibraryFilterTestSkills(c, user, "Skills/B.skl", newTestSkill("Banana", fxp.One, nil))
	d := openTestLibraryFilter(t, screen, &gurps.LibrarySourceSet{
		Entries: []gurps.LibraryFile{{Library: user.Key(), Path: "Skills"}},
	})
	content := skillLibraryFilterContent(t, d)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The test reports the change itself, so that the watches don't report it a second time.
	stopNavigatorWatches(t, screen, n)

	release := make(chan struct{})
	var releaseOnce sync.Once
	releaseSecond := func() { releaseOnce.Do(func() { close(release) }) }
	t.Cleanup(releaseSecond) // Runs ahead of the view's cleanup, which waits for the loads.
	var lock sync.Mutex
	reads := make(map[string]int)
	readsOf := func(p string) int {
		lock.Lock()
		defer lock.Unlock()
		return reads[p]
	}
	screen.Do(func() {
		parse := content.parse
		content.parse = func(file libraryFilterFile) ([]*gurps.Skill, error) {
			lock.Lock()
			reads[file.Path]++
			hold := file.Path == second.Path && reads[file.Path] == 1
			lock.Unlock()
			if hold {
				<-release
			}
			return parse(file)
		}
		d.reload()
	})
	awaitLibraryFilter(t, screen, "the first load to read both files", func() bool {
		return readsOf(first.Path) == 1 && readsOf(second.Path) == 1
	})
	gen := content.gen.Load()
	screen.Do(func() { libraryFilterPathChanged(first.fullPath) })
	awaitLibraryFilter(t, screen, "the load that supersedes it", func() bool { return content.gen.Load() == gen+1 })
	releaseSecond()
	d.waitForLoads()
	screen.Sync()

	c.Equal(1, readsOf(first.Path), "the file the superseded load read isn't read again")
	screen.Do(func() { c.Equal([]string{"Apple", "Banana"}, shownSkillNames(content.table)) })
}

// TestLibraryFilterCopyToSheetCarriesLibrarySource verifies that a row copied from the view to a sheet records the
// library file it came from and its ID there as its source, as a row copied from that file's own list does, even when
// the view gave it an ID of its own.
func TestLibraryFilterCopyToSheetCarriesLibrarySource(t *testing.T) {
	c := check.New(t)
	master, user := uxtest.UseTestLibraries(t, c)
	_, content := newTestLibraryFilter(t)
	sheet := newTestSheetForTemplate(t)
	alpha := newTestSkill("Alpha Skill", fxp.One, nil)
	beta := newTestSkill("Beta Skill", fxp.One, nil)
	beta.TID = alpha.TID
	alphaFile := writeLibraryFilterTestSkills(c, master, "Skills/Alpha.skl", alpha)
	betaFile := writeLibraryFilterTestSkills(c, user, "Skills/Beta.skl", beta)
	installLibraryFilterFiles(content, alphaFile, betaFile)

	content.table.SelectAll()
	copySelectionTo(content.table, []*Sheet{sheet})
	sources := make(map[string]gurps.Source)
	for _, one := range sheet.Entity().Skills {
		sources[one.Name] = one.Source
	}
	c.Equal(gurps.Source{LibraryFile: alphaFile.LibraryFile, TID: alpha.TID}, sources["Alpha Skill"])
	c.Equal(gurps.Source{LibraryFile: betaFile.LibraryFile, TID: alpha.TID}, sources["Beta Skill"],
		"a row the view gave an ID of its own records the ID it has in its file")
}

// openSkillLists returns the open skill lists, by the path of their files.
func openSkillLists() map[string]*TableDockable[*gurps.Skill] {
	lists := make(map[string]*TableDockable[*gurps.Skill])
	for _, one := range AllDockables() {
		if list, ok := one.(*TableDockable[*gurps.Skill]); ok {
			lists[list.BackingFilePath()] = list
		}
	}
	return lists
}

// TestLibraryFilterShowInSourceFile verifies that Show in Source File, which the view's context menu offers after Open
// Detail Editor, opens the list file each selected row came from and selects there exactly those rows, by the IDs they
// have in their files, a row the view gave an ID of its own among them, opening the closed container above one; that
// the file of the first selected row ends up with the focus, though other files were opened too; that the user is told
// how many of the rows weren't found, such as one whose file has changed since the view read it; and that more than ten
// files are opened only once the user has agreed to it.
func TestLibraryFilterShowInSourceFile(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	var warned []int
	uxtest.SwapForTest(t, &warnRowsNotInSourceFiles, func(count int) { warned = append(warned, count) })
	user := gurps.GlobalSettings().Libraries.User()
	alpha := newTestSkill("Alpha", fxp.One, nil)
	hidden := newTestSkill("Hidden Xray", fxp.One, nil)
	group := newTestSkillContainer("Group", hidden)
	other := newTestSkill("Other Xray", fxp.One, nil)
	other.TID = alpha.TID
	first := writeLibraryFilterTestSkills(c, user, "Skills/A.skl", alpha, group)
	second := writeLibraryFilterTestSkills(c, user, "Skills/B.skl", other)
	third := writeLibraryFilterTestSkills(c, user, "Skills/C.skl", newTestSkill("Changed Xray", fxp.One, nil))
	for i := range maxSourceFilesWithoutAsking + 1 {
		writeLibraryFilterTestSkills(c, user, fmt.Sprintf("Many/%d.skl", i),
			newTestSkill(fmt.Sprintf("Many %d", i), fxp.One, nil))
	}
	group.SetOpen(false) // The open state is kept in the settings, by ID, for the list file's container too.
	d := openTestLibraryFilter(t, screen,
		&gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{{Library: user.Key(), Path: "Skills"}}})
	content := skillLibraryFilterContent(t, d)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The view is to go on showing the row the third file no longer holds.
	stopNavigatorWatches(t, screen, n)
	writeLibraryFilterTestSkills(c, user, third.Path, newTestSkill("Changed Xray", fxp.One, nil))

	screen.Do(func() {
		items := content.decorator.ContextMenuItems()
		i := slices.IndexFunc(items, func(item ContextMenuItem) bool { return item.ID == OpenEditorItemID })
		c.True(i != -1 && i+1 < len(items) && items[i+1].ID == ShowInSourceFileItemID,
			"the context menu offers the command after Open Detail Editor")
		// The filter shows the closed container as context, with its row beneath it, and leaves it closed.
		content.filterField.SetText("xray")
		rows := libraryFilterRowsByName(content.table)
		content.table.SetSelectionMap(map[tid.TID]bool{
			rows["Hidden Xray"].ID():  true,
			rows["Other Xray"].ID():   true,
			rows["Changed Xray"].ID(): true,
		})
		c.Equal(3, content.table.SelectionCount(), "the rows are selected in the view")
		content.table.PerformCmd(nil, ShowInSourceFileItemID)
	})
	screen.Sync()
	screen.Do(func() {
		lists := openSkillLists()
		firstList, secondList := lists[first.fullPath], lists[second.fullPath]
		if firstList == nil || secondList == nil || lists[third.fullPath] == nil {
			t.Errorf("every file must be open: %v", slices.Collect(maps.Keys(lists)))
			return
		}
		c.Equal(map[tid.TID]bool{hidden.TID: true}, firstList.table.CopySelectionMap(),
			"the first file has the row within the container selected")
		c.True(group.IsOpen(), "the container is opened")
		c.Equal(map[tid.TID]bool{alpha.TID: true}, secondList.table.CopySelectionMap(),
			"the second file has the row the view gave an ID of its own selected, by the ID it has there")
		active := ActiveDockable()
		c.True(active != nil && active.AsPanel() == firstList.AsPanel(), "the first selected row's file has the focus")
	})
	c.Equal([]int{1}, warned, "the user is told once that one row wasn't found")

	var asked int
	uxtest.SwapForTest(t, &confirmShowInSourceFiles, func(count int) bool {
		asked = count
		return false
	})
	screen.Do(func() {
		content.setQuickFilterText("")
		d.chooseSet(librarySkillsKey,
			&gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{{Library: user.Key(), Path: "Many"}}}, true)
	})
	awaitLibraryFilter(t, screen, "the rows of the many files", func() bool {
		return len(shownSkillNames(content.table)) == maxSourceFilesWithoutAsking+1
	})
	var before int
	screen.Do(func() {
		before = len(openSkillLists())
		content.table.SelectAll()
		content.table.PerformCmd(nil, ShowInSourceFileItemID)
	})
	screen.Sync()
	c.Equal(maxSourceFilesWithoutAsking+1, asked, "the user is asked before so many files are opened")
	screen.Do(func() { c.Equal(before, len(openSkillLists()), "declining opens nothing") })
	c.Equal([]int{1}, warned, "nor tells of rows not found")
}

// TestLibraryFilterReloadClosesEditorsOfReplacedRows verifies that reading a changed file again closes the editors open
// on its rows, which would otherwise go on showing the rows as they were, even though the new rows have the same IDs,
// while an editor open on a row of a file that wasn't read again stays open.
func TestLibraryFilterReloadClosesEditorsOfReplacedRows(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	user := gurps.GlobalSettings().Libraries.User()
	changed := newTestSkill("Changed", fxp.One, nil)
	changedFile := writeLibraryFilterTestSkills(c, user, "Skills/Changed.skl", changed)
	writeLibraryFilterTestSkills(c, user, "Skills/Kept.skl", newTestSkill("Kept", fxp.One, nil))
	d := openTestLibraryFilter(t, screen, nil)
	content := skillLibraryFilterContent(t, d)
	var n *Navigator
	screen.Do(func() { n = Workspace.Navigator })
	// The test reads the file again itself, so that the watches don't have it read a second time.
	stopNavigatorWatches(t, screen, n)

	screen.Do(func() {
		rows := libraryFilterRowsByName(content.table)
		for _, name := range []string{"Changed", "Kept"} {
			content.table.SetSelectionMap(map[tid.TID]bool{rows[name].ID(): true})
			content.table.PerformCmd(nil, OpenEditorItemID)
		}
	})
	editorOpenOn := func(name string) bool {
		var open bool
		screen.Do(func() {
			for _, one := range AllDockables() {
				e, ok := one.AsPanel().Self.(*editor[*gurps.Skill, *gurps.SkillEditData])
				open = open || (ok && e.target.Name == name)
			}
		})
		return open
	}
	c.True(editorOpenOn("Changed"), "an editor is open on the row of the file to be changed")
	c.True(editorOpenOn("Kept"), "and on the row of the other file")

	changed.Name = "Changed Again" // Under the same ID
	writeLibraryFilterTestSkills(c, user, changedFile.Path, changed)
	screen.Do(d.resolveAndLoad)
	awaitLibraryFilter(t, screen, "the changed file's rows", func() bool {
		return slices.Contains(shownSkillNames(content.table), "Changed Again")
	})
	c.False(editorOpenOn("Changed"), "the editor on the row read again is closed")
	c.True(editorOpenOn("Kept"), "the editor on the row of the file not read again stays open")
}
