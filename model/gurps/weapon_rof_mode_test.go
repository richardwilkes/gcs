// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/check"
)

var rofModeSameTests = []string{
	"",
	"1",
	"10!",
	"10",
	"1x100",
	"2x12",
	"9#",
}

var rofModeAdjustedTests = []rofAdjustedCase{
	{"-", ""},
	{"0", ""},
	{"1(5)", "1"},
	{"1×7", "1x7"},
	{"2.9", "2x9"},
	{"?", ""},
	{"x100", "1x100"},
}

type rofAdjustedCase struct {
	input    string
	expected string
}

func TestWeaponRoFModeSame(t *testing.T) {
	for i, s := range rofModeSameTests {
		check.Equal(t, s, gurps.ParseWeaponRoFMode(s).String(), "test %d", i)
	}
}

func TestWeaponRoFModeAdjusted(t *testing.T) {
	for i, c := range rofModeAdjustedTests {
		check.Equal(t, c.expected, gurps.ParseWeaponRoFMode(c.input).String(), "test %d", i)
	}
}
