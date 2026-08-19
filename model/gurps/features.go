// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/v2/errs"
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

// AnySwitchable returns true if any of the features are switchable.
func (f Features) AnySwitchable() bool {
	for _, one := range f {
		if one.IsSwitchable() {
			return true
		}
	}
	return false
}

// Active returns the features that currently take effect for an owner whose switch is in the given state: every
// feature that is not switchable, plus the switchable ones only when switchedOn is true. The receiver is returned
// unchanged (no allocation) when nothing needs to be filtered out.
func (f Features) Active(switchedOn bool) Features {
	if switchedOn || !f.AnySwitchable() {
		return f
	}
	result := make(Features, 0, len(f))
	for _, one := range f {
		if !one.IsSwitchable() {
			result = append(result, one)
		}
	}
	return result
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (f *Features) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var v []jsontext.Value
	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return errs.Wrap(err)
	}
	*f = make([]Feature, len(v))
	for i, one := range v {
		var justTypeData struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(one, &justTypeData, dec.Options()); err != nil {
			return errs.Wrap(err)
		}
		// Note that the type is extracted as a string and resolved with ExtractKnownType rather than being unmarshaled
		// directly into a feature.Type: the enum's UnmarshalText maps anything it doesn't recognize onto the first
		// value, which would turn a feature written by a newer version of GCS into a bogus AttributeBonus.
		featureType, known := feature.ExtractKnownType(justTypeData.Type)
		if !known {
			(*f)[i] = NewUnknownFeature(justTypeData.Type, one)
			continue
		}
		var feat Feature
		switch featureType {
		case feature.AttributeBonus:
			feat = &AttributeBonus{}
		case feature.ConditionalModifier:
			feat = &ConditionalModifierBonus{}
		case feature.ContainedWeightReduction:
			feat = &ContainedWeightReduction{}
		case feature.CostReduction:
			feat = &CostReduction{}
		case feature.EquipmentMaxUsesBonus:
			feat = &EquipmentMaxUsesBonus{}
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
		case feature.TraitBonus:
			feat = &TraitBonus{}
		case feature.TraitMaxLevelBonus:
			feat = &TraitMaxLevelBonus{}
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
		case feature.SelectorOverride:
			feat = &SelectorOverride{}
		default:
			// A known type that has no case above, or the Unknown type itself. Preserve rather than discard.
			(*f)[i] = NewUnknownFeature(justTypeData.Type, one)
			continue
		}
		if err := json.Unmarshal(one, &feat, dec.Options()); err != nil {
			return errs.Wrap(err)
		}
		(*f)[i] = feat
	}
	return nil
}
