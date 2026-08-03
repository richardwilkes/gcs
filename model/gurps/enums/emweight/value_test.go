// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emweight_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestFractionExtractionMatchesClassification verifies that every spelling of a multiplier that ValueFromString
// classifies as one is also understood by ExtractFraction. ValueFromString explicitly accepts an uppercase "X" (via
// ToLower) and the Unicode multiplication sign, so those leading characters must be stripped before the fraction is
// parsed; otherwise the fraction comes back as 0/1 and multiplies the equipment's weight by zero.
func TestFractionExtractionMatchesClassification(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input     string
		expected  emweight.Value
		formatted string
	}{
		{"x2", emweight.Multiplier, "x2"},
		{"X2", emweight.Multiplier, "x2"},
		{"×2", emweight.Multiplier, "x2"},
		{"  X2  ", emweight.Multiplier, "x2"},
		{"2x", emweight.Multiplier, "x2"},
		{"2×", emweight.Multiplier, "x2"},
		{"x1/2", emweight.Multiplier, "x1/2"},
		{"X1/2", emweight.Multiplier, "x1/2"},
		{"×1/2", emweight.Multiplier, "x1/2"},
		{"x50%", emweight.PercentageMultiplier, "x50%"},
		{"X50%", emweight.PercentageMultiplier, "x50%"},
		{"×50%", emweight.PercentageMultiplier, "x50%"},
		{"+10%", emweight.PercentageAdder, "+10%"},
		{"-10%", emweight.PercentageAdder, "-10%"},
		{"+5", emweight.Addition, "+5"},
		{"-5", emweight.Addition, "-5"},
	} {
		v := emweight.ValueFromString(one.input)
		c.Equal(one.expected, v, "test %d: %q", i, one.input)
		c.Equal(one.formatted, v.Format(v.ExtractFraction(one.input)), "test %d: %q", i, one.input)
	}
}

// TestMultiplierFractionValue verifies that the extracted fraction actually carries the typed multiplier, since a
// value of 0 would silently zero out the equipment's weight.
func TestMultiplierFractionValue(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{"x2", "X2", "×2", "2x", "2×"} {
		c.Equal(fxp.Two, emweight.Multiplier.ExtractFraction(s).Value(), "test %d: %q", i, s)
	}
	for i, s := range []string{"x50%", "X50%", "×50%"} {
		c.Equal(fxp.Fifty, emweight.PercentageMultiplier.ExtractFraction(s).Value(), "test %d: %q", i, s)
	}
}
