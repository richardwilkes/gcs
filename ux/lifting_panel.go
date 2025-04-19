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

// LiftingPanel holds the contents of the lifting block on the sheet.
type LiftingPanel struct {
	unison.Panel
	entity *gurps.Entity
}

// NewLiftingPanel creates a new lifting panel.
func NewLiftingPanel(entity *gurps.Entity) *LiftingPanel {
	p := &LiftingPanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
		HAlign:   align.Middle,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Lifting & Moving Things")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p, gc, rect, 0, 2, nil) }
	InstallTintFunc(p, colors.TintLifting)
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.BasicLift()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Basic Lift"), i18n.Text("The weight that can be lifted overhead with one hand in one second"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.OneHandedLift()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("One-Handed Lift"), i18n.Text("The weight that can be lifted overhead with one hand in two seconds"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.TwoHandedLift()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Two-Handed Lift"),
		i18n.Text("The weight that can be lifted overhead with both hands in four seconds"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.ShoveAndKnockOver()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Shove & Knock Over"), i18n.Text("The weight of an object that can be shoved and knocked over"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.RunningShoveAndKnockOver()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Running Shove & Knock Over"),
		i18n.Text("The weight of an object that can be shoved and knocked over with a running start"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.CarryOnBack()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Carry On Back"), i18n.Text("The weight that can be carried slung across the back"))
	p.addFieldAndLabel(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.ShiftSlightly()); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Shift Slightly"), i18n.Text("The weight that can be shifted slightly on a floor"))
	return p
}

func (p *LiftingPanel) addFieldAndLabel(field *NonEditablePageField, title, tooltip string) {
	field.Tooltip = newWrappedTooltip(tooltip)
	p.AddChild(field)
	label := NewPageLabel(title)
	label.Tooltip = newWrappedTooltip(tooltip)
	p.AddChild(label)
}
