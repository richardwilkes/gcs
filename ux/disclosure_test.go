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
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// fakeDiscloser stands in for a list in the toggle tests, recording every state it is put in.
type fakeDiscloser struct {
	open      bool
	exists    bool
	noteState int
	setOpen   []bool
	setClosed []bool
}

func (f *fakeDiscloser) FirstDisclosureState() (open, exists bool) { return f.open, f.exists }
func (f *fakeDiscloser) SetDisclosureState(open bool)              { f.setOpen = append(f.setOpen, open) }
func (f *fakeDiscloser) FirstNoteState() int                       { return f.noteState }
func (f *fakeDiscloser) ApplyNoteState(closed bool)                { f.setClosed = append(f.setClosed, closed) }

// TestToggleHierarchy verifies that toggling the hierarchy of a group takes its cue from the first member that has a
// container at all -- an empty list ahead of it doesn't count -- and puts every member, including that empty one, into
// the opposite state, and that a group with no container anywhere is opened.
func TestToggleHierarchy(t *testing.T) {
	c := check.New(t)
	empty := &fakeDiscloser{}
	open := &fakeDiscloser{open: true, exists: true}
	closed := &fakeDiscloser{open: false, exists: true}

	c.False(toggleHierarchy(empty, open, closed), "an open first container must close the group")
	for _, d := range []*fakeDiscloser{empty, open, closed} {
		c.Equal([]bool{false}, d.setOpen, "every member must be closed")
	}

	empty.setOpen, open.setOpen, closed.setOpen = nil, nil, nil
	c.True(toggleHierarchy(empty, closed, open), "a closed first container must open the group")
	for _, d := range []*fakeDiscloser{empty, open, closed} {
		c.Equal([]bool{true}, d.setOpen, "every member must be opened")
	}

	none := &fakeDiscloser{}
	c.True(toggleHierarchy(none), "with no container to go by, the group is opened")
	c.Equal([]bool{true}, none.setOpen, "the state is applied even so")

	c.True(toggleHierarchy[*fakeDiscloser](), "an empty group is harmless")
}

// TestToggleNotes verifies that toggling the notes of a group takes its cue from the first member that has a note at
// all, hides every member's notes when that one is shown and shows them when it is hidden, and that a group without
// a note anywhere is left alone and reported as such.
func TestToggleNotes(t *testing.T) {
	c := check.New(t)
	none := &fakeDiscloser{}
	shown := &fakeDiscloser{noteState: 1}
	hidden := &fakeDiscloser{noteState: -1}

	c.True(toggleNotes(none, shown, hidden), "a shown first note must hide the group's notes")
	for _, d := range []*fakeDiscloser{none, shown, hidden} {
		c.Equal([]bool{true}, d.setClosed, "every member's notes must be hidden")
	}

	none.setClosed, shown.setClosed, hidden.setClosed = nil, nil, nil
	c.True(toggleNotes(none, hidden, shown), "a hidden first note must show the group's notes")
	for _, d := range []*fakeDiscloser{none, shown, hidden} {
		c.Equal([]bool{false}, d.setClosed, "every member's notes must be shown")
	}

	none.setClosed = nil
	c.False(toggleNotes(none), "with no note to act on, nothing is changed")
	c.Nil(none.setClosed, "no state is applied")
	c.False(toggleNotes[*fakeDiscloser](), "an empty group has nothing to act on")
}

// newTestContainer returns a container trait of the given name holding one child, open or closed as asked.
func newTestContainer(name string, open bool) *gurps.Trait {
	container := gurps.NewTrait(nil, nil, true)
	container.Name = name
	child := gurps.NewTrait(nil, container, false)
	child.Name = name + " Child"
	container.Children = []*gurps.Trait{child}
	container.SetOpen(open)
	return container
}

// TestTemplateToggleHierarchyFlipsEveryList verifies that a template's hierarchy button flips all of its lists together,
// going by the first container it finds: an open container at the head of the traits list closes the skills list's
// container as well, and the next press opens both again.
func TestTemplateToggleHierarchyFlipsEveryList(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	trait := newTestContainer("Traits", true)
	data.Traits = []*gurps.Trait{trait}
	skill := gurps.NewSkill(nil, nil, true)
	skill.Name = "Skills"
	skillChild := gurps.NewSkill(nil, skill, false)
	skillChild.Name = "Skill Child"
	skill.Children = []*gurps.Skill{skillChild}
	skill.SetOpen(false)
	data.Skills = []*gurps.Skill{skill}
	template := newTestTemplateDockable("Toggle", data)

	template.toggleHierarchy()
	c.False(trait.IsOpen(), "the open trait container sets the direction, so it must be closed")
	c.False(skill.IsOpen(), "the skill container must be closed along with it")
	open, exists := template.Skills.FirstDisclosureState()
	c.True(exists, "the skills list must still see its container")
	c.False(open, "the skills list must report the container closed")

	template.toggleHierarchy()
	c.True(trait.IsOpen(), "the next press must open the trait container again")
	c.True(skill.IsOpen(), "and the skill container with it")
}

// TestTableDockableToggles verifies that a list dockable's hierarchy and note buttons work the same way as a page
// list's, going by the first container or note in the table and applying the opposite state to every row.
func TestTableDockableToggles(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	first := newTestContainer("First", true)
	second := newTestContainer("Second", false)
	second.LocalNotes = "A note"
	dockable := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{first, second})

	dockable.toggleHierarchy()
	c.False(first.IsOpen(), "the open first container sets the direction, so it must be closed")
	c.False(second.IsOpen(), "the closed second container must stay closed")
	dockable.toggleHierarchy()
	c.True(first.IsOpen(), "the next press must open the first container")
	c.True(second.IsOpen(), "and the second one with it")

	c.Equal(1, dockable.FirstNoteState(), "the note starts out shown")
	dockable.toggleNotes()
	c.Equal(-1, dockable.FirstNoteState(), "the first press must hide the note")
	dockable.toggleNotes()
	c.Equal(1, dockable.FirstNoteState(), "the next press must show it again")

	second.LocalNotes = ""
	c.Equal(0, dockable.FirstNoteState(), "with no note anywhere there is nothing to report")
	dockable.toggleNotes() // Nothing to act on; must not panic or change anything.
}

// TestOpenStateChangesReachOtherViews verifies that opening or closing containers and notes in one view brings the
// other views showing rows with the same IDs, whose open states are kept with them, into line, as a library filter view
// and the list of a file it shows do: the list shows what a container the view opened holds and sizes a row for the
// note the view hid. The edits in a list that open a container to show what they put in it, and undoing and redoing
// them, which close and open it again, do so as well, here for the list of a copy of the file.
func TestOpenStateChangesReachOtherViews(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().Closed, make(map[string]int64))
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	box := newTestSkillContainer("Box", newTestSkill("Inside", fxp.One, nil))
	box.SetOpen(false)
	noted := newTestSkill("Noted", fxp.One, nil)
	noted.LocalNotes = "A note\nof three\nlines"
	file := writeLibraryFilterTestSkills(c, gurps.GlobalSettings().Libraries.User(), "Skills/X.skl", box, noted)
	d := openTestLibraryFilter(t, screen, &gurps.LibrarySourceSet{Entries: []gurps.LibraryFile{file.LibraryFile}})
	content := skillLibraryFilterContent(t, d)
	var list *TableDockable[*gurps.Skill]
	screen.Do(func() {
		one, _ := OpenFile(file.fullPath, gurps.PageInfo{})
		if opened, ok := one.(*TableDockable[*gurps.Skill]); ok {
			list = opened
		}
	})
	if list == nil {
		t.Fatal("the file must open as a list of skills")
	}

	screen.Do(func() {
		c.Equal([]string{"Box", "Noted"}, shownSkillNames(list.table), "the list starts with the container closed")
		content.toggleHierarchy()
	})
	screen.Sync()
	var height float32
	screen.Do(func() {
		c.Equal([]string{"Box", "Inside", "Noted"}, shownSkillNames(list.table),
			"the list shows what the container the view opened holds")
		height = list.table.RowHeights()[list.table.LastRowIndex()]
		content.toggleNotes()
	})
	screen.Sync()
	screen.Do(func() {
		c.True(list.table.RowHeights()[list.table.LastRowIndex()] < height,
			"the list's row is no taller than the note the view hid leaves it")
	})

	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialListLocked, false)
	first := gurps.NewTrait(nil, nil, false)
	first.Name = "First"
	container := newTestContainer("Box", false)
	var lists []*TableDockable[*gurps.Trait]
	for _, name := range []string{"X" + gurps.TraitsExt, "Y" + gurps.TraitsExt} {
		path := filepath.Join(t.TempDir(), name)
		c.NoError(gurps.SaveTraits([]*gurps.Trait{first, container}, path))
		screen.Do(func() {
			one, _ := OpenFile(path, gurps.PageInfo{})
			if opened, ok := one.(*TableDockable[*gurps.Trait]); ok {
				lists = append(lists, opened)
			}
		})
	}
	if len(lists) != 2 {
		t.Fatal("both files must open as lists of traits")
	}
	edited, other := lists[0], lists[1]
	var mgr *unison.UndoManager
	screen.Do(func() { mgr = unison.UndoManagerFor(edited.table) })
	otherShowsContents := func(shows bool, what string) {
		t.Helper()
		screen.Sync()
		want := []string{"First", "Box"}
		if shows {
			want = append(want, "Box Child")
		}
		names := make([]string, 0, len(want))
		screen.Do(func() {
			for i := range other.table.LastRowIndex() + 1 {
				names = append(names, other.table.RowFromIndex(i).Data().Name)
			}
		})
		c.Equal(want, names, what)
	}
	undoAndRedo := func(edit string) {
		t.Helper()
		screen.Do(mgr.Undo)
		otherShowsContents(false, "undoing "+edit+" closes the container in the other list")
		screen.Do(mgr.Redo)
		otherShowsContents(true, "redoing "+edit+" opens the container in the other list again")
		screen.Do(mgr.Undo)
	}

	screen.Do(func() {
		edited.table.SetSelectionMap(map[tid.TID]bool{first.ID(): true})
		MoveSelection(edited.table, MoveIntoContainer)
	})
	otherShowsContents(true, "Move Into Container opens the container in the other list")
	undoAndRedo("Move Into Container")

	screen.Do(func() {
		modifier := gurps.NewTraitModifier(nil, nil, false)
		modifier.Name = "Fine"
		target := edited.table.RootRows()[1].Data().Children[0]
		c.True(applyModifiersTo([]*unison.Table[*Node[*gurps.Trait]]{edited.table}, []*gurps.Trait{target},
			[]*gurps.TraitModifier{modifier}, gurps.LibraryFile{}))
	})
	otherShowsContents(true, "applying a modifier to a row in the container opens it in the other list")
	undoAndRedo("applying a modifier")

	screen.Do(func() {
		c.True(applyDrop(&unison.TableDragData[*Node[*gurps.Trait]]{Table: other.table, Rows: other.table.RootRows()[:1]},
			edited.table, edited.table.RootRows()[1], 0))
	})
	otherShowsContents(true, "a drop into the container opens it in the other list")
	undoAndRedo("a drop into the container")
}

// TestTableDockableTogglesReachNestedRows verifies that a list dockable's hierarchy and note buttons reach the rows
// within containers at every depth: a container within another is opened and closed along with it, and a note found
// only on a row within a container both sets the direction and is shown and hidden.
func TestTableDockableTogglesReachNestedRows(t *testing.T) {
	c := check.New(t)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	outer := gurps.NewTrait(nil, nil, true)
	outer.Name = "Outer"
	outer.SetOpen(true)
	inner := newTestContainer("Inner", true)
	inner.SetParent(outer)
	outer.Children = []*gurps.Trait{inner}
	inner.Children[0].LocalNotes = "A note"
	dockable := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{outer})

	dockable.toggleHierarchy()
	c.False(outer.IsOpen(), "the open outer container sets the direction, so it must be closed")
	c.False(inner.IsOpen(), "the container within it must be closed along with it")
	dockable.toggleHierarchy()
	c.True(outer.IsOpen(), "the next press must open the outer container")
	c.True(inner.IsOpen(), "and the container within it")

	c.Equal(1, dockable.FirstNoteState(), "the note of a row within a container must be found")
	dockable.toggleNotes()
	c.Equal(-1, dockable.FirstNoteState(), "the first press must hide the note of the row within a container")
	dockable.toggleNotes()
	c.Equal(1, dockable.FirstNoteState(), "the next press must show it again")
}
