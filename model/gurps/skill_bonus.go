/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SkillBonus{}

// SkillBonus holds an adjustment to a skill.
type SkillBonus struct {
	Type                   FeatureType        `json:"type"`
	SelectionType          SkillSelectionType `json:"selection_type"`
	NameCriteria           criteria.String    `json:"name,omitempty"`
	SpecializationCriteria criteria.String    `json:"specialization,omitempty"`
	TagsCriteria           criteria.String    `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewSkillBonus creates a new SkillBonus.
func NewSkillBonus() *SkillBonus {
	return &SkillBonus{
		Type:          SkillBonusFeatureType,
		SelectionType: NameSkillSelectionType,
		NameCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Is,
			},
		},
		SpecializationCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Any,
			},
		},
		TagsCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Any,
			},
		},
		LeveledAmount: LeveledAmount{Amount: fxp.One},
	}
}

// FeatureType implements Feature.
func (s *SkillBonus) FeatureType() FeatureType {
	return s.Type
}

// Clone implements Feature.
func (s *SkillBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SkillBonus) FillWithNameableKeys(m map[string]string) {
	Extract(s.SpecializationCriteria.Qualifier, m)
	if s.SelectionType != ThisWeaponSkillSelectionType {
		Extract(s.NameCriteria.Qualifier, m)
		Extract(s.TagsCriteria.Qualifier, m)
	}
}

// ApplyNameableKeys implements Feature.
func (s *SkillBonus) ApplyNameableKeys(m map[string]string) {
	s.SpecializationCriteria.Qualifier = Apply(s.SpecializationCriteria.Qualifier, m)
	if s.SelectionType != ThisWeaponSkillSelectionType {
		s.NameCriteria.Qualifier = Apply(s.NameCriteria.Qualifier, m)
		s.TagsCriteria.Qualifier = Apply(s.TagsCriteria.Qualifier, m)
	}
}

// Owner implements Bonus.
func (s *SkillBonus) Owner() fmt.Stringer {
	return s.owner
}

// SetOwner implements Bonus.
func (s *SkillBonus) SetOwner(owner fmt.Stringer) {
	s.owner = owner
}

// SetLevel implements Bonus.
func (s *SkillBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SkillBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	basicAddToTooltip(s.owner, &s.LeveledAmount, buffer)
}
