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

func TestBlock(t *testing.T) {
	parse := func(s string) string { return attack.ParseBlock(s).String() }
	same := []string{
		"0",
		"-1",
		"No",
	}
	adjusted := []parseCase{
		{"", "No"},
		{"-", "No"},
		{"–", "No"},
		{"—", "No"},
		{"+1", "1"},
	}
	checkFieldParsing(check.New(t), parse, same, adjusted)
}

// TestBlockValidate verifies that a weapon that can't block keeps no modifier.
func TestBlockValidate(t *testing.T) {
	checkValidate(check.New(t), []validateCase[attack.Block]{
		{name: "can't block", in: attack.Block{Modifier: fxp.Two}, expected: attack.Block{}},
	})
}
