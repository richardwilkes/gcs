// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
)

// Features holds a list of features.
type Features []Feature

// Clone creates a copy of the features.
func (f Features) Clone() Features {
	if len(f) == 0 {
		return nil
	}
	result := make([]Feature, 0, len(f))
	for _, one := range f {
		result = append(result, one.Clone())
	}
	return result
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (f *Features) UnmarshalJSON(data []byte) error {
	var s []*json.RawMessage
	if err := json.Unmarshal(data, &s); err != nil {
		return errs.Wrap(err)
	}
	*f = make([]Feature, len(s))
	for i, one := range s {
		var justTypeData struct {
			Type feature.Type `json:"type"`
		}
		if err := json.Unmarshal(*one, &justTypeData); err != nil {
			return errs.Wrap(err)
		}
		var feat Feature
		switch justTypeData.Type {
		case feature.AttributeBonus:
			feat = &AttributeBonus{}
		case feature.ConditionalModifier:
			feat = &ConditionalModifierBonus{}
		case feature.ContainedWeightReduction:
			feat = &ContainedWeightReduction{}
		case feature.CostReduction:
			feat = &CostReduction{}
		case feature.DRBonus:
			feat = &DRBonus{}
		case feature.ReactionBonus:
			feat = &ReactionBonus{}
		case feature.SkillBonus:
			feat = &SkillBonus{}
		case feature.SkillPointBonus:
			feat = &SkillPointBonus{}
		case feature.SpellBonus:
			feat = &SpellBonus{}
		case feature.SpellPointBonus:
			feat = &SpellPointBonus{}
		case feature.WeaponBonus,
			feature.WeaponAccBonus,
			feature.WeaponScopeAccBonus,
			feature.WeaponDRDivisorBonus,
			feature.WeaponEffectiveSTBonus,
			feature.WeaponMinSTBonus,
			feature.WeaponMinReachBonus,
			feature.WeaponMaxReachBonus,
			feature.WeaponHalfDamageRangeBonus,
			feature.WeaponMinRangeBonus,
			feature.WeaponMaxRangeBonus,
			feature.WeaponBulkBonus,
			feature.WeaponRecoilBonus,
			feature.WeaponParryBonus,
			feature.WeaponBlockBonus,
			feature.WeaponRofMode1ShotsBonus,
			feature.WeaponRofMode1SecondaryBonus,
			feature.WeaponRofMode2ShotsBonus,
			feature.WeaponRofMode2SecondaryBonus,
			feature.WeaponSwitch,
			feature.WeaponNonChamberShotsBonus,
			feature.WeaponChamberShotsBonus,
			feature.WeaponShotDurationBonus,
			feature.WeaponReloadTimeBonus:
			feat = &WeaponBonus{}
		default:
			return errs.Newf(i18n.Text("Unknown feature type: %s"), justTypeData.Type)
		}
		if err := json.Unmarshal(*one, &feat); err != nil {
			return errs.Wrap(err)
		}
		(*f)[i] = feat
	}
	return nil
}
