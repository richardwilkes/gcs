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
