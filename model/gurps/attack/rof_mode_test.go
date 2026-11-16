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

// TestRoFModeValidate verifies that shot and projectile counts are rounded up, that a mode without shots keeps nothing
// else, and that a parsed mode is validated.
func TestRoFModeValidate(t *testing.T) {
	c := check.New(t)
	checkValidate(c, []validateCase[attack.RoFMode]{
		{
			name:     "fractional",
			in:       attack.RoFMode{ShotsPerAttack: fxp.TwoAndAHalf, SecondaryProjectiles: fxp.OneAndAHalf},
			expected: attack.RoFMode{ShotsPerAttack: fxp.Three, SecondaryProjectiles: fxp.Two},
		},
		{
			name: "no shots",
			in: attack.RoFMode{
				ShotsPerAttack:             fxp.NegOne,
				SecondaryProjectiles:       fxp.Two,
				FullAutoOnly:               true,
				HighCyclicControlledBursts: true,
			},
			expected: attack.RoFMode{},
		},
		{
			name:     "negative projectiles",
			in:       attack.RoFMode{ShotsPerAttack: fxp.One, SecondaryProjectiles: fxp.NegOne},
			expected: attack.RoFMode{ShotsPerAttack: fxp.One},
		},
	})
	c.Equal(attack.RoFMode{}, attack.ParseRoFMode("-3!"), "parsed")
}

// TestRoFModeTooltip verifies the singular and plural forms of the projectile note and that the minimum RoF of a full
// automatic weapon is a quarter of its shots, rounded up.
func TestRoFModeTooltip(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		mode string
		want string
	}{
		{mode: "3x9", want: "This weapon fires 3 shots per attack and each shot releases 9 smaller projectiles."},
		{mode: "1x1", want: "This weapon fires 1 shot per attack and each shot releases 1 smaller projectile."},
		{mode: "9!", want: "!: This weapon can only fire on full automatic. Minimum RoF is 3."},
		{mode: "9#", want: "#: This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1."},
		{
			mode: "2x3!#",
			want: "This weapon fires 2 shots per attack and each shot releases 3 smaller projectiles.\n" +
				"!: This weapon can only fire on full automatic. Minimum RoF is 1.\n" +
				"#: This weapon can fire in high cyclic controlled bursts, reducing Recoil to 1.",
		},
	} {
		c.Equal(tc.want, attack.ParseRoFMode(tc.mode).Tooltip(), "mode %q", tc.mode)
	}
}
