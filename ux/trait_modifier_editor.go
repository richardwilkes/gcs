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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/affects"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/tmcost"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// EditTraitModifier displays the editor for a trait modifier.
func EditTraitModifier(owner Rebuildable, modifier *gurps.TraitModifier) {
	displayEditor(owner, modifier, svg.GCSTraitModifiers,
		"md:Help/Interface/Trait Modifiers", nil, initTraitModifierEditor, nil)
}

func initTraitModifierEditor(e *editor[*gurps.TraitModifier, *gurps.TraitModifierEditData], content *unison.Panel) func() {
	addNameLabelAndField(content, &e.editorData.Name)
	addLabelAndMultiLineStringField(content, i18n.Text("Notes"), "", &e.editorData.LocalNotes)
	content.AddChild(unison.NewPanel())
	addCheckBox(content, i18n.Text("Also show notes in weapon usage"), &e.editorData.ShowNotesOnWeapon)
	addVTTNotesLabelAndField(content, &e.editorData.VTTNotes)
	if !e.target.Container() {
		content.AddChild(unison.NewPanel())
		addInvertedCheckBox(content, i18n.Text("Enabled"), &e.editorData.Disabled)
		costLabel := i18n.Text("Cost")
		wrapper := addFlowWrapper(content, costLabel, 3)
		addDecimalField(wrapper, nil, "", costLabel, "", &e.editorData.Cost, -fxp.MaxBasePoints, fxp.MaxBasePoints)
		costTypePopup := unison.NewPopupMenu[tmcost.Type]()
		costTypePopup.AddItem(tmcost.Types...)
		costTypePopup.Select(e.editorData.CostType)
		wrapper.AddChild(costTypePopup)
		affectsPopup := addPopup(wrapper, affects.Options, &e.editorData.Affects)
		levelLabel := i18n.Text("Level")
		wrapper = addFlowWrapper(content, levelLabel, 2)
		levels := addDecimalField(wrapper, nil, "", levelLabel, "", &e.editorData.Levels, 0, fxp.Thousand)
		box := addCheckBox(wrapper, i18n.Text("Use level from owner"), &e.editorData.UseLevelFromTrait)
		box.OnSet = func() { adjustFieldBlank(levels, e.editorData.UseLevelFromTrait) }
		adjustFieldBlank(levels, e.editorData.UseLevelFromTrait)
		total := NewNonEditableField(func(field *NonEditableField) {
			enabled := true
			costMultiplier := gurps.CostMultiplierForTraitModifier(e.editorData.Levels, e.target.OwningTrait(),
				e.editorData.UseLevelFromTrait)
			costType, ok := costTypePopup.Selected()
			if ok {
				switch costType {
				case tmcost.Percentage:
					field.SetTitle(e.editorData.Cost.Mul(costMultiplier).StringWithSign() + tmcost.Percentage.String())
				case tmcost.Points:
					field.SetTitle(e.editorData.Cost.Mul(costMultiplier).StringWithSign())
				case tmcost.Multiplier:
					field.SetTitle(tmcost.Multiplier.String() + e.editorData.Cost.Mul(costMultiplier).String())
					affectsPopup.Select(affects.Total)
					enabled = false
				default:
					ok = false
				}
			}
			if !ok {
				errs.Log(errs.New("unhandled cost type"), "index", costTypePopup.SelectedIndex())
				field.SetTitle(e.editorData.Cost.Mul(costMultiplier).StringWithSign() + tmcost.Percentage.String())
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
		costTypePopup.SelectionChangedCallback = func(popup *unison.PopupMenu[tmcost.Type]) {
			if what, ok := popup.Selected(); ok {
				e.editorData.CostType = what
				MarkModified(popup)
			}
		}
	}
	addTagsLabelAndField(content, &e.editorData.Tags)
	addPageRefLabelAndField(content, &e.editorData.PageRef)
	addPageRefHighlightLabelAndField(content, &e.editorData.PageRefHighlight)
	addSourceFields(content, &e.target.SourcedID)
	if !e.target.Container() {
		content.AddChild(newFeaturesPanel(gurps.EntityFromNode(e.target), e.target, &e.editorData.Features, false))
	}
	return nil
}
