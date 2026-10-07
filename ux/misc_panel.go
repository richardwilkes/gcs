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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// MiscPanel holds the contents of the miscellaneous block on the sheet.
type MiscPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *TargetMgr
	prefix    string
}

// NewMiscPanel creates a new miscellaneous panel.
func NewMiscPanel(entity *gurps.Entity, targetMgr *TargetMgr) *MiscPanel {
	m := &MiscPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	initTitledPagePanel(m, i18n.Text("Miscellaneous"), 2, true, colors.TintMisc)

	m.AddChild(NewPageLabelEnd(i18n.Text("Created")))
	m.AddChild(NewNonEditablePageFieldFor(func() string { return m.entity.CreatedOn.String() }))

	m.AddChild(NewPageLabelEnd(i18n.Text("Modified")))
	m.AddChild(NewNonEditablePageFieldFor(func() string { return m.entity.ModifiedOn.String() }))

	addLabeledStringPageField(m, m.targetMgr, m.prefix+"player", i18n.Text("Player"), NewPageLabelEnd,
		func() string { return m.entity.Profile.PlayerName },
		func(s string) { m.entity.Profile.PlayerName = s })

	return m
}

// UpdateModified updates the current modification timestamp.
func (m *MiscPanel) UpdateModified() {
	m.entity.ModifiedOn = jio.Now()
}

// SelectableTextField is a field whose text can be replaced wholesale and then selected. Both *unison.Field and the
// fields built on top of it, such as *NumericField, satisfy it. It exists so that SetTextAndMarkModified can be handed
// the outer field rather than the *unison.Field embedded within it, since the outer field may do work of its own when
// its text is replaced.
type SelectableTextField interface {
	unison.Paneler
	SetText(text string)
	SelectAll()
}

// SetTextAndMarkModified sets the field to the given text, selects it, requests focus, then calls MarkModified().
func SetTextAndMarkModified(field SelectableTextField, text string) {
	field.SetText(text)
	field.SelectAll()
	panel := field.AsPanel()
	panel.RequestFocus()
	panel.Parent().MarkForLayoutAndRedraw()
	MarkModified(field)
}
