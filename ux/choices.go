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
func addChoices[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], parent *unison.Panel, templateOnly bool) (
	typePopup *unison.PopupMenu[picker.Type],
	comparisonPopup *unison.PopupMenu[string],
	field unison.Paneler,
) {
	if templateOnly && !HasOwner[*Template](parent) {
		return typePopup, comparisonPopup, field
	}
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
	comparisonPopup, field = addNumericCriteriaPanel(wrapper, nil, "", "", i18n.Text("Choice"), &tp.Qualifier, fxp.Min,
		fxp.Max, 1, false, false)
	return typePopup, comparisonPopup, field
}
