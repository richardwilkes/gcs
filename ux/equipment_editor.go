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

package ux

import (
	"strconv"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/measure"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditEquipment displays the editor for equipment.
func EditEquipment(owner Rebuildable, equipment *model.Equipment, carried bool) {
	displayEditor[*model.Equipment, *model.EquipmentEditData](owner, equipment, svg.GCSEquipment,
		func(e *editor[*model.Equipment, *model.EquipmentEditData], content *unison.Panel) func() {
			addNameLabelAndField(content, &e.editorData.Name)
			addNotesLabelAndField(content, &e.editorData.LocalNotes)
			addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
			addLabelAndStringField(content, i18n.Text("Tech Level"), model.TechLevelInfo, &e.editorData.TechLevel)
			addLabelAndStringField(content, i18n.Text("Legality Class"), model.LegalityClassInfo, &e.editorData.LegalityClass)
			qtyLabel := i18n.Text("Quantity")
			if carried {
				wrapper := addFlowWrapper(content, qtyLabel, 2)
				addDecimalField(wrapper, nil, "", qtyLabel, "", &e.editorData.Quantity, 0, fxp.Max-1)
				addCheckBox(wrapper, i18n.Text("Equipped"), &e.editorData.Equipped)
			} else {
				addLabelAndDecimalField(content, nil, "", qtyLabel, "", &e.editorData.Quantity, 0, fxp.Max-1)
			}
			valueLabel := i18n.Text("Value")
			wrapper := addFlowWrapper(content, valueLabel, 3)
			addDecimalField(wrapper, nil, "", valueLabel, "", &e.editorData.Value, 0, fxp.Max-1)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Extended")))
			wrapper.AddChild(NewNonEditableField(func(field *NonEditableField) {
				var value fxp.Int
				if e.editorData.Quantity > 0 {
					value = model.ValueAdjustedForModifiers(e.editorData.Value, e.editorData.Modifiers)
					if e.target.Container() {
						for _, one := range e.target.Children {
							value += one.ExtendedValue()
						}
					}
					value = value.Mul(e.editorData.Quantity)
				}
				field.Text = value.Comma()
				field.MarkForLayoutAndRedraw()
			}))
			weightLabel := i18n.Text("Weight")
			wrapper = addFlowWrapper(content, weightLabel, 3)
			addWeightField(wrapper, nil, "", weightLabel, "", e.target.Entity, &e.editorData.Weight, false)
			wrapper.AddChild(NewFieldInteriorLeadingLabel(i18n.Text("Extended")))
			wrapper.AddChild(NewNonEditableField(func(field *NonEditableField) {
				var weight measure.Weight
				defUnits := model.SheetSettingsFor(e.target.Entity).DefaultWeightUnits
				if e.editorData.Quantity > 0 {
					weight = model.ExtendedWeightAdjustedForModifiers(defUnits, e.editorData.Quantity, e.editorData.Weight,
						e.editorData.Modifiers, e.editorData.Features, e.target.Children, false, false)
				}
				field.Text = defUnits.Format(weight)
				field.MarkForLayoutAndRedraw()
			}))
			content.AddChild(unison.NewPanel())
			addCheckBox(content, i18n.Text("Ignore weight for skills"), &e.editorData.WeightIgnoredForSkills)
			usesLabel := i18n.Text("Uses")
			wrapper = addFlowWrapper(content, usesLabel, 3)
			usesField := addIntegerField(wrapper, nil, "", usesLabel, "", &e.editorData.Uses, 0, 9999999)
			maxUsesLabel := i18n.Text("Maximum Uses")
			wrapper.AddChild(NewFieldInteriorLeadingLabel(maxUsesLabel))
			addIntegerField(wrapper, nil, "", maxUsesLabel, "", &e.editorData.MaxUses, 0, 9999999)
			addTagsLabelAndField(content, &e.editorData.Tags)
			addPageRefLabelAndField(content, &e.editorData.PageRef)
			adjustFieldBlank(usesField, e.editorData.MaxUses <= 0)
			content.AddChild(newPrereqPanel(e.target.Entity, &e.editorData.Prereq))
			content.AddChild(newFeaturesPanel(e.target.Entity, e.target, &e.editorData.Features))
			modifiersPanel := newEquipmentModifiersPanel(e.target.Entity, &e.editorData.Modifiers)
			content.AddChild(modifiersPanel)
			for _, wt := range model.AllWeaponType {
				content.AddChild(newWeaponsPanel(e, e.target, wt, &e.editorData.Weapons))
			}
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
		})
}
