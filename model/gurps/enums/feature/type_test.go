// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package feature_test

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestIsWeaponBonusMatchesWeaponKeys verifies that the hand-maintained IsWeaponBonus agrees with the generated enum:
// every type whose key starts with "weapon_" is a weapon bonus and nothing else is. A weapon type missing from
// IsWeaponBonus loads as an UnknownFeature and is absent from the feature editor.
func TestIsWeaponBonusMatchesWeaponKeys(t *testing.T) {
	c := check.New(t)
	count := 0
	for _, one := range feature.Types {
		expected := strings.HasPrefix(one.Key(), "weapon_")
		c.Equal(expected, one.IsWeaponBonus(), "IsWeaponBonus for %s", one.Key())
		if expected {
			count++
		}
	}
	c.Equal(24, count, "expected number of weapon bonus types")
}
