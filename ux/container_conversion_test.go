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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// newSheetWithConvertibleNotes returns a sheet whose notes list holds a plain note, an empty container and a container
// with a child, so that a selection of all three exercises every branch of the conversion checks: the plain note can
// only become a container, the empty container can only become a plain note, and the container with a child cannot be
// converted at all.
func newSheetWithConvertibleNotes(t *testing.T) (sheet *Sheet, plain, empty, parent *gurps.Note) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	plain = gurps.NewNote(entity, nil, false)
	plain.MarkDown = "Plain"
	empty = gurps.NewNote(entity, nil, true)
	empty.MarkDown = "Empty"
	parent = gurps.NewNote(entity, nil, true)
	parent.MarkDown = "Parent"
	child := gurps.NewNote(entity, parent, false)
	child.MarkDown = "Child"
	parent.Children = []*gurps.Note{child}
	entity.Notes = []*gurps.Note{plain, empty, parent}
	sheet.Rebuild(true)
	return sheet, plain, empty, parent
}

// TestContainerConversionChecksHonorDirectionAndChildren verifies that the enablement checks only report a convertible
// selection when a selected row can actually move in the requested direction, and that a container holding children is
// never offered up for conversion to a non-container.
func TestContainerConversionChecksHonorDirectionAndChildren(t *testing.T) {
	c := check.New(t)
	sheet, plain, empty, parent := newSheetWithConvertibleNotes(t)
	table := sheet.Notes.Table

	table.SetSelectionMap(nil)
	c.False(CanConvertToContainer(table), "an empty selection cannot be converted to containers")
	c.False(CanConvertToNonContainer(table), "an empty selection cannot be converted to non-containers")

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true})
	c.True(CanConvertToContainer(table), "a plain note can become a container")
	c.False(CanConvertToNonContainer(table), "a plain note is already a non-container")

	table.SetSelectionMap(map[tid.TID]bool{empty.ID(): true})
	c.False(CanConvertToContainer(table), "an empty container is already a container")
	c.True(CanConvertToNonContainer(table), "an empty container can become a plain note")

	table.SetSelectionMap(map[tid.TID]bool{parent.ID(): true})
	c.False(CanConvertToContainer(table), "a container with children is already a container")
	c.False(CanConvertToNonContainer(table), "a container with children cannot become a plain note")

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	c.True(CanConvertToContainer(table), "a mixed selection can be converted to containers if any row can")
	c.True(CanConvertToNonContainer(table), "a mixed selection can be converted to non-containers if any row can")
}

// TestContainerConversionIsUndoableAsOneEdit verifies that converting a mixed selection only touches the rows that can
// move in the requested direction, that the whole conversion is recorded as a single undo edit carrying the action's
// title, and that undo and redo each put every touched row back where it belongs.
func TestContainerConversionIsUndoableAsOneEdit(t *testing.T) {
	c := check.New(t)
	sheet, plain, empty, parent := newSheetWithConvertibleNotes(t)
	table := sheet.Notes.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.False(mgr.CanUndo(), "nothing must be undoable before any conversion")
	sheet.Entity().ModifiedOn = jio.Time{}

	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	ConvertToContainer(sheet, table)
	c.True(plain.Container(), "the plain note must have become a container")
	c.True(empty.Container(), "the empty container must remain a container")
	c.True(parent.Container(), "the container with children must remain a container")
	c.Equal(1, len(parent.Children), "the container's children must be untouched")
	c.True(mgr.CanUndo(), "the conversion must be undoable")
	c.Equal("Undo "+convertToContainerAction.Title, mgr.UndoTitle(), "the undo edit must carry the action's title")
	c.NotEqual(jio.Time{}, sheet.Entity().ModifiedOn, "the conversion must mark the sheet as modified")

	mgr.Undo()
	c.False(plain.Container(), "undo must turn the note back into a plain note")
	c.True(empty.Container(), "undo must leave the untouched container alone")
	c.False(mgr.CanUndo(), "the whole conversion must have been a single undo edit")

	c.True(mgr.CanRedo(), "the conversion must be redoable")
	mgr.Redo()
	c.True(plain.Container(), "redo must turn the note back into a container")
	mgr.Undo()

	// Now the other direction: both the plain note and the empty container are selected along with the container that
	// has children, but only the empty container can become a plain note.
	table = sheet.Notes.Table
	table.SetSelectionMap(map[tid.TID]bool{plain.ID(): true, empty.ID(): true, parent.ID(): true})
	ConvertToNonContainer(sheet, table)
	c.False(plain.Container(), "the plain note must remain a plain note")
	c.False(empty.Container(), "the empty container must have become a plain note")
	c.True(parent.Container(), "the container with children must not be converted")
	c.Equal("Undo "+convertToNonContainerAction.Title, mgr.UndoTitle(), "the undo edit must carry the action's title")

	mgr.Undo()
	c.True(empty.Container(), "undo must turn the note back into a container")
	c.False(plain.Container(), "undo must leave the untouched plain note alone")
	c.True(parent.Container(), "undo must leave the container with children alone")

	// A selection with nothing convertible must not record an edit.
	table = sheet.Notes.Table
	table.SetSelectionMap(map[tid.TID]bool{parent.ID(): true})
	ConvertToNonContainer(sheet, table)
	c.True(parent.Container(), "a container with children must not be converted")
	c.True(mgr.CanRedo(), "a conversion that changes nothing must not have recorded an edit that clears the redo stack")
}

// TestContainerConversionClosesEditorsAndUndoesTheKind verifies that converting a displayed equipment group to a plain
// item closes its editor first, and that undo brings back a group, not a physical container, even though the sheet is
// checked for modification, and so written out, in between. Undo and redo each discard an editor opened on the row in
// between, whose ID's kind the conversion changes, and redo brings back the legality class new equipment starts with.
func TestContainerConversionClosesEditorsAndUndoesTheKind(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	entity := gurps.NewEntity()
	group := gurps.NewEquipmentGroup(entity, nil)
	group.Name = "Kit"
	entity.CarriedEquipment = []*gurps.Equipment{group}
	var sheet *Sheet
	screen.Do(func() {
		sheet = NewSheet("test"+gurps.SheetExt, entity)
		DisplayNewDockable(sheet)
	})
	editorsOpen := func() int {
		var count int
		screen.Do(func() {
			count = len(AllMatchingDockables(func(d unison.Dockable) bool {
				e, ok := d.(*editor[*gurps.Equipment, *gurps.EquipmentEditData])
				return ok && e.target == group
			}))
		})
		return count
	}
	var mgr *unison.UndoManager
	screen.Do(func() {
		mgr = unison.UndoManagerFor(sheet.CarriedEquipment.Table)
		EditEquipment(sheet, group, true)
	})
	c.NotNil(mgr, "the sheet must have an undo manager")
	c.Equal(1, editorsOpen(), "the group's editor must be open")

	screen.Do(func() {
		table := sheet.CarriedEquipment.Table
		table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		ConvertToNonContainer(sheet, table)
	})
	c.False(group.Container(), "the group must have become a plain item")
	c.Equal("4", group.LegalityClass, "the item must start out with the default legality class")
	c.Equal(0, editorsOpen(), "converting must close the group's editor")

	screen.Do(func() { EditEquipment(sheet, group, true) })
	c.Equal(1, editorsOpen(), "the item's editor must be open")
	screen.Do(func() { _ = sheet.Modified() })
	screen.Do(mgr.Undo)
	c.True(group.IsGroup(), "undo must bring back a group, not a physical container")
	c.Equal(0, editorsOpen(), "undo must discard the editor opened on the item")
	c.Equal("", group.LegalityClass, "undo must take the legality class away again")

	screen.Do(func() { EditEquipment(sheet, group, true) })
	c.Equal(1, editorsOpen(), "the group's editor must be open")
	screen.Do(mgr.Redo)
	c.False(group.Container(), "redo must convert the group again")
	c.Equal("4", group.LegalityClass)
	c.Equal(0, editorsOpen(), "redo must discard the editor opened on the group")
}

// newTestTemplateWithEquipment returns a template dockable holding the equipment.
func newTestTemplateWithEquipment(equipment ...*gurps.Equipment) *Template {
	data := gurps.NewTemplate()
	data.Equipment = equipment
	return newTestTemplateDockable("Source", data)
}

// newTestSheetWithEquipment returns a sheet dockable carrying the equipment.
func newTestSheetWithEquipment(t *testing.T, equipment ...*gurps.Equipment) *Sheet {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	for _, one := range equipment {
		one.SetDataOwner(sheet.Entity())
	}
	sheet.Entity().CarriedEquipment = equipment
	sheet.Rebuild(true)
	return sheet
}

// equipmentConversions returns whether each of the conversion commands is available for the selected equipment.
func equipmentConversions(table *unison.Table[*Node[*gurps.Equipment]], selected ...*gurps.Equipment) (toChoice, toGroup, toContainer, toNonContainer bool) {
	selection := make(map[tid.TID]bool)
	for _, one := range selected {
		selection[one.ID()] = true
	}
	table.SetSelectionMap(selection)
	return table.CanPerformCmd(table, ConvertToChoiceContainerItemID),
		table.CanPerformCmd(table, ConvertToGroupContainerItemID),
		table.CanPerformCmd(table, ConvertToContainerItemID),
		table.CanPerformCmd(table, ConvertToNonContainerItemID)
}

// TestEquipmentConversionCommandsFollowTheSelection verifies that each kind of equipment offers exactly the
// conversions it can make, and that only a template offers to make a choice.
func TestEquipmentConversionCommandsFollowTheSelection(t *testing.T) {
	c := check.New(t)
	item := gurps.NewEquipment(nil, nil, false)
	backpack := gurps.NewEquipment(nil, nil, true)
	group := gurps.NewEquipmentGroup(nil, nil)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	template := newTestTemplateWithEquipment(item, backpack, group, choice)
	table := template.Equipment.Table

	toChoice, toGroup, toContainer, toNonContainer := equipmentConversions(table, item)
	c.False(toChoice || toGroup, "an item can only become a container")
	c.True(toContainer)
	c.False(toNonContainer)

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, backpack)
	c.False(toChoice, "a physical container must become a group before it can become a choice")
	c.True(toGroup)
	c.False(toContainer, "a physical container already is one")
	c.True(toNonContainer, "an empty physical container can become an item")

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, group)
	c.True(toChoice)
	c.False(toGroup, "a group already is one")
	c.True(toContainer, "a group can become a physical container")
	c.True(toNonContainer, "an empty group can become an item")

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, choice)
	c.False(toChoice, "a choice already is one")
	c.True(toGroup)
	c.False(toContainer, "a choice must become a group before it can become a physical container")
	c.False(toNonContainer, "a choice must become a group before it can become an item")

	sheetBackpack := gurps.NewEquipment(nil, nil, true)
	sheetGroup := gurps.NewEquipmentGroup(nil, nil)
	sheet := newTestSheetWithEquipment(t, sheetBackpack, sheetGroup)
	sheetTable := sheet.CarriedEquipment.Table
	_, toGroup, _, _ = equipmentConversions(sheetTable, sheetBackpack)
	c.True(toGroup, "a sheet may hold groups")
	toChoice, _, toContainer, _ = equipmentConversions(sheetTable, sheetGroup)
	c.False(toChoice, "only a template may hold choices")
	c.True(toContainer, "a sheet may hold physical containers")
}

// TestEquipmentGroupConversionWarnsAndIsUndoable verifies that converting a physical container to a group asks before
// removing what a group can't hold, that converting back asks nothing, and that undo restores each step.
func TestEquipmentGroupConversionWarnsAndIsUndoable(t *testing.T) {
	c := check.New(t)
	backpack := gurps.NewEquipment(nil, nil, true)
	backpack.BaseValue = "60"
	backpack.BaseWeight = "3 lb"
	sheet := newTestSheetWithEquipment(t, backpack)
	table := sheet.CarriedEquipment.Table
	table.SetSelectionMap(map[tid.TID]bool{backpack.ID(): true})
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)

	var asked int
	answer := false
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return answer
	})

	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing data must be confirmed first")
	c.True(backpack.IsPhysicalContainer(), "declining must leave the container alone")
	c.False(mgr.CanUndo(), "declining must not record an edit")

	answer = true
	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.True(backpack.IsGroup(), "the container must have become a group")
	c.Equal("", backpack.BaseValue, "the value must be removed")
	c.Equal("", backpack.BaseWeight, "the weight must be removed")

	mgr.Undo()
	c.True(backpack.IsPhysicalContainer(), "undo must turn it back into a physical container")
	c.Equal("60", backpack.BaseValue, "undo must restore the value")
	c.Equal("3 lb", backpack.BaseWeight, "undo must restore the weight")

	mgr.Redo()
	c.True(backpack.IsGroup(), "redo must convert it again")

	asked = 0
	table.SetSelectionMap(map[tid.TID]bool{backpack.ID(): true})
	table.PerformCmd(table, ConvertToContainerItemID)
	c.Equal(0, asked, "nothing is lost, so nothing may be asked")
	c.True(backpack.IsPhysicalContainer(), "the group must have become a physical container")
	mgr.Undo()
	c.True(backpack.IsGroup(), "undo must turn it back into a group")
}

// TestConvertToContainerTakesItemsAndGroupsTogether verifies that "Convert to Container" turns the selected items into
// containers and the selected groups into physical containers at once, as a single undoable edit, both on a sheet and
// in an equipment library, where the library's own command performs the shared conversion.
func TestConvertToContainerTakesItemsAndGroupsTogether(t *testing.T) {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	type equipmentTable = *unison.Table[*Node[*gurps.Equipment]]
	for _, one := range []struct {
		owner string
		open  func(t *testing.T, item, group *gurps.Equipment) (cmds *unison.Panel, table equipmentTable)
	}{
		{"sheet", func(t *testing.T, item, group *gurps.Equipment) (*unison.Panel, equipmentTable) {
			table := newTestSheetWithEquipment(t, item, group).CarriedEquipment.Table
			panel := table.AsPanel()
			return panel, table
		}},
		{"library", func(_ *testing.T, item, group *gurps.Equipment) (*unison.Panel, equipmentTable) {
			d := NewEquipmentTableDockable("Gear"+gurps.EquipmentExt, []*gurps.Equipment{group, item})
			return d.AsPanel(), d.table
		}},
	} {
		t.Run(one.owner, func(t *testing.T) {
			c := check.New(t)
			item := gurps.NewEquipment(nil, nil, false)
			group := gurps.NewEquipmentGroup(nil, nil)
			cmds, table := one.open(t, item, group)
			table.SetSelectionMap(map[tid.TID]bool{item.ID(): true, group.ID(): true})
			mgr := unison.UndoManagerFor(table)
			c.NotNil(mgr)

			c.True(cmds.CanPerformCmd(nil, ConvertToContainerItemID))
			cmds.PerformCmd(nil, ConvertToContainerItemID)
			c.True(item.IsPhysicalContainer(), "the item must have become a physical container")
			c.True(group.IsPhysicalContainer(), "the group must have become a physical container")

			mgr.Undo()
			c.False(item.Container(), "undo must turn the item back")
			c.True(group.IsGroup(), "the same undo must turn the group back")
			c.False(mgr.CanUndo(), "the conversion must have been a single edit")
		})
	}
}
