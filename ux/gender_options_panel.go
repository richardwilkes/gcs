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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// genderOptionsPanel is one row of the ancestry editor's gender list: a drag handle, a remove button, and the full set
// of options for the gender, starting with its weight and name.
type genderOptionsPanel struct {
	unison.Panel
	dockable *ancestryEditorDockable
	gender   *gurps.WeightedAncestryOptions
}

func newGenderOptionsPanel(d *ancestryEditorDockable, gender *gurps.WeightedAncestryOptions) *genderOptionsPanel {
	p := &genderOptionsPanel{dockable: d, gender: gender}
	p.Self = p
	configureEditorRow(p.AsPanel(), 3)
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: d,
		row:    p.AsPanel(),
		title:  i18n.Text("Gender Drag"),
		move: func(to int) bool {
			return moveEntry(&d.model.GenderOptions, slices.Index(d.model.GenderOptions, gender), to)
		},
	}))
	deleteButton := unison.NewSVGButton(unison.TrashSVG)
	deleteButton.ClickCallback = p.removeGender
	deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove gender"))
	deleteButton.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
	p.AddChild(deleteButton)
	p.AddChild(newAncestryOptionsPanel(d, gender.Value, gender))
	return p
}

// removeGender removes this gender from the ancestry. Removing the last one is allowed: with no genders, the common
// options are used for everything.
func (p *genderOptionsPanel) removeGender() {
	d := p.dockable
	i := slices.Index(d.model.GenderOptions, p.gender)
	if i == -1 {
		return
	}
	d.editStructure(i18n.Text("Remove Gender"), func() {
		d.model.GenderOptions = slices.Delete(d.model.GenderOptions, i, i+1)
	}, "")
}
