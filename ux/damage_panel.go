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
	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// DamagePanel holds the contents of the damage block on the sheet.
type DamagePanel struct {
	unison.Panel
	entity              *gurps.Entity
	targetMgr           *TargetMgr
	showLiftingSTDamage bool
	showEffectDamage    bool
	showTKDamage        bool
}

// NewDamagePanel creates a new damage panel.
func NewDamagePanel(entity *gurps.Entity, targetMgr *TargetMgr) *DamagePanel {
	p := &DamagePanel{
		entity:    entity,
		targetMgr: targetMgr,
	}
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
	InstallTintFunc(p, colors.TintDamage)
	p.rebuild()
	return p
}

func (p *DamagePanel) rebuild() {
	p.RemoveAllChildren()
	p.addDamageField(i18n.Text("Basic Thrust"), func() string { return p.entity.Thrust().String() })
	p.addDamageField(i18n.Text("Basic Swing"), func() string { return p.entity.Swing().String() })
	if p.showLiftingSTDamage = p.entity.SheetSettings.ShowLiftingSTDamage; p.showLiftingSTDamage {
		p.addDamageField(i18n.Text("Lifting Thrust"), func() string { return p.entity.LiftingThrust().String() })
		p.addDamageField(i18n.Text("Lifting Swing"), func() string { return p.entity.LiftingSwing().String() })
	}
	if p.showEffectDamage = p.entity.SheetSettings.ShowIQBasedDamage; p.showEffectDamage {
		p.addDamageField(i18n.Text("IQ-based Thrust"), func() string { return p.entity.IQThrust().String() })
		p.addDamageField(i18n.Text("IQ-based Swing"), func() string { return p.entity.IQSwing().String() })
	}
	if p.showTKDamage = p.entity.TelekineticStrength() > 0; p.showTKDamage {
		p.addDamageField(i18n.Text("Telekinetic Thrust"), func() string { return p.entity.TelekineticThrust().String() })
		p.addDamageField(i18n.Text("Telekinetic Swing"), func() string { return p.entity.TelekineticSwing().String() })
	}
}

func (p *DamagePanel) addDamageField(title string, f func() string) {
	p.AddChild(NewNonEditablePageFieldEnd(func(field *NonEditablePageField) {
		field.SetTitle(f())
		MarkForLayoutWithinDockable(field)
	}))
	p.AddChild(NewPageLabel(title))
}

// Sync the panel to the current data.
func (p *DamagePanel) Sync() {
	if p.showLiftingSTDamage != p.entity.SheetSettings.ShowLiftingSTDamage ||
		p.showEffectDamage != p.entity.SheetSettings.ShowIQBasedDamage ||
		p.showTKDamage != (p.entity.TelekineticStrength() > 0) {
		p.forceSync()
	}
}

func (p *DamagePanel) forceSync() {
	p.rebuild()
	MarkForLayoutWithinDockable(p)
}
