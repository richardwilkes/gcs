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

func TestRealLengthConversion(t *testing.T) {
	c := check.New(t)
	c.Equal(`0.25 in`, paper.Length{Length: 0.25, Units: paper.Inch}.String())
	c.Equal(float32(18), paper.Length{Length: 0.25, Units: paper.Inch}.Pixels())
	c.Equal(`1 in`, paper.Length{Length: 1, Units: paper.Inch}.String())
	c.Equal(float32(72), paper.Length{Length: 1, Units: paper.Inch}.Pixels())
	c.Equal(`15 in`, paper.Length{Length: 15, Units: paper.Inch}.String())
	c.Equal(float32(1080), paper.Length{Length: 15, Units: paper.Inch}.Pixels())
	c.Equal("1 cm", paper.Length{Length: 1, Units: paper.Centimeter}.String())
	c.Equal(float32(28.3464566929), paper.Length{Length: 1, Units: paper.Centimeter}.Pixels())
	c.Equal("1 mm", paper.Length{Length: 1, Units: paper.Millimeter}.String())
	c.Equal(float32(2.8346456693), paper.Length{Length: 1, Units: paper.Millimeter}.Pixels())
}

// TestParseLengthUnitCaseInsensitivity verifies that the unit suffix is matched without regard to case, matching
// fxp.LengthFromString and fxp.WeightFromString. An upper-cased suffix used to be left on the text, so the value failed
// to parse at all rather than being read as inches.
func TestParseLengthUnitCaseInsensitivity(t *testing.T) {
	c := check.New(t)

	for _, text := range []string{"8.5 in", "8.5 IN", "8.5 In", "8.5in", " 8.5 IN ", "+8.5 IN"} {
		length, err := paper.ParseLengthFromString(text)
		c.NoError(err, "%q should parse", text)
		c.Equal(8.5, length.Length, "%q should yield 8.5", text)
		c.Equal(paper.Inch, length.Units, "%q should be inches", text)
	}
	for _, text := range []string{"21 cm", "21 CM", "21 Cm"} {
		length, err := paper.ParseLengthFromString(text)
		c.NoError(err, "%q should parse", text)
		c.Equal(21.0, length.Length, "%q should yield 21", text)
		c.Equal(paper.Centimeter, length.Units, "%q should be centimeters", text)
	}
	for _, text := range []string{"5 mm", "5 MM", "5 Mm"} {
		length, err := paper.ParseLengthFromString(text)
		c.NoError(err, "%q should parse", text)
		c.Equal(5.0, length.Length, "%q should yield 5", text)
		c.Equal(paper.Millimeter, length.Units, "%q should be millimeters", text)
	}

	// No suffix at all still means inches.
	length, err := paper.ParseLengthFromString("2")
	c.NoError(err, "a bare number should parse")
	c.Equal(2.0, length.Length, "a bare number should yield its value")
	c.Equal(paper.Inch, length.Units, "a bare number should be inches")
}

// TestParseLengthRejectsNonFinite verifies that the values strconv.ParseFloat accepts but that aren't real measurements
// are refused. Neither +Inf nor NaN is < 0, so the "value must be zero or greater" check let them through, and a
// non-finite margin then poisoned the page layout arithmetic and survived a save/load round trip.
func TestParseLengthRejectsNonFinite(t *testing.T) {
	c := check.New(t)

	for _, text := range []string{
		"inf", "Inf", "INF", "+inf", "-inf", "infinity", "Infinity", "+Infinity",
		"nan", "NaN", "NAN", "+nan", "-nan",
		"inf in", "+Inf in", "NaN in", "inf cm", "nan mm",
	} {
		length, err := paper.ParseLengthFromString(text)
		c.HasError(err, "%q must be rejected", text)
		c.False(math.IsInf(length.Length, 0) || math.IsNaN(length.Length),
			"%q must not yield a non-finite length", text)

		// The forgiving form must not leak a non-finite value either.
		length = paper.LengthFromString(text)
		c.False(math.IsInf(length.Length, 0) || math.IsNaN(length.Length),
			"LengthFromString(%q) must not yield a non-finite length", text)
		c.False(math.IsInf(float64(length.Pixels()), 0) || math.IsNaN(float64(length.Pixels())),
			"LengthFromString(%q).Pixels() must be finite", text)
	}

	// Negative values are still rejected, and ordinary values still parse.
	_, err := paper.ParseLengthFromString("-1 in")
	c.HasError(err, "a negative value must be rejected")
	length, err := paper.ParseLengthFromString("0.25 in")
	c.NoError(err, "a normal value must still parse")
	c.Equal(0.25, length.Length, "a normal value must keep its value")
}

// TestLengthEnsureValidity verifies that a non-finite length that reached a settings object by some other route is
// corrected rather than passed along, since neither +Inf nor NaN is < 0.
func TestLengthEnsureValidity(t *testing.T) {
	c := check.New(t)

	for _, value := range []float64{math.Inf(1), math.Inf(-1), math.NaN(), -1} {
		length := paper.Length{Length: value, Units: paper.Inch}
		length.EnsureValidity()
		c.Equal(0.0, length.Length, "%v should be corrected to 0", value)
		c.Equal(paper.Inch, length.Units, "%v should keep its units", value)
	}

	// A valid length is left alone.
	length := paper.Length{Length: 0.25, Units: paper.Centimeter}
	length.EnsureValidity()
	c.Equal(0.25, length.Length, "a valid length is unchanged")
	c.Equal(paper.Centimeter, length.Units, "a valid length keeps its units")

	// An unknown unit is still corrected.
	length = paper.Length{Length: 1, Units: paper.Unit(200)}
	length.EnsureValidity()
	c.Equal(paper.Inch, length.Units, "an unknown unit falls back to inches")
}
