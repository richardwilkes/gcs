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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/mod"
)

var (
	_ FileBackedDockable = &Campaign{}
	_ unison.TabCloser   = &Campaign{}
	_ KeyedDockable      = &Campaign{}
)

// Campaign holds the view for a GURPS campaign.
type Campaign struct {
	fileBackedPanel
	toolbar  *unison.Panel
	scroll   *unison.ScrollPanel
	content  *unison.Panel
	campaign *gurps.Campaign
	scale    int
}

// NewCampaignFromFile loads a GURPS campaign file and creates a new unison.Dockable for it.
func NewCampaignFromFile(filePath string) (unison.Dockable, error) {
	return openDockableFromFile(filePath, gurps.NewCampaignFromFile, NewCampaign)
}

// NewCampaign creates a new unison.Dockable for GURPS campaign files.
func NewCampaign(filePath string, campaign *gurps.Campaign) *Campaign {
	c := &Campaign{
		scroll:   unison.NewScrollPanel(),
		campaign: campaign,
		scale:    gurps.GlobalSettings().General.InitialEditorUIScale,
	}
	c.Self = c
	c.initFileEditor(c, filePath, gurps.CampaignExt, campaign.Save, campaign)
	c.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})
	c.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
		c.RequestFocus()
		return false
	}
	c.scroll.SetContent(c.createContent(), behavior.Unmodified, behavior.Unmodified)
	c.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	c.toolbar = newToolbar()
	c.toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(c.toolbar, func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
		func() int { return c.scale }, func(scale int) { c.scale = scale }, false, c.scroll)
	finishToolbarLayout(c.toolbar)

	c.AddChild(c.toolbar)
	c.AddChild(c.scroll)
	return c
}

func (c *Campaign) createContent() unison.Paneler {
	c.content = unison.NewPanel()
	return c.content
}
