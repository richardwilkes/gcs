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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/toolbox/v2/check"
)

// fakeDiscloser stands in for a list in the toggle tests, recording the state it was last put in.
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
	swapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
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

// TestTableDockableTogglesAreOffWhileFiltered verifies that a list dockable turns its hierarchy and note buttons off
// while a filter is applied, and that the toggles leave the rows alone should they be reached all the same. A filtered
// table shows every container it keeps as open whatever the container's own open state, so a toggle would silently
// change the disclosure states without anything to show for it until the filter was cleared. Both the quick filter and
// a saved filter have to turn the buttons off, and clearing either has to turn them back on.
func TestTableDockableTogglesAreOffWhileFiltered(t *testing.T) {
	c := check.New(t)
	swapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	first := newTestContainer("First", true)
	first.LocalNotes = "A note"
	second := newTestContainer("Second", false)
	dockable := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{first, second})
	c.True(dockable.hierarchyButton.Enabled(), "the hierarchy button starts out on")
	c.True(dockable.noteToggleButton.Enabled(), "the note button starts out on")

	dockable.filterField.SetText("first")
	c.True(dockable.table.IsFiltered(), "typing in the quick filter must filter the table")
	c.False(dockable.hierarchyButton.Enabled(), "the hierarchy button must be off while the quick filter is applied")
	c.False(dockable.noteToggleButton.Enabled(), "the note button must be off while the quick filter is applied")
	dockable.toggleHierarchy()
	c.True(first.IsOpen(), "the matched container must be left open")
	c.False(second.IsOpen(), "the unmatched container must be left closed")
	dockable.toggleNotes()
	c.Equal(1, dockable.FirstNoteState(), "the note must be left shown")

	dockable.filterField.SetText("")
	c.False(dockable.table.IsFiltered(), "emptying the quick filter must show everything")
	c.True(dockable.hierarchyButton.Enabled(), "clearing the quick filter must turn the hierarchy button back on")
	c.True(dockable.noteToggleButton.Enabled(), "clearing the quick filter must turn the note button back on")

	dockable.chooseFilter(newNameContainsFilter("First", "first"))
	c.True(dockable.table.IsFiltered(), "the saved filter must filter the table")
	c.False(dockable.hierarchyButton.Enabled(), "the hierarchy button must be off while a saved filter is applied")
	c.False(dockable.noteToggleButton.Enabled(), "the note button must be off while a saved filter is applied")
	dockable.toggleHierarchy()
	c.True(first.IsOpen(), "the matched container must be left open under the saved filter as well")
	c.False(second.IsOpen(), "the unmatched container must be left closed under the saved filter as well")

	dockable.chooseFilter(nil)
	c.False(dockable.table.IsFiltered(), "dropping the saved filter must show everything")
	c.True(dockable.hierarchyButton.Enabled(), "dropping the saved filter must turn the hierarchy button back on")
	c.True(dockable.noteToggleButton.Enabled(), "dropping the saved filter must turn the note button back on")
	dockable.toggleHierarchy()
	c.False(first.IsOpen(), "with no filter, the open first container sets the direction, so it must be closed")
	c.False(second.IsOpen(), "and the second one must be closed along with it")
}
