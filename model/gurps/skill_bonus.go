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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SkillBonus{}

// SkillBonus holds an adjustment to a skill.
type SkillBonus struct {
	SkillBonusData
	// selfControlOf is, for the Merchant penalty Entity.processFeatures derives from a trait's self-control roll, that
	// trait, whose roll gives the amount when it is read and whose adjustment whether it applies; see AdjustedAmount
	// and applies.
	selfControlOf *Trait
	// equipmentPenaltyOf is, for the missing-equipment penalty Entity.applyEquipmentPenalties gives a skill, that
	// skill, whose optional specialization the penalty matches and whose grant decides whether it applies; see
	// matchesOptionalSpecialization and applies.
	equipmentPenaltyOf *Skill
}

// SkillBonusData holds an adjustment to a skill which is persisted.
type SkillBonusData struct {
	Type feature.Type `json:"type"`
	FeatureSwitch
	SelectionType                  skillsel.Type `json:"selection_type"`
	NameCriteria                   criteria.Text `json:"name,omitzero"`
	SpecializationCriteria         criteria.Text `json:"specialization,omitzero"`
	OptionalSpecializationCriteria criteria.Text `json:"optional_specialization,omitzero"`
	DamageTypeCriteria             criteria.Text `json:"damage_type,omitzero"`
	TagsCriteria                   criteria.Text `json:"tags,omitzero"`
	WeaponOwnerCriteria
	LeveledAmount
	BonusOwner `json:"-"`
}

// NewSkillBonus creates a new SkillBonus.
func NewSkillBonus() *SkillBonus {
	var s SkillBonus
	s.Type = feature.SkillBonus
	s.SelectionType = skillsel.Name
	s.NameCriteria.Compare = criteria.IsText
	s.SpecializationCriteria.Compare = criteria.AnyText
	s.OptionalSpecializationCriteria.Compare = criteria.AnyText
	s.DamageTypeCriteria.Compare = criteria.AnyText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return &s
}

// FeatureType implements Feature.
func (s *SkillBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *SkillBonus) Clone() Feature {
	return clonePtr(s)
}

// FillWithNameableKeys implements Feature. Only the criteria the selection consults contribute keys: the specialization
// criterion holds the usage when the bonus selects weapons, and only a selection of skills by name has an optional
// specialization.
func (s *SkillBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.SpecializationCriteria.Qualifier)
	if s.SelectionType == skillsel.Name {
		nameable.Extract(m, existing, s.OptionalSpecializationCriteria.Qualifier)
	}
	if s.SelectionType != skillsel.ThisWeapon {
		nameable.Extract(
			m, existing,
			s.NameCriteria.Qualifier,
			s.TagsCriteria.Qualifier,
		)
	}
	if s.SelectionType != skillsel.Name {
		nameable.Extract(m, existing, s.DamageTypeCriteria.Qualifier)
	}
	if s.SelectionType == skillsel.WeaponsWithName {
		s.fillWithNameableKeys(m, existing)
	}
}

// SetLeveledOwner implements Bonus.
func (s *SkillBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AdjustedAmount implements Bonus. The Merchant penalty derived from a trait's self-control roll reads the roll afresh,
// so that a granted trait's is the one the prerequisites being judged see (see Trait.ResolvedSelfControl), as a
// per-level amount reads the level they see.
func (s *SkillBonus) AdjustedAmount() fxp.Int {
	if s.selfControlOf != nil {
		return fxp.FromInteger(s.selfControlOf.ResolvedSelfControl(nil).Penalty())
	}
	return s.LeveledAmount.AdjustedAmount()
}

// applies returns false for a penalty derived from a trait or skill that does not call for it as the prerequisites
// being judged see that trait or skill: the Merchant penalty of a trait whose self-control adjustment is not a major
// cost of living increase (see Trait.ResolvedSelfControlAdjustment), and the missing-equipment penalty of a skill whose
// details come from a grant that treats its prerequisites as met (see prereqsBypassed). Every other bonus applies.
func (s *SkillBonus) applies() bool {
	switch {
	case s.selfControlOf != nil:
		return s.selfControlOf.ResolvedSelfControlAdjustment(nil) == selfctrl.MajorCostOfLivingIncrease
	case s.equipmentPenaltyOf != nil:
		return !prereqsBypassed(s.equipmentPenaltyOf)
	default:
		return true
	}
}

// matchesOptionalSpecialization returns true if the optional specialization criterion matches the value. That of the
// missing-equipment penalty of a skill is the skill's optional specialization as the prerequisites being judged see it
// (see Skill.OptionalSpecializationWithReplacements).
func (s *SkillBonus) matchesOptionalSpecialization(replacements map[string]string, optionalSpecialization string) bool {
	if s.equipmentPenaltyOf != nil {
		return criteria.Text{
			Compare:   criteria.IsText,
			Qualifier: s.equipmentPenaltyOf.OptionalSpecializationWithReplacements(),
		}.Matches(replacements, optionalSpecialization)
	}
	return s.OptionalSpecializationCriteria.Matches(replacements, optionalSpecialization)
}

// AddToTooltip implements Bonus.
func (s *SkillBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if s.selfControlOf != nil {
		s.basicAddToTooltip(&LeveledAmount{Amount: s.AdjustedAmount()}, buffer)
		return
	}
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// Describe implements Feature.
func (s *SkillBonus) Describe(_ *Entity, replacements map[string]string, em func(string) string) string {
	text := describeAmount(s.Amount.StringWithSign(), s.PerLevel, em) + i18n.Text(" to ")
	switch s.SelectionType {
	case skillsel.ThisWeapon:
		text += describeThisWeapon(i18n.Text("skill"), s.SpecializationCriteria, s.DamageTypeCriteria, replacements, em)
	case skillsel.WeaponsWithName:
		text += i18n.Text("the %s of %s", i18n.Text("skill"),
			describeNamedWeapons(s.NameCriteria, s.SpecializationCriteria, s.DamageTypeCriteria, s.TagsCriteria,
				&s.WeaponOwnerCriteria, replacements, em))
	default:
		text += describeTarget(i18n.Text("skill %s"), i18n.Text("skills"), i18n.Text("whose name"), s.NameCriteria,
			replacements, em) +
			describeSpecialization(s.SpecializationCriteria, s.OptionalSpecializationCriteria, replacements, em) +
			describeTags(s.TagsCriteria, replacements, em)
	}
	return describeSwitchable(s.Switchable, text)
}

// Hash writes this object's contents into the hasher.
func (s *SkillBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	xhash.Bool(h, s.Switchable)
	xhash.Num8(h, s.SelectionType)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.OptionalSpecializationCriteria.Hash(h)
	s.DamageTypeCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.WeaponOwnerCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SkillBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SkillBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SkillBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return unmarshalWithLegacyTags(dec, &s.SkillBonusData, &s.TagsCriteria)
}
