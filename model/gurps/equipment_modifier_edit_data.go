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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*EquipmentModifier] = &EquipmentModifierEditData{}

// EquipmentModifierEditData holds the EquipmentModifier data that can be edited by the UI detail editor.
type EquipmentModifierEditData struct {
	Name             string   `json:"name,omitempty"`
	PageRef          string   `json:"reference,omitempty"`
	PageRefHighlight string   `json:"reference_highlight,omitempty"`
	LocalNotes       string   `json:"notes,omitempty"`
	VTTNotes         string   `json:"vtt_notes,omitempty"`
	Tags             []string `json:"tags,omitempty"`
	EquipmentModifierEditDataNonContainerOnly
}

// EquipmentModifierEditDataNonContainerOnly holds the EquipmentModifier data that is only applicable to
// EquipmentModifiers that aren't containers.
type EquipmentModifierEditDataNonContainerOnly struct {
	CostType     emcost.Type   `json:"cost_type,omitempty"`
	WeightType   emweight.Type `json:"weight_type,omitempty"`
	Disabled     bool          `json:"disabled,omitempty"`
	TechLevel    string        `json:"tech_level,omitempty"`
	CostAmount   string        `json:"cost,omitempty"`
	WeightAmount string        `json:"weight,omitempty"`
	Features     Features      `json:"features,omitempty"`
}

// CopyFrom implements node.EditorData.
func (d *EquipmentModifierEditData) CopyFrom(mod *EquipmentModifier) {
	d.copyFrom(&mod.EquipmentModifierEditData)
}

// ApplyTo implements node.EditorData.
func (d *EquipmentModifierEditData) ApplyTo(mod *EquipmentModifier) {
	mod.EquipmentModifierEditData.copyFrom(d)
}

func (d *EquipmentModifierEditData) copyFrom(other *EquipmentModifierEditData) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Features = other.Features.Clone()
}
