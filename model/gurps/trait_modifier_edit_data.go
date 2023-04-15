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
	"github.com/richardwilkes/toolbox/txt"
)

var _ EditorData[*TraitModifier] = &TraitModifierEditData{}

// TraitModifierEditData holds the TraitModifier data that can be edited by the UI detail editor.
type TraitModifierEditData struct {
	Name       string                `json:"name,omitempty"`
	PageRef    string                `json:"reference,omitempty"`
	LocalNotes string                `json:"notes,omitempty"`
	VTTNotes   string                `json:"vtt_notes,omitempty"`
	Tags       []string              `json:"tags,omitempty"`
	Cost       fxp.Int               `json:"cost,omitempty"`      // Non-container only
	Levels     fxp.Int               `json:"levels,omitempty"`    // Non-container only
	Affects    Affects               `json:"affects,omitempty"`   // Non-container only
	CostType   TraitModifierCostType `json:"cost_type,omitempty"` // Non-container only
	Disabled   bool                  `json:"disabled,omitempty"`  // Non-container only
	Features   Features              `json:"features,omitempty"`  // Non-container only
}

// CopyFrom implements node.EditorData.
func (d *TraitModifierEditData) CopyFrom(mod *TraitModifier) {
	d.copyFrom(&mod.TraitModifierEditData)
}

// ApplyTo implements node.EditorData.
func (d *TraitModifierEditData) ApplyTo(mod *TraitModifier) {
	mod.TraitModifierEditData.copyFrom(d)
}

func (d *TraitModifierEditData) copyFrom(other *TraitModifierEditData) {
	*d = *other
	d.Tags = txt.CloneStringSlice(d.Tags)
	d.Features = other.Features.Clone()
}
