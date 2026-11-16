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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

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

// TestLengthExtremesRoundTrip verifies that the extreme lengths render as feet and inches rather than as a bare "-",
// and parse back to the original value: Min has no positive counterpart, so taking its magnitude before splitting off
// the feet used to leave it negative and every part was then skipped as "not positive". LengthFromString saturates
// absurd input to Min, so a hand-edited file could reach it and then be re-saved with a height of "-".
func TestLengthExtremesRoundTrip(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		expected string
		length   fxp.Length
	}{
		{length: fxp.Length(fxp.Min), expected: `-76,861,433,640,456'5.5808"`},
		{length: fxp.Length(fxp.Min + 1), expected: `-76,861,433,640,456'5.5807"`},
		{length: fxp.Length(fxp.Max), expected: `76,861,433,640,456'5.5807"`},
	} {
		c.Equal(d.expected, d.length.String())
		parsed, err := fxp.LengthFromString(d.expected, fxp.Inch)
		c.NoError(err, "parsing %q", d.expected)
		c.Equal(d.length, parsed, "parsing %q", d.expected)
	}
}

// TestLengthFromStringUnitSuffixes verifies that every unit's key (other than FeetAndInches, which is parsed via the
// '/" notation), when appended to a number in lower case, in upper case, or in upper case after a space, parses back to
// that same unit. This guards against the suffix-matching being sensitive to the order in which the enum is declared
// (e.g. "m" must not greedily match the "cm" / "km" / "m" overlap).
func TestLengthFromStringUnitSuffixes(t *testing.T) {
	c := check.New(t)
	for _, unit := range fxp.LengthUnits {
		if unit == fxp.FeetAndInches {
			continue
		}
		upper := strings.ToUpper(unit.Key())
		for _, text := range []string{"3" + unit.Key(), "3" + upper, "3 " + upper} {
			w, err := fxp.LengthFromString(text, fxp.Meter)
			c.NoError(err, "parsing %q", text)
			c.Equal(fxp.LengthFromInteger(3, unit), w, "parsing %q", text)
		}
	}
}

// TestLengthFromString verifies the inches a text parses to, which pins each unit's conversion to inches: a bare number
// takes the default unit, a leading plus and any whitespace are ignored, and the inches portion of a feet-and-inches
// value is honored even when the closing inch marker hasn't been typed yet. The length fields validate on every
// keystroke and reject the rune when parsing fails, so `6'` and `6'2` must be valid intermediate states -- and the
// latter must mean 6'2", not 6' with the inches silently discarded. The table also holds the strings the formatter
// produces for negative lengths, where the sign applies to the feet and inches combined. LengthFromStringForced
// returns the same value, or zero for text that does not parse.
func TestLengthFromString(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		text     string
		def      fxp.LengthUnit
		expected string
	}{
		{text: "2.5", def: fxp.Centimeter, expected: "1"},
		{text: " +32   yd  ", def: fxp.Inch, expected: "1152"},
		{text: "0.5m", def: fxp.Inch, expected: "20"},
		{text: "1ft", def: fxp.Inch, expected: "12"},
		{text: "1mi", def: fxp.Inch, expected: "63360"},
		{text: "1km", def: fxp.Inch, expected: "40000"},
		{text: `6'2`, def: fxp.Inch, expected: "74"},
		{text: `6'2"`, def: fxp.Inch, expected: "74"},
		{text: `6'  2`, def: fxp.Inch, expected: "74"},
		{text: `6'`, def: fxp.Inch, expected: "72"},
		{text: `12"`, def: fxp.Inch, expected: "12"},
		{text: `-2'6`, def: fxp.Inch, expected: "-30"},
		{text: `-6"`, def: fxp.Inch, expected: "-6"},
		{text: `-1'`, def: fxp.Inch, expected: "-12"},
		{text: `-2'6"`, def: fxp.Inch, expected: "-30"},
	} {
		w, err := fxp.LengthFromString(d.text, d.def)
		c.NoError(err, "parsing %q", d.text)
		c.Equal(d.expected, fxp.Int(w).String(), "parsing %q", d.text)
	}
	c.Equal(fxp.LengthFromInteger(1, fxp.Inch), fxp.LengthFromStringForced("2.5", fxp.Centimeter))
	c.Equal(fxp.Length(0), fxp.LengthFromStringForced("junk", fxp.Inch))
}

// TestLengthFromStringRejectsTrailingText verifies that the feet-and-inches parser consumes the entire input. Trailing
// text used to be silently discarded (e.g. `6'2"junk` parsed as 6'2" and `12"3` as 12"), storing a value the user never
// typed instead of reporting the input as malformed.
func TestLengthFromStringRejectsTrailingText(t *testing.T) {
	c := check.New(t)
	for _, text := range []string{`6'2"junk`, `12"3`, `6'2"3`, `6'2""`, `6'x`, `6'2'3`, `12"3'`, `"`} {
		_, err := fxp.LengthFromString(text, fxp.Inch)
		c.HasError(err, "parsing %q should fail", text)
	}
}
