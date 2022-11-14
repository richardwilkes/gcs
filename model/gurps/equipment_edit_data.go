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
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*Equipment] = &EquipmentEditData{}

// EquipmentEditData holds the Equipment data that can be edited by the UI detail editor.
type EquipmentEditData struct {
	Name                   string               `json:"description,omitempty"`
	PageRef                string               `json:"reference,omitempty"`
	LocalNotes             string               `json:"notes,omitempty"`
	VTTNotes               string               `json:"vtt_notes,omitempty"`
	TechLevel              string               `json:"tech_level,omitempty"`
	LegalityClass          string               `json:"legality_class,omitempty"`
	Tags                   []string             `json:"tags,omitempty"`
	Modifiers              []*EquipmentModifier `json:"modifiers,omitempty"`
	Quantity               fxp.Int              `json:"quantity,omitempty"`
	Value                  fxp.Int              `json:"value,omitempty"`
	Weight                 measure.Weight       `json:"weight,omitempty"`
	MaxUses                int                  `json:"max_uses,omitempty"`
	Uses                   int                  `json:"uses,omitempty"`
	Prereq                 *PrereqList          `json:"prereqs,omitempty"`
	Weapons                []*Weapon            `json:"weapons,omitempty"`
	Features               Features             `json:"features,omitempty"`
	Equipped               bool                 `json:"equipped,omitempty"`
	WeightIgnoredForSkills bool                 `json:"ignore_weight_for_skills,omitempty"`
}

// CopyFrom implements node.EditorData.
func (d *EquipmentEditData) CopyFrom(e *Equipment) {
	d.copyFrom(e.Entity, &e.EquipmentEditData, false)
}

// ApplyTo implements node.EditorData.
func (d *EquipmentEditData) ApplyTo(e *Equipment) {
	e.EquipmentEditData.copyFrom(e.Entity, d, true)
}

func (d *EquipmentEditData) copyFrom(entity *Entity, other *EquipmentEditData, isApply bool) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Modifiers = nil
	if len(other.Modifiers) != 0 {
		d.Modifiers = make([]*EquipmentModifier, 0, len(other.Modifiers))
		for _, one := range other.Modifiers {
			d.Modifiers = append(d.Modifiers, one.Clone(entity, nil, true))
		}
	}
	d.Prereq = d.Prereq.CloneResolvingEmpty(false, isApply)
	d.Weapons = nil
	if len(other.Weapons) != 0 {
		d.Weapons = make([]*Weapon, 0, len(other.Weapons))
		for _, one := range other.Weapons {
			d.Weapons = append(d.Weapons, one.Clone(entity, nil, true))
		}
	}
	d.Features = other.Features.Clone()
}
