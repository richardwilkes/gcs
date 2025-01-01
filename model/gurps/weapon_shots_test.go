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

func TestWeaponShots(t *testing.T) {
	for i, s := range []string{
		"",
		"1",
		"1(1)",
		"1(10)",
		"1(10i)",
		"10",
		"10(3)",
		"10(30i)",
		"10+1(3)",
		"10x1s",
		"1x120s",
		"1x1s",
		"2+1(2i)",
		"3x3s",
	} {
		check.Equal(t, s, gurps.ParseWeaponShots(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"1(-)", "1"},
		{"1(10", "1(10)"},
		{"1,000(2i)", "1000(2i)"},
		{"1,000(3)", "1000(3)"},
		{"2 (5)", "2(5)"},
		{"2 FP", ""},
		{"2D/24hrs.", ""},
		{"3/day", ""},
		{"30+1 (3)", "30+1(3)"},
		{"500-1,000", "500"},
		{"T(1)", "T"},
		{"T(spec)", "T"},
		{"n/a", ""},
	}
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponShots(c.input).String(), "test %d", i)
	}
}
