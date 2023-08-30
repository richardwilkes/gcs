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
	"github.com/richardwilkes/toolbox/errs"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// PointPoolsPanel holds the contents of the point pools block on the sheet.
type PointPoolsPanel struct {
	unison.Panel
	entity      *gurps.Entity
	targetMgr   *TargetMgr
	stateLabels map[string]*unison.Label
	prefix      string
	crc         uint64
}

// NewPointPoolsPanel creates a new point pools panel.
func NewPointPoolsPanel(entity *gurps.Entity, targetMgr *TargetMgr) *PointPoolsPanel {
	p := &PointPoolsPanel{
		entity:    entity,
		targetMgr: targetMgr,
		prefix:    targetMgr.NextPrefix(),
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  6,
		HSpacing: 4,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		VAlign: unison.FillAlignment,
		HSpan:  2,
	})
	p.SetBorder(unison.NewCompoundBorder(&TitledBorder{Title: i18n.Text("Point Pools")}, unison.NewEmptyBorder(unison.Insets{
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

func (p *PointPoolsPanel) rebuild(attrs *gurps.AttributeDefs) {
	focusRefKey := p.targetMgr.CurrentFocusRef()
	p.RemoveAllChildren()
	p.stateLabels = make(map[string]*unison.Label)
	for _, def := range attrs.List(false) {
		if def.Pool() {
			if def.Type == gurps.PoolSeparatorAttributeType {
				p.AddChild(NewPageInternalHeader(def.Name, 6))
			} else {
				attr, ok := p.entity.Attributes.Set[def.ID()]
				if !ok {
					errs.Log(errs.New("unable to locate attribute data"), "id", def.ID())
					continue
				}
				p.AddChild(p.createPointsField(attr))

				var currentField *DecimalField
				currentField = NewDecimalPageField(p.targetMgr, p.prefix+attr.AttrID+":cur",
					i18n.Text("Point Pool Current"), func() fxp.Int {
						if currentField != nil {
							currentField.SetMinMax(currentField.Min(), attr.Maximum())
						}
						return attr.Current()
					},
					func(v fxp.Int) { attr.Damage = (attr.Maximum() - v).Max(0) }, fxp.Min, attr.Maximum(), true)
				p.AddChild(currentField)

				p.AddChild(NewPageLabel(i18n.Text("of")))

				maximumField := NewDecimalPageField(p.targetMgr, p.prefix+attr.AttrID+":max",
					i18n.Text("Point Pool Maximum"), func() fxp.Int { return attr.Maximum() },
					func(v fxp.Int) {
						attr.SetMaximum(v)
						currentField.SetMinMax(currentField.Min(), v)
						currentField.Sync()
					}, fxp.Min, fxp.Max, true)
				p.AddChild(maximumField)

				name := NewPageLabel(def.Name)
				if def.FullName != "" {
					name.Tooltip = newWrappedTooltip(def.FullName)
				}
				p.AddChild(name)

				if threshold := attr.CurrentThreshold(); threshold != nil {
					state := NewPageLabel("[" + threshold.State + "]")
					if threshold.Explanation != "" {
						state.Tooltip = newWrappedTooltip(threshold.Explanation)
					}
					p.AddChild(state)
					p.stateLabels[def.ID()] = state
				} else {
					p.AddChild(unison.NewPanel())
				}
			}
		}
	}
	if p.targetMgr != nil {
		if sheet := unison.Ancestor[*Sheet](p); sheet != nil {
			p.targetMgr.ReacquireFocus(focusRefKey, sheet.toolbar, sheet.scroll.Content())
		}
	}
}

func (p *PointPoolsPanel) createPointsField(attr *gurps.Attribute) *NonEditablePageField {
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

// Sync the panel to the current data.
func (p *PointPoolsPanel) Sync() {
	attrs := gurps.SheetSettingsFor(p.entity).Attributes
	if crc := attrs.CRC64(); crc != p.crc {
		p.crc = crc
		p.rebuild(attrs)
		MarkForLayoutWithinDockable(p)
	} else {
		for _, def := range attrs.List(false) {
			if def.Pool() && def.Type != gurps.PoolSeparatorAttributeType {
				id := def.ID()
				if label, exists := p.stateLabels[id]; exists {
					if attr, ok := p.entity.Attributes.Set[id]; ok {
						if threshold := attr.CurrentThreshold(); threshold != nil {
							label.Text = "[" + threshold.State + "]"
							if threshold.Explanation != "" {
								label.Tooltip = newWrappedTooltip(threshold.Explanation)
							}
						} else {
							label.Text = ""
							label.Tooltip = nil
						}
					}
				}
			}
		}
		MarkForLayoutWithinDockable(p)
	}
}
