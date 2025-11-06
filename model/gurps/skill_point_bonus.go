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
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Bonus = &SkillPointBonus{}

// SkillPointBonus holds an adjustment to a skill's points.
type SkillPointBonus struct {
	SkillPointBonusData
}

// SkillPointBonusData holds an adjustment to a skill's points which is persisted.
type SkillPointBonusData struct {
	Type                   feature.Type  `json:"type"`
	NameCriteria           criteria.Text `json:"name,omitzero"`
	SpecializationCriteria criteria.Text `json:"specialization,omitzero"`
	TagsCriteria           criteria.Text `json:"tags,omitzero"`
	LeveledAmount
	BonusOwner `json:"-"`
}

// NewSkillPointBonus creates a new SkillPointBonus.
func NewSkillPointBonus() *SkillPointBonus {
	var s SkillPointBonus
	s.Type = feature.SkillPointBonus
	s.NameCriteria.Compare = criteria.IsText
	s.SpecializationCriteria.Compare = criteria.AnyText
	s.TagsCriteria.Compare = criteria.AnyText
	s.Amount = fxp.One
	return &s
}

// FeatureType implements Feature.
func (s *SkillPointBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *SkillPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SkillPointBonus) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(s.NameCriteria.Qualifier, m, existing)
	nameable.Extract(s.SpecializationCriteria.Qualifier, m, existing)
	nameable.Extract(s.TagsCriteria.Qualifier, m, existing)
}

// SetLeveledOwner implements Bonus.
func (s *SkillPointBonus) SetLeveledOwner(owner LeveledOwner) {
	s.LeveledOwner = owner
}

// AddToTooltip implements Bonus.
func (s *SkillPointBonus) AddToTooltip(buffer *xbytes.InsertBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(s.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(s.Format())
		if s.AdjustedAmount() == fxp.One {
			buffer.WriteString(i18n.Text(" pt]"))
		} else {
			buffer.WriteString(i18n.Text(" pts]"))
		}
	}
}

// Hash writes this object's contents into the hasher.
func (s *SkillPointBonus) Hash(h hash.Hash) {
	if s == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, s.Type)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}

// MarshalJSONTo implements json.MarshalerTo.
func (s *SkillPointBonus) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &s.SkillPointBonusData)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (s *SkillPointBonus) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	var content struct {
		SkillPointBonusData
		OldTagsCriteria criteria.Text `json:"category"`
	}
	if err := json.UnmarshalDecode(dec, &content); err != nil {
		return err
	}
	s.SkillPointBonusData = content.SkillPointBonusData
	if s.TagsCriteria.IsZero() && !content.OldTagsCriteria.IsZero() {
		s.TagsCriteria = content.OldTagsCriteria
	}
	return nil
}
