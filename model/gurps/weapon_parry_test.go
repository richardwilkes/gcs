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

func TestWeaponParry(t *testing.T) {
	parse := func(s string) string { return gurps.ParseWeaponParry(s).String() }
	same := []string{
		"0",
		"-1",
		"10",
		"0U",
		"-1U",
		"9U",
		"0F",
		"-2F",
		"8F",
		"0FU",
		"-2FU",
		"8FU",
		"No",
	}
	adjusted := []parseCase{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
		{"0 (x5)", "0"},
		{"0U / 0", "0U"},
		{"0U/ 0", "0U"},
		{"13 (x5)", "13"},
	}
	checkWeaponFieldParsing(check.New(t), parse, same, adjusted)
}

// TestWeaponParryResolve runs the Resolve() regression tests shared by both defenses against the parry.
func TestWeaponParryResolve(t *testing.T) {
	runDefenseResolveTests(t, parryUnderTest)
}
