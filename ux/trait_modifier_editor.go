// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditTraitModifier displays the editor for a trait modifier.
func EditTraitModifier(owner Rebuildable, modifier *gurps.TraitModifier) {
	displayEditor[*gurps.TraitModifier, *gurps.TraitModifierEditData](owner, modifier, svg.GCSTraitModifiers,
		"md:Help/Interface/Trait Modifiers", nil, initTraitModifierEditor, nil)
}

func initTraitModifierEditor(e *editor[*gurps.TraitModifier, *gurps.TraitModifierEditData], content *unison.Panel) func() {
	addNameLabelAndField(content, &e.editorData.Name)
	addNotesLabelAndField(content, &e.editorData.LocalNotes)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	if !e.target.Container() {
		content.AddChild(unison.NewPanel())
		addInvertedCheckBox(content, i18n.Text("Enabled"), &e.editorData.Disabled)
		costLabel := i18n.Text("Cost")
		wrapper := addFlowWrapper(content, costLabel, 3)
		addDecimalField(wrapper, nil, "", costLabel, "", &e.editorData.Cost, -fxp.MaxBasePoints, fxp.MaxBasePoints)
		costTypePopup := addCostTypePopup(wrapper, e)
		affectsPopup := addPopup(wrapper, affects.Options, &e.editorData.Affects)
		levels := addLabelAndDecimalField(content, nil, "", i18n.Text("Level"), "", &e.editorData.Levels, 0, fxp.Thousand)
		adjustFieldBlank(levels, !e.target.IsLeveled())
		total := NewNonEditableField(func(field *NonEditableField) {
			enabled := true
			switch costTypePopup.SelectedIndex() - 1 {
			case -1:
				field.SetTitle(e.editorData.Cost.Mul(e.editorData.Levels).StringWithSign() + tmcost.Percentage.String())
			case int(tmcost.Percentage):
				field.SetTitle(e.editorData.Cost.StringWithSign() + tmcost.Percentage.String())
			case int(tmcost.Points):
				field.SetTitle(e.editorData.Cost.StringWithSign())
			case int(tmcost.Multiplier):
				field.SetTitle(tmcost.Multiplier.String() + e.editorData.Cost.String())
				affectsPopup.Select(affects.Total)
				enabled = false
			default:
				errs.Log(errs.New("unhandled cost type"), "index", costTypePopup.SelectedIndex())
				field.SetTitle(e.editorData.Cost.StringWithSign() + tmcost.Percentage.String())
			}
			affectsPopup.SetEnabled(enabled)
			field.MarkForLayoutAndRedraw()
		})
		insets := total.Border().Insets()
		total.SetLayoutData(&unison.FlexLayoutData{
			MinSize: unison.NewSize(total.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
		})
		content.AddChild(NewFieldLeadingLabel(i18n.Text("Total"), false))
		content.AddChild(total)
		costTypePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[string]) {
			index := popup.SelectedIndex()
			if index == 0 {
				e.editorData.CostType = tmcost.Percentage
				if e.editorData.Levels < fxp.One {
					levels.SetText("1")
				}
			} else {
				e.editorData.CostType = tmcost.Types[index-1]
				e.editorData.Levels = 0
			}
			adjustFieldBlank(levels, index != 0)
			MarkModified(wrapper)
		}
	}
	addTagsLabelAndField(content, &e.editorData.Tags)
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	if !e.target.Container() {
		content.AddChild(newFeaturesPanel(gurps.EntityFromNode(e.target), e.target, &e.editorData.Features, false))
	}
	return nil
}

func addCostTypePopup(parent *unison.Panel, e *editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]) *unison.PopupMenu[string] {
	popup := unison.NewPopupMenu[string]()
	popup.AddItem(i18n.Text("% per level"))
	for _, one := range tmcost.Types {
		popup.AddItem(one.String())
	}
	if e.target.IsLeveled() {
		popup.SelectIndex(0)
	} else {
		popup.Select(e.editorData.CostType.String())
	}
	parent.AddChild(popup)
	return popup
}
