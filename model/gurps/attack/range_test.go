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

func TestRange(t *testing.T) {
	parse := func(s string) string { return attack.ParseRange(s).String(false) }
	same := []string{
		"",
		"1",
		"1-6",
		"1/10",
		"5/5-10",
		"x0.2",
		"x0.5/x1",
		"1,000/3,000",
	}
	adjusted := []parseCase{
		{"0", ""},
		{"0/5", "5"},
		{"×2", "x2"},
		{"1,000/3,000 mi.", "1,000/3,000 mi"},
		{"1000/3000", "1,000/3,000"},
		{"1/2D 5, Max 10", "5/10"},
		{"6-1", "1-6"},
		{"1/5-10", "5-10"},
		{"10/1", "1"},
		{"5 per point", ""},
		{"B550-551", ""},
		{"C/1", "1"},
		{"2 PBAoE", ""},
		{"ST + Skill/5", ""},
		{"5mi", "5 mi"},
		{"STx2.5", "x2.5"},
		{"100 (Sight)", ""},
		{"max 100", "100"},
		{"10 (spec)", ""},
		{"x0.5/x0.5", "x0.5"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestRangeValidate verifies that negative ranges become zero.
func TestRangeValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.Range]{
		{
			name:     "negative",
			in:       attack.Range{HalfDamage: fxp.NegOne, Min: -fxp.Two, Max: -fxp.Three},
			expected: attack.Range{},
		},
	})
}
