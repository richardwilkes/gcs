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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/role"
)

// testPageBlock stands in for one of the sheet's block panels.
type testPageBlock struct {
	unison.Panel
}

// TestInitTitledPagePanel verifies the shared setup of a titled sheet page block: the panel becomes its own Self, gets
// the titled border with the standard insets inside it, a grid of the requested columns, layout data that fills its
// cell, the banded background and the tint only when asked for, and hands back the layout and layout data installed.
func TestInitTitledPagePanel(t *testing.T) {
	c := check.New(t)
	block := &testPageBlock{}
	layout, layoutData := initTitledPagePanel(block, "Title", 3, true, colors.TintIdentity)
	c.Equal(block, block.Self, "the block becomes its own Self")
	c.Equal(newTitledBlockBorder("Title").Insets().Add(titledPagePanelInsets), block.Border().Insets(),
		"the standard insets sit inside the titled border")
	installed, ok := block.Layout().(*blockLayout)
	c.True(ok, "a titled block's layout stretches its heading across the block")
	c.Equal(installed.FlexLayout, layout, "the layout handed back is the one installed")
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

// A titled block's heading covers the painted title strip without moving the content, and follows the strip's height
// when the title's font changes.
func TestTitledPagePanelHeadingCoversTheTitle(t *testing.T) {
	c := check.New(t)
	block := &testPageBlock{}
	initTitledPagePanel(block, "Identity", 2, true, nil)
	block.AddChild(NewPageLabel("Name"))
	block.AddChild(NewPageLabel("Value"))
	children := block.Children()
	c.Equal(3, len(children), "the heading is a child of the block")
	heading := children[0]
	c.Equal(role.Heading, heading.Accessibility.Role)
	c.Equal(1, heading.Accessibility.Level)
	c.Equal("Identity", heading.Accessibility.Name)
	c.Equal("Identity", block.Accessibility.Name, "the block is still named for the focus arriving in it")
	c.Equal(heading, blockHeading(block.AsPanel()))
	c.Equal(1, blockRowsStart(block.AsPanel()), "the rows start after the heading")
	c.Equal(children[1:], blockRows(block.AsPanel()))

	layout, ok := block.Layout().(*blockLayout)
	if !ok {
		t.Fatalf("a titled block's layout must be a blockLayout, not a %T", block.Layout())
	}
	border := layout.border
	plain := &TitledBorder{Title: "Identity"}
	c.Equal(float32(1), border.Insets().Top, "the block's border leaves the title strip to the content")
	c.Equal(plain.TitleHeight()+1, plain.Insets().Top, "a border without a heading keeps the title strip")
	c.Equal(plain.Insets(), border.titleInsets(), "both paint the same area")

	_, pref, _ := block.Sizes(geom.Size{})
	block.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	block.ValidateLayout()
	strip := geom.NewRect(1, 1, pref.Width-2, border.TitleHeight())
	c.Equal(strip, border.TitleStrip(pref), "the strip lies inside the lines at the top and sides of the border")
	c.Equal(strip, heading.FrameRect(), "the heading covers the painted title strip")
	insets := block.Border().Insets()
	c.True(heading.FrameRect().X < insets.Left && heading.FrameRect().Y < insets.Top,
		"the heading reaches outside the insets the content sits within")
	c.Equal(plain.Insets().Top+titledPagePanelInsets.Top, children[1].FrameRect().Y,
		"the content starts where it did when the border kept the title strip")
	c.Equal(children[1].FrameRect().Y, children[2].FrameRect().Y, "the labels share the first content row")

	before := border.TitleHeight()
	border.Font = border.font().Face().Font(border.font().Size() * 2)
	c.True(border.TitleHeight() > before, "a larger font makes the strip taller")
	block.MarkForLayoutRecursively()
	block.ValidateLayout()
	c.Equal(border.TitleStrip(block.FrameRect().Size), heading.FrameRect(), "the heading follows the strip's height")
	c.True(children[1].FrameRect().Y >= heading.FrameRect().Y+heading.FrameRect().Height,
		"the first row is laid out below the taller strip")

	// A block without a title has no heading.
	untitled := &testPageBlock{}
	initTitledPagePanel(untitled, "", 2, true, nil)
	c.Equal(0, len(untitled.Children()))
	c.Equal(plain.TitleHeight()+1+titledPagePanelInsets.Top, untitled.Border().Insets().Top,
		"the border keeps the title strip")
	_, ok = untitled.Layout().(*unison.FlexLayout)
	c.True(ok, "the layout is the plain FlexLayout, not a %T", untitled.Layout())
	c.Nil(blockHeading(untitled.AsPanel()))
	c.Equal(0, blockRowsStart(untitled.AsPanel()), "the banded rows start with the first child")
	untitled.AddChild(NewPageLabel("Name"))
	c.Equal(untitled.Children(), blockRows(untitled.AsPanel()))
}

// The attribute and basic damage blocks rebuild their rows, which must leave the heading first, or the first row would
// be laid out under the title.
func TestRebuildingBlocksKeepTheirHeading(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.Recalculate()
	targetMgr := NewTargetMgr(unison.NewPanel())
	attrs := gurps.SheetSettingsFor(entity).Attributes
	primary := NewPrimaryAttrPanel(entity, targetMgr)
	secondary := NewSecondaryAttrPanel(entity, targetMgr)
	pools := NewPointPoolsPanel(entity, targetMgr)
	damage := NewDamagePanel(entity, targetMgr)
	for _, one := range []struct {
		name    string
		block   unison.Paneler
		rebuild func()
	}{
		{"primary attributes", primary, func() { primary.rebuild(attrs) }},
		{"secondary attributes", secondary, func() { secondary.rebuild(attrs) }},
		{"point pools", pools, func() { pools.rebuild(attrs) }},
		{"basic damage", damage, damage.rebuild},
	} {
		for _, pass := range []string{"built", "rebuilt"} {
			if pass == "rebuilt" {
				one.rebuild()
			}
			panel := one.block.AsPanel()
			children := panel.Children()
			if len(children) < 2 {
				t.Fatalf("%s %s: the block has %d children; expected the heading and its rows", one.name, pass,
					len(children))
			}
			heading := children[0]
			c.Equal(role.Heading, heading.Accessibility.Role, "%s %s: the heading comes first", one.name, pass)
			c.Equal(panel.Accessibility.Name, heading.Accessibility.Name, "%s %s: the heading carries the title", one.name, pass)
			// The sheet may give a block more width than its columns need; the heading must still span the block.
			_, pref, _ := panel.Sizes(geom.Size{})
			panel.SetFrameRect(geom.NewRect(0, 0, pref.Width*2, pref.Height))
			panel.ValidateLayout()
			content := panel.ContentRect(false)
			c.Equal(float32(1), heading.FrameRect().X, "%s %s: the heading starts inside the border's left line", one.name, pass)
			c.Equal(pref.Width*2-2, heading.FrameRect().Width, "%s %s: the heading spans the block", one.name, pass)
			bottom := heading.FrameRect().Y + heading.FrameRect().Height
			for _, child := range children[1:] {
				c.True(child.FrameRect().Y >= bottom, "%s %s: a %T at %v is laid out under the title strip, which ends at %v",
					one.name, pass, child.Self, child.FrameRect(), bottom)
				// A separator row would grab the spare width, but the factory attributes have none.
				c.True(child.FrameRect().Width < content.Width, "%s %s: the rows keep their own widths", one.name, pass)
			}
		}
	}
}
