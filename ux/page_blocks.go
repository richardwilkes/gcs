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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

func createPageTopBlock(entity *gurps.Entity, targetMgr *TargetMgr) (page *Page, modifiedFunc, syncDisclosureFunc func()) {
	page = NewPage(entity)
	var top *unison.Panel
	top, modifiedFunc = createPageFirstRow(entity, targetMgr)
	page.AddChild(top)
	var bottom *unison.Panel
	bottom, syncDisclosureFunc = createPageSecondRow(entity, targetMgr)
	page.AddChild(bottom)
	return page, modifiedFunc, syncDisclosureFunc
}

func createPageFirstRow(entity *gurps.Entity, targetMgr *TargetMgr) (top *unison.Panel, modifiedFunc func()) {
	right := unison.NewPanel()
	right.SetLayout(&unison.FlexLayout{
		Columns:  3,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	right.AddChild(NewIdentityPanel(entity, targetMgr))
	miscPanel := NewMiscPanel(entity, targetMgr)
	right.AddChild(miscPanel)
	right.AddChild(NewPointsPanel(entity, targetMgr))
	right.AddChild(NewDescriptionPanel(entity, targetMgr))

	top = unison.NewPanel()
	portraitPanel := NewPortraitPanel(entity)
	top.SetLayout(&portraitLayout{
		portrait: portraitPanel,
		rest:     right,
	})
	top.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})
	top.AddChild(portraitPanel)
	top.AddChild(right)

	return top, miscPanel.UpdateModified
}

func createPageSecondRow(entity *gurps.Entity, targetMgr *TargetMgr) (p *unison.Panel, syncDisclosureFunc func()) {
	p = unison.NewPanel()
	p.SetLayout(&unison.FlexLayout{
		Columns:  4,
		HSpacing: 1,
		VSpacing: 1,
		HAlign:   align.Fill,
		VAlign:   align.Fill,
	})
	p.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
	})

	endWrapper := unison.NewPanel()
	endWrapper.SetLayout(&unison.FlexLayout{
		Columns:  1,
		VSpacing: 1,
	})
	endWrapper.SetLayoutData(&unison.FlexLayoutData{
		VSpan:  3,
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	endWrapper.AddChild(NewEncumbrancePanel(entity))
	endWrapper.AddChild(NewLiftingPanel(entity))

	primaryAttrPanel := NewPrimaryAttrPanel(entity, targetMgr)
	p.AddChild(primaryAttrPanel)
	secondaryAttrPanel := NewSecondaryAttrPanel(entity, targetMgr)
	p.AddChild(secondaryAttrPanel)
	bodyPanel := NewBodyPanel(entity, targetMgr)
	p.AddChild(bodyPanel)
	p.AddChild(endWrapper)
	damagePanel := NewDamagePanel(entity, targetMgr)
	p.AddChild(damagePanel)
	poolPanel := NewPointPoolsPanel(entity, targetMgr)
	p.AddChild(poolPanel)

	return p, func() {
		primaryAttrPanel.forceSync()
		secondaryAttrPanel.forceSync()
		poolPanel.forceSync()
		damagePanel.forceSync()
		bodyPanel.sync(true)
	}
}
