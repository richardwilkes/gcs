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

func TestWeaponReach(t *testing.T) {
	for i, s := range []string{
		"",
		"1",
		"1*",
		"1-2",
		"3-4",
		"1-2*",
		"C",
		"C,1",
		"C,1-2",
	} {
		check.Equal(t, s, gurps.ParseWeaponReach(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"1,2", "1-2"},
		{"1,2*", "1-2*"},
		{"1, 2", "1-2"},
		{"1, 2*", "1-2*"},
		{"1/point", "1"},
		{"5/10", "5"},
		{"C, 1", "C,1"},
		{"C,1,2", "C,1-2"},
		{"C-5", "C,1-5"},
		{"Special", ""},
		{"  1 , 3 ", "1-3"},
	}
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponReach(c.input).String(), "test %d", i)
	}
}
