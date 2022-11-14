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

package ux

import (
	"fmt"
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison"
)

// BodyPanel holds the contents of the body block on the sheet.
type BodyPanel struct {
	unison.Panel
	entity        *model.Entity
	titledBorder  *TitledBorder
	row           []unison.Paneler
	sepLayoutData []*unison.FlexLayoutData
	crc           uint64
}

// NewBodyPanel creates a new body panel.
func NewBodyPanel(entity *model.Entity) *BodyPanel {
	p := &BodyPanel{entity: entity}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  6,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		VSpan:  3,
	})
	locations := model.SheetSettingsFor(entity).BodyType
	p.crc = locations.CRC64()
	p.titledBorder = &TitledBorder{Title: locations.Name}
	p.SetBorder(unison.NewCompoundBorder(p.titledBorder, unison.NewEmptyBorder(unison.Insets{
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, unison.Fill))
		r := p.Children()[0].FrameRect()
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, theme.HeaderColor.Paint(gc, r, unison.Fill))
		for i, row := range p.row {
			var ink unison.Ink
			if i&1 == 1 {
				ink = unison.BandingColor
			} else {
				ink = unison.ContentColor
			}
			r = row.AsPanel().FrameRect()
			r.X = rect.X
			r.Width = rect.Width
			gc.DrawRect(r, ink.Paint(gc, r, unison.Fill))
		}
	}
	p.addContent(locations)
	return p
}

func (p *BodyPanel) addContent(locations *model.Body) {
	p.RemoveAllChildren()
	p.AddChild(NewPageHeader(i18n.Text("Roll"), 1))
	p.AddChild(NewInteriorSeparator())
	p.AddChild(NewPageHeader(i18n.Text("Location"), 2))
	p.AddChild(NewInteriorSeparator())
	p.AddChild(NewPageHeader(i18n.Text("DR"), 1))
	p.row = nil
	p.sepLayoutData = nil
	p.addTable(locations, 0)
	for _, one := range p.sepLayoutData {
		one.VSpan = len(p.row)
	}
}

func (p *BodyPanel) addTable(bodyType *model.Body, depth int) {
	hasSubTable := false
	if depth == 0 {
		for _, location := range bodyType.Locations {
			if location.SubTable != nil {
				hasSubTable = true
				break
			}
		}
	}
	for i, location := range bodyType.Locations {
		rollRange := location.RollRange
		if rollRange == "-" {
			rollRange = ""
		}
		var label *unison.Label
		if hasSubTable || depth != 0 {
			label = NewPageLabel(rollRange)
		} else {
			label = NewPageLabelCenter(rollRange)
		}
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
		if depth > 0 {
			label.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: float32(10 * depth)}))
		}
		p.AddChild(label)

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		name := NewPageLabel(location.TableName)
		if depth > 0 {
			name.SetBorder(unison.NewEmptyBorder(unison.Insets{Left: float32(10 * depth)}))
		}
		if strings.TrimSpace(location.Description) != "" {
			name.Tooltip = unison.NewTooltipWithText(location.Description)
		}
		name.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
		p.row = append(p.row, name)
		p.AddChild(name)
		p.AddChild(p.createHitPenaltyField(location))

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		p.AddChild(p.createDRField(location))

		if location.SubTable != nil {
			p.addTable(location.SubTable, depth+1)
		}
	}
}

func (p *BodyPanel) createHitPenaltyField(location *model.HitLocation) unison.Paneler {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.Text = fmt.Sprintf("%+d", location.HitPenalty)
		MarkForLayoutWithinDockable(f)
	})
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	return field
}

func (p *BodyPanel) createDRField(location *model.HitLocation) unison.Paneler {
	field := NewNonEditablePageFieldCenter(func(f *NonEditablePageField) {
		var tooltip xio.ByteBuffer
		f.Text = location.DisplayDR(p.entity, &tooltip)
		f.Tooltip = unison.NewTooltipWithText(fmt.Sprintf(i18n.Text("The DR covering the %s hit location%s"),
			location.TableName, tooltip.String()))
		MarkForLayoutWithinDockable(f)
	})
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.FillAlignment})
	return field
}

func (p *BodyPanel) addSeparator() {
	sep := unison.NewSeparator()
	sep.Vertical = true
	sep.LineInk = unison.InteriorDividerColor
	layoutData := &unison.FlexLayoutData{
		HAlign: unison.MiddleAlignment,
		VAlign: unison.FillAlignment,
		VGrab:  true,
	}
	sep.SetLayoutData(layoutData)
	p.sepLayoutData = append(p.sepLayoutData, layoutData)
	p.AddChild(sep)
}

// Sync the panel to the current data.
func (p *BodyPanel) Sync() {
	locations := model.SheetSettingsFor(p.entity).BodyType
	if crc := locations.CRC64(); crc != p.crc {
		p.crc = crc
		p.titledBorder.Title = locations.Name
		p.addContent(locations)
		MarkForLayoutWithinDockable(p)
	}
}
