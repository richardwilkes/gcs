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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

const hitLocationDragDataKey = "drag.body"

type bodySettingsPanel struct {
	unison.Panel
	dockable *bodySettingsDockable
}

func newBodySettingsPanel(d *bodySettingsDockable) *bodySettingsPanel {
	p := &bodySettingsPanel{
		dockable: d,
	}
	p.Self = p
	p.SetBorder(unison.NewEmptyBorder(unison.Insets{
		Top:    unison.StdVSpacing,
		Left:   unison.StdHSpacing,
		Bottom: unison.StdVSpacing,
		Right:  unison.StdHSpacing * 2,
	}))
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})

	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

	return p
}

func (p *bodySettingsPanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})

	addButton := unison.NewSVGButton(svg.CircledAdd)
	addButton.ClickCallback = p.addHitLocation
	addButton.Tooltip = newWrappedTooltip(i18n.Text("Add hit location"))
	buttons.AddChild(addButton)
	return buttons
}

func (p *bodySettingsPanel) addHitLocation() {
	undo := p.dockable.prepareUndo(i18n.Text("Add Hit Location"))
	location := gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix())
	p.dockable.body.AddLocation(location)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
	if focus := p.dockable.targetMgr.Find(location.KeyPrefix + "id"); focus != nil {
		focus.RequestFocus()
	}
}

func (p *bodySettingsPanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})

	text := i18n.Text("Name")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field := NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"name", text,
		func() string { return p.dockable.body.Name },
		func(s string) { p.dockable.body.Name = s })
	field.SetMinimumTextWidthUsing(prototypeMinNameWidth)
	field.Tooltip = newWrappedTooltip(i18n.Text("The name of this body type"))
	content.AddChild(field)

	text = i18n.Text("Roll")
	content.AddChild(NewFieldLeadingLabel(text, false))
	field = NewStringField(p.dockable.targetMgr, p.dockable.body.KeyPrefix+"roll", text,
		func() string { return p.dockable.body.Roll.String() },
		func(s string) { p.dockable.body.Roll = dice.New(s) })
	field.SetMinimumTextWidthUsing("100d1000")
	field.Tooltip = newWrappedTooltip(i18n.Text("The dice to roll on the table"))
	content.AddChild(field)

	wrapper := unison.NewPanel()
	wrapper.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))
	wrapper.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
	})
	wrapper.SetLayout(&unison.FlexLayout{Columns: 1})
	content.AddChild(wrapper)

	for _, loc := range p.dockable.body.Locations {
		wrapper.AddChild(newHitLocationSettingsPanel(p.dockable, loc))
	}
	return content
}
