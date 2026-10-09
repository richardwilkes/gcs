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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
)

// DemolitionDamageType is what an explosive charge inflicts: crushing damage with the Explosion modifier (BX415).
const DemolitionDamageType = "cr ex"

// BlastSituation is where the target was when the explosion went off, which decides whether the collateral damage is
// divided at all (BX414-BX415).
type BlastSituation byte

// The possible BlastSituation values.
const (
	CaughtInBlast         BlastSituation = iota // Some distance from the center, taking the divided collateral damage.
	StruckDirectly                              // Hit by the explosive itself, taking the listed damage.
	ThrewSelfOnExplosive                        // Threw himself on the explosive: maximum damage, DR protecting normally.
	ExplosiveInsideTarget                       // The explosive went off inside him: an attack on the vitals with no DR.
)

// BlastReach says how an explosion's collateral damage reaches a target, which BlastDivisor reports alongside the
// divisor.
type BlastReach byte

// The possible BlastReach values.
const (
	BlastCollateral BlastReach = iota // Caught in the blast some distance from the center: the damage is divided.
	BlastUndivided                    // Struck directly, threw himself on the explosive or had it go off inside him.
	BlastAtCenter                     // Caught in the blast at its very center, where nothing divides the damage.
	BlastOutOfRange                   // Beyond the collateral damage radius, and so unhurt by the blast.
)

// BlastDivisor returns what the rolled damage of an explosion is divided by for a target in the given situation at the
// given distance in yards from the center of the blast, along with how the blast reaches it (BX414-BX415). A divisor
// of one leaves the damage alone, which is what every reach but BlastCollateral comes to.
func BlastDivisor(blast dice.Dice, situation BlastSituation, env ExplosionEnvironment, distanceYards fxp.Int) (divisor fxp.Int, reach BlastReach) {
	if situation != CaughtInBlast {
		return fxp.One, BlastUndivided
	}
	if distanceYards > fxp.FromInteger(CollateralDamageRadius(blast)) {
		return fxp.One, BlastOutOfRange
	}
	if distanceYards <= 0 {
		return fxp.One, BlastAtCenter
	}
	return env.CollateralDivisor(distanceYards), BlastCollateral
}

// FragmentationResult is what the fragments of an explosion do to a target (BX414-BX415).
type FragmentationResult struct {
	Radius         int  // The radius in yards the fragments reach.
	Skill          int  // The effective skill the fragments attack at; meaningful only when a roll is needed.
	RangePenalty   int  // The range penalty that skill includes.
	PosturePenalty int  // The posture penalty that skill includes; zero for an airburst.
	Automatic      bool // Whether a fragment hits without a roll, because the explosive struck the target itself.
	OutOfRange     bool // Whether the target is beyond the fragments' reach.
	GroundBurst    bool // Whether the blast was at ground level some distance from the target, so posture protects.
	HalfExposed    bool // Whether a crawling or lying-down target shows the fragments only half a torso (BX551).
}

// Fragmentation works out what the fragments of an explosion do to a target in the given situation and posture at the
// given distance in yards from the center of the blast, with the given size modifier (BX414-BX415). A blast on the
// ground is at the target's own elevation and, unless the target is at its center, farther away than any attacker's
// height, so a crawling or lying-down target shows it only half a torso (BX551); an airburst comes from above, so
// posture does not protect against it at all. Call this only for an explosive that lists fragmentation, i.e. when
// DiceOfDamage(fragments) is more than zero.
func Fragmentation(fragments dice.Dice, situation BlastSituation, posture TargetPosture, distanceYards fxp.Int, sizeModifier int, airburst bool) FragmentationResult {
	r := FragmentationResult{Radius: FragmentationRadius(fragments)}
	switch {
	case situation != CaughtInBlast:
		r.Automatic = true
	case distanceYards > fxp.FromInteger(r.Radius):
		r.OutOfRange = true
	default:
		r.GroundBurst = !airburst && distanceYards > 0
		if !airburst {
			r.PosturePenalty = posture.FragmentPenalty(r.GroundBurst)
		}
		r.RangePenalty = gurps.SpeedRangePenalty(distanceYards)
		r.Skill = FragmentationSkill(distanceYards, r.PosturePenalty, sizeModifier)
		r.HalfExposed = r.GroundBurst && posture == ProneTarget
	}
	return r
}

// DissipationDivisor returns what an area-effect or cone attack's damage is divided by for a target the given spread
// from its center or apex -- AreaDamageDivisor for an area, ConeWidth for a cone -- along with any bonus to a HT roll
// instead (BX414). An attack that does not dissipate is never divided. One that is resisted by a HT roll eases the
// roll by the spread instead of dividing the damage, so the divisor is one and htBonus carries the spread.
func DissipationDivisor(spread fxp.Int, dissipates, htResisted bool) (divisor, htBonus fxp.Int) {
	switch {
	case !dissipates:
		return fxp.One, 0
	case htResisted:
		return fxp.One, spread
	default:
		return spread, 0
	}
}

// BaseDamageType returns the type DR is looked up against for the given damage type: its first word, the "cr" of
// "cr ex", or "" for a blank type.
func BaseDamageType(damageType string) string {
	if fields := strings.Fields(damageType); len(fields) > 0 {
		return fields[0]
	}
	return ""
}

// ExplosionDamageType returns the damage type to show for an explosion, which always carries the Explosion modifier
// (B104): the given type, trimmed, with "ex" added when it does not say so already.
func ExplosionDamageType(damageType string) string {
	damageType = strings.TrimSpace(damageType)
	switch {
	case gurps.IsExplosiveDamageType(damageType):
		return damageType
	case damageType == "":
		return "ex"
	default:
		return damageType + " ex"
	}
}

// ExposedFilter returns the test LargeAreaDR uses to decide which locations face the attack: nil when every location
// is exposed, as it is to a true area effect or when no location is named, and otherwise one that accepts the named
// location alone.
func ExposedFilter(location *gurps.HitLocation, areaEffect bool) func(*gurps.HitLocation) bool {
	if location == nil || areaEffect {
		return nil
	}
	return func(loc *gurps.HitLocation) bool { return loc == location }
}
