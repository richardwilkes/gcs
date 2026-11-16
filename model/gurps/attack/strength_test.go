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
	"strings"
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
		"12B",
		"13R",
		"10†",
		"10‡",
		"M",
		"12BMR†",
		"12BMR‡",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"?", ""},
		{"0", ""},
		{"5*", "5†"},
		{"7†[10]", "7†"},
		{"12BMR†‡", "12BMR‡"},
		{"   2 m b R † ", "2BMR†"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestStrengthValidate verifies that a parsed strength is validated: a negative minimum becomes zero and ‡ replaces †.
func TestStrengthValidate(t *testing.T) {
	check.New(t).Equal(attack.Strength{TwoHandedUnready: true}, attack.ParseStrength("-5†‡"))
}

// TestStrengthTooltip verifies which notes appear, in what order, and the ST each derives from the minimum: two thirds
// of it with a bipod, and one and a half, two or three times it for the two-handed thresholds, rounded up.
func TestStrengthTooltip(t *testing.T) {
	c := check.New(t)
	const (
		ratedNote        = "The weapon has a rated ST of 12, which is used instead of the user's ST for all calculations except minimum ST."
		minimumNote      = "The weapon has a minimum ST of %d. If your ST is less than this, you will suffer a -1 to weapon skill per point of ST you lack and lose one extra FP at the end of any fight that lasts long enough to cost FP."
		bipodNote        = "B: Has an attached bipod. When used from a prone position, treats the attack as braced (add +1 to Accuracy)."
		bipodReducedNote = "B: Has an attached bipod. When used from a prone position, reduces the ST requirement to %d and treats the attack as braced (add +1 to Accuracy)."
		mountedNote      = "M: Mounted. Ignore listed ST and Bulk when firing from its mount. Takes at least 3 Ready maneuvers to unmount or remount the weapon."
		musketRestNote   = "R: Uses a Musket Rest. Any aimed shot fired while stationary and standing up is automatically braced (add +1 to Accuracy)."
		twoHandedNote    = "†: Requires two hands. If you have at least ST %d, you can use it one-handed, but it becomes unready after you attack with it. If you have at least ST %d, you can use it one-handed with no readiness penalty."
		unreadyNote      = "‡: Requires two hands and becomes unready after you attack with it. If you have at least ST %d, you can use it two-handed without it becoming unready. If you have at least ST %d, you can use it one-handed with no readiness penalty."
	)
	for _, tc := range []struct {
		strength string
		ratedST  fxp.Int
		want     []string
	}{
		{},
		{ratedST: fxp.Twelve, want: []string{ratedNote}},
		{strength: "10", ratedST: fxp.Twelve, want: []string{ratedNote, fmt.Sprintf(minimumNote, 10)}},
		{
			strength: "11BMR†",
			want: []string{
				fmt.Sprintf(minimumNote, 11),
				fmt.Sprintf(bipodReducedNote, 8),
				mountedNote,
				musketRestNote,
				fmt.Sprintf(twoHandedNote, 17, 22),
			},
		},
		{strength: "11‡", want: []string{fmt.Sprintf(minimumNote, 11), fmt.Sprintf(unreadyNote, 17, 33)}},
		{strength: "1B", want: []string{fmt.Sprintf(minimumNote, 1), bipodNote}},
		{strength: "B", want: []string{bipodNote}},
		{strength: "M", want: []string{mountedNote}},
		{strength: "R", want: []string{musketRestNote}},
	} {
		c.Equal(strings.Join(tc.want, "\n\n"), attack.ParseStrength(tc.strength).Tooltip(tc.ratedST),
			"strength %q, rated ST %v", tc.strength, tc.ratedST)
	}
}
