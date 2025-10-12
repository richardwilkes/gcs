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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/emweight"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// EditTraitModifier displays the editor for a trait modifier.
func EditTraitModifier(owner Rebuildable, modifier *gurps.TraitModifier) {
	displayEditor(owner, modifier, svg.GCSTraitModifiers,
		"md:User%20Guide/Trait%20Modifiers", nil, initTraitModifierEditor, nil)
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
		wrapper := addFlowWrapper(content, costLabel, 2)
		field := NewStringField(nil, "", costLabel,
			func() string {
				v := emweight.ValueFromString(e.editorData.CostAdj)
				return v.Format(v.ExtractFraction(e.editorData.CostAdj))
			},
			func(value string) {
				v := emweight.ValueFromString(value)
				e.editorData.CostAdj = v.Format(v.ExtractFraction(value))
				MarkModified(wrapper)
			})
		field.SetMinimumTextWidthUsing("x1000000")
		field.Tooltip = newWrappedTooltip(i18n.Text("Enter a cost adjustment, such as +5, -5, +50%, -25%, x2, x2/3, x10%."))
		wrapper.AddChild(field)
		affectsPopup := addPopup(wrapper, affects.Options, &e.editorData.Affects)
		levelLabel := i18n.Text("Level")
		wrapper = addFlowWrapper(content, levelLabel, 2)
		levels := addDecimalField(wrapper, nil, "", levelLabel, "", &e.editorData.Levels, 0, fxp.Thousand)
		box := addCheckBox(wrapper, i18n.Text("Use level from owner"), &e.editorData.UseLevelFromTrait)
		box.OnSet = func() { adjustFieldBlank(levels, e.editorData.UseLevelFromTrait) }
		adjustFieldBlank(levels, e.editorData.UseLevelFromTrait)
		total := NewNonEditableField(func(field *NonEditableField) {
			costMultiplier := gurps.CostMultiplierForTraitModifier(e.editorData.Levels, e.target.OwningTrait(),
				e.editorData.UseLevelFromTrait)
			v := emweight.ValueFromString(e.editorData.CostAdj)
			f := v.ExtractFraction(e.editorData.CostAdj)
			f.Numerator = f.Numerator.Mul(costMultiplier)
			f.Normalize()
			f = f.Simplify()
			field.SetTitle(v.Format(f))
			enabled := v != emweight.Multiplier && v != emweight.PercentageMultiplier
			if !enabled {
				affectsPopup.Select(affects.Total)
			}
			affectsPopup.SetEnabled(enabled)
			field.MarkForLayoutAndRedraw()
		})
		insets := total.Border().Insets()
		total.SetLayoutData(&unison.FlexLayoutData{
			MinSize: geom.NewSize(total.Font.SimpleWidth((-fxp.MaxBasePoints*2).String())+insets.Left+insets.Right, 0),
		})
		content.AddChild(NewFieldLeadingLabel(i18n.Text("Total"), false))
		content.AddChild(total)
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
