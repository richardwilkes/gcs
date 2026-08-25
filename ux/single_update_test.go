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

// These tests pin down the edits that used to report themselves more than once -- marking the sheet as modified and
// then rebuilding it, or rebuilding it once per table an edit touched -- so that each now updates the sheet a single
// time while still bumping its modification timestamp, which is the one thing a rebuild on its own leaves out.

// TestInsertItemsSyncsTheSheetOnlyOnce verifies that inserting an item into one of a sheet's lists updates the sheet
// once, by rebuilding it, rather than marking it as modified and then rebuilding it as well.
func TestInsertItemsSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := newSwitchableTrait(entity, "Claws")
	before := len(entity.Traits)
	entity.ModifiedOn = jio.Time{}
	counter := installSyncCounter(sheet)

	InsertItems(sheet, sheet.Traits.Table, entity.TraitList, entity.SetTraitList,
		func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
			return sheet.Traits.provider.RootRows()
		},
		trait)

	c.Equal(before+1, len(entity.Traits), "the item must be inserted into the entity")
	c.Equal(1, counter.count, "an insertion must sync the sheet exactly once")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the rebuild must bring the switch column in with the switchable item")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "an insertion must bump the modification timestamp")
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.True(mgr.CanUndo(), "the insertion must be undoable")
}

// TestSheetSettingsUpdatedSyncsTheSheetOnlyOnce verifies that a change to the sheet settings updates the sheet once,
// by rebuilding it, rather than marking it as modified and then rebuilding it as well.
func TestSheetSettingsUpdatedSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	for _, fullRebuild := range []bool{false, true} {
		entity.ModifiedOn = jio.Time{}
		counter := installSyncCounter(sheet)
		sheet.SheetSettingsUpdated(entity, fullRebuild)
		c.Equal(1, counter.count, "a sheet settings change must sync the sheet exactly once (fullRebuild=%v)",
			fullRebuild)
		c.NotEqual(jio.Time{}, entity.ModifiedOn,
			"a sheet settings change must bump the modification timestamp (fullRebuild=%v)", fullRebuild)
		sheet.AsPanel().RemoveChild(counter)
	}

	// Settings for some other entity are none of this sheet's business.
	counter := installSyncCounter(sheet)
	sheet.SheetSettingsUpdated(gurps.NewEntity(), true)
	c.Equal(0, counter.count, "another entity's settings change must not touch the sheet")
}

// TestDropIntoASheetSyncsTheSheetOnlyOnce verifies that a drop into one of a sheet's lists is reported by the single
// rebuild the drop completion performs, which also bumps the modification timestamp, and that the table's own drop
// notification can tell that it has been, so that it doesn't mark the sheet as modified on top of that. A drop into a
// template, by contrast, isn't reported by a rebuild, so the notification has to do it.
func TestDropIntoASheetSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	eqp := newSwitchableEquipment(entity, "Powered Armor")
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)
	from := sheet.CarriedEquipment.Table
	to := sheet.OtherEquipment.Table
	c.Equal(sheet, dropRebuilder(to), "a drop into a sheet's list must be reported by rebuilding the sheet")
	c.Equal(sheet, dropRebuilder(from), "a drop into a sheet's list must be reported by rebuilding the sheet")

	// Drive a move between the two lists the way unison's drop handling does, up to the drop completion.
	undo := willDropCallback(from, to, true)
	c.NotNil(undo, "the drop must be undoable")
	entity.CarriedEquipment = nil
	entity.OtherEquipment = []*gurps.Equipment{eqp}
	from.SyncToModel()
	to.SyncToModel()
	to.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
	entity.ModifiedOn = jio.Time{}
	counter := installSyncCounter(sheet)
	didDropCallback(undo, from, to, true)
	c.Equal(1, counter.count, "the drop completion must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "the drop completion must bump the modification timestamp")
	c.NotNil(dropRebuilder(liveTable(to)),
		"the drop notification must still be able to tell that the drop was reported by a rebuild")

	template := newTestTemplateWithBodyType("Humanoid")
	c.Nil(dropRebuilder(template.Traits.Table), "a drop into a template's list is not reported by a rebuild")
	c.Nil(dropRebuilder(nil), "no table means nothing to rebuild")
}

// TestRowEditsReportThemselvesOnce verifies that the row-level edits that report themselves by rebuilding the sheet --
// deleting and duplicating rows, here -- do so exactly once and bump the modification timestamp as they go, which a
// bare rebuild used to leave out, so that the sheet's "Modified" date stood still through them.
func TestRowEditsReportThemselvesOnce(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	before := len(entity.Traits)

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	entity.ModifiedOn = jio.Time{}
	counter := installSyncCounter(sheet)
	DuplicateSelection(sheet.Traits.Table)
	c.Equal(before+1, len(entity.Traits), "duplicating must add a row")
	c.Equal(1, counter.count, "duplicating must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "duplicating must bump the modification timestamp")

	sheet.Traits.Table.SetSelectionMap(map[tid.TID]bool{trait.ID(): true})
	entity.ModifiedOn = jio.Time{}
	counter.count = 0
	DeleteSelection(sheet.Traits.Table, true)
	c.Equal(before, len(entity.Traits), "deleting must remove the row")
	c.Equal(1, counter.count, "deleting must sync the sheet exactly once")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "deleting must bump the modification timestamp")
}
