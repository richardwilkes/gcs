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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestParsePageSizeUnits verifies that width and height come back in the units the size itself specifies, not always in
// inches.
func TestParsePageSizeUnits(t *testing.T) {
	c := check.New(t)

	width, height, valid := ParsePageSize("letter")
	c.True(valid)
	c.Equal(paper.Length{Length: 8.5, Units: paper.Inch}, width)
	c.Equal(paper.Length{Length: 11, Units: paper.Inch}, height)

	width, height, valid = ParsePageSize("a4")
	c.True(valid)
	c.Equal(paper.Length{Length: 210, Units: paper.Millimeter}, width)
	c.Equal(paper.Length{Length: 297, Units: paper.Millimeter}, height)

	// The "na-" and "iso-" prefixes are ignored, and matching is case-insensitive.
	width, height, valid = ParsePageSize("  ISO-A4 ")
	c.True(valid)
	c.Equal(paper.Length{Length: 210, Units: paper.Millimeter}, width)
	c.Equal(paper.Length{Length: 297, Units: paper.Millimeter}, height)

	// An explicit size carries whatever units were given.
	width, height, valid = ParsePageSize("100mm x 20cm")
	c.True(valid)
	c.Equal(paper.Length{Length: 100, Units: paper.Millimeter}, width)
	c.Equal(paper.Length{Length: 20, Units: paper.Centimeter}, height)

	width, height = MustParsePageSize("a4")
	c.Equal(paper.Length{Length: 210, Units: paper.Millimeter}, width)
	c.Equal(paper.Length{Length: 297, Units: paper.Millimeter}, height)
}

// TestParsePageSizeInvalid verifies that a size that is unknown, malformed or outside the usable range in either
// dimension is refused, and that MustParsePageSize falls back to the first standard paper size for it.
func TestParsePageSizeInvalid(t *testing.T) {
	c := check.New(t)
	for _, one := range []string{"", "bogus", "1x", "x1", "0.01in x 1in", "1000in x 1in", "8in x", "8in x 1000in"} {
		_, _, valid := ParsePageSize(one)
		c.False(valid, one)
		width, height := MustParsePageSize(one)
		c.Equal(StdPaperSizes[0].Width, width, one)
		c.Equal(StdPaperSizes[0].Height, height, one)
	}
}

// TestEnsurePageSizeIsValid verifies that a page size is stored in canonical form: a standard size by its name, whether
// given by name or by dimensions, any other usable size as its two dimensions, and anything else as the first standard
// size.
func TestEnsurePageSizeIsValid(t *testing.T) {
	c := check.New(t)
	for in, want := range map[string]string{
		" NA-Legal ":    "legal",
		"210mm x 297mm": "a4",
		"100mm x 20cm":  "100 mm x 20 cm",
		"bogus":         "letter",
	} {
		c.Equal(want, EnsurePageSizeIsValid(in), in)
	}
	c.Equal("letter", ToPageSize(paper.Length{Length: 1, Units: paper.Inch}, paper.Length{Length: 11, Units: paper.Inch}),
		"dimensions outside the usable range become the first standard size")
}
