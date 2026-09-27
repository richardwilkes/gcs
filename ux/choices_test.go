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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// newChoices builds the "Choices" row an item editor shows for the container, returning the two popups and the
// qualifier field it is made of, all nil when the editor shows no such row.
func newChoices(trait *gurps.Trait) (typePopup *unison.PopupMenu[picker.Type], comparisonPopup *unison.PopupMenu[string], field unison.Paneler) {
	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait, editorData: &gurps.TraitEditData{}}
	e.editorData.CopyFrom(trait)
	return addChoices(e, unison.NewPanel())
}

// newChoiceContainer returns a trait choice container whose picker is set as given.
func newChoiceContainer(pickerType picker.Type, compare criteria.NumericComparison) *gurps.Trait {
	trait := gurps.NewTraitChoiceContainer(nil, nil)
	trait.TemplatePicker.Type = pickerType
	trait.TemplatePicker.Qualifier.Compare = compare
	trait.TemplatePicker.Qualifier.Qualifier = fxp.One
	return trait
}

// TestChoicesOnlyForChoiceContainers verifies that only a choice container's editor offers choices, and that it can't
// take them out of use. A container becomes a choice container only by being created as one, and stays one.
func TestChoicesOnlyForChoiceContainers(t *testing.T) {
	c := check.New(t)
	typePopup, _, _ := newChoices(gurps.NewTrait(nil, nil, true))
	c.Nil(typePopup, "a plain container must not offer choices")

	typePopup, _, _ = newChoices(newChoiceContainer(picker.Count, criteria.EqualsNumber))
	c.NotNil(typePopup, "a choice container must offer choices")
	c.Equal(-1, typePopup.IndexOfItem(picker.NotApplicable), "a choice container must not offer to stop being one")
}

// TestChoicesForEquipment verifies that an equipment choice container's editor offers choices by count, value or
// weight, taking a weight with its units, and that a group's does not.
func TestChoicesForEquipment(t *testing.T) {
	c := check.New(t)
	newEquipmentChoices := func(eqp *gurps.Equipment) (*unison.PopupMenu[picker.Type], unison.Paneler, *unison.Panel) {
		e := &editor[*gurps.Equipment, *gurps.EquipmentEditData]{target: eqp, editorData: &gurps.EquipmentEditData{}}
		e.editorData.CopyFrom(eqp)
		parent := unison.NewPanel()
		typePopup, _, field := addChoices(e, parent)
		return typePopup, field, parent
	}
	typePopup, _, _ := newEquipmentChoices(gurps.NewEquipmentGroup(nil, nil))
	c.Nil(typePopup, "a group must not offer choices")

	typePopup, field, _ := newEquipmentChoices(gurps.NewEquipmentChoiceContainer(nil, nil))
	c.NotNil(typePopup, "a choice container must offer choices")
	c.Equal(3, typePopup.ItemCount(), "equipment must be picked by count, value or weight")
	c.Equal(-1, typePopup.IndexOfItem(picker.Points), "equipment has no points to pick by")
	_, isWeight := field.(*WeightField)
	c.False(isWeight, "a count is not a weight")

	wrapper := field.AsPanel().Parent().Parent()
	weightFieldIn := func() bool {
		found := false
		for _, one := range wrapper.Children() {
			for _, child := range one.Children() {
				if _, ok := child.Self.(*WeightField); ok {
					found = true
				}
			}
		}
		return found
	}
	typePopup.Select(picker.Weight)
	c.True(weightFieldIn(), "a weight must be entered with its units")
	typePopup.Select(picker.Value)
	c.False(weightFieldIn(), "a value is not a weight")
}
