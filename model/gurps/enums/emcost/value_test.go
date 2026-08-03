// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/toolbox/v2/check"
)

const costFactorExample = "+2 CF"

// TestValueExtractionMatchesClassification verifies that every spelling of a multiplier that FromString classifies as
// one is also understood by ExtractValue. FromString lowercases before classifying and accepts the Unicode
// multiplication sign, so "X2" and "×2" must extract 2, not fall through to the non-positive-multiplier coercion that
// turns them into "x1" (or, for "×2", into an addition of 0).
func TestValueExtractionMatchesClassification(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input     string
		expected  emcost.Value
		formatted string
	}{
		{"x2", emcost.Multiplier, "x2"},
		{"X2", emcost.Multiplier, "x2"},
		{"×2", emcost.Multiplier, "x2"},
		{"  X2.5  ", emcost.Multiplier, "x2.5"},
		{"2x", emcost.Multiplier, "x2"},
		{"2X", emcost.Multiplier, "x2"},
		{"2×", emcost.Multiplier, "x2"},
		{"x0", emcost.Multiplier, "x1"},
		{"x", emcost.Multiplier, "x1"},
		{"+3", emcost.Addition, "+3"},
		{"-3", emcost.Addition, "-3"},
		{"", emcost.Addition, "+0"},
		{"+10%", emcost.Percentage, "+10%"},
		{"-10%", emcost.Percentage, "-10%"},
		{costFactorExample, emcost.CostFactor, costFactorExample},
		{"+2 cf", emcost.CostFactor, costFactorExample},
	} {
		v := emcost.Addition.FromString(one.input)
		c.Equal(one.expected, v, "test %d: %q", i, one.input)
		c.Equal(one.formatted, v.Format(v.ExtractValue(one.input)), "test %d: %q", i, one.input)
	}
}
