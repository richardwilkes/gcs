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

// TestBodySettingsAddButtonAppendsAHitLocation verifies that the body type's add button appends a hit location to the
// body, with a row of its own, as a single undo edit.
func TestBodySettingsAddButtonAppendsAHitLocation(t *testing.T) {
	c := check.New(t)
	d := newTestBodySettingsDockable("a", "b")
	panels := uxtest.PanelsOfType[*bodySettingsPanel](d.AsPanel())
	c.Equal(1, len(panels), "the body has a panel")
	if len(panels) != 1 {
		return
	}
	toolbarButtonWithTip(t, panels[0].AsPanel(), "Add hit location").ClickCallback()
	c.Equal(3, len(d.model.Locations), "a location is added to the body")
	if len(d.model.Locations) == 3 {
		added := d.model.Locations[2]
		c.Equal([]string{"a", "b", added.LocID}, hitLocationIDs(d.model), "after the others")
		c.NotEqual("", added.KeyPrefix, "with a target key prefix of its own")
		c.NotNil(d.targetMgr.Find(added.KeyPrefix+"id"), "whose ID field is registered")
	}
	c.Equal(3, len(uxtest.PanelsOfType[*hitLocationSettingsPanel](d.AsPanel())), "the location has a row")
	d.undoMgr.Undo()
	c.Equal([]string{"a", "b"}, hitLocationIDs(d.model), "undo takes the location away")
	c.False(d.undoMgr.CanUndo(), "the add is a single edit")
}
