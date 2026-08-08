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
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestEquipmentRatedStrength verifies that equipment reports its user-settable rated ST. Unlike traits, skills and
// spells -- whose RatedStrength always returns 0 -- equipment genuinely carries one, and weapon strength calculations
// depend on getting it back.
func TestEquipmentRatedStrength(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	equipment := NewEquipment(e, nil, false)
	c.Equal(fxp.Int(0), equipment.RatedStrength(), "equipment with no rated ST reports 0")
	equipment.RatedST = fxp.FromInteger(12)
	c.Equal(fxp.FromInteger(12), equipment.RatedStrength(), "equipment reports the rated ST it was given")
}

// TestNewEquipmentFromFileAttachesContainerData verifies that loading an equipment list attaches the owner to
// container rows as well as to leaf rows, so that a container's own weapons get their sub-version fixups applied and
// its modifiers can resolve nameable placeholders.
func TestNewEquipmentFromFileAttachesContainerData(t *testing.T) {
	c := check.New(t)
	// container_with_own_data.eqp holds a leveled container that carries a weapon written before weapon
	// sub-versioning existed (no "sv" key) and a modifier with a nameable placeholder, plus a child item with the
	// same content.
	rows, err := NewEquipmentFromFile(os.DirFS("testdata"), "container_with_own_data.eqp")
	c.NoError(err)
	c.Equal(1, len(rows), "the list has a single top-level row")
	if len(rows) != 1 {
		return
	}
	container := rows[0]
	c.Equal(true, container.Container(), "the top-level row is a container")
	c.Equal(1, len(container.Children), "the container has a child")
	if len(container.Weapons) != 1 || len(container.Modifiers) != 1 || len(container.Children) != 1 {
		return
	}
	child := container.Children[0]

	// The container's own weapon must be attached and migrated exactly like the child's.
	for _, one := range []*Equipment{container, child} {
		w := one.Weapons[0]
		c.Equal(WeaponOwner(one), w.Owner, one.Name+": the weapon's owner is set")
		c.Equal(currentWeaponSubVersion, w.SubVersion, one.Name+": the weapon is stamped with the current sub-version")
		c.Equal("", w.Damage.Base, one.Name+": the leveled-damage migration cleared the flat base damage")
		c.Equal("1d", w.Damage.BaseLeveled, one.Name+": the leveled-damage migration moved the base damage")
		c.Equal(w, w.Damage.Owner, one.Name+": the weapon's damage points back at the weapon")
	}

	// The container's own modifiers must be attached so nameable placeholders resolve.
	c.Equal("Oak plating", container.Modifiers[0].NameWithReplacements(),
		"the container's modifier resolves the container's replacements")
	c.Equal("Iron plating", child.Modifiers[0].NameWithReplacements(),
		"the child's modifier resolves the child's replacements")
}

// TestEquipmentEditDataResolvesModifierNameables verifies that the modifier copies an equipment editor works with are
// pointed at their equipment, so their nameable placeholders resolve. Without that, the accessors fall back to the raw
// text and the editor's modifier rows read "@Material@" instead of the replacement the user chose.
func TestEquipmentEditDataResolvesModifierNameables(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	eqp.Replacements = map[string]string{"Material": "Steel"}
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.LocalNotes = "Forged from @Material@"
	eqp.Modifiers = []*EquipmentModifier{mod}

	// What the editor is handed when it opens.
	var edit EquipmentEditData
	edit.CopyFrom(eqp)
	c.Equal(1, len(edit.Modifiers))
	c.True(eqp == edit.Modifiers[0].OwningEquipment(), "the copy points at the equipment being edited")
	c.Equal("Steel plating", edit.Modifiers[0].NameWithReplacements())
	c.Equal("Forged from Steel", edit.Modifiers[0].LocalNotesWithReplacements())

	// The copy must be a copy: editing it leaves the original alone.
	edit.Modifiers[0].Name = "@Material@ mesh"
	c.Equal("@Material@ plating", mod.Name, "the original modifier is untouched")

	// What applying the editor's data back produces.
	target := NewEquipment(entity, nil, false)
	edit.ApplyTo(target)
	c.Equal(1, len(target.Modifiers))
	c.True(target == target.Modifiers[0].OwningEquipment(), "the applied copy points at the equipment it was applied to")
	c.Equal("Steel mesh", target.Modifiers[0].NameWithReplacements())
}
