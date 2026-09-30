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

// The Extended Value preview must honor the level currently typed into the editor rather than the level the equipment
// was opened with. The preview's modifier context used to be the unedited target, so a "per level" cost modifier kept
// multiplying by the original level and the preview never moved when the Level field changed.
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

// A "per pound" cost modifier in the Extended Value preview must see the weight currently typed into the editor. The
// multiplier comes from the equipment's own weight, so passing the unedited target left the preview stuck on the
// original weight.
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

// The same fix for the Extended Weight preview, whose "per level" weight modifiers were likewise multiplied by the
// target's original level instead of the editor's pending one.
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

// Toggling a modifier off in the editor must be reflected by both previews. The modifier list itself was already the
// editor's, but the equipment context it was evaluated against was not, so anything the multiplier derived from the
// equipment stayed stale.
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

// A zero quantity yields zero for both previews, matching what the sheet reports for such a row.
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

// TestExtendedPreviewsShowAChoicesRange verifies that the editor of a container holding a choice yet to be made
// previews the range the choice may come to, as the list does, while a container holding no choice previews a single
// value and weight.
func TestExtendedPreviewsShowAChoicesRange(t *testing.T) {
	c := check.New(t)
	backpack := gurps.NewEquipment(nil, nil, true)
	backpack.BaseValue = "5"
	backpack.BaseWeight = "2 lb"
	choice := gurps.NewEquipmentChoiceContainer(nil, backpack)
	for _, one := range []struct{ value, weight string }{{"10", "1 lb"}, {"30", "3 lb"}} {
		option := gurps.NewEquipment(nil, choice, false)
		option.BaseValue = one.value
		option.BaseWeight = one.weight
		choice.Children = append(choice.Children, option)
	}
	backpack.Children = []*gurps.Equipment{choice}
	var data gurps.EquipmentEditData
	data.CopyFrom(backpack)
	c.Equal("15~35", extendedValueTextForEditor(backpack, &data))
	c.Equal("3~5 lb", extendedWeightTextForEditor(backpack, &data, fxp.Pound))

	backpack.Children = nil
	data.CopyFrom(backpack)
	c.Equal("5", extendedValueTextForEditor(backpack, &data))
	c.Equal("2 lb", extendedWeightTextForEditor(backpack, &data, fxp.Pound))
}

// newEditorEquipmentWithChoice creates a $100, 2 lb piece of equipment outside a sheet whose only modifier is a
// mandatory choice yet to be made, with an option for each of the given cost and weight pairs, then returns both it and
// an editor data overlay copied from it.
func newEditorEquipmentWithChoice(options ...[2]string) (*gurps.Equipment, *gurps.EquipmentEditData) {
	equipment := gurps.NewEquipment(nil, nil, false)
	equipment.Quantity = fxp.One
	equipment.BaseValue = "100"
	equipment.BaseWeight = "2 lb"
	choice := gurps.NewEquipmentModifierChoice(nil, nil)
	for _, one := range options {
		option := gurps.NewEquipmentModifier(nil, choice, false)
		option.CostType = emcost.Original
		option.CostAmount = one[0]
		option.WeightType = emweight.Original
		option.WeightAmount = one[1]
		option.Disabled = true
		choice.Children = append(choice.Children, option)
	}
	equipment.Modifiers = []*gurps.EquipmentModifier{choice}
	var data gurps.EquipmentEditData
	data.CopyFrom(equipment)
	return equipment, &data
}

// TestExtendedPreviewsCountAnOpenChoice verifies that the editor of equipment with a mandatory modifier choice yet to
// be made previews what the list shows while its options all come to the same, and the range the choice may come to
// once they don't.
func TestExtendedPreviewsCountAnOpenChoice(t *testing.T) {
	c := check.New(t)
	equipment, data := newEditorEquipmentWithChoice([2]string{"+10", "+1 lb"}, [2]string{"+10", "+1 lb"})
	c.Equal(fxp.FromInteger(110), equipment.ExtendedValue(), "the list counts the choice")
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), equipment.ExtendedWeight(false, fxp.Pound), "the list weighs it")
	c.Equal("110", extendedValueTextForEditor(equipment, data))
	c.Equal("3 lb", extendedWeightTextForEditor(equipment, data, fxp.Pound))

	equipment, data = newEditorEquipmentWithChoice([2]string{"+10", "+1 lb"}, [2]string{"+30", "+3 lb"})
	c.Equal("110~130", extendedValueTextForEditor(equipment, data))
	c.Equal("3~5 lb", extendedWeightTextForEditor(equipment, data, fxp.Pound))
}

// TestExtendedPreviewsLeaveTheModifiersOnTheEditedItem verifies that working out the editor's previews, which cost a
// throwaway copy of the item, leaves the editor's modifiers, the pick of a choice among them included, modifying the
// item being edited.
func TestExtendedPreviewsLeaveTheModifiersOnTheEditedItem(t *testing.T) {
	c := check.New(t)
	equipment, _ := newEditorEquipmentWithChoice([2]string{"", "+1 lb"}, [2]string{"", "+1 lb"})
	equipment.Modifiers[0].Children[0].Disabled = false
	mod := gurps.NewEquipmentModifier(nil, nil, false)
	mod.CostType = emcost.Original
	mod.CostAmount = "+10"
	equipment.Modifiers = append([]*gurps.EquipmentModifier{mod}, equipment.Modifiers...)
	e, _ := buildEditorContent(nil, equipment, initEquipmentEditor(true))
	c.Equal(fxp.FromInteger(110), extendedValueForEditor(e.target, e.editorData))
	c.Equal(fxp.WeightFromInteger(3, fxp.Pound), extendedWeightForEditor(e.target, e.editorData, fxp.Pound))
	var count int
	gurps.Traverse(func(mod *gurps.EquipmentModifier) bool {
		count++
		c.True(mod.Target() == equipment, "%s still modifies the edited item", mod.Name)
		return false
	}, false, false, e.editorData.Modifiers...)
	c.Equal(4, count)
}
