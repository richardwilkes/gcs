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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

const (
	prototypeMinIDWidth   = "Abcdefghijklmnopqrstuvwxyz"
	prototypeMinNameWidth = prototypeMinIDWidth + prototypeMinIDWidth
)

type hitLocationSettingsPanel struct {
	unison.Panel
	dockable     *bodySettingsDockable
	loc          *gurps.HitLocation
	addButton    *unison.Button
	deleteButton *unison.Button
}

func newHitLocationSettingsPanel(dockable *bodySettingsDockable, loc *gurps.HitLocation) *hitLocationSettingsPanel {
	p := &hitLocationSettingsPanel{
		dockable: dockable,
		loc:      loc,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing,
	}))
	p.DrawCallback = func(gc *unison.Canvas, rect unison.Rect) {
		var ink unison.Ink
		if p.Parent().IndexOfChild(p)%2 == 1 {
			ink = unison.ThemeSurface
		} else {
			ink = unison.ThemeBelowSurface
		}
		gc.DrawRect(rect, ink.Paint(gc, rect, paintstyle.Fill))
	}

	p.AddChild(NewDragHandle(map[string]any{hitLocationDragDataKey: p}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

	return p
}

func (p *hitLocationSettingsPanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})

	p.deleteButton = unison.NewSVGButton(svg.Trash)
	p.deleteButton.ClickCallback = p.removeHitLocation
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove hit location"))
	owningTable := p.loc.OwningTable()
	p.deleteButton.SetEnabled(owningTable != nil && len(owningTable.Locations) > 1)
	buttons.AddChild(p.deleteButton)

	p.addButton = unison.NewSVGButton(svg.CircledAdd)
	p.addButton.ClickCallback = p.addSubTable
	p.addButton.Tooltip = newWrappedTooltip(i18n.Text("Add sub-table"))
	p.addButton.SetEnabled(p.loc.SubTable == nil)
	buttons.AddChild(p.addButton)
	return buttons
}

func (p *hitLocationSettingsPanel) addSubTable() {
	undo := p.dockable.prepareUndo(i18n.Text("Add Sub-Table"))
	p.loc.SubTable = &gurps.Body{
		BodyData:  gurps.BodyData{Roll: dice.New("1d")},
		KeyPrefix: p.dockable.targetMgr.NextPrefix(),
	}
	p.loc.SubTable.SetOwningLocation(p.loc)
	p.loc.SubTable.Update(p.dockable.Entity())
	p.loc.SubTable.AddLocation(gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix()))
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
	if focus := p.dockable.targetMgr.Find(p.loc.SubTable.KeyPrefix + "subroll"); focus != nil {
		focus.RequestFocus()
	}
}

func (p *hitLocationSettingsPanel) removeHitLocation() {
	undo := p.dockable.prepareUndo(i18n.Text("Remove Hit Location"))
	p.loc.OwningTable().RemoveLocation(p.loc)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
}

func (p *hitLocationSettingsPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})

	text := i18n.Text("ID")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field := NewStringField(p.dockable.targetMgr, p.loc.KeyPrefix+"id", text,
		func() string { return p.loc.LocID },
		func(s string) {
			if p.validateLocID(s) {
				p.loc.SetID(strings.TrimSpace(strings.ToLower(s)))
			}
		})
	field.ValidateCallback = func(field *StringField, _ *gurps.HitLocation) func() bool {
		return func() bool { return p.validateLocID(field.Text()) }
	}(field, p.loc)
	field.SetMinimumTextWidthUsing(prototypeMinIDWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("An ID for the hit location"))
	content.AddChild(field)

	text = i18n.Text("Choice Name")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field = NewStringField(p.dockable.targetMgr, p.loc.KeyPrefix+"choice_name", text,
		func() string { return p.loc.ChoiceName },
		func(s string) { p.loc.ChoiceName = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("The name of this hit location as it should appear in choice lists"))
	content.AddChild(field)

	text = i18n.Text("Table Name")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field = NewStringField(p.dockable.targetMgr, p.loc.KeyPrefix+"table_name", text,
		func() string { return p.loc.TableName },
		func(s string) { p.loc.TableName = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("The name of this hit location as it should appear in the hit location table"))
	content.AddChild(field)

	text = i18n.Text("Slots")
	content.AddChild(NewFieldLeadingLabel(text, false))
	intField := NewIntegerField(p.dockable.targetMgr, p.loc.KeyPrefix+"slots", text,
		func() int { return p.loc.Slots },
		func(v int) { p.loc.Slots = v },
		0, 999999, false, false)
	intField.Tooltip = newWrappedTooltip(i18n.Text("The number of consecutive numbers this hit location fills in the table"))
	content.AddChild(intField)

	text = i18n.Text("Hit Penalty")
	content.AddChild(NewFieldLeadingLabel(text, false))
	intField = NewIntegerField(p.dockable.targetMgr, p.loc.KeyPrefix+"hit_penalty", text,
		func() int { return p.loc.HitPenalty },
		func(v int) { p.loc.HitPenalty = v },
		-100, 100, true, false)
	intField.Tooltip = newWrappedTooltip(i18n.Text("The skill adjustment for this hit location"))
	content.AddChild(intField)

	text = i18n.Text("DR Bonus")
	content.AddChild(NewFieldLeadingLabel(text, false))
	intField = NewIntegerField(p.dockable.targetMgr, p.loc.KeyPrefix+"dr_bonus", text,
		func() int { return p.loc.DRBonus },
		func(v int) { p.loc.DRBonus = v },
		0, 100, false, false)
	intField.Tooltip = newWrappedTooltip(i18n.Text("The amount of DR this hit location grants due to natural toughness"))
	content.AddChild(intField)

	text = i18n.Text("Description")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field = NewMultiLineStringField(p.dockable.targetMgr, p.loc.KeyPrefix+"desc", text,
		func() string { return p.loc.Description },
		func(s string) { p.loc.Description = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("A description of any special effects for hits to this location"))
	content.AddChild(field)

	if p.loc.SubTable != nil {
		text = i18n.Text("Sub-Roll")
		content.AddChild(NewFieldLeadingLabel(text, false))
		field = NewStringField(p.dockable.targetMgr, p.loc.SubTable.KeyPrefix+"subroll", text,
			func() string { return p.loc.SubTable.Roll.String() },
			func(s string) { p.loc.SubTable.Roll = dice.New(s) })
		field.SetMinimumTextWidthUsing("100d1000")
		field.Tooltip = newWrappedTooltip(i18n.Text("The dice to roll on the sub-table"))
		content.AddChild(field)

		content.AddChild(newBodySettingsSubTablePanel(p.dockable, p.loc.SubTable))
	}
	return content
}

func (p *hitLocationSettingsPanel) validateLocID(locID string) bool {
	if key := strings.TrimSpace(strings.ToLower(locID)); key != "" {
		return key == gurps.SanitizeID(key, false, gurps.ReservedIDs...)
	}
	return false
}
