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
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio/fs"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ FileBackedDockable = &Campaign{}
	_ unison.TabCloser   = &Campaign{}
	_ KeyedDockable      = &Campaign{}
)

// Campaign holds the view for a GURPS campaign.
type Campaign struct {
	unison.Panel
	path              string
	toolbar           *unison.Panel
	scroll            *unison.ScrollPanel
	content           *unison.Panel
	campaign          *gurps.Campaign
	hash              uint64
	scale             int
	needsSaveAsPrompt bool
}

// NewCampaignFromFile loads a GURPS campaign file and creates a new unison.Dockable for it.
func NewCampaignFromFile(filePath string) (unison.Dockable, error) {
	campaign, err := gurps.NewCampaignFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	t := NewCampaign(filePath, campaign)
	t.needsSaveAsPrompt = false
	return t, nil
}

// NewCampaign creates a new unison.Dockable for GURPS campaign files.
func NewCampaign(filePath string, campaign *gurps.Campaign) *Campaign {
	c := &Campaign{
		path:              filePath,
		scroll:            unison.NewScrollPanel(),
		campaign:          campaign,
		hash:              gurps.Hash64(campaign),
		scale:             gurps.GlobalSettings().General.InitialEditorUIScale,
		needsSaveAsPrompt: true,
	}
	c.Self = c
	c.SetLayout(&unison.FlexLayout{
		Columns: 1,
		HAlign:  align.Fill,
		VAlign:  align.Fill,
	})
	c.MouseDownCallback = func(_ unison.Point, _, _ int, _ unison.Modifiers) bool {
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

	helpButton := unison.NewSVGButton(svg.Help)
	helpButton.Tooltip = newWrappedTooltip(i18n.Text("Help"))
	helpButton.ClickCallback = func() { HandleLink(nil, "md:Help/Interface/Campaign") }

	c.toolbar = unison.NewPanel()
	c.toolbar.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, 0, unison.Insets{Bottom: 1},
		false), unison.NewEmptyBorder(unison.StdInsets())))
	c.toolbar.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	c.toolbar.AddChild(NewDefaultInfoPop())
	c.toolbar.AddChild(helpButton)
	c.toolbar.AddChild(
		NewScaleField(
			gurps.InitialUIScaleMin,
			gurps.InitialUIScaleMax,
			func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
			func() int { return c.scale },
			func(scale int) { c.scale = scale },
			nil,
			false,
			c.scroll,
		),
	)
	c.toolbar.SetLayout(&unison.FlexLayout{
		Columns:  len(c.toolbar.Children()),
		HSpacing: unison.StdHSpacing,
	})

	c.AddChild(c.toolbar)
	c.AddChild(c.scroll)
	return c
}

// DockKey implements KeyedDockable.
func (c *Campaign) DockKey() string {
	return filePrefix + c.path
}

func (c *Campaign) createContent() unison.Paneler {
	c.content = unison.NewPanel()
	return c.content
}

// TitleIcon implements workspace.FileBackedDockable
func (c *Campaign) TitleIcon(suggestedSize unison.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  gurps.FileInfoFor(c.path).SVG,
		Size: suggestedSize,
	}
}

// Title implements workspace.FileBackedDockable
func (c *Campaign) Title() string {
	return fs.BaseName(c.path)
}

func (c *Campaign) String() string {
	return c.Title()
}

// Tooltip implements workspace.FileBackedDockable
func (c *Campaign) Tooltip() string {
	return c.path
}

// Modified implements workspace.FileBackedDockable
func (c *Campaign) Modified() bool {
	return c.hash != gurps.Hash64(c.campaign)
}

// MayAttemptClose implements unison.TabCloser
func (c *Campaign) MayAttemptClose() bool {
	return MayAttemptCloseOfGroup(c)
}

// AttemptClose implements unison.TabCloser
func (c *Campaign) AttemptClose() bool {
	if AttemptSaveForDockable(c) {
		return AttemptCloseForDockable(c)
	}
	return false
}

// BackingFilePath implements workspace.FileBackedDockable
func (c *Campaign) BackingFilePath() string {
	return c.path
}

// SetBackingFilePath implements workspace.FileBackedDockable
func (c *Campaign) SetBackingFilePath(p string) {
	c.path = p
	UpdateTitleForDockable(c)
}

func (c *Campaign) save(forceSaveAs bool) bool {
	success := false
	if forceSaveAs || c.needsSaveAsPrompt {
		success = SaveDockableAs(c, gurps.CampaignExt, c.campaign.Save, func(path string) {
			c.hash = gurps.Hash64(c.campaign)
			c.path = path
		})
	} else {
		success = SaveDockable(c, c.campaign.Save, func() { c.hash = gurps.Hash64(c.campaign) })
	}
	if success {
		c.needsSaveAsPrompt = false
	}
	return success
}
