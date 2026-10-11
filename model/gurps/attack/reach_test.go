// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

//nolint:goconst // I'd rather have the strings inline than extracted out into a constant for the tests.
package attack_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestReach(t *testing.T) {
	parse := func(s string) string { return attack.ParseReach(s).String() }
	same := []string{
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
	}
	adjusted := []parseCase{
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
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestReachRoundTripsThroughJSON guards against the close combat marker in the first component ("C,2-3") being
// read as a minimum reach of 0, which Validate turned into 1, silently rewriting the minimum on every save/load.
func TestReachRoundTripsThroughJSON(t *testing.T) {
	c := check.New(t)
	for i, original := range []attack.Reach{
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
		var reloaded attack.Reach
		c.NoError(jio.Unmarshal(data, &reloaded), "test %d", i)
		c.Equal(original, reloaded, "test %d (%s)", i, original.String())
	}
}
