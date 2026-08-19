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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestUndoOfDeleteSurvivesSwitchColumnDisappearing verifies that deleting the only switchable row -- which drops the
// switch column and so forces the sheet to rebuild the traits page list from scratch -- doesn't sever undo from the
// table the user is actually looking at.
func TestUndoOfDeleteSurvivesSwitchColumnDisappearing(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "the switch column must start out present")

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(sheet.Traits.Table, true)
	c.Equal(0, len(entity.Traits), "the trait must be gone from the entity")
	c.Equal(-1, sheet.Traits.Table.LastRowIndex(), "the trait must be gone from the table")
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the switch column must go away with the last switchable row")

	c.True(mgr.CanUndo(), "the deletion must be undoable")
	mgr.Undo()
	c.Equal(1, len(entity.Traits), "undo must put the trait back into the entity")
	c.Equal(0, sheet.Traits.Table.LastRowIndex(), "undo must put the row back into the visible table")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "undo must bring the switch column back")

	c.True(mgr.CanRedo(), "the deletion must be redoable")
	mgr.Redo()
	c.Equal(0, len(entity.Traits), "redo must remove the trait from the entity again")
	c.Equal(-1, sheet.Traits.Table.LastRowIndex(), "redo must remove the row from the visible table again")
}

// TestUndoOfDropSurvivesSwitchColumnAppearing verifies the other direction: dropping the first switchable row onto a
// sheet brings the switch column into view, which replaces the traits page list mid-drop. Both the undo edit that gets
// registered afterward and the undo itself have to end up pointed at the table that took its place.
func TestUndoOfDropSurvivesSwitchColumnAppearing(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = nil
	sheet.Rebuild(true)
	stale := sheet.Traits.Table
	mgr := unison.UndoManagerFor(stale)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.NotEqual(gurps.TraitSwitchColumn, stale.Columns[0].ID, "the switch column must start out absent")

	// Drive the drop the way unison does: collect the undo data, put the dropped row into the model and select it,
	// then hand off to the drop completion.
	undo := willDropCallback(nil, stale, false)
	c.NotNil(undo, "the drop must be undoable")
	trait := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{trait}
	stale.SyncToModel()
	stale.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	didDropCallback(undo, nil, stale, false)

	c.NotEqual(stale, sheet.Traits.Table, "the drop must have replaced the traits table")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "the drop must bring in the switch column")
	c.True(mgr.CanUndo(), "the drop must have registered an undo edit")

	mgr.Undo()
	c.Equal(0, len(entity.Traits), "undo must take the trait back out of the entity")
	c.Equal(-1, sheet.Traits.Table.LastRowIndex(), "undo must take the row back out of the visible table")
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "undo must take the switch column away again")

	c.True(mgr.CanRedo(), "the drop must be redoable")
	mgr.Redo()
	c.Equal(1, len(entity.Traits), "redo must put the trait back into the entity")
	c.Equal(0, sheet.Traits.Table.LastRowIndex(), "redo must put the row back into the visible table")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "redo must bring the switch column back")
}

// TestLiveTableResolvesReplacedTables verifies the lookup that keeps the above working: a table that its owner has
// since replaced resolves to the replacement, while a table that is still the current one -- or that has no owner to
// ask, as in an editor or a library list -- resolves to itself.
func TestLiveTableResolvesReplacedTables(t *testing.T) {
	c := check.New(t)
	sheet, _ := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	stale := sheet.Traits.Table
	c.Equal(stale, liveTable(stale), "a table that is still in use must resolve to itself")

	entity.Traits = nil
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits.Table, "losing the switch column must replace the traits table")
	c.Equal(sheet.Traits.Table, liveTable(stale), "a replaced table must resolve to the one that took its place")

	// Tables belonging to something that never replaces them, such as an editor's, have no owner recorded and must be
	// left alone.
	_, orphan := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	c.Equal(orphan, liveTable(orphan), "a table with no recorded owner must resolve to itself")
}

// columnsMatchProvider returns true if the columns a table is showing are the ones its provider now calls for, i.e.
// that the list it belongs to doesn't have to be built anew for the user to see the right set of columns.
func columnsMatchProvider[T gurps.Node[T]](table *unison.Table[*Node[T]]) bool {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	return ok && !columnsOutOfSync(provider.ColumnIDs(), table.Columns)
}

// syncCounter counts the deep syncs that pass over it, making the number of times a single edit re-syncs an entire
// sheet observable. The optional onSync hook runs with each of them, so that a test can also look at the state the
// sheet is being brought up to date against.
type syncCounter struct {
	unison.Panel
	onSync func()
	count  int
}

// Sync implements Syncer.
func (s *syncCounter) Sync() {
	s.count++
	if s.onSync != nil {
		s.onSync()
	}
}

// installSyncCounter adds a sync counter to a sheet, where its deep syncs will reach it. It goes directly on the sheet,
// whose layout is a single column and which a rebuild never touches, so that it survives the page lists being replaced.
func installSyncCounter(sheet *Sheet) *syncCounter {
	counter := &syncCounter{}
	counter.Self = counter
	sheet.AsPanel().AddChild(counter)
	return counter
}

// TestUndoOfColumnChangingEditSyncsTheSheetOnlyOnce verifies that an undo that has to bring a column into or out of
// view doesn't pay for the update twice. Putting the data back and marking the table as modified would recalculate the
// entity and re-sync every table, and the rebuild that the changed column set then requires does all of that over
// again. Only one of the two is warranted, and on a sheet with many rows the difference is the whole cost of the edit.
func TestUndoOfColumnChangingEditSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	counter := installSyncCounter(sheet)

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(sheet.Traits.Table, true)
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the switch column must go away with the last switchable row")

	counter.count = 0
	mgr.Undo()
	c.Equal(1, len(entity.Traits), "undo must put the trait back into the entity")
	c.Equal(0, sheet.Traits.Table.LastRowIndex(), "undo must put the row back into the visible table")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "undo must bring the switch column back")
	c.Equal(1, counter.count, "an undo that brings a column into view must sync the sheet exactly once")

	counter.count = 0
	mgr.Redo()
	c.Equal(0, len(entity.Traits), "redo must remove the trait from the entity again")
	c.Equal(-1, sheet.Traits.Table.LastRowIndex(), "redo must remove the row from the visible table again")
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"redo must take the switch column away again")
	c.Equal(1, counter.count, "a redo that takes a column out of view must sync the sheet exactly once")
}

// TestUndoOfOrdinaryEditSyncsTheSheetOnlyOnce verifies the other half of the same choice: an undo that leaves the
// column set alone doesn't rebuild, so marking the table as modified is what has to do the syncing -- and it still has
// to happen, exactly once.
func TestUndoOfOrdinaryEditSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Plain"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	counter := installSyncCounter(sheet)

	table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(table, true)
	c.Equal(0, len(entity.Traits), "the trait must be gone from the entity")

	counter.count = 0
	mgr.Undo()
	c.Equal(table, sheet.Traits.Table, "an unchanged column set must leave the traits table in place")
	c.Equal(1, len(entity.Traits), "undo must put the trait back into the entity")
	c.Equal(0, table.LastRowIndex(), "undo must put the row back into the visible table")
	c.Equal(1, counter.count, "an undo that leaves the columns alone must still sync the sheet exactly once")
}

// TestUndoOfColumnChangingEditBumpsTheModificationTimestamp verifies that an undo which has to bring a column into or
// out of view still reports the sheet as having been changed. Such an undo rebuilds the sheet instead of marking the
// table as modified, and the rebuild does everything marking as modified would do except bump the modification
// timestamp, so the undo has to bump it itself. Otherwise undoing the deletion of the last switchable row would leave
// the sheet claiming it hadn't been touched since it was last saved, while the same undo on any other row would not.
func TestUndoOfColumnChangingEditBumpsTheModificationTimestamp(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(sheet.Traits.Table, true)
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the switch column must go away with the last switchable row")

	entity.ModifiedOn = jio.Time{}
	mgr.Undo()
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID, "undo must bring the switch column back")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "an undo that rebuilds the sheet must bump the modification timestamp")

	entity.ModifiedOn = jio.Time{}
	mgr.Redo()
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"redo must take the switch column away again")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "a redo that rebuilds the sheet must bump the modification timestamp")
}

// TestUndoOfOrdinaryEditBumpsTheModificationTimestamp verifies the other half of that: an undo that leaves the column
// set alone marks the table as modified rather than rebuilding, which has to bump the modification timestamp as well,
// so that both paths look the same to the user.
func TestUndoOfOrdinaryEditBumpsTheModificationTimestamp(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Plain"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

	table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(table, true)
	c.Equal(0, len(entity.Traits), "the trait must be gone from the entity")

	entity.ModifiedOn = jio.Time{}
	mgr.Undo()
	c.Equal(table, sheet.Traits.Table, "an unchanged column set must leave the traits table in place")
	c.Equal(1, len(entity.Traits), "undo must put the trait back into the entity")
	c.NotEqual(jio.Time{}, entity.ModifiedOn,
		"an undo that leaves the columns alone must bump the modification timestamp, too")
}

// newSwitchableSkill returns a non-container skill carrying a single switchable +1 ST bonus, so that a sheet's skills
// list needs the switch column while the skill is in it.
func newSwitchableSkill(entity *gurps.Entity, name string) *gurps.Skill {
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = name
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetSwitchable(true)
	bonus.SetOwner(skill)
	skill.Features = gurps.Features{bonus}
	return skill
}

// TestUndoSpanningEveryListSyncsTheSheetOnlyOnce verifies that an undo that puts every one of a sheet's lists back at
// once -- what "Sync With All Sources" registers -- updates the sheet a single time, even when more than one of those
// lists has to gain or lose its switch column. Restoring each list and reporting it right away would recalculate the
// entity and re-sync every table on the sheet up to six times over for the one undo.
func TestUndoSpanningEveryListSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the traits switch column must start out absent")
	c.NotEqual(gurps.SkillSwitchColumn, sheet.Skills.Table.Columns[0].ID,
		"the skills switch column must start out absent")

	traitCount := len(entity.Traits)
	skillCount := len(entity.Skills)
	before := newSheetTablesUndoData(sheet)
	entity.Traits = append(entity.Traits, newSwitchableTrait(entity, "Claws"))
	entity.Skills = append(entity.Skills, newSwitchableSkill(entity, "Body Sense"))
	sheet.Traits.Table.SyncToModel()
	sheet.Skills.Table.SyncToModel()
	sheet.Rebuild(true)
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the switchable trait must bring in the traits switch column")
	c.Equal(gurps.SkillSwitchColumn, sheet.Skills.Table.Columns[0].ID,
		"the switchable skill must bring in the skills switch column")
	after := newSheetTablesUndoData(sheet)
	counter := installSyncCounter(sheet)

	counter.count = 0
	before.Apply()
	c.Equal(traitCount, len(entity.Traits), "undo must take the trait back out of the entity")
	c.Equal(skillCount, len(entity.Skills), "undo must take the skill back out of the entity")
	c.Equal(1, counter.count, "an undo spanning every list must sync the sheet exactly once")
	c.NotEqual(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"undo must take the traits switch column away again")
	c.NotEqual(gurps.SkillSwitchColumn, sheet.Skills.Table.Columns[0].ID,
		"undo must take the skills switch column away again")
	c.True(columnsMatchProvider(sheet.Traits.Table), "the traits table's columns must match its provider after undo")
	c.True(columnsMatchProvider(sheet.Skills.Table), "the skills table's columns must match its provider after undo")

	counter.count = 0
	after.Apply()
	c.Equal(traitCount+1, len(entity.Traits), "redo must put the trait back into the entity")
	c.Equal(skillCount+1, len(entity.Skills), "redo must put the skill back into the entity")
	c.Equal(1, counter.count, "a redo spanning every list must sync the sheet exactly once")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"redo must bring the traits switch column back")
	c.Equal(gurps.SkillSwitchColumn, sheet.Skills.Table.Columns[0].ID,
		"redo must bring the skills switch column back")
	c.True(columnsMatchProvider(sheet.Traits.Table), "the traits table's columns must match its provider after redo")
	c.True(columnsMatchProvider(sheet.Skills.Table), "the skills table's columns must match its provider after redo")
}

// newSwitchableEquipment returns a non-container piece of equipment carrying a single switchable +1 ST bonus, so that
// the sheet's list holding it needs the switch column while it is there.
func newSwitchableEquipment(entity *gurps.Entity, name string) *gurps.Equipment {
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = name
	eqp.Features = gurps.Features{switchableSTBonus(nil)}
	return eqp
}

// TestUndoOfMoveBetweenEquipmentListsSyncsTheSheetOnlyOnce verifies that undoing a move of rows from one of a sheet's
// lists to another -- what dragging between the carried and other equipment lists and the "Move to Other Equipment"
// command both register -- updates the sheet a single time, and only once both lists are back in place. Restoring the
// list the rows landed in and reporting it right away would recalculate the entity and re-sync the whole sheet while
// the moved item was in neither list, and restoring the list it came from would then pay for the same update again.
func TestUndoOfMoveBetweenEquipmentListsSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	eqp := newSwitchableEquipment(entity, "Powered Armor")
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)
	mgr := unison.UndoManagerFor(sheet.CarriedEquipment.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.Equal(gurps.EquipmentSwitchColumn, sheet.CarriedEquipment.Table.Columns[1].ID,
		"the switchable item must bring the switch column into the carried equipment list")
	c.NotEqual(gurps.EquipmentSwitchColumn, sheet.OtherEquipment.Table.Columns[0].ID,
		"the other equipment list must start out without the switch column")

	sheet.CarriedEquipment.Table.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
	moveSelectedEquipment(sheet.CarriedEquipment.Table, sheet.OtherEquipment.Table)
	c.Equal(0, len(entity.CarriedEquipment), "the move must take the item out of the carried equipment")
	c.Equal(1, len(entity.OtherEquipment), "the move must put the item into the other equipment")
	c.NotEqual(gurps.EquipmentSwitchColumn, sheet.CarriedEquipment.Table.Columns[1].ID,
		"the switch column must leave the carried equipment list with the item")
	c.Equal(gurps.EquipmentSwitchColumn, sheet.OtherEquipment.Table.Columns[0].ID,
		"the switch column must arrive in the other equipment list with the item")
	c.True(mgr.CanUndo(), "the move must be undoable")

	counter := installSyncCounter(sheet)
	counter.onSync = func() {
		c.Equal(1, len(entity.CarriedEquipment)+len(entity.OtherEquipment),
			"the item must be in exactly one of the two lists whenever the sheet is brought up to date")
	}
	mgr.Undo()
	c.Equal(1, len(entity.CarriedEquipment), "undo must put the item back into the carried equipment")
	c.Equal(0, len(entity.OtherEquipment), "undo must take the item back out of the other equipment")
	c.Equal(1, counter.count, "an undo that puts both lists back must sync the sheet exactly once")
	c.True(columnsMatchProvider(sheet.CarriedEquipment.Table),
		"the carried equipment list's columns must match its provider after undo")
	c.True(columnsMatchProvider(sheet.OtherEquipment.Table),
		"the other equipment list's columns must match its provider after undo")

	counter.count = 0
	c.True(mgr.CanRedo(), "the move must be redoable")
	mgr.Redo()
	c.Equal(0, len(entity.CarriedEquipment), "redo must take the item back out of the carried equipment")
	c.Equal(1, len(entity.OtherEquipment), "redo must put the item back into the other equipment")
	c.Equal(1, counter.count, "a redo that puts both lists back must sync the sheet exactly once")
	c.True(columnsMatchProvider(sheet.CarriedEquipment.Table),
		"the carried equipment list's columns must match its provider after redo")
	c.True(columnsMatchProvider(sheet.OtherEquipment.Table),
		"the other equipment list's columns must match its provider after redo")
}

// TestUndoOfDragBetweenSheetsLeavesTheSourceSheetAlone verifies the invariant that lets a drag's undo report itself
// just once for both of the lists it puts back: they always belong to the same owner. A drag from one sheet to another
// is a copy rather than a move -- only tables within a single dockable move their data (see the providers'
// DropShouldMoveData) -- so the source sheet never changes, there is nothing to put back into the list the rows came
// from, and undoing the drop updates the sheet they landed on alone, exactly once.
func TestUndoOfDragBetweenSheetsLeavesTheSourceSheetAlone(t *testing.T) {
	c := check.New(t)
	source := newTestSheetForTemplate(t)
	dest := newTestSheetForTemplate(t)
	sourceEntity := source.Entity()
	destEntity := dest.Entity()
	sourceEntity.CarriedEquipment = []*gurps.Equipment{newSwitchableEquipment(sourceEntity, "Powered Armor")}
	source.Rebuild(true)
	provider, ok := source.CarriedEquipment.Table.ClientData()[TableProviderClientKey].(TableProvider[*gurps.Equipment])
	c.True(ok, "the table must have its provider recorded")
	move := provider.DropShouldMoveData(source.CarriedEquipment.Table, dest.CarriedEquipment.Table)
	c.False(move, "a drag between two sheets must be a copy rather than a move")

	// Drive the drop the way unison does: collect the undo data, put the dropped row into the destination's model and
	// select it, then hand off to the drop completion.
	undo := willDropCallback(source.CarriedEquipment.Table, dest.CarriedEquipment.Table, move)
	c.NotNil(undo, "the drop must be undoable")
	c.Nil(undo.BeforeData.From, "a copy leaves nothing to put back into the list the rows came from")
	dropped := newSwitchableEquipment(destEntity, "Powered Armor")
	destEntity.CarriedEquipment = []*gurps.Equipment{dropped}
	dest.CarriedEquipment.Table.SyncToModel()
	dest.CarriedEquipment.Table.SetSelectionMap(map[tid.TID]bool{dropped.ID(): true})
	didDropCallback(undo, source.CarriedEquipment.Table, dest.CarriedEquipment.Table, move)
	c.Equal(gurps.EquipmentSwitchColumn, dest.CarriedEquipment.Table.Columns[1].ID,
		"the dropped item must bring the switch column into the destination sheet")

	sourceCounter := installSyncCounter(source)
	destCounter := installSyncCounter(dest)
	undo.BeforeData.Apply()
	c.Equal(0, len(destEntity.CarriedEquipment), "undo must take the item back out of the destination sheet")
	c.Equal(1, len(sourceEntity.CarriedEquipment), "undo must leave the source sheet's item where it is")
	c.Equal(1, destCounter.count, "undoing the drop must sync the destination sheet exactly once")
	c.Equal(0, sourceCounter.count, "undoing the drop must not touch the source sheet at all")
	c.True(columnsMatchProvider(dest.CarriedEquipment.Table),
		"the destination's carried equipment list's columns must match its provider after undo")
}

// TestUndoOfApplyTemplateSyncsTheSheetOnlyOnce verifies the same thing for the other multi-table undo: undoing an
// "Apply Template" puts five of the sheet's lists back and then rebuilds the sheet, and that rebuild is all the
// reporting the undo needs. It still has to bump the modification timestamp, since that is the one thing marking the
// sheet as modified would have done that the rebuild does not.
func TestUndoOfApplyTemplateSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traitCount := len(entity.Traits)
	template := newTestTemplateWithBodyType("Template Body")
	c.True(template.applyTemplateToSheet(sheet, true), "the template must be applied")
	c.Equal(traitCount+1, len(entity.Traits), "the template's trait must have been added to the sheet")
	c.True(sheet.undoMgr.CanUndo(), "applying a template must be undoable")
	counter := installSyncCounter(sheet)

	counter.count = 0
	entity.ModifiedOn = jio.Time{}
	sheet.undoMgr.Undo()
	c.Equal(traitCount, len(entity.Traits), "undo must take the template's trait back out of the entity")
	c.Equal(traitCount, len(sheet.Traits.Table.RootRows()), "undo must take the row back out of the visible table")
	c.Equal(1, counter.count, "undoing an apply template must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "undo must still bump the modification timestamp")

	counter.count = 0
	sheet.undoMgr.Redo()
	c.Equal(traitCount+1, len(entity.Traits), "redo must put the template's trait back into the entity")
	c.Equal(1, counter.count, "redoing an apply template must sync the sheet exactly once")
}

// listAttachedToSheet returns true if the given page list is currently part of the sheet's panel tree, i.e. that it is
// one of the lists the sheet is showing. Several of a sheet's lists are only carried on the page while they have rows,
// and which ones those are is decided when the sheet creates its lists.
func listAttachedToSheet(sheet *Sheet, list unison.Paneler) bool {
	target := list.AsPanel()
	return sheet.AsPanel().HasInSelfOrDescendants(func(one *unison.Panel) bool { return one == target })
}

// TestUndoRestoresAConditionallyPresentList verifies that an undo puts back the parts of a sheet that only creating the
// lists anew can. The melee weapons list is carried on the page solely while the character has a melee weapon, so
// deleting the only trait that supplies one takes the whole list off the page. The traits table's own columns are
// untouched by that deletion, so an undo that only rebuilt when the restored table's columns had changed would settle
// for marking the traits table as modified, leaving the melee weapons list missing until some unrelated later edit
// happened to rebuild the sheet.
func TestUndoRestoresAConditionallyPresentList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.Equal(1, sheet.MeleeWeapons.Table.RootRowCount(), "the trait's weapon must be in the melee weapons list")
	c.True(listAttachedToSheet(sheet, sheet.MeleeWeapons), "the melee weapons list must start out on the page")

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	DeleteSelection(sheet.Traits.Table, true)
	c.Equal(0, len(entity.Traits), "the trait must be gone from the entity")
	c.Equal(0, sheet.MeleeWeapons.Table.RootRowCount(), "the weapon must go with the trait that supplied it")
	c.False(listAttachedToSheet(sheet, sheet.MeleeWeapons),
		"an empty melee weapons list must come off the page")

	c.True(mgr.CanUndo(), "the deletion must be undoable")
	mgr.Undo()
	c.Equal(1, len(entity.Traits), "undo must put the trait back into the entity")
	c.Equal(1, sheet.MeleeWeapons.Table.RootRowCount(), "undo must bring the weapon back")
	c.True(listAttachedToSheet(sheet, sheet.MeleeWeapons), "undo must put the melee weapons list back on the page")

	c.True(mgr.CanRedo(), "the deletion must be redoable")
	mgr.Redo()
	c.Equal(0, len(entity.Traits), "redo must remove the trait from the entity again")
	c.False(listAttachedToSheet(sheet, sheet.MeleeWeapons),
		"redo must take the melee weapons list back off the page")
}
