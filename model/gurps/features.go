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

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
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
	list, err := unmarshalTypedList(dec, feature.ExtractKnownType, allocFeature,
		func(kind string, raw jsontext.Value) Feature { return NewUnknownFeature(kind, raw) })
	if err != nil {
		return err
	}
	*f = list
	return nil
}

// allocFeature returns an empty Feature of the concrete type that represents featureType, or nil if there is none.
func allocFeature(featureType feature.Type) Feature {
	if featureType.IsWeaponBonus() {
		return &WeaponBonus{}
	}
	switch featureType {
	case feature.AttributeBonus:
		return &AttributeBonus{}
	case feature.ConditionalModifier:
		return &ConditionalModifierBonus{}
	case feature.ContainedWeightReduction:
		return &ContainedWeightReduction{}
	case feature.CostReduction:
		return &CostReduction{}
	case feature.EquipmentMaxUsesBonus:
		return &EquipmentMaxUsesBonus{}
	case feature.DRBonus:
		return &DRBonus{}
	case feature.ReactionBonus:
		return &ReactionBonus{}
	case feature.SkillBonus:
		return &SkillBonus{}
	case feature.SkillPointBonus:
		return &SkillPointBonus{}
	case feature.SpellBonus:
		return &SpellBonus{}
	case feature.SpellPointBonus:
		return &SpellPointBonus{}
	case feature.TraitBonus:
		return &TraitBonus{}
	case feature.TraitMaxLevelBonus:
		return &TraitMaxLevelBonus{}
	case feature.SelectorOverride:
		return &SelectorOverride{}
	default:
		return nil
	}
}
