// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emcost"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/check"
)

const equipmentCostAndWeightPrototype = "-99.99 CF"

// EditEquipmentModifier displays the editor for an equipment modifier.
func EditEquipmentModifier(owner Rebuildable, modifier *gurps.EquipmentModifier) {
	displayEditor(owner, modifier, svg.GCSEquipmentModifiers, "md:Help/Interface/Equipment Modifiers", nil,
		initEquipmentModifierEditor, nil)
}

func initEquipmentModifierEditor(e *editor[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData], content *unison.Panel) func() {
	addNameLabelAndField(content, &e.editorData.Name)
	if !e.target.Container() {
		addLabelAndStringField(content, i18n.Text("Tech Level"), gurps.TechLevelInfo(), &e.editorData.TechLevel)
	}
	addLabelAndMultiLineStringField(content, i18n.Text("Notes"), "", &e.editorData.LocalNotes)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Also show notes in weapon usage"), &e.editorData.ShowNotesOnWeapon)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	if !e.target.Container() {
		content.AddChild(unison.NewPanel())
		addInvertedCheckBox(content, i18n.Text("Enabled"), &e.editorData.Disabled)
		addEquipmentCostFields(content, e)
		addEquipmentWeightFields(content, e)
	}
	addTagsLabelAndField(content, &e.editorData.Tags)
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)
	if !e.target.Container() {
		content.AddChild(newFeaturesPanel(gurps.EntityFromNode(e.target), e.target, &e.editorData.Features, true))
	}
	return nil
}

func addEquipmentCostFields(parent *unison.Panel, e *editor[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData]) {
	label := i18n.Text("Cost Modifier")
	wrapper := addFlowWrapper(parent, label, 3)
	field := NewStringField(nil, "", label,
		func() string { return e.editorData.CostType.Format(e.editorData.CostAmount) },
		func(value string) {
			e.editorData.CostAmount = e.editorData.CostType.Format(value)
			MarkModified(parent)
		})
	field.SetMinimumTextWidthUsing(equipmentCostAndWeightPrototype)
	wrapper.AddChild(field)
	popup := unison.NewPopupMenu[string]()
	for _, one := range emcost.Types {
		popup.AddItem(one.StringWithExample())
	}
	popup.SelectIndex(int(e.editorData.CostType))
	wrapper.AddChild(popup)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		e.editorData.CostType = emcost.Types[p.SelectedIndex()]
		field.SetText(e.editorData.CostType.Format(field.Text()))
		MarkModified(wrapper)
	}
	wrapper.AddChild(NewCheckBox(nil, "", i18n.Text("Per Level"),
		func() check.Enum { return check.FromBool(e.editorData.CostIsPerLevel) },
		func(in check.Enum) { e.editorData.CostIsPerLevel = in == check.On }))
}

func addEquipmentWeightFields(parent *unison.Panel, e *editor[*gurps.EquipmentModifier, *gurps.EquipmentModifierEditData]) {
	units := gurps.SheetSettingsFor(gurps.EntityFromNode(e.target)).DefaultWeightUnits
	label := i18n.Text("Weight Modifier")
	wrapper := addFlowWrapper(parent, label, 3)
	field := NewStringField(nil, "", label,
		func() string { return e.editorData.WeightType.Format(e.editorData.WeightAmount, units) },
		func(value string) {
			e.editorData.WeightAmount = e.editorData.WeightType.Format(value, units)
			MarkModified(parent)
		})
	field.SetMinimumTextWidthUsing(equipmentCostAndWeightPrototype)
	wrapper.AddChild(field)
	popup := unison.NewPopupMenu[string]()
	for _, one := range emweight.Types {
		popup.AddItem(one.StringWithExample())
	}
	popup.SelectIndex(int(e.editorData.WeightType))
	wrapper.AddChild(popup)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		e.editorData.WeightType = emweight.Types[p.SelectedIndex()]
		field.SetText(e.editorData.WeightType.Format(field.Text(), units))
		MarkModified(wrapper)
	}
	wrapper.AddChild(NewCheckBox(nil, "", i18n.Text("Per Level"),
		func() check.Enum { return check.FromBool(e.editorData.WeightIsPerLevel) },
		func(in check.Enum) { e.editorData.WeightIsPerLevel = in == check.On }))
}
