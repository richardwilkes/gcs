// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestBlastDivisor verifies how the collateral damage of a 6dx2 blast reaches a target (BX414-BX415): divided by the
// environment's multiple of the distance, undivided for a target struck directly or at the very center, and not at
// all beyond the 24 yards the blast reaches.
func TestBlastDivisor(t *testing.T) {
	c := check.New(t)
	blast := dice.Dice{Count: 6, Sides: 6, Multiplier: 2}
	for _, tc := range []struct {
		name      string
		situation BlastSituation
		env       ExplosionEnvironment
		distance  fxp.Int
		divisor   fxp.Int
		reach     BlastReach
	}{
		{"air at 3 yards", CaughtInBlast, ExplosionInAir, fxp.Three, fxp.Nine, BlastCollateral},
		{"underwater at 3 yards", CaughtInBlast, ExplosionUnderwater, fxp.Three, fxp.Three, BlastCollateral},
		{"vacuum at 3 yards", CaughtInBlast, ExplosionInVacuum, fxp.Three, fxp.Thirty, BlastCollateral},
		{"at the center", CaughtInBlast, ExplosionInAir, 0, fxp.One, BlastAtCenter},
		{"at the edge", CaughtInBlast, ExplosionInAir, fxp.FromInteger(24), fxp.FromInteger(72), BlastCollateral},
		{"beyond the edge", CaughtInBlast, ExplosionInAir, fxp.FromStringForced("24.0001"), fxp.One, BlastOutOfRange},
		{"struck directly", StruckDirectly, ExplosionInAir, fxp.Hundred, fxp.One, BlastUndivided},
		{"threw himself on it", ThrewSelfOnExplosive, ExplosionInAir, fxp.Hundred, fxp.One, BlastUndivided},
		{"inside him", ExplosiveInsideTarget, ExplosionInAir, fxp.Hundred, fxp.One, BlastUndivided},
	} {
		divisor, reach := BlastDivisor(blast, tc.situation, tc.env, tc.distance)
		c.Equal(tc.divisor, divisor, "%s: divisor", tc.name)
		c.Equal(tc.reach, reach, "%s: reach", tc.name)
	}
}

// TestFragmentation verifies the fragmentation roll of BX415 for a [2d] attack: 15 less the range penalty, the
// posture penalty of a ground burst but not an airburst, and the size modifier, with no roll for a target the
// explosive struck and none at all beyond the 10 yards the fragments reach.
func TestFragmentation(t *testing.T) {
	c := check.New(t)
	fragments := dice.Dice{Count: 2, Sides: 6}

	r := Fragmentation(fragments, CaughtInBlast, StandingTarget, fxp.Three, 0, false)
	c.Equal(FragmentationResult{Radius: 10, Skill: 14, RangePenalty: -1, GroundBurst: true}, r,
		"standing 3 yards from a ground burst, the fragments roll against 15 - 1")

	r = Fragmentation(fragments, CaughtInBlast, ProneTarget, fxp.Three, 0, false)
	c.Equal(FragmentationResult{Radius: 10, Skill: 10, RangePenalty: -1, PosturePenalty: -4, GroundBurst: true, HalfExposed: true}, r,
		"lying down 3 yards from a ground burst shows it only half a torso, for -4")

	r = Fragmentation(fragments, CaughtInBlast, ProneTarget, 0, 0, false)
	c.Equal(FragmentationResult{Radius: 10, Skill: 13, PosturePenalty: -2}, r,
		"lying down at the center of the blast is not a ground burst, so the posture alone counts")

	r = Fragmentation(fragments, CaughtInBlast, ProneTarget, fxp.Three, 0, true)
	c.Equal(FragmentationResult{Radius: 10, Skill: 14, RangePenalty: -1}, r,
		"an airburst ignores posture")

	r = Fragmentation(fragments, CaughtInBlast, StandingTarget, fxp.Three, 2, false)
	c.Equal(16, r.Skill, "the size modifier adds to the roll")

	r = Fragmentation(fragments, CaughtInBlast, StandingTarget, fxp.FromStringForced("10.5"), 0, false)
	c.Equal(FragmentationResult{Radius: 10, OutOfRange: true}, r, "beyond 10 yards the fragments do not reach")

	r = Fragmentation(fragments, StruckDirectly, StandingTarget, fxp.Hundred, 0, false)
	c.Equal(FragmentationResult{Radius: 10, Automatic: true}, r, "a target the explosive struck is hit without a roll")
}

// TestDissipationDivisor verifies the three ways an area attack's damage can fall off with distance (BX414).
func TestDissipationDivisor(t *testing.T) {
	c := check.New(t)
	divisor, htBonus := DissipationDivisor(fxp.Three, false, false)
	c.Equal(fxp.One, divisor, "an attack that does not dissipate is not divided")
	c.Equal(fxp.Int(0), htBonus, "nor eased")
	divisor, htBonus = DissipationDivisor(fxp.Three, true, false)
	c.Equal(fxp.Three, divisor, "one that dissipates is divided by the spread")
	c.Equal(fxp.Int(0), htBonus, "and not eased")
	divisor, htBonus = DissipationDivisor(fxp.Three, true, true)
	c.Equal(fxp.One, divisor, "one that HT resists is not divided")
	c.Equal(fxp.Three, htBonus, "the spread eases the roll instead")
}

// TestDamageTypes verifies the damage type helpers: the base type DR is looked up against, and the Explosion modifier
// an explosion's type always shows.
func TestDamageTypes(t *testing.T) {
	c := check.New(t)
	c.Equal("cr", BaseDamageType("cr ex"), "the first word is the base type")
	c.Equal("cr", BaseDamageType(" cr ex "), "padding is ignored")
	c.Equal("", BaseDamageType(""), "a blank type has no base")
	c.Equal("", BaseDamageType("   "), "nor does a padded blank")

	c.Equal("cr ex", ExplosionDamageType("cr ex"), "a type that says it explodes is kept")
	c.Equal("cr ex", ExplosionDamageType(" cr ex "), "and trimmed")
	c.Equal("cr ex", ExplosionDamageType("cr"), "one that does not has the modifier added")
	c.Equal("ex", ExplosionDamageType(""), "a blank type is the modifier alone")
	c.Equal("burn ex*", ExplosionDamageType("burn ex*"), "a decorated modifier still counts")
	c.Equal("cr exp ex", ExplosionDamageType("cr exp"), "a longer word is not the modifier")
}

// TestExposedFilter verifies the test LargeAreaDR is given: none when every location is exposed, otherwise one that
// accepts the named location alone.
func TestExposedFilter(t *testing.T) {
	c := check.New(t)
	c.Nil(ExposedFilter(nil, false), "with no location named, every location is exposed")
	loc := gurps.NewHitLocation(nil, "")
	c.Nil(ExposedFilter(loc, true), "every location is exposed to a true area effect")
	filter := ExposedFilter(loc, false)
	c.NotNil(filter, "a named location gives a filter")
	c.True(filter(loc), "which accepts that location")
	c.False(filter(gurps.NewHitLocation(nil, "")), "and no other")
}

// TestExplosionChoices verifies the explosion tables: their order, which the calculator's indexes depend on, and the
// facts each entry carries.
func TestExplosionChoices(t *testing.T) {
	c := check.New(t)

	environments := ExplosionEnvironmentChoices()
	checkNamed(c, "environment", environments, func(e ExplosionEnvironmentChoice) string { return e.Name })
	c.Equal([]ExplosionEnvironment{ExplosionInAir, ExplosionUnderwater, ExplosionInVacuum},
		[]ExplosionEnvironment{environments[0].Environment, environments[1].Environment, environments[2].Environment},
		"air comes first, as the usual case")

	postures := TargetPostureChoices()
	checkNamed(c, "posture", postures, func(p TargetPostureChoice) string { return p.Name })
	c.Equal([]TargetPosture{StandingTarget, CrouchingTarget, ProneTarget},
		[]TargetPosture{postures[0].Posture, postures[1].Posture, postures[2].Posture},
		"standing comes first, as the usual case")

	situations := BlastSituationChoices()
	checkNamed(c, "situation", situations, func(s BlastSituationChoice) string { return s.Name })
	c.Equal([]BlastSituation{CaughtInBlast, StruckDirectly, ThrewSelfOnExplosive, ExplosiveInsideTarget},
		[]BlastSituation{situations[0].Situation, situations[1].Situation, situations[2].Situation, situations[3].Situation},
		"being caught in the blast comes first, as the usual case")

	causes := ScatterCauses()
	checkNamed(c, "scatter cause", causes, func(s ScatterCause) string { return s.Name })
	c.Equal([]bool{false, true, false}, []bool{causes[0].Squared, causes[1].Squared, causes[2].Squared},
		"only the squared miss squares the margin; a dodge never does")
}
