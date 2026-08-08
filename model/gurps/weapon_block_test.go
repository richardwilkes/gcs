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

func TestWeaponBlock(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"0",
		"-1",
		"10",
		"No",
	} {
		c.Equal(s, gurps.ParseWeaponBlock(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponBlock(one.input).String(), "test %d", i)
	}
}

// TestWeaponBlockResolveBlockTypeDefault verifies that a block-type weapon default does not have the +3 and the
// entity's block bonus applied to it twice. SkillLevelFast() already turns a block-type default into a block level, so
// a weapon defaulting to "Cloak Block" must resolve to the same value as one defaulting to the "Cloak" skill itself.
func TestWeaponBlockResolveBlockTypeDefault(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)

	// A skill-type default halves the skill level, then adds the +3 and the entity's block bonus: 14/2 + 3 + 1.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.SkillID)}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "skill-type default")

	// A block-type default has already been through that conversion, so it must yield the same block, not 15.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "block-type default")

	// Both kinds of default together still resolve to the same block.
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.BlockID),
	}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "both defaults")

	// The weapon's own block modifier and the default's modifier still apply on top.
	w.Block.Modifier = fxp.Two
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.BlockID)}
	w.Defaults[0].Modifier = -fxp.One
	c.Equal("12", w.Block.Resolve(w, nil).String(), "block-type default with modifiers")
}

// TestWeaponBlockResolveParryTypeDefault verifies that a parry-type weapon default contributes to a block what a plain
// skill default to the same skill would. A weapon's Defaults list feeds the parry and block calculations alike, so a
// parry-type default reaches the block calculation; converting the parry level it already carries into a block level
// halved the skill a second time and folded the parry bonus into the block.
func TestWeaponBlockResolveParryTypeDefault(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)

	// A parry-type default must yield the same block as the skill-type default it is derived from: 14/2 + 3 + 1.
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.ParryID)}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "parry-type default")

	// Mixing it with a skill-type default changes nothing, since both resolve to the same block.
	w.Defaults = []*gurps.SkillDefault{
		newDefenseTestDefault(gurps.SkillID),
		newDefenseTestDefault(gurps.ParryID),
	}
	c.Equal("11", w.Block.Resolve(w, nil).String(), "both defaults")

	// The weapon's own block modifier and the default's modifier still apply on top: 11 - 1 + 2.
	w.Block.Modifier = fxp.Two
	w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.ParryID)}
	w.Defaults[0].Modifier = -fxp.One
	c.Equal("12", w.Block.Resolve(w, nil).String(), "parry-type default with modifiers")
}

// TestWeaponBlockResolveMinSTPenaltyAtSkillScale verifies that a skill-level adjustment -- here the penalty for
// wielding a weapon whose minimum ST exceeds the character's -- is folded into the skill level before it is halved
// into a block, whichever kind of default the block resolves from. A defense-type default arrives already converted,
// so adding the adjustment to it counted the adjustment at twice the weight the skill-type path gives it.
func TestWeaponBlockResolveMinSTPenaltyAtSkillScale(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeapon(c)
	w.Strength = gurps.WeaponStrength{Min: fxp.FromInteger(12)} // -2 to skill for the ST 10 character

	// (14 - 2)/2 + 3 + 1, not (14/2 + 3 + 1) - 2.
	for _, defaultType := range []string{gurps.SkillID, gurps.BlockID, gurps.ParryID} {
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
		c.Equal("10", w.Block.Resolve(w, nil).String(), "%q default", defaultType)
	}
}

// TestWeaponBlockResolveThisWeaponSkillBonusAtSkillScale verifies the same for a skill bonus aimed at this weapon, the
// other half of the base adjustment, with both kinds of default present so that an over-weighted adjustment on the
// defense-type one would win the max in Resolve() and inflate the displayed block.
func TestWeaponBlockResolveThisWeaponSkillBonusAtSkillScale(t *testing.T) {
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
		newDefenseTestDefault(gurps.BlockID),
	}

	// (14 + 1)/2 + 3 + 1 for both defaults, so neither can out-bid the other: 11, not 12.
	c.Equal("11", w.Block.Resolve(w, nil).String(), "a this-weapon skill bonus must not count double")
}

// TestWeaponBlockResolveAppliesBlockBonus verifies that the block calculation applies the entity's block bonus and not
// its parry bonus, whichever kind of default it resolves from. The two bonuses are equal in the default fixture, so
// this uses one where they differ.
func TestWeaponBlockResolveAppliesBlockBonus(t *testing.T) {
	c := check.New(t)
	w := newDefenseTestWeaponWithBonuses(c, 2, 1) // +2 parry, +1 block

	// Every kind of default resolves to the same block, built from the block bonus alone: 14/2 + 3 + 1.
	for _, defaultType := range []string{gurps.SkillID, gurps.BlockID, gurps.ParryID} {
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
		c.Equal("11", w.Block.Resolve(w, nil).String(), "%q default uses the block bonus", defaultType)
	}
}
