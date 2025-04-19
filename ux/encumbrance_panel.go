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
	"fmt"
	"strconv"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var _ unison.ColorProvider = &encRowColor{}

// EncumbrancePanel holds the contents of the encumbrance block on the sheet.
type EncumbrancePanel struct {
	unison.Panel
	entity     *gurps.Entity
	row        []unison.Paneler
	current    int
	overloaded bool
}

// NewEncumbrancePanel creates a new encumbrance panel.
func NewEncumbrancePanel(entity *gurps.Entity) *EncumbrancePanel {
	p := &EncumbrancePanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  9,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	p.SetBorder(&TitledBorder{Title: i18n.Text("Encumbrance, Move & Dodge")})
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		r := p.Children()[0].FrameRect()
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, colors.Header.Paint(gc, r, paintstyle.Fill))
		p.current = int(entity.EncumbranceLevel(false))
		p.overloaded = entity.WeightCarried(false) > entity.MaximumCarry(encumbrance.ExtraHeavy)
		for i, row := range p.row {
			var ink unison.Ink
			switch {
			case p.current == i:
				if p.overloaded {
					ink = unison.ThemeWarning
				} else {
					ink = unison.ThemeFocus
				}
			case i&1 == 1:
				ink = unison.ThemeBanding
			default:
				ink = unison.ThemeBelowSurface
			}
			r = row.AsPanel().FrameRect()
			r.X = rect.X
			r.Width = rect.Width
			gc.DrawRect(r, ink.Paint(gc, r, paintstyle.Fill))
		}
	}

	p.AddChild(NewPageHeader(i18n.Text("Level"), 3))
	p.AddChild(unison.NewPanel())
	p.AddChild(NewPageHeader(i18n.Text("Max Load"), 1))
	p.AddChild(unison.NewPanel())
	p.AddChild(NewPageHeader(i18n.Text("Move"), 1))
	p.AddChild(unison.NewPanel())
	p.AddChild(NewPageHeader(i18n.Text("Dodge"), 1))

	for i, enc := range encumbrance.Levels {
		rowColor := &encRowColor{
			owner: p,
			index: i,
		}
		p.AddChild(p.createMarker(entity, enc, rowColor))
		p.AddChild(p.createLevelField(enc, rowColor))
		name := NewPageLabelWithInk(enc.String(), rowColor)
		name.SetLayoutData(&unison.FlexLayoutData{
			HAlign: align.Fill,
			VAlign: align.Middle,
			HGrab:  true,
		})
		p.AddChild(name)
		p.row = append(p.row, name)
		if i == 0 {
			p.addSeparator()
		}
		p.AddChild(p.createMaxCarryField(enc, rowColor))
		if i == 0 {
			p.addSeparator()
		}
		p.AddChild(p.createMoveField(enc, rowColor))
		if i == 0 {
			p.addSeparator()
		}
		p.AddChild(p.createDodgeField(enc, rowColor))
	}

	InstallTintFunc(p, colors.TintEncumbrance)
	return p
}

func (p *EncumbrancePanel) createMarker(entity *gurps.Entity, enc encumbrance.Level, rowColor *encRowColor) *unison.Label {
	marker := unison.NewLabel()
	marker.Font = fonts.PageLabelPrimary
	marker.OnBackgroundInk = rowColor
	marker.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	marker.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: 4}))
	baseline := marker.Font.Baseline()
	marker.Drawable = &unison.DrawableSVG{
		SVG:  svg.Weight,
		Size: unison.Size{Width: baseline, Height: baseline},
	}
	marker.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		if enc == entity.EncumbranceLevel(false) {
			marker.DefaultDraw(gc, rect)
		}
	}
	return marker
}

func (p *EncumbrancePanel) createLevelField(enc encumbrance.Level, rowColor *encRowColor) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(_ *NonEditablePageField) {})
	field.OnBackgroundInk = rowColor
	field.SetTitle(strconv.Itoa(int(enc)))
	field.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
	})
	return field
}

func (p *EncumbrancePanel) createMaxCarryField(enc encumbrance.Level, rowColor *encRowColor) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := p.entity.SheetSettings.DefaultWeightUnits.Format(p.entity.MaximumCarry(enc)); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	})
	field.OnBackgroundInk = rowColor
	field.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("The maximum load that can be carried and still remain within the %s encumbrance level"), enc.String()))
	field.Text.AdjustDecorations(func(d *unison.TextDecoration) { d.OnBackgroundInk = field.OnBackgroundInk })
	return field
}

func (p *EncumbrancePanel) createMoveField(enc encumbrance.Level, rowColor *encRowColor) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := strconv.Itoa(p.entity.Move(enc)); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	})
	field.OnBackgroundInk = rowColor
	field.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("The ground movement rate for the %s encumbrance level"), enc.String()))
	field.Text.AdjustDecorations(func(d *unison.TextDecoration) { d.OnBackgroundInk = field.OnBackgroundInk })
	return field
}

func (p *EncumbrancePanel) createDodgeField(enc encumbrance.Level, rowColor *encRowColor) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := strconv.Itoa(p.entity.Dodge(enc)); text != f.Text.String() {
			f.SetTitle(text)
			MarkForLayoutWithinDockable(f)
		}
	})
	field.OnBackgroundInk = rowColor
	field.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("The dodge for the %s encumbrance level"), enc.String()))
	field.SetBorder(unison.NewEmptyBorder(unison.Insets{Right: 4}))
	field.Text.AdjustDecorations(func(d *unison.TextDecoration) { d.OnBackgroundInk = field.OnBackgroundInk })
	return field
}

func (p *EncumbrancePanel) addSeparator() {
	sep := unison.NewSeparator()
	sep.Vertical = true
	sep.LineInk = unison.ThemeSurfaceEdge
	sep.SetLayoutData(&unison.FlexLayoutData{
		VSpan:  len(encumbrance.Levels),
		HAlign: align.Middle,
		VAlign: align.Fill,
		VGrab:  true,
	})
	p.AddChild(sep)
}

type encRowColor struct {
	owner *EncumbrancePanel
	index int
}

func (c *encRowColor) GetColor() unison.Color {
	switch c.owner.current {
	case c.index:
		if c.owner.overloaded {
			return unison.ThemeOnWarning.GetColor()
		}
		return unison.ThemeOnFocus.GetColor()
	default:
		return unison.ThemeOnSurface.GetColor()
	}
}

func (c *encRowColor) Paint(canvas *unison.Canvas, rect unison.Rect, style paintstyle.Enum) *unison.Paint {
	return c.GetColor().Paint(canvas, rect, style)
}
