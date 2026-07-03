// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestGURPSLengthConversion(t *testing.T) {
	c := check.New(t)
	c.Equal(`1"`, fxp.FeetAndInches.Format(fxp.LengthFromInteger(1, fxp.Inch)))
	c.Equal(`1'3"`, fxp.FeetAndInches.Format(fxp.LengthFromInteger(15, fxp.Inch)))
	c.Equal("2.5 cm", fxp.Centimeter.Format(fxp.LengthFromStringForced("2.5", fxp.Centimeter)))
	c.Equal("37.5 cm", fxp.Centimeter.Format(fxp.LengthFromStringForced("37.5", fxp.Centimeter)))

	w, err := fxp.LengthFromString("1", fxp.Inch)
	c.NoError(err)
	c.Equal(`1"`, fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString(`6'         2"`, fxp.Inch)
	c.NoError(err)
	c.Equal(`6'2"`, fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString(" +32   yd  ", fxp.Inch)
	c.NoError(err)
	c.Equal("96'", fxp.FeetAndInches.Format(w))
	w, err = fxp.LengthFromString("0.5m", fxp.Inch)
	c.NoError(err)
	c.Equal("50 cm", fxp.Centimeter.Format(w))
	w, err = fxp.LengthFromString("1cm", fxp.Inch)
	c.NoError(err)
	c.Equal("1 cm", fxp.Centimeter.Format(w))
}

// TestFeetAndInchesFormat verifies that the FeetAndInches formatter handles zero, positive, and negative lengths. It
// guards against the negative-length regression where floor-toward-negative-infinity dropped the sign and mangled the
// remainder (e.g. -30 inches formatting as `6"` and -12 inches formatting as an empty string).
func TestFeetAndInchesFormat(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		inches   int
		expected string
	}{
		{inches: 0, expected: "0'"},
		{inches: 1, expected: `1"`},
		{inches: 12, expected: "1'"},
		{inches: 15, expected: `1'3"`},
		{inches: 30, expected: `2'6"`},
		{inches: -1, expected: `-1"`},
		{inches: -12, expected: "-1'"},
		{inches: -15, expected: `-1'3"`},
		{inches: -30, expected: `-2'6"`},
	} {
		c.Equal(d.expected, fxp.FeetAndInches.Format(fxp.LengthFromInteger(d.inches, fxp.Inch)),
			"formatting %d inches", d.inches)
		// Length.String() routes through the same formatter, so it must agree.
		c.Equal(d.expected, fxp.LengthFromInteger(d.inches, fxp.Inch).String(), "String() of %d inches", d.inches)
	}
}

// TestLengthFeetAndInchesRoundTrip verifies that formatting a length and parsing the result back yields the original
// value. It guards against the regression where the parser applied the sign to the feet contribution alone while the
// formatter emitted both magnitudes unsigned, corrupting any value <= -12 inches (e.g. -30 inches -> `-2'6"` ->
// -2*12+6 = -18 inches).
func TestLengthFeetAndInchesRoundTrip(t *testing.T) {
	c := check.New(t)
	for _, inches := range []int{0, 1, 6, 12, 15, 24, 30, -1, -6, -12, -15, -24, -30, -144} {
		original := fxp.LengthFromInteger(inches, fxp.Inch)
		text := original.String()
		parsed, err := fxp.LengthFromString(text, fxp.Inch)
		c.NoError(err, "parsing %q", text)
		c.Equal(original, parsed, "round-trip of %d inches (formatted as %q)", inches, text)
	}
}

// TestLengthFromStringNegativeFeetAndInches verifies the exact feet-and-inches strings the formatter produces for
// negative lengths parse back to the correct negative inch counts.
func TestLengthFromStringNegativeFeetAndInches(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		text     string
		expected int
	}{
		{text: `-6"`, expected: -6},
		{text: `-1'`, expected: -12},
		{text: `-1'3"`, expected: -15},
		{text: `-2'`, expected: -24},
		{text: `-2'6"`, expected: -30},
	} {
		w, err := fxp.LengthFromString(d.text, fxp.Inch)
		c.NoError(err, "parsing %q", d.text)
		c.Equal(fxp.LengthFromInteger(d.expected, fxp.Inch), w, "parsing %q", d.text)
	}
}

// TestLengthFromStringUnitSuffixes verifies that every unit's key (other than FeetAndInches, which is parsed via the
// '/" notation), when appended to a number, parses back to that same unit. This guards against the suffix-matching
// being sensitive to the order in which the enum is declared (e.g. "m" must not greedily match the "cm" / "km" / "m"
// overlap).
func TestLengthFromStringUnitSuffixes(t *testing.T) {
	c := check.New(t)
	for _, unit := range fxp.LengthUnits {
		if unit == fxp.FeetAndInches {
			continue
		}
		text := "3" + unit.Key()
		w, err := fxp.LengthFromString(text, fxp.Meter)
		c.NoError(err, "parsing %q", text)
		c.Equal(fxp.LengthFromInteger(3, unit), w, "parsing %q", text)
	}
}

// TestLengthFromStringCaseInsensitive verifies that unit suffixes are matched without regard to case.
func TestLengthFromStringCaseInsensitive(t *testing.T) {
	c := check.New(t)
	for _, text := range []string{"0.5M", "0.5 M"} {
		w, err := fxp.LengthFromString(text, fxp.Inch)
		c.NoError(err, "parsing %q", text)
		c.Equal("50 cm", fxp.Centimeter.Format(w), "parsing %q", text)
	}
	w, err := fxp.LengthFromString("32 YD", fxp.Inch)
	c.NoError(err)
	c.Equal("96'", fxp.FeetAndInches.Format(w))
}
