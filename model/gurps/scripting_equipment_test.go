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
	"github.com/richardwilkes/gcs/v5/model/fxp"
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

// TestScriptEquipmentEquippedOnEditorClone verifies that the working clone the equipment editor makes to preview
// Extended Value and Extended Weight resolves scripts the same way the row it was cloned from does. The editor clones
// with the row's own parent, which is nil for a top-level row, so the clone is rooted in neither of the entity's
// equipment lists. Such a root has to be treated as carried, or a script reading self.equipped would answer one way on
// the sheet and another in the editor's preview.
func TestScriptEquipmentEquippedOnEditorClone(t *testing.T) {
	c := check.New(t)

	const script = "self.equipped ? 100 : 50"
	e := NewEntity()
	carried := NewEquipment(e, nil, false)
	carried.Name = "Sword"
	carried.BaseValue = script
	e.CarriedEquipment = append(e.CarriedEquipment, carried)
	other := NewEquipment(e, nil, false)
	other.Name = "Spare Sword"
	other.BaseValue = script
	otherBag := NewEquipment(e, nil, true)
	otherBag.Name = "Storage Chest"
	otherChild := NewEquipment(e, otherBag, false)
	otherChild.Name = "Stored Shield"
	otherChild.BaseValue = script
	otherBag.Children = []*Equipment{otherChild}
	e.OtherEquipment = append(e.OtherEquipment, other, otherBag)
	e.Recalculate()

	// This is how ux.cloneEquipmentWithOverlay builds the clone the editor previews with.
	cloneForEditor := func(eqp *Equipment) *Equipment {
		return eqp.Clone(eqp.Source.LibraryFile, eqp.DataOwner(), eqp.Parent(), true)
	}

	// The clone preserves the ID of the row it came from, and script results are cached per entity by (self ID, script
	// text), so the cache has to be discarded between the two resolutions or the second would just get the first's
	// answer back.
	baseValue := func(eqp *Equipment) fxp.Int {
		e.DiscardCaches()
		return eqp.ResolvedBaseValue()
	}

	c.True(carried.IsCarried(), "a top-level item in the carried list is carried")
	carriedClone := cloneForEditor(carried)
	c.True(carriedClone.IsCarried(), "the editor's clone of a top-level carried item is still carried")
	c.Equal(fxp.FromInteger(100), baseValue(carried), "the sheet resolves the carried item as equipped")
	c.Equal(baseValue(carried), baseValue(carriedClone),
		"the editor's preview of a carried item agrees with the sheet")

	c.False(otherChild.IsCarried(), "a nested item in the other list is not carried")
	otherChildClone := cloneForEditor(otherChild)
	c.False(otherChildClone.IsCarried(),
		"the editor's clone of a nested other item keeps its real parent chain, so it is still not carried")
	c.Equal(fxp.FromInteger(50), baseValue(otherChild), "the sheet resolves the nested other item as unequipped")
	c.Equal(baseValue(otherChild), baseValue(otherChildClone),
		"the editor's preview of a nested other item agrees with the sheet")

	// A top-level item in the other list is the one case the clone cannot get right: nothing on it says which list it
	// came from, and treating an unknown root as carried is what keeps every other case in agreement. The editor's
	// preview of such an item therefore reports it as equipped, which is what it did before the equipped property
	// started consulting the list at all.
	c.False(other.IsCarried(), "a top-level item in the other list is not carried")
	otherClone := cloneForEditor(other)
	c.True(otherClone.IsCarried(), "the editor's clone of a top-level other item is rooted in neither list")
	c.Equal(fxp.FromInteger(50), baseValue(other), "the sheet resolves the top-level other item as unequipped")
	c.Equal(fxp.FromInteger(100), baseValue(otherClone), "the editor's preview treats it as equipped")
}
