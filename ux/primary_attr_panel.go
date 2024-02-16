/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// PrimaryAttrPanel holds the contents of the primary attributes block on the sheet.
type PrimaryAttrPanel struct {
	unison.Panel
	entity    *gurps.Entity
	targetMgr *TargetMgr
	prefix    string
	crc       uint64
}

// NewPrimaryAttrPanel creates a new primary attributes panel.
func NewPrimaryAttrPanel(entity *gurps.Entity, targetMgr *TargetMgr) *PrimaryAttrPanel {
	p := &PrimaryAttrPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Primary Attributes")}, unison.NewEmptyBorder(unison.Insets{
		Top:    1,
		Left:   2,
		Bottom: 1,
		Right:  2,
	})))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		gc.DrawRect(rect, unison.ContentColor.Paint(gc, rect, paintstyle.Fill))
	}
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	p.crc = attrs.CRC64()
	p.rebuild(attrs)
	return p
}

func (p *PrimaryAttrPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := p.targetMgr.CurrentFocusRef()
	p.RemoveAllChildren()
	for _, def := range attrs.List(false) {
		if def.Primary() {
			if def.Type == attribute.PrimarySeparator {
				p.AddChild(NewPageInternalHeader(def.CombinedName(), 3))
			} else {
				attr, ok := p.entity.Attributes.Set[def.ID()]
				if !ok {
					errs.Log(errs.New("unable to locate attribute data"), "id", def.ID())
					continue
				}
				if def.Type == attribute.IntegerRef || def.Type == attribute.DecimalRef {
					field := NewNonEditablePageFieldEnd(func(field *NonEditablePageField) {
						field.Text = attr.Maximum().String()
					})
					field.SetLayoutData(&unison.FlexLayoutData{
						HSpan:  2,
						HAlign: align.Fill,
						VAlign: align.Middle,
					})
					p.AddChild(field)
				} else {
					p.AddChild(p.createPointsField(attr))
					p.AddChild(p.createValueField(def, attr))
				}
				p.AddChild(NewPageLabel(def.CombinedName()))
			}
		}
	}
	if p.targetMgr != nil {
		if sheet := unison.Ancestor[*Sheet](p); sheet != nil {
			p.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (p *PrimaryAttrPanel) createPointsField(attr *gurps.Attribute) *NonEditablePageField {
	field := NewNonEditablePageFieldEnd(func(f *NonEditablePageField) {
		if text := "[" + attr.PointCost().String() + "]"; text != f.Text {
			f.Text = text
			MarkForLayoutWithinDockable(f)
		}
		if def := attr.AttributeDef(); def != nil {
			f.Tooltip = newWrappedTooltip(fmt.Sprintf(i18n.Text("Points spent on %s"), def.CombinedName()))
		}
	})
	field.Font = gurps.PageFieldSecondaryFont
	return field
}

func (p *PrimaryAttrPanel) createValueField(def *gurps.AttributeDef, attr *gurps.Attribute) *DecimalField {
	return NewDecimalPageField(p.targetMgr, p.prefix+attr.AttrID, def.CombinedName(),
		func() fxp.Int { return attr.Maximum() },
		func(v fxp.Int) { attr.SetMaximum(v) }, fxp.Min, fxp.Max, true)
}

// Sync the panel to the current data.
func (p *PrimaryAttrPanel) Sync() {
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	if crc := attrs.CRC64(); crc != p.crc {
		p.crc = crc
		p.rebuild(attrs)
		MarkForLayoutWithinDockable(p)
	}
}
