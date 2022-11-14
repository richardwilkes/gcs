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
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*EquipmentModifier] = &EquipmentModifierEditData{}

// EquipmentModifierEditData holds the EquipmentModifier data that can be edited by the UI detail editor.
type EquipmentModifierEditData struct {
	Name         string                      `json:"name,omitempty"`
	PageRef      string                      `json:"reference,omitempty"`
	LocalNotes   string                      `json:"notes,omitempty"`
	VTTNotes     string                      `json:"vtt_notes,omitempty"`
	Tags         []string                    `json:"tags,omitempty"`
	CostType     EquipmentModifierCostType   `json:"cost_type,omitempty"`   // Non-container only
	WeightType   EquipmentModifierWeightType `json:"weight_type,omitempty"` // Non-container only
	Disabled     bool                        `json:"disabled,omitempty"`    // Non-container only
	TechLevel    string                      `json:"tech_level,omitempty"`  // Non-container only
	CostAmount   string                      `json:"cost,omitempty"`        // Non-container only
	WeightAmount string                      `json:"weight,omitempty"`      // Non-container only
	Features     Features                    `json:"features,omitempty"`    // Non-container only
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
