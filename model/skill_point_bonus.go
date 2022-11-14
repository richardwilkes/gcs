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

package model

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Bonus = &SkillPointBonus{}

// SkillPointBonus holds an adjustment to a skill's points.
type SkillPointBonus struct {
	Type                   FeatureType     `json:"type"`
	NameCriteria           criteria.String `json:"name,omitempty"`
	SpecializationCriteria criteria.String `json:"specialization,omitempty"`
	TagsCriteria           criteria.String `json:"tags,alt=category,omitempty"`
	LeveledAmount
	owner fmt.Stringer
}

// NewSkillPointBonus creates a new SkillPointBonus.
func NewSkillPointBonus() *SkillPointBonus {
	return &SkillPointBonus{
		Type: SkillPointBonusFeatureType,
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
func (s *SkillPointBonus) FeatureType() FeatureType {
	return s.Type
}

// Clone implements Feature.
func (s *SkillPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FillWithNameableKeys implements Feature.
func (s *SkillPointBonus) FillWithNameableKeys(m map[string]string) {
	Extract(s.NameCriteria.Qualifier, m)
	Extract(s.SpecializationCriteria.Qualifier, m)
	Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SkillPointBonus) ApplyNameableKeys(m map[string]string) {
	s.NameCriteria.Qualifier = Apply(s.NameCriteria.Qualifier, m)
	s.SpecializationCriteria.Qualifier = Apply(s.SpecializationCriteria.Qualifier, m)
	s.TagsCriteria.Qualifier = Apply(s.TagsCriteria.Qualifier, m)
}

// Owner implements Bonus.
func (s *SkillPointBonus) Owner() fmt.Stringer {
	return s.owner
}

// SetOwner implements Bonus.
func (s *SkillPointBonus) SetOwner(owner fmt.Stringer) {
	s.owner = owner
}

// SetLevel implements Bonus.
func (s *SkillPointBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SkillPointBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(parentName(s.owner))
		buffer.WriteString(" [")
		buffer.WriteString(s.LeveledAmount.FormatWithLevel(false))
		if s.AdjustedAmount() == fxp.One {
			buffer.WriteString(i18n.Text(" pt]"))
		} else {
			buffer.WriteString(i18n.Text(" pts]"))
		}
	}
}
