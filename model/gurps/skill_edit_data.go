/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*Skill] = &SkillEditData{}

// SkillEditData holds the Skill data that can be edited by the UI detail editor.
type SkillEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	SkillNonContainerOnlyEditData
	SkillContainerOnlyEditData
}

// SkillNonContainerOnlyEditData holds the Skill data that is only applicable to skills that aren't containers.
type SkillNonContainerOnlyEditData struct {
	Specialization               string              `json:"specialization,omitempty"`
	TechLevel                    *string             `json:"tech_level,omitempty"`
	Difficulty                   AttributeDifficulty `json:"difficulty,omitempty"`
	Points                       fxp.Int             `json:"points,omitempty"`
	EncumbrancePenaltyMultiplier fxp.Int             `json:"encumbrance_penalty_multiplier,omitempty"`
	DefaultedFrom                *SkillDefault       `json:"defaulted_from,omitempty"`
	Defaults                     []*SkillDefault     `json:"defaults,omitempty"`
	TechniqueDefault             *SkillDefault       `json:"default,omitempty"`
	TechniqueLimitModifier       *fxp.Int            `json:"limit,omitempty"`
	Prereq                       *PrereqList         `json:"prereqs,omitempty"`
	Weapons                      []*Weapon           `json:"weapons,omitempty"`
	Features                     Features            `json:"features,omitempty"`
	Study                        []*Study            `json:"study,omitempty"`
	StudyHoursNeeded             study.Level         `json:"study_hours_needed,omitempty"`
}

// SkillContainerOnlyEditData holds the Skill data that is only applicable to skills that are containers.
type SkillContainerOnlyEditData struct {
	TemplatePicker *TemplatePicker `json:"template_picker,omitempty"`
}

// CopyFrom implements node.EditorData.
func (d *SkillEditData) CopyFrom(s *Skill) {
	d.copyFrom(s.Entity, &s.SkillEditData, s.Container(), false)
}

// ApplyTo implements node.EditorData.
func (d *SkillEditData) ApplyTo(s *Skill) {
	s.SkillEditData.copyFrom(s.Entity, d, s.Container(), true)
}

func (d *SkillEditData) copyFrom(entity *Entity, other *SkillEditData, isContainer, isApply bool) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	if other.TechLevel != nil {
		tl := *other.TechLevel
		d.TechLevel = &tl
	}
	if other.DefaultedFrom != nil {
		def := *other.DefaultedFrom
		d.DefaultedFrom = &def
	}
	d.Defaults = nil
	if len(other.Defaults) != 0 {
		d.Defaults = make([]*SkillDefault, len(other.Defaults))
		for i, def := range other.Defaults {
			def2 := *def
			d.Defaults[i] = &def2
		}
	}
	if other.TechniqueDefault != nil {
		def := *other.TechniqueDefault
		d.TechniqueDefault = &def
		if !DefaultTypeIsSkillBased(other.TechniqueDefault.DefaultType) {
			d.TechniqueDefault.Name = ""
			d.TechniqueDefault.Specialization = ""
		}
	}
	if other.TechniqueLimitModifier != nil {
		mod := *other.TechniqueLimitModifier
		d.TechniqueLimitModifier = &mod
	}
	d.Prereq = d.Prereq.CloneResolvingEmpty(isContainer, isApply)
	d.Weapons = nil
	if len(other.Weapons) != 0 {
		d.Weapons = make([]*Weapon, len(other.Weapons))
		for i := range other.Weapons {
			d.Weapons[i] = other.Weapons[i].Clone(entity, nil, true)
		}
	}
	d.Features = other.Features.Clone()
	if len(other.Study) != 0 {
		d.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			d.Study[i] = other.Study[i].Clone()
		}
	}
	d.TemplatePicker = other.TemplatePicker.Clone()
}
