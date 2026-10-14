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
)

// The edits that undo by putting a serialized snapshot of a table back -- a drag, a move, a deletion, a duplication, an
// applied template -- used to replace every object in the list with a fresh one, while the edits that undo by changing
// the object they were made with -- an editor's apply, a toggle -- kept pointing at the old ones. Once an undo or redo
// of a snapshot edit had passed, the other kind changed objects no list held, and the change never showed. The tests
// here interleave the two kinds around one item and check that the live item is the one that changes.

// newTestSheetWithOtherEquipment returns a sheet whose other equipment list holds two items, the first valued at $10.
func newTestSheetWithOtherEquipment(t *testing.T) (sheet *Sheet, first, second *gurps.Equipment) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	first = gurps.NewEquipment(entity, nil, false)
	first.Name = "First"
	first.BaseValue = "10"
	second = gurps.NewEquipment(entity, nil, false)
	second.Name = "Second"
	entity.OtherEquipment = []*gurps.Equipment{first, second}
	sheet.Rebuild(true)
	return sheet, first, second
}

// liveEquipment returns the item with the ID in the list, or nil.
func liveEquipment(list []*gurps.Equipment, id tid.TID) *gurps.Equipment {
	for _, one := range list {
		if one.ID() == id {
			return one
		}
	}
	return nil
}

// applyValueEdit applies an editor that changes the item's value, recording the edit with the sheet's undo manager.
func applyValueEdit(sheet *Sheet, item *gurps.Equipment, value string) {
	e, _ := buildEditorContent(sheet, item, initEquipmentEditor(false))
	e.editorData.BaseValue = value
	e.applyEdits()
}

// dragSecond simulates dragging the item to the end of the sheet's other equipment list, the way unison drives a drop
// within a table: the undo data is collected, the model is rearranged and the drop completion is called.
func dragSecond(c check.Checker, sheet *Sheet, item, other *gurps.Equipment) {
	c.Helper()
	table := sheet.OtherEquipment.Table
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[*gurps.Equipment])
	c.True(ok, "the table must have its provider recorded")
	move := provider.DropShouldMoveData(table, table)
	c.True(move, "a drag within a table must be a move")
	undo := willDropCallback(table, table, move)
	c.NotNil(undo, "the drop must be undoable")
	sheet.Entity().OtherEquipment = []*gurps.Equipment{other, item}
	table.SyncToModel()
	table.SetSelectionMap(map[tid.TID]bool{item.ID(): true})
	didDropCallback(undo, table, table, move)
}

// checkEditSurvivesTheRearrangement undoes and redoes an edit of the item's value to $20 and the rearrangement that
// put it second before the edit, in the order a user would: the edit back, the rearrangement back, the rearrangement
// again, the edit again.
func checkEditSurvivesTheRearrangement(c check.Checker, sheet *Sheet, mgr *unison.UndoManager, id tid.TID) {
	c.Helper()
	entity := sheet.Entity()
	live := func() *gurps.Equipment { return liveEquipment(entity.OtherEquipment, id) }
	c.Equal("20", live().BaseValue, "the edit must have changed the value")
	mgr.Undo()
	c.Equal("10", live().BaseValue, "undoing the edit must put the value back")
	c.Equal(id, entity.OtherEquipment[1].ID(), "undoing the edit must leave the item second")
	mgr.Undo()
	c.Equal(id, entity.OtherEquipment[0].ID(), "undoing the rearrangement must put the item back first")
	mgr.Redo()
	c.Equal(id, entity.OtherEquipment[1].ID(), "redoing the rearrangement must put the item second again")
	c.Equal("10", live().BaseValue, "redoing the rearrangement must leave the value as it was")
	mgr.Redo()
	c.Equal("20", live().BaseValue, "redoing the edit must change the item that is in the list")
}

func TestRedoOfAnEditReachesTheItemADragPutBack(t *testing.T) {
	c := check.New(t)
	sheet, first, second := newTestSheetWithOtherEquipment(t)
	mgr := unison.UndoManagerFor(sheet.OtherEquipment.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	dragSecond(c, sheet, first, second)
	c.Equal(first.ID(), sheet.Entity().OtherEquipment[1].ID(), "the drag must have put the item second")
	applyValueEdit(sheet, sheet.Entity().OtherEquipment[1], "20")
	checkEditSurvivesTheRearrangement(c, sheet, mgr, first.ID())
}

func TestRedoOfAnEditReachesTheItemAMovePutBack(t *testing.T) {
	c := check.New(t)
	sheet, first, _ := newTestSheetWithOtherEquipment(t)
	table := sheet.OtherEquipment.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	table.SetSelectionMap(map[tid.TID]bool{first.ID(): true})
	MoveSelection(table, MoveDown)
	c.Equal(first.ID(), sheet.Entity().OtherEquipment[1].ID(), "the move must have put the item second")
	applyValueEdit(sheet, sheet.Entity().OtherEquipment[1], "20")
	checkEditSurvivesTheRearrangement(c, sheet, mgr, first.ID())
}

func TestUndoOfAnEditReachesTheItemADeletionPutBack(t *testing.T) {
	c := check.New(t)
	sheet, first, _ := newTestSheetWithOtherEquipment(t)
	entity := sheet.Entity()
	table := sheet.OtherEquipment.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	applyValueEdit(sheet, first, "20")
	c.Equal("20", first.BaseValue, "the edit must have changed the value")
	table.SetSelectionMap(map[tid.TID]bool{first.ID(): true})
	DeleteSelection(table, true)
	c.Equal(1, len(entity.OtherEquipment), "the item must be gone")
	mgr.Undo()
	c.Equal(2, len(entity.OtherEquipment), "undoing the deletion must put the item back")
	c.Equal("20", entity.OtherEquipment[0].BaseValue, "the item must come back with the edit")
	mgr.Undo()
	c.Equal("10", entity.OtherEquipment[0].BaseValue, "undoing the edit must reach the item that was put back")
	mgr.Redo()
	c.Equal("20", entity.OtherEquipment[0].BaseValue, "redoing the edit must reach the item that was put back")
}

func TestRedoOfAnEditReachesTheItemACreationPutBack(t *testing.T) {
	c := check.New(t)
	sheet, _, _ := newTestSheetWithOtherEquipment(t)
	entity := sheet.Entity()
	table := sheet.OtherEquipment.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	created := gurps.NewEquipment(entity, nil, false)
	created.Name = "Created"
	created.BaseValue = "10"
	InsertItems(sheet, table, entity.OtherEquipmentList, entity.SetOtherEquipmentList,
		func(_ *unison.Table[*Node[*gurps.Equipment]]) []*Node[*gurps.Equipment] { return table.RootRows() }, created)
	c.Equal(3, len(entity.OtherEquipment), "the item must have been added")
	applyValueEdit(sheet, created, "20")
	live := func() *gurps.Equipment { return liveEquipment(entity.OtherEquipment, created.ID()) }
	c.Equal("20", live().BaseValue, "the edit must have changed the value")
	mgr.Undo()
	c.Equal("10", live().BaseValue, "undoing the edit must put the value back")
	mgr.Undo()
	c.Nil(live(), "undoing the creation must take the item out")
	mgr.Redo()
	c.Equal("10", live().BaseValue, "redoing the creation must put the item back")
	mgr.Redo()
	c.Equal("20", live().BaseValue, "redoing the edit must reach the item that was put back")
}

func TestUndoOfAToggleReachesTheItemAMovePutBack(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	first := gurps.NewEquipment(entity, nil, false)
	first.Name = "First"
	second := gurps.NewEquipment(entity, nil, false)
	second.Name = "Second"
	entity.CarriedEquipment = []*gurps.Equipment{first, second}
	sheet.Rebuild(true)
	table := sheet.CarriedEquipment.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	live := func() *gurps.Equipment { return liveEquipment(entity.CarriedEquipment, first.ID()) }
	table.SetSelectionMap(map[tid.TID]bool{first.ID(): true})
	toggleEquipped(sheet, table)
	c.False(live().Equipped, "the toggle must have unequipped the item")
	table.SetSelectionMap(map[tid.TID]bool{first.ID(): true})
	MoveSelection(table, MoveDown)
	c.Equal(first.ID(), entity.CarriedEquipment[1].ID(), "the move must have put the item second")
	mgr.Undo()
	mgr.Undo()
	c.True(live().Equipped, "undoing the toggle must reach the item the move's undo put back")
	mgr.Redo()
	mgr.Redo()
	c.False(live().Equipped, "redoing the toggle must reach the item the move's redo put back")
}

// An editor left open while the item it edits is put back by an undo must still apply to the item on the page.
func TestAnEditorLeftOpenAcrossAnUndoAppliesToTheItemOnThePage(t *testing.T) {
	c := check.New(t)
	sheet, first, second := newTestSheetWithOtherEquipment(t)
	entity := sheet.Entity()
	mgr := unison.UndoManagerFor(sheet.OtherEquipment.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	e, _ := buildEditorContent(sheet, first, initEquipmentEditor(false))
	dragSecond(c, sheet, first, second)
	mgr.Undo()
	c.Equal(first.ID(), entity.OtherEquipment[0].ID(), "undoing the drag must put the item back first")
	e.editorData.BaseValue = "20"
	e.applyEdits()
	c.Equal("20", liveEquipment(entity.OtherEquipment, first.ID()).BaseValue,
		"the editor must apply to the item on the page")
}

// The tables inside an editor put their lists back the same way, so a modifier's editor must reach the modifier a
// move's undo and redo put back.
func TestRedoOfAModifierEditReachesTheModifierAMovePutBack(t *testing.T) {
	c := check.New(t)
	sheet, first, _ := newTestSheetWithOtherEquipment(t)
	entity := sheet.Entity()
	fine := gurps.NewEquipmentModifier(entity, nil, false)
	fine.Name = "Fine"
	cheap := gurps.NewEquipmentModifier(entity, nil, false)
	cheap.Name = "Cheap"
	first.Modifiers = []*gurps.EquipmentModifier{fine, cheap}
	gurps.AttachModifiers(first, first.Modifiers)
	e, content := buildEditorContent(sheet, first, initEquipmentEditor(false))
	var panel *equipmentModifiersPanel
	for _, child := range content.Children() {
		if p, ok := child.Self.(*equipmentModifiersPanel); ok {
			panel = p
		}
	}
	c.NotNil(panel, "the editor must hold the modifiers panel")
	mgr := unison.UndoManagerFor(panel.table)
	c.True(mgr == e.undoMgr, "the modifiers table must use the editor's undo manager")
	modifiers := func() []*gurps.EquipmentModifier { return e.editorData.Modifiers }
	editedID := modifiers()[0].ID()
	live := func() *gurps.EquipmentModifier {
		for _, one := range modifiers() {
			if one.ID() == editedID {
				return one
			}
		}
		return nil
	}
	panel.table.SetSelectionMap(map[tid.TID]bool{editedID: true})
	MoveSelection(panel.table, MoveDown)
	c.Equal(editedID, modifiers()[1].ID(), "the move must have put the modifier second")
	me, _ := buildEditorContent(e, live(), initEquipmentModifierEditor)
	me.editorData.Name = "Very Fine"
	me.applyEdits()
	c.Equal("Very Fine", live().Name, "the edit must have renamed the modifier")
	mgr.Undo()
	c.Equal("Fine", live().Name, "undoing the edit must put the name back")
	mgr.Undo()
	c.Equal(editedID, modifiers()[0].ID(), "undoing the move must put the modifier back first")
	mgr.Redo()
	c.Equal(editedID, modifiers()[1].ID(), "redoing the move must put the modifier second again")
	mgr.Redo()
	c.Equal("Very Fine", live().Name, "redoing the edit must reach the modifier that is in the list")
}

// A list whose nodes share an ID -- which nothing on load rejects -- gives a restore no way to tell which object each
// of them came from, so it must leave them as deserialized rather than hand one node's object to another: an editor
// open on one of them must not overwrite the other.
func TestAnEditorOnAnItemSharingAnIDDoesNotOverwriteTheOtherAfterAnUndo(t *testing.T) {
	c := check.New(t)
	sheet, first, second := newTestSheetWithOtherEquipment(t)
	entity := sheet.Entity()
	second.TID = first.TID
	second.BaseValue = "15"
	third := gurps.NewEquipment(entity, nil, false)
	third.Name = "Third"
	entity.OtherEquipment = append(entity.OtherEquipment, third)
	sheet.Rebuild(true)
	table := sheet.OtherEquipment.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	e, _ := buildEditorContent(sheet, second, initEquipmentEditor(false))
	table.SetSelectionMap(map[tid.TID]bool{third.ID(): true})
	MoveSelection(table, MoveUp)
	c.Equal("Third", entity.OtherEquipment[1].Name, "the move must have put the third item second")
	mgr.Undo()
	c.Equal("Third", entity.OtherEquipment[2].Name, "undoing the move must put the third item back last")
	e.editorData.BaseValue = "20"
	e.applyEdits()
	c.Equal("First", entity.OtherEquipment[0].Name, "the first item must keep its name")
	c.Equal("10", entity.OtherEquipment[0].BaseValue, "the first item must keep its value")
	c.Equal("Second", entity.OtherEquipment[1].Name, "the second item must keep its name")
}
