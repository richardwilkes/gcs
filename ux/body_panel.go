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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// BodyPanel holds the contents of the body block on the sheet.
type BodyPanel struct {
	unison.Panel
	entity        *gurps.Entity
	targetMgr     *TargetMgr
	titledBorder  *TitledBorder
	row           []unison.Paneler
	sepLayoutData []*unison.FlexLayoutData
	crc           uint64
}

// NewBodyPanel creates a new body panel.
func NewBodyPanel(entity *gurps.Entity, targetMgr *TargetMgr) *BodyPanel {
	p := &BodyPanel{
		entity:    entity,
		targetMgr: targetMgr,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  8,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		VSpan:  3,
	})
	locations := gurps.SheetSettingsFor(entity).BodyType
	p.crc = locations.CRC64()
	p.titledBorder = &TitledBorder{Title: locations.Name}
	p.SetBorder(unison.NewCompoundBorder(p.titledBorder, unison.NewEmptyBorder(unison.Insets{
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ThemeBelowSurface.Paint(gc, rect, paintstyle.Fill))
		r := p.Children()[0].FrameRect()
		r.X = rect.X
		r.Width = rect.Width
		gc.DrawRect(r, gurps.ThemeHeader.Paint(gc, r, paintstyle.Fill))
		for i, row := range p.row {
			var ink unison.Ink
			if i&1 == 1 {
				ink = unison.ThemeSurface
			} else {
				ink = unison.ThemeBelowSurface
			}
			r = row.AsPanel().FrameRect()
			r.X = rect.X
			r.Width = rect.Width
			gc.DrawRect(r, ink.Paint(gc, r, paintstyle.Fill))
		}
	}
	p.addContent(locations)
	return p
}

func (p *BodyPanel) addContent(locations *gurps.Body) {
	p.RemoveAllChildren()
	p.AddChild(NewPageHeader(i18n.Text("Roll"), 1))
	p.AddChild(unison.NewPanel())
	p.AddChild(NewPageHeader(i18n.Text("Location"), 2))
	p.AddChild(unison.NewPanel())
	header := NewPageHeader(i18n.Text("DR"), 1)
	header.Tooltip = newWrappedTooltip(i18n.Text("Damage Resistance for the hit location"))
	p.AddChild(header)
	p.AddChild(unison.NewPanel())
	header = NewPageHeader("", 1)
	header.Tooltip = newWrappedTooltip(i18n.Text("Notes for the hit location"))
	baseline := header.Font.Baseline() * 0.8
	header.Drawable = &unison.DrawableSVG{
		SVG:  svg.FirstAidKit,
		Size: unison.NewSize(baseline, baseline),
	}
	p.AddChild(header)
	p.row = nil
	p.sepLayoutData = nil
	p.addTable(locations, 0)
	for _, one := range p.sepLayoutData {
		one.VSpan = len(p.row)
	}
}

func (p *BodyPanel) addTable(bodyType *gurps.Body, depth int) {
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
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
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
			name.Tooltip = newWrappedTooltip(location.Description)
		}
		name.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
		p.row = append(p.row, name)
		p.AddChild(name)
		p.AddChild(p.createHitPenaltyField(location))

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		p.AddChild(p.createDRField(location))

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		title := fmt.Sprintf(i18n.Text("Notes for %s"), location.TableName)
		notesField := NewStringPageField(p.targetMgr, "body:"+location.ID(), title,
			func() string { return location.Notes }, func(value string) { location.Notes = value })
		notesField.Tooltip = newWrappedTooltip(title)
		p.AddChild(notesField)

		if location.SubTable != nil {
			p.addTable(location.SubTable, depth+1)
		}
	}
}

func (p *BodyPanel) createHitPenaltyField(location *gurps.HitLocation) unison.Paneler {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		f.SetTitle(fmt.Sprintf("%+d", location.HitPenalty))
		MarkForLayoutWithinDockable(f)
	})
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	return field
}

func (p *BodyPanel) createDRField(location *gurps.HitLocation) unison.Paneler {
	field := NewNonEditablePageFieldCenter(func(f *NonEditablePageField) {
		var tooltip xio.ByteBuffer
		f.SetTitle(location.DisplayDR(p.entity, &tooltip))
		f.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("The DR covering the %s hit location%s"),
			location.TableName, tooltip.String()))
		MarkForLayoutWithinDockable(f)
	})
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	return field
}

func (p *BodyPanel) addSeparator() {
	sep := unison.NewSeparator()
	sep.Vertical = true
	sep.LineInk = unison.ThemeSurfaceEdge
	layoutData := &unison.FlexLayoutData{
		HAlign: align.Middle,
		VAlign: align.Fill,
		VGrab:  true,
	}
	sep.SetLayoutData(layoutData)
	p.sepLayoutData = append(p.sepLayoutData, layoutData)
	p.AddChild(sep)
}

// Sync the panel to the current data.
func (p *BodyPanel) Sync() {
	locations := gurps.SheetSettingsFor(p.entity).BodyType
	if crc := locations.CRC64(); crc != p.crc {
		p.crc = crc
		p.titledBorder.Title = locations.Name
		p.addContent(locations)
		MarkForLayoutWithinDockable(p)
	}
}
