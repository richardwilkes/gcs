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

package feature

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/nameables"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

const (
	// SkillPointsID holds the ID for skill point lookups.
	SkillPointsID = "skill.points"
)

var _ Bonus = &SkillPointBonus{}

// SkillPointBonus holds an adjustment to a skill's points.
type SkillPointBonus struct {
	Type                   Type            `json:"type"`
	Parent                 fmt.Stringer    `json:"-"`
	NameCriteria           criteria.String `json:"name,omitempty"`
	SpecializationCriteria criteria.String `json:"specialization,omitempty"`
	TagsCriteria           criteria.String `json:"tags,alt=category,omitempty"`
	LeveledAmount
}

// NewSkillPointBonus creates a new SkillPointBonus.
func NewSkillPointBonus() *SkillPointBonus {
	return &SkillPointBonus{
		Type: SkillPointBonusType,
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
func (s *SkillPointBonus) FeatureType() Type {
	return s.Type
}

// Clone implements Feature.
func (s *SkillPointBonus) Clone() Feature {
	other := *s
	return &other
}

// FeatureMapKey implements Feature.
func (s *SkillPointBonus) FeatureMapKey() string {
	if s.NameCriteria.Compare == criteria.Is &&
		(s.SpecializationCriteria.Compare == criteria.Any && s.TagsCriteria.Compare == criteria.Any) {
		return SkillPointsID + "/" + s.NameCriteria.Qualifier
	}
	return SkillPointsID + "*"
}

// FillWithNameableKeys implements Feature.
func (s *SkillPointBonus) FillWithNameableKeys(m map[string]string) {
	nameables.Extract(s.NameCriteria.Qualifier, m)
	nameables.Extract(s.SpecializationCriteria.Qualifier, m)
	nameables.Extract(s.TagsCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Feature.
func (s *SkillPointBonus) ApplyNameableKeys(m map[string]string) {
	s.NameCriteria.Qualifier = nameables.Apply(s.NameCriteria.Qualifier, m)
	s.SpecializationCriteria.Qualifier = nameables.Apply(s.SpecializationCriteria.Qualifier, m)
	s.TagsCriteria.Qualifier = nameables.Apply(s.TagsCriteria.Qualifier, m)
}

// SetParent implements Bonus.
func (s *SkillPointBonus) SetParent(parent fmt.Stringer) {
	s.Parent = parent
}

// SetLevel implements Bonus.
func (s *SkillPointBonus) SetLevel(level fxp.Int) {
	s.Level = level
}

// AddToTooltip implements Bonus.
func (s *SkillPointBonus) AddToTooltip(buffer *xio.ByteBuffer) {
	if buffer != nil {
		buffer.WriteByte('\n')
		buffer.WriteString(parentName(s.Parent))
		buffer.WriteString(" [")
		buffer.WriteString(s.LeveledAmount.FormatWithLevel(false))
		if s.AdjustedAmount() == fxp.One {
			buffer.WriteString(i18n.Text(" pt]"))
		} else {
			buffer.WriteString(i18n.Text(" pts]"))
		}
	}
}
