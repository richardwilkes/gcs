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

// TestEquipmentCloneAttachesModifiersToTheClone verifies that a clone's modifier copies are pointed at the clone rather
// than at the equipment they were cloned from, so that they resolve their nameable placeholders with the clone's
// replacements. Duplicating a nested row in a standalone equipment list never re-attaches the data owner, so a
// mispointed modifier would keep rendering from the original's replacements for the rest of the session.
func TestEquipmentCloneAttachesModifiersToTheClone(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	eqp.Replacements = map[string]string{"Material": "Steel"}
	mod := NewEquipmentModifier(entity, nil, true)
	mod.Name = "@Material@ plating"
	child := NewEquipmentModifier(entity, mod, false)
	child.Name = "@Material@ rivets"
	mod.Children = []*EquipmentModifier{child}
	eqp.Modifiers = []*EquipmentModifier{mod}
	eqp.SetDataOwner(entity)

	clone := eqp.Clone(LibraryFile{}, entity, nil, false)
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 || len(clone.Modifiers[0].Children) != 1 {
		return
	}
	c.True(clone == clone.Modifiers[0].OwningEquipment(), "the clone's modifier points at the clone")
	c.True(clone == clone.Modifiers[0].Children[0].OwningEquipment(), "the clone's child modifier points at the clone")

	// Diverging the clone's replacements must move the clone's modifiers only.
	clone.Replacements["Material"] = "Bronze"
	c.Equal("Bronze plating", clone.Modifiers[0].NameWithReplacements())
	c.Equal("Bronze rivets", clone.Modifiers[0].Children[0].NameWithReplacements())
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements(), "the original's modifier is unaffected")
	c.Equal("Steel rivets", eqp.Modifiers[0].Children[0].NameWithReplacements(),
		"the original's child modifier is unaffected")
}

// TestEquipmentCloneMigratesLegacyModifierReplacements verifies that cloning equipment whose modifiers still carry the
// legacy per-modifier replacements migrates them into the clone rather than into the equipment being cloned.
func TestEquipmentCloneMigratesLegacyModifierReplacements(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.Replacements = map[string]string{"Material": "Steel"}
	eqp.Modifiers = []*EquipmentModifier{mod}

	clone := eqp.Clone(LibraryFile{}, entity, nil, false)
	c.Equal(0, len(eqp.Replacements), "the equipment being cloned isn't mutated")
	c.Equal(1, len(clone.Modifiers))
	if len(clone.Modifiers) != 1 {
		return
	}
	c.Equal("Steel", clone.Replacements["Material"], "the legacy replacements were migrated into the clone")
	c.Equal("Steel plating", clone.Modifiers[0].NameWithReplacements())

	// The original's modifier still carries its own replacements, so it resolves once it is attached normally.
	eqp.SetDataOwner(entity)
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements())
}

// TestEquipmentEditDataCapturesMigratedReplacements verifies that populating an editor from equipment whose modifiers
// still carry the legacy per-modifier replacements captures the migrated replacements in the editor's snapshot.
// Without that, applying the editor's data back writes the pre-migration snapshot and silently drops them.
func TestEquipmentEditDataCapturesMigratedReplacements(t *testing.T) {
	c := check.New(t)
	entity := NewEntity()
	eqp := NewEquipment(entity, nil, false)
	eqp.Name = "Armor"
	mod := NewEquipmentModifier(entity, nil, false)
	mod.Name = "@Material@ plating"
	mod.Replacements = map[string]string{"Material": "Steel"}
	eqp.Modifiers = []*EquipmentModifier{mod}

	var edit EquipmentEditData
	edit.CopyFrom(eqp)
	c.Equal("Steel", edit.Replacements["Material"], "the editor's snapshot has the migrated replacements")

	edit.ApplyTo(eqp)
	c.Equal("Steel", eqp.Replacements["Material"], "applying the editor's data preserves the migrated replacements")
	c.Equal(1, len(eqp.Modifiers))
	if len(eqp.Modifiers) != 1 {
		return
	}
	c.Equal("Steel plating", eqp.Modifiers[0].NameWithReplacements())
}
