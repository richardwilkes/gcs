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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
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
	configureEditorRow(p.AsPanel(), 3, false)
	p.AddChild(NewDragHandle(editorRowDragKey, &editorRowDragData{
		editor: dockable,
		row:    p.AsPanel(),
		title:  i18n.Text("Hit Location Drag"),
		move:   func(to int) bool { return dockable.moveHitLocation(loc, to) },
	}))
	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

	return p
}

func (p *hitLocationSettingsPanel) createButtons() *unison.Panel {
	p.deleteButton = unison.NewSVGButton(unison.TrashSVG)
	p.deleteButton.ClickCallback = p.removeHitLocation
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove hit location"))
	owningTable := p.loc.OwningTable()
	p.deleteButton.SetEnabled(owningTable != nil && len(owningTable.Locations) > 1)

	p.addButton = unison.NewSVGButton(unison.CircledAddSVG)
	p.addButton.ClickCallback = p.addSubTable
	p.addButton.Tooltip = newWrappedTooltip(i18n.Text("Add sub-table"))
	p.addButton.SetEnabled(p.loc.SubTable == nil)
	return newEditorRowButtonColumn(p.deleteButton, p.addButton)
}

func (p *hitLocationSettingsPanel) addSubTable() {
	keyPrefix := p.dockable.targetMgr.NextPrefix()
	p.dockable.editStructure(i18n.Text("Add Sub-Table"), func() {
		p.loc.SubTable = &gurps.Body{
			Roll:      gurps.Roller.Parse("1d"),
			KeyPrefix: keyPrefix,
		}
		p.loc.SubTable.SetOwningLocation(p.loc)
		p.loc.SubTable.Update(p.dockable.Entity())
		p.loc.SubTable.AddLocation(gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix()))
	}, keyPrefix+"subroll")
}

func (p *hitLocationSettingsPanel) removeHitLocation() {
	p.dockable.editStructure(i18n.Text("Remove Hit Location"), func() { p.loc.OwningTable().RemoveLocation(p.loc) }, "")
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

	mgr := p.dockable.targetMgr
	addLabelAndTargetedStringField(content, mgr, p.loc.KeyPrefix+"choice_name", i18n.Text("Choice Name"),
		i18n.Text("The name of this hit location as it should appear in choice lists"), prototypeMinNameWidth,
		func() string { return p.loc.ChoiceName },
		func(s string) { p.loc.ChoiceName = s })
	addLabelAndTargetedStringField(content, mgr, p.loc.KeyPrefix+"table_name", i18n.Text("Table Name"),
		i18n.Text("The name of this hit location as it should appear in the hit location table"), prototypeMinNameWidth,
		func() string { return p.loc.TableName },
		func(s string) { p.loc.TableName = s })
	addLabelAndTargetedIntegerField(content, mgr, p.loc.KeyPrefix+"slots", i18n.Text("Slots"),
		i18n.Text("The number of consecutive numbers this hit location fills in the table"),
		func() int { return p.loc.Slots },
		func(v int) { p.loc.Slots = v }, 0, 999999, false)
	addLabelAndTargetedIntegerField(content, mgr, p.loc.KeyPrefix+"hit_penalty", i18n.Text("Hit Penalty"),
		i18n.Text("The skill adjustment for this hit location"),
		func() int { return p.loc.HitPenalty },
		func(v int) { p.loc.HitPenalty = v }, -100, 100, true)
	addLabelAndTargetedIntegerField(content, mgr, p.loc.KeyPrefix+"dr_bonus", i18n.Text("DR Bonus"),
		i18n.Text("The amount of DR this hit location grants due to natural toughness"),
		func() int { return p.loc.DRBonus },
		func(v int) { p.loc.DRBonus = v }, 0, 100, false)
	addLabelAndTargetedMultiLineStringField(content, mgr, p.loc.KeyPrefix+"desc", i18n.Text("Description"),
		i18n.Text("A description of any special effects for hits to this location"), prototypeMinNameWidth,
		func() string { return p.loc.Description },
		func(s string) { p.loc.Description = s })

	if p.loc.SubTable != nil {
		addLabelAndTargetedStringField(content, mgr, p.loc.SubTable.KeyPrefix+"subroll", i18n.Text("Sub-Roll"),
			i18n.Text("The dice to roll on the sub-table"), "100d1000",
			func() string { return gurps.Roller.Format(p.loc.SubTable.Roll) },
			func(s string) { p.loc.SubTable.Roll = gurps.Roller.Parse(s) })
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
