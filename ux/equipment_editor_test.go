// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newEditorEquipment creates a piece of equipment carrying a single non-container modifier, then returns both it and
// an editor data overlay copied from it, mirroring what displayEditor() hands to the equipment editor.
func newEditorEquipment(mutate func(e *gurps.Equipment, mod *gurps.EquipmentModifier)) (*gurps.Equipment, *gurps.EquipmentEditData) {
	equipment := gurps.NewEquipment(nil, nil, false)
	equipment.Name = "Test Item"
	equipment.Quantity = fxp.One
	equipment.Level = fxp.One
	mod := gurps.NewEquipmentModifier(nil, nil, false)
	mod.Name = "Test Modifier"
	mutate(equipment, mod)
	equipment.Modifiers = []*gurps.EquipmentModifier{mod}
	var data gurps.EquipmentEditData
	data.CopyFrom(equipment)
	return equipment, &data
}

// TestExtendedValuePreviewUsesPendingLevel verifies that the Extended Value preview honors the level currently typed
// into the editor rather than the level the equipment was opened with. The preview's modifier context used to be the
// unedited target, so a "per level" cost modifier kept multiplying by the original level and the preview never moved
// when the Level field changed -- then the sheet showed a different value than the editor had just displayed.
func TestExtendedValuePreviewUsesPendingLevel(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipment(func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
		e.BaseValue = "100"
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
		mod.CostIsPerLevel = true
	})

	// At level 1 the modifier adds 10 to the base value of 100.
	c.Equal(fxp.FromInteger(110), extendedValueForEditor(equipment, data), "preview matches the unedited state")

	// Raising the level in the editor must immediately scale the per-level modifier, even though the target still
	// holds the original level.
	data.Level = fxp.FromInteger(3)
	c.Equal(fxp.FromInteger(130), extendedValueForEditor(equipment, data), "preview follows the edited level")
	c.Equal(fxp.One, equipment.Level, "the target is left untouched by the preview")

	// The preview must agree with what the sheet will show once the edits are applied.
	data.ApplyTo(equipment)
	c.Equal(fxp.FromInteger(130), equipment.ExtendedValue(), "the applied value matches the preview")
}

// TestExtendedValuePreviewUsesPendingWeight verifies that a "per pound" cost modifier in the Extended Value preview
// sees the weight currently typed into the editor. The multiplier comes from the equipment's own weight, so passing
// the unedited target left the preview stuck on the original weight.
func TestExtendedValuePreviewUsesPendingWeight(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipment(func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
		e.BaseValue = "100"
		e.BaseWeight = "2 lb"
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
		mod.CostIsPerPound = true
	})

	// At 2 lb the modifier adds 10 per pound, so 20 on top of the base value of 100.
	c.Equal(fxp.FromInteger(120), extendedValueForEditor(equipment, data), "preview matches the unedited state")

	data.BaseWeight = "5 lb"
	c.Equal(fxp.FromInteger(150), extendedValueForEditor(equipment, data), "preview follows the edited weight")

	data.ApplyTo(equipment)
	c.Equal(fxp.FromInteger(150), equipment.ExtendedValue(), "the applied value matches the preview")
}

// TestExtendedWeightPreviewUsesPendingLevel verifies the same fix for the Extended Weight preview, whose "per level"
// weight modifiers were likewise multiplied by the target's original level instead of the editor's pending one.
func TestExtendedWeightPreviewUsesPendingLevel(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipment(func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
		e.BaseWeight = "2 lb"
		mod.WeightType = emweight.Original
		mod.WeightAmount = "+1 lb"
		mod.WeightIsPerLevel = true
	})

	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), extendedWeightForEditor(equipment, data, fxp.Pound),
		"preview matches the unedited state")

	data.Level = fxp.FromInteger(3)
	c.Equal(fxp.WeightFromInteger(5, fxp.Pound), extendedWeightForEditor(equipment, data, fxp.Pound),
		"preview follows the edited level")

	data.ApplyTo(equipment)
	c.Equal(fxp.WeightFromInteger(5, fxp.Pound), equipment.ExtendedWeight(false, fxp.Pound),
		"the applied weight matches the preview")
}

// TestExtendedPreviewsHonorModifierEnablement verifies that toggling a modifier off in the editor is reflected by both
// previews. The modifier list itself was already the editor's, but the equipment context it was evaluated against was
// not, so anything the multiplier derived from the equipment stayed stale.
func TestExtendedPreviewsHonorModifierEnablement(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipment(func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
		e.BaseValue = "100"
		e.BaseWeight = "2 lb"
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
		mod.CostIsPerPound = true
		mod.WeightType = emweight.Original
		mod.WeightAmount = "+1 lb"
		mod.WeightIsPerLevel = true
	})

	// The modifier adds a pound, so the per-pound cost multiplier is 3, not 2.
	c.Equal(fxp.FromInteger(130), extendedValueForEditor(equipment, data), "value preview with the modifier enabled")
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), extendedWeightForEditor(equipment, data, fxp.Pound),
		"weight preview with the modifier enabled")

	data.Modifiers[0].Disabled = true
	c.Equal(fxp.FromInteger(100), extendedValueForEditor(equipment, data), "value preview with the modifier disabled")
	c.Equal(fxp.WeightFromInteger(2, fxp.Pound), extendedWeightForEditor(equipment, data, fxp.Pound),
		"weight preview with the modifier disabled")
}

// TestExtendedPreviewsWithZeroQuantity verifies that a zero quantity yields zero for both previews, matching what the
// sheet reports for such a row.
func TestExtendedPreviewsWithZeroQuantity(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipment(func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
		e.BaseValue = "100"
		e.BaseWeight = "2 lb"
		mod.CostType = emcost.Original
		mod.CostAmount = "+10"
	})

	data.Quantity = 0
	c.Equal(fxp.Int(0), extendedValueForEditor(equipment, data), "zero quantity yields no value")
	c.Equal(fxp.Weight(0), extendedWeightForEditor(equipment, data, fxp.Pound), "zero quantity yields no weight")
}
