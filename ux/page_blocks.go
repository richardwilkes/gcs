// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
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

func createPageTopBlock(entity *gurps.Entity, targetMgr *TargetMgr) (page *Page, modifiedFunc func()) {
	page = NewPage(entity)
	var top *unison.Panel
	top, modifiedFunc = createPageFirstRow(entity, targetMgr)
	page.AddChild(top)
	page.AddChild(createPageSecondRow(entity, targetMgr))
	return page, modifiedFunc
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

func createPageSecondRow(entity *gurps.Entity, targetMgr *TargetMgr) *unison.Panel {
	p := unison.NewPanel()
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

	p.AddChild(NewPrimaryAttrPanel(entity, targetMgr))
	p.AddChild(NewSecondaryAttrPanel(entity, targetMgr))
	p.AddChild(NewBodyPanel(entity, targetMgr))
	p.AddChild(endWrapper)
	p.AddChild(NewDamagePanel(entity))
	p.AddChild(NewPointPoolsPanel(entity, targetMgr))

	return p
}
