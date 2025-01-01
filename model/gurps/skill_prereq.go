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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var _ Prereq = &SkillPrereq{}

// SkillPrereq holds a prerequisite for a skill.
type SkillPrereq struct {
	Parent                 *PrereqList     `json:"-"`
	Type                   prereq.Type     `json:"type"`
	Has                    bool            `json:"has"`
	NameCriteria           criteria.Text   `json:"name,omitempty"`
	LevelCriteria          criteria.Number `json:"level,omitempty"`
	SpecializationCriteria criteria.Text   `json:"specialization,omitempty"`
}

// NewSkillPrereq creates a new SkillPrereq.
func NewSkillPrereq() *SkillPrereq {
	var p SkillPrereq
	p.Type = prereq.Skill
	p.NameCriteria.Compare = criteria.IsText
	p.LevelCriteria.Compare = criteria.AtLeastNumber
	p.SpecializationCriteria.Compare = criteria.AnyText
	p.Has = true
	return &p
}

// PrereqType implements Prereq.
func (p *SkillPrereq) PrereqType() prereq.Type {
	return p.Type
}

// ParentList implements Prereq.
func (p *SkillPrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *SkillPrereq) Clone(parent *PrereqList) Prereq {
	clone := *p
	clone.Parent = parent
	return &clone
}

// FillWithNameableKeys implements Prereq.
func (p *SkillPrereq) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(p.NameCriteria.Qualifier, m, existing)
	nameable.Extract(p.SpecializationCriteria.Qualifier, m, existing)
}

// Satisfied implements Prereq.
func (p *SkillPrereq) Satisfied(entity *Entity, exclude any, tooltip *xio.ByteBuffer, prefix string, _ *bool) bool {
	var replacements map[string]string
	if na, ok := exclude.(nameable.Accesser); ok {
		replacements = na.NameableReplacements()
	}
	satisfied := false
	var techLevel *string
	if sk, ok := exclude.(*Skill); ok {
		techLevel = sk.TechLevel
	}
	Traverse(func(sk *Skill) bool {
		if exclude == sk || !p.NameCriteria.Matches(replacements, sk.NameWithReplacements()) ||
			!p.SpecializationCriteria.Matches(replacements, sk.SpecializationWithReplacements()) {
			return false
		}
		satisfied = p.LevelCriteria.Matches(sk.LevelData.Level)
		if satisfied && techLevel != nil {
			satisfied = sk.TechLevel == nil || *techLevel == *sk.TechLevel
		}
		return satisfied
	}, false, true, entity.Skills...)
	if !p.Has {
		satisfied = !satisfied
	}
	if !satisfied && tooltip != nil {
		tooltip.WriteString(prefix)
		tooltip.WriteString(HasText(p.Has))
		tooltip.WriteString(i18n.Text(" a skill whose name "))
		tooltip.WriteString(p.NameCriteria.String(replacements))
		if p.SpecializationCriteria.Compare != criteria.AnyText {
			tooltip.WriteString(i18n.Text(", specialization "))
			tooltip.WriteString(p.SpecializationCriteria.String(replacements))
			tooltip.WriteByte(',')
		}
		if techLevel == nil {
			tooltip.WriteString(i18n.Text(" and level "))
			tooltip.WriteString(p.LevelCriteria.String())
		} else {
			if p.SpecializationCriteria.Compare != criteria.AnyText {
				tooltip.WriteByte(',')
			}
			tooltip.WriteString(i18n.Text(" level "))
			tooltip.WriteString(p.LevelCriteria.String())
			tooltip.WriteString(i18n.Text(" and tech level matches"))
		}
	}
	return satisfied
}

// Hash writes this object's contents into the hasher.
func (p *SkillPrereq) Hash(h hash.Hash) {
	if p == nil {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num8(h, p.Type)
	hashhelper.Bool(h, p.Has)
	p.NameCriteria.Hash(h)
	p.LevelCriteria.Hash(h)
	p.SpecializationCriteria.Hash(h)
}
