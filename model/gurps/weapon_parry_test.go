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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
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

// TestWeaponParryResolveMinSTPenaltyAtSkillScale verifies that a skill-level adjustment -- here the penalty for
// wielding a weapon whose minimum ST exceeds the character's -- is folded into the skill level before it is halved
// into a parry, whichever kind of default the parry resolves from. A defense-type default arrives already converted,
// so adding the adjustment to it counted the adjustment at twice the weight the skill-type path gives it.
func TestWeaponParryResolveMinSTPenaltyAtSkillScale(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	w.Strength = gurps.WeaponStrength{Min: fxp.FromInteger(12)} // -2 to skill for the ST 10 character

	// (14 - 2)/2 + 3 + 1, not (14/2 + 3 + 1) - 2.
	for _, defaultType := range []string{gurps.SkillID, gurps.ParryID, gurps.BlockID} {
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
		c.Equal("10", w.Parry.Resolve(w, nil).String(), "%q default", defaultType)
	}
}

// TestWeaponParryResolveOddMinSTPenaltyAtSkillScale is TestWeaponParryResolveMinSTPenaltyAtSkillScale with an odd skill
// level and an odd penalty, where merely halving the adjustment before applying it to an already-converted default
// would still lose a point that the skill-type path keeps.
func TestWeaponParryResolveOddMinSTPenaltyAtSkillScale(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	e := w.Entity()
	e.Skills[0].Points = fxp.FromInteger(20) // DX+5, i.e. level 15
	e.Recalculate()
	c.Equal(fxp.FromInteger(15), e.Skills[0].LevelData.Level, "the Cloak skill should be at level 15")
	w.Strength = gurps.WeaponStrength{Min: fxp.FromInteger(11)} // -1 to skill for the ST 10 character

	// (15 - 1)/2 + 3 + 1, which the odd skill level rounds up to the same 7 as 15/2 does.
	for _, defaultType := range []string{gurps.SkillID, gurps.ParryID, gurps.BlockID} {
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
		c.Equal("11", w.Parry.Resolve(w, nil).String(), "%q default", defaultType)
	}
}

// TestWeaponParryResolveThisWeaponSkillBonusAtSkillScale verifies the same for a skill bonus aimed at this weapon, the
// other half of the base adjustment. This is the layout the High Tech library's Tonfa uses -- a skill default paired
// with a matching parry-type default -- so an over-weighted adjustment on the defense-type default won the max in
// Resolve() and inflated the displayed parry.
func TestWeaponParryResolveThisWeaponSkillBonusAtSkillScale(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	owner, ok := w.Owner.(*gurps.Trait)
	c.True(ok, "the test weapon should be owned by a trait")
	bonus := gurps.NewSkillBonus()
	bonus.SelectionType = skillsel.ThisWeapon
	bonus.Amount = fxp.One
	owner.Features = append(owner.Features, bonus)
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.ParryID),
	}

	// (14 + 1)/2 + 3 + 1 for both defaults, so neither can out-bid the other: 11, not 12.
	c.Equal("11", w.Parry.Resolve(w, nil).String(), "a this-weapon skill bonus must not count double")
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
