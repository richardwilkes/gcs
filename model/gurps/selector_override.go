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
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/frequency"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selector"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stdmg"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// SelectorScope identifies the kind of node a selector field lives on, which determines how a SelectorOverride matches
// and which criteria the authoring panel offers.
type SelectorScope byte

// Possible SelectorScope values. The zero value is the weapon scope, so existing weapon descriptors need no change.
const (
	SelectorScopeWeapon SelectorScope = iota
	SelectorScopeTrait
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

// MatchesTrait returns true if this override targets a trait-scoped field and applies to the given trait. Traits have
// no usage, so only the name and tag criteria participate.
func (o *SelectorOverride) MatchesTrait(t *Trait) bool {
	replacements := t.NameableReplacements()
	return o.NameCriteria.Matches(replacements, t.NameWithReplacements()) &&
		o.TagsCriteria.MatchesList(replacements, t.Tags...)
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

// SelectorFieldDescriptor describes a selector field: the values worth suggesting in the authoring UI, an optional
// human label for each of those values, an optional validator, and whether values outside the suggested set are
// permitted. This is the domain knowledge that turns a bare enum into an editable field.
type SelectorFieldDescriptor struct {
	SuggestedStates []string
	StateTitle      func(state string) string // optional human label in the picker; nil means the state itself
	Validate        func(value string) bool   // optional; nil means any value is accepted
	Field           selector.Field
	Scope           SelectorScope
	FreeForm        bool
}

// damageTypeStates are the common GURPS damage type abbreviations, shared by the damage type and fragmentation type
// fields. Both are free-form strings in this model, so these are offered as suggestions but anything is accepted.
var damageTypeStates = []string{"cr", "cut", "imp", "pi-", "pi", "pi+", "pi++", "burn", "cor", "fat", "tox"}

// validFixedPoint reports whether value parses as a fixed-point number, used to validate the numeric damage fields.
func validFixedPoint(value string) bool {
	_, err := fxp.FromString(value)
	return err == nil
}

// frequencyStateKey and frequencyFromStateKey convert a frequency roll to and from the canonical string stored in a
// frequency override. The roll's own byte value is used as the key, since frequency.Roll has no string key of its own.
func frequencyStateKey(r frequency.Roll) string { return strconv.Itoa(int(r)) }

func frequencyFromStateKey(state string) frequency.Roll {
	n, err := strconv.Atoi(state)
	if err != nil {
		return frequency.None
	}
	return frequency.Roll(n).EnsureValid()
}

// traitFrequencyStates lists every frequency roll as a stored key, in the order the frequency enum defines them.
var traitFrequencyStates = func() []string {
	states := make([]string, len(frequency.Rolls))
	for i, r := range frequency.Rolls {
		states[i] = frequencyStateKey(r)
	}
	return states
}()

// selfControlRollStateKey and selfControlRollFromStateKey convert a self-control roll to and from its stored key. Like
// frequency, selfctrl.Roll has no string key of its own, so its byte value is used.
func selfControlRollStateKey(r selfctrl.Roll) string { return strconv.Itoa(int(r)) }

func selfControlRollFromStateKey(state string) selfctrl.Roll {
	n, err := strconv.Atoi(state)
	if err != nil {
		return selfctrl.None
	}
	return selfctrl.Roll(n).EnsureValid()
}

// traitSelfControlRollStates lists every self-control roll as a stored key, in enum order.
var traitSelfControlRollStates = func() []string {
	states := make([]string, len(selfctrl.Rolls))
	for i, r := range selfctrl.Rolls {
		states[i] = selfControlRollStateKey(r)
	}
	return states
}()

// traitSelfControlAdjustmentStates lists every self-control adjustment by its enum key, in enum order.
var traitSelfControlAdjustmentStates = func() []string {
	states := make([]string, len(selfctrl.Adjustments))
	for i, a := range selfctrl.Adjustments {
		states[i] = a.Key()
	}
	return states
}()

// selectorFieldDescriptors holds the descriptor for every selector.Field.
var selectorFieldDescriptors = map[selector.Field]SelectorFieldDescriptor{
	selector.WeaponDamageType: {
		Field:           selector.WeaponDamageType,
		SuggestedStates: damageTypeStates,
		FreeForm:        true,
	},
	selector.WeaponFragmentationType: {
		Field:           selector.WeaponFragmentationType,
		SuggestedStates: damageTypeStates,
		FreeForm:        true,
	},
	selector.WeaponDamageStrengthBasis: {
		Field: selector.WeaponDamageStrengthBasis,
		SuggestedStates: []string{
			stdmg.None.Key(), stdmg.Thrust.Key(), stdmg.LiftingThrust.Key(), stdmg.TelekineticThrust.Key(),
			stdmg.IQThrust.Key(), stdmg.Swing.Key(), stdmg.LiftingSwing.Key(), stdmg.TelekineticSwing.Key(),
			stdmg.IQSwing.Key(),
		},
		StateTitle: func(state string) string { return stdmg.ExtractOption(state).String() },
		FreeForm:   false,
	},
	// Dice-spec fields are free-form strings (they may even hold a script), so no validator is applied.
	selector.WeaponBaseDamageDice:         {Field: selector.WeaponBaseDamageDice, FreeForm: true},
	selector.WeaponBaseDamageDicePerLevel: {Field: selector.WeaponBaseDamageDicePerLevel, FreeForm: true},
	selector.WeaponFragmentationDice:      {Field: selector.WeaponFragmentationDice, FreeForm: true},
	// Numeric fields accept any fixed-point value.
	selector.WeaponArmorDivisor:              {Field: selector.WeaponArmorDivisor, FreeForm: true, Validate: validFixedPoint},
	selector.WeaponFragmentationArmorDivisor: {Field: selector.WeaponFragmentationArmorDivisor, FreeForm: true, Validate: validFixedPoint},
	selector.WeaponDamageStrengthMultiplier:  {Field: selector.WeaponDamageStrengthMultiplier, FreeForm: true, Validate: validFixedPoint},
	selector.WeaponDamagePerDieModifier:      {Field: selector.WeaponDamagePerDieModifier, FreeForm: true, Validate: validFixedPoint},
	selector.TraitSelfControlRoll: {
		Field:           selector.TraitSelfControlRoll,
		SuggestedStates: traitSelfControlRollStates,
		StateTitle:      func(state string) string { return selfControlRollFromStateKey(state).String() },
		Scope:           SelectorScopeTrait,
		FreeForm:        false,
	},
	selector.TraitSelfControlAdjustment: {
		Field:           selector.TraitSelfControlAdjustment,
		SuggestedStates: traitSelfControlAdjustmentStates,
		StateTitle:      func(state string) string { return selfctrl.ExtractAdjustment(state).String() },
		Scope:           SelectorScopeTrait,
		FreeForm:        false,
	},
	selector.TraitFrequency: {
		Field:           selector.TraitFrequency,
		SuggestedStates: traitFrequencyStates,
		StateTitle:      func(state string) string { return frequencyFromStateKey(state).String() },
		Scope:           SelectorScopeTrait,
		FreeForm:        false,
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
