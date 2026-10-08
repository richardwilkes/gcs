// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import "github.com/richardwilkes/gcs/v5/model/fxp"

// FallResult is what a fall comes to (BX431): the velocity reached, the distance it was worked out from and what, if
// anything, held it down.
type FallResult struct {
	Velocity        fxp.Int // The velocity reached, in yards per second, after any terminal velocity is applied.
	Distance        fxp.Int // The distance the velocity was worked out from: the fall, less 5 yards when Controlled.
	Terminal        fxp.Int // The terminal velocity the fall was held to; meaningful only when TerminalLimited.
	Controlled      bool    // Whether a successful Acrobatics roll shortened the fall by 5 yards.
	NoGravity       bool    // Whether there was no gravity to fall under, in which case nothing else was considered.
	Vacuum          bool    // Whether a terminal velocity was asked for in a vacuum, where there is none.
	TerminalLimited bool    // Whether the terminal velocity held the velocity down.
}

// Fall works out the velocity reached in a fall of the given distance in yards under the given gravity in Gs (BX431).
// A controlled fall, one softened by a successful Acrobatics roll, counts as 5 yards shorter. terminalBase is the
// terminal velocity at one gravity in one atmosphere for the faller's shape, or zero or less for no limit; when there
// is one, TerminalVelocity scales it to the gravity and the pressure in atmospheres and the velocity is held to it,
// unless the pressure is zero or less, since in a vacuum there is no terminal velocity.
func Fall(distanceYards, gravity, terminalBase, pressure fxp.Int, controlled bool) FallResult {
	r := FallResult{Distance: distanceYards, Controlled: controlled}
	if controlled {
		r.Distance = (distanceYards - fxp.Five).Max(0)
	}
	r.Velocity = FallingVelocity(r.Distance, gravity)
	if gravity <= 0 {
		r.NoGravity = true
		return r
	}
	if terminalBase <= 0 {
		return r
	}
	limit, unlimited := TerminalVelocity(terminalBase, gravity, pressure)
	switch {
	case unlimited:
		r.Vacuum = true
	case r.Velocity > limit:
		r.Velocity = limit
		r.Terminal = limit
		r.TerminalLimited = true
	}
	return r
}

// ImmovableCollisionDice returns the fractional dice of damage a collision with something that will not move inflicts
// on both the mover and the obstacle (BX430-BX431): the mover's HP, doubled for a hard obstacle, at the mover's
// velocity, halved if the mover is shaped to do half damage. Convert the count to dice with CollisionDamageDice.
func ImmovableCollisionDice(mover CollisionObject, hard bool) fxp.Int {
	return CollisionObject{HP: ImmovableCollisionHP(mover.HP, hard), HalfDamage: mover.HalfDamage}.diceFor(mover.Velocity)
}

// OverrunST returns the ST whose thrust damage a striking object inflicts on top of the collision when it is at least
// two Size Modifier steps bigger than the object it hits, which it overruns (BX432): half its ST, rounded down, or half
// its HP when it has no ST score, as a vehicle has not. overruns is false when the striker is not big enough.
func OverrunST(strikerSM, struckSM int, strikerST, strikerHP fxp.Int) (st fxp.Int, overruns bool) {
	if strikerSM < struckSM+2 {
		return 0, false
	}
	st = strikerST
	if st <= 0 {
		st = strikerHP
	}
	return st.Div(fxp.Two).Floor(), true
}

// DroppedObjectHampers reports whether an object dropped onto a victim is at least as big as the victim, in which case
// the victim may move only one yard on his next turn and his active defenses are at -3 (BX431).
func DroppedObjectHampers(objectSM, victimSM int) bool {
	return objectSM >= victimSM
}
