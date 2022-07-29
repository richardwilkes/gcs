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
	"github.com/richardwilkes/gcs/v5/model/gurps/attribute"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/unison"
)

// PrimaryAttrPanel holds the contents of the primary attributes block on the sheet.
type PrimaryAttrPanel struct {
	unison.Panel
	entity *gurps.Entity
	crc    uint64
}

// NewPrimaryAttrPanel creates a new primary attributes panel.
func NewPrimaryAttrPanel(entity *gurps.Entity) *PrimaryAttrPanel {
	p := &PrimaryAttrPanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
	})
	p.SetBorder(unison.NewCompoundBorder(&widget.TitledBorder{Title: i18n.Text("Primary Attributes")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
	}
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	p.crc = attrs.CRC64()
	p.rebuild(attrs)
	return p
}

func (p *PrimaryAttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	p.RemoveAllChildren()
	for _, def := range attrs.List(false) {
		if def.Type == attribute.Pool || !def.Primary() {
			continue
		}
		if def.Type == attribute.PrimarySeparator {
			p.AddChild(newPageInternalHeader(def.Name, 3))
		} else {
			attr, ok := p.entity.Attributes.Set[def.ID()]
			if !ok {
				jot.Warnf("unable to locate attribute data for '%s'", def.ID())
				continue
			}
			p.AddChild(p.createPointsField(attr))
			p.AddChild(p.createValueField(def, attr))
			p.AddChild(widget.NewPageLabel(def.CombinedName()))
		}
	}
}

func newPageInternalHeader(title string, span int) unison.Paneler {
	layoutData := &unison.FlexLayoutData{
		HSpan:  span,
		HAlign: unison.FillAlignment,
	}
	border := unison.NewEmptyBorder(unison.NewVerticalInsets(2))
	if title == "" {
		sep := unison.NewSeparator()
		sep.SetBorder(border)
		sep.SetLayoutData(layoutData)
		return sep
	}
	label := unison.NewLabel()
	label.Text = title
	label.Font = theme.PageLabelSecondaryFont
	label.HAlign = unison.MiddleAlignment
	label.OnBackgroundInk = theme.OnPageColor
	label.SetLayoutData(layoutData)
	label.SetBorder(border)
	label.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		_, pref, _ := label.Sizes(unison.Size{})
		paint := unison.DividerColor.Paint(gc, rect, unison.Stroke)
		paint.SetStrokeWidth(1)
		half := (rect.Width - pref.Width) / 2
		gc.DrawLine(rect.X, rect.CenterY(), rect.X+half-2, rect.CenterY(), paint)
		gc.DrawLine(2+rect.Right()-half, rect.CenterY(), rect.Right(), rect.CenterY(), paint)
		label.DefaultDraw(gc, rect)
	}
	return label
}

func (p *PrimaryAttrPanel) createPointsField(attr *gurps.Attribute) *widget.NonEditablePageField {
	field := widget.NewNonEditablePageFieldEnd(func(f *widget.NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text {
			f.Text = text
			widget.MarkForLayoutWithinDockable(f)
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = theme.PageFieldSecondaryFont
	return field
}

func (p *PrimaryAttrPanel) createValueField(def *gurps.AttributeDef, attr *gurps.Attribute) *widget.DecimalField {
	field := widget.NewDecimalPageField(nil, "", def.CombinedName(),
		func() fxp.Int { return attr.Maximum() },
		func(v fxp.Int) { attr.SetMaximum(v) }, fxp.Min, fxp.Max, true)
	return field
}

// Sync the panel to the current data.
func (p *PrimaryAttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	if crc := attrs.CRC64(); crc != p.crc {
		p.crc = crc
		p.rebuild(attrs)
		widget.MarkForLayoutWithinDockable(p)
	}
}
