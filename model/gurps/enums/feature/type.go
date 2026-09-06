// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package feature

// SelectableTypes holds the possible Type values that may be chosen by the user. Unknown is excluded, since it exists
// only to hold onto feature data this version of GCS doesn't understand.
var SelectableTypes []Type

// SelectableTypesWithoutContainedWeightReduction holds the same values as SelectableTypes, minus the
// ContainedWeightReduction.
var SelectableTypesWithoutContainedWeightReduction []Type

func init() {
	SelectableTypes = make([]Type, 0, len(Types)-1)
	SelectableTypesWithoutContainedWeightReduction = make([]Type, 0, len(Types)-2)
	for _, one := range Types {
		if one == Unknown {
			continue
		}
		SelectableTypes = append(SelectableTypes, one)
		if one != ContainedWeightReduction {
			SelectableTypesWithoutContainedWeightReduction = append(SelectableTypesWithoutContainedWeightReduction, one)
		}
	}
}

// IsWeaponBonus returns true if the type is one of the weapon bonuses, all of which are represented by the same
// WeaponBonus feature and differ only in which weapon stat they adjust.
func (enum Type) IsWeaponBonus() bool {
	switch enum {
	case WeaponBonus,
		WeaponAccBonus,
		WeaponScopeAccBonus,
		WeaponDRDivisorBonus,
		WeaponEffectiveSTBonus,
		WeaponMinSTBonus,
		WeaponMinReachBonus,
		WeaponMaxReachBonus,
		WeaponHalfDamageRangeBonus,
		WeaponMinRangeBonus,
		WeaponMaxRangeBonus,
		WeaponRecoilBonus,
		WeaponBulkBonus,
		WeaponParryBonus,
		WeaponBlockBonus,
		WeaponRofMode1ShotsBonus,
		WeaponRofMode1SecondaryBonus,
		WeaponRofMode2ShotsBonus,
		WeaponRofMode2SecondaryBonus,
		WeaponNonChamberShotsBonus,
		WeaponChamberShotsBonus,
		WeaponShotDurationBonus,
		WeaponReloadTimeBonus,
		WeaponSwitch:
		return true
	default:
		return false
	}
}
