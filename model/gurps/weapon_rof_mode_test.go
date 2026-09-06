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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

var rofModeSameTests = []string{
	"",
	"1",
	"10!",
	"10",
	"1x100",
	"2x12",
	"9#",
}

var rofModeAdjustedTests = []parseCase{
	{"-", ""},
	{"0", ""},
	{"1(5)", "1"},
	{"1×7", "1x7"},
	{"2.9", "2x9"},
	{"?", ""},
	{"x100", "1x100"},
}

func TestWeaponRoFMode(t *testing.T) {
	parse := func(s string) string { return gurps.ParseWeaponRoFMode(s).String() }
	checkWeaponFieldParsing(check.New(t), parse, rofModeSameTests, rofModeAdjustedTests)
}
