// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SkillPointBonus{}

// SkillPointBonus holds an adjustment to a skill's points.
type SkillPointBonus struct {
	Type                   feature.Type   `json:"type"`
	NameCriteria           StringCriteria `json:"name,omitempty"`
	SpecializationCriteria StringCriteria `json:"specialization,omitempty"`
	TagsCriteria           StringCriteria `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewSkillPointBonus creates a new SkillPointBonus.
func NewSkillPointBonus() *SkillPointBonus {
	return &SkillPointBonus{
		Type: feature.SkillPointBonus,
		NameCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: IsString,
			},
		},
		SpecializationCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		TagsCriteria: StringCriteria{
			StringCriteriaData: StringCriteriaData{
				Compare: AnyString,
			},
		},
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
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
	ExtractNameables(s.NameCriteria.Qualifier, m, existing)
	ExtractNameables(s.SpecializationCriteria.Qualifier, m, existing)
	ExtractNameables(s.TagsCriteria.Qualifier, m, existing)
}

// SetLevel implements Bonus.
func (s *SkillPointBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SkillPointBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(s.parentName())
		buffer.WriteString(" [")
		buffer.WriteString(s.LeveledAmount.Format(false))
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
		return
	}
	_ = binary.Write(h, binary.LittleEndian, s.Type)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}
