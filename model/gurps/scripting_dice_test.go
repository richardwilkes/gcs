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
	"testing"

	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestIntFromScript verifies that a number supplied by a script is narrowed to an int identically on every
// architecture. Go leaves the conversion of an out-of-range float64 to an integer implementation-defined: arm64
// saturates to MaxInt64 while amd64 produces MinInt64. goja performs that conversion itself when it maps a JS number
// onto a Go int parameter, so any binding that accepts an int is handed a different value depending on the machine.
func TestIntFromScript(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		value float64
		want  int
	}{
		{value: math.MaxFloat64, want: math.MaxInt},
		{value: -math.MaxFloat64, want: math.MinInt},
		{value: 1e300, want: math.MaxInt},
		{value: -1e300, want: math.MinInt},
		{value: math.Inf(1), want: math.MaxInt},
		{value: math.Inf(-1), want: math.MinInt},
		{value: math.NaN(), want: 0}, // NaN compares false against everything, so it needs its own case.
		{value: 0, want: 0},
		{value: 6, want: 6},
		{value: -6, want: -6},
		{value: 6.9, want: 6},   // Narrowing truncates toward zero.
		{value: -6.9, want: -6}, //
		{value: 1e15, want: 1000000000000000},
	} {
		c.Equal(tc.want, intFromScript(tc.value), "value %v", tc.value)
	}
}

// TestScriptDiceFromIsArchitectureIndependent verifies that dice.from clamps a component to the Roller's configured
// maximum rather than wrapping to the opposite extreme. Letting goja narrow the arguments made dice.from(1e300, 6)
// produce "999999d" on arm64 and "0" on amd64.
func TestScriptDiceFromIsArchitectureIndependent(t *testing.T) {
	c := check.New(t)
	cfg := Roller.Config()
	for _, tc := range []struct {
		expr string
		want string
	}{
		// Each component in turn, pushed past what an int can hold.
		{expr: "dice.from(1e300, 6, 0, 1)", want: Roller.Format(diceOf(cfg.MaxCount, 6, 0, 1))},
		{expr: "dice.from(2, 1e300, 0, 1)", want: Roller.Format(diceOf(2, cfg.MaxSides, 0, 1))},
		{expr: "dice.from(2, 6, 1e300, 1)", want: Roller.Format(diceOf(2, 6, cfg.MaxModifier, 1))},
		{expr: "dice.from(2, 6, 0, 1e300)", want: Roller.Format(diceOf(2, 6, 0, cfg.MaxMultiplier))},
		{expr: "dice.from(Infinity, 6, 0, 1)", want: Roller.Format(diceOf(cfg.MaxCount, 6, 0, 1))},

		// The negative extreme clamps to the low end, not the high one.
		{expr: "dice.from(-1e300, 6, 0, 1)", want: Roller.Format(diceOf(0, 6, 0, 1))},
		{expr: "dice.from(2, 6, -1e300, 1)", want: Roller.Format(diceOf(2, 6, -cfg.MaxModifier, 1))},

		// Ordinary values, and the dice(sides) shorthand, are untouched.
		{expr: "dice.from(3, 6, 2, 1)", want: Roller.Format(diceOf(3, 6, 2, 1))},
		{expr: "dice.from(6)", want: Roller.Format(diceOf(1, 6, 0, 1))},
		{expr: "dice.from(3, 6)", want: Roller.Format(diceOf(3, 6, 0, 1))},
	} {
		v, err := runScript(0, tc.expr)
		c.NoError(err, "expr %q", tc.expr)
		c.Equal(tc.want, v, "expr %q", tc.expr)
	}
}

// diceOf builds a dice.Dice from its components, so the expectations above can be written in terms of the Roller's
// configured limits rather than hard-coded strings that would drift if those limits changed.
func diceOf(count, sides, modifier, multiplier int) dice.Dice {
	return dice.Dice{Count: count, Sides: sides, Modifier: modifier, Multiplier: multiplier}
}
