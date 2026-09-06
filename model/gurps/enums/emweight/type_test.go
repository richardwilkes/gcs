// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emweight_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTypeFromStringRestrictsToPermitted verifies that a Type keeps a classification it permits and otherwise falls
// back to its first permitted Value, and that an addition carries its weight unit through formatting.
func TestTypeFromStringRestrictsToPermitted(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		weightType emweight.Type
		input      string
		expected   emweight.Value
		formatted  string
	}{
		{emweight.Original, "+2", emweight.Addition, "+2 lb"},
		{emweight.Original, "+2 kg", emweight.Addition, "+2 kg"},
		{emweight.Original, "+10%", emweight.PercentageAdder, "+10%"},
		{emweight.Original, "×2", emweight.Addition, "+2 lb"},
		{emweight.Original, "x50%", emweight.Addition, "+50 lb"},
		{emweight.Base, "+2", emweight.Addition, "+2 lb"},
		{emweight.Base, "+10%", emweight.Addition, "+10 lb"},
		{emweight.Base, "X1/2", emweight.Multiplier, "x1/2"},
		{emweight.Base, "×50%", emweight.PercentageMultiplier, "x50%"},
		{emweight.Final, "2x", emweight.Multiplier, "x2"},
	} {
		c.Equal(one.expected, one.weightType.FromString(one.input), "test %d: %q", i, one.input)
		c.Equal(one.formatted, one.weightType.Format(one.input, fxp.Pound), "test %d: %q", i, one.input)
	}
}
