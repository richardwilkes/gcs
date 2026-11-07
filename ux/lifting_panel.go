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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// LiftingPanel holds the contents of the lifting block on the sheet.
type LiftingPanel struct {
	unison.Panel
	entity   *gurps.Entity
	fields   []*NonEditablePageField
	tooltips []string
	// adjustments holds the lines naming the bonuses that adjust lifting ST that the tooltips of the fields were last
	// built with.
	adjustments string
}

// NewLiftingPanel creates a new lifting panel.
func NewLiftingPanel(entity *gurps.Entity) *LiftingPanel {
	p := &LiftingPanel{entity: entity}
	layout, layoutData := initTitledPagePanel(p, i18n.Text("Lifting & Moving Things"), 2, true, colors.TintLifting)
	layout.HAlign = align.Middle
	layoutData.HGrab = true
	layoutData.VGrab = true
	for _, one := range []struct {
		get     func() fxp.Weight
		title   string
		tooltip string
	}{
		{
			get:     entity.BasicLift,
			title:   i18n.Text("Basic Lift"),
			tooltip: i18n.Text("The weight that can be lifted overhead with one hand in one second"),
		},
		{
			get:     entity.OneHandedLift,
			title:   i18n.Text("One-Handed Lift"),
			tooltip: i18n.Text("The weight that can be lifted overhead with one hand in two seconds"),
		},
		{
			get:     entity.TwoHandedLift,
			title:   i18n.Text("Two-Handed Lift"),
			tooltip: i18n.Text("The weight that can be lifted overhead with both hands in four seconds"),
		},
		{
			get:     entity.ShoveAndKnockOver,
			title:   i18n.Text("Shove & Knock Over"),
			tooltip: i18n.Text("The weight of an object that can be shoved and knocked over"),
		},
		{
			get:     entity.RunningShoveAndKnockOver,
			title:   i18n.Text("Running Shove & Knock Over"),
			tooltip: i18n.Text("The weight of an object that can be shoved and knocked over with a running start"),
		},
		{
			get:     entity.CarryOnBack,
			title:   i18n.Text("Carry On Back"),
			tooltip: i18n.Text("The weight that can be carried slung across the back"),
		},
		{
			get:     entity.ShiftSlightly,
			title:   i18n.Text("Shift Slightly"),
			tooltip: i18n.Text("The weight that can be shifted slightly on a floor"),
		},
	} {
		p.addFieldAndLabel(NewNonEditablePageFieldEndFor(func() string {
			return p.entity.SheetSettings.DefaultWeightUnits.Format(one.get())
		}), one.title, one.tooltip)
	}
	p.Sync()
	return p
}

// Sync implements Syncer. The fields keep their own values current, so this rebuilds their tooltips when the bonuses
// that adjust lifting ST, which those values are figured from, change.
func (p *LiftingPanel) Sync() {
	adjustments := p.entity.CurrentAdjustmentTooltip(gurps.StrengthID, stlimit.LiftingOnly)
	if adjustments == p.adjustments {
		return
	}
	p.adjustments = adjustments
	for i, field := range p.fields {
		field.Tooltip = newWrappedTooltip(adjustedTooltip(p.tooltips[i], adjustments))
	}
}

func (p *LiftingPanel) addFieldAndLabel(field *NonEditablePageField, title, tooltip string) {
	field.Tooltip = newWrappedTooltip(tooltip)
	p.fields = append(p.fields, field)
	p.tooltips = append(p.tooltips, tooltip)
	p.AddChild(field)
	label := NewPageLabel(title)
	label.Tooltip = newWrappedTooltip(tooltip)
	p.AddChild(label)
}
