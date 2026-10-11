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

	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestBulk(t *testing.T) {
	parse := func(s string) string { return attack.ParseBulk(s).String() }
	same := []string{
		"",
		"-1",
		"-10",
		"-11/-14",
		"-3*",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"–", ""},
		{"—", ""},
		{"?", ""},
		{"0", ""},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestBulkTooltip verifies the folded-stock figures: bulk and accuracy worsen by one, any recoil above one rises by one
// and the minimum ST rises by a fifth, rounded up.
func TestBulkTooltip(t *testing.T) {
	c := check.New(t)
	accuracy := attack.ParseAccuracy("4")
	recoil := attack.ParseRecoil("3")
	strength := attack.ParseStrength("10")
	c.Equal("", attack.ParseBulk("-3").Tooltip(accuracy, recoil, strength))
	c.Equal("*: Has a retracting stock. With the stock folded, the weapon's stats change to Bulk -2*, Accuracy 3, Recoil 4, and minimum ST 12. Folding or unfolding the stock takes one Ready maneuver.",
		attack.ParseBulk("-3*").Tooltip(accuracy, recoil, strength))
}
