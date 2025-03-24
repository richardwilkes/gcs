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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
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
		HAlign:   align.Middle,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Basic Damage")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p, gc, rect, 0, 2, nil) }

	// Add basic thrust damage
	p.AddChild(NewPageLabel(i18n.Text("Striking Thrust")))
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.Thrust().String())
		MarkForLayoutWithinDockable(f)
	}))

	// Add basic swing damage
	p.AddChild(NewPageLabel(i18n.Text("Striking Swing")))
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.Swing().String())
		MarkForLayoutWithinDockable(f)
	}))

	// Add Lifting Strength-based damage
	p.AddChild(NewPageLabel(i18n.Text("Lifting Thrust")))
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.LiftingThrust().String())
		MarkForLayoutWithinDockable(f)
	}))

	// Add Lifting Strength-based damage
	p.AddChild(NewPageLabel(i18n.Text("Lifting Swing")))
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.LiftingSwing().String())
		MarkForLayoutWithinDockable(f)
	}))

	// Add IQ-based damage
	if entity.SheetSettings.ShowIQBasedDamage {
		p.AddChild(NewPageLabel(i18n.Text("IQ-Based Damage")))
		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			iqST := entity.ResolveAttributeCurrent("iq")
			iqDamage := entity.ThrustFor(fxp.As[int](iqST)).String()
			f.SetTitle(iqDamage)
			MarkForLayoutWithinDockable(f)
		}))
	}

	// Add TK Strength-based damage
	tkST := entity.TelekineticStrength()
	if tkST > 0 {
		p.AddChild(NewPageLabel(i18n.Text("TK Thrust")))
		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			tkThrust := p.entity.ThrustFor(fxp.As[int](tkST)).String()
			f.SetTitle(tkThrust)
			MarkForLayoutWithinDockable(f)
		}))

		// Add TK Strength-based damage
		p.AddChild(NewPageLabel(i18n.Text("TK Swing")))
		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			tkSwing := p.entity.SwingFor(fxp.As[int](tkST)).String()
			f.SetTitle(tkSwing)
			MarkForLayoutWithinDockable(f)
		}))
	}

	return p
}
