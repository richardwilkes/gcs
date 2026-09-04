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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// moveTestTraits holds the traits newMoveTestSheet puts onto its sheet: a container holding three plain traits, sitting
// between two more plain traits.
type moveTestTraits struct {
	first, box, one, two, three, last *gurps.Trait
}

// newMoveTestSheet returns a sheet whose traits list is First, Box[One, Two, Three], Last, along with those traits.
func newMoveTestSheet(t *testing.T) (*Sheet, moveTestTraits) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	newTrait := func(name string, parent *gurps.Trait, container bool) *gurps.Trait {
		trait := gurps.NewTrait(entity, parent, container)
		trait.Name = name
		return trait
	}
	traits := moveTestTraits{
		first: newTrait("First", nil, false),
		box:   newTrait("Box", nil, true),
		last:  newTrait("Last", nil, false),
	}
	traits.one = newTrait("One", traits.box, false)
	traits.two = newTrait("Two", traits.box, false)
	traits.three = newTrait("Three", traits.box, false)
	traits.box.Children = []*gurps.Trait{traits.one, traits.two, traits.three}
	entity.Traits = []*gurps.Trait{traits.first, traits.box, traits.last}
	sheet.Rebuild(true)
	return sheet, traits
}

// selectTraits makes the given traits the table's selection.
func selectTraits(table *unison.Table[*Node[*gurps.Trait]], traits ...*gurps.Trait) {
	selMap := make(map[tid.TID]bool, len(traits))
	for _, trait := range traits {
		selMap[trait.ID()] = true
	}
	table.SetSelectionMap(selMap)
}

// selectedTraitNames returns the names of the selected traits, in table order.
func selectedTraitNames(table *unison.Table[*Node[*gurps.Trait]]) []string {
	return traitNames(ExtractNodeDataFromList(table.SelectedRows(false)))
}

// TestMoveSelectionUpAndDownWithinContainer verifies that Move Up and Move Down shift a row one place among its
// siblings through the command the menu and key binding route to, keep it selected, and become unavailable once it
// reaches the end of its list.
func TestMoveSelectionUpAndDownWithinContainer(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	table := sheet.Traits.Table
	selectTraits(table, traits.two)

	c.True(table.CanPerformCmd(table, MoveUpItemID), "a row with a sibling above it must be able to move up")
	table.PerformCmd(table, MoveUpItemID)
	c.Equal([]string{"Two", "One", "Three"}, traitNames(traits.box.Children), "Move Up must swap the row with the one above")
	c.Equal([]string{"Two"}, selectedTraitNames(sheet.Traits.Table), "the moved row must stay selected")
	c.Equal(traits.box, traits.two.Parent(), "moving within a container must not change the row's parent")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveUpItemID),
		"the first row in a list must not be able to move up")

	table = sheet.Traits.Table
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"One", "Two", "Three"}, traitNames(traits.box.Children), "Move Down must swap the row with the one below")
	table = sheet.Traits.Table
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"One", "Three", "Two"}, traitNames(traits.box.Children), "Move Down must keep going one place at a time")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveDownItemID),
		"the last row in a list must not be able to move down")
	c.Equal([]string{"First", "Box", "Last"}, traitNames(sheet.Entity().Traits),
		"moving within a container must leave the top-level list alone")
}

// TestMoveSelectionKeepsSelectedRowsInOrder verifies how several selected siblings move together: a run of them moves
// as a block, so their order is preserved, and a selected row that has reached the end of its list holds back the
// selected rows directly behind it rather than letting them leapfrog it.
func TestMoveSelectionKeepsSelectedRowsInOrder(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	table := sheet.Traits.Table

	selectTraits(table, traits.one, traits.two)
	c.False(table.CanPerformCmd(table, MoveUpItemID),
		"a run of selected rows at the top of its list has nowhere to go")
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"Three", "One", "Two"}, traitNames(traits.box.Children),
		"a run of selected rows must move down as a block, keeping its order")
	c.Equal([]string{"One", "Two"}, selectedTraitNames(sheet.Traits.Table), "both moved rows must stay selected")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveDownItemID),
		"a run of selected rows at the bottom of its list has nowhere to go")

	// Rows that aren't next to each other move independently, and a selected row blocked at the top stops the selected
	// row that arrives directly beneath it.
	table = sheet.Traits.Table
	selectTraits(table, traits.three, traits.two)
	table.PerformCmd(table, MoveUpItemID)
	c.Equal([]string{"Three", "Two", "One"}, traitNames(traits.box.Children),
		"a selected row must move up past an unselected sibling while the selected row above it stays put")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveUpItemID),
		"selected rows stacked at the top of the list must not be able to move up")
}

// TestMoveSelectionOutOfContainer verifies that Move Out of Container lifts a row out of its container to the place
// just above it, that rows lifted out together keep their order, and that a top-level row can't be lifted any further.
func TestMoveSelectionOutOfContainer(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table

	selectTraits(table, traits.two)
	c.True(table.CanPerformCmd(table, MoveOutOfContainerItemID), "a row in a container must be able to move out of it")
	table.PerformCmd(table, MoveOutOfContainerItemID)
	c.Equal([]string{"First", "Two", "Box", "Last"}, traitNames(entity.Traits),
		"the row must land immediately above the container it came out of")
	c.Equal([]string{"First", "Two", "Box", "Last"}, rootRowNames(sheet.Traits.Table), "the table must show the new order")
	c.Equal([]string{"One", "Three"}, traitNames(traits.box.Children), "the row must be gone from the container")
	c.Nil(traits.two.Parent(), "a row moved to the top level must have no parent")
	c.Equal([]string{"Two"}, selectedTraitNames(sheet.Traits.Table), "the moved row must stay selected")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveOutOfContainerItemID),
		"a top-level row has no container to move out of")

	table = sheet.Traits.Table
	selectTraits(table, traits.one, traits.three)
	table.PerformCmd(table, MoveOutOfContainerItemID)
	c.Equal([]string{"First", "Two", "One", "Three", "Box", "Last"}, traitNames(entity.Traits),
		"rows moved out together must keep their order, above the container")
	c.Equal(0, len(traits.box.Children), "the container must be empty once everything has been moved out")
	c.Equal([]string{"One", "Three"}, selectedTraitNames(sheet.Traits.Table), "both moved rows must stay selected")
}

// TestMoveSelectionIntoContainer verifies that Move Into Container drops a row into the container directly below it as
// its first child, that a run of rows above the container enters it in order, and that nothing happens when what sits
// below the row isn't a container, is a selected one, or is nothing at all.
func TestMoveSelectionIntoContainer(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	zero := gurps.NewTrait(entity, nil, false)
	zero.Name = "Zero"
	entity.Traits = []*gurps.Trait{zero, traits.first, traits.box, traits.last}
	sheet.Rebuild(true)
	table := sheet.Traits.Table

	selectTraits(table, traits.last)
	c.False(table.CanPerformCmd(table, MoveIntoContainerItemID), "the last row has nothing below it to move into")
	selectTraits(table, zero)
	c.False(table.CanPerformCmd(table, MoveIntoContainerItemID),
		"a row with a plain row directly below it has no container to move into")
	selectTraits(table, traits.first, traits.box)
	c.False(table.CanPerformCmd(table, MoveIntoContainerItemID), "a row must not be moved into a selected container")
	selectTraits(table, traits.one)
	c.False(table.CanPerformCmd(table, MoveIntoContainerItemID),
		"a row inside a container with plain siblings below it has nowhere to go")

	selectTraits(table, traits.first)
	c.True(table.CanPerformCmd(table, MoveIntoContainerItemID), "a row directly above a container must be able to move in")
	table.PerformCmd(table, MoveIntoContainerItemID)
	c.Equal([]string{"Zero", "Box", "Last"}, traitNames(entity.Traits), "the row must be gone from the top level")
	c.Equal([]string{"First", "One", "Two", "Three"}, traitNames(traits.box.Children),
		"the row must become the container's first child")
	c.Equal(traits.box, traits.first.Parent(), "the row's parent must now be the container")
	c.Equal([]string{"First"}, selectedTraitNames(sheet.Traits.Table), "the moved row must stay selected")
	c.False(sheet.Traits.Table.CanPerformCmd(sheet.Traits.Table, MoveIntoContainerItemID),
		"once inside, the row has a plain sibling below it and can't go any further in")

	// For what was the container's first child, Move Out of Container puts things back exactly as they were.
	table = sheet.Traits.Table
	table.PerformCmd(table, MoveOutOfContainerItemID)
	c.Equal([]string{"Zero", "First", "Box", "Last"}, traitNames(entity.Traits), "moving out must undo moving in")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(traits.box.Children), "moving out must undo moving in")

	// A run of selected rows above the container enters it as a block, keeping its order.
	table = sheet.Traits.Table
	selectTraits(table, zero, traits.first)
	table.PerformCmd(table, MoveIntoContainerItemID)
	c.Equal([]string{"Box", "Last"}, traitNames(entity.Traits), "both rows must be gone from the top level")
	c.Equal([]string{"Zero", "First", "One", "Two", "Three"}, traitNames(traits.box.Children),
		"rows moved in together must keep their order at the front of the container")
	c.Equal([]string{"Zero", "First"}, selectedTraitNames(sheet.Traits.Table), "both moved rows must stay selected")

	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	mgr.Undo()
	c.Equal([]string{"Zero", "First", "Box", "Last"}, traitNames(entity.Traits), "undo must put the rows back")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(entity.Traits[2].Children),
		"undo must take the rows back out of the container")
	c.True(traits.box.IsOpen(), "undo must leave a container that was already open as it was")
}

// TestMoveIntoContainerPutsRowAtFront verifies that Move Into Container always makes the row the container's first
// child, so that moving a row out and then back in only returns it to its old place if it was the first child to begin
// with: Move Out puts the row directly above the container no matter where inside it the row was.
func TestMoveIntoContainerPutsRowAtFront(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table
	selectTraits(table, traits.two)
	table.PerformCmd(table, MoveOutOfContainerItemID)
	c.Equal([]string{"First", "Two", "Box", "Last"}, traitNames(entity.Traits), "the row must land directly above the container")
	c.Equal([]string{"One", "Three"}, traitNames(traits.box.Children), "the row must be gone from the middle of the container")

	table = sheet.Traits.Table
	table.PerformCmd(table, MoveIntoContainerItemID)
	c.Equal([]string{"First", "Box", "Last"}, traitNames(entity.Traits), "the row must be back in the container")
	c.Equal([]string{"Two", "One", "Three"}, traitNames(traits.box.Children),
		"the row must go to the front of the container rather than back to where it came from")
}

// TestMoveSelectionIntoClosedContainerOpensIt verifies that moving a row into a closed container opens the container,
// so that the row stays in view and selected rather than disappearing into it, that undo closes the container again
// along with putting the row back, and that redo opens it once more so that the row it moves back in is showing.
func TestMoveSelectionIntoClosedContainerOpensIt(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	traits.box.SetOpen(false)
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	c.Equal(2, table.LastRowIndex(), "the closed container must hide its children")

	selectTraits(table, traits.first)
	table.PerformCmd(table, MoveIntoContainerItemID)
	c.True(traits.box.IsOpen(), "moving a row into a closed container must open it")
	c.Equal([]string{"First", "One", "Two", "Three"}, traitNames(traits.box.Children),
		"the row must become the container's first child")
	c.Equal([]string{"First"}, selectedTraitNames(sheet.Traits.Table), "the moved row must stay selected and visible")

	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	mgr.Undo()
	c.False(traits.box.IsOpen(), "undo must close the container the move opened")
	c.Equal([]string{"First", "Box", "Last"}, traitNames(entity.Traits), "undo must put the row back")
	c.Equal(2, sheet.Traits.Table.LastRowIndex(), "the container must be hiding its children again")
	c.Equal([]string{"First"}, selectedTraitNames(sheet.Traits.Table), "undo must restore the selection")

	mgr.Redo()
	c.True(traits.box.IsOpen(), "redo must open the container again so that the row it moves is showing")
	c.Equal([]string{"First", "One", "Two", "Three"}, traitNames(entity.Traits[0].Children),
		"redo must move the row back into the container")
	c.Equal([]string{"First"}, selectedTraitNames(sheet.Traits.Table), "redo must select the moved row")
}

// TestMoveSelectionMovesContainerWithContents verifies that a selected container moves as a unit, taking its contents
// along, and that selecting something inside it as well doesn't move that thing separately.
func TestMoveSelectionMovesContainerWithContents(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table

	selectTraits(table, traits.box)
	table.PerformCmd(table, MoveUpItemID)
	c.Equal([]string{"Box", "First", "Last"}, traitNames(entity.Traits), "the container must move up as a unit")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(traits.box.Children), "the container's contents must come along")

	table = sheet.Traits.Table
	selectTraits(table, traits.box, traits.two)
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"First", "Box", "Last"}, traitNames(entity.Traits), "the container must move down as a unit")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(traits.box.Children),
		"a selected row inside a selected container must not be moved on its own")

	table = sheet.Traits.Table
	selectTraits(table, traits.box, traits.two)
	c.False(table.CanPerformCmd(table, MoveOutOfContainerItemID),
		"a selected row inside a selected container must not offer to move out on its own")
}

// TestMoveSelectionIsUndoable verifies that a move is recorded as a single undo edit which puts both the order and the
// selection back, and that it can be redone.
func TestMoveSelectionIsUndoable(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")

	selectTraits(table, traits.one, traits.three)
	table.PerformCmd(table, MoveOutOfContainerItemID)
	c.Equal([]string{"First", "One", "Three", "Box", "Last"}, traitNames(entity.Traits), "the move must have happened")
	c.True(mgr.CanUndo(), "the move must be undoable")

	mgr.Undo()
	c.Equal([]string{"First", "Box", "Last"}, traitNames(entity.Traits), "undo must put the top-level list back")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(entity.Traits[1].Children),
		"undo must put the rows back into the container")
	c.Equal([]string{"First", "Box", "Last"}, rootRowNames(sheet.Traits.Table), "undo must put the table back")
	c.Equal([]string{"One", "Three"}, selectedTraitNames(sheet.Traits.Table), "undo must restore the selection")
	c.False(mgr.CanUndo(), "the move must have been recorded as a single edit")

	c.True(mgr.CanRedo(), "the move must be redoable")
	mgr.Redo()
	c.Equal([]string{"First", "One", "Three", "Box", "Last"}, traitNames(entity.Traits), "redo must move the rows again")
	c.Equal([]string{"One", "Three"}, selectedTraitNames(sheet.Traits.Table), "redo must keep the moved rows selected")
}

// TestMoveSelectionUnavailableWithoutSelectionOrWhileFiltered verifies that none of the commands is offered when there
// is nothing selected or while the table is showing search results, whose flat list has no order to rearrange.
func TestMoveSelectionUnavailableWithoutSelectionOrWhileFiltered(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	table := sheet.Traits.Table
	ids := []int{MoveUpItemID, MoveDownItemID, MoveOutOfContainerItemID, MoveIntoContainerItemID}

	table.ClearSelection()
	for _, id := range ids {
		c.False(table.CanPerformCmd(table, id), "command %d must be unavailable with nothing selected", id)
	}

	table.ApplyFilter(func(row *Node[*gurps.Trait]) bool { return row.Data() != traits.two })
	selectTraits(table, traits.two)
	c.True(table.HasSelection(), "the filtered row must be selectable")
	for _, id := range ids {
		c.False(table.CanPerformCmd(table, id), "command %d must be unavailable while the table is filtered", id)
	}
	table.ApplyFilter(nil)
	// No single row can go every way at once, so the check that the commands come back uses two: First can move down
	// and into the container beneath it, while Two, inside that container, can move up and out of it.
	selectTraits(table, traits.first, traits.two)
	for _, id := range ids {
		c.True(table.CanPerformCmd(table, id), "command %d must be available again once the filter is cleared", id)
	}
}

// TestMoveSelectionFinishesLikeADrop verifies that a move ends the way a drag within the same table does: the provider
// gets to look over the rows that moved, which for a skill on a sheet fills in a blank tech level from the character,
// and the Preconfigured flag is cleared on them -- except in a template, where the flag is there to be set.
func TestMoveSelectionFinishesLikeADrop(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	entity.Profile.TechLevel = "9"
	blank := ""
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Skill"
	skill.TechLevel = &blank
	other := gurps.NewSkill(entity, nil, false)
	other.Name = "Other"
	entity.Skills = []*gurps.Skill{skill, other}
	traits.first.SetPreconfigured(true)
	sheet.Rebuild(true)

	skills := sheet.Skills.Table
	skills.SetSelectionMap(map[tid.TID]bool{skill.ID(): true})
	skills.PerformCmd(skills, MoveDownItemID)
	c.Equal([]string{"Other", "Skill"}, skillNames(entity.Skills), "the skill must have moved")
	c.Equal("9", *skill.TechLevel, "a blank tech level must be filled in from the character, as a drop fills it in")

	table := sheet.Traits.Table
	selectTraits(table, traits.first)
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"Box", "First", "Last"}, traitNames(entity.Traits), "the trait must have moved")
	c.False(traits.first.IsPreconfigured(), "moving a row on a sheet must clear its Preconfigured flag, as a drop does")

	data := gurps.NewTemplate()
	a := gurps.NewTrait(nil, nil, false)
	a.Name = "A"
	a.SetPreconfigured(true)
	b := gurps.NewTrait(nil, nil, false)
	b.Name = "B"
	data.Traits = []*gurps.Trait{a, b}
	template := newTestTemplateDockable("Move", data)
	table = template.Traits.Table
	selectTraits(table, a)
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"B", "A"}, traitNames(data.Traits), "the move must happen in a template as well")
	c.True(a.IsPreconfigured(), "moving a row in a template must leave its Preconfigured flag alone, as a drop does")
}

// TestMoveOutOfParentLeavesStrayRowAlone verifies that a row whose parent doesn't list it among its children -- which
// consistent data never produces -- is left where it is rather than being inserted above a container it was never in,
// matching what the other movers do when a row isn't in the list they expect it in.
func TestMoveOutOfParentLeavesStrayRowAlone(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	entity := sheet.Entity()
	stray := gurps.NewTrait(entity, traits.box, false)
	stray.Name = "Stray"
	c.False(moveOutOfParent(sheet.Traits.provider, stray), "a row that isn't among its parent's children must not move")
	c.Equal([]string{"First", "Box", "Last"}, traitNames(entity.Traits), "nothing must have been inserted above the container")
	c.Equal([]string{"One", "Two", "Three"}, traitNames(traits.box.Children), "the container must be untouched")
	c.Equal(traits.box, stray.Parent(), "the row's parent must be left as it was")
}

// TestMoveShortcutAtTopOfListLeavesSelectionAlone verifies that pressing the Move Up shortcut once too often does
// nothing. With the row already first in its list the command is disabled, so the menu declines the key and it reaches
// the table, whose default handling used to take a command-modified arrow for plain navigation and shift the selection
// to the row above. unison now leaves navigation keys with unrecognized modifiers alone; this checks that the node
// tables' own key handling doesn't undo that, for the shortcuts of Move Up, Increase Uses and Move Out of Container.
func TestMoveShortcutAtTopOfListLeavesSelectionAlone(t *testing.T) {
	c := check.New(t)
	sheet, traits := newMoveTestSheet(t)
	table := sheet.Traits.Table
	selectTraits(table, traits.one)
	c.False(table.CanPerformCmd(table, MoveUpItemID), "the first row in its list must not be able to move up")

	table.KeyDownCallback(unison.KeyUp, moveUpAction.KeyBinding.Modifiers, false)
	c.Equal([]string{"One"}, selectedTraitNames(table), "a declined Move Up must leave the selection alone")
	table.KeyDownCallback(unison.KeyUp, mod.OSMenuCommand(), false)
	c.Equal([]string{"One"}, selectedTraitNames(table), "a declined Increase Uses must leave the selection alone")
	table.KeyDownCallback(unison.KeyLeft, moveOutOfContainerAction.KeyBinding.Modifiers, false)
	c.True(traits.box.IsOpen(), "a declined Move Out of Container must not close the container")
	selectTraits(table, traits.last)
	traits.box.SetOpen(false)
	table.SyncToModel()
	table.KeyDownCallback(unison.KeyRight, moveIntoContainerAction.KeyBinding.Modifiers, false)
	c.False(traits.box.IsOpen(), "a declined Move Into Container must not open the container")
	c.Equal([]string{"Last"}, selectedTraitNames(table), "a declined Move Into Container must leave the selection alone")
	traits.box.SetOpen(true)
	table.SyncToModel()
	selectTraits(table, traits.one)

	c.True(table.KeyDownCallback(unison.KeyUp, mod.None, false), "an unmodified Up must be handled")
	c.Equal([]string{"Box"}, selectedTraitNames(table), "an unmodified Up must still move the selection")
	c.True(table.KeyDownCallback(unison.KeyDown, mod.Shift, false), "a shifted Down must be handled")
	c.Equal([]string{"Box", "One"}, selectedTraitNames(table), "a shifted Down must still extend the selection")
}

// TestMoveSelectionInEditorTable verifies that the commands reach the tables inside the detail editors, which accept
// drags as well, and that a move there is applied to the editor's copy of the data and can be undone.
func TestMoveSelectionInEditorTable(t *testing.T) {
	c := check.New(t)
	e, table, modifier := newTraitEditorWithModifiers(t)
	c.Equal([]string{"Sharp", "Variations"}, modifierNames(e.editorData.Modifiers), "the editor starts with the modifier first")
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the editor's undo manager")

	rows := table.RootRows()
	c.Equal(2, len(rows), "the table must show both modifiers")
	table.SetSelectionMap(map[tid.TID]bool{rows[0].ID(): true})
	c.True(table.CanPerformCmd(table, MoveDownItemID), "the modifier must be able to move down")
	c.False(table.CanPerformCmd(table, MoveOutOfContainerItemID), "a top-level modifier has no container to leave")
	table.PerformCmd(table, MoveDownItemID)
	c.Equal([]string{"Variations", "Sharp"}, modifierNames(e.editorData.Modifiers),
		"the move must be applied to the editor's copy of the modifiers")
	c.Equal([]string{"Variations", "Sharp"}, modifierNames(ExtractNodeDataFromList(table.RootRows())),
		"the table must show the new order")
	c.True(table.IsRowSelected(1), "the moved modifier must stay selected")
	c.Equal(modifier.Name, table.RootRows()[1].Data().Name, "the selected row must be the modifier that moved")

	c.True(mgr.CanUndo(), "the move must be undoable")
	mgr.Undo()
	c.Equal([]string{"Sharp", "Variations"}, modifierNames(e.editorData.Modifiers), "undo must put the order back")
}

// skillNames returns the names of the skills, in order.
func skillNames(skills []*gurps.Skill) []string {
	names := make([]string, 0, len(skills))
	for _, skill := range skills {
		names = append(names, skill.Name)
	}
	return names
}

// modifierNames returns the names of the trait modifiers, in order.
func modifierNames(modifiers []*gurps.TraitModifier) []string {
	names := make([]string, 0, len(modifiers))
	for _, modifier := range modifiers {
		names = append(names, modifier.Name)
	}
	return names
}
