// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper_test

import (
	"math"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestLengthStringAndPixels pins the text form, which is also the form saved to files, and each unit's conversion to
// 72-per-inch pixels.
func TestLengthStringAndPixels(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		expected string
		length   paper.Length
		pixels   float32
	}{
		{length: paper.Length{Length: 0.25, Units: paper.Inch}, expected: "0.25 in", pixels: 18},
		{length: paper.Length{Length: 1, Units: paper.Inch}, expected: "1 in", pixels: 72},
		{length: paper.Length{Length: 1, Units: paper.Centimeter}, expected: "1 cm", pixels: 28.3464566929},
		{length: paper.Length{Length: 1, Units: paper.Millimeter}, expected: "1 mm", pixels: 2.8346456693},
	} {
		c.Equal(d.expected, d.length.String())
		c.Equal(d.pixels, d.length.Pixels(), "pixels of %s", d.expected)
	}
}

// TestParseLengthFromString pins the accepted spellings and the rejected values. strconv.ParseFloat accepts "inf",
// "infinity" and "nan", none of which is < 0, so the "zero or greater" check alone would let them through, and it
// returns ±Inf alongside its error on overflow.
func TestParseLengthFromString(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		text     string
		expected paper.Length
		rejected bool
	}{
		{text: "8.5 in", expected: paper.Length{Length: 8.5, Units: paper.Inch}},
		{text: "8.5in", expected: paper.Length{Length: 8.5, Units: paper.Inch}},
		{text: " 8.5 IN ", expected: paper.Length{Length: 8.5, Units: paper.Inch}},
		{text: "+8.5 in", expected: paper.Length{Length: 8.5, Units: paper.Inch}},
		{text: "21 Cm", expected: paper.Length{Length: 21, Units: paper.Centimeter}},
		{text: "5 MM", expected: paper.Length{Length: 5, Units: paper.Millimeter}},
		{text: "2", expected: paper.Length{Length: 2, Units: paper.Inch}},
		{text: "0", expected: paper.Length{Units: paper.Inch}},
		{text: "abc", rejected: true},
		{text: "-1 in", rejected: true},
		{text: "inf", rejected: true},
		{text: "-inf", rejected: true},
		{text: "infinity", rejected: true},
		{text: "nan", rejected: true},
		{text: "inf cm", rejected: true},
		{text: "1e999", rejected: true},
	} {
		length, err := paper.ParseLengthFromString(d.text)
		if d.rejected {
			c.HasError(err, "%q must be rejected", d.text)
			c.False(math.IsInf(length.Length, 0) || math.IsNaN(length.Length),
				"%q must not yield a non-finite length", d.text)
			continue
		}
		c.NoError(err, "%q should parse", d.text)
		c.Equal(d.expected, length, "parsing %q", d.text)
	}
}

// TestLengthEnsureValidity verifies that a negative or non-finite length is reset to 0 and an unknown unit to inches,
// each without disturbing the other field. Neither +Inf nor NaN is < 0.
func TestLengthEnsureValidity(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		value    float64
		units    paper.Unit
		expected paper.Length
	}{
		{value: math.Inf(1), units: paper.Centimeter, expected: paper.Length{Units: paper.Centimeter}},
		{value: math.Inf(-1), units: paper.Centimeter, expected: paper.Length{Units: paper.Centimeter}},
		{value: math.NaN(), units: paper.Centimeter, expected: paper.Length{Units: paper.Centimeter}},
		{value: -1, units: paper.Centimeter, expected: paper.Length{Units: paper.Centimeter}},
		{value: 0.25, units: paper.Centimeter, expected: paper.Length{Length: 0.25, Units: paper.Centimeter}},
		{value: 1, units: paper.Unit(200), expected: paper.Length{Length: 1, Units: paper.Inch}},
	} {
		length := paper.Length{Length: d.value, Units: d.units}
		length.EnsureValidity()
		c.Equal(d.expected, length, "validating %v with unit %d", d.value, d.units)
	}
}
