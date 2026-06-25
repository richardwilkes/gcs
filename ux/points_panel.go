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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
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
	unspentField *NonEditablePageField
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
	hdr.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, colors.Header.Paint(gc, rect, paintstyle.Fill))
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
	p.total.Text = unison.NewSmallCapsText(fmt.Sprintf(i18n.Text("%s Points"), overallTotal), &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: colors.OnHeader,
	})
	hdri.AddChild(p.total)
	height := fonts.PageLabelPrimary.Baseline() - 2
	editButton := unison.NewSVGButton(svg.Edit)
	editButton.OnBackgroundInk = colors.OnHeader
	editButton.OnSelectionInk = colors.OnHeader
	editButton.Font = fonts.PageLabelPrimary
	if dsvg, ok := editButton.Drawable.(*unison.DrawableSVG); ok {
		dsvg.Size = geom.NewSize(height, height)
	}
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

	p.ptsList.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(colors.Header, geom.Size{}, geom.Insets{
		Top:    0,
		Left:   1,
		Bottom: 1,
		Right:  1,
	}, false), unison.NewEmptyBorder(geom.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.ptsList.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		drawBandedBackground(p.ptsList, gc, rect, 0, 2, func(rowIndex int, ink unison.Ink) unison.Ink {
			if rowIndex == 0 && p.overSpent == -1 {
				return unison.ThemeError
			}
			return ink
		})
	}

	p.unspentField = NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.UnspentPoints().String(); text != f.Text.String() {
			f.SetTitle(text)
			p.adjustUnspent()
			MarkForLayoutWithinDockable(f)
		}
	})
	p.unspentLabel = p.addPointsField(p.unspentField, i18n.Text("Unspent"), i18n.Text("Points earned but not yet spent"))
	for _, one := range []struct {
		get     func(*gurps.PointsBreakdown) fxp.Int
		title   string
		tooltip string
	}{
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Ancestry },
			title:   i18n.Text("Ancestry"),
			tooltip: i18n.Text("Total points spent on an ancestry package"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Attributes },
			title:   i18n.Text("Attributes"),
			tooltip: i18n.Text("Total points spent on attributes"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Advantages },
			title:   i18n.Text("Advantages"),
			tooltip: i18n.Text("Total points spent on advantages"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Disadvantages },
			title:   i18n.Text("Disadvantages"),
			tooltip: i18n.Text("Total points spent on disadvantages"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Quirks },
			title:   i18n.Text("Quirks"),
			tooltip: i18n.Text("Total points spent on quirks"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Skills },
			title:   i18n.Text("Skills"),
			tooltip: i18n.Text("Total points spent on skills"),
		},
		{
			get:     func(pb *gurps.PointsBreakdown) fxp.Int { return pb.Spells },
			title:   i18n.Text("Spells"),
			tooltip: i18n.Text("Total points spent on spells"),
		},
	} {
		p.addPointsField(NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			if text := one.get(p.entity.PointsBreakdown()).String(); text != f.Text.String() {
				f.SetTitle(text)
				MarkForLayoutWithinDockable(f)
			}
		}), one.title, one.tooltip)
	}
	p.adjustUnspent()

	InstallTintFunc(p, colors.TintPoints)
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
				p.unspentField.OnBackgroundInk = unison.ThemeOnError
				p.unspentField.Text.AdjustDecorations(func(decoration *unison.TextDecoration) {
					decoration.OnBackgroundInk = unison.ThemeOnError
				})
				p.unspentLabel.Text = unison.NewSmallCapsText(i18n.Text("Overspent"), &unison.TextDecoration{
					Font:            fonts.PageLabelPrimary,
					OnBackgroundInk: unison.ThemeOnError,
				})
			}
		} else {
			if p.overSpent != 1 {
				p.overSpent = 1
				p.unspentField.OnBackgroundInk = unison.DefaultLabelTheme.OnBackgroundInk
				p.unspentField.Text.AdjustDecorations(func(decoration *unison.TextDecoration) {
					decoration.OnBackgroundInk = unison.DefaultLabelTheme.OnBackgroundInk
				})
				p.unspentLabel.Text = unison.NewSmallCapsText(i18n.Text("Unspent"), &unison.TextDecoration{
					Font:            fonts.PageLabelPrimary,
					OnBackgroundInk: unison.DefaultLabelTheme.OnBackgroundInk,
				})
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
	p.total.Text = unison.NewSmallCapsText(fmt.Sprintf(i18n.Text("%s Points"), overallTotal), &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: colors.OnHeader,
	})
	p.MarkForLayoutAndRedraw()
}
