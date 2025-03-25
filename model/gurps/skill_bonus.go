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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/skillsel"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Bonus = &SkillBonus{}

// SkillBonus holds an adjustment to a skill.
type SkillBonus struct {
	Type                   feature.Type  `json:"type"`
	SelectionType          skillsel.Type `json:"selection_type"`
	NameCriteria           criteria.Text `json:"name,omitempty"`
	SpecializationCriteria criteria.Text `json:"specialization,omitempty"`
	TagsCriteria           criteria.Text `json:"tags,alt=category,omitempty"`
	LeveledAmount
	BonusOwner
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
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, s.Type)
	hashhelper.Num8(h, s.SelectionType)
	s.NameCriteria.Hash(h)
	s.SpecializationCriteria.Hash(h)
	s.TagsCriteria.Hash(h)
	s.LeveledAmount.Hash(h)
}
