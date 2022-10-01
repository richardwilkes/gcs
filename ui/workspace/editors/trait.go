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

package editors

import (
	"github.com/richardwilkes/gcs/v5/constants"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/ancestry"
	"github.com/richardwilkes/gcs/v5/model/gurps/trait"
	"github.com/richardwilkes/gcs/v5/model/gurps/weapon"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditTrait displays the editor for a trait.
func EditTrait(owner widget.Rebuildable, t *gurps.Trait) {
	displayEditor[*gurps.Trait, *gurps.TraitEditData](owner, t, res.GCSTraitsSVG, initTraitEditor)
}

func initTraitEditor(e *editor[*gurps.Trait, *gurps.TraitEditData], content *unison.Panel) func() {
	addNameLabelAndField(content, &e.editorData.Name)
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	addUserDescLabelAndField(content, &e.editorData.UserDesc)
	addTagsLabelAndField(content, &e.editorData.Tags)
	content.AddChild(unison.NewPanel())
	addInvertedCheckBox(content, i18n.Text("Enabled"), &e.editorData.Disabled)
	var levelField *widget.DecimalField
	if !e.target.Container() {
		wrapper := addFlowWrapper(content, i18n.Text("Point Cost"), 8)
		cost := widget.NewNonEditableField(func(field *widget.NonEditableField) {
			field.Text = gurps.AdjustedPoints(e.target.Entity, e.editorData.BasePoints, e.editorData.Levels,
				e.editorData.PointsPerLevel, e.editorData.CR, e.editorData.Modifiers,
				e.editorData.RoundCostDown).String()
			field.MarkForLayoutAndRedraw()
		})
		insets := cost.Border().Insets()
		cost.SetLayoutData(&unison.FlexLayoutData{
			MinSize: unison.NewSize(cost.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
		})
		wrapper.AddChild(cost)
		addCheckBox(wrapper, i18n.Text("Round Down"), &e.editorData.RoundCostDown)
		baseCost := i18n.Text("Base Cost")
		wrapper = addFlowWrapper(content, baseCost, 8)
		addDecimalField(wrapper, nil, "", baseCost, "", &e.editorData.BasePoints, -fxp.MaxBasePoints,
			fxp.MaxBasePoints)
		addLabelAndDecimalField(wrapper, nil, "", i18n.Text("Per Level"), "", &e.editorData.PointsPerLevel,
			-fxp.MaxBasePoints, fxp.MaxBasePoints)
		levelField = addLabelAndDecimalField(wrapper, nil, "", i18n.Text("Level"), "", &e.editorData.Levels, 0,
			fxp.MaxBasePoints)
		adjustFieldBlank(levelField, e.editorData.PointsPerLevel == 0)
	}
	addLabelAndPopup(content, i18n.Text("Self-Control Roll"), "", trait.AllSelfControlRolls, &e.editorData.CR)
	crAdjPopup := addLabelAndPopup(content, i18n.Text("CR Adjustment"), i18n.Text("Self-Control Roll Adjustment"),
		gurps.AllSelfControlRollAdj, &e.editorData.CRAdj)
	if e.editorData.CR == trait.None {
		crAdjPopup.SetEnabled(false)
	}
	var ancestryPopup *unison.PopupMenu[string]
	if e.target.Container() {
		addLabelAndPopup(content, i18n.Text("Container Type"), "", trait.AllContainerType,
			&e.editorData.ContainerType)
		var choices []string
		for _, lib := range ancestry.AvailableAncestries(gurps.SettingsProvider.Libraries()) {
			for _, one := range lib.List {
				choices = append(choices, one.Name)
			}
		}
		ancestryPopup = addLabelAndPopup(content, i18n.Text("Ancestry"), "", choices, &e.editorData.Ancestry)
		adjustPopupBlank(ancestryPopup, e.editorData.ContainerType != trait.Race)
	}
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	modifiersPanel := newTraitModifiersPanel(e.target.Entity, &e.editorData.Modifiers)
	if e.target.Container() {
		content.AddChild(modifiersPanel)
	} else {
		content.AddChild(newPrereqPanel(e.target.Entity, &e.editorData.Prereq))
		content.AddChild(newFeaturesPanel(e.target.Entity, e.target, &e.editorData.Features))
		content.AddChild(modifiersPanel)
		for _, wt := range weapon.AllType {
			content.AddChild(newWeaponsPanel(e, e.target, wt, &e.editorData.Weapons))
		}
	}
	e.InstallCmdHandlers(constants.NewTraitModifierItemID, unison.AlwaysEnabled,
		func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, ntable.NoItemVariant) })
	e.InstallCmdHandlers(constants.NewTraitContainerModifierItemID, unison.AlwaysEnabled,
		func(_ any) { modifiersPanel.provider.CreateItem(e, modifiersPanel.table, ntable.ContainerItemVariant) })
	return func() {
		if levelField != nil {
			adjustFieldBlank(levelField, e.editorData.PointsPerLevel == 0)
		}
		if e.editorData.CR == trait.None {
			crAdjPopup.SetEnabled(false)
			crAdjPopup.Select(gurps.NoCRAdj)
		} else {
			crAdjPopup.SetEnabled(true)
		}
		if ancestryPopup != nil {
			if e.editorData.ContainerType == trait.Race {
				if !ancestryPopup.Enabled() {
					adjustPopupBlank(ancestryPopup, false)
					if ancestryPopup.IndexOfItem(e.editorData.Ancestry) == -1 {
						e.editorData.Ancestry = ancestry.Default
					}
					ancestryPopup.Select(e.editorData.Ancestry)
				}
			} else {
				adjustPopupBlank(ancestryPopup, true)
			}
		}
	}
}
