// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func parseRoF(s string) string {
	return gurps.ParseWeaponRoF(s).String()
}

// TestWeaponRoF verifies that a single-mode rate of fire parses exactly as a mode does, and accepts "Jet" besides.
func TestWeaponRoF(t *testing.T) {
	checkWeaponFieldParsing(check.New(t), parseRoF, append(slices.Clone(rofModeSameTests), "Jet"),
		rofModeAdjustedTests)
}

func TestWeaponRoFMultiMode(t *testing.T) {
	cases := make([]parseCase, 0, len(rofModeSameTests)*len(rofModeAdjustedTests))
	for _, c1 := range rofModeSameTests {
		if c1 != "" {
			for _, c2 := range rofModeAdjustedTests {
				if c2.expected != "" {
					cases = append(cases, parseCase{
						input:    c1 + "/" + c2.input,
						expected: c1 + "/" + c2.expected,
					})
				}
			}
		}
	}
	cases = append(
		cases,
		parseCase{
			input:    "1/",
			expected: "1",
		},
		parseCase{
			input:    "/1",
			expected: "1",
		},
	)
	checkWeaponFieldParsing(check.New(t), parseRoF, nil, cases)
}
