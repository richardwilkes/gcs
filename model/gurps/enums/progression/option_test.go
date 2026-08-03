// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package progression_test

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestDamageProgressionsAlwaysSetDiceMultiplier verifies that every progression produces dice with a multiplier of 1.
// A zero multiplier is not merely a display nuisance: addDice in model/gurps/weapon_damage.go multiplies each side's
// average by its multiplier when combining dice with differing side counts, so a zero would silently erase the entire
// ST-based damage contribution.
func TestDamageProgressionsAlwaysSetDiceMultiplier(t *testing.T) {
	c := check.New(t)
	for _, option := range progression.Options {
		for st := 1; st <= 100; st++ {
			c.Equal(1, option.Thrust(st).Multiplier, "%s thrust at ST %d", option.Key(), st)
			c.Equal(1, option.Swing(st).Multiplier, "%s swing at ST %d", option.Key(), st)
		}
	}
}
