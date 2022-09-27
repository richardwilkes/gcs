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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// PointsPanel holds the contents of the points block on the sheet.
type PointsPanel struct {
	unison.Panel
	entity       *gurps.Entity
	targetMgr    *widget.TargetMgr
	prefix       string
	pointsBorder *widget.TitledBorder
	unspent      *widget.DecimalField
}

// NewPointsPanel creates a new points panel.
func NewPointsPanel(entity *gurps.Entity, targetMgr *widget.TargetMgr) *PointsPanel {
	p := &PointsPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.EndAlignment,
		VAlign: unison.FillAlignment,
		VSpan:  2,
	})
	var overallTotal string
	if p.entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		overallTotal = p.entity.SpentPoints().String()
	} else {
		overallTotal = p.entity.TotalPoints.String()
	}
	p.pointsBorder = &widget.TitledBorder{Title: fmt.Sprintf(i18n.Text("%s Points"), overallTotal)}
	p.SetBorder(unison.NewCompoundBorder(p.pointsBorder, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p, gc, rect, 0, 2) }

	p.unspent = widget.NewDecimalPageField(p.targetMgr, p.prefix+"unspent", i18n.Text("Unspent Points"),
		func() fxp.Int { return p.entity.UnspentPoints() },
		func(v fxp.Int) { p.entity.SetUnspentPoints(v) }, fxp.Min, fxp.Max, true)
	p.unspent.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.MiddleAlignment,
	})
	p.unspent.Tooltip = unison.NewTooltipWithText(i18n.Text("Points earned but not yet spent"))
	p.AddChild(p.unspent)
	p.AddChild(widget.NewPageLabel(i18n.Text("Unspent")))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		_, _, race, _ := p.entity.TraitPoints()
		if text := race.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Race"), i18n.Text("Total points spent on a racial package"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		if text := p.entity.AttributePoints().String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Attributes"), i18n.Text("Total points spent on attributes"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		ad, _, _, _ := p.entity.TraitPoints()
		if text := ad.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Advantages"), i18n.Text("Total points spent on advantages"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		_, disad, _, _ := p.entity.TraitPoints()
		if text := disad.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Disadvantages"), i18n.Text("Total points spent on disadvantages"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		_, _, _, quirk := p.entity.TraitPoints()
		if text := quirk.String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Quirks"), i18n.Text("Total points spent on quirks"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		if text := p.entity.SkillPoints().String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Skills"), i18n.Text("Total points spent on skills"))
	p.addPointsField(widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		if text := p.entity.SpellPoints().String(); text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Spells"), i18n.Text("Total points spent on spells"))

	return p
}

func (p *PointsPanel) addPointsField(field *widget.NonEditablePageField, title, tooltip string) {
	field.Tooltip = unison.NewTooltipWithText(tooltip)
	p.AddChild(field)
	label := widget.NewPageLabel(title)
	label.Tooltip = unison.NewTooltipWithText(tooltip)
	p.AddChild(label)
}

// Sync the panel to the current data.
func (p *PointsPanel) Sync() {
	p.unspent.Sync()
	var overallTotal string
	if p.entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		overallTotal = p.entity.SpentPoints().String()
	} else {
		overallTotal = p.entity.TotalPoints.String()
	}
	p.pointsBorder.Title = fmt.Sprintf(i18n.Text("%s Points"), overallTotal)
}
