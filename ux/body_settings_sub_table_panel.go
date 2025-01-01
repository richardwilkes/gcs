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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

type bodySettingsSubTablePanel struct {
	unison.Panel
	dockable     *bodySettingsDockable
	body         *gurps.Body
	addButton    *unison.Button
	deleteButton *unison.Button
}

func newBodySettingsSubTablePanel(d *bodySettingsDockable, body *gurps.Body) *bodySettingsSubTablePanel {
	p := &bodySettingsSubTablePanel{
		dockable: d,
		body:     body,
	}
	p.Self = p
	p.SetLayout(&unison.FlexLayout{
		Columns:  2,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HSpan:  2,
		HAlign: align.Fill,
	})

	p.AddChild(p.createButtons())
	p.AddChild(p.createContent())

	return p
}

func (p *bodySettingsSubTablePanel) createButtons() *unison.Panel {
	buttons := unison.NewPanel()
	buttons.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	buttons.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Middle})

	p.deleteButton = unison.NewSVGButton(svg.Trash)
	p.deleteButton.ClickCallback = p.removeSubTable
	p.deleteButton.Tooltip = newWrappedTooltip(i18n.Text("Remove sub-table"))
	buttons.AddChild(p.deleteButton)

	p.addButton = unison.NewSVGButton(svg.CircledAdd)
	p.addButton.ClickCallback = p.addHitLocation
	p.addButton.Tooltip = newWrappedTooltip(i18n.Text("Add hit location"))
	buttons.AddChild(p.addButton)
	return buttons
}

func (p *bodySettingsSubTablePanel) addHitLocation() {
	undo := p.dockable.prepareUndo(i18n.Text("Add Hit Location"))
	location := gurps.NewHitLocation(p.dockable.Entity(), p.dockable.targetMgr.NextPrefix())
	p.body.AddLocation(location)
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
	if focus := p.dockable.targetMgr.Find(location.KeyPrefix + "id"); focus != nil {
		focus.RequestFocus()
	}
}

func (p *bodySettingsSubTablePanel) removeSubTable() {
	undo := p.dockable.prepareUndo(i18n.Text("Remove Sub-Table"))
	p.body.OwningLocation().SubTable = nil
	p.dockable.finishAndPostUndo(undo)
	p.dockable.sync()
}

func (p *bodySettingsSubTablePanel) createContent() *unison.Panel {
	content := unison.NewPanel()
	content.SetLayout(&unison.FlexLayout{
		Columns:  1,
		HSpacing: unison.StdHSpacing,
		VSpacing: unison.StdVSpacing,
	})
	content.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill})
	content.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.NewUniformInsets(1), false))

	for _, loc := range p.body.Locations {
		content.AddChild(newHitLocationSettingsPanel(p.dockable, loc))
	}

	return content
}
