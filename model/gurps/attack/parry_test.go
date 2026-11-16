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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/attack"
	"github.com/richardwilkes/toolbox/v2/check"
)

func TestParry(t *testing.T) {
	parse := func(s string) string { return attack.ParseParry(s).String() }
	same := []string{
		"0",
		"-1",
		"10",
		"9U",
		"-2F",
		"0FU",
		"No",
	}
	adjusted := []parseCase{
		{"", "No"},
		{"-", "No"},
		{"–", "No"},
		{"—", "No"},
		{"+1", "1"},
		{"0U / 0", "0U"},
		{"13 (x5)", "13"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestParryValidate verifies that a weapon that can't parry keeps neither a modifier nor its markers.
func TestParryValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.Parry]{
		{
			name:     "can't parry",
			in:       attack.Parry{Fencing: true, Unbalanced: true, Modifier: fxp.Two},
			expected: attack.Parry{},
		},
	})
}

// TestParryTooltip verifies that the fencing note gives the retreating parry as 3 above the parry rather than the
// usual 1 above it, and that the unbalanced note follows it.
func TestParryTooltip(t *testing.T) {
	c := check.New(t)
	const (
		fencingNote    = "F: Fencing weapon. When retreating, your Parry is %d (instead of %d). You suffer a -2 cumulative penalty for multiple parries after the first on the same turn instead of the usual -4. Flails cannot be parried by this weapon."
		unbalancedNote = "U: Unbalanced weapon. You cannot use it to parry if you have already used it to attack this turn (or vice-versa)."
	)
	for _, tc := range []struct {
		parry string
		want  string
	}{
		{parry: "2F", want: fmt.Sprintf(fencingNote, 5, 3)},
		{parry: "-1U", want: unbalancedNote},
		{parry: "0FU", want: fmt.Sprintf(fencingNote, 3, 1) + "\n\n" + unbalancedNote},
	} {
		c.Equal(tc.want, attack.ParseParry(tc.parry).Tooltip(), "parry %q", tc.parry)
	}
}
