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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// testPageBlock stands in for one of the sheet's block panels.
type testPageBlock struct {
	unison.Panel
}

// TestInitTitledPagePanel verifies the shared setup of a titled sheet page block: the panel becomes its own Self, gets
// the titled border with the standard insets inside it, a grid of the requested columns, layout data that fills its
// cell, the banded background and the tint only when asked for, and hands back the layout and layout data it installed
// so that a block can adjust them.
func TestInitTitledPagePanel(t *testing.T) {
	c := check.New(t)
	block := &testPageBlock{}
	layout, layoutData := initTitledPagePanel(block, "Title", 3, true, colors.TintIdentity)
	c.Equal(block, block.Self, "the block becomes its own Self")
	c.Equal(newTitledPageBorder("Title").Insets(), block.Border().Insets())
	c.Equal((&TitledBorder{Title: "Title"}).Insets().Add(titledPagePanelInsets), block.Border().Insets(),
		"the standard insets sit inside the titled border")
	c.Equal(block.Layout(), layout, "the layout handed back is the one installed")
	c.Equal(3, layout.Columns)
	c.Equal(float32(4), layout.HSpacing)
	c.Equal(block.LayoutData(), layoutData, "the layout data handed back is the one installed")
	c.Equal(align.Fill, layoutData.HAlign)
	c.Equal(align.Fill, layoutData.VAlign)
	c.False(layoutData.HGrab)
	c.NotNil(block.DrawCallback, "a banded block draws its rows")
	c.NotNil(block.DrawOverCallback, "a tinted block draws its tint")

	plain := &testPageBlock{}
	initTitledPagePanel(plain, "Plain", 2, false, nil)
	c.Nil(plain.DrawCallback, "a block that isn't banded draws nothing of its own")
	c.Nil(plain.DrawOverCallback, "a block without a tint draws no tint")
}
