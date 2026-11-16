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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestBulk(t *testing.T) {
	parse := func(s string) string { return attack.ParseBulk(s).String() }
	same := []string{
		"",
		"-1",
		"-11/-14",
		"-3*",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"?", ""},
		{"0", ""},
		{"3/-2", "0/-2"},
		{"-2/3", "-2"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestBulkTooltip verifies the folded-stock figures: bulk and accuracy worsen by one, accuracy stopping at zero, any
// recoil above one rises by one and the minimum ST rises by a fifth, rounded up.
func TestBulkTooltip(t *testing.T) {
	c := check.New(t)
	const format = "*: Has a retracting stock. With the stock folded, the weapon's stats change to Bulk %s, Accuracy %s, Recoil %s, and minimum ST %s. Folding or unfolding the stock takes one Ready maneuver."
	for _, tc := range []struct {
		bulk     string
		accuracy string
		recoil   string
		strength string
		want     string
	}{
		{bulk: "-3", accuracy: "4", recoil: "3", strength: "10"},
		{
			bulk:     "-3/-5*",
			accuracy: "4",
			recoil:   "1/2",
			strength: "11",
			want:     fmt.Sprintf(format, "-2/-4*", "3", "1/3", "14"),
		},
		{
			bulk:     "-4*",
			accuracy: "0",
			recoil:   "2/1",
			strength: "10",
			want:     fmt.Sprintf(format, "-3*", "0", "3/1", "12"),
		},
	} {
		c.Equal(tc.want, attack.ParseBulk(tc.bulk).Tooltip(attack.ParseAccuracy(tc.accuracy),
			attack.ParseRecoil(tc.recoil), attack.ParseStrength(tc.strength)), "bulk %q", tc.bulk)
	}
}
