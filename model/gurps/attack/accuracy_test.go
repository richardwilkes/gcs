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

func TestAccuracy(t *testing.T) {
	parse := func(s string) string { return attack.ParseAccuracy(s).String() }
	same := []string{
		"0",
		"1",
		"1+3",
		"0+3",
		"Jet",
	}
	adjusted := []parseCase{
		{"", "0"},
		{"-", "0"},
		{"?", "0"},
		{"-1", "0"},
		{"+1", "1"},
		{"+1+3", "1+3"},
		{"1+-2", "1"},
		{"1+0", "1"},
		{"51,", "51"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestAccuracyValidate verifies that a jet keeps neither a base nor a scope accuracy.
func TestAccuracyValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.Accuracy]{
		{
			name:     "jet",
			in:       attack.Accuracy{Base: fxp.Three, Scope: fxp.Two, Jet: true},
			expected: attack.Accuracy{Jet: true},
		},
	})
}
