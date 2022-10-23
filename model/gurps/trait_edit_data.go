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
	"github.com/richardwilkes/gcs/v5/model/gurps/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/trait"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*Trait] = &TraitEditData{}

// TraitEditData holds the Trait data that can be edited by the UI detail editor.
type TraitEditData struct {
	Name           string                `json:"name,omitempty"`
	PageRef        string                `json:"reference,omitempty"`
	LocalNotes     string                `json:"notes,omitempty"`
	VTTNotes       string                `json:"vtt_notes,omitempty"`
	Ancestry       string                `json:"ancestry,omitempty"` // Container only
	UserDesc       string                `json:"userdesc,omitempty"`
	Tags           []string              `json:"tags,omitempty"`
	Modifiers      []*TraitModifier      `json:"modifiers,omitempty"`
	BasePoints     fxp.Int               `json:"base_points,omitempty"`      // Non-container only
	Levels         fxp.Int               `json:"levels,omitempty"`           // Non-container only
	PointsPerLevel fxp.Int               `json:"points_per_level,omitempty"` // Non-container only
	Prereq         *PrereqList           `json:"prereqs,omitempty"`          // Non-container only
	Weapons        []*Weapon             `json:"weapons,omitempty"`          // Non-container only
	Features       feature.Features      `json:"features,omitempty"`         // Non-container only
	Study          []*Study              `json:"study,omitempty"`            // Non-container only
	TemplatePicker *TemplatePicker       `json:"template_picker,omitempty"`  // Container only
	CR             trait.SelfControlRoll `json:"cr,omitempty"`
	CRAdj          SelfControlRollAdj    `json:"cr_adj,omitempty"`
	ContainerType  trait.ContainerType   `json:"container_type,omitempty"` // Container only
	Disabled       bool                  `json:"disabled,omitempty"`
	RoundCostDown  bool                  `json:"round_down,omitempty"` // Non-container only
	CanLevel       bool                  `json:"can_level,omitempty"`  // Non-container only
}

// CopyFrom implements node.EditorData.
func (d *TraitEditData) CopyFrom(t *Trait) {
	d.copyFrom(t.Entity, &t.TraitEditData, t.Container(), false)
}

// ApplyTo implements node.EditorData.
func (d *TraitEditData) ApplyTo(t *Trait) {
	t.TraitEditData.copyFrom(t.Entity, d, t.Container(), true)
}

func (d *TraitEditData) copyFrom(entity *Entity, other *TraitEditData, isContainer, isApply bool) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Modifiers = nil
	if len(other.Modifiers) != 0 {
		d.Modifiers = make([]*TraitModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			d.Modifiers = append(d.Modifiers, one.Clone(entity, nil, true))
		}
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
	d.TemplatePicker = d.TemplatePicker.Clone()
}
