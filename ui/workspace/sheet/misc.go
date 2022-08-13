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

package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// MiscPanel holds the contents of the miscellaneous block on the sheet.
type MiscPanel struct {
	unison.Panel
	sheet  *Sheet
	prefix string
}

// NewMiscPanel creates a new miscellaneous panel.
func NewMiscPanel(sheet *Sheet) *MiscPanel {
	m := &MiscPanel{
		sheet:  sheet,
		prefix: sheet.targetMgr.NextPrefix(),
	}
	m.Self = m
	m.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	m.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
	})
	m.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Miscellaneous")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	m.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}

	m.AddChild(widget.NewPageLabelEnd(i18n.Text("Created")))
	m.AddChild(widget.NewNonEditablePageField(func(f *widget.NonEditablePageField) {
		if text := m.sheet.entity.CreatedOn.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}))

	m.AddChild(widget.NewPageLabelEnd(i18n.Text("Modified")))
	m.AddChild(widget.NewNonEditablePageField(func(f *widget.NonEditablePageField) {
		if text := m.sheet.entity.ModifiedOn.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}))

	title := i18n.Text("Player")
	m.AddChild(widget.NewPageLabelEnd(title))
	m.AddChild(widget.NewStringPageFieldNoGrab(m.sheet.targetMgr, m.prefix+"player", title,
		func() string { return m.sheet.entity.Profile.PlayerName },
		func(s string) { m.sheet.entity.Profile.PlayerName = s }))

	return m
}

// UpdateModified updates the current modification timestamp.
func (m *MiscPanel) UpdateModified() {
	m.sheet.entity.ModifiedOn = jio.Now()
}

// SetTextAndMarkModified sets the field to the given text, selects it, requests focus, then calls MarkModified().
func SetTextAndMarkModified(field *unison.Field, text string) {
	field.SetText(text)
	field.SelectAll()
	field.RequestFocus()
	field.Parent().MarkForLayoutAndRedraw()
	widget.MarkModified(field)
}
