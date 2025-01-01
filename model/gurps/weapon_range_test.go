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

func TestWeaponRange(t *testing.T) {
	for i, s := range []string{
		"",
		"1",
		"1-6",
		"1/10",
		"10",
		"10/100",
		"2-4",
		"x0.2",
		"x0.5/x1",
		"x0.6/x1.2",
		"x10/x15",
		"1,000",
		"1,000/3,000",
		"10/1,000",
	} {
		check.Equal(t, s, gurps.ParseWeaponRange(s).String(false), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"0", ""},
		{"0/5", "5"},
		{"×2", "x2"},
		{"1000", "1,000"},
		{"1,000/3,000 mi.", "1,000/3,000 mi"},
		{"1000/3000", "1,000/3,000"},
		{"10/1000", "10/1,000"},
		{"1/2D 5, Max 10", "5/10"},
		{"1/3 mi.", "1/3 mi"},
		{"10/1", "1"},
		{"10/30 mi.", "10/30 mi"},
		{"5/point", ""},
		{"B550", ""},
		{"C/1", "1"},
		{"PBAoE 2", ""},
		{"ST + Skill/5", ""},
		{"ST + skill/5", ""},
		{"ST/2 + skill/5", ""},
		{"5mi", "5 mi"},
		{"STx2.5", "x2.5"},
		{"Sight", ""},
		{"Within Sight", ""},
		{"max 100", "100"},
		{"spec", ""},
		{"spec.", ""},
		{"x0.5/x0.5", "x0.5"},
	}
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponRange(c.input).String(false), "test %d", i)
	}
}
