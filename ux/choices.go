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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// addChoices adds the "Choices" row to the editor of a template choice container, and nothing to any other editor. The
// row doesn't offer to take the picker out of use, since that would leave a plain container behind; turning a choice
// container back into a group is the job of the "Convert to Group" command, which warns before removing the choices.
// Only a template may hold a choice container, so the row needs no check of where the editor was opened.
func addChoices[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], parent *unison.Panel) (
	typePopup *unison.PopupMenu[picker.Type],
	comparisonPopup *unison.PopupMenu[string],
	field unison.Paneler,
) {
	if xreflect.IsNil(e.target) || !gurps.IsTemplateChoiceContainer(e.target) {
		return typePopup, comparisonPopup, field
	}
	pickable, ok := any(e.editorData).(gurps.TemplatePickerProvider)
	if !ok {
		return typePopup, comparisonPopup, field
	}
	types, tp := pickable.TemplatePickerData()
	types = slices.DeleteFunc(slices.Clone(types), func(one picker.Type) bool { return one == picker.NotApplicable })
	wrapper, label := addFlowWrapper(parent, i18n.Text("Choices"), 3)
	typePopup = labelControl(addPopup(wrapper, types, &tp.Type), label)
	entity := gurps.EntityFromNode(e.target)
	comparisonPopup, field = addChoiceQualifier(wrapper, entity, tp)
	// The qualifier's field depends on the type: a weight is entered with its units, and only points may be less than
	// nothing. So the field is rebuilt whenever the type changes. The qualifier itself is kept as it is, since it is a
	// bare number either way, a weight being held in canonical units, unless it is below what the new type allows.
	current := field
	selected := typePopup.SelectionChangedCallback
	typePopup.SelectionChangedCallback = func(p *unison.PopupMenu[picker.Type]) {
		was := tp.Type
		selected(p)
		if tp.Type != was {
			tp.Qualifier.Qualifier = max(tp.Qualifier.Qualifier, choiceQualifierMinimum(tp.Type))
			current.AsPanel().Parent().RemoveFromParent()
			_, current = addChoiceQualifier(wrapper, entity, tp)
			wrapper.MarkForLayoutRecursivelyUpward()
		}
	}
	return typePopup, comparisonPopup, field
}

// choiceQualifierMinimum returns the least a picker of the given type may ask for. Only points may be less than
// nothing, since a disadvantage costs negative points; a count, a value or a weight never can be.
func choiceQualifierMinimum(pickerType picker.Type) fxp.Int {
	if pickerType == picker.Points {
		return fxp.Min
	}
	return 0
}

// addChoiceQualifier adds the comparison and the qualifier the picker is to meet, the qualifier being entered as a
// weight when the picker is made by weight.
//
// Both kinds of field store the qualifier through a setter that holds it to what the picker's type allows at the time
// it is called, not when the field was built. The field is rebuilt whenever the type changes, but the undo edits a
// field records still go to that field after it has been replaced, so an undo could otherwise bring back a qualifier
// the new type forbids, such as a negative one for a count.
func addChoiceQualifier(parent *unison.Panel, entity *gurps.Entity, tp *gurps.TemplatePicker) (popup *unison.PopupMenu[string], field unison.Paneler) {
	panel := newCriteriaPanel(parent, 1, false)
	comparisonName, undoTitle := criteriaTitles(i18n.Text("Choice"))
	popup = newComparisonPopup(comparisonName, criteria.PrefixedNumericComparisonChoices(""),
		int(tp.Qualifier.Compare.EnsureValid()))
	panel.AddChild(popup)
	set := func(value fxp.Int) {
		tp.Qualifier.Qualifier = max(value, choiceQualifierMinimum(tp.Type))
		MarkModified(panel)
	}
	if tp.Type == picker.Weight {
		field = NewWeightField(nil, "", undoTitle, entity,
			func() fxp.Weight { return fxp.Weight(tp.Qualifier.Qualifier) },
			func(value fxp.Weight) { set(fxp.Int(value)) }, 0, fxp.Weight(fxp.Max), false)
	} else {
		field = NewDecimalField(nil, "", undoTitle, func() fxp.Int { return tp.Qualifier.Qualifier }, set,
			choiceQualifierMinimum(tp.Type), fxp.Max, false, false)
	}
	panel.AddChild(field)
	popup.SelectionChangedCallback = func(p *unison.PopupMenu[string]) {
		tp.Qualifier.Compare = criteria.NumericComparisons[p.SelectedIndex()]
		adjustFieldBlank(field, tp.Qualifier.Compare == criteria.AnyNumber)
		MarkModified(panel)
	}
	adjustFieldBlank(field, tp.Qualifier.Compare == criteria.AnyNumber)
	return popup, field
}
