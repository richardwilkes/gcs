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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMarkModifiedRecalculates verifies that telling the sheet something changed brings the entity's derived state up
// to date. Everything MarkModified goes on to do -- the panels, the tables and the calculator -- displays that state.
// It used to be refreshed only as a side effect of the tab asking whether the sheet had unsaved changes, which
// recalculated the entity on its way to hashing it; hashing no longer does that, since recalculating rewrites part of
// what gets saved.
func TestMarkModifiedRecalculates(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	entity.Recalculate()
	before := skill.LevelData.Level

	entity.Attributes.Set["dx"].Adjustment = fxp.Four
	sheet.MarkModified(nil)

	c.Equal(before+fxp.Four, skill.LevelData.Level, "the skill level must reflect the raised attribute")
}

// TestNewItemCommandUsesTheLiveList verifies that the "New Trait" command adds its item to the list the user is
// looking at, even after the sheet has had to replace that list. A list can only change its set of columns by being
// built anew -- which is what the arrival of the first switchable feature forces, since it brings the switch column in
// -- so a command that captured the list when the sheet was created would afterwards be creating items in an orphan:
// the model would gain the trait, but an orphaned table can't find the undo manager, so the insertion wouldn't be
// undoable and the user's next undo would silently take back the edit before it instead, and the new row would be
// neither selected nor scrolled into view in the list that is actually on screen.
func TestNewItemCommandUsesTheLiveList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Traits = nil
	sheet.Rebuild(true)
	stale := sheet.Traits
	mgr := sheet.UndoManager()
	c.NotNil(mgr, "the sheet must have an undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")

	// Giving the sheet its first switchable feature brings in the switch column, which replaces the traits list.
	claws := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{claws}
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits, "gaining the switch column must have replaced the traits list")

	sheet.AsPanel().PerformCmd(nil, NewTraitItemID)
	c.Equal(2, len(entity.Traits), "the command must have added a trait to the entity")
	added := entity.Traits[1]
	c.NotEqual(claws.ID(), added.ID(), "the added trait must be the new one, not the one that was already there")
	c.Equal(2, sheet.Traits.Table.RootRowCount(), "the new row must be in the list that is on screen")
	c.True(sheet.Traits.Table.CopySelectionMap()[added.ID()],
		"the new row must be selected in the list that is on screen")

	c.True(mgr.CanUndo(), "creating an item must be undoable")
	mgr.Undo()
	c.Equal(1, len(entity.Traits), "undo must take the new trait back out of the entity")
	c.Equal(claws.ID(), entity.Traits[0].ID(), "undo must leave the trait that was already there alone")
	c.Equal(1, sheet.Traits.Table.RootRowCount(), "undo must take the row back out of the list that is on screen")
}
