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

// TestModifierClassification verifies that every spelling of a multiplier and percentage marker is recognized,
// regardless of case and surrounding whitespace, and that plain additions match neither.
func TestModifierClassification(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input   string
		prefix  bool
		marker  bool
		percent bool
	}{
		{input: "x2", prefix: true, marker: true},
		{input: "X2", prefix: true, marker: true},
		{input: "×2", prefix: true, marker: true},
		{input: "  x2  ", prefix: true, marker: true},
		{input: "2x", marker: true},
		{input: "2X", marker: true},
		{input: "2×", marker: true},
		{input: "  2×  ", marker: true},
		{input: "x", prefix: true, marker: true},
		{input: "x50%", prefix: true, marker: true, percent: true},
		{input: "×50%", prefix: true, marker: true, percent: true},
		{input: "+10%", percent: true},
		{input: " -10% ", percent: true},
		{input: "+2"},
		{input: "-2"},
		{input: "+2 CF"},
		{input: "+5 lb"},
		{input: ""},
		{input: "   "},
	} {
		c.Equal(one.prefix, fxp.HasMultiplierPrefix(one.input), "prefix %d: %q", i, one.input)
		c.Equal(one.marker, fxp.HasMultiplierMarker(one.input), "marker %d: %q", i, one.input)
		c.Equal(one.percent, fxp.HasPercentSuffix(one.input), "percent %d: %q", i, one.input)
	}
}

// TestStripMultiplierLeaders verifies that every leading multiplier marker is removed along with surrounding
// whitespace, while trailing markers and the numeric portion are left alone.
func TestStripMultiplierLeaders(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input    string
		expected string
	}{
		{"x2", "2"},
		{"X2", "2"},
		{"×2", "2"},
		{"  ×2.5  ", "2.5"},
		{"xx2", "2"},
		{"2x", "2x"},
		{"x1/2", "1/2"},
		{"x50%", "50%"},
		{"+3", "+3"},
		{"", ""},
	} {
		c.Equal(one.expected, fxp.StripMultiplierLeaders(one.input), "test %d: %q", i, one.input)
	}
}

// TestExtractModifierValue verifies that the numeric portion of every spelling is extracted, and that a non-positive
// multiplier is coerced to 1 only when the value is being read as a multiplier.
func TestExtractModifierValue(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		input      string
		multiplier bool
		expected   fxp.Int
	}{
		{"x2", true, fxp.Two},
		{"X2", true, fxp.Two},
		{"×2", true, fxp.Two},
		{"2x", true, fxp.Two},
		{"2×", true, fxp.Two},
		{"  X2.5  ", true, fxp.TwoAndAHalf},
		{"x0", true, fxp.One},
		{"x-2", true, fxp.One},
		{"x", true, fxp.One},
		{"", true, fxp.One},
		{"+3", false, fxp.Three},
		{"-3", false, -fxp.Three},
		{"+10%", false, fxp.Ten},
		{"-10%", false, -fxp.Ten},
		{"+2 CF", false, fxp.Two},
		{"0", false, 0},
		{"", false, 0},
	} {
		c.Equal(one.expected, fxp.ExtractModifierValue(one.input, one.multiplier), "test %d: %q", i, one.input)
	}
}
