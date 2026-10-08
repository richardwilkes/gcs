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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// ThrowingTier is a level of a throwing skill, relative to DX, and what it adds to a throw (BX355).
type ThrowingTier struct {
	Name     string
	Distance int // What it adds to the ST the distance is worked out from.
	Damage   int // What it adds per die to the damage.
}

// String implements fmt.Stringer.
func (t ThrowingTier) String() string {
	return t.Name
}

// ThrowingTiers returns the levels of the Throwing skill that matter to a throw (BX355): at DX+1 it adds 1 to the ST
// the distance is worked out from, and at DX+2 or better it adds 2. It does nothing for the damage.
func ThrowingTiers() []ThrowingTier {
	return []ThrowingTier{
		{Name: i18n.Text("None, or below DX+1")},
		{Name: i18n.Text("DX+1"), Distance: 1},
		{Name: i18n.Text("DX+2 or better"), Distance: 2},
	}
}

// ThrowingArtTiers returns the levels of the Throwing Art skill that matter to a throw (BX355): at DX it adds 1 to the
// ST the distance is worked out from and 1 per die to the damage, and at DX+1 or better it adds 2 of each.
func ThrowingArtTiers() []ThrowingTier {
	return []ThrowingTier{
		{Name: i18n.Text("None, or below DX")},
		{Name: i18n.Text("DX"), Distance: 1, Damage: 1},
		{Name: i18n.Text("DX+1 or better"), Distance: 2, Damage: 2},
	}
}

// Thrower is the character throwing an object (BX355).
type Thrower struct {
	Entity      *gurps.Entity // Supplies the Basic Lift and damage progression; nil uses the global default settings.
	ST          fxp.Int       // The ST the distance is worked out from.
	StrikingST  fxp.Int       // The Striking ST the damage is worked out from.
	Throwing    ThrowingTier  // The level of the Throwing skill, from ThrowingTiers.
	ThrowingArt ThrowingTier  // The level of the Throwing Art skill, from ThrowingArtTiers.
}

// ThrowResult is what a throw comes to (BX355).
type ThrowResult struct {
	Distance fxp.Int   // The distance thrown, in inches; zero when nothing is thrown.
	Damage   dice.Dice // The damage the object does when it lands; meaningful only when Distance is more than zero.
	TooHeavy bool      // Whether the object is too heavy to throw at all.
}

// Throw returns the distance the object of the given weight is thrown and the damage it does, with the given extra
// effort penalty adding 5% per -1 to the ST both are worked out from (BX355, BX357). A weight of zero or less throws
// nothing and yields the zero result.
func (t *Thrower) Throw(objectWeight fxp.Weight, extraEffortPenalty int) ThrowResult {
	if objectWeight <= 0 {
		return ThrowResult{}
	}
	distanceBonus := max(t.Throwing.Distance, t.ThrowingArt.Distance)
	damageBonus := t.ThrowingArt.Damage

	st := extraEffortST(t.ST, extraEffortPenalty) + fxp.FromInteger(distanceBonus)
	basicLift := BasicLiftFor(t.Entity, st)
	var weightRatio fxp.Int
	if basicLift > 0 {
		weightRatio = fxp.Int(objectWeight).Div(fxp.Int(basicLift))
	}
	inches := st.Mul(throwingDistanceModifier(weightRatio)).Mul(fxp.ThirtySix).Floor()
	if inches <= fxp.One {
		return ThrowResult{TooHeavy: true}
	}

	thrust := ThrustFor(t.Entity, extraEffortST(t.StrikingST, extraEffortPenalty))
	thrust.Modifier += thrust.Count * damageBonus
	basicLift = BasicLiftFor(t.Entity, st-fxp.FromInteger(distanceBonus))
	if basicLift > 0 {
		weightRatio = fxp.Int(objectWeight).Div(fxp.Int(basicLift))
	} else {
		weightRatio = 0
	}
	switch {
	case weightRatio <= fxp.Eighth:
		thrust.Modifier -= thrust.Count * 2
	case weightRatio <= fxp.Quarter:
		thrust.Modifier -= thrust.Count
	case weightRatio <= fxp.Half:
	case weightRatio <= fxp.One:
		thrust.Modifier += thrust.Count
	case weightRatio <= fxp.Two:
	case weightRatio <= fxp.Four:
		thrust.Modifier -= thrust.Count / 2
	default:
		thrust.Modifier -= thrust.Count
	}
	return ThrowResult{Distance: inches, Damage: thrust}
}

// throwingDistanceModifier returns what the ST is multiplied by to get the distance in yards, from the Throwing
// Distance Table (BX355), for an object weighing the given multiple of the thrower's Basic Lift. An object heavier
// than the table reaches cannot be thrown, so its modifier is zero.
func throwingDistanceModifier(weightRatio fxp.Int) fxp.Int {
	switch {
	case weightRatio <= fxp.Twentieth:
		return fxp.ThreeAndAHalf
	case weightRatio <= fxp.Tenth:
		return fxp.TwoAndAHalf
	case weightRatio <= fxp.PointOneFive:
		return fxp.Two
	case weightRatio <= fxp.Fifth:
		return fxp.OneAndAHalf
	case weightRatio <= fxp.Quarter:
		return fxp.OnePointTwo
	case weightRatio <= fxp.ThreeTenths:
		return fxp.OnePointOne
	case weightRatio <= fxp.TwoFifths:
		return fxp.One
	case weightRatio <= fxp.Half:
		return fxp.FourFifths
	case weightRatio <= fxp.ThreeQuarters:
		return fxp.SevenTenths
	case weightRatio <= fxp.One:
		return fxp.ThreeFifths
	case weightRatio <= fxp.OneAndAHalf:
		return fxp.TwoFifths
	case weightRatio <= fxp.Two:
		return fxp.ThreeTenths
	case weightRatio <= fxp.TwoAndAHalf:
		return fxp.Quarter
	case weightRatio <= fxp.Three:
		return fxp.Fifth
	case weightRatio <= fxp.Four:
		return fxp.PointOneFive
	case weightRatio <= fxp.Five:
		return fxp.PointOneTwo
	case weightRatio <= fxp.Six:
		return fxp.Tenth
	case weightRatio <= fxp.Seven:
		return fxp.PointZeroNine
	case weightRatio <= fxp.Eight:
		return fxp.PointZeroEight
	case weightRatio <= fxp.Nine:
		return fxp.PointZeroSeven
	case weightRatio <= fxp.Ten:
		return fxp.PointZeroSix
	case weightRatio <= fxp.Twelve:
		return fxp.Twentieth
	default:
		return 0
	}
}

// extraEffortST returns the ST with the given extra effort penalty applied, which adds 5% per -1 (BX357).
func extraEffortST(st fxp.Int, extraEffortPenalty int) fxp.Int {
	if extraEffortPenalty >= 0 {
		return st
	}
	return st.Mul(fxp.FromInteger(-5*extraEffortPenalty).Div(fxp.Hundred) + fxp.One).Floor()
}

// ThrowerStats is what a character sheet supplies about a thrower.
type ThrowerStats struct {
	ST               fxp.Int // The ST the distance is worked out from: the lifting ST without its lifting-only bonus.
	StrikingST       fxp.Int
	ThrowingIndex    int // The index into ThrowingTiers of the character's level in Throwing.
	ThrowingArtIndex int // The index into ThrowingArtTiers of the character's level in Throwing Art.
}

// ThrowerStatsFromEntity reads the thrower's numbers from the sheet's character. The entity must not be nil.
func ThrowerStatsFromEntity(entity *gurps.Entity) ThrowerStats {
	return ThrowerStats{
		ST:               (entity.LiftingStrength() - entity.LiftingStrengthBonus).Max(0),
		StrikingST:       entity.StrikingStrength().Max(0),
		ThrowingIndex:    throwingTierIndex(entity, "Throwing", fxp.One),
		ThrowingArtIndex: throwingTierIndex(entity, "Throwing Art", 0),
	}
}

// throwingTierIndex returns the tier the entity's level in the named skill falls in: 0 below the given relative level,
// 1 at it, and 2 a full level or more above it.
func throwingTierIndex(entity *gurps.Entity, name string, first fxp.Int) int {
	sk := entity.BestSkillNamed(name, "", false, nil)
	if sk == nil {
		return 0
	}
	switch relative := sk.CalculateLevel(nil).RelativeLevel; {
	case relative >= first+fxp.One:
		return 2
	case relative >= first:
		return 1
	default:
		return 0
	}
}
