// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//nolint:goconst // I'd rather have the strings inline than extracted out into a constant for the tests.
package gurps_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestWeaponReach(t *testing.T) {
	c := check.New(t)
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
		"C,2",
		"C,2-3",
		"C,2-3*",
		"C,5",
	} {
		c.Equal(s, gurps.ParseWeaponReach(s).String(), "test %d", i)
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
		{"C, 2 - 3", "C,2-3"},
		{"C,2,3", "C,2-3"},
		{"c,2-3*", "C,2-3*"},
		{"Special", ""},
		{"  1 , 3 ", "1-3"},
	}
	for i, one := range cases {
		c.Equal(one.expected, gurps.ParseWeaponReach(one.input).String(), "test %d", i)
	}
}

// TestWeaponReachRoundTripsThroughJSON verifies that a reach set in the editor survives being saved and reloaded. A
// close combat weapon serializes its marker into the first component of the string ("C,2-3"), where the marker was
// read as the minimum reach and produced 0, which Validate then turned into 1: every save/load cycle silently rewrote
// the weapon's minimum reach.
func TestWeaponReachRoundTripsThroughJSON(t *testing.T) {
	c := check.New(t)
	for i, original := range []gurps.WeaponReach{
		{},
		{Min: fxp.One, Max: fxp.One},
		{Min: fxp.One, Max: fxp.Two},
		{CloseCombat: true},
		{Min: fxp.One, Max: fxp.One, CloseCombat: true},
		{Min: fxp.Two, Max: fxp.Three, CloseCombat: true},
		{Min: fxp.Two, Max: fxp.Three, CloseCombat: true, ChangeRequiresReady: true},
		{Min: fxp.Five, Max: fxp.Five, CloseCombat: true},
	} {
		data, err := jio.Marshal(original)
		c.NoError(err, "test %d", i)
		var reloaded gurps.WeaponReach
		c.NoError(jio.Unmarshal(data, &reloaded), "test %d", i)
		c.Equal(original, reloaded, "test %d (%s)", i, original.String())
	}
}
