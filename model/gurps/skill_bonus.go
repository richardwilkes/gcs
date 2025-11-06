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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SkillBonus{}

// SkillBonus holds an adjustment to a skill.
type SkillBonus struct {
	SkillBonusData
}

// SkillBonusData holds an adjustment to a skill which is persisted.
type SkillBonusData struct {
	Type                   feature.Type  `json:"type"`
	SelectionType          skillsel.Type `json:"selection_type"`
	NameCriteria           criteria.Text `json:"name,omitzero"`
	SpecializationCriteria criteria.Text `json:"specialization,omitzero"`
	TagsCriteria           criteria.Text `json:"tags,omitzero"`
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
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SkillBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(s.SpecializationCriteria.Qualifier, m, existing)
	if s.SelectionType != skillsel.ThisWeapon {
		nameable.Extract(s.NameCriteria.Qualifier, m, existing)
		nameable.Extract(s.TagsCriteria.Qualifier, m, existing)
	}
}

// SetLeveledOwner implements Bonus.
func (s *SkillBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *SkillBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (s *SkillBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	xhash.Num8(h, s.SelectionType)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SkillBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SkillBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SkillBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		SkillBonusData
		OldTagsCriteria criteria.Text `json:"category"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.SkillBonusData = content.SkillBonusData
	if s.TagsCriteria.IsZero() && !content.OldTagsCriteria.IsZero() {
		s.TagsCriteria = content.OldTagsCriteria
	}
	return nil
}
