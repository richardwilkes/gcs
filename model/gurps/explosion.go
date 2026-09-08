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
	"math"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
)

// ExplosionEnvironment is the medium the blast spreads through, which sets the divisor applied to the collateral
// damage reaching anything the explosion did not strike directly (BX414-BX415).
type ExplosionEnvironment byte

// The possible ExplosionEnvironment values.
const (
	ExplosionInAir      ExplosionEnvironment = iota // Divide the damage by 3 x the distance in yards (BX414).
	ExplosionUnderwater                             // Water carries the blast, so divide by the distance alone (BX415).
	ExplosionInVacuum                               // Vacuum or a trace atmosphere: divide by 10 x the distance (BX415).
)

// Modifiers the explosion and area attack rules apply.
const (
	// FragmentationBaseSkill is the skill fragments attack at when the explosive did not strike the target itself
	// (BX415). Only three modifiers apply to that roll: range, the target's posture and the target's size modifier.
	FragmentationBaseSkill = 15
	// ExtraFragmentMargin is how many full points of success on the fragmentation roll each fragment past the first
	// costs (BX415).
	ExtraFragmentMargin = 3
	// NonStandingTargetPenalty is the penalty to be hit at range that any posture other than standing imposes (BX551).
	NonStandingTargetPenalty = -2
	// HalfExposedTorsoPenalty is the further penalty to hit a crawling or lying-down target from the same or a lower
	// elevation and farther away than the attacker's own height, when only half of its torso is exposed (the Posture
	// Table's note, BX551).
	HalfExposedTorsoPenalty = -2
	// AreaAttackBonus is the bonus to hit when deliberately attacking an area rather than a target standing in it
	// (BX414). The area cannot defend, though anyone in it may dive for cover.
	AreaAttackBonus = 4
)

// TargetPosture is how a target is placed when the fragments of an explosion arrive, grouped by the modifier to be hit
// at range that the Posture Table gives (BX551). It is the only posture modifier the fragmentation roll takes (BX415).
type TargetPosture byte

// The possible TargetPosture values.
const (
	StandingTarget  TargetPosture = iota // No modifier.
	CrouchingTarget                      // Crouching, kneeling or sitting: NonStandingTargetPenalty.
	ProneTarget                          // Crawling or lying down: NonStandingTargetPenalty, plus HalfExposedTorsoPenalty from a ground burst.
)

// FragmentPenalty returns the posture modifier the fragments take against a target in this posture (BX551). groundBurst
// reports whether the blast is at the same or a lower elevation than the target and farther away than the attacker's
// own height, which a blast on the ground some distance from the target always is; a crawling or lying-down target then
// shows only half of its torso and takes HalfExposedTorsoPenalty on top of its posture's own penalty, and the groin,
// legs and feet cannot be hit at all. Posture does not protect against an airburst (BX415), so the caller passes
// StandingTarget for one. An unrecognized posture counts as standing.
func (p TargetPosture) FragmentPenalty(groundBurst bool) int {
	switch p {
	case CrouchingTarget:
		return NonStandingTargetPenalty
	case ProneTarget:
		if groundBurst {
			return NonStandingTargetPenalty + HalfExposedTorsoPenalty
		}
		return NonStandingTargetPenalty
	default:
		return 0
	}
}

// CollateralDivisor returns the amount the rolled damage is divided by for something the given distance in yards from
// the center of the blast (BX414-BX415). A distance of zero or less means the target was struck directly and so takes
// the listed damage: fxp.One is returned, since dividing by one leaves the damage alone. An unrecognized environment
// is treated as air.
func (e ExplosionEnvironment) CollateralDivisor(distanceYards fxp.Int) fxp.Int {
	if distanceYards <= 0 {
		return fxp.One
	}
	switch e {
	case ExplosionUnderwater:
		return distanceYards
	case ExplosionInVacuum:
		return fxp.Ten.Mul(distanceYards)
	default: // ExplosionInAir, and anything unrecognized.
		return fxp.Three.Mul(distanceYards)
	}
}

// IsExplosiveDamageType reports whether a damage type carries the Explosion modifier (B104), i.e. an "ex" token, as in
// the "cr ex" of a grenade (BX414).
//
// Each whitespace-separated token is matched on its leading letters alone, because the damage types in the data
// libraries decorate the token in several ways -- "burn ex* rad sur", "cr ex/2", "ex," -- and every one of those is
// still an explosion. Matching only the leading letters keeps a longer word that merely starts with them, such as
// "exp", from counting.
func IsExplosiveDamageType(damageType string) bool {
	for token := range strings.FieldsSeq(strings.ToLower(damageType)) {
		i := 0
		for i < len(token) && token[i] >= 'a' && token[i] <= 'z' {
			i++
		}
		if token[:i] == "ex" {
			return true
		}
	}
	return false
}

// DiceOfDamage returns the number of dice of damage the explosion radii are measured in (BX414). A dice multiplier
// counts toward it, so 6dx2 is twelve dice of damage rather than six. A negative count is treated as zero and a
// multiplier below one as one, since dice are never rolled a negative number of times.
func DiceOfDamage(d dice.Dice) int {
	return max(d.Count, 0) * max(d.Multiplier, 1)
}

// CollateralDamageRadius returns the radius in yards within which everything is vulnerable to an explosion's collateral
// damage: twice the dice of damage (BX414). A 6dx2 blast is twelve dice, so everything within 24 yards is at risk.
func CollateralDamageRadius(d dice.Dice) int {
	return 2 * DiceOfDamage(d)
}

// FragmentationRadius returns the radius in yards within which everything is vulnerable to an explosion's fragments:
// five times the dice of fragmentation damage (BX414). A [2d] fragmentation attack reaches 10 yards.
func FragmentationRadius(d dice.Dice) int {
	return 5 * DiceOfDamage(d)
}

// FragmentationSkill returns the effective skill the fragments from an explosion attack a target at, given its distance
// in yards from the center of the blast, the penalty for its posture and its size modifier (BX415). Fragments start at
// FragmentationBaseSkill and take only three modifiers: the range penalty from the Size and Speed/Range Table (BX550),
// the posture penalty (TargetPosture.FragmentPenalty, or zero for an airburst or a standing target) and the size modifier.
// Every full ExtraFragmentMargin points of success beyond the first hit lands one more fragment.
//
// A target the explosive attack struck directly is hit by a fragment automatically and does not roll at all.
func FragmentationSkill(distanceYards fxp.Int, posturePenalty, sizeModifier int) int {
	return FragmentationBaseSkill + SpeedRangePenalty(distanceYards) + posturePenalty + sizeModifier
}

// DividedDamage returns the least, average and greatest damage the given dice inflict once divided by the given divisor
// and rounded down (BX414). Pass the divisor from ExplosionEnvironment.CollateralDivisor or AreaDamageDivisor; a divisor
// of one or less leaves the damage alone, which is what a direct hit and an area attack that does not dissipate both
// want.
//
// The three values are computed from the dice's own fields rather than rolled, so this needs no roller: the least roll
// puts a 1 on every die, the greatest the number of sides, and the average is halfway between, with the modifier added
// and the multiplier applied to each. A total below zero is reported as zero, since damage never heals the target.
func DividedDamage(d dice.Dice, divisor fxp.Int) (minimum, average, maximum int) {
	count := fxp.FromInteger(max(d.Count, 0))
	sides := fxp.FromInteger(max(d.Sides, 0))
	modifier := fxp.FromInteger(d.Modifier)
	multiplier := fxp.FromInteger(max(d.Multiplier, 1))
	divide := func(damage fxp.Int) int {
		if divisor > fxp.One {
			damage = damage.Div(divisor)
		}
		return damage.Max(0).Floor().AsInteger[int]()
	}
	return divide((count + modifier).Mul(multiplier)),
		divide((count.Mul(sides+fxp.One).Div(fxp.Two) + modifier).Mul(multiplier)),
		divide((count.Mul(sides) + modifier).Mul(multiplier))
}

// LargeAreaDR returns the effective DR against a large-area injury (BX400), such as the collateral damage from an
// explosion or a cone or area-effect attack: the average of the torso DR and the DR of the least-protected exposed
// location, rounded up. damageType picks the DR when it varies by attack ("" means DR against everything); exposed,
// when not nil, reports whether a location faces the attack, and a nil exposed counts every location, as a true area
// effect does. Negative DR counts as zero. Zero is returned for an entity with no body type or no exposed location.
//
// Only call this with a real entity: a nil one would be answered from the global default body type rather than the
// target's own.
func LargeAreaDR(entity *Entity, damageType string, exposed func(*HitLocation) bool) int {
	body := SheetSettingsFor(entity).BodyType
	if body == nil {
		return 0
	}
	drFor := func(loc *HitLocation) int {
		// The DR map is keyed by lowercased specialization, with the DR that applies to everything under AllID and
		// each specialized bonus stacking on top of it.
		m := loc.DR(entity, nil, nil)
		dr := m[AllID]
		if t := strings.ToLower(strings.TrimSpace(damageType)); t != "" && t != AllID {
			dr += m[t]
		}
		return max(dr, 0)
	}
	least, found := math.MaxInt, false
	for _, loc := range body.UniqueHitLocations(entity) { // Includes sub-table locations; primes the lookup itself.
		if exposed == nil || exposed(loc) {
			least, found = min(least, drFor(loc)), true
		}
	}
	if !found {
		return 0
	}
	torso := least // With no torso location, the least-protected DR stands in for both halves of the average.
	if loc := body.LookupLocationByID(entity, TorsoID); loc != nil {
		torso = drFor(loc)
	}
	return (torso + least + 1) / 2 // BX400 rounds the average up.
}

// CoverDRFromBody returns the DR the body of someone who threw himself onto an explosive gives everyone else: his torso
// DR plus his HP (BX415). He himself takes the maximum possible damage, with his own DR protecting normally. Negative
// values count as zero and a fractional HP total is rounded down.
func CoverDRFromBody(torsoDR int, hp fxp.Int) int {
	return max(torsoDR, 0) + hp.Max(0).Floor().AsInteger[int]()
}

// TNTForBlast returns the pounds of TNT needed to produce a 6dxn explosion: (n x n)/4 (BX415). An n of zero or less
// needs nothing.
func TNTForBlast(n fxp.Int) fxp.Int {
	if n <= 0 {
		return 0
	}
	return n.Mul(n).Div(fxp.Four)
}

// ExplosiveForBlast returns the pounds of an explosive with the given relative explosive force needed to produce a 6dxn
// explosion: the weight of TNT the blast needs, divided by the REF (BX415). The BX415 example: 6dx8 of dynamite, whose
// REF is 0.8, is (8x8)/(4x0.8) = 20 lbs. A REF of zero or less -- an explosive that does not explode -- yields zero
// rather than an unbounded weight.
func ExplosiveForBlast(n, ref fxp.Int) fxp.Int {
	if ref <= 0 {
		return 0
	}
	return TNTForBlast(n).Div(ref)
}

// BlastForExplosive returns the n of the 6dxn explosion a given weight in pounds of an explosive with the given
// relative explosive force produces, i.e. TNTForBlast and ExplosiveForBlast run backwards: n = sqrt(4 x pounds x REF)
// (BX415). A weight or REF of zero or less yields zero.
//
// The square root has no fixed-point form, so it is computed in floating point, as FallingVelocity does. The result is
// deliberately left fractional, since a weight rarely lands on a whole number of dice and the caller decides how to
// present the remainder.
func BlastForExplosive(pounds, ref fxp.Int) fxp.Int {
	if pounds <= 0 || ref <= 0 {
		return 0
	}
	return fxp.FromFloat(math.Sqrt(4 * pounds.AsFloat[float64]() * ref.AsFloat[float64]()))
}

// ScatterDistance returns how many yards an attack that missed lands from where it was aimed, and whether that distance
// was held down by the cap (BX414). The distance is the margin of failure in yards, or the margin of success if the
// target dodged instead. Set squared for the cases that scatter by the square of the margin: a flying or underwater
// target, or an Artillery or Dropping attack on a target the attacker cannot see -- never for a dodge.
//
// Whatever the case, the scatter cannot exceed half the distance to the target, rounded up, so an attack that misses
// badly still lands closer to the target than to the attacker. A margin of zero or less does not scatter at all.
//
// The direction is rolled separately: 1d, where a 1 is the direction the attacker faces and each higher number turns 60
// degrees further clockwise.
func ScatterDistance(margin int, distanceYards fxp.Int, squared bool) (yards fxp.Int, capped bool) {
	if margin <= 0 {
		return 0, false
	}
	if squared {
		margin *= margin
	}
	scatter := fxp.FromInteger(margin)
	limit := distanceYards.Max(0).Div(fxp.Two).Ceil()
	if scatter > limit {
		return limit, true
	}
	return scatter, false
}

// ConeWidth returns how many yards wide a cone attack is at the given distance from its apex (BX413). A cone is one yard
// wide where it starts and spreads by its maximum width divided by its maximum range for every yard of range, so at 60
// yards a cone with a maximum range of 100 yards and a maximum width of 5 yards is 3 yards wide. A cone whose width is
// unspecified -- a maximum width or range of zero or less -- spreads a yard per yard of range instead. The width never
// drops below the one yard the cone starts at.
//
// The width is left fractional; Dissipation (BX414) divides a cone's damage by it, and AreaDamageDivisor does the same
// job for an area effect measured from its center.
func ConeWidth(distanceYards, maxRange, maxWidth fxp.Int) fxp.Int {
	distance := distanceYards.Max(0)
	if maxWidth <= 0 || maxRange <= 0 {
		return distance.Max(fxp.One)
	}
	return distance.Mul(maxWidth).Div(maxRange).Max(fxp.One)
}

// AreaDamageDivisor returns the amount a dissipating area-effect attack's damage is divided by at the given distance in
// yards from the center of the area: the distance itself, never less than one (Dissipation, BX414). For an attack
// resisted by a HT roll the same number is a bonus to that roll instead of a divisor, so 2 yards from the center is +2
// to HT rather than half damage.
func AreaDamageDivisor(distanceYards fxp.Int) fxp.Int {
	return distanceYards.Max(fxp.One)
}

// ExplosiveType is a row of the Relative Explosive Force Table (BX415). The REF scales a weight of the explosive against
// the same weight of TNT, so ExplosiveForBlast divides by it.
type ExplosiveType struct {
	Name        string  // The name of the explosive, as printed.
	Description string  // The note the table carries for it.
	REF         fxp.Int // Its relative explosive force, with TNT as 1.
	TL          int     // The tech level it becomes available at.
}

// ExplosiveTypes is the Relative Explosive Force Table (BX415), in the order it is printed. Black Powder appears twice,
// since its REF improves at TL5.
var ExplosiveTypes = []ExplosiveType{
	{Name: "Serpentine Powder", Description: "Standard gunpowder, pre-1600.", REF: fxp.ThreeTenths, TL: 3},
	{Name: "Ammonium Nitrate", Description: "Common improvised explosive.", REF: fxp.TwoFifths, TL: 4},
	{Name: "Black Powder", Description: "Standard gunpowder, 1600-1850.", REF: fxp.TwoFifths, TL: 4},
	{Name: "Black Powder", Description: "Standard gunpowder, 1850-1890.", REF: fxp.Half, TL: 5},
	{Name: "Diesel Fuel/Nitrate Fertilizer", Description: "Common improvised explosive.", REF: fxp.Half, TL: 6},
	{Name: "Dynamite", Description: "Commercially available for mining, demolition.", REF: fxp.FourFifths, TL: 6},
	{Name: "TNT", Description: "The basic, stable, high explosive.", REF: fxp.One, TL: 6},
	{Name: "Amatol", Description: "TNT-ammonium nitrate. Fills bombs & shells in WWII.", REF: fxp.OnePointTwo, TL: 6},
	{Name: "Nitroglycerine", Description: "Unstable! If dropped, detonates on 13+ on 3d.", REF: fxp.OneAndAHalf, TL: 6},
	{
		Name:        "Tetryl",
		Description: "Common for smaller explosive shells and bullets.",
		REF:         fxp.FromStringForced("1.3"),
		TL:          7,
	},
	{
		Name:        "Composition B",
		Description: "Another common explosive filler.",
		REF:         fxp.FromStringForced("1.4"),
		TL:          7,
	},
	{
		Name:        "C4 Plastic Explosive",
		Description: "Standard military and covert-ops explosive.",
		REF:         fxp.FromStringForced("1.4"),
		TL:          7,
	},
	{Name: "Octanitrocubane", Description: "Theoretical advanced explosive.", REF: fxp.Four, TL: 9},
	{Name: "Stabilized Metallic Hydrogen", Description: "Exotic science-fiction explosive.", REF: fxp.Six, TL: 10},
}
