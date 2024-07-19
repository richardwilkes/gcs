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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SkillBonus{}

// SkillBonus holds an adjustment to a skill.
type SkillBonus struct {
	Type                   feature.Type   `json:"type"`
	SelectionType          skillsel.Type  `json:"selection_type"`
	NameCriteria           StringCriteria `json:"name,omitempty"`
	SpecializationCriteria StringCriteria `json:"specialization,omitempty"`
	TagsCriteria           StringCriteria `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
}

// NewSkillBonus creates a new SkillBonus.
func NewSkillBonus() *SkillBonus {
	return &SkillBonus{
		Type:          feature.SkillBonus,
		SelectionType: skillsel.Name,
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
func (s *SkillBonus) FeatureType() feature.Type {
	return s.Type
}

// Clone implements Feature.
func (s *SkillBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SkillBonus) FillWithNameableKeys(m map[string]string) {
	ExtractNameables(s.SpecializationCriteria.Qualifier, m)
	if s.SelectionType != skillsel.ThisWeapon {
		ExtractNameables(s.NameCriteria.Qualifier, m)
		ExtractNameables(s.TagsCriteria.Qualifier, m)
	}
}

// SetLevel implements Bonus.
func (s *SkillBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SkillBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	s.basicAddToTooltip(&s.LeveledAmount, buffer)
}

// Hash writes this object's contents into the hasher.
func (s *SkillBonus) Hash(h hash.Hash) {
	if s == nil {
		return
	}
	_ = binary.Write(h, binary.LittleEndian, s.Type)
	_ = binary.Write(h, binary.LittleEndian, s.SelectionType)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}
