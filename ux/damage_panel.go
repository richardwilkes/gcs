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
	entity        *gurps.Entity
	targetMgr     *TargetMgr
	titledBorder  *TitledBorder
	row           []unison.Paneler
	sepLayoutData []*unison.FlexLayoutData
	hash          uint64
}

// NewDamagePanel creates a new damage panel.
func NewDamagePanel(entity *gurps.Entity, targetMgr *TargetMgr) *DamagePanel {
	p := &DamagePanel{entity: entity, targetMgr: targetMgr}
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

	// Populate the panel
	p.populate()

	return p
}

// Sync the panel to the current data.
func (p *DamagePanel) Sync() {
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	if hash := gurps.Hash64(attrs); hash != p.hash {
		p.hash = hash
		p.clearAndPopulate()
	}
	MarkForLayoutWithinDockable(p)
}

// forceSync forces the panel to update.
func (a *DamagePanel) forceSync() {
	attrs := gurps.SheetSettingsFor(a.entity).Attributes
	a.hash = gurps.Hash64(attrs)
	a.clearAndPopulate() // Replace rebuild with clearAndPopulate
	MarkForLayoutWithinDockable(a)
}

// SheetSettingsUpdated handles updates to the sheet settings.
func (p *DamagePanel) SheetSettingsUpdated(entity *gurps.Entity, blockLayout bool) {
	if entity == p.entity {
		p.Sync() // Trigger a sync to update the panel
	}
}

// clearAndPopulate clears the panel and repopulates it with updated data.
func (p *DamagePanel) clearAndPopulate() {
	p.RemoveAllChildren()          // Clear all existing children
	p.populate()                   // Repopulate the panel
	MarkForLayoutWithinDockable(p) // Ensure layout is updated
}

// populate adds the content to the panel.
func (p *DamagePanel) populate() {
	// Add basic thrust damage
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.Thrust().String())
		MarkForLayoutWithinDockable(f)
	}))
	p.AddChild(NewPageLabel(i18n.Text("Basic Thrust")))

	// Add basic swing damage
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.Swing().String())
		MarkForLayoutWithinDockable(f)
	}))
	p.AddChild(NewPageLabel(i18n.Text("Basic Swing")))

	// Add Lifting Strength-based damage
	p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(p.entity.LiftingThrust().String())
		MarkForLayoutWithinDockable(f)
	}))
	p.AddChild(NewPageLabel(i18n.Text("Lifting Damage")))

	// Add IQ-based damage if enabled
	if p.entity.SheetSettings.ShowIQBasedDamage {

		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			iqST := p.entity.ResolveAttributeCurrent("iq")
			iqDamage := p.entity.ThrustFor(fxp.As[int](iqST)).String()
			f.SetTitle(iqDamage)
			MarkForLayoutWithinDockable(f)
		}))
		p.AddChild(NewPageLabel(i18n.Text("IQ-Based Damage")))
	}

	// Add TK Strength-based damage if non-zero
	tkST := p.entity.TelekineticStrength()
	if tkST > 0 {
		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			tkThrust := p.entity.ThrustFor(fxp.As[int](tkST)).String()
			f.SetTitle(tkThrust)
			MarkForLayoutWithinDockable(f)
		}))
		p.AddChild(NewPageLabel(i18n.Text("TK Thrust")))

		p.AddChild(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			tkSwing := p.entity.SwingFor(fxp.As[int](tkST)).String()
			f.SetTitle(tkSwing)
			MarkForLayoutWithinDockable(f)
		}))
		p.AddChild(NewPageLabel(i18n.Text("TK Swing")))
	}
}
