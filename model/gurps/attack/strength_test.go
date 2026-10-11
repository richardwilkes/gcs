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

func TestStrength(t *testing.T) {
	parse := func(s string) string { return attack.ParseStrength(s).String() }
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
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestStrengthTooltip verifies that the rated ST note appears only when there is a rated ST and that the minimum ST
// note follows it.
func TestStrengthTooltip(t *testing.T) {
	c := check.New(t)
	c.Equal("", attack.ParseStrength("").Tooltip(0))
	rated := attack.ParseStrength("").Tooltip(fxp.FromInteger(12))
	c.HasPrefix(rated, "The weapon has a rated ST of 12")
	c.NotContains(rated, "minimum ST of")
	both := attack.ParseStrength("10").Tooltip(fxp.FromInteger(12))
	c.HasPrefix(both, "The weapon has a rated ST of 12")
	c.Contains(both, "\n\nThe weapon has a minimum ST of 10")
}
