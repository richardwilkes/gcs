/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/check"
)

func TestWeaponRoFSame(t *testing.T) {
	var w gurps.Weapon
	for i, s := range rofModeSameTests {
		check.Equal(t, s, gurps.ParseWeaponRoF(s).String(&w), "test %d", i)
	}
}

func TestWeaponRoFAdjusted(t *testing.T) {
	var w gurps.Weapon
	for i, c := range rofModeAdjustedTests {
		check.Equal(t, c.expected, gurps.ParseWeaponRoF(c.input).String(&w), "test %d", i)
	}
}

func TestWeaponRoFMultiMode(t *testing.T) {
	var w gurps.Weapon
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
	for i, c := range cases {
		check.Equal(t, c.expected, gurps.ParseWeaponRoF(c.input).String(&w), "test %d", i)
	}
}
