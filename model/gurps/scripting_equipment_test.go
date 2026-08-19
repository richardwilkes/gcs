// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptEquipmentEquipped verifies that the `equipped` property a script sees means "affecting the character".
// Equipment in the other equipment list contributes nothing to the character, yet its equipped flag is typically true,
// since new equipment starts out equipped and nothing clears the flag when an item is created in or moved to that
// list, so the property must not simply report that flag.
func TestScriptEquipmentEquipped(t *testing.T) {
	c := check.New(t)

	equipped := func(eqp *Equipment) string {
		v, err := runScript(0, "String(self.equipped)", ScriptArg{
			Name:  "self",
			Value: func(r *goja.Runtime) any { return newScriptEquipment(r, eqp) },
		})
		c.NoError(err, "the script should run")
		return v
	}

	e := NewEntity()
	carried := NewEquipment(e, nil, false)
	carried.Name = "Sword"
	e.CarriedEquipment = append(e.CarriedEquipment, carried)
	other := NewEquipment(e, nil, false)
	other.Name = "Spare Sword"
	otherBag := NewEquipment(e, nil, true)
	otherBag.Name = "Storage Chest"
	otherChild := NewEquipment(e, otherBag, false)
	otherChild.Name = "Stored Shield"
	otherBag.Children = []*Equipment{otherChild}
	e.OtherEquipment = append(e.OtherEquipment, other, otherBag)
	carriedBag := NewEquipment(e, nil, true)
	carriedBag.Name = "Backpack"
	carriedChild := NewEquipment(e, carriedBag, false)
	carriedChild.Name = "Carried Shield"
	carriedBag.Children = []*Equipment{carriedChild}
	e.CarriedEquipment = append(e.CarriedEquipment, carriedBag)
	e.Recalculate()

	c.Equal("true", equipped(carried), "an equipped item in the carried list is equipped")

	carried.Equipped = false
	c.Equal("false", equipped(carried), "clearing the equipped flag on a carried item makes it unequipped")
	carried.Equipped = true

	c.Equal("false", equipped(other), "an equipped item in the other list is not equipped")
	c.Equal("false", equipped(otherChild),
		"an equipped child of an equipped container in the other list is not equipped")
	c.Equal("true", equipped(carriedChild), "an equipped child of an equipped carried container is equipped")

	// Equipment with no owning entity -- a library list, a template or a loot sheet -- has no other list to be told
	// apart from, so its equipped flag is all there is to go on.
	c.Equal("true", equipped(NewEquipment(nil, nil, false)), "an equipped item with no owning entity is equipped")
}
