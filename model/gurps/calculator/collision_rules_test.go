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
// Table, the 5 yards a controlled fall takes off, the terminal velocity that holds a long fall down but not a short
// one, the vacuum that has no terminal velocity, and the absence of a fall without gravity.
func TestFall(t *testing.T) {
	c := check.New(t)
	seventeen := fxp.FromInteger(17)
	for _, tc := range []struct {
		name         string
		distance     fxp.Int
		gravity      fxp.Int
		terminalBase fxp.Int
		pressure     fxp.Int
		controlled   bool
		want         FallResult
	}{
		{
			name:     "17 yards reaches 19 yards/second",
			distance: seventeen,
			gravity:  fxp.One,
			pressure: fxp.One,
			want:     FallResult{Velocity: fxp.Nineteen, Distance: seventeen},
		},
		{
			name:       "a controlled fall counts as 5 yards shorter, and 12 yards reaches 16 yards/second",
			distance:   seventeen,
			gravity:    fxp.One,
			pressure:   fxp.One,
			controlled: true,
			want:       FallResult{Velocity: fxp.Sixteen, Distance: fxp.Twelve, Controlled: true},
		},
		{
			name:       "a controlled fall never counts as less than nothing",
			distance:   fxp.Three,
			gravity:    fxp.One,
			pressure:   fxp.One,
			controlled: true,
			want:       FallResult{Controlled: true},
		},
		{
			name:         "a long fall is held to the terminal velocity",
			distance:     fxp.Thousand,
			gravity:      fxp.One,
			terminalBase: fxp.Sixty,
			pressure:     fxp.One,
			want:         FallResult{Velocity: fxp.Sixty, Distance: fxp.Thousand, Terminal: fxp.Sixty, TerminalLimited: true},
		},
		{
			name:         "a short fall stays under its terminal velocity",
			distance:     seventeen,
			gravity:      fxp.One,
			terminalBase: fxp.Sixty,
			pressure:     fxp.One,
			want:         FallResult{Velocity: fxp.Nineteen, Distance: seventeen},
		},
		{
			name:         "in a vacuum there is no terminal velocity, so the formula's velocity stands",
			distance:     fxp.Thousand,
			gravity:      fxp.One,
			terminalBase: fxp.Sixty,
			want:         FallResult{Velocity: fxp.FromInteger(146), Distance: fxp.Thousand, Vacuum: true},
		},
		{
			name:     "without a terminal velocity to apply, the vacuum is not mentioned",
			distance: fxp.Thousand,
			gravity:  fxp.One,
			want:     FallResult{Velocity: fxp.FromInteger(146), Distance: fxp.Thousand},
		},
		{
			name:         "without gravity there is no fall, and the terminal velocity is not considered",
			distance:     seventeen,
			terminalBase: fxp.Sixty,
			want:         FallResult{Distance: seventeen, NoGravity: true},
		},
	} {
		c.Equal(tc.want, Fall(tc.distance, tc.gravity, tc.terminalBase, tc.pressure, tc.controlled), tc.name)
	}
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
