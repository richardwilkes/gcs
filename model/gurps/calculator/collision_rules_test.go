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
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestFall verifies the fall rules of BX431 as the calculator combines them: the velocity from the Falling Velocity
// Table, the 5 yards a controlled fall takes off, the terminal velocity that holds a long fall down, the vacuum that
// has no terminal velocity, and the absence of a fall without gravity.
func TestFall(t *testing.T) {
	c := check.New(t)
	seventeen := fxp.FromInteger(17)

	r := Fall(seventeen, fxp.One, 0, fxp.One, false)
	c.Equal(fxp.FromInteger(19), r.Velocity, "17 yards reaches 19 yards/second")
	c.Equal(seventeen, r.Distance, "the whole fall counts")
	c.False(r.Controlled || r.NoGravity || r.Vacuum || r.TerminalLimited, "nothing limited the fall")

	r = Fall(seventeen, fxp.One, 0, fxp.One, true)
	c.Equal(fxp.Twelve, r.Distance, "a controlled fall counts as 5 yards shorter")
	c.Equal(fxp.Sixteen, r.Velocity, "12 yards reaches 16 yards/second")
	c.True(r.Controlled, "the controlled fall is reported")

	r = Fall(fxp.Three, fxp.One, 0, fxp.One, true)
	c.Equal(fxp.Int(0), r.Distance, "a controlled fall never counts as less than nothing")
	c.Equal(fxp.Int(0), r.Velocity, "and reaches no velocity")

	r = Fall(fxp.Thousand, fxp.One, fxp.Sixty, fxp.One, false)
	c.Equal(fxp.Sixty, r.Velocity, "a long fall is held to the terminal velocity")
	c.Equal(fxp.Sixty, r.Terminal, "which is reported")
	c.True(r.TerminalLimited, "along with the fact that it applied")

	r = Fall(fxp.Thousand, fxp.One, fxp.Sixty, 0, false)
	c.True(r.Vacuum, "in a vacuum there is no terminal velocity")
	c.False(r.TerminalLimited, "so nothing holds the fall down")
	c.Equal(FallingVelocity(fxp.Thousand, fxp.One), r.Velocity, "and the formula's velocity stands")

	r = Fall(fxp.Thousand, fxp.One, 0, 0, false)
	c.False(r.Vacuum, "without a terminal velocity to apply, the vacuum is not mentioned")

	r = Fall(seventeen, 0, fxp.Sixty, 0, false)
	c.True(r.NoGravity, "without gravity there is no fall")
	c.Equal(fxp.Int(0), r.Velocity, "and no velocity")
	c.False(r.Vacuum, "and the terminal velocity is not considered")
}

// TestImmovableCollisionDice verifies that hitting something immovable uses the mover's HP, doubled for a hard
// obstacle, and halves the dice for a mover shaped to do half damage.
func TestImmovableCollisionDice(t *testing.T) {
	c := check.New(t)
	mover := CollisionObject{HP: fxp.Ten, Velocity: fxp.FromInteger(19)}
	c.Equal(fxp.FromStringForced("3.8"), ImmovableCollisionDice(mover, true), "a hard surface doubles the HP: 20 x 19 / 100")
	c.Equal(fxp.FromStringForced("1.9"), ImmovableCollisionDice(mover, false), "a soft one does not: 10 x 19 / 100")
	mover.HalfDamage = true
	c.Equal(fxp.FromStringForced("1.9"), ImmovableCollisionDice(mover, true), "a spiked mover does half damage")
}

// TestOverrunST verifies the overrun rule of BX432: a striker at least two Size Modifier steps bigger than what it
// hits inflicts thrust damage for half its ST, or half its HP when it has no ST score.
func TestOverrunST(t *testing.T) {
	c := check.New(t)
	st, overruns := OverrunST(2, 0, 0, fxp.Sixty)
	c.True(overruns, "a 60 HP car at SM +2 overruns a pedestrian at SM 0")
	c.Equal(fxp.Thirty, st, "with no ST score, half its HP stands in")
	st, overruns = OverrunST(4, 2, fxp.Fifteen, fxp.Sixty)
	c.True(overruns, "the two steps are relative to the struck object's size")
	c.Equal(fxp.Seven, st, "half of ST 15, rounded down")
	_, overruns = OverrunST(1, 0, fxp.Fifteen, fxp.Sixty)
	c.False(overruns, "one step bigger is not enough")
}

// TestDroppedObjectHampers verifies that a dropped object hampers the victim when it is at least as big.
func TestDroppedObjectHampers(t *testing.T) {
	c := check.New(t)
	c.True(DroppedObjectHampers(0, 0), "an object as big as the victim hampers him")
	c.True(DroppedObjectHampers(1, 0), "as does a bigger one")
	c.False(DroppedObjectHampers(-1, 0), "a smaller one does not")
}
