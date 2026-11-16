// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestShots(t *testing.T) {
	parse := func(s string) string { return attack.ParseShots(s).String() }
	same := []string{
		"",
		"1",
		"1(1)",
		"1(10i)",
		"10+1(3)",
		"10x1s",
		"3x3s(2)",
		"T(1)",
		"T(2i)",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"1(-)", "1"},
		{"1(10", "1(10)"},
		{"1,000(2i)", "1000(2i)"},
		{"2 FP", ""},
		{"2D/24hrs.", ""},
		{"3/day", ""},
		{"30+1 (3)", "30+1(3)"},
		{"500-1,000", "500"},
		{"T(-)", "T"},
		{"T(spec)", "T"},
		{"3(spec)", ""},
		{"n/a", ""},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestShotsValidate verifies that a thrown weapon keeps only its reload time, that a weapon with nothing to fire keeps
// nothing, that negative values become zero, and that parsed shots are validated.
func TestShotsValidate(t *testing.T) {
	c := check.New(t)
	checkValidate(c, []validateCase[attack.Shots]{
		{
			name: "thrown",
			in: attack.Shots{
				Count:      fxp.One,
				InChamber:  fxp.Two,
				Duration:   fxp.Three,
				ReloadTime: fxp.Four,
				Thrown:     true,
			},
			expected: attack.Shots{ReloadTime: fxp.Four, Thrown: true},
		},
		{
			name: "nothing to fire",
			in: attack.Shots{
				Count:      fxp.NegOne,
				InChamber:  fxp.NegOne,
				Duration:   fxp.Three,
				ReloadTime: fxp.Four,
			},
			expected: attack.Shots{},
		},
		{
			name:     "negative duration",
			in:       attack.Shots{Count: fxp.One, Duration: fxp.NegOne, ReloadTime: fxp.Two},
			expected: attack.Shots{Count: fxp.One, ReloadTime: fxp.Two},
		},
		{
			name:     "negative reload time",
			in:       attack.Shots{Count: fxp.One, ReloadTime: fxp.NegOne},
			expected: attack.Shots{Count: fxp.One},
		},
	})
	c.Equal(attack.Shots{}, attack.ParseShots("0(3)"), "parsed")
}
