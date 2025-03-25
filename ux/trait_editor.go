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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// EditTrait displays the editor for a trait.
func EditTrait(owner Rebuildable, t *gurps.Trait) {
	displayEditor(owner, t, svg.GCSTraits, "md:Help/Interface/Trait", nil,
		initTraitEditor, nil)
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
	entity := gurps.EntityFromNode(e.target)
	if !e.target.Container() {
		wrapper := addFlowWrapper(content, i18n.Text("Point Cost"), 2)
		costField := NewNonEditableField(func(field *NonEditableField) {
			field.SetTitle(gurps.AdjustedPoints(entity, e.target, e.editorData.CanLevel, e.editorData.BasePoints,
				e.editorData.Levels, e.editorData.PointsPerLevel, e.editorData.CR, e.editorData.Modifiers,
				e.editorData.RoundCostDown).String())
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
			HAlign: align.End,
			VAlign: align.Middle,
		})
		wrapper = unison.NewPanel()
		wrapper.SetLayout(&unison.FlexLayout{
			Columns:  3,
			HSpacing: unison.StdHSpacing,
			VSpacing: unison.StdVSpacing,
			VAlign:   align.Middle,
		})
		content.AddChild(wrapper)
		levelField = addDecimalField(wrapper, nil, "", i18n.Text("Level"), "", &e.editorData.Levels, 0,
			fxp.MaxBasePoints)
		perLevelField = addLabelAndDecimalField(wrapper, nil, "", i18n.Text("Cost Per Level"), "",
			&e.editorData.PointsPerLevel, -fxp.MaxBasePoints, fxp.MaxBasePoints)
		adjustFieldBlank(perLevelField, !e.editorData.CanLevel)
		adjustFieldBlank(levelField, !e.editorData.CanLevel)
	}
	addLabelAndPopup(content, i18n.Text("Self-Control Roll"), "", selfctrl.Rolls, &e.editorData.CR)
	crAdjPopup := addLabelAndPopup(content, i18n.Text("CR Adjustment"), i18n.Text("Self-Control Roll Adjustment"),
		selfctrl.Adjustments, &e.editorData.CRAdj)
	if e.editorData.CR == selfctrl.NoCR {
		crAdjPopup.SetEnabled(false)
	}
	var ancestryPopup *unison.PopupMenu[string]
	if e.target.Container() {
		addLabelAndPopup(content, i18n.Text("Container Type"), "", container.Types,
			&e.editorData.ContainerType)
		var choices []string
		for _, lib := range gurps.AvailableAncestries(gurps.GlobalSettings().Libraries()) {
			for _, one := range lib.List {
				choices = append(choices, one.Name)
			}
		}
		ancestryPopup = addLabelAndPopup(content, i18n.Text("Ancestry"), "", choices, &e.editorData.Ancestry)
		adjustPopupBlank(ancestryPopup, e.editorData.ContainerType != container.Ancestry)
		addTemplateChoices(content, nil, "", &e.editorData.TemplatePicker)
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)
	modifiersPanel := newTraitModifiersPanel(entity, &e.editorData.Modifiers)
	content.AddChild(newPrereqPanel(entity, &e.editorData.Prereq, prereq.TypesForNonEquipment))
	if e.target.Container() {
		content.AddChild(modifiersPanel)
	} else {
		content.AddChild(newFeaturesPanel(entity, e.target, &e.editorData.Features, false))
		content.AddChild(modifiersPanel)
		content.AddChild(newWeaponsPanel(e, e.target, true, &e.editorData.Weapons))
		content.AddChild(newWeaponsPanel(e, e.target, false, &e.editorData.Weapons))
		content.AddChild(newStudyPanel(entity, &e.editorData.StudyHoursNeeded, &e.editorData.Study))
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
		if e.editorData.CR == selfctrl.NoCR {
			crAdjPopup.SetEnabled(false)
			crAdjPopup.Select(selfctrl.NoCRAdj)
		} else {
			crAdjPopup.SetEnabled(true)
		}
		if ancestryPopup != nil {
			if e.editorData.ContainerType == container.Ancestry {
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
