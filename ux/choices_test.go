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

// newEquipmentChoices builds the "Choices" row for an equipment choice container whose picker is set as given,
// returning the editor data and the row's two popups and qualifier field.
func newEquipmentChoices(pickerType picker.Type, compare criteria.NumericComparison, qualifier fxp.Int) (
	data *gurps.EquipmentEditData, typePopup *unison.PopupMenu[picker.Type], comparison *unison.PopupMenu[string],
	field unison.Paneler,
) {
	eqp := gurps.NewEquipmentChoiceContainer(nil, nil)
	eqp.TemplatePicker.Type = pickerType
	eqp.TemplatePicker.Qualifier.Compare = compare
	eqp.TemplatePicker.Qualifier.Qualifier = qualifier
	e := &editor[*gurps.Equipment, *gurps.EquipmentEditData]{target: eqp, editorData: &gurps.EquipmentEditData{}}
	e.editorData.CopyFrom(eqp)
	typePopup, comparison, field = addChoices(e, unison.NewPanel())
	return e.editorData, typePopup, comparison, field
}

// TestChoicesOnlyForChoiceContainers verifies that only a choice container's editor offers choices, and that it can't
// take them out of use, which is left to the "Convert to Group" command.
func TestChoicesOnlyForChoiceContainers(t *testing.T) {
	c := check.New(t)
	typePopup, _, _ := newChoices(gurps.NewTrait(nil, nil, true))
	c.Nil(typePopup, "a plain container must not offer choices")

	typePopup, _, _ = newChoices(newChoiceContainer(picker.Count, criteria.EqualsNumber))
	c.NotNil(typePopup, "a choice container must offer choices")
	c.Equal(-1, typePopup.IndexOfItem(picker.NotApplicable), "a choice container must not offer to stop being one")
}

// TestChoicesOpeningState verifies that a freshly opened editor blanks the picker's qualifier field exactly when its
// comparison takes no qualifier, whether that is a plain number field or the weight field a choice made by weight is
// entered with, and that choosing a comparison that takes a qualifier then offers it.
func TestChoicesOpeningState(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		pickerType picker.Type
		compare    criteria.NumericComparison
		choose     criteria.NumericComparison // Chosen afterwards when compare takes no qualifier
	}{
		{pickerType: picker.Count, compare: criteria.AnyNumber, choose: criteria.EqualsNumber},
		{pickerType: picker.Points, compare: criteria.AtLeastNumber},
		{pickerType: picker.Weight, compare: criteria.AnyNumber, choose: criteria.AtMostNumber},
		{pickerType: picker.Weight, compare: criteria.AtLeastNumber},
	} {
		var comparison *unison.PopupMenu[string]
		var field unison.Paneler
		if one.pickerType == picker.Weight { // Only equipment is picked by weight
			_, _, comparison, field = newEquipmentChoices(one.pickerType, one.compare, fxp.One)
			_, isWeight := field.(*WeightField)
			c.True(isWeight, "a weight must be entered with its units")
		} else {
			_, comparison, field = newChoices(newChoiceContainer(one.pickerType, one.compare))
		}
		c.True(comparison.Enabled(), "%s: a choice container must offer a comparison", one.pickerType)
		if one.compare != criteria.AnyNumber {
			c.True(field.AsPanel().Enabled(), "%s: a comparison that takes a qualifier must offer one", one.pickerType)
			continue
		}
		c.False(field.AsPanel().Enabled(), "%s: a comparison that takes no qualifier must not offer one", one.pickerType)
		comparison.SelectIndex(int(one.choose))
		c.True(field.AsPanel().Enabled(), "%s: choosing a comparison that takes a qualifier must offer one",
			one.pickerType)
	}
}

// TestChoicesForEquipment verifies that an equipment choice container's editor offers choices by count, value or
// weight, taking a weight with its units, and that a group's does not. Changing the type between one entered as a
// number and one entered as a weight keeps the qualifier, and its comparison, as they were.
func TestChoicesForEquipment(t *testing.T) {
	c := check.New(t)
	group := gurps.NewEquipmentGroup(nil, nil)
	e := &editor[*gurps.Equipment, *gurps.EquipmentEditData]{target: group, editorData: &gurps.EquipmentEditData{}}
	e.editorData.CopyFrom(group)
	typePopup, _, _ := addChoices(e, unison.NewPanel())
	c.Nil(typePopup, "a group must not offer choices")

	data, typePopup, _, field := newEquipmentChoices(picker.Count, criteria.AtMostNumber, fxp.FromInteger(25))
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
	c.Equal(fxp.FromInteger(25), data.TemplatePicker.Qualifier.Qualifier, "the qualifier must be kept")
	c.Equal(criteria.AtMostNumber, data.TemplatePicker.Qualifier.Compare, "the comparison must be kept")
	typePopup.Select(picker.Value)
	c.False(weightFieldIn(), "a value is not a weight")
	c.Equal(fxp.FromInteger(25), data.TemplatePicker.Qualifier.Qualifier, "the qualifier must be kept")
	c.Equal(criteria.AtMostNumber, data.TemplatePicker.Qualifier.Compare, "the comparison must be kept")
}

// TestChoiceQualifierMinimum verifies that only a choice made by points may ask for less than nothing: a count, a value
// or a weight can't, so its field won't go below zero, and changing a choice to one of those raises a qualifier that
// is below zero to zero.
func TestChoiceQualifierMinimum(t *testing.T) {
	c := check.New(t)
	decimalFieldIn := func(wrapper *unison.Panel) *DecimalField {
		var found *DecimalField
		for _, one := range wrapper.Children() {
			for _, child := range one.Children() {
				if field, ok := child.Self.(*DecimalField); ok {
					found = field
				}
			}
		}
		return found
	}

	trait := newChoiceContainer(picker.Points, criteria.AtMostNumber)
	trait.TemplatePicker.Qualifier.Qualifier = fxp.FromInteger(-10)
	e := &editor[*gurps.Trait, *gurps.TraitEditData]{target: trait, editorData: &gurps.TraitEditData{}}
	e.editorData.CopyFrom(trait)
	typePopup, _, field := addChoices(e, unison.NewPanel())
	wrapper := field.AsPanel().Parent().Parent()
	c.Equal(fxp.Min, decimalFieldIn(wrapper).Min(), "points may be less than nothing")

	typePopup.Select(picker.Count)
	c.Equal(fxp.Int(0), decimalFieldIn(wrapper).Min(), "a count may not be less than nothing")
	c.Equal(fxp.Int(0), e.editorData.TemplatePicker.Qualifier.Qualifier, "the qualifier must be raised to zero")

	eqp := gurps.NewEquipmentChoiceContainer(nil, nil)
	eqp.TemplatePicker.Type = picker.Value
	eqpEditor := &editor[*gurps.Equipment, *gurps.EquipmentEditData]{target: eqp, editorData: &gurps.EquipmentEditData{}}
	eqpEditor.editorData.CopyFrom(eqp)
	_, _, field = addChoices(eqpEditor, unison.NewPanel())
	decimal, ok := field.(*DecimalField)
	c.True(ok, "a value is entered as a number")
	c.Equal(fxp.Int(0), decimal.Min(), "a value may not be less than nothing")

	eqp.TemplatePicker.Type = picker.Weight
	eqpEditor.editorData.CopyFrom(eqp)
	_, _, field = addChoices(eqpEditor, unison.NewPanel())
	weight, ok := field.(*WeightField)
	c.True(ok, "a weight is entered with its units")
	c.Equal(fxp.Weight(0), weight.Min(), "a weight may not be less than nothing")
}
