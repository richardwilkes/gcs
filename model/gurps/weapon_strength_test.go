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

func TestWeaponStrength(t *testing.T) {
	parse := func(s string) string { return gurps.ParseWeaponStrength(s).String() }
	same := []string{
		"",
		"10",
		"125M",
		"100M†",
		"12B",
		"10B†",
		"13R",
		"10R†",
		"10†",
		"10‡",
		"M",
		"B",
		"R",
		"†",
		"‡",
		"12BMR†",
		"12BMR‡",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"–", ""},
		{"—", ""},
		{"?", ""},
		{"0", ""},
		{"5*", "5†"},
		{"7†[10]", "7†"},
		{"12BMR†‡", "12BMR‡"},
		{"   2 m b R † ", "2BMR†"},
		{"spec", ""},
	}
	checkWeaponFieldParsing(check.New(t), parse, same, adjusted)
}
