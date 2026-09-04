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
		wrapper = addFlowWrapper(content, levelLabel, 3)
		levels := addDecimalField(wrapper, nil, "", levelLabel, "", &e.editorData.Levels, 0, fxp.Thousand)
		box := addCheckBox(wrapper, i18n.Text("Use level from owner"), &e.editorData.UseLevelFromTrait)
		box.OnSet = func() { adjustFieldBlank(levels, e.editorData.UseLevelFromTrait) }
		adjustFieldBlank(levels, e.editorData.UseLevelFromTrait)
		multiplyBox := addInvertedCheckBox(wrapper, i18n.Text("Multiply cost by level"), &e.editorData.CostIgnoresLevel)
		multiplyBox.Tooltip = newWrappedTooltip(i18n.Text("When checked, a leveled modifier's cost adjustment is multiplied by its level. For a modifier that takes its level from its owner, that is the owner's purchased level, not counting any levels granted by bonuses. Uncheck it to leave the cost adjustment unmultiplied, while any per-level features the modifier carries still scale with the level. Either way, a point adjustment that affects levels only is added to the owner's cost per level, so it is still charged for each of the owner's levels."))
		total := NewNonEditableField(func(field *NonEditableField) {
			v := emweight.ValueFromString(e.editorData.CostAdj)
			modifier := traitModifierWithOverlay(e.target, e.editorData)
			field.SetTitle(v.Format(modifier.CostModifierForTrait(owningTraitWithPendingEdits(e.owner, e.target)).Simplify()))
			// The option only means something for a leveled modifier. A "use level from owner" modifier counts as
			// leveled even when it has no owner yet, as in a library list, since it takes its level from whichever
			// trait it is eventually attached to.
			multiplyBox.SetEnabled(e.editorData.UseLevelFromTrait || e.editorData.Levels > 0)
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

// traitModifierWithOverlay returns a copy of the modifier with the editor's working data applied, so that cost
// computations reflect what has been entered rather than what was last applied.
func traitModifierWithOverlay(t *gurps.TraitModifier, overlay *gurps.TraitModifierEditData) *gurps.TraitModifier {
	clone := *t
	clone.TraitModifierEditData = *overlay
	return &clone
}

// owningTraitWithPendingEdits returns the trait the modifier belongs to. When the modifier is being edited from within
// that trait's own editor, the trait is returned as it currently stands in that editor, so that a level change which
// has not yet been applied is still reflected in the modifier's cost.
func owningTraitWithPendingEdits(owner Rebuildable, modifier *gurps.TraitModifier) *gurps.Trait {
	trait := modifier.OwningTrait()
	if trait == nil {
		return nil
	}
	if traitEditor, ok := owner.(*editor[*gurps.Trait, *gurps.TraitEditData]); ok && traitEditor.target == trait {
		return cloneTraitWithOverlay(trait, traitEditor.editorData)
	}
	return trait
}
