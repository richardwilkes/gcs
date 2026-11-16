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

// TestWeightFromString verifies the pounds a text parses to, which pins the conversion of kg, lb, oz, tn and g to
// pounds: a bare number takes the default unit, and a leading plus and any whitespace are ignored.
func TestWeightFromString(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		text     string
		def      fxp.WeightUnit
		expected string
	}{
		{text: "1", def: fxp.Kilogram, expected: "2 lb"},
		{text: " +22.34   lb  ", def: fxp.Pound, expected: "22.34 lb"},
		{text: "1 oz", def: fxp.Pound, expected: "0.0625 lb"},
		{text: "1 tn", def: fxp.Pound, expected: "2,000 lb"},
		{text: "500 g", def: fxp.Pound, expected: "1 lb"},
	} {
		w, err := fxp.WeightFromString(d.text, d.def)
		c.NoError(err, "parsing %q", d.text)
		c.Equal(d.expected, w.String(), "parsing %q", d.text)
	}
}

// TestWeightFromStringUnitSuffixes verifies that every unit's key, when appended to a number in lower, upper or mixed
// case, or in upper case after a space, parses back to that same unit. This guards against the suffix-matching being
// sensitive to the order in which the enum is declared (e.g. "g" must not greedily match the "kg" / "g" overlap, nor
// "t" the "tn" / "t" overlap).
func TestWeightFromStringUnitSuffixes(t *testing.T) {
	c := check.New(t)
	for _, unit := range fxp.WeightUnits {
		key := unit.Key()
		upper := strings.ToUpper(key)
		mixed := strings.ToUpper(key[:1]) + key[1:]
		for _, text := range []string{"3" + key, "3" + upper, "3" + mixed, "3 " + upper} {
			w, err := fxp.WeightFromString(text, fxp.Gram)
			c.NoError(err, "parsing %q", text)
			c.Equal(fxp.WeightFromInteger(3, unit), w, "parsing %q", text)
			c.Equal(unit, fxp.TrailingWeightUnitFromString(text, fxp.Gram), "trailing unit of %q", text)
		}
	}
}
