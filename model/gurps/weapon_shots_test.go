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

func TestWeaponShots(t *testing.T) {
	c := check.New(t)
	for i, s := range []string{
		"",
		"1",
		"1(1)",
		"1(10)",
		"1(10i)",
		"10",
		"10(3)",
		"10(30i)",
		"10+1(3)",
		"10x1s",
		"1x120s",
		"1x1s",
		"2+1(2i)",
		"3x3s",
		"3x3s(2)",
		"T(1)",
		"T(2i)",
	} {
		c.Equal(s, gurps.ParseWeaponShots(s).String(), "test %d", i)
	}

	cases := []struct {
		input    string
		expected string
	}{
		{"-", ""},
		{"–", ""},
		{"—", ""},
		{"1(-)", "1"},
		{"1(10", "1(10)"},
		{"1,000(2i)", "1000(2i)"},
		{"1,000(3)", "1000(3)"},
		{"2 (5)", "2(5)"},
		{"2 FP", ""},
		{"2D/24hrs.", ""},
		{"3/day", ""},
		{"30+1 (3)", "30+1(3)"},
		{"500-1,000", "500"},
		{"T (2)", "T(2)"},
		{"T(-)", "T"},
		{"T(spec)", "T"},
		{"n/a", ""},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponShots(one.input).String(), "test %d", i)
	}
}

// TestWeaponShotsReloadTimeParsing verifies that the reload time is recovered even when something separates it from
// the shot count, i.e. the thrown marker or the duration's trailing "s". Both are emitted by String(), so failing to
// parse them back loses the reload time on every save/load round trip.
func TestWeaponShotsReloadTimeParsing(t *testing.T) {
	c := check.New(t)

	ws := gurps.ParseWeaponShots("T(2)")
	c.True(ws.Thrown)
	c.Equal(fxp.Two, ws.ReloadTime)
	c.False(ws.ReloadTimeIsPerShot)

	ws = gurps.ParseWeaponShots("T(2i)")
	c.True(ws.Thrown)
	c.Equal(fxp.Two, ws.ReloadTime)
	c.True(ws.ReloadTimeIsPerShot)

	ws = gurps.ParseWeaponShots("3x3s(2)")
	c.False(ws.Thrown)
	c.Equal(fxp.Three, ws.Count)
	c.Equal(fxp.Three, ws.Duration)
	c.Equal(fxp.Two, ws.ReloadTime)
}
