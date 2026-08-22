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
	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

func addChoices[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], parent *unison.Panel, templateOnly bool) (
	typePopup *unison.PopupMenu[picker.Type],
	comparisonPopup *unison.PopupMenu[string],
	field unison.Paneler,
) {
	if templateOnly && !HasOwner[*Template](parent) {
		return
	}

	if xreflect.IsNil(e.target) || !e.target.Container() {
		return
	}

	var tp *gurps.TemplatePicker
	if pickable, ok := any(e.target).(gurps.TemplatePickerProvider); ok {
		tp = pickable.TemplatePickerData()
	} else {
		return
	}

	if tp == nil {
		tp = &gurps.TemplatePicker{}
	}

	last := tp.Type
	wrapper := addFlowWrapper(parent, i18n.Text("Choices"), 3)
	typePopup = addPopup(wrapper, picker.Types, &tp.Type)
	text := i18n.Text("Choice Quantifier")
	comparisonPopup, field = addNumericCriteriaPanel(wrapper, nil, "", "", text, &tp.Qualifier, fxp.Min, fxp.Max, 1, false, false)

	// A picker that isn't in use has nothing to quantify, so both the comparison and the qualifier are blanked out. The
	// qualifier is blanked as well whenever the comparison doesn't use one. The opening state must be settled the same
	// way the selection callback settles it, or an untouched editor lets the user alter a picker that will be dropped
	// on save, or refuses edits to one that will be kept.
	adjust := func(pickerType picker.Type) {
		notApplicable := pickerType == picker.NotApplicable
		adjustPopupBlank(comparisonPopup, notApplicable)
		adjustFieldBlank(field, notApplicable || tp.Qualifier.Compare == criteria.AnyNumber)
	}

	typePopup.SelectionChangedCallback = func(p *unison.PopupMenu[picker.Type]) {
		if item, ok := p.Selected(); ok {
			tp.Type = item
			if last == picker.NotApplicable && item != picker.NotApplicable {
				tp.Qualifier.Qualifier = fxp.One
				if syncer, ok2 := field.(Syncer); ok2 {
					syncer.Sync()
				}
			}
			last = item
			adjust(item)
			MarkModified(parent)
		}
	}

	adjust(tp.Type)

	return
}
