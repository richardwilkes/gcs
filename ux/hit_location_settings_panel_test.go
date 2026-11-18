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

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestHitLocationButtonsAddASubTableAndRemoveTheLocation verifies that a hit location's add button gives it a sub-table
// holding one location, after which it offers no other, and that its delete button removes it from the table that owns
// it, which may not lose its last location, each as a single undo edit.
func TestHitLocationButtonsAddASubTableAndRemoveTheLocation(t *testing.T) {
	c := check.New(t)
	d := newTestBodySettingsDockable("a", "b")
	rows := func() []*hitLocationSettingsPanel { return uxtest.PanelsOfType[*hitLocationSettingsPanel](d.AsPanel()) }
	c.True(rows()[0].addButton.Enabled(), "a location without a sub-table may be given one")
	c.True(rows()[0].deleteButton.Enabled(), "a location with a sibling may be deleted")

	rows()[0].addButton.ClickCallback()
	if sub := d.model.Locations[0].SubTable; sub == nil {
		c.NotNil(sub, "the location is given a sub-table")
	} else {
		c.Equal(1, len(sub.Locations), "holding one location")
	}
	c.Equal("Undo Add Sub-Table", d.undoMgr.UndoTitle())
	all := rows()
	c.Equal(3, len(all), "the sub-table's location has a row")
	if len(all) != 3 {
		return
	}
	c.False(all[0].addButton.Enabled(), "a location with a sub-table may not be given another")
	c.False(all[1].deleteButton.Enabled(), "the sub-table's only location may not be deleted")

	all[2].deleteButton.ClickCallback()
	c.Equal([]string{"a"}, hitLocationIDs(d.model), "the location is removed")
	c.Equal("Undo Remove Hit Location", d.undoMgr.UndoTitle())
	c.False(rows()[0].deleteButton.Enabled(), "the body's only location may not be deleted")

	d.undoMgr.Undo()
	c.Equal([]string{"a", "b"}, hitLocationIDs(d.model), "undo brings the location back")
	d.undoMgr.Undo()
	c.Nil(d.model.Locations[0].SubTable, "undo takes the sub-table away")
	c.False(d.undoMgr.CanUndo(), "each is a single edit")
}
