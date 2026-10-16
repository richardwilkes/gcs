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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// SelectorScope identifies the kind of node a selector field lives on, which determines how a SelectorOverride matches
// and which criteria the authoring panel offers.
type SelectorScope byte

// Possible SelectorScope values. The zero value is the weapon scope, so weapon descriptors need not set Scope.
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
// Matching is a self-contained criteria comparison, with one matcher per scope. A trait-scoped field matches traits by
// name and tags (MatchesTrait). A weapon-scoped field selects weapons the way a WeaponBonus does, by SelectionType: the
// weapons used with a required skill (matchesRequiredSkill), the weapons of the item carrying the override
// (matchesThisWeapon), or the weapons with a name (MatchesWeapon). Adding a field that lives on some other kind of node
// means giving that scope its own matcher and resolve site; the Override design (priority + specificity ladder, tooltip
// contest) is unchanged.
type SelectorOverride struct {
	SelectorOverrideData
	BonusOwner
}

// SelectorOverrideData is the persisted portion. SelectionType says how a weapon-scoped override selects its weapons
// and is ignored by a trait-scoped one. The specialization and relative level criteria only take part when selecting
// by required skill; the usage criterion does in every selection, unlike a weapon bonus's.
type SelectorOverrideData struct { //nolint:govet // Field order is kept readable for the on-disk JSON rather than packed
	Type feature.Type `json:"type"`
	FeatureSwitch
	Field                  selector.Field  `json:"field"`
	Priority               int             `json:"priority,omitzero"`
	Value                  string          `json:"value"`
	SelectionType          wsel.Type       `json:"selection_type"`
	NameCriteria           criteria.Text   `json:"name,omitzero"`
	SpecializationCriteria criteria.Text   `json:"specialization,omitzero"`
	UsageCriteria          criteria.Text   `json:"usage,omitzero"`
	TagsCriteria           criteria.Text   `json:"tags,omitzero"`
	RelativeLevelCriteria  criteria.Number `json:"level,omitzero"`
}

// NewSelectorOverride creates a new SelectorOverride targeting the given field. A weapon-scoped one starts out
// selecting weapons by their required skill, as a new weapon bonus does.
func NewSelectorOverride(field selector.Field) *SelectorOverride {
	var o SelectorOverride
	o.Type = feature.SelectorOverride
	o.Field = field
	o.SelectionType = wsel.WithName
	if o.weaponScoped() {
		o.SelectionType = wsel.WithRequiredSkill
	}
	o.NameCriteria.Compare = criteria.IsText
	o.SpecializationCriteria.Compare = criteria.AnyText
	o.UsageCriteria.Compare = criteria.AnyText
	o.TagsCriteria.Compare = criteria.AnyText
	o.RelativeLevelCriteria.Compare = criteria.AnyNumber
	if d := SelectorFieldDescriptorFor(field); len(d.SuggestedStates) != 0 {
		o.Value = d.SuggestedStates[0]
	}
	return &o
}

// weaponScoped reports whether the override's field lives on weapons, in which case SelectionType says which ones.
func (o *SelectorOverride) weaponScoped() bool {
	return SelectorFieldDescriptorFor(o.Field).Scope != SelectorScopeTrait
}

// FeatureType implements Feature.
func (o *SelectorOverride) FeatureType() feature.Type {
	return o.Type
}

// Clone implements Feature.
func (o *SelectorOverride) Clone() Feature {
	return clonePtr(o)
}

// FillWithNameableKeys implements Feature. Only the criteria the override matches on are consulted.
func (o *SelectorOverride) FillWithNameableKeys(m, existing map[string]string) {
	switch {
	case !o.weaponScoped():
		nameable.Extract(m, existing, o.NameCriteria.Qualifier, o.TagsCriteria.Qualifier)
	case o.SelectionType == wsel.ThisWeapon:
		nameable.Extract(m, existing, o.UsageCriteria.Qualifier)
	case o.SelectionType == wsel.WithName:
		nameable.Extract(m, existing, o.NameCriteria.Qualifier, o.UsageCriteria.Qualifier, o.TagsCriteria.Qualifier)
	default:
		nameable.Extract(m, existing, o.NameCriteria.Qualifier, o.SpecializationCriteria.Qualifier,
			o.UsageCriteria.Qualifier, o.TagsCriteria.Qualifier)
	}
}

// OverridePriority implements Override.
func (o *SelectorOverride) OverridePriority() int {
	return o.Priority
}

// OverrideSpecificity implements Override. Each criterion that actually constrains something (isn't "any") makes the
// match one notch more specific, so a rule that pins name+usage out-ranks one that only pins name. Only the criteria
// the override actually matches on are counted -- a trait-scoped override never consults the usage criterion (see
// MatchesTrait), and a weapon-scoped one consults the specialization and relative skill level only when selecting by
// required skill -- so a leftover criterion must not win a tie-break it doesn't constrain. An override selecting the
// weapons of its own item pins their name and tags by identity, so it counts as having pinned both.
func (o *SelectorOverride) OverrideSpecificity() int {
	var list []criteria.Text
	specificity := 0
	switch {
	case !o.weaponScoped():
		list = []criteria.Text{o.NameCriteria, o.TagsCriteria}
	case o.SelectionType == wsel.ThisWeapon:
		specificity = 2
		list = []criteria.Text{o.UsageCriteria}
	case o.SelectionType == wsel.WithName:
		list = []criteria.Text{o.NameCriteria, o.UsageCriteria, o.TagsCriteria}
	default:
		list = []criteria.Text{o.NameCriteria, o.SpecializationCriteria, o.UsageCriteria, o.TagsCriteria}
		if !o.RelativeLevelCriteria.IsZero() {
			specificity++
		}
	}
	for _, c := range list {
		if !c.IsZero() {
			specificity++
		}
	}
	return specificity
}

// MatchesWeapon returns true if this override's name, usage and tag criteria match the given weapon, which is how an
// override selecting weapons by name picks them. The criteria are filled from the override owner's substitutions, as
// with any bonus. The caller is responsible for checking that the override targets the field it is interested in.
func (o *SelectorOverride) MatchesWeapon(w *Weapon) bool {
	replacements := bonusReplacements(o)
	return o.NameCriteria.Matches(replacements, w.String()) &&
		o.UsageCriteria.Matches(replacements, w.UsageWithReplacements()) &&
		o.TagsCriteria.MatchesList(replacements, w.Owner.TagList()...)
}

// matchesThisWeapon returns true if this override, carried by the weapon's owner and selecting that owner's weapons,
// applies to the given weapon: only the usage criterion takes part.
func (o *SelectorOverride) matchesThisWeapon(w *Weapon) bool {
	return o.UsageCriteria.Matches(bonusReplacements(o), w.UsageWithReplacements())
}

// matchesRequiredSkill returns true if this override, selecting weapons by their required skill, matches a weapon
// with the given usage and tags that is used with the named skill, which is empty when the weapon uses none. rsl
// supplies that skill's relative level, and is only called when the relative level criterion constrains it, since
// finding it means walking the skills.
func (o *SelectorOverride) matchesRequiredSkill(name, specialization, usage string, tags []string, rsl func() fxp.Int) bool {
	replacements := bonusReplacements(o)
	return matchesRequiredSkillName(o.NameCriteria, replacements, name) &&
		o.SpecializationCriteria.Matches(replacements, specialization) &&
		o.UsageCriteria.Matches(replacements, usage) &&
		o.TagsCriteria.MatchesList(replacements, tags...) &&
		(o.RelativeLevelCriteria.IsZero() || o.RelativeLevelCriteria.Matches(rsl()))
}

// MatchesTrait returns true if this override's criteria match the given trait. Traits have no usage, so only the name
// and tag criteria participate. The criteria are filled from the override owner's substitutions, as with any bonus.
// The caller is responsible for checking that the override targets the field it is interested in.
func (o *SelectorOverride) MatchesTrait(t *Trait) bool {
	replacements := bonusReplacements(o)
	return o.NameCriteria.Matches(replacements, t.NameWithReplacements()) &&
		o.TagsCriteria.MatchesList(replacements, t.Tags...)
}

// Describe implements Feature. The value is shown by its title, where its field has titles for its values, and as an
// empty pair of quotes while it is empty, as an empty qualifier is.
func (o *SelectorOverride) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	d := SelectorFieldDescriptorFor(o.Field)
	value := o.Value
	if d.StateTitle != nil {
		value = d.StateTitle(value)
	}
	if value == "" {
		value = `""`
	} else {
		value = em(value)
	}
	var target string
	switch {
	case d.Scope == SelectorScopeTrait:
		target = describeTarget(i18n.Text("traits named %s"), i18n.Text("traits"), i18n.Text("whose name"),
			o.NameCriteria, replacements, em) + describeTags(o.TagsCriteria, replacements, em)
	case o.SelectionType == wsel.ThisWeapon:
		target = i18n.Text("this weapon") + describeWhose(i18n.Text("when its usage"), o.UsageCriteria, replacements, em)
	case o.SelectionType == wsel.WithName:
		target = describeNamedWeapons(o.NameCriteria, o.UsageCriteria, o.TagsCriteria, replacements, em)
	default:
		target = describeSkillSelectedWeapons(o.NameCriteria, o.SpecializationCriteria, o.UsageCriteria,
			o.TagsCriteria, o.RelativeLevelCriteria, replacements, em)
	}
	return describeSwitchable(o.Switchable, i18n.Text("Sets %s to %s on %s (priority %d)",
		em(o.Field.String()), value, target, o.Priority))
}

// Hash writes this object's contents into the hasher.
func (o *SelectorOverride) Hash(h hash.Hash) {
	if o == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, o.Type)
	xhash.Bool(h, o.Switchable)
	xhash.Num8(h, o.Field)
	xhash.Num64(h, int64(o.Priority))
	xhash.StringWithLen(h, o.Value)
	xhash.Num8(h, o.SelectionType)
	o.NameCriteria.Hash(h)
	o.SpecializationCriteria.Hash(h)
	o.UsageCriteria.Hash(h)
	o.TagsCriteria.Hash(h)
	o.RelativeLevelCriteria.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (o *SelectorOverride) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &o.SelectorOverrideData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. Before there was a selection type, an override selected weapons
// by name, so a file without one reads as that.
func (o *SelectorOverride) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	o.SelectionType = wsel.WithName
	return json.UnmarshalDecode(dec, &o.SelectorOverrideData)
}

// SelectorFieldDescriptor describes a selector field: the values worth suggesting in the authoring UI, an optional
// human label for each of those values, an optional validator, and whether values outside the suggested set are
// permitted.
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
	// The lookup deliberately uses the field as-is rather than field.EnsureValid(), since that would map every unknown
	// field onto the first enum value and hand back that field's descriptor instead of the safe free-form one.
	if d, ok := selectorFieldDescriptors[field]; ok {
		return d
	}
	return SelectorFieldDescriptor{Field: field, FreeForm: true}
}
