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

	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestModifierHideNotes checks that a modifier with HideNotes set is listed in its owner's notes without its notes,
// while its full description and its weapon usage notes keep them.
func TestModifierHideNotes(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	allies := addTraitWithFeatures(e, "Allies")
	unwilling := NewTraitModifier(e, nil, false)
	unwilling.Name = "Unwilling"
	unwilling.LocalNotes = "Must be persuaded"
	unwilling.HideNotes = true
	unwilling.ShowNotesOnWeapon = true
	summonable := NewTraitModifier(e, nil, false)
	summonable.Name = "Summonable"
	summonable.LocalNotes = "Takes a turn"
	allies.SetModifiers([]*TraitModifier{unwilling, summonable})
	allies.Weapons = []*Weapon{NewWeapon(allies, true)}
	c.True(unwilling.Enabled(), "precondition: the unwilling modifier is enabled")
	c.True(summonable.Enabled(), "precondition: the summonable modifier is enabled")

	c.Equal("Unwilling; Summonable (Takes a turn)", allies.ModifierNotes(),
		"hidden notes are left out, while the modifier is still listed")
	c.Equal("Unwilling (Must be persuaded)", unwilling.FullDescription(), "the full description keeps hidden notes")
	c.Equal("Must be persuaded", allies.Weapons[0].Notes(), "weapon usage notes ignore HideNotes")

	sword := addCarriedEquipmentWithFeatures(e, "Broadsword")
	silvered := NewEquipmentModifier(e, nil, false)
	silvered.Name = "Silvered"
	silvered.LocalNotes = "Hurts werewolves"
	silvered.HideNotes = true
	sword.SetModifiers([]*EquipmentModifier{silvered})
	c.True(silvered.Enabled(), "precondition: the silvered modifier is enabled")

	c.Equal("Silvered", sword.ModifierNotes(), "equipment modifiers hide their notes too")
	c.Equal("Silvered (Hurts werewolves)", silvered.FullDescription(), "the full description keeps hidden notes")
}

// TestModifierHideNotesRoundTripAndHash checks that HideNotes and ShowInTitle on both kinds of modifier, and
// CostIgnoresLevel on a trait modifier, are each saved only when set, so that existing files re-save unchanged, load
// back, and are hashed, so that a library row differing only in one is reported as mismatched.
func TestModifierHideNotesRoundTripAndHash(t *testing.T) {
	c := check.New(t)
	traitMod := NewTraitModifier(nil, nil, false)
	traitMod.Name = "Unwilling"
	equipmentMod := NewEquipmentModifier(nil, nil, false)
	equipmentMod.Name = "Silvered"
	for _, one := range []struct {
		name string
		key  string
		set  func(bool)
		node Hashable
	}{
		{"trait modifier", "hide_notes", func(on bool) { traitMod.HideNotes = on }, traitMod},
		{"trait modifier", "show_in_title", func(on bool) { traitMod.ShowInTitle = on }, traitMod},
		{"trait modifier", "cost_ignores_level", func(on bool) { traitMod.CostIgnoresLevel = on }, traitMod},
		{"equipment modifier", "hide_notes", func(on bool) { equipmentMod.HideNotes = on }, equipmentMod},
		{"equipment modifier", "show_in_title", func(on bool) { equipmentMod.ShowInTitle = on }, equipmentMod},
	} {
		data, err := jio.Marshal(one.node)
		c.NoError(err)
		c.NotContains(string(data), one.key, "%s: an unset %s isn't saved", one.name, one.key)
		hash := Hash64(one.node)
		one.set(true)
		c.NotEqual(hash, Hash64(one.node), "%s: %s is source data and is hashed", one.name, one.key)
		data, err = jio.Marshal(one.node)
		c.NoError(err)
		c.Contains(string(data), `"`+one.key+`":true`, "%s", one.name)
	}

	data, err := jio.Marshal(traitMod)
	c.NoError(err)
	var restoredTraitMod TraitModifier
	c.NoError(jio.Unmarshal(data, &restoredTraitMod))
	c.True(restoredTraitMod.HideNotes && restoredTraitMod.ShowInTitle && restoredTraitMod.CostIgnoresLevel,
		"the trait modifier's flags load back")
	data, err = jio.Marshal(equipmentMod)
	c.NoError(err)
	var restoredEquipmentMod EquipmentModifier
	c.NoError(jio.Unmarshal(data, &restoredEquipmentMod))
	c.True(restoredEquipmentMod.HideNotes && restoredEquipmentMod.ShowInTitle,
		"the equipment modifier's flags load back")
}
