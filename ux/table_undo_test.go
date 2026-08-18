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
