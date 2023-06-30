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

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditTrait displays the editor for a trait.
func EditTrait(owner Rebuildable, t *gurps.Trait) {
	displayEditor[*gurps.Trait, *gurps.TraitEditData](owner, t, svg.GCSTraits, "md:Help/Interface/Trait", nil,
		initTraitEditor)
}

func initTraitEditor(e *editor[*gurps.Trait, *gurps.TraitEditData], content *unison.Panel) func() {
	addNameLabelAndField(content, &e.editorData.Name)
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	addUserDescLabelAndField(content, &e.editorData.UserDesc)
	addTagsLabelAndField(content, &e.editorData.Tags)
	content.AddChild(unison.NewPanel())
	addInvertedCheckBox(content, i18n.Text("Enabled"), &e.editorData.Disabled)
	var perLevelField, levelField *DecimalField
	if !e.target.Container() {
		wrapper := addFlowWrapper(content, i18n.Text("Point Cost"), 2)
		costField := NewNonEditableField(func(field *NonEditableField) {
			field.Text = gurps.AdjustedPoints(e.target.Entity, e.editorData.CanLevel, e.editorData.BasePoints,
				e.editorData.Levels, e.editorData.PointsPerLevel, e.editorData.CR, e.editorData.Modifiers,
				e.editorData.RoundCostDown).String()
			field.MarkForLayoutAndRedraw()
		})
		insets := costField.Border().Insets()
		costField.SetLayoutData(&unison.FlexLayoutData{
			MinSize: unison.NewSize(costField.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
		})
		wrapper.AddChild(costField)
		addCheckBox(wrapper, i18n.Text("Round Down"), &e.editorData.RoundCostDown)

		addLabelAndDecimalField(content, nil, "", i18n.Text("Base Cost"), "", &e.editorData.BasePoints,
			-fxp.MaxBasePoints, fxp.MaxBasePoints)

		hasLevelsCheckBox := addCheckBox(content, i18n.Text("Levels"), &e.editorData.CanLevel)
		hasLevelsCheckBox.SetLayoutData(&unison.FlexLayoutData{
			HAlign: unison.EndAlignment,
			VAlign: unison.MiddleAlignment,
		})
		wrapper = unison.NewPanel()
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  3,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
			VAlign:   unison.MiddleAlignment,
		})
		content.AddChild(wrapper)
		levelField = addDecimalField(wrapper, nil, "", i18n.Text("Level"), "", &e.editorData.Levels, 0,
			fxp.MaxBasePoints)
		perLevelField = addLabelAndDecimalField(wrapper, nil, "", i18n.Text("Cost Per Level"), "",
			&e.editorData.PointsPerLevel, -fxp.MaxBasePoints, fxp.MaxBasePoints)
		adjustFieldBlank(perLevelField, !e.editorData.CanLevel)
		adjustFieldBlank(levelField, !e.editorData.CanLevel)
	}
	addLabelAndPopup(content, i18n.Text("Self-Control Roll"), "", gurps.AllSelfControlRolls, &e.editorData.CR)
	crAdjPopup := addLabelAndPopup(content, i18n.Text("CR Adjustment"), i18n.Text("Self-Control Roll Adjustment"),
		gurps.AllSelfControlRollAdj, &e.editorData.CRAdj)
	if e.editorData.CR == gurps.NoCR {
		crAdjPopup.SetEnabled(false)
	}
	var ancestryPopup *unison.PopupMenu[string]
	if e.target.Container() {
		addLabelAndPopup(content, i18n.Text("Container Type"), "", gurps.AllContainerType,
			&e.editorData.ContainerType)
		var choices []string
		for _, lib := range gurps.AvailableAncestries(gurps.GlobalSettings().Libraries()) {
			for _, one := range lib.List {
				choices = append(choices, one.Name)
			}
		}
		ancestryPopup = addLabelAndPopup(content, i18n.Text("Ancestry"), "", choices, &e.editorData.Ancestry)
		adjustPopupBlank(ancestryPopup, e.editorData.ContainerType != gurps.RaceContainerType)
		addTemplateChoices(content, nil, "", &e.editorData.TemplatePicker)
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	modifiersPanel := newTraitModifiersPanel(e.target.Entity, &e.editorData.Modifiers)
	if e.target.Container() {
		content.AddChild(modifiersPanel)
	} else {
		content.AddChild(newPrereqPanel(e.target.Entity, &e.editorData.Prereq))
		content.AddChild(newFeaturesPanel(e.target.Entity, e.target, &e.editorData.Features, false))
		content.AddChild(modifiersPanel)
		for _, wt := range gurps.AllWeaponType {
			content.AddChild(newWeaponsPanel(e, e.target, wt, &e.editorData.Weapons))
		}
		content.AddChild(newStudyPanel(e.target.Entity, &e.editorData.StudyHoursNeeded, &e.editorData.Study))
	}
	e.InstallCmdHandlers(NewTraitModifierItemID, unison.AlwaysEnabled,
		func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, NoItemVariant) })
	e.InstallCmdHandlers(NewTraitContainerModifierItemID, unison.AlwaysEnabled,
		func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, ContainerItemVariant) })
	return func() {
		if perLevelField != nil {
			adjustFieldBlank(perLevelField, !e.editorData.CanLevel)
		}
		if levelField != nil {
			adjustFieldBlank(levelField, !e.editorData.CanLevel)
		}
		if e.editorData.CR == gurps.NoCR {
			crAdjPopup.SetEnabled(false)
			crAdjPopup.Select(gurps.NoCRAdj)
		} else {
			crAdjPopup.SetEnabled(true)
		}
		if ancestryPopup != nil {
			if e.editorData.ContainerType == gurps.RaceContainerType {
				if !ancestryPopup.Enabled() {
					adjustPopupBlank(ancestryPopup, false)
					if ancestryPopup.IndexOfItem(e.editorData.Ancestry) == -1 {
						e.editorData.Ancestry = gurps.DefaultAncestry
					}
					ancestryPopup.Select(e.editorData.Ancestry)
				}
			} else {
				adjustPopupBlank(ancestryPopup, true)
			}
		}
	}
}
