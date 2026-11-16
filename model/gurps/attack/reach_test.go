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

func TestReach(t *testing.T) {
	parse := func(s string) string { return attack.ParseReach(s).String() }
	same := []string{
		"",
		"1",
		"1*",
		"1-2",
		"C",
		"C,1",
		"C,2",
		"C,2-3",
		"C,2-3*",
	}
	adjusted := []parseCase{
		{"-", ""},
		{"1,2", "1-2"},
		{"5/10", "5"},
		{"C-5", "C,1-5"},
		{"C, 2 - 3", "C,2-3"},
		{"C,2,3", "C,2-3"},
		{"Special", ""},
		{"  1 , 3 ", "1-3"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestReachValidate verifies that negative reaches become zero and that the maximum is never below the minimum.
func TestReachValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.Reach]{
		{name: "negative", in: attack.Reach{Min: fxp.NegOne, Max: -fxp.Two}, expected: attack.Reach{}},
		{
			name:     "maximum below minimum",
			in:       attack.Reach{Min: fxp.Three, Max: fxp.One},
			expected: attack.Reach{Min: fxp.Three, Max: fxp.Three},
		},
	})
}
