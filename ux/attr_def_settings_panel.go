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
	"strings"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/id"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type attrDefSettingsPanel struct {
	unison.Panel
	dockable           *attributeSettingsDockable
	def                *model.AttributeDef
	deleteButton       *unison.Button
	addThresholdButton *unison.Button
	poolPanel          *poolSettingsPanel
}

func newAttrDefSettingsPanel(dockable *attributeSettingsDockable, def *model.AttributeDef) *attrDefSettingsPanel {
	p := &attrDefSettingsPanel{
		dockable: dockable,
		def:      def,
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
	}))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		color := unison.ContentColor
		if p.Parent().IndexOfChild(p)%2 == 1 {
			color = unison.BandingColor
		}
		gc.DrawRect(rect, color.Paint(gc, rect, unison.Fill))
	}
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	p.AddChild(NewDragHandle(map[string]any{attributeSettingsDragDataKey: &attributeSettingsDragData{
		owner: dockable.Entity(),
		def:   def,
	}}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())
	return p
}

func (p *attrDefSettingsPanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{HAlign: unison.MiddleAlignment})

	p.deleteButton = unison.NewSVGButton(svg.Trash)
	p.deleteButton.ClickCallback = p.deleteAttrDef
	p.deleteButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Remove attribute"))
	buttons.AddChild(p.deleteButton)

	p.addThresholdButton = unison.NewSVGButton(svg.CircledAdd)
	p.addThresholdButton.ClickCallback = func() { p.poolPanel.addThreshold() }
	p.addThresholdButton.Tooltip = unison.NewTooltipWithText(i18n.Text("Add pool threshold"))
	p.addThresholdButton.SetEnabled(p.def.Type == model.PoolAttributeType)
	buttons.AddChild(p.addThresholdButton)
	return buttons
}

func (p *attrDefSettingsPanel) deleteAttrDef() {
	attrPanel := p.Parent()
	p.RemoveFromParent()
	children := attrPanel.Children()
	if len(children) == 1 {
		children[0].Self.(*attrDefSettingsPanel).deleteButton.SetEnabled(false)
	}
	undo := &unison.UndoEdit[*model.AttributeDefs]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Delete Attribute"),
		UndoFunc:   func(e *unison.UndoEdit[*model.AttributeDefs]) { p.dockable.applyAttrDefs(e.BeforeData) },
		RedoFunc:   func(e *unison.UndoEdit[*model.AttributeDefs]) { p.dockable.applyAttrDefs(e.AfterData) },
		AbsorbFunc: func(e *unison.UndoEdit[*model.AttributeDefs], other unison.Undoable) bool { return false },
	}
	undo.BeforeData = p.dockable.defs.Clone()
	delete(p.dockable.defs.Set, p.def.DefID)
	undo.AfterData = p.dockable.defs.Clone()
	p.dockable.UndoManager().Add(undo)
	p.dockable.MarkModified(nil)
}

func (p *attrDefSettingsPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})

	text := i18n.Text("ID")
	content.AddChild(NewFieldLeadingLabel(text))
	field := NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"id", text,
		func() string { return p.def.DefID },
		func(s string) {
			if p.validateAttrID(s) {
				delete(p.dockable.defs.Set, p.def.DefID)
				p.def.DefID = strings.TrimSpace(strings.ToLower(s))
				p.dockable.defs.Set[p.def.DefID] = p.def
			}
		})
	field.ValidateCallback = func(field *StringField, def *model.AttributeDef) func() bool {
		return func() bool { return p.validateAttrID(field.Text()) }
	}(field, p.def)
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.Tooltip = unison.NewTooltipWithText(i18n.Text("A unique ID for the attribute"))
	content.AddChild(field)

	if p.def.IsSeparator() {
		text = i18n.Text("Name")
		content.AddChild(NewFieldLeadingLabel(text))
		field = NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"name", text,
			func() string { return p.def.Name },
			func(s string) { p.def.Name = s })
		field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("A title to use with the separator"))
		content.AddChild(field)
	} else {
		text = i18n.Text("Short Name")
		content.AddChild(NewFieldLeadingLabel(text))
		field = NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"name", text,
			func() string { return p.def.Name },
			func(s string) { p.def.Name = s })
		field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The name of this attribute, often an abbreviation"))
		content.AddChild(field)

		text = i18n.Text("Full Name")
		content.AddChild(NewFieldLeadingLabel(text))
		field = NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"fullname", text,
			func() string { return p.def.FullName },
			func(s string) { p.def.FullName = s })
		field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The full name of this attribute (may be omitted, in which case the Short Name will be used instead)"))
		content.AddChild(field)

		text = i18n.Text("Base Value")
		content.AddChild(NewFieldLeadingLabel(text))
		field = NewStringField(p.dockable.targetMgr, p.def.KeyPrefix+"base", text,
			func() string { return p.def.AttributeBase },
			func(s string) { p.def.AttributeBase = s })
		field.SetMinimumTextWidthUsing("floor($basic_speed)")
		field.Tooltip = unison.NewTooltipWithText(i18n.Text("The base value, which may be a number or a formula"))
		content.AddChild(field)

		text = i18n.Text("Cost per Point")
		content.AddChild(NewFieldLeadingLabel(text))
		numField := NewIntegerField(p.dockable.targetMgr, p.def.KeyPrefix+"cost", text,
			func() int { return fxp.As[int](p.def.CostPerPoint) },
			func(v int) { p.def.CostPerPoint = fxp.From(v) },
			0, 9999, false, false)
		numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The cost per point difference from the base"))
		content.AddChild(numField)

		text = i18n.Text("SM Reduction")
		content.AddChild(NewFieldLeadingLabel(text))
		numField = NewPercentageField(p.dockable.targetMgr, p.def.KeyPrefix+"sm", text,
			func() int { return fxp.As[int](p.def.CostAdjPercentPerSM) },
			func(v int) { p.def.CostAdjPercentPerSM = fxp.From(v) },
			0, 80, false, false)
		numField.Tooltip = unison.NewTooltipWithText(i18n.Text("The reduction in cost for each SM greater than 0"))
		content.AddChild(numField)
	}

	text = i18n.Text("Attribute Type")
	content.AddChild(NewFieldLeadingLabel(text))
	content.AddChild(NewPopup[model.AttributeType](p.dockable.targetMgr, p.def.KeyPrefix+"type", text,
		func() model.AttributeType { return p.def.Type },
		func(typ model.AttributeType) { p.applyAttributeType(typ) },
		model.AllAttributeType...))

	if p.def.Type == model.PoolAttributeType {
		p.poolPanel = newPoolSettingsPanel(p.dockable, p.def)
		content.AddChild(p.poolPanel)
	} else {
		p.poolPanel = nil
	}
	return content
}

func (p *attrDefSettingsPanel) validateAttrID(attrID string) bool {
	if key := strings.TrimSpace(strings.ToLower(attrID)); key != "" {
		if key != id.Sanitize(key, false, model.ReservedIDs...) {
			return false
		}
		if key == p.def.DefID {
			return true
		}
		_, exists := p.dockable.defs.Set[key]
		return !exists
	}
	return false
}

func (p *attrDefSettingsPanel) applyAttributeType(attrType model.AttributeType) {
	p.def.Type = attrType
	if p.def.Type == model.PoolAttributeType && len(p.def.Thresholds) == 0 {
		p.def.Thresholds = append(p.def.Thresholds, &model.PoolThreshold{KeyPrefix: p.dockable.targetMgr.NextPrefix()})
	} else if p.def.IsSeparator() {
		p.def.FullName = ""
		p.def.AttributeBase = ""
		p.def.CostPerPoint = 0
		p.def.CostAdjPercentPerSM = 0
		p.def.Thresholds = nil
	}
	p.dockable.sync()
}
