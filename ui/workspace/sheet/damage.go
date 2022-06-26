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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// DamagePanel holds the contents of the damage block on the sheet.
type DamagePanel struct {
	unison.Panel
	entity *gurps.Entity
}

// NewDamagePanel creates a new damage panel.
func NewDamagePanel(entity *gurps.Entity) *DamagePanel {
	p := &DamagePanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
		HAlign:   unison.MiddleAlignment,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
	})
	p.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Basic Damage")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p, gc, rect, 0, 2) }

	p.AddChild(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		f.Text = p.entity.Thrust().String()
		widget.MarkForLayoutWithinDockable(f)
	}))
	p.AddChild(widget.NewPageLabel(i18n.Text("Basic Thrust")))

	p.AddChild(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		f.Text = p.entity.Swing().String()
		widget.MarkForLayoutWithinDockable(f)
	}))
	p.AddChild(widget.NewPageLabel(i18n.Text("Basic Swing")))

	return p
}
