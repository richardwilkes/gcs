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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestPoolThresholdAddAndDeleteAreUndoable verifies that adding a threshold to a pool with the button of the pool's row
// and deleting one with the threshold's own button are each recorded as a single undo edit of the whole set of
// definitions, that the rows are rebuilt to match after each step, and that the guard against deleting the last
// threshold follows the count.
func TestPoolThresholdAddAndDeleteAreUndoable(t *testing.T) {
	c := check.New(t)
	defs := testAttrDefs("st", "hp")
	pool := defs.Set["hp"]
	pool.Type = attribute.Pool
	pool.Thresholds = []*gurps.PoolThreshold{{State: "Reeling", KeyPrefix: "r"}}
	d := newTestAttributeSettingsDockableFor(defs)
	initTestSettingsContent(&d.undoableSettingsDockable)
	expandAllAttributes(d)
	rows := uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "the last threshold may not be deleted")

	toolbarButtonWithTip(t, d.rows()[1].AsPanel(), "Add pool threshold").ClickCallback()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "the threshold is added to the pool")
	c.Equal("Undo Add Pool Threshold", d.undoMgr.UndoTitle())
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows), "the rows are rebuilt")
	if len(rows) != 2 {
		return
	}
	c.True(rows[0].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.True(rows[1].deleteButton.Enabled(), "with two thresholds, either may be deleted")
	c.NotEqual("", rows[1].threshold.KeyPrefix, "the added threshold is given a target key prefix")
	c.NotNil(d.targetMgr.Find(rows[1].threshold.KeyPrefix+"state"), "the added threshold's state field is registered")

	d.undoMgr.Undo()
	c.Equal([]string{"Reeling"}, thresholdStates(d.model.Set["hp"]), "undo removes the added threshold")
	c.False(d.undoMgr.CanUndo(), "the add is a single edit")
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "undoing back to one threshold restores the guard")

	d.undoMgr.Redo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "redo adds it back")
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(2, len(rows))

	rows[0].deleteButton.ClickCallback()
	c.Equal(1, len(d.model.Set["hp"].Thresholds), "the threshold is deleted")
	c.Equal("", d.model.Set["hp"].Thresholds[0].State, "the first threshold is the one deleted")
	c.Equal("Undo Delete Pool Threshold", d.undoMgr.UndoTitle())
	rows = uxtest.PanelsOfType[*thresholdSettingsPanel](d.AsPanel())
	c.Equal(1, len(rows))
	c.False(rows[0].deleteButton.Enabled(), "deleting down to one threshold restores the guard")

	d.undoMgr.Undo()
	c.Equal(2, len(d.model.Set["hp"].Thresholds), "undo restores the deleted threshold")
	c.Equal("Reeling", d.model.Set["hp"].Thresholds[0].State, "in its original position")
	c.Equal([]string{"st", "hp"}, attrDefIDs(d.model), "the definitions are untouched throughout")
}
