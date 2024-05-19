/*
 * Copyright ©1998-2024 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// PointsPanel holds the contents of the points block on the sheet.
type PointsPanel struct {
	unison.Panel
	entity       *gurps.Entity
	targetMgr    *TargetMgr
	prefix       string
	total        *unison.Label
	ptsList      *unison.Panel
	unspentLabel *unison.Label
	overSpent    int8
}

// NewPointsPanel creates a new points panel.
func NewPointsPanel(entity *gurps.Entity, targetMgr *TargetMgr) *PointsPanel {
	p := &PointsPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{Columns: 1})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Fill,
		VSpan:  2,
		VGrab:  true,
	})

	hdr := unison.NewPanel()
	hdr.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Middle,
	})
	hdr.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	hdr.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, gurps.ThemeHeader.Paint(gc, rect, paintstyle.Fill))
	}

	hdri := unison.NewPanel()
	hdri.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	hdri.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})
	hdr.AddChild(hdri)

	var overallTotal string
	if p.entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		overallTotal = p.entity.PointsBreakdown().Total().String()
	} else {
		overallTotal = p.entity.TotalPoints.String()
	}
	p.total = unison.NewLabel()
	p.total.Font = gurps.PageLabelPrimaryFont
	p.total.Text = fmt.Sprintf(i18n.Text("%s Points"), overallTotal)
	p.total.OnBackgroundInk = gurps.OnThemeHeader
	hdri.AddChild(p.total)
	height := p.total.Font.Baseline() - 2
	editButton := unison.NewSVGButton(svg.Edit)
	editButton.OnBackgroundInk = gurps.OnThemeHeader
	editButton.OnSelectionInk = gurps.OnThemeHeader
	editButton.Font = gurps.PageLabelPrimaryFont
	editButton.Drawable.(*unison.DrawableSVG).Size = unison.NewSize(height, height)
	editButton.ClickCallback = func() {
		displayPointsEditor(unison.AncestorOrSelf[Rebuildable](p), p.entity)
	}
	hdri.AddChild(editButton)
	p.AddChild(hdr)

	p.ptsList = unison.NewPanel()
	p.ptsList.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: 4,
	})
	p.ptsList.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.End,
		VAlign: align.Fill,
		VSpan:  2,
		VGrab:  true,
	})
	p.AddChild(p.ptsList)

	p.ptsList.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(gurps.ThemeHeader, 0, unison.Insets{
		Top:    0,
		Left:   1,
		Bottom: 1,
		Right:  1,
	}, false), unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.ptsList.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) { drawBandedBackground(p.ptsList, gc, rect, 0, 2) }

	p.unspentLabel = p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.UnspentPoints().String(); text != f.Text {
			f.Text = text
			p.adjustUnspent()
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Unspent"), i18n.Text("Points earned but not yet spent"))
	p.unspentLabel.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		if p.overSpent == -1 {
			gc.DrawRect(rect, unison.ThemeError.Paint(gc, rect, paintstyle.Fill))
		}
		p.unspentLabel.DefaultDraw(gc, rect)
	}
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Ancestry.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Ancestry"), i18n.Text("Total points spent on an ancestry package"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Attributes.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Attributes"), i18n.Text("Total points spent on attributes"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Advantages.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Advantages"), i18n.Text("Total points spent on advantages"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Disadvantages.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Disadvantages"), i18n.Text("Total points spent on disadvantages"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Quirks.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Quirks"), i18n.Text("Total points spent on quirks"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Skills.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Skills"), i18n.Text("Total points spent on skills"))
	p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.PointsBreakdown().Spells.String(); text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
	}), i18n.Text("Spells"), i18n.Text("Total points spent on spells"))
	p.adjustUnspent()
	return p
}

func (p *PointsPanel) addPointsField(field *NonEditablePageField, title, tooltip string) *unison.Label {
	field.Tooltip = newWrappedTooltip(tooltip)
	p.ptsList.AddChild(field)
	label := NewPageLabel(title)
	label.Tooltip = newWrappedTooltip(tooltip)
	p.ptsList.AddChild(label)
	return label
}

func (p *PointsPanel) adjustUnspent() {
	if p.unspentLabel != nil {
		last := p.overSpent
		if p.entity.UnspentPoints() < 0 {
			if p.overSpent != -1 {
				p.overSpent = -1
				p.unspentLabel.OnBackgroundInk = unison.ThemeOnError
				p.unspentLabel.Text = i18n.Text("Overspent")
			}
		} else {
			if p.overSpent != 1 {
				p.overSpent = 1
				p.unspentLabel.OnBackgroundInk = gurps.OnThemeHeader
				p.unspentLabel.Text = i18n.Text("Unspent")
			}
		}
		if last != p.overSpent {
			MarkForLayoutWithinDockable(p)
		}
	}
}

// Sync the panel to the current data.
func (p *PointsPanel) Sync() {
	var overallTotal string
	if p.entity.SheetSettings.ExcludeUnspentPointsFromTotal {
		overallTotal = p.entity.PointsBreakdown().Total().String()
	} else {
		overallTotal = p.entity.TotalPoints.String()
	}
	p.total.Text = fmt.Sprintf(i18n.Text("%s Points"), overallTotal)
	p.MarkForLayoutAndRedraw()
}
