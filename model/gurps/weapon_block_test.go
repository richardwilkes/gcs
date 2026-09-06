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

func TestWeaponBlock(t *testing.T) {
	parse := func(s string) string { return gurps.ParseWeaponBlock(s).String() }
	same := []string{
		"0",
		"-1",
		"10",
		"No",
	}
	adjusted := []parseCase{
		{"", "No"},
		{"-", "No"},
		{"+0", "0"},
		{"+1", "1"},
	}
	checkWeaponFieldParsing(check.New(t), parse, same, adjusted)
}

// TestWeaponBlockResolve runs the Resolve() regression tests shared by both defenses against the block.
func TestWeaponBlockResolve(t *testing.T) {
	runDefenseResolveTests(t, blockUnderTest)
}
