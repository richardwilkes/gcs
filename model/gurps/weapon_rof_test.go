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
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponRoFSame(t *testing.T) {
	c := check.New(t)
	for i, s := range rofModeSameTests {
		c.Equal(s, gurps.ParseWeaponRoF(s).String(), "test %d", i)
	}
	c.Equal("Jet", gurps.ParseWeaponRoF("Jet").String())
}

func TestWeaponRoFAdjusted(t *testing.T) {
	c := check.New(t)

	for i, one := range rofModeAdjustedTests {
		c.Equal(one.expected, gurps.ParseWeaponRoF(one.input).String(), "test %d", i)
	}
}

func TestWeaponRoFMultiMode(t *testing.T) {
	c := check.New(t)
	cases := make([]rofAdjustedCase, 0, len(rofModeSameTests)*len(rofModeAdjustedTests))
	for _, c1 := range rofModeSameTests {
		if c1 != "" {
			for _, c2 := range rofModeAdjustedTests {
				if c2.expected != "" {
					cases = append(cases, rofAdjustedCase{
						input:    c1 + "/" + c2.input,
						expected: c1 + "/" + c2.expected,
					})
				}
			}
		}
	}
	cases = append(cases,
		rofAdjustedCase{
			input:    "1/",
			expected: "1",
		},
		rofAdjustedCase{
			input:    "/1",
			expected: "1",
		},
	)
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponRoF(one.input).String(), "test %d", i)
	}
}
