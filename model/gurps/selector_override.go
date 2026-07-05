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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Override = &SelectorOverride{}

// SelectorOverride is the general Override for replacing a multi-state field with a chosen value. A single feature type
// covers every such field; which one it targets is carried in Field, and the value it sets is a string interpreted by
// that field (see SelectorFieldDescriptorFor for the field's suggested states). This subsumes what would otherwise be a
// separate feature type per field — damage type is just Field == selector.WeaponDamageType.
//
// NOTE: matching is currently a self-contained name/usage/tags comparison suitable for weapon-scoped fields. Adding a
// field that lives somewhere other than a weapon means giving that field its own matcher and resolve site; the Override
// design (priority + specificity ladder, tooltip contest) is unchanged.
type SelectorOverride struct {
	SelectorOverrideData
	BonusOwner
}

// SelectorOverrideData is the persisted portion.
type SelectorOverrideData struct { //nolint:govet // Field order is kept readable for the on-disk JSON rather than packed
	Type          feature.Type   `json:"type"`
	Field         selector.Field `json:"field"`
	Priority      int            `json:"priority,omitzero"`
	Value         string         `json:"value"`
	NameCriteria  criteria.Text  `json:"name,omitzero"`
	UsageCriteria criteria.Text  `json:"usage,omitzero"`
	TagsCriteria  criteria.Text  `json:"tags,omitzero"`
}

// NewSelectorOverride creates a new SelectorOverride targeting the given field.
func NewSelectorOverride(field selector.Field) *SelectorOverride {
	var o SelectorOverride
	o.Type = feature.SelectorOverride
	o.Field = field
	o.NameCriteria.Compare = criteria.IsText
	o.UsageCriteria.Compare = criteria.AnyText
	o.TagsCriteria.Compare = criteria.AnyText
	if d := SelectorFieldDescriptorFor(field); len(d.SuggestedStates) != 0 {
		o.Value = d.SuggestedStates[0]
	}
	return &o
}

// FeatureType implements Feature.
func (o *SelectorOverride) FeatureType() feature.Type {
	return o.Type
}

// Clone implements Feature.
func (o *SelectorOverride) Clone() Feature {
	other := *o
	return &other
}

// FillWithNameableKeys implements Feature.
func (o *SelectorOverride) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(o.NameCriteria.Qualifier, m, existing)
	nameable.Extract(o.UsageCriteria.Qualifier, m, existing)
	nameable.Extract(o.TagsCriteria.Qualifier, m, existing)
}

// OverridePriority implements Override.
func (o *SelectorOverride) OverridePriority() int {
	return o.Priority
}

// OverrideSpecificity implements Override. Each criterion that actually constrains something (isn't "any") makes the
// match one notch more specific, so a rule that pins name+usage out-ranks one that only pins name.
func (o *SelectorOverride) OverrideSpecificity() int {
	specificity := 0
	for _, c := range []criteria.Text{o.NameCriteria, o.UsageCriteria, o.TagsCriteria} {
		if !c.IsZero() {
			specificity++
		}
	}
	return specificity
}

// MatchesWeapon returns true if this override targets a weapon-scoped field and applies to the given weapon.
func (o *SelectorOverride) MatchesWeapon(w *Weapon) bool {
	replacements := w.NameableReplacements()
	return o.NameCriteria.Matches(replacements, w.String()) &&
		o.UsageCriteria.Matches(replacements, w.UsageWithReplacements()) &&
		o.TagsCriteria.MatchesList(replacements, w.Owner.TagList()...)
}

// Hash writes this object's contents into the hasher.
func (o *SelectorOverride) Hash(h hash.Hash) {
	if o == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, o.Type)
	xhash.Num8(h, o.Field)
	xhash.Num64(h, int64(o.Priority))
	xhash.StringWithLen(h, o.Value)
	o.NameCriteria.Hash(h)
	o.UsageCriteria.Hash(h)
	o.TagsCriteria.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (o *SelectorOverride) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &o.SelectorOverrideData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (o *SelectorOverride) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return json.UnmarshalDecode(dec, &o.SelectorOverrideData)
}

// SelectorFieldDescriptor describes a selector field: the values worth suggesting in the authoring UI and whether
// values outside that set are permitted. This is the domain knowledge that turns a bare enum into an editable field.
type SelectorFieldDescriptor struct {
	SuggestedStates []string
	Field           selector.Field
	FreeForm        bool
}

// selectorFieldDescriptors holds the descriptor for every selector.Field. GURPS damage types are a free-form string in
// this model, so WeaponDamageType offers the common abbreviations as suggestions but still accepts anything.
var selectorFieldDescriptors = map[selector.Field]SelectorFieldDescriptor{
	selector.WeaponDamageType: {
		Field:           selector.WeaponDamageType,
		SuggestedStates: []string{"cr", "cut", "imp", "pi-", "pi", "pi+", "pi++", "burn", "cor", "fat", "tox"},
		FreeForm:        true,
	},
}

// SelectorFieldDescriptorFor returns the descriptor for the given field, or a zero-value free-form descriptor if the
// field is unknown.
func SelectorFieldDescriptorFor(field selector.Field) SelectorFieldDescriptor {
	if d, ok := selectorFieldDescriptors[field.EnsureValid()]; ok {
		return d
	}
	return SelectorFieldDescriptor{Field: field, FreeForm: true}
}
