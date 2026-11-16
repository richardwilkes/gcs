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

func TestRoF(t *testing.T) {
	parse := func(s string) string { return attack.ParseRoF(s).String() }
	same := []string{
		"",
		"1",
		"10!",
		"1x100",
		"9#",
		"3/9#",
		"Jet",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"0", ""},
		{"1(5)", "1"},
		{"1×7", "1x7"},
		{"2.9", "2x9"},
		{"?", ""},
		{"x100", "1x100"},
		{"1×7/x100", "1x7/1x100"},
		{"1/", "1"},
		{"/1", "1"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestRoFValidate verifies that a jet keeps no modes and that the modes of anything else are validated.
func TestRoFValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.RoF]{
		{
			name: "jet",
			in: attack.RoF{
				Mode1: attack.RoFMode{ShotsPerAttack: fxp.Three},
				Mode2: attack.RoFMode{ShotsPerAttack: fxp.One},
				Jet:   true,
			},
			expected: attack.RoF{Jet: true},
		},
		{
			name: "modes",
			in: attack.RoF{
				Mode1: attack.RoFMode{FullAutoOnly: true},
				Mode2: attack.RoFMode{HighCyclicControlledBursts: true},
			},
			expected: attack.RoF{},
		},
	})
}

// TestRoFTooltip verifies that the modes' notes are labeled by mode only when both modes have notes.
func TestRoFTooltip(t *testing.T) {
	c := check.New(t)
	const (
		fullAutoNote = "!: This weapon can only fire on full automatic. Minimum RoF is 3."
		burstsNote   = "#: This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."
	)
	for _, tc := range []struct {
		rof  string
		want string
	}{
		{rof: "9!/9#", want: "First mode:\n" + fullAutoNote + "\n\nSecond mode:\n" + burstsNote},
		{rof: "9!/3", want: fullAutoNote},
		{rof: "3/9#", want: burstsNote},
	} {
		c.Equal(tc.want, attack.ParseRoF(tc.rof).Tooltip(), "rof %q", tc.rof)
	}
}
