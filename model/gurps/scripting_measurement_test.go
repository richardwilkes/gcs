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
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestSSRTToYardsTableValues pins the Size/Speed/Range Table entries ssrtToYards produces, including the repeating
// 10/15/20/30/50/70 pattern that steps up by a factor of ten every six entries.
func TestSSRTToYardsTableValues(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		value int
		want  fxp.Int
	}{
		{value: -15, want: fxp.One.Div(fxp.Five).Div(fxp.ThirtySix)},
		{value: -11, want: fxp.One.Div(fxp.ThirtySix)},
		{value: -5, want: fxp.One.Div(fxp.Three)},
		{value: -2, want: fxp.One},
		{value: 0, want: fxp.Two},
		{value: 3, want: fxp.Seven},
		{value: 4, want: fxp.Ten},
		{value: 5, want: fxp.Fifteen},
		{value: 6, want: fxp.Twenty},
		{value: 7, want: fxp.Thirty},
		{value: 8, want: fxp.Fifty},
		{value: 9, want: fxp.Seventy},
		{value: 10, want: fxp.Hundred},
		{value: 16, want: fxp.Thousand},
		{value: maxSSRTValue, want: fxp.FromInteger(700000000000000)},
	} {
		c.Equal(tc.want, ssrtToYards(tc.value), "value %d", tc.value)
	}
}

// TestSSRTToYardsClamps verifies that values beyond either end of the table are clamped to it. The low end has always
// been clamped; the high end must be too, since the value arrives straight from a script and the yardage grows without
// bound while fxp.Int does not.
func TestSSRTToYardsClamps(t *testing.T) {
	c := check.New(t)
	low := ssrtToYards(minSSRTValue)
	for _, value := range []int{minSSRTValue - 1, -100, -1000000, math.MinInt} {
		c.Equal(low, ssrtToYards(value), "value %d", value)
	}
	high := ssrtToYards(maxSSRTValue)
	for _, value := range []int{maxSSRTValue + 1, 100, 1000000, 1_000_000_000_000_000, math.MaxInt} {
		c.Equal(high, ssrtToYards(value), "value %d", value)
	}
}

// TestSSRTToYardsNeverSaturates verifies that every value the table accepts yields an exact result rather than one that
// hit the top of the fixed-point range. fxp.Int.Mul saturates at fxp.Max instead of wrapping, so an unbounded value
// would silently report ~9.2e14 yards for every input past 87 rather than reporting anything meaningful. Strict
// monotonicity across the whole range is checked at the same time, since a saturated entry would show up as a
// repeated value.
func TestSSRTToYardsNeverSaturates(t *testing.T) {
	c := check.New(t)
	prev := fxp.Int(0)
	for value := minSSRTValue; value <= maxSSRTValue; value++ {
		got := ssrtToYards(value)
		c.True(got > 0, "value %d yielded %v", value, got)
		c.True(got < fxp.Max, "value %d saturated at %v", value, got)
		c.True(got > prev, "value %d (%v) should exceed value %d (%v)", value, got, value-1, prev)
		prev = got
	}

	// One past the end would need 1e15 yards, which fxp.Int cannot hold; that is why maxSSRTValue is where it is.
	c.Equal(fxp.Max, fxp.FromInteger(100000000000000).Mul(fxp.Ten))
}

// TestScriptModifierToYardsIsBounded verifies that measure.modifierToYards returns promptly no matter what a script
// hands it. The conversion loop runs (value-4)/6 times inside a Go function, and goja only honors an Interrupt between
// VM instructions, so the per-script timeout cannot preempt it: before the value was clamped, a call such as
// measure.modifierToYards(1e15) would have spun here for weeks with the application unresponsive.
func TestScriptModifierToYardsIsBounded(t *testing.T) {
	c := check.New(t)

	// Warm the VM pool and the compiled-program cache so the timings below measure the conversion itself.
	_, err := runScript(0, "measure.modifierToYards(0)")
	c.NoError(err)

	for _, tc := range []struct {
		expr string
		want string
	}{
		{expr: "measure.modifierToYards(4)", want: "10"},
		{expr: "measure.modifierToYards(87)", want: "700000000000000"},
		{expr: "measure.modifierToYards(88)", want: "700000000000000"},
		{expr: "measure.modifierToYards(1e8)", want: "700000000000000"},
		{expr: "measure.modifierToYards(1e15)", want: "700000000000000"},
		{expr: "measure.modifierToYards(1e300)", want: "700000000000000"},
		{expr: "measure.modifierToYards(-1e300)", want: "0.0055"},
	} {
		start := time.Now()
		var v string
		v, err = runScript(0, tc.expr)
		elapsed := time.Since(start)
		c.NoError(err, "expr %q", tc.expr)
		c.Equal(tc.want, v, "expr %q", tc.expr)
		// Generously loose: the clamped conversion is microseconds, while an unclamped 1e15 would take weeks. The
		// bound only has to separate "returns" from "never returns" without being flaky on a loaded machine.
		c.True(elapsed < 5*time.Second, "expr %q took %v", tc.expr, elapsed)
	}
}

// TestSSRTRoundTrip verifies that the yardage ssrtToYards produces maps back to the value it came from through
// ssrtInchesToValue, tying the two directions of the table together so that a change to either one that drifts from the
// other is caught.
//
// The round trip stops short of maxSSRTValue because ssrtInchesToValue works in inches: multiplying the yardage by 36
// saturates once the result passes fxp.Max, at which point the remaining entries are indistinguishable from one
// another. That is a limit of the inches-based direction and not of the table — ssrtToYards returns yards, and
// measure.modifierToYards, its only caller, is unaffected — so the loop covers everything up to that point.
func TestSSRTRoundTrip(t *testing.T) {
	c := check.New(t)
	covered := 0
	for value := minSSRTValue; value <= maxSSRTValue; value++ {
		yards := ssrtToYards(value)
		inches := fxp.Yard.ToInches(yards)
		if inches >= fxp.Max {
			break
		}
		c.Equal(value, ssrtInchesToValue(inches, true), "value %d (%v yards)", value, yards)
		covered++
	}
	c.Equal(94, covered, "the round trip should cover -15 through 78")
}
