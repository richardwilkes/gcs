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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestWeaponDRDivisorBonus verifies that a DR divisor bonus adjusts the armor divisor. A percentage bonus must add a
// percentage of the divisor rather than replacing the divisor with that percentage of itself, which is what the
// tooltip ("+10% to armor divisor") describes.
func TestWeaponDRDivisorBonus(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		amount   fxp.Int
		percent  bool
		divisor  fxp.Int
		expected string
	}{
		{amount: fxp.One, divisor: fxp.Two, expected: "1d(3) cr"},
		{amount: fxp.NegOne, divisor: fxp.Two, expected: "1d cr"}, // A divisor of 1 isn't shown
		{amount: fxp.FromInteger(10), percent: true, divisor: fxp.Two, expected: "1d(2.2) cr"},
		{amount: fxp.Hundred, percent: true, divisor: fxp.Two, expected: "1d(4) cr"},
		{amount: fxp.FromInteger(-50), percent: true, divisor: fxp.Two, expected: "1d cr"},
		{amount: fxp.Hundred, percent: true, divisor: fxp.Half, expected: "1d cr"},
		{amount: fxp.FromInteger(10), percent: true, divisor: fxp.Half, expected: "1d(0.55) cr"},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponDRDivisorBonus)
		bonus.Amount = one.amount
		bonus.Percent = one.percent
		w := newWeaponWithBonuses(false, bonus)
		w.Damage.ArmorDivisor = one.divisor
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d", i)
	}
}

// TestWeaponPerLevelBaseDamageWithDiceMultiplier verifies that a per-level base damage specification carrying a dice
// multiplier is scaled by the level count exactly once. Dice evaluate as ((sum of Count dice) + Modifier) * Multiplier,
// so scaling Count and Multiplier both would apply the level count twice, turning a 3-level "1dx2" into "3dx6".
func TestWeaponPerLevelBaseDamageWithDiceMultiplier(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		baseLeveled string
		levels      int
		expected    string
	}{
		{baseLeveled: "1dx2", levels: 1, expected: "1dx2 cr"},
		{baseLeveled: "1dx2", levels: 3, expected: "3dx2 cr"},
		{baseLeveled: "1dx3", levels: 2, expected: "2dx3 cr"},
		{baseLeveled: "1d+1", levels: 3, expected: "3d+3 cr"},
		{baseLeveled: "1d", levels: 4, expected: "4d cr"},
		{baseLeveled: "2x2", levels: 3, expected: "6x2 cr"}, // 3 * (2 * 2) == 12 == 6 * 2
	} {
		w := newLeveledWeapon(one.levels)
		w.Damage.BaseLeveled = one.baseLeveled
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d (%s at %d levels)", i, one.baseLeveled, one.levels)
	}
}

// newLeveledWeapon builds an entity with a leveled trait at the given level that owns a single ranged weapon, so that
// the weapon's per-level base damage is the only thing contributing to its damage.
func newLeveledWeapon(levels int) *gurps.Weapon {
	e := gurps.NewEntity()
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Gadget"
	owner.CanLevel = true
	owner.Levels = fxp.FromInteger(levels)
	w := gurps.NewWeapon(owner, false)
	w.Damage.Base = "" // Leave only the per-level portion contributing to the damage
	owner.Weapons = []*gurps.Weapon{w}
	e.Traits = append(e.Traits, owner)
	e.Recalculate()
	return w
}

// TestWeaponDamageSpecIsScriptOrCompleteDice verifies that a damage field is only treated as a plain dice
// specification when the dice grammar consumes all of it. The dice parser stops at the first character it cannot use
// and silently discards the remainder, so a script expression written without spaces (which is how anyone writing
// arithmetic naturally writes it) must not be truncated to the dice specification it happens to start with.
func TestWeaponDamageSpecIsScriptOrCompleteDice(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		base     string
		expected string
	}{
		// Script expressions that start with something the dice parser would happily consume. Truncating instead of
		// evaluating these would yield "2 cr", "9 cr", "4 cr" and "3 cr", respectively.
		{base: "2*3", expected: "6 cr"},
		{base: "10-1-1", expected: "8 cr"},
		{base: "4/2", expected: "2 cr"},
		{base: "1+2*3", expected: "7 cr"},
		// Complete dice specifications must still be taken as dice rather than handed to the script engine, which
		// would fail to parse them.
		{base: "1d", expected: "1d cr"},
		{base: "1d6", expected: "1d cr"},
		{base: "2d6+1", expected: "2d+1 cr"},
		{base: "1d4-1", expected: "1d4-1 cr"},
		{base: "2x3", expected: "2x3 cr"},
		{base: "1d+", expected: "1d cr"}, // A dangling sign is an empty modifier to the dice parser
		{base: "+1d6", expected: "1d cr"},
		{base: "3", expected: "3 cr"},
		{base: "-2", expected: "-2 cr"},
		{base: "0", expected: "cr"},
	} {
		w := newWeaponWithBonuses(false)
		w.Damage.Base = one.base
		c.Equal(one.expected, w.Damage.ResolvedDamage(nil), "test %d (%s)", i, one.base)
	}
}

// TestResolveDamageMatchesResolvedDamage verifies that the resolved damage structure and the formatted damage string
// stay in step: ResolvedDamage is nothing more than ResolveDamage formatted, so the two must agree for a weapon that
// exercises every piece of the formatting -- a damage bonus, an armor divisor, an explosive damage type and
// fragmentation with an armor divisor and type of its own. A weapon with no owner has nothing to resolve against, so
// ResolveDamage answers nil and ResolvedDamage falls back on the unresolved form.
func TestResolveDamageMatchesResolvedDamage(t *testing.T) {
	c := check.New(t)
	bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
	bonus.Amount = fxp.Two
	w := newWeaponWithBonuses(false, bonus)
	w.Damage.Type = "cr ex"
	w.Damage.ArmorDivisor = fxp.Two
	w.Damage.Fragmentation = "2d"
	w.Damage.FragmentationArmorDivisor = fxp.Three
	w.Damage.FragmentationType = "cut"

	resolved := w.Damage.ResolveDamage(nil)
	c.NotNil(resolved, "a weapon with an entity resolves")
	c.Equal("1d+2(2) cr ex [2d(3) cut]", resolved.String(), "the resolved damage formats as it always has")
	c.Equal(w.Damage.ResolvedDamage(nil), resolved.String(), "ResolvedDamage is ResolveDamage formatted")
	c.Equal(2, resolved.Dice.Modifier, "the damage bonus landed on the dice")
	c.Equal(fxp.Two, resolved.ArmorDivisor, "the armor divisor came through")
	c.Equal(2, resolved.Fragmentation.Count, "the fragmentation dice came through")
	c.Equal(fxp.Three, resolved.FragmentationArmorDivisor, "the fragmentation armor divisor came through")
	c.Equal("cut", resolved.FragmentationType, "the fragmentation type came through")
	c.True(resolved.HasFragmentation, "a weapon that throws fragments has fragmentation")

	unowned := &gurps.WeaponDamage{}
	c.Nil(unowned.ResolveDamage(nil), "a weapon damage with no owner cannot be resolved")
	c.Equal(unowned.String(), unowned.ResolvedDamage(nil), "and its damage falls back on the unresolved form")
}

// TestResolvedWeaponDamageIsExplosive verifies that damage counts as an explosion (BX414) when its type carries the
// Explosion modifier or when it throws fragments, and not otherwise.
func TestResolvedWeaponDamageIsExplosive(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name          string
		damageType    string
		fragmentation string
		want          bool
	}{
		{name: "an explosive damage type", damageType: "cr ex", want: true},
		{name: "a decorated explosive damage type", damageType: "burn ex*", want: true},
		{name: "fragmentation without the explosion modifier", damageType: "cr", fragmentation: "2d", want: true},
		{name: "both", damageType: "cr ex", fragmentation: "2d", want: true},
		{name: "neither", damageType: "cut", want: false},
		{name: "fragmentation that formats as nothing", damageType: "cut", fragmentation: "0", want: false},
	} {
		w := newWeaponWithBonuses(false)
		w.Damage.Type = tc.damageType
		w.Damage.Fragmentation = tc.fragmentation
		resolved := w.Damage.ResolveDamage(nil)
		c.NotNil(resolved, tc.name)
		c.Equal(tc.want, resolved.IsExplosive(), tc.name)
	}
}

// TestWeaponDamageBonusDice verifies that a damage bonus given as dice is folded into the weapon's damage the way a
// base damage specification is: dice add to the dice, a modifier and multiplier come along, dice with other sides are
// averaged in, taking away every die leaves no damage, and a percentage bonus cannot carry dice.
func TestWeaponDamageBonusDice(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name    string
		base    string
		dice    string
		percent bool
		want    string
	}{
		{name: "adds a die", base: "1d", dice: "+1d", want: "2d cr"},
		{name: "adds dice and a modifier", base: "1d", dice: "1d+2", want: "2d+2 cr"},
		{name: "adds dice with a multiplier", base: "1d", dice: "2dx3", want: "3dx3 cr"},
		{name: "adds dice, a modifier and a multiplier", base: "1d", dice: "2d+1x3", want: "3d+1x3 cr"},
		{name: "takes a die away", base: "2d+1", dice: "-1d", want: "1d+1 cr"},
		{name: "takes every die away", base: "1d", dice: "-1d", want: "cr"},
		{name: "takes more dice away than there are", base: "1d+3", dice: "-2d", want: "cr"},
		{name: "averages in dice with other sides", base: "1d", dice: "+1d3", want: "2d3+2 cr"},
		{name: "ignores dice on a percentage bonus", base: "1d", dice: "+1d", percent: true, want: "1d cr"},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.Amount = 0
		bonus.Percent = tc.percent
		var ok bool
		bonus.Dice, ok = gurps.ParseBonusDice(tc.dice)
		c.True(ok, "%s: %q parses", tc.name, tc.dice)
		w := newWeaponWithBonuses(false, bonus)
		w.Damage.Base = tc.base
		c.Equal(tc.want, w.Damage.ResolvedDamage(nil), tc.name)
	}
}

// TestWeaponDamageBonusDiceScaled verifies that dice on a per-level or per-die damage bonus are multiplied the way a
// flat amount is, that several dice bonuses add up, and that the tooltip reports the dice before and after scaling.
func TestWeaponDamageBonusDiceScaled(t *testing.T) {
	c := check.New(t)

	perLevel := gurps.NewWeaponBonus(feature.WeaponBonus)
	perLevel.Amount = 0
	perLevel.Dice, _ = gurps.ParseBonusDice("1d+2")
	perLevel.PerLevel = true
	w := newWeaponWithBonuses(false, perLevel)
	owner, ok := w.Owner.(*gurps.Trait)
	c.True(ok)
	owner.CanLevel = true
	owner.Levels = fxp.Three
	c.Equal("4d+6 cr", w.Damage.ResolvedDamage(nil), "a per-level dice bonus scales with the level")
	tooltip := w.Damage.DamageTooltip()
	c.True(strings.Contains(tooltip, "Gadget 3 [+3d+6 (+1d+2 per level) to damage]"),
		"the tooltip reports the scaled and the per-level dice: %s", tooltip)

	perDie := gurps.NewWeaponBonus(feature.WeaponBonus)
	perDie.Amount = 0
	perDie.Dice, _ = gurps.ParseBonusDice("1d")
	perDie.PerDie = true
	flat := gurps.NewWeaponBonus(feature.WeaponBonus)
	flat.Amount = fxp.One
	w = newWeaponWithBonuses(false, perDie, flat)
	w.Damage.Base = "2d"
	c.Equal("4d+1 cr", w.Damage.ResolvedDamage(nil),
		"a per-die dice bonus adds a die for each base die, and a flat bonus still adds to the modifier")
	tooltip = w.Damage.DamageTooltip()
	c.True(strings.Contains(tooltip, "Gadget [+2d (+1d per die) to damage]"),
		"the tooltip reports the scaled and the per-die dice: %s", tooltip)
	c.True(strings.Contains(tooltip, "Gadget [+1 to damage]"), "the flat bonus reports as it always has: %s", tooltip)
}

// TestWeaponDamageBonusDiceAddedInBestOrder verifies that several dice bonuses are folded into the damage in the order
// that rounds least: the dice with the base's sides first, so they add exactly, and the rest grouped by their sides and
// summed before being averaged in, whatever order the bonuses were collected in.
func TestWeaponDamageBonusDiceAddedInBestOrder(t *testing.T) {
	c := check.New(t)
	newDiceBonus := func(spec string) *gurps.WeaponBonus {
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.Amount = 0
		var ok bool
		bonus.Dice, ok = gurps.ParseBonusDice(spec)
		c.True(ok, "%q parses", spec)
		return bonus
	}
	for _, tc := range []struct {
		name string
		base string
		dice []string
		want string
	}{
		// 2d6+1d3 averages 9, which 4d3+1 is exactly; averaging the d3 in first would give 3d3+4, averaging 10.
		{name: "matching dice first", base: "1d", dice: []string{"+1d3", "+1d"}, want: "4d3+1 cr"},
		{name: "matching dice first, collected the other way round", base: "1d", dice: []string{"+1d", "+1d3"}, want: "4d3+1 cr"},
		// 1d3+2d6 averages 9, which 4d3+1 is exactly; averaging the d6s in one at a time would give 4d3+2.
		{name: "same-sided dice summed before averaging", base: "1d3", dice: []string{"+1d", "+1d"}, want: "4d3+1 cr"},
		{name: "same-sided dice cancel before averaging", base: "1d3", dice: []string{"+1d", "-1d"}, want: "1d3 cr"},
		{name: "matching dice taken away first", base: "2d", dice: []string{"+1d3", "-1d"}, want: "2d3+2 cr"},
	} {
		bonuses := make([]*gurps.WeaponBonus, 0, len(tc.dice))
		for _, spec := range tc.dice {
			bonuses = append(bonuses, newDiceBonus(spec))
		}
		w := newWeaponWithBonuses(false, bonuses...)
		w.Damage.Base = tc.base
		c.Equal(tc.want, w.Damage.ResolvedDamage(nil), tc.name)
	}
}

// TestWeaponDamageBonusDicePhoenixFlame verifies that, under the Phoenix Flame D3 progression, the modifier of a
// per-level or per-die dice bonus is halved the same way a flat per-level or per-die amount is, while its dice are
// added in full.
func TestWeaponDamageBonusDicePhoenixFlame(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name     string
		dice     string
		amount   fxp.Int
		perLevel bool
		perDie   bool
		want     string
	}{
		{name: "flat per-level amount is halved", amount: fxp.Two, perLevel: true, want: "2d3+3 cr"},
		{name: "per-level dice keep their dice and halve their modifier", dice: "1d+2", perLevel: true, want: "7d3+4 cr"},
		{name: "flat per-die amount is halved", amount: fxp.Two, perDie: true, want: "2d3+2 cr"},
		{name: "per-die dice keep their dice and halve their modifier", dice: "1d+2", perDie: true, want: "5d3+3 cr"},
		{name: "a plain dice bonus is untouched", dice: "1d+2", want: "3d3+4 cr"},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponBonus)
		bonus.Amount = tc.amount
		bonus.PerLevel = tc.perLevel
		bonus.PerDie = tc.perDie
		if tc.dice != "" {
			var ok bool
			bonus.Dice, ok = gurps.ParseBonusDice(tc.dice)
			c.True(ok, "%s: %q parses", tc.name, tc.dice)
		}
		w := newWeaponWithBonuses(false, bonus)
		w.Damage.Base = "2d3"
		owner, ok := w.Owner.(*gurps.Trait)
		c.True(ok)
		owner.CanLevel = true
		owner.Levels = fxp.Three
		w.Entity().SheetSettings.DamageProgression = progression.PhoenixFlameD3
		c.Equal(tc.want, w.Damage.ResolvedDamage(nil), tc.name)
	}
}
