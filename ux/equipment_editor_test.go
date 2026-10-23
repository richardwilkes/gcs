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

// Both previews must follow what is typed into the editor -- the level a "per level" modifier and the weight a "per
// pound" one is multiplied by, the quantity, and whether a modifier is enabled -- rather than the unedited target,
// which they leave untouched, and must agree with what the sheet shows once the edits are applied.
func TestExtendedPreviewsHonorModifierEnablement(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		setup                func(e *gurps.Equipment, mod *gurps.EquipmentModifier)
		edit                 func(data *gurps.EquipmentEditData)
		value, editedValue   fxp.Int
		weight, editedWeight fxp.Weight
	}{
		{
			name: "per level cost follows the edited level",
			setup: func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
				e.BaseValue = "100"
				mod.CostType = emcost.Original
				mod.CostAmount = "+10"
				mod.CostIsPerLevel = true
			},
			edit:        func(data *gurps.EquipmentEditData) { data.Level = fxp.FromInteger(3) },
			value:       fxp.FromInteger(110),
			editedValue: fxp.FromInteger(130),
		},
		{
			name: "per pound cost follows the edited weight",
			setup: func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
				e.BaseValue = "100"
				e.BaseWeight = "2 lb"
				mod.CostType = emcost.Original
				mod.CostAmount = "+10"
				mod.CostIsPerPound = true
			},
			edit:         func(data *gurps.EquipmentEditData) { data.BaseWeight = "5 lb" },
			value:        fxp.FromInteger(120),
			editedValue:  fxp.FromInteger(150),
			weight:       fxp.WeightFromInteger(2, fxp.Pound),
			editedWeight: fxp.WeightFromInteger(5, fxp.Pound),
		},
		{
			name: "per level weight follows the edited level",
			setup: func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
				e.BaseWeight = "2 lb"
				mod.WeightType = emweight.Original
				mod.WeightAmount = "+1 lb"
				mod.WeightIsPerLevel = true
			},
			edit:         func(data *gurps.EquipmentEditData) { data.Level = fxp.FromInteger(3) },
			weight:       fxp.WeightFromInteger(3, fxp.Pound),
			editedWeight: fxp.WeightFromInteger(5, fxp.Pound),
		},
		{
			name: "a disabled modifier drops out",
			setup: func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
				e.BaseValue = "100"
				e.BaseWeight = "2 lb"
				mod.CostType = emcost.Original
				mod.CostAmount = "+10"
				mod.CostIsPerPound = true
				mod.WeightType = emweight.Original
				mod.WeightAmount = "+1 lb"
				mod.WeightIsPerLevel = true
			},
			edit: func(data *gurps.EquipmentEditData) { data.Modifiers[0].SetDisabled(true) },
			// The modifier adds a pound, so the per-pound cost multiplier is 3, not 2.
			value:        fxp.FromInteger(130),
			editedValue:  fxp.FromInteger(100),
			weight:       fxp.WeightFromInteger(3, fxp.Pound),
			editedWeight: fxp.WeightFromInteger(2, fxp.Pound),
		},
		{
			name: "a zero quantity yields nothing",
			setup: func(e *gurps.Equipment, mod *gurps.EquipmentModifier) {
				e.BaseValue = "100"
				e.BaseWeight = "2 lb"
				mod.CostType = emcost.Original
				mod.CostAmount = "+10"
			},
			edit:   func(data *gurps.EquipmentEditData) { data.Quantity = 0 },
			value:  fxp.FromInteger(110),
			weight: fxp.WeightFromInteger(2, fxp.Pound),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			equipment, data := newEditorEquipment(tc.setup)
			c.Equal(tc.value, extendedValueForEditor(equipment, data), "value preview matches the unedited state")
			c.Equal(tc.weight, extendedWeightForEditor(equipment, data, fxp.Pound),
				"weight preview matches the unedited state")

			tc.edit(data)
			c.Equal(tc.editedValue, extendedValueForEditor(equipment, data), "value preview follows the edit")
			c.Equal(tc.editedWeight, extendedWeightForEditor(equipment, data, fxp.Pound),
				"weight preview follows the edit")
			c.Equal(tc.value, equipment.ExtendedValue(), "the target's value is left untouched by the preview")
			c.Equal(tc.weight, equipment.ExtendedWeight(false, fxp.Pound),
				"the target's weight is left untouched by the preview")

			data.ApplyTo(equipment)
			c.Equal(tc.editedValue, equipment.ExtendedValue(), "the applied value matches the preview")
			c.Equal(tc.editedWeight, equipment.ExtendedWeight(false, fxp.Pound),
				"the applied weight matches the preview")
		})
	}
}

// TestExtendedPreviewsShowAChoicesRange verifies that the editor of a container holding a choice yet to be made, or of
// equipment whose mandatory modifier choice is yet to be made, previews the range the choice may come to, as the list
// does, while one holding no choice, or whose options all come to the same, previews the single value and weight the
// list counts.
func TestExtendedPreviewsShowAChoicesRange(t *testing.T) {
	newBackpack := func(withChoice bool) (*gurps.Equipment, *gurps.EquipmentEditData) {
		backpack := gurps.NewEquipment(nil, nil, true)
		backpack.BaseValue = "5"
		backpack.BaseWeight = "2 lb"
		if withChoice {
			choice := gurps.NewEquipmentChoiceContainer(nil, backpack)
			for _, one := range []struct{ value, weight string }{{"10", "1 lb"}, {"30", "3 lb"}} {
				option := gurps.NewEquipment(nil, choice, false)
				option.BaseValue = one.value
				option.BaseWeight = one.weight
				choice.Children = append(choice.Children, option)
			}
			backpack.Children = []*gurps.Equipment{choice}
		}
		var data gurps.EquipmentEditData
		data.CopyFrom(backpack)
		return backpack, &data
	}
	for _, tc := range []struct {
		name          string
		build         func() (*gurps.Equipment, *gurps.EquipmentEditData)
		value, weight string
	}{
		{
			name:   "a container holding an equipment choice",
			build:  func() (*gurps.Equipment, *gurps.EquipmentEditData) { return newBackpack(true) },
			value:  "15~35",
			weight: "3~5 lb",
		},
		{
			name:   "a container holding no choice",
			build:  func() (*gurps.Equipment, *gurps.EquipmentEditData) { return newBackpack(false) },
			value:  "5",
			weight: "2 lb",
		},
		{
			name: "a modifier choice whose options agree",
			build: func() (*gurps.Equipment, *gurps.EquipmentEditData) {
				return newEditorEquipmentWithChoice([2]string{"+10", "+1 lb"}, [2]string{"+10", "+1 lb"})
			},
			value:  "110",
			weight: "3 lb",
		},
		{
			name: "a modifier choice whose options differ",
			build: func() (*gurps.Equipment, *gurps.EquipmentEditData) {
				return newEditorEquipmentWithChoice([2]string{"+10", "+1 lb"}, [2]string{"+30", "+3 lb"})
			},
			value:  "110~130",
			weight: "3~5 lb",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			equipment, data := tc.build()
			c.Equal(tc.value, extendedValueTextForEditor(equipment, data))
			c.Equal(tc.weight, extendedWeightTextForEditor(equipment, data, fxp.Pound))
			if _, settled := equipment.ExtendedValueRange().Settled(); settled {
				c.Equal(tc.value, equipment.ExtendedValue().Comma(), "the list counts the same value")
				c.Equal(tc.weight, fxp.Pound.Format(equipment.ExtendedWeight(false, fxp.Pound)),
					"the list weighs the same weight")
			}
		})
	}
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
		option.SetDisabled(true)
		choice.Children = append(choice.Children, option)
	}
	equipment.Modifiers = []*gurps.EquipmentModifier{choice}
	var data gurps.EquipmentEditData
	data.CopyFrom(equipment)
	return equipment, &data
}

// TestExtendedPreviewsLeaveTheModifiersOnTheEditedItem verifies that working out the editor's previews, which cost a
// throwaway copy of the item, leaves the editor's modifiers, the pick of a choice among them included, modifying the
// item being edited.
func TestExtendedPreviewsLeaveTheModifiersOnTheEditedItem(t *testing.T) {
	c := check.New(t)
	equipment, _ := newEditorEquipmentWithChoice([2]string{"", "+1 lb"}, [2]string{"", "+1 lb"})
	equipment.Modifiers[0].Children[0].SetDisabled(false)
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
