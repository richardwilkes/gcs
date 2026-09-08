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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/rpgtools/dice"
)

// CollisionAngle names the geometry of a collision, which decides the velocity the two objects meet at (BX430).
type CollisionAngle byte

// The possible CollisionAngle values.
const (
	HeadOnCollision  CollisionAngle = iota // Closing head-on: the velocities add.
	RearEndCollision                       // The striker overtakes the struck object: the velocities subtract.
	SideOnCollision                        // A side-on hit, or a hit on a stationary object: the striker's velocity.
)

// fallingVelocityTableMaxYards is the largest distance the printed Falling Velocity Table (BX431) covers. Past this
// point FallingVelocity switches to the formula, and the seam is clean: the table's last row and the formula both give
// 49 yards per second.
const fallingVelocityTableMaxYards = 112

// fallingVelocityTable is the printed Falling Velocity Table (BX431) for one gravity, one row per printed entry, giving
// the largest distance in yards the row covers and the velocity in yards per second reached over it.
//
// The table is encoded rather than computed because it does not match the formula the same page offers as an
// alternative. Twenty-nine of its 112 entries are exactly one yard per second higher than round(sqrt(21.4 x distance)),
// and no adjustment of the constant reproduces the whole table, so the printed values win wherever they exist.
var fallingVelocityTable = []struct {
	maxYards int
	velocity int
}{
	{maxYards: 1, velocity: 5},
	{maxYards: 2, velocity: 7},
	{maxYards: 3, velocity: 8},
	{maxYards: 4, velocity: 9},
	{maxYards: 5, velocity: 10},
	{maxYards: 6, velocity: 11},
	{maxYards: 7, velocity: 12},
	{maxYards: 8, velocity: 13},
	{maxYards: 9, velocity: 14},
	{maxYards: 11, velocity: 15},
	{maxYards: 12, velocity: 16},
	{maxYards: 14, velocity: 17},
	{maxYards: 15, velocity: 18},
	{maxYards: 17, velocity: 19},
	{maxYards: 19, velocity: 20},
	{maxYards: 21, velocity: 21},
	{maxYards: 23, velocity: 22},
	{maxYards: 25, velocity: 23},
	{maxYards: 27, velocity: 24},
	{maxYards: 29, velocity: 25},
	{maxYards: 32, velocity: 26},
	{maxYards: 34, velocity: 27},
	{maxYards: 37, velocity: 28},
	{maxYards: 39, velocity: 29},
	{maxYards: 42, velocity: 30},
	{maxYards: 45, velocity: 31},
	{maxYards: 48, velocity: 32},
	{maxYards: 51, velocity: 33},
	{maxYards: 54, velocity: 34},
	{maxYards: 57, velocity: 35},
	{maxYards: 61, velocity: 36},
	{maxYards: 64, velocity: 37},
	{maxYards: 67, velocity: 38},
	{maxYards: 71, velocity: 39},
	{maxYards: 75, velocity: 40},
	{maxYards: 79, velocity: 41},
	{maxYards: 82, velocity: 42},
	{maxYards: 86, velocity: 43},
	{maxYards: 90, velocity: 44},
	{maxYards: 95, velocity: 45},
	{maxYards: 99, velocity: 46},
	{maxYards: 103, velocity: 47},
	{maxYards: 108, velocity: 48},
	{maxYards: fallingVelocityTableMaxYards, velocity: 49},
}

// FallingVelocity returns the velocity in yards per second reached by falling the given distance in yards under the
// given gravity, expressed in Gs (BX431). A distance or gravity of zero or less yields zero.
//
// At one gravity, and out to the 112 yards the printed Falling Velocity Table covers, the table is consulted: the
// distance is rounded up to a whole yard and the first row that reaches it supplies the velocity. Everywhere else the
// alternative formula on the same page is used instead: round(sqrt(21.4 x gravity x distance)), computed in floating
// point since the square root has no fixed-point form.
func FallingVelocity(distanceYards, gravity fxp.Int) fxp.Int {
	if distanceYards <= 0 || gravity <= 0 {
		return 0
	}
	if gravity == fxp.One {
		if yards := distanceYards.Ceil().AsInteger[int](); yards <= fallingVelocityTableMaxYards {
			for _, row := range fallingVelocityTable {
				if yards <= row.maxYards {
					return fxp.FromInteger(row.velocity)
				}
			}
		}
	}
	return fxp.FromFloat(math.Round(math.Sqrt(21.4 * gravity.AsFloat[float64]() * distanceYards.AsFloat[float64]())))
}

// TerminalVelocity returns the velocity in yards per second at which a fall stops accelerating, given the base velocity
// for the falling object's shape and posture -- 60 for a spread-eagle human, 100 for one in a swan dive, and 200 or
// more for something dense and streamlined -- the gravity in Gs and the atmospheric pressure in atmospheres (BX431). The
// base velocity is scaled by the square root of the gravity and divided by the square root of the pressure.
//
// A pressure of zero or less is a vacuum, where a fall never stops accelerating, so unlimited is returned as true and
// the velocity as zero. A base velocity or gravity of zero or less yields a velocity of zero that is not unlimited.
func TerminalVelocity(baseVelocity, gravity, pressure fxp.Int) (velocity fxp.Int, unlimited bool) {
	if pressure <= 0 {
		return 0, true
	}
	if baseVelocity <= 0 || gravity <= 0 {
		return 0, false
	}
	return fxp.FromFloat(baseVelocity.AsFloat[float64]() * math.Sqrt(gravity.AsFloat[float64]()) /
		math.Sqrt(pressure.AsFloat[float64]())), false
}

// CollisionDiceCount returns the number of dice of crushing damage a collision inflicts, as (HP x velocity)/100 (BX430).
// Negative HP or velocity is treated as zero.
//
// The count is deliberately left fractional, since the rules act on the fraction: a count below one die maps onto the
// 1d-3 / 1d-2 / 1d-1 steps, and the caps that hold one object's damage down to the other's compare counts rather than
// dice. Callers halve this count for a bullet-shaped, sharp or spiked object and apply those caps to it before handing
// it to CollisionDamageDice, so that the halving, the caps and the sub-one-die steps all work on the same scale. Note
// that fxp.Int carries four decimal places, so the product truncates at the fourth place.
func CollisionDiceCount(hp, velocity fxp.Int) fxp.Int {
	return hp.Max(0).Mul(velocity.Max(0)).Div(fxp.Hundred)
}

// CollisionDamageDice converts a fractional dice count from CollisionDiceCount into the dice of crushing damage
// actually rolled (BX430). Below a full die the count maps onto fixed steps: a quarter of a die or less is 1d-3, a half
// or less is 1d-2, and anything short of a full die is 1d-1. From one die up, the count is rounded to the nearest whole
// die, with a fraction of exactly a half rounding up. A count of zero or less produces no dice at all.
func CollisionDamageDice(count fxp.Int) dice.Dice {
	switch {
	case count <= 0:
		return dice.Dice{Sides: 6, Multiplier: 1}
	case count <= fxp.Quarter:
		return dice.Dice{Count: 1, Sides: 6, Modifier: -3, Multiplier: 1}
	case count <= fxp.Half:
		return dice.Dice{Count: 1, Sides: 6, Modifier: -2, Multiplier: 1}
	case count < fxp.One:
		return dice.Dice{Count: 1, Sides: 6, Modifier: -1, Multiplier: 1}
	default:
		// fxp.Int.Round rounds half away from zero, which is what "round a fraction of 0.5 or more up" asks for.
		return dice.Dice{Count: count.Round().AsInteger[int](), Sides: 6, Multiplier: 1}
	}
}

// CollisionVelocity returns the velocity in yards per second that the two objects in a collision meet at, given the
// angle of the collision and each object's own velocity (BX430). Negative velocities are treated as zero, as is a
// rear-end collision in which the struck object is the faster of the two.
func CollisionVelocity(angle CollisionAngle, strikerVelocity, struckVelocity fxp.Int) fxp.Int {
	striker := strikerVelocity.Max(0)
	struck := struckVelocity.Max(0)
	switch angle {
	case HeadOnCollision:
		return striker + struck
	case RearEndCollision:
		return (striker - struck).Max(0)
	default: // SideOnCollision, and anything unrecognized: only the striker's velocity counts.
		return striker
	}
}

// CollisionObject describes one of the two objects in a collision.
type CollisionObject struct {
	HP         fxp.Int // Its HP.
	Velocity   fxp.Int // Its own velocity in yards per second, before the angle of the collision is applied.
	HalfDamage bool    // Whether it is bullet-shaped, sharp or spiked, and so does half damage (BX430).
}

// diceFor returns the fractional dice this object inflicts at the given collision velocity, halved if it is shaped to
// do half damage. The halving happens here, before any cap is applied, so that a spiked striker cannot be out-damaged
// by an object that was capped at its full, unhalved count.
func (c CollisionObject) diceFor(velocity fxp.Int) fxp.Int {
	count := CollisionDiceCount(c.HP, velocity)
	if c.HalfDamage {
		count = count.Div(fxp.Two)
	}
	return count
}

// CollisionResult holds what a collision between two moving objects does to each of them.
type CollisionResult struct {
	Velocity      fxp.Int // The velocity the two met at, in yards per second.
	StrikerDice   fxp.Int // The fractional dice the striker inflicts on the struck object, after halving and any cap.
	StruckDice    fxp.Int // The fractional dice the struck object inflicts on the striker, after halving and any cap.
	StrikerCapped bool    // Whether the striker's dice were held down to the struck object's (head-on, striker slower).
	StruckCapped  bool    // Whether the struck object's dice were held down to the striker's.
}

// Collision returns the damage a collision between two objects does to each of them (BX430). Each object inflicts the
// dice its own HP and the collision velocity call for, halved if it is shaped to do half damage, and then one of them
// may be capped.
//
// In a head-on collision the slower of the two -- the one with the lower velocity of its own -- cannot inflict more
// dice than the faster one; if the two are moving at the same speed, neither is capped. In a rear-end or side-on
// collision the struck object cannot inflict more dice than the striker. Whichever side was held down is reported by
// StrikerCapped or StruckCapped.
//
// Convert the resulting counts to dice with CollisionDamageDice.
func Collision(angle CollisionAngle, striker, struck CollisionObject) CollisionResult {
	velocity := CollisionVelocity(angle, striker.Velocity, struck.Velocity)
	result := CollisionResult{
		Velocity:    velocity,
		StrikerDice: striker.diceFor(velocity),
		StruckDice:  struck.diceFor(velocity),
	}
	if angle == HeadOnCollision {
		strikerOwn := striker.Velocity.Max(0)
		struckOwn := struck.Velocity.Max(0)
		if strikerOwn == struckOwn {
			return result // Neither is the slower one, so neither is capped.
		}
		if strikerOwn < struckOwn {
			if result.StrikerDice > result.StruckDice {
				result.StrikerDice = result.StruckDice
				result.StrikerCapped = true
			}
			return result
		}
	}
	if result.StruckDice > result.StrikerDice {
		result.StruckDice = result.StrikerDice
		result.StruckCapped = true
	}
	return result
}

// ImmovableCollisionHP returns the HP to compute collision damage from when an object slams into something that will
// not move (BX430). The mover inflicts its usual collision damage on the obstacle and takes the same damage itself, but
// a hard obstacle -- the ground, concrete, a building -- doubles the mover's HP for that calculation, while a soft one
// leaves it alone. Negative HP is treated as zero.
func ImmovableCollisionHP(moverHP fxp.Int, hard bool) fxp.Int {
	if hard {
		return moverHP.Max(0).Mul(fxp.Two)
	}
	return moverHP.Max(0)
}

// BluntTraumaFromFall returns the injury that leaks through armor that stopped the given amount of falling damage
// (BX431). Every five full points stopped inflicts one point of injury, since the falling rules count all armor as
// flexible. Negative damage is treated as zero.
func BluntTraumaFromFall(damageStopped fxp.Int) int {
	return damageStopped.Max(0).Div(fxp.Five).Floor().AsInteger[int]()
}
