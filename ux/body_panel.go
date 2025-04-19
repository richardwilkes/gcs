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
	"strings"
	"time"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/side"
)

// BodyPanel holds the contents of the body block on the sheet.
type BodyPanel struct {
	unison.Panel
	entity        *gurps.Entity
	targetMgr     *TargetMgr
	titledBorder  *TitledBorder
	row           []unison.Paneler
	sepLayoutData []*unison.FlexLayoutData
	hash          uint64
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
	p.hash = gurps.Hash64(locations)
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
		gc.DrawRect(r, colors.Header.Paint(gc, r, paintstyle.Fill))
		for i, row := range p.row {
			var ink unison.Ink
			if i&1 == 1 {
				ink = unison.ThemeBanding
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
	InstallTintFunc(p, colors.TintBody)
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
			rollRange = " "
		}
		var roll *unison.Label
		if hasSubTable || depth != 0 {
			roll = NewPageLabel(rollRange)
		} else {
			roll = NewPageLabelCenter(rollRange)
		}
		border := unison.NewEmptyBorder(unison.Insets{Left: float32(10 * depth), Bottom: 1})
		roll.SetBorder(border)
		roll.SetLayoutData(&unison.FlexLayoutData{})
		p.AddChild(roll)

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		indexes := append(bodyType.ParentIndexes(), i)
		if location.SubTable != nil {
			name := unison.NewButton()
			name.SetFocusable(false)
			name.HideBase = true
			name.HAlign = align.Start
			name.Side = side.Right
			name.HMargin = 0
			name.VMargin = 0
			name.Font = fonts.PageLabelPrimary
			name.Text = unison.NewSmallCapsText(location.TableName, &unison.TextDecoration{
				Font:            name.Font,
				OnBackgroundInk: name.OnBackgroundInk,
			})
			var rotation float32
			if location.IsOpen(indexes) {
				rotation = 90
			}
			size := max(fonts.PageLabelPrimary.Baseline()-2, 6)
			name.Drawable = &unison.DrawableSVG{
				SVG:             unison.CircledChevronRightSVG,
				Size:            unison.NewSize(size, size),
				RotationDegrees: rotation,
			}
			name.SetLayoutData(&unison.FlexLayoutData{})
			name.SetBorder(border)
			if strings.TrimSpace(location.Description) != "" {
				name.Tooltip = newWrappedTooltip(location.Description)
			}
			name.ClickCallback = func() {
				location.SetOpen(indexes, !location.IsOpen(indexes))
				p.sync(true)
				unison.InvokeTaskAfter(p.Window().UpdateCursorNow, time.Millisecond)
			}
			p.row = append(p.row, name)
			p.AddChild(name)
		} else {
			name := NewPageLabel(location.TableName)
			name.SetLayoutData(&unison.FlexLayoutData{})
			name.SetBorder(border)
			if strings.TrimSpace(location.Description) != "" {
				name.Tooltip = newWrappedTooltip(location.Description)
			}
			p.row = append(p.row, name)
			p.AddChild(name)
		}
		penalty := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
			f.SetTitle(fmt.Sprintf("%+d", location.HitPenalty))
			MarkForLayoutWithinDockable(f)
		})
		penalty.SetLayoutData(&unison.FlexLayoutData{})
		p.AddChild(penalty)

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		dr := NewNonEditablePageFieldCenter(func(f *NonEditablePageField) {
			var tooltip xio.ByteBuffer
			f.SetTitle(location.DisplayDR(p.entity, &tooltip))
			f.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("The DR covering the %s hit location%s"),
				location.TableName, tooltip.String()))
			MarkForLayoutWithinDockable(f)
		})
		dr.SetLayoutData(&unison.FlexLayoutData{})
		p.AddChild(dr)

		if i == 0 && depth == 0 {
			p.addSeparator()
		}

		title := fmt.Sprintf(i18n.Text("Notes for %s"), location.TableName)
		notesField := NewStringPageField(p.targetMgr, "body:"+location.ID(), title,
			func() string { return location.Notes }, func(value string) { location.Notes = value })
		notesField.Tooltip = newWrappedTooltip(title)
		notesField.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
		p.AddChild(notesField)

		if location.IsOpen(indexes) {
			p.addTable(location.SubTable, depth+1)
		}
	}
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
	p.sync(false)
}

func (p *BodyPanel) sync(force bool) {
	locations := gurps.SheetSettingsFor(p.entity).BodyType
	if hash := gurps.Hash64(locations); force || hash != p.hash {
		p.hash = hash
		p.titledBorder.Title = locations.Name
		p.addContent(locations)
		MarkForLayoutWithinDockable(p)
	}
}
