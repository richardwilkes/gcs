// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponParry(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"0",
		"-1",
		"10",
		"0U",
		"-1U",
		"9U",
		"0F",
		"-2F",
		"8F",
		"0FU",
		"-2FU",
		"8FU",
		"No",
	} {
		c.Equal(s, gurps.ParseWeaponParry(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
		{"0 (x5)", "0"},
		{"0U / 0", "0U"},
		{"0U/ 0", "0U"},
		{"13 (x5)", "13"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponParry(one.input).String(), "test %d", i)
	}
}

// TestWeaponParryResolveParryTypeDefault verifies that a parry-type weapon default does not have the +3 and the
// entity's parry bonus applied to it twice. SkillLevelFast() already turns a parry-type default into a parry level, so
// a weapon defaulting to "Cloak Parry" must resolve to the same value as one defaulting to the "Cloak" skill itself.
func TestWeaponParryResolveParryTypeDefault(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)

	// A skill-type default halves the skill level, then adds the +3 and the entity's parry bonus: 14/2 + 3 + 1.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.SkillID)}
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "skill-type default")

	// A parry-type default has already been through that conversion, so it must yield the same parry, not 15.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.ParryID)}
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "parry-type default")

	// Both kinds of default together still resolve to the same parry.
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.ParryID),
	}
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "both defaults")

	// The weapon's own parry modifier and the default's modifier still apply on top.
	w.Parry.Modifier = fxp.Two
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.ParryID)}
	w.Defaults[0].Modifier = -fxp.One
	c.Equal("12", w.Parry.Resolve(w, nil).String(), "parry-type default with modifiers")
}

// TestWeaponParryResolveBlockTypeDefault verifies that a block-type weapon default contributes to a parry what a plain
// skill default to the same skill would. A weapon's Defaults list feeds the parry and block calculations alike, so a
// block-type default reaches the parry calculation; converting the block level it already carries into a parry level
// halved the skill a second time and folded the block bonus into the parry.
func TestWeaponParryResolveBlockTypeDefault(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)

	// A block-type default must yield the same parry as the skill-type default it is derived from: 14/2 + 3 + 1.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "block-type default")

	// Mixing it with a skill-type default changes nothing, since both resolve to the same parry.
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.BlockID),
	}
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "both defaults")

	// The weapon's own parry modifier and the default's modifier still apply on top: 11 - 1 + 2.
	w.Parry.Modifier = fxp.Two
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	w.Defaults[0].Modifier = -fxp.One
	c.Equal("12", w.Parry.Resolve(w, nil).String(), "block-type default with modifiers")
}

// TestWeaponParryResolveAppliesParryBonus verifies that the parry calculation applies the entity's parry bonus and not
// its block bonus, whichever kind of default it resolves from. The two bonuses are equal in the default fixture, so
// this uses one where they differ.
func TestWeaponParryResolveAppliesParryBonus(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeaponWithBonuses(c, 2, 1) // +2 parry, +1 block

	// Every kind of default resolves to the same parry, built from the parry bonus alone: 14/2 + 3 + 2.
	for _, defaultType := range []string{gurps.SkillID, gurps.ParryID, gurps.BlockID} {
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
		c.Equal("12", w.Parry.Resolve(w, nil).String(), "%q default uses the parry bonus", defaultType)
	}
}
