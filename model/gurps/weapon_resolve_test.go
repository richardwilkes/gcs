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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestWeaponAccuracyMixedBonusResolution verifies that when several bonus types are collected in one pass, each
// type's flat and percentage amounts are kept together and applied only to their own field, with the flat amounts
// applied before the percentage.
func TestWeaponAccuracyMixedBonusResolution(t *testing.T) {
	c := check.New(t)

	flatAcc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	flatAcc.Amount = fxp.Two
	percentAcc := gurps.NewWeaponBonus(feature.WeaponAccBonus)
	percentAcc.Percent = true
	percentAcc.Amount = fxp.FromInteger(50)
	percentScope := gurps.NewWeaponBonus(feature.WeaponScopeAccBonus)
	percentScope.Percent = true
	percentScope.Amount = fxp.Hundred
	w := newWeaponWithBonuses(false, flatAcc, percentAcc, percentScope)

	w.Accuracy = attack.ParseAccuracy("4+2")
	c.Equal("9+4", w.ResolvedAccuracy(nil).String(), "(4+2)*1.5 base, 2*2 scope")

	w.Accuracy = attack.ParseAccuracy("4")
	c.Equal("9", w.ResolvedAccuracy(nil).String(), "a percentage of an absent scope stays absent")
}

// TestWeaponBulkBonusResolution verifies that a flat or percentage bulk bonus adjusts the existing values rather than
// replacing them, and doesn't materialize a giant bulk on a weapon that doesn't have one (a stored giant bulk of 0
// means "no separate giant bulk"), while still adjusting one that is present.
func TestWeaponBulkBonusResolution(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		percent bool
		amount  fxp.Int
		bulk    string
		want    string
		msg     string
	}{
		{amount: fxp.NegOne, bulk: "-3", want: "-4", msg: "a weapon with no giant bulk should not gain one"},
		{amount: fxp.NegOne, bulk: "-3*", want: "-4*", msg: "the retracting stock marker should be preserved"},
		{amount: fxp.NegOne, bulk: "-3/-5", want: "-4/-6", msg: "an existing giant bulk should be adjusted"},
		{
			percent: true,
			amount:  fxp.FromInteger(50),
			bulk:    "-4",
			want:    "-6",
			msg:     "a percentage should adjust the bulk without adding a giant bulk",
		},
		{
			percent: true,
			amount:  fxp.FromInteger(50),
			bulk:    "-4/-6",
			want:    "-6/-9",
			msg:     "a percentage should adjust an existing giant bulk too",
		},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponBulkBonus)
		bonus.Percent = tc.percent
		bonus.Amount = tc.amount
		w := newWeaponWithBonuses(false, bonus)
		w.Bulk = attack.ParseBulk(tc.bulk)
		c.Equal(tc.want, w.ResolvedBulk(nil).String(), tc.msg)
	}
}

// TestWeaponBulkRetractingStockSwitch verifies that the "retracting stock" weapon switch sets whether the bulk has a
// retracting stock to the switch's value: on, it adds one, and off, it removes one. A bulk bonus still applies to a
// weapon whose stock the switch added.
func TestWeaponBulkRetractingStockSwitch(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		on    bool
		bulk  string
		bonus fxp.Int
		want  string
	}{
		{on: true, bulk: "-3", want: "-3*"},
		{on: false, bulk: "-3*", want: "-3"},
		{on: true, bulk: "-3", bonus: fxp.NegOne, want: "-4*"},
	} {
		sw := gurps.NewWeaponBonus(feature.WeaponSwitch)
		sw.SwitchType = wswitch.RetractingStock
		sw.SwitchTypeValue = tc.on
		bonuses := []*gurps.WeaponBonus{sw}
		if tc.bonus != 0 {
			bonus := gurps.NewWeaponBonus(feature.WeaponBulkBonus)
			bonus.Amount = tc.bonus
			bonuses = append(bonuses, bonus)
		}
		w := newWeaponWithBonuses(false, bonuses...)
		w.Bulk = attack.ParseBulk(tc.bulk)
		c.Equal(tc.want, w.ResolvedBulk(nil).String(), "switch %v on bulk %s with a %v bonus", tc.on, tc.bulk, tc.bonus)
	}
}

// TestWeaponRecoilBonusResolution verifies that a recoil bonus leaves a recoil of 0, which means the recoil isn't used,
// at 0, and doesn't take a recoil that is used below 1.
func TestWeaponRecoilBonusResolution(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		amount fxp.Int
		recoil string
		want   string
		msg    string
	}{
		{amount: fxp.One, recoil: "0/3", want: "0/4", msg: "an unused shot recoil stays unused"},
		{amount: fxp.One, recoil: "2/0", want: "3", msg: "an unused slug recoil stays unused"},
		{amount: -fxp.Five, recoil: "3/5", want: "1/1", msg: "a used recoil stays at least 1"},
	} {
		bonus := gurps.NewWeaponBonus(feature.WeaponRecoilBonus)
		bonus.Amount = tc.amount
		w := newWeaponWithBonuses(false, bonus)
		w.Recoil = attack.ParseRecoil(tc.recoil)
		c.Equal(tc.want, w.ResolvedRecoil(nil).String(), tc.msg)
	}
}

// TestWeaponRateOfFireTooltip verifies that the modifiers tooltip of a rate of fire lists the bonuses to each mode, and
// names the modes only when both have bonuses.
func TestWeaponRateOfFireTooltip(t *testing.T) {
	c := check.New(t)
	const mode1 = "\nGadget [+1 to shots per attack (mode 1)]"
	const mode2 = "\nGadget [+1 to secondary projectiles (mode 2)]"
	tooltipFor := func(types ...feature.Type) string {
		bonuses := make([]*gurps.WeaponBonus, 0, len(types))
		for _, one := range types {
			bonuses = append(bonuses, gurps.NewWeaponBonus(one))
		}
		w := newWeaponWithBonuses(false, bonuses...)
		w.RateOfFire = attack.ParseRoF("3/2")
		var tooltip xbytes.InsertBuffer
		w.ResolvedRateOfFire(&tooltip)
		return tooltip.String()
	}
	c.Equal(mode1, tooltipFor(feature.WeaponRofMode1ShotsBonus), "a bonus to the first mode alone")
	c.Equal(mode2, tooltipFor(feature.WeaponRofMode2SecondaryBonus), "a bonus to the second mode alone")
	both := tooltipFor(feature.WeaponRofMode1ShotsBonus, feature.WeaponRofMode2SecondaryBonus)
	for _, want := range []string{"First mode:", mode1, "Second mode:", mode2} {
		c.Contains(both, want, "bonuses to both modes")
	}
}

// TestWeaponDefenseResolve runs the resolution regression tests shared by both defenses against the parry and the
// block.
func TestWeaponDefenseResolve(t *testing.T) {
	for _, d := range []*defenseUnderTest{parryUnderTest, blockUnderTest} {
		t.Run(d.name, func(t *testing.T) { runDefenseResolveTests(t, d) })
	}
}

// defenseUnderTest describes one of the two weapon defenses for runDefenseResolveTests, which exercises the parts of
// the parry and block calculations that are the same for both.
type defenseUnderTest struct {
	name          string
	ownID         string       // the default type that resolves directly to this defense
	otherID       string       // the default type of the other defense
	distinctBonus int          // the bonus newDefenseTestWeaponWithBonuses(c, 2, 1) gives this defense
	bonusType     feature.Type // the type of a weapon bonus to this defense
	resolve       func(w *gurps.Weapon) string
	setModifier   func(w *gurps.Weapon, modifier fxp.Int)
}

var (
	parryUnderTest = &defenseUnderTest{
		name:          "parry",
		ownID:         gurps.ParryID,
		otherID:       gurps.BlockID,
		distinctBonus: 2,
		bonusType:     feature.WeaponParryBonus,
		resolve:       func(w *gurps.Weapon) string { return w.ResolvedParry(nil).String() },
		setModifier:   func(w *gurps.Weapon, modifier fxp.Int) { w.Parry.Modifier = modifier },
	}
	blockUnderTest = &defenseUnderTest{
		name:          "block",
		ownID:         gurps.BlockID,
		otherID:       gurps.ParryID,
		distinctBonus: 1,
		bonusType:     feature.WeaponBlockBonus,
		resolve:       func(w *gurps.Weapon) string { return w.ResolvedBlock(nil).String() },
		setModifier:   func(w *gurps.Weapon, modifier fxp.Int) { w.Block.Modifier = modifier },
	}
)

// runDefenseResolveTests runs the resolution regression tests shared by the parry and block calculations against the
// given defense.
func runDefenseResolveTests(t *testing.T, d *defenseUnderTest) {
	t.Helper()
	allDefaultTypes := []string{gurps.SkillID, d.ownID, d.otherID}

	// A default of the defense's own type must not have the +3 and the entity's defense bonus applied to it twice.
	// SkillLevelFast() already turns such a default into a defense level, so a weapon defaulting to "Cloak Parry" must
	// resolve to the same value as one defaulting to the "Cloak" skill itself.
	t.Run("OwnTypeDefault", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)

		// A skill-type default halves the skill level, then adds the +3 and the entity's defense bonus: 14/2 + 3 + 1.
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.SkillID)}
		c.Equal("11", d.resolve(w), "skill-type default")

		// A default of the defense's own type has already been through that conversion, so it must yield the same
		// value, not 15.
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(d.ownID)}
		c.Equal("11", d.resolve(w), "%s-type default", d.name)

		// Both kinds of default together still resolve to the same value.
		w.Defaults = []*gurps.SkillDefault{
			newDefenseTestDefault(gurps.SkillID),
			newDefenseTestDefault(d.ownID),
		}
		c.Equal("11", d.resolve(w), "both defaults")

		// The weapon's own modifier and the default's modifier still apply on top.
		d.setModifier(w, fxp.Two)
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(d.ownID)}
		w.Defaults[0].Modifier = -fxp.One
		c.Equal("12", d.resolve(w), "%s-type default with modifiers", d.name)
	})

	// A default of the other defense's type contributes what a plain skill default to the same skill would. A weapon's
	// Defaults list feeds the parry and block calculations alike, so a parry-type default reaches the block calculation
	// and vice versa; converting the defense level it already carries into the other defense halved the skill a second
	// time and folded the wrong bonus in.
	t.Run("OtherTypeDefault", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)

		// The other defense's default must yield the same value as the skill-type default it is derived from:
		// 14/2 + 3 + 1.
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(d.otherID)}
		c.Equal("11", d.resolve(w), "%s-type default", d.otherID)

		// Mixing it with a skill-type default changes nothing, since both resolve to the same value.
		w.Defaults = []*gurps.SkillDefault{
			newDefenseTestDefault(gurps.SkillID),
			newDefenseTestDefault(d.otherID),
		}
		c.Equal("11", d.resolve(w), "both defaults")

		// The weapon's own modifier and the default's modifier still apply on top: 11 - 1 + 2.
		d.setModifier(w, fxp.Two)
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(d.otherID)}
		w.Defaults[0].Modifier = -fxp.One
		c.Equal("12", d.resolve(w), "%s-type default with modifiers", d.otherID)
	})

	// A skill-level adjustment -- here the penalty for wielding a weapon whose minimum ST exceeds the character's -- is
	// folded into the skill level before it is halved into a defense, whichever kind of default the defense resolves
	// from. A defense-type default arrives already converted, so adding the adjustment to it counted the adjustment at
	// twice the weight the skill-type path gives it.
	t.Run("MinSTPenaltyAtSkillScale", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)
		w.Strength = attack.Strength{Min: fxp.FromInteger(12)} // -2 to skill for the ST 10 character

		// (14 - 2)/2 + 3 + 1, not (14/2 + 3 + 1) - 2.
		for _, defaultType := range allDefaultTypes {
			w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
			c.Equal("10", d.resolve(w), "%q default", defaultType)
		}
	})

	// The same with an odd skill level and an odd penalty, where merely halving the adjustment before applying it to an
	// already-converted default would still lose a point that the skill-type path keeps.
	t.Run("OddMinSTPenaltyAtSkillScale", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)
		e := w.Entity()
		e.Skills[0].Points = fxp.FromInteger(20) // DX+5, i.e. level 15
		e.Recalculate()
		c.Equal(fxp.FromInteger(15), e.Skills[0].LevelData.Level, "the Cloak skill should be at level 15")
		w.Strength = attack.Strength{Min: fxp.FromInteger(11)} // -1 to skill for the ST 10 character

		// (15 - 1)/2 + 3 + 1: the penalty costs nothing, since 15/2 also rounds down to 7.
		for _, defaultType := range allDefaultTypes {
			w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
			c.Equal("11", d.resolve(w), "%q default", defaultType)
		}
	})

	// The same for a skill bonus aimed at this weapon, the other half of the base adjustment. This is the layout the
	// High Tech library's Tonfa uses -- a skill default paired with a matching parry-type default -- so an over-weighted
	// adjustment on the defense-type default won the max in the defense resolution and inflated the displayed defense.
	t.Run("ThisWeaponSkillBonusAtSkillScale", func(t *testing.T) {
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
			newDefenseTestDefault(d.ownID),
		}

		// (14 + 1)/2 + 3 + 1 for both defaults, so neither can out-bid the other: 11, not 12.
		c.Equal("11", d.resolve(w), "a this-weapon skill bonus must not count double")
	})

	// The calculation applies the entity's bonus for this defense and not the other one's, whichever kind of default it
	// resolves from. The two bonuses are equal in the default fixture, so this uses one where they differ.
	t.Run("AppliesOwnBonus", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeaponWithBonuses(c, 2, 1) // +2 parry, +1 block

		// Every kind of default resolves to the same value, built from this defense's bonus alone: 14/2 + 3 + bonus.
		expected := fxp.FromInteger(14/2 + 3 + d.distinctBonus).String()
		for _, defaultType := range allDefaultTypes {
			w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
			c.Equal(expected, d.resolve(w), "%q default uses the %s bonus", defaultType, d.name)
		}
	})

	// A multiplier to the defense applies after every amount added to it, the weapon's own modifier, the default's
	// modifier and the weapon bonuses included, as AttributeBonus.Multiply has it.
	t.Run("MultiplierAfterWeaponAmounts", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)
		e := w.Entity()
		halved := gurps.NewTrait(e, nil, false)
		halved.Name = "Halved Defenses"
		multiplier := gurps.NewAttributeBonus(d.ownID)
		multiplier.Multiply = true
		multiplier.Amount = fxp.Half
		halved.Features = gurps.Features{multiplier}
		e.Traits = append(e.Traits, halved)
		owner, ok := w.Owner.(*gurps.Trait)
		c.True(ok, "the test weapon should be owned by a trait")
		bonus := gurps.NewWeaponBonus(d.bonusType)
		bonus.SelectionType = wsel.ThisWeapon
		bonus.Amount = fxp.One
		owner.Features = append(owner.Features, bonus)
		d.setModifier(w, fxp.Two)
		e.Recalculate()

		// (14/2 + 3 + 1 + 2 + 1) × 0.5, rounded, not (14/2 + 3 + 1) × 0.5, rounded, + 2 + 1.
		for _, defaultType := range allDefaultTypes {
			w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(defaultType)}
			c.Equal("7", d.resolve(w), "%q default", defaultType)
		}

		// The default's modifier is one of those amounts too: (14/2 + 3 + 1 + 2 + 1 - 2) × 0.5, rounded.
		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(d.ownID)}
		w.Defaults[0].Modifier = -fxp.Two
		e.DiscardCaches()
		c.Equal("6", d.resolve(w), "%s-type default with a modifier", d.name)
	})

	// A default of the defense's own type naming a skill the character lacks is passed over, even when an encumbered
	// fencing weapon's penalty is added afterward, which would otherwise wrap the missing level around to a huge one.
	// The parry is the case that matters; the block only takes that penalty because its resolution currently shares it.
	t.Run("AbsentSkillDefaultIsPassedOver", func(t *testing.T) {
		c := check.New(t)
		w := newDefenseTestWeapon(c)
		w.Parry.Fencing = true
		e := w.Entity()
		pack := gurps.NewEquipment(e, nil, false)
		pack.BaseWeight = "30 lb"
		e.CarriedEquipment = []*gurps.Equipment{pack}
		e.Recalculate()
		c.Equal(encumbrance.Light, e.EncumbranceLevel(true), "precondition: the character is encumbered")

		w.Defaults = []*gurps.SkillDefault{newDefenseTestDefault(gurps.SkillID)}
		want := d.resolve(w)
		absent := newDefenseTestDefault(d.ownID)
		absent.Name.Qualifier = "Absent"
		w.Defaults = append(w.Defaults, absent)
		c.Equal(want, d.resolve(w), "a %s-type default naming an absent skill changes nothing", d.name)
	})
}

// TestWeaponPercentBonusFloorsIncrement pins the documented rounding of weapon percentage bonuses: the increment is
// floored, i.e. rounded toward negative infinity, not toward zero. A -50% bonus on a recoil of 5 yields an increment
// of -2.5, which floors to -3, for a result of 2.
func TestWeaponPercentBonusFloorsIncrement(t *testing.T) {
	c := check.New(t)

	bonus := gurps.NewWeaponBonus(feature.WeaponRecoilBonus)
	bonus.Percent = true
	bonus.Amount = fxp.FromInteger(-50)
	w := newWeaponWithBonuses(false, bonus)

	w.Recoil = attack.ParseRecoil("5")
	c.Equal("2", w.ResolvedRecoil(nil).String())

	w.Recoil = attack.ParseRecoil("5/7")
	c.Equal("2/3", w.ResolvedRecoil(nil).String())
}

// newStrengthBasisWeapon builds a ST 14 entity carrying a "Telekinesis" trait at the given level and a trait that owns
// a single thrust-based weapon. When overrideTo is non-empty, the owning trait also carries a WeaponDamageStrengthBasis
// override, matching any weapon, that retargets the weapon's strength basis to that value.
func newStrengthBasisWeapon(melee bool, tkLevels int, overrideTo string) *gurps.Weapon {
	e := gurps.NewEntity()
	e.Attributes.Set[gurps.StrengthID].Adjustment = fxp.Four // ST 14
	owner := gurps.NewTrait(e, nil, false)
	owner.Name = "Gadget"
	w := gurps.NewWeapon(owner, melee)
	w.Damage.StrengthType = stdmg.Thrust
	owner.Weapons = []*gurps.Weapon{w}
	if overrideTo != "" {
		override := gurps.NewSelectorOverride(selector.WeaponDamageStrengthBasis)
		override.SelectionType = wsel.WithName
		override.Value = overrideTo
		override.NameCriteria.Compare = criteria.AnyText
		owner.Features = append(owner.Features, override)
	}
	e.Traits = append(e.Traits, owner)
	if tkLevels != 0 {
		tk := gurps.NewTrait(e, nil, false)
		tk.Name = "Telekinesis"
		tk.CanLevel = true
		tk.Levels = fxp.FromInteger(tkLevels)
		e.Traits = append(e.Traits, tk)
	}
	e.Recalculate()
	return w
}

// TestWeaponRangeUsesResolvedStrengthBasis verifies that a muscle-powered range is multiplied by the strength the
// weapon's *resolved* damage strength basis names, so a WeaponDamageStrengthBasis override to telekinetic strength
// moves the range along with the damage instead of leaving it on throwing ST.
func TestWeaponRangeUsesResolvedStrengthBasis(t *testing.T) {
	c := check.New(t)

	// Without an override, the thrust basis falls through to throwing ST, which is the character's ST of 14.
	w := newStrengthBasisWeapon(false, 5, "")
	w.Range = attack.ParseRange("x10/x15")
	c.Equal("140/210", w.ResolvedRange(nil).String(true), "the unoverridden range comes from throwing ST 14")

	// Retargeting the strength basis to telekinetic must drive the range from TK 5 instead.
	w = newStrengthBasisWeapon(false, 5, stdmg.TelekineticThrust.Key())
	w.Range = attack.ParseRange("x10/x15")
	c.Equal("50/75", w.ResolvedRange(nil).String(true), "an override to telekinetic drives the range from TK 5")

	// The IQ basis is honored through the same path.
	w = newStrengthBasisWeapon(false, 5, stdmg.IQThrust.Key())
	w.Range = attack.ParseRange("x10/x15")
	c.Equal("100/150", w.ResolvedRange(nil).String(true), "an override to IQ drives the range from IQ 10")
}
