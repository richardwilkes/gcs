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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*Spell] = &SpellEditData{}

// SpellEditData holds the Spell data that can be edited by the UI detail editor.
type SpellEditData struct {
	Name              string              `json:"name,omitempty"`
	PageRef           string              `json:"reference,omitempty"`
	LocalNotes        string              `json:"notes,omitempty"`
	VTTNotes          string              `json:"vtt_notes,omitempty"`
	Tags              []string            `json:"tags,omitempty"`
	TechLevel         *string             `json:"tech_level,omitempty"`       // Non-container only
	Difficulty        AttributeDifficulty `json:"difficulty,omitempty"`       // Non-container only
	College           CollegeList         `json:"college,omitempty"`          // Non-container only
	PowerSource       string              `json:"power_source,omitempty"`     // Non-container only
	Class             string              `json:"spell_class,omitempty"`      // Non-container only
	Resist            string              `json:"resist,omitempty"`           // Non-container only
	CastingCost       string              `json:"casting_cost,omitempty"`     // Non-container only
	MaintenanceCost   string              `json:"maintenance_cost,omitempty"` // Non-container only
	CastingTime       string              `json:"casting_time,omitempty"`     // Non-container only
	Duration          string              `json:"duration,omitempty"`         // Non-container only
	RitualSkillName   string              `json:"base_skill,omitempty"`       // Non-container only
	RitualPrereqCount int                 `json:"prereq_count,omitempty"`     // Non-container only
	Points            fxp.Int             `json:"points,omitempty"`           // Non-container only
	Prereq            *PrereqList         `json:"prereqs,omitempty"`          // Non-container only
	Weapons           []*Weapon           `json:"weapons,omitempty"`          // Non-container only
	TemplatePicker    *TemplatePicker     `json:"template_picker,omitempty"`  // Container only
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
		d.Weapons = make([]*Weapon, 0, len(other.Weapons))
		for _, one := range other.Weapons {
			d.Weapons = append(d.Weapons, one.Clone(entity, nil, true))
		}
	}
	d.TemplatePicker = d.TemplatePicker.Clone()
}
