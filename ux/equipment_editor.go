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
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditEquipment displays the editor for equipment.
func EditEquipment(owner Rebuildable, equipment *gurps.Equipment, carried bool) {
	displayEditor(owner, equipment, svg.GCSEquipment,
		"md:Help/Interface/Equipment", nil,
		func(e *editor[*gurps.Equipment, *gurps.EquipmentEditData], content *unison.Panel) func() {
			addNameLabelAndField(content, &e.editorData.Name)
			addNotesLabelAndField(content, &e.editorData.LocalNotes)
			addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
			addLabelAndStringField(content, i18n.Text("Tech Level"), gurps.TechLevelInfo(), &e.editorData.TechLevel)
			addLabelAndStringField(content, i18n.Text("Legality Class"),
				i18n.Text("LC0: Banned\nLC1: Military\nLC2: Restricted\nLC3: Licensed\nLC4: Open"),
				&e.editorData.LegalityClass)
			qtyLabel := i18n.Text("Quantity")
			if carried {
				wrapper := addFlowWrapper(content, qtyLabel, 2)
				addDecimalField(wrapper, nil, "", qtyLabel, "", &e.editorData.Quantity, 0, fxp.Max-1)
				addCheckBox(wrapper, i18n.Text("Equipped"), &e.editorData.Equipped)
			} else {
				addLabelAndDecimalField(content, nil, "", qtyLabel, "", &e.editorData.Quantity, 0, fxp.Max-1)
			}
			valueLabel := i18n.Text("Value")
			content.AddChild(NewFieldLeadingLabel(valueLabel, false))
			addScriptField(content, nil, "", valueLabel,
				i18n.Text("The value, which may be a number or a script expression"),
				func() string { return e.editorData.BaseValue },
				func(value string) {
					e.editorData.BaseValue = value
					MarkModified(content)
				})
			entity := gurps.EntityFromNode(e.target)
			extendedValueLabel := i18n.Text("Extended Value")
			content.AddChild(NewFieldLeadingLabel(extendedValueLabel, false))
			content.AddChild(NewNonEditableField(func(field *NonEditableField) {
				var value fxp.Int
				if e.editorData.Quantity > 0 {
					value = gurps.ValueAdjustedForModifiers(e.target,
						cloneEquipmentWithOverlay(e.target, e.editorData).ResolvedBaseValue(), e.editorData.Modifiers)
					if e.target.Container() {
						for _, one := range e.target.Children {
							value += one.ExtendedValue()
						}
					}
					value = value.Mul(e.editorData.Quantity)
				}
				field.SetTitle(value.Comma())
				field.MarkForLayoutAndRedraw()
			}))
			weightLabel := i18n.Text("Weight")
			content.AddChild(NewFieldLeadingLabel(weightLabel, false))
			addScriptField(content, nil, "", weightLabel,
				i18n.Text("The weight, which may be a number with optional units or a script expression"),
				func() string { return e.editorData.BaseWeight },
				func(value string) {
					e.editorData.BaseWeight = value
					MarkModified(content)
				})
			extendedWeightLabel := i18n.Text("Extended Weight")
			content.AddChild(NewFieldLeadingLabel(extendedWeightLabel, false))
			content.AddChild(NewNonEditableField(func(field *NonEditableField) {
				var weight fxp.Weight
				defUnits := gurps.SheetSettingsFor(entity).DefaultWeightUnits
				if e.editorData.Quantity > 0 {
					weight = gurps.ExtendedWeightAdjustedForModifiers(e.target, defUnits, e.editorData.Quantity,
						cloneEquipmentWithOverlay(e.target, e.editorData).ResolvedBaseWeight(), e.editorData.Modifiers, e.editorData.Features, e.target.Children, false,
						false)
				}
				field.SetTitle(defUnits.Format(weight))
				field.MarkForLayoutAndRedraw()
			}))
			content.AddChild(unison.NewPanel())
			addCheckBox(content, i18n.Text("Ignore weight for skills"), &e.editorData.WeightIgnoredForSkills)
			usesLabel := i18n.Text("Uses")
			wrapper := addFlowWrapper(content, usesLabel, 3)
			usesField := addIntegerField(wrapper, nil, "", usesLabel, "", &e.editorData.Uses, 0, 9999999)
			maxUsesLabel := i18n.Text("Maximum Uses")
			wrapper.AddChild(NewFieldInteriorLeadingLabel(maxUsesLabel, false))
			addIntegerField(wrapper, nil, "", maxUsesLabel, "", &e.editorData.MaxUses, 0, 9999999)
			addLabelAndDecimalField(content, nil, "", i18n.Text("Rated ST"), i18n.Text("Equipment with a rated ST use this value instead of the user's ST"), &e.editorData.RatedST, 0, fxp.Max)
			addLabelAndDecimalField(content, nil, "", i18n.Text("Level"), i18n.Text("Level can be used with features and modifiers that have per-level effects"), &e.editorData.Level, 0, fxp.Max)
			addTagsLabelAndField(content, &e.editorData.Tags)
			addPageRefLabelAndField(content, &e.editorData.PageRef)
			addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
			addSourceFields(content, &e.target.SourcedID)
			adjustFieldBlank(usesField, e.editorData.MaxUses <= 0)
			content.AddChild(newPrereqPanel(entity, &e.editorData.Prereq, prereq.TypesForEquipment))
			content.AddChild(newFeaturesPanel(entity, e.target, &e.editorData.Features, false))
			modifiersPanel := newEquipmentModifiersPanel(entity, &e.editorData.Modifiers)
			content.AddChild(modifiersPanel)
			content.AddChild(newWeaponsPanel(e, e.target, true, &e.editorData.Weapons))
			content.AddChild(newWeaponsPanel(e, e.target, false, &e.editorData.Weapons))
			e.InstallCmdHandlers(NewEquipmentModifierItemID, unison.AlwaysEnabled,
				func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, NoItemVariant) })
			e.InstallCmdHandlers(NewEquipmentContainerModifierItemID, unison.AlwaysEnabled,
				func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, ContainerItemVariant) })
			return func() {
				if e.editorData.Uses > e.editorData.MaxUses {
					usesField.SetText(strconv.Itoa(e.editorData.MaxUses))
				}
				adjustFieldBlank(usesField, e.editorData.MaxUses <= 0)
			}
		}, nil)
}

func cloneEquipmentWithOverlay(e *gurps.Equipment, overlay *gurps.EquipmentEditData) *gurps.Equipment {
	clone := e.Clone(e.Source.LibraryFile, e.DataOwner(), e.Parent(), true)
	clone.EquipmentEditData = *overlay
	return clone
}
