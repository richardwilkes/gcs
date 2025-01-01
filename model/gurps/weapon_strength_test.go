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

func TestWeaponStrength(t *testing.T) {
	for i, s := range []string{
		"",
		"10",
		"125M",
		"100M†",
		"12B",
		"10B†",
		"13R",
		"10R†",
		"10†",
		"10‡",
		"M",
		"B",
		"R",
		"†",
		"‡",
		"12BMR†",
		"12BMR‡",
	} {
		check.Equal(t, s, gurps.ParseWeaponStrength(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"?", ""},
		{"0", ""},
		{"5*", "5†"},
		{"7†[10]", "7†"},
		{"12BMR†‡", "12BMR‡"},
		{"   2 m b R † ", "2BMR†"},
		{"spec", ""},
	}
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponStrength(c.input).String(), "test %d", i)
	}
}
