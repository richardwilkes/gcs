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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponRecoil(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"",
		"1",
		"1/3",
		"1/4",
		"1/5",
		"1/6",
		"1/7",
		"10",
		"2",
		"3",
		"4",
		"5",
		"6",
		"7",
		"8",
	} {
		c.Equal(s, gurps.ParseWeaponRecoil(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"—", ""},
		{"?", ""},
		{"0", ""},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponRecoil(one.input).String(), "test %d", i)
	}
}

// TestWeaponPercentBonusFloorsIncrement pins the documented rounding of weapon percentage bonuses: the increment is
// floored, i.e. rounded toward negative infinity, not toward zero. A -50% bonus on a recoil of 5 yields an increment
// of -2.5, which floors to -3, for a result of 2.
func TestWeaponPercentBonusFloorsIncrement(t *testing.T) {
	c := check.New(t)

	bonus := gurps.NewWeaponRecoilBonus()
	bonus.Percent = true
	bonus.Amount = fxp.FromInteger(-50)
	w := newWeaponWithBonuses(false, bonus)

	w.Recoil = gurps.ParseWeaponRecoil("5")
	c.Equal("2", w.Recoil.Resolve(w, nil).String())

	w.Recoil = gurps.ParseWeaponRecoil("5/7")
	c.Equal("2/3", w.Recoil.Resolve(w, nil).String())
}
