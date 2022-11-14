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
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var _ Prereq = &SkillPrereq{}

// SkillPrereq holds a prerequisite for a skill.
type SkillPrereq struct {
	Parent                 *PrereqList      `json:"-"`
	Type                   PrereqType       `json:"type"`
	Has                    bool             `json:"has"`
	NameCriteria           criteria.String  `json:"name,omitempty"`
	LevelCriteria          criteria.Numeric `json:"level,omitempty"`
	SpecializationCriteria criteria.String  `json:"specialization,omitempty"`
}

// NewSkillPrereq creates a new SkillPrereq.
func NewSkillPrereq() *SkillPrereq {
	return &SkillPrereq{
		Type: SkillPrereqType,
		NameCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Is,
			},
		},
		LevelCriteria: criteria.Numeric{
			NumericData: criteria.NumericData{
				Compare: criteria.AtLeast,
			},
		},
		SpecializationCriteria: criteria.String{
			StringData: criteria.StringData{
				Compare: criteria.Any,
			},
		},
		Has: true,
	}
}

// PrereqType implements Prereq.
func (s *SkillPrereq) PrereqType() PrereqType {
	return s.Type
}

// ParentList implements Prereq.
func (s *SkillPrereq) ParentList() *PrereqList {
	return s.Parent
}

// Clone implements Prereq.
func (s *SkillPrereq) Clone(parent *PrereqList) Prereq {
	clone := *s
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (s *SkillPrereq) FillWithNameableKeys(m map[string]string) {
	Extract(s.NameCriteria.Qualifier, m)
	Extract(s.SpecializationCriteria.Qualifier, m)
}

// ApplyNameableKeys implements Prereq.
func (s *SkillPrereq) ApplyNameableKeys(m map[string]string) {
	s.NameCriteria.Qualifier = Apply(s.NameCriteria.Qualifier, m)
	s.SpecializationCriteria.Qualifier = Apply(s.SpecializationCriteria.Qualifier, m)
}

// Satisfied implements Prereq.
func (s *SkillPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	satisfied := false
	var techLevel *string
	if sk, ok := exclude.(*Skill); ok {
		techLevel = sk.TechLevel
	}
	Traverse(func(sk *Skill) bool {
		if exclude == sk || !s.NameCriteria.Matches(sk.Name) || !s.SpecializationCriteria.Matches(sk.Specialization) {
			return false
		}
		satisfied = s.LevelCriteria.Matches(sk.LevelData.Level)
		if satisfied && techLevel != nil {
			satisfied = sk.TechLevel == nil || *techLevel == *sk.TechLevel
		}
		return satisfied
	}, false, true, entity.Skills...)
	if !s.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(s.Has))
		tooltip.WriteString(i18n.Text(" a skill whose name "))
		tooltip.WriteString(s.NameCriteria.String())
		if s.SpecializationCriteria.Compare != criteria.Any {
			tooltip.WriteString(i18n.Text(", specialization "))
			tooltip.WriteString(s.SpecializationCriteria.String())
			tooltip.WriteByte(',')
		}
		if techLevel == nil {
			tooltip.WriteString(i18n.Text(" and level "))
			tooltip.WriteString(s.LevelCriteria.String())
		} else {
			if s.SpecializationCriteria.Compare != criteria.Any {
				tooltip.WriteByte(',')
			}
			tooltip.WriteString(i18n.Text(" level "))
			tooltip.WriteString(s.LevelCriteria.String())
			tooltip.WriteString(i18n.Text(" and tech level matches"))
		}
	}
	return satisfied
}
