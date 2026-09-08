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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestIsExplosiveDamageType verifies that the Explosion modifier (B104) is recognized in every form the data libraries
// write it in -- decorated with an asterisk, a comma or a slashed suffix, and in any position among the other damage
// modifiers -- while a longer word that merely starts with the same two letters is not mistaken for it.
func TestIsExplosiveDamageType(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name       string
		damageType string
		want       bool
	}{
		{name: "a plain grenade", damageType: "cr ex", want: true},
		{name: "decorated and buried among other modifiers", damageType: "burn ex* rad sur", want: true},
		{name: "following another modifier", damageType: "cr dkb ex", want: true},
		{name: "with a divisor suffix", damageType: "cr ex/2", want: true},
		{name: "with a per-point suffix", damageType: "cr ex/point", want: true},
		{name: "trailing punctuation", damageType: "ex,", want: true},
		{name: "on its own", damageType: "ex", want: true},
		{name: "upper case", damageType: "CR EX", want: true},
		{name: "not explosive", damageType: "cut", want: false},
		{name: "a longer word starting with ex", damageType: "exp", want: false},
		{name: "ex is only part of a word", damageType: "cr exotic", want: false},
		{name: "empty", damageType: "", want: false},
	} {
		c.Equal(tc.want, IsExplosiveDamageType(tc.damageType), tc.name)
	}
}

// TestExplosionRadii verifies the two radii BX414 measures in dice: everything within twice the dice of damage is
// vulnerable to the collateral damage, and everything within five times the dice of fragmentation damage is vulnerable
// to the fragments. A dice multiplier counts toward the dice, which is what makes the worked example's 6dx2 reach 24
// yards rather than 12.
func TestExplosionRadii(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name        string
		d           dice.Dice
		wantDice    int
		wantBlast   int
		wantFragged int
	}{
		{
			name: "the BX414 worked example, 6dx2", d: dice.Dice{Count: 6, Sides: 6, Multiplier: 2},
			wantDice: 12, wantBlast: 24, wantFragged: 60,
		},
		{
			name: "a [2d] fragmentation attack", d: dice.Dice{Count: 2, Sides: 6, Multiplier: 1},
			wantDice: 2, wantBlast: 4, wantFragged: 10,
		},
		{
			name: "a modifier does not add dice", d: dice.Dice{Count: 6, Sides: 6, Modifier: 12, Multiplier: 1},
			wantDice: 6, wantBlast: 12, wantFragged: 30,
		},
		{
			name: "an unset multiplier counts as one", d: dice.Dice{Count: 3, Sides: 6},
			wantDice: 3, wantBlast: 6, wantFragged: 15,
		},
		{name: "no dice", d: dice.Dice{Sides: 6, Multiplier: 1}, wantDice: 0, wantBlast: 0, wantFragged: 0},
		{
			name: "a negative count cannot produce a negative radius",
			d:    dice.Dice{Count: -6, Sides: 6, Multiplier: 2}, wantDice: 0, wantBlast: 0, wantFragged: 0,
		},
	} {
		c.Equal(tc.wantDice, DiceOfDamage(tc.d), "%s: dice of damage", tc.name)
		c.Equal(tc.wantBlast, CollateralDamageRadius(tc.d), "%s: collateral damage radius", tc.name)
		c.Equal(tc.wantFragged, FragmentationRadius(tc.d), "%s: fragmentation radius", tc.name)
	}
}

// TestCollateralDivisor verifies the divisor each environment applies to an explosion's collateral damage: 3 x the
// distance in air (BX414), the distance alone underwater and 10 x the distance in vacuum (BX415). A distance of zero or
// less is a direct hit, which takes the listed damage undivided.
func TestCollateralDivisor(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name        string
		environment ExplosionEnvironment
		distance    fxp.Int
		want        fxp.Int
	}{
		{name: "3 yards in air", environment: ExplosionInAir, distance: fxp.Three, want: fxp.Nine},
		{name: "3 yards underwater", environment: ExplosionUnderwater, distance: fxp.Three, want: fxp.Three},
		{name: "3 yards in vacuum", environment: ExplosionInVacuum, distance: fxp.Three, want: fxp.Thirty},
		{
			name: "a fractional distance in air", environment: ExplosionInAir, distance: fxp.OneAndAHalf,
			want: fxp.FromStringForced("4.5"),
		},
		{name: "struck directly, in air", environment: ExplosionInAir, distance: 0, want: fxp.One},
		{name: "struck directly, underwater", environment: ExplosionUnderwater, distance: 0, want: fxp.One},
		{name: "struck directly, in vacuum", environment: ExplosionInVacuum, distance: 0, want: fxp.One},
		{
			name: "a negative distance is a direct hit", environment: ExplosionInAir, distance: fxp.NegOne,
			want: fxp.One,
		},
		{
			name: "an unrecognized environment behaves as air", environment: ExplosionEnvironment(200),
			distance: fxp.Three, want: fxp.Nine,
		},
	} {
		c.Equal(tc.want, tc.environment.CollateralDivisor(tc.distance), tc.name)
	}
}

// TestDividedDamage verifies the least, average and greatest damage an explosion inflicts once divided and rounded
// down (BX414), including the BX414 worked example: 6dx2 three yards away in air is divided by 9, so its 12 to 72 points
// (42 on average) become 1 to 8 (4 on average). Damage is never negative, so a total below zero is reported as zero.
func TestDividedDamage(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name        string
		d           dice.Dice
		divisor     fxp.Int
		wantMinimum int
		wantAverage int
		wantMaximum int
	}{
		{
			name: "the BX414 worked example, 6dx2 at 3 yards in air",
			d:    dice.Dice{Count: 6, Sides: 6, Multiplier: 2}, divisor: fxp.Nine,
			wantMinimum: 1, wantAverage: 4, wantMaximum: 8,
		},
		{
			name: "the same blast, struck directly", d: dice.Dice{Count: 6, Sides: 6, Multiplier: 2},
			divisor: fxp.One, wantMinimum: 12, wantAverage: 42, wantMaximum: 72,
		},
		{
			name: "a divisor below one leaves the damage alone", d: dice.Dice{Count: 6, Sides: 6, Multiplier: 2},
			divisor: fxp.Half, wantMinimum: 12, wantAverage: 42, wantMaximum: 72,
		},
		{
			name: "a zero divisor leaves the damage alone", d: dice.Dice{Count: 6, Sides: 6, Multiplier: 2},
			divisor: 0, wantMinimum: 12, wantAverage: 42, wantMaximum: 72,
		},
		{
			name:    "a modifier applies before the multiplier",
			d:       dice.Dice{Count: 2, Sides: 6, Modifier: 2, Multiplier: 3},
			divisor: fxp.One, wantMinimum: 12, wantAverage: 27, wantMaximum: 42,
		},
		{
			name: "a fractional average rounds down", d: dice.Dice{Count: 1, Sides: 6, Multiplier: 1},
			divisor: fxp.One, wantMinimum: 1, wantAverage: 3, wantMaximum: 6,
		},
		{
			name: "a negative modifier cannot drive the damage below zero",
			d:    dice.Dice{Count: 1, Sides: 6, Modifier: -3, Multiplier: 1}, divisor: fxp.One,
			wantMinimum: 0, wantAverage: 0, wantMaximum: 3,
		},
		{
			name: "a modifier that wipes out the dice entirely",
			d:    dice.Dice{Count: 1, Sides: 6, Modifier: -10, Multiplier: 1}, divisor: fxp.One,
			wantMinimum: 0, wantAverage: 0, wantMaximum: 0,
		},
		{
			name: "a big blast far away rounds down to nothing",
			d:    dice.Dice{Count: 6, Sides: 6, Multiplier: 2}, divisor: fxp.Hundred,
			wantMinimum: 0, wantAverage: 0, wantMaximum: 0,
		},
		{
			name: "no dice at all", d: dice.Dice{Sides: 6, Multiplier: 1}, divisor: fxp.Three,
			wantMinimum: 0, wantAverage: 0, wantMaximum: 0,
		},
	} {
		minimum, average, maximum := DividedDamage(tc.d, tc.divisor)
		c.Equal(tc.wantMinimum, minimum, "%s: minimum", tc.name)
		c.Equal(tc.wantAverage, average, "%s: average", tc.name)
		c.Equal(tc.wantMaximum, maximum, "%s: maximum", tc.name)
	}

	// The example again, spelled out the way BX414 does: the divisor comes from the environment and the distance.
	minimum, average, maximum := DividedDamage(dice.Dice{Count: 6, Sides: 6, Multiplier: 2},
		ExplosionInAir.CollateralDivisor(fxp.Three))
	c.Equal(1, minimum, "6dx2 at 3 yards, minimum")
	c.Equal(4, average, "6dx2 at 3 yards, average")
	c.Equal(8, maximum, "6dx2 at 3 yards, maximum")
}

// TestFragmentationSkill verifies that fragments attack at skill 15 with only the three modifiers BX415 allows: the
// range penalty from the Size and Speed/Range Table (BX550), the target's posture and its size modifier.
func TestFragmentationSkill(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name     string
		distance fxp.Int
		posture  int
		sm       int
		want     int
	}{
		{name: "at the center of the blast", distance: 0, want: FragmentationBaseSkill},
		{name: "2 yards away is still no range penalty", distance: fxp.Two, want: 15},
		{name: "2 yards away, crouching", distance: fxp.Two, posture: NonStandingTargetPenalty, want: 13},
		{name: "10 yards away", distance: fxp.Ten, want: 11},
		{name: "10 yards away, crouching", distance: fxp.Ten, posture: NonStandingTargetPenalty, want: 9},
		{
			name:     "10 yards away, prone to a ground burst",
			distance: fxp.Ten,
			posture:  NonStandingTargetPenalty + HalfExposedTorsoPenalty,
			want:     7,
		},
		{name: "10 yards away, crouching, and large", distance: fxp.Ten, posture: NonStandingTargetPenalty, sm: 2, want: 11},
		{name: "10 yards away, crouching, and small", distance: fxp.Ten, posture: NonStandingTargetPenalty, sm: -2, want: 7},
		{name: "an airburst ignores posture", distance: fxp.Ten, want: 11},
		{name: "100 yards away", distance: fxp.Hundred, want: 5},
	} {
		c.Equal(tc.want, FragmentationSkill(tc.distance, tc.posture, tc.sm), tc.name)
	}
}

// TestTargetPostureFragmentPenalty pins the Target column of the Posture Table (BX551): standing is unmodified, every
// other posture is -2, and a crawling or lying-down target is a further -2 from a ground burst, since the table's note
// leaves only half of its torso exposed to an attacker at ground level; nothing changes for the other postures.
func TestTargetPostureFragmentPenalty(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name        string
		posture     TargetPosture
		groundBurst bool
		want        int
	}{
		{name: "standing", posture: StandingTarget, want: 0},
		{name: "standing, ground burst", posture: StandingTarget, groundBurst: true, want: 0},
		{name: "crouching", posture: CrouchingTarget, want: -2},
		{name: "crouching, ground burst", posture: CrouchingTarget, groundBurst: true, want: -2},
		{name: "prone, blast above or within reach", posture: ProneTarget, want: -2},
		{name: "prone, ground burst", posture: ProneTarget, groundBurst: true, want: -4},
		{name: "unknown posture counts as standing", posture: TargetPosture(99), groundBurst: true, want: 0},
	} {
		c.Equal(tc.want, tc.posture.FragmentPenalty(tc.groundBurst), tc.name)
	}
}

// TestDemolition verifies the Demolition math of BX415 in both directions, including its worked example: a 6dx8 blast
// needs (8x8)/4 = 16 lbs of TNT, or 16/0.8 = 20 lbs of dynamite, and 20 lbs of dynamite makes a 6dx8 blast again.
func TestDemolition(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.Sixteen, TNTForBlast(fxp.Eight), "6dx8 needs (8x8)/4 lbs of TNT")
	c.Equal(fxp.Twenty, ExplosiveForBlast(fxp.Eight, fxp.FourFifths), "the BX415 example: 6dx8 of dynamite is 20 lbs")
	c.Equal(fxp.Eight, BlastForExplosive(fxp.Twenty, fxp.FourFifths), "20 lbs of dynamite makes a 6dx8 blast")
	c.Equal(fxp.One, BlastForExplosive(fxp.Quarter, fxp.One), "a quarter pound of TNT makes a 6d blast")
	c.Equal(fxp.Quarter, TNTForBlast(fxp.One), "and a 6d blast needs a quarter pound of TNT")

	for _, tc := range []struct {
		name   string
		n      fxp.Int
		ref    fxp.Int
		wantT  fxp.Int
		wantEx fxp.Int
	}{
		{name: "TNT is the yardstick", n: fxp.Six, ref: fxp.One, wantT: fxp.Nine, wantEx: fxp.Nine},
		{
			name: "a weaker explosive takes more of it", n: fxp.Six, ref: fxp.Half, wantT: fxp.Nine,
			wantEx: fxp.FromInteger(18),
		},
		{
			name: "a stronger explosive takes less", n: fxp.Six, ref: fxp.Six, wantT: fxp.Nine,
			wantEx: fxp.OneAndAHalf,
		},
		{name: "no blast needs no explosive", n: 0, ref: fxp.One, wantT: 0, wantEx: 0},
		{name: "a negative blast needs no explosive", n: fxp.NegOne, ref: fxp.One, wantT: 0, wantEx: 0},
		{name: "an inert explosive can never do it", n: fxp.Six, ref: 0, wantT: fxp.Nine, wantEx: 0},
	} {
		c.Equal(tc.wantT, TNTForBlast(tc.n), "%s: TNT", tc.name)
		c.Equal(tc.wantEx, ExplosiveForBlast(tc.n, tc.ref), "%s: explosive", tc.name)
	}

	for _, tc := range []struct {
		name   string
		pounds fxp.Int
		ref    fxp.Int
		want   fxp.Int
	}{
		{name: "9 lbs of TNT", pounds: fxp.Nine, ref: fxp.One, want: fxp.Six},
		{name: "a fractional result is left fractional", pounds: fxp.One, ref: fxp.One, want: fxp.Two},
		{
			name: "1 lb of dynamite", pounds: fxp.One, ref: fxp.FourFifths,
			want: fxp.FromFloat(1.7888), // sqrt(3.2), truncated to four decimal places
		},
		{name: "no explosive makes no blast", pounds: 0, ref: fxp.One, want: 0},
		{name: "a negative weight makes no blast", pounds: fxp.NegOne, ref: fxp.One, want: 0},
		{name: "an inert explosive makes no blast", pounds: fxp.Nine, ref: 0, want: 0},
	} {
		c.Equal(tc.want, BlastForExplosive(tc.pounds, tc.ref), tc.name)
	}
}

// TestScatterDistance verifies that a miss scatters by the margin of failure in yards -- or the square of it for a
// flying or underwater target, or an unseen one attacked with Artillery or Dropping -- and that the scatter is always
// held down to half the distance to the target, rounded up (BX414).
func TestScatterDistance(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name       string
		margin     int
		distance   fxp.Int
		squared    bool
		want       fxp.Int
		wantCapped bool
	}{
		{name: "missed by 3 at 20 yards", margin: 3, distance: fxp.Twenty, want: fxp.Three},
		{
			name: "missed by 9 at 5 yards", margin: 9, distance: fxp.Five, want: fxp.Three, wantCapped: true,
		}, // Half of 5 rounds up to 3.
		{name: "exactly at the cap", margin: 10, distance: fxp.Twenty, want: fxp.Ten},
		{name: "a yard past the cap", margin: 11, distance: fxp.Twenty, want: fxp.Ten, wantCapped: true},
		{
			name: "missed by 4 at a flier 100 yards away", margin: 4, distance: fxp.Hundred, squared: true,
			want: fxp.Sixteen,
		},
		{
			name: "the square is capped too", margin: 5, distance: fxp.Twenty, squared: true, want: fxp.Ten,
			wantCapped: true,
		},
		{name: "no margin, no scatter", margin: 0, distance: fxp.Twenty, want: 0},
		{name: "a negative margin does not scatter", margin: -3, distance: fxp.Twenty, want: 0},
		{
			name: "no distance leaves nowhere to scatter to", margin: 3, distance: 0, want: 0, wantCapped: true,
		},
	} {
		yards, capped := ScatterDistance(tc.margin, tc.distance, tc.squared)
		c.Equal(tc.want, yards, "%s: yards", tc.name)
		c.Equal(tc.wantCapped, capped, "%s: capped", tc.name)
	}
}

// TestConeWidthAndAreaDivisor verifies the cone geometry of BX413, including its worked example -- a cone with a maximum
// range of 100 yards and a maximum width of 5 yards is 3 yards wide at 60 yards -- and the Dissipation divisors of
// BX414: a cone divides its damage by its width at the target, an area effect by the distance from its center. Neither
// divisor ever drops below one.
func TestConeWidthAndAreaDivisor(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name     string
		distance fxp.Int
		maxRange fxp.Int
		maxWidth fxp.Int
		want     fxp.Int
	}{
		{
			name: "the BX413 worked example", distance: fxp.Sixty, maxRange: fxp.Hundred, maxWidth: fxp.Five,
			want: fxp.Three,
		},
		{
			name: "at the maximum range it reaches its maximum width", distance: fxp.Hundred, maxRange: fxp.Hundred,
			maxWidth: fxp.Five, want: fxp.Five,
		},
		{
			name: "a fractional width is left fractional", distance: fxp.Seventy, maxRange: fxp.Hundred,
			maxWidth: fxp.Five, want: fxp.ThreeAndAHalf,
		},
		{
			name: "at the apex it is a yard wide", distance: 0, maxRange: fxp.Hundred, maxWidth: fxp.Five,
			want: fxp.One,
		},
		{
			name: "close in it is still a yard wide", distance: fxp.Ten, maxRange: fxp.Hundred, maxWidth: fxp.Five,
			want: fxp.One,
		},
		{
			name: "an unspecified width spreads a yard per yard", distance: fxp.Sixty, maxRange: fxp.Hundred,
			maxWidth: 0, want: fxp.Sixty,
		},
		{
			name: "an unspecified range spreads a yard per yard too", distance: fxp.Sixty, maxRange: 0,
			maxWidth: fxp.Five, want: fxp.Sixty,
		},
		{name: "a negative distance is the apex", distance: fxp.NegOne, maxRange: 0, maxWidth: 0, want: fxp.One},
	} {
		c.Equal(tc.want, ConeWidth(tc.distance, tc.maxRange, tc.maxWidth), tc.name)
	}

	for _, tc := range []struct {
		name     string
		distance fxp.Int
		want     fxp.Int
	}{
		{name: "at the center", distance: 0, want: fxp.One},
		{name: "a yard out", distance: fxp.One, want: fxp.One},
		{name: "2 yards out halves the damage twice over", distance: fxp.Two, want: fxp.Two},
		{name: "10 yards out", distance: fxp.Ten, want: fxp.Ten},
		{name: "a negative distance is the center", distance: fxp.NegOne, want: fxp.One},
	} {
		c.Equal(tc.want, AreaDamageDivisor(tc.distance), tc.name)
	}

	// A cone 3 yards wide at the target divides a 6d attack by 3 (Dissipation, BX414).
	minimum, average, maximum := DividedDamage(dice.Dice{Count: 6, Sides: 6, Multiplier: 1},
		ConeWidth(fxp.Sixty, fxp.Hundred, fxp.Five))
	c.Equal(2, minimum, "6d divided by a 3 yard wide cone, minimum")
	c.Equal(7, average, "6d divided by a 3 yard wide cone, average")
	c.Equal(12, maximum, "6d divided by a 3 yard wide cone, maximum")
}

// TestCoverDRFromBody verifies that someone who throws himself onto an explosive shields everyone else with his torso
// DR plus his HP (BX415), and that neither piece can subtract from the total.
func TestCoverDRFromBody(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name    string
		torsoDR int
		hp      fxp.Int
		want    int
	}{
		{name: "an unarmored 10 HP human", torsoDR: 0, hp: fxp.Ten, want: 10},
		{name: "wearing mail", torsoDR: 4, hp: fxp.Ten, want: 14},
		{name: "a fractional HP total rounds down", torsoDR: 4, hp: fxp.FromStringForced("10.5"), want: 14},
		{name: "negative HP counts as none", torsoDR: 4, hp: fxp.FromInteger(-10), want: 4},
		{name: "negative DR counts as none", torsoDR: -4, hp: fxp.Ten, want: 10},
	} {
		c.Equal(tc.want, CoverDRFromBody(tc.torsoDR, tc.hp), tc.name)
	}
}

// TestLargeAreaDR verifies the Large-Area Injury DR of BX400: the average of the torso DR and the DR of the
// least-protected exposed location, rounded up. It pins the rounding, the way a damage type picks the DR that varies
// by attack, and the answers for a body with no torso, no exposed location or no body type at all.
func TestLargeAreaDR(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addCarriedEquipmentWithFeatures(e, "Mail Hauberk", newTestDRBonus(fxp.Five, AllID, TorsoID))
	e.Recalculate()
	torso := e.SheetSettings.BodyType.LookupLocationByID(e, TorsoID)
	c.NotNil(torso, "the default body has a torso")
	c.Equal(5, torso.DR(e, nil, nil)[AllID], "the hauberk gives the torso DR 5")

	onlyTorso := func(loc *HitLocation) bool { return loc.LocID == TorsoID }

	// Everything is exposed, so the least-protected location is a bare one, such as the eyes: (5 + 0)/2 rounds up to 3.
	c.Equal(3, LargeAreaDR(e, "", nil), "torso 5 and a bare location average to 3, rounded up")
	c.Equal(3, LargeAreaDR(e, "cr", nil), "a damage type with no specialized DR changes nothing")

	// Only the torso faces the blast, so it is both halves of the average.
	c.Equal(5, LargeAreaDR(e, "", onlyTorso), "with only the torso exposed, its own DR stands")

	// DR that only stops one kind of attack counts for that attack alone.
	addCarriedEquipmentWithFeatures(e, "Fire Cloak", newTestDRBonus(fxp.Four, "burn", TorsoID))
	e.Recalculate()
	c.Equal(9, LargeAreaDR(e, "burn", onlyTorso), "the burn-specialized DR stacks on top of the general DR")
	c.Equal(5, LargeAreaDR(e, "cr", onlyTorso), "it does nothing against crushing damage")
	c.Equal(5, LargeAreaDR(e, "", onlyTorso), "nor when no damage type is named")
	c.Equal(5, LargeAreaDR(e, AllID, onlyTorso), `a damage type of "all" is the general DR, not a second helping`)
	c.Equal(5, LargeAreaDR(e, " BURN ", nil), "the type is matched with the case and padding stripped: (9 + 0)/2")

	// No location is exposed at all, so there is nothing to average.
	c.Equal(0, LargeAreaDR(e, "", func(_ *HitLocation) bool { return false }), "nothing exposed, no DR")

	// A body with no torso: the least-protected exposed location stands in for both halves of the average.
	e2 := NewEntity()
	arm := NewHitLocation(e2, "")
	arm.LocID = "arm"
	arm.ChoiceName = "Arm"
	arm.TableName = "Arm"
	arm.Slots = 1
	arm.DRBonus = 4
	armless := NewHitLocation(e2, "")
	armless.LocID = "tail"
	armless.ChoiceName = "Tail"
	armless.TableName = "Tail"
	armless.Slots = 1
	armless.DRBonus = -3 // Negative DR is nonsense, but the data allows it, so it must not subtract from the average.
	body := &Body{Roll: dice.Dice{Count: 3, Sides: 6}, Locations: []*HitLocation{arm, armless}}
	body.Update(e2)
	e2.SheetSettings.BodyType = body
	e2.Recalculate()
	c.Equal(0, LargeAreaDR(e2, "", nil), "the negative DR clamps to zero, and with no torso it is both halves")
	c.Equal(4, LargeAreaDR(e2, "", func(loc *HitLocation) bool { return loc.LocID == "arm" }),
		"with only the armored location exposed, its own DR stands in for the missing torso")

	// No body type at all yields no DR rather than falling back on the global default body.
	e2.SheetSettings.BodyType = nil
	c.Equal(0, LargeAreaDR(e2, "", nil), "an entity with no body type has no DR to average")
}

// TestExplosiveTypes verifies that the Relative Explosive Force Table (BX415) is transcribed in the printed order, with
// TNT as the yardstick, and that every row can actually be used to work out a demolition charge.
func TestExplosiveTypes(t *testing.T) {
	c := check.New(t)
	c.Equal(14, len(ExplosiveTypes), "the table has 14 rows")
	c.Equal("Serpentine Powder", ExplosiveTypes[0].Name, "the table starts at TL3")
	c.Equal(3, ExplosiveTypes[0].TL, "the table starts at TL3")
	c.Equal(fxp.ThreeTenths, ExplosiveTypes[0].REF, "serpentine powder is REF 0.3")
	c.Equal("Stabilized Metallic Hydrogen", ExplosiveTypes[13].Name, "the table ends at TL10")
	c.Equal(10, ExplosiveTypes[13].TL, "the table ends at TL10")
	c.Equal(fxp.Six, ExplosiveTypes[13].REF, "stabilized metallic hydrogen is REF 6")

	// Black Powder is printed twice, since it improves at TL5, so the name alone does not identify a row.
	blackPowder := make([]ExplosiveType, 0, 2)
	for _, one := range ExplosiveTypes {
		if one.Name == "Black Powder" {
			blackPowder = append(blackPowder, one)
		}
	}
	c.Equal(2, len(blackPowder), "Black Powder appears at both TL4 and TL5")
	c.Equal(4, blackPowder[0].TL, "the first Black Powder row is TL4")
	c.Equal(fxp.TwoFifths, blackPowder[0].REF, "TL4 Black Powder is REF 0.4")
	c.Equal(5, blackPowder[1].TL, "the second Black Powder row is TL5")
	c.Equal(fxp.Half, blackPowder[1].REF, "TL5 Black Powder is REF 0.5")

	tl := 0
	for _, one := range ExplosiveTypes {
		c.True(one.REF > 0, "%s has a usable REF", one.Name)
		c.NotEqual("", one.Description, "%s has a description", one.Name)
		c.True(one.TL >= tl, "%s does not go backward in TL", one.Name)
		tl = one.TL
		if one.Name == "TNT" {
			c.Equal(fxp.One, one.REF, "TNT is the yardstick, so its REF is 1")
			c.Equal(TNTForBlast(fxp.Eight), ExplosiveForBlast(fxp.Eight, one.REF), "TNT needs no conversion")
		}
		if one.Name == "Dynamite" {
			c.Equal(fxp.FourFifths, one.REF, "dynamite is REF 0.8")
			c.Equal(fxp.Twenty, ExplosiveForBlast(fxp.Eight, one.REF), "the BX415 example: 20 lbs for 6dx8")
		}
	}
}
