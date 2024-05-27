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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/study"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*Spell] = &SpellEditData{}

// SpellEditData holds the Spell data that can be edited by the UI detail editor.
type SpellEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	SpellNonContainerOnlyEditData
	SkillContainerOnlyEditData
}

// SpellNonContainerOnlyEditData holds the Spell data that is only applicable to spells that aren't containers.
type SpellNonContainerOnlyEditData struct {
	TechLevel         *string             `json:"tech_level,omitempty"`
	Difficulty        AttributeDifficulty `json:"difficulty,omitempty"`
	College           CollegeList         `json:"college,omitempty"`
	PowerSource       string              `json:"power_source,omitempty"`
	Class             string              `json:"spell_class,omitempty"`
	Resist            string              `json:"resist,omitempty"`
	CastingCost       string              `json:"casting_cost,omitempty"`
	MaintenanceCost   string              `json:"maintenance_cost,omitempty"`
	CastingTime       string              `json:"casting_time,omitempty"`
	Duration          string              `json:"duration,omitempty"`
	RitualSkillName   string              `json:"base_skill,omitempty"`
	RitualPrereqCount int                 `json:"prereq_count,omitempty"`
	Points            fxp.Int             `json:"points,omitempty"`
	Prereq            *PrereqList         `json:"prereqs,omitempty"`
	Weapons           []*Weapon           `json:"weapons,omitempty"`
	Study             []*Study            `json:"study,omitempty"`
	StudyHoursNeeded  study.Level         `json:"study_hours_needed,omitempty"`
}

// CopyFrom implements node.EditorData.
func (d *SpellEditData) CopyFrom(s *Spell) {
	d.copyFrom(s.Entity, &s.SpellEditData, s.Container(), false)
}

// ApplyTo implements node.EditorData.
func (d *SpellEditData) ApplyTo(s *Spell) {
	s.SpellEditData.copyFrom(s.Entity, d, s.Container(), true)
}

func (d *SpellEditData) copyFrom(entity *Entity, other *SpellEditData, isContainer, isApply bool) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	if other.TechLevel != nil {
		tl := *other.TechLevel
		d.TechLevel = &tl
	}
	d.College = txt.CloneStringSlice(d.College)
	d.Prereq = d.Prereq.CloneResolvingEmpty(isContainer, isApply)
	d.Weapons = nil
	if len(other.Weapons) != 0 {
		d.Weapons = make([]*Weapon, len(other.Weapons))
		for i := range other.Weapons {
			d.Weapons[i] = other.Weapons[i].Clone(entity, nil, true)
		}
	}
	if len(other.Study) != 0 {
		d.Study = make([]*Study, len(other.Study))
		for i := range other.Study {
			d.Study[i] = other.Study[i].Clone()
		}
	}
	d.TemplatePicker = d.TemplatePicker.Clone()
}
