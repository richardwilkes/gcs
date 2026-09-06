// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestTypeFromStringRestrictsToPermitted verifies that a Type keeps a classification it permits and otherwise falls
// back to its first permitted Value.
func TestTypeFromStringRestrictsToPermitted(t *testing.T) {
	c := check.New(t)
	for i, one := range []struct {
		costType  emcost.Type
		input     string
		expected  emcost.Value
		formatted string
	}{
		{emcost.Original, "+2", emcost.Addition, "+2"},
		{emcost.Original, "+10%", emcost.Percentage, "+10%"},
		{emcost.Original, "×2", emcost.Multiplier, "x2"},
		{emcost.Original, "+2 CF", emcost.Addition, "+2"},
		{emcost.Base, "+2 CF", emcost.CostFactor, "+2 CF"},
		{emcost.Base, "X2", emcost.Multiplier, "x2"},
		{emcost.Base, "+10%", emcost.CostFactor, "+10 CF"},
		{emcost.Base, "+2", emcost.CostFactor, "+2 CF"},
		{emcost.Final, "+2 CF", emcost.Addition, "+2"},
	} {
		c.Equal(one.expected, one.costType.FromString(one.input), "test %d: %q", i, one.input)
		c.Equal(one.formatted, one.costType.Format(one.input), "test %d: %q", i, one.input)
	}
}
