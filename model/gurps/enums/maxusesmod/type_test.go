// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package maxusesmod_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/maxusesmod"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestNormalize verifies that extraction understands every spelling of a multiplier that FromString classifies as one.
// FromString lowercases before matching, so "X2" must normalize to "x2" rather than losing the typed value to the
// non-positive-multiplier coercion and becoming "x1".
func TestNormalize(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input    string
		expected string
	}{
		{"x2", "x2"},
		{"X2", "x2"},
		{"×2", "x2"},
		{"  X2  ", "x2"},
		{"2x", "x2"},
		{"2X", "x2"},
		{"2×", "x2"},
		{"x0", "x1"},
		{"x-2", "x1"},
		{"x", "x1"},
		{"+3", "+3"},
		{"-3", "-3"},
		{"", "+0"},
		{"+10%", "+10%"},
		{"-10%", "-10%"},
	} {
		c.Equal(one.expected, maxusesmod.Normalize(one.input), "test %d: %q", i, one.input)
	}
}

// TestFromString verifies the classification of each supported spelling.
func TestFromString(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input    string
		expected maxusesmod.Type
	}{
		{"x2", maxusesmod.Multiplier},
		{"X2", maxusesmod.Multiplier},
		{"×2", maxusesmod.Multiplier},
		{"2x", maxusesmod.Multiplier},
		{"2×", maxusesmod.Multiplier},
		{"+3", maxusesmod.Addition},
		{"+10%", maxusesmod.Percentage},
	} {
		c.Equal(one.expected, maxusesmod.FromString(one.input), "test %d: %q", i, one.input)
	}
}
