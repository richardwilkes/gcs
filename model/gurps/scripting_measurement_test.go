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
	"strconv"
	"strings"
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

// TestScriptModifierUsesSaturatedLength verifies that measure.modifier agrees with measure.rangeModifier and
// measure.sizeModifier for the same physical length, including one too large for the fixed-point type to hold.
// measure.modifier used to format the length into a string and parse it back, which turned the range error that
// conversion now reports into a flat 0 — so measure.modifier(1e300, "yd", false) answered 0 while
// measure.rangeModifier(1e300) answered -79.
func TestScriptModifierUsesSaturatedLength(t *testing.T) {
	c := check.New(t)
	for _, yards := range []string{"0.5", "2", "100", "1e6", "9.2e14", "1e15", "1e300", "-1e300"} {
		rangeMod, err := runScript(0, "measure.rangeModifier("+yards+")")
		c.NoError(err, "yards %s", yards)
		sizeMod, err := runScript(0, "measure.sizeModifier("+yards+")")
		c.NoError(err, "yards %s", yards)

		v, err := runScript(0, "measure.modifier("+yards+", 'yd', false)")
		c.NoError(err, "yards %s", yards)
		c.Equal(rangeMod, v, "measure.modifier(%s, 'yd', false) should match measure.rangeModifier", yards)

		v, err = runScript(0, "measure.modifier("+yards+", 'yd', true)")
		c.NoError(err, "yards %s", yards)
		c.Equal(sizeMod, v, "measure.modifier(%s, 'yd', true) should match measure.sizeModifier", yards)
	}

	// The out-of-range case specifically: the modifier for a saturated length, not the 0 a failed parse produced.
	saturated := ssrtInchesToValue(fxp.Yard.ToInches(fxp.Max), true)
	v, err := runScript(0, "measure.modifier(1e300, 'yd', true)")
	c.NoError(err)
	c.Equal(strconv.Itoa(saturated), v)
	c.True(saturated > 0, "a saturated length should yield a large positive size modifier, got %d", saturated)
}

// TestScriptModifierUnitResolution verifies that every length unit a script can name resolves to the same conversion
// the other measurement bindings use, and that an unrecognized or empty unit still falls back to yards, as it did when
// these units were resolved by parsing a formatted string.
func TestScriptModifierUnitResolution(t *testing.T) {
	c := check.New(t)
	for _, unit := range fxp.LengthUnits {
		c.Equal(unit, lengthUnitForScript(unit.Key()), "key %q", unit.Key())
		c.Equal(unit, lengthUnitForScript(strings.ToUpper(unit.Key())), "key %q uppercased", unit.Key())
		c.Equal(unit, lengthUnitForScript(" "+unit.Key()+" "), "key %q padded", unit.Key())
	}
	for _, units := range []string{"", "   ", "yards", "bogus", "'"} {
		c.Equal(fxp.Yard, lengthUnitForScript(units), "units %q should fall back to yards", units)
	}

	// End to end: naming a unit must scale the length by that unit before the modifier is computed.
	for _, tc := range []struct {
		units string
		value string
	}{
		{units: "yd", value: "100"},
		{units: "ft", value: "300"},
		{units: "in", value: "3600"},
		{units: "mi", value: "0.0568"},
	} {
		want := ssrtInchesToValue(lengthUnitForScript(tc.units).ToInches(fxp.FromStringForced(tc.value)), true)
		v, err := runScript(0, "measure.modifier("+tc.value+", '"+tc.units+"', true)")
		c.NoError(err, "units %q", tc.units)
		c.Equal(strconv.Itoa(want), v, "units %q", tc.units)
	}
}

// TestSSRTValueFromScriptIsArchitectureIndependent verifies that a script-supplied number reaches the table as the same
// value on every architecture.
//
// Go leaves the conversion of an out-of-range float64 to an integer implementation-defined. arm64 saturates to
// MaxInt64, which clamps to the top of the table; amd64 produces MinInt64, which clamps to the bottom. Narrowing before
// clamping therefore made measure.modifierToYards(1e300) return 700000000000000 on Apple silicon and 0.0055 on Intel —
// a split that passed CI on all three arm64 runners and failed on all three amd64 runners.
func TestSSRTValueFromScriptIsArchitectureIndependent(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		value float64
		want  int
	}{
		{value: math.MaxFloat64, want: maxSSRTValue},
		{value: -math.MaxFloat64, want: minSSRTValue},
		{value: 1e300, want: maxSSRTValue},
		{value: -1e300, want: minSSRTValue},
		{value: math.Inf(1), want: maxSSRTValue},
		{value: math.Inf(-1), want: minSSRTValue},
		{value: math.NaN(), want: 0}, // NaN compares false against everything, so it must be handled explicitly.
		{value: maxSSRTValue, want: maxSSRTValue},
		{value: maxSSRTValue + 1, want: maxSSRTValue},
		{value: minSSRTValue, want: minSSRTValue},
		{value: minSSRTValue - 1, want: minSSRTValue},
		{value: 0, want: 0},
		{value: 4, want: 4},
		{value: 4.7, want: 4}, // Narrowing truncates toward zero, as it always has.
		{value: -4.7, want: -4},
	} {
		c.Equal(tc.want, intFromScriptClampedToSSRT(tc.value), "value %v", tc.value)
	}

	// End to end through the script binding: the extremes must map to the ends of the table on every architecture.
	top := fxp.AsFloat[float64](ssrtToYards(maxSSRTValue))
	bottom := fxp.AsFloat[float64](ssrtToYards(minSSRTValue))
	for _, tc := range []struct {
		expr string
		want float64
	}{
		{expr: "measure.modifierToYards(1e300)", want: top},
		{expr: "measure.modifierToYards(-1e300)", want: bottom},
		{expr: "measure.modifierToYards(Infinity)", want: top},
		{expr: "measure.modifierToYards(-Infinity)", want: bottom},
		{expr: "measure.modifierToYards(NaN)", want: fxp.AsFloat[float64](ssrtToYards(0))},
	} {
		v, err := runScript(0, tc.expr)
		c.NoError(err, "expr %q", tc.expr)
		got, err := fxp.FromString(v)
		c.NoError(err, "expr %q result %q", tc.expr, v)
		c.Equal(fxp.FromFloat(tc.want), got, "expr %q", tc.expr)
	}
}

// intFromScriptClampedToSSRT mirrors what ModifierToYards does: narrow the script's number, then let ssrtToYards clamp
// it to the table. Expressed here so the table above can assert the value the table actually acts on.
func intFromScriptClampedToSSRT(value float64) int {
	return min(max(intFromScript(value), minSSRTValue), maxSSRTValue)
}
