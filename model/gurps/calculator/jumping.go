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
	"math"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
)

// Jumper is the character making a jump (BX352).
type Jumper struct {
	Entity       *gurps.Entity     // Supplies the Basic Lift for LiftingST; nil uses the global default sheet settings.
	BasicMove    fxp.Int           // Basic Move.
	LiftingST    fxp.Int           // Lifting ST.
	Weight       fxp.Weight        // Body weight; a Basic Lift above it allows ST/4 as the Basic Move.
	Encumbrance  encumbrance.Level // Each level takes 20% off the distance.
	JumpingSkill int               // The Jumping skill level, or 0 for none; half of it may replace the Basic Move.
	EnhancedMove fxp.Int           // Levels of Enhanced Move (Ground), which multiply the Basic Move of a running start.
	SuperJump    fxp.Int           // Levels of Super Jump, each of which doubles the distance.
}

// JumperFromEntity reads the jumper's numbers from the sheet's character, which becomes its Entity. The entity must
// not be nil.
func JumperFromEntity(entity *gurps.Entity) Jumper {
	j := Jumper{
		Entity:      entity,
		BasicMove:   entity.ResolveAttributeCurrent(gurps.BasicMoveID).Max(0),
		LiftingST:   entity.LiftingStrength().Max(0),
		Weight:      entity.Profile.Weight,
		Encumbrance: entity.EncumbranceLevel(false),
	}
	if sk := entity.BestSkillNamed("Jumping", "", false, nil); sk != nil {
		j.JumpingSkill = sk.CalculateLevel(nil).Level.Max(0).AsInteger[int]()
	}
	j.EnhancedMove, _ = entity.TraitLevels("enhanced move (ground)")
	j.SuperJump, _ = entity.TraitLevels("super jump")
	return j
}

// HighJump returns the height of a high jump in inches (BX352), with the given running start in yards (a meter counts
// as a yard) and the given extra effort penalty adding 5% per -1 to the distance but not to the Basic Move it is worked
// out from (BX356).
func (j *Jumper) HighJump(runningStart fxp.Int, extraEffortPenalty int) fxp.Int {
	return j.jump(false, runningStart, extraEffortPenalty)
}

// BroadJump returns the distance of a broad jump in inches (BX352), with the given running start in yards (a meter
// counts as a yard) and the given extra effort penalty adding 5% per -1 to the distance but not to the Basic Move it
// is worked out from (BX356).
func (j *Jumper) BroadJump(runningStart fxp.Int, extraEffortPenalty int) fxp.Int {
	return j.jump(true, runningStart, extraEffortPenalty)
}

// jump returns the distance of a high jump, or a broad jump when broad is set, in inches.
func (j *Jumper) jump(broad bool, runningStart fxp.Int, extraEffortPenalty int) fxp.Int {
	basicMove := j.BasicMove
	basicMoveWithoutRun := basicMove

	if runningStart > 0 {
		basicMove += runningStart
		if j.EnhancedMove > 0 {
			if adjusted := basicMoveWithoutRun.Mul(j.EnhancedMove + fxp.One); adjusted > basicMove {
				basicMove = adjusted
			}
		}
	}

	if j.JumpingSkill > 0 {
		level := fxp.FromInteger(j.JumpingSkill).Div(fxp.Two).Floor()
		basicMove = basicMove.Max(level)
		basicMoveWithoutRun = basicMoveWithoutRun.Max(level)
	}

	// Adjust Basic Move for high strength
	if basicLift := BasicLiftFor(j.Entity, j.LiftingST); basicLift > j.Weight {
		adjusted := j.LiftingST.Div(fxp.Four).Floor()
		basicMove = basicMove.Max(adjusted)
		basicMoveWithoutRun = basicMoveWithoutRun.Max(adjusted)
	}

	// The base distance, which a running start can at most double.
	var multiplier, reduction fxp.Int
	if broad {
		multiplier = fxp.Two
		reduction = fxp.Three
	} else {
		multiplier = fxp.Six
		reduction = fxp.Ten
	}
	distance := (basicMove.Mul(multiplier) - reduction).Min((basicMoveWithoutRun.Mul(multiplier) - reduction).Mul(fxp.Two))

	distance = distance.Mul(fxp.One - fxp.FromInteger(int(j.Encumbrance)).Mul(fxp.Two).Div(fxp.Ten))

	if j.SuperJump > 0 {
		distance = distance.Mul(fxp.FromFloat(math.Pow(2, j.SuperJump.AsFloat[float64]())))
	}

	if extraEffortPenalty < 0 {
		distance = distance.Mul(fxp.FromInteger(-5*extraEffortPenalty).Div(fxp.Hundred) + fxp.One)
	}
	if broad {
		distance = distance.Mul(fxp.Twelve)
	}
	return distance.Floor()
}
