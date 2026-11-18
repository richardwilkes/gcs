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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

func testBlockNode(key string, weight fxp.Int) *gurps.SheetLayoutNode {
	return &gurps.SheetLayoutNode{Type: layoutnode.Block, Key: key, Weight: weight}
}

func testContainerNode(nodeType layoutnode.Type, weight fxp.Int,
	children ...*gurps.SheetLayoutNode,
) *gurps.SheetLayoutNode {
	return &gurps.SheetLayoutNode{Type: nodeType, Weight: weight, Children: children}
}

// testLeafFunc returns a leaf function that hands back a fresh panel for each of the given keys and nil for anything
// else, along with the map of the panels it made.
func testLeafFunc(keys ...string) (leaf layoutLeafFunc, panelsByKey map[string]*unison.Panel) {
	panels := make(map[string]*unison.Panel, len(keys))
	for _, key := range keys {
		panels[key] = newSizedPanel(10, 40, 20)
	}
	return func(key string) unison.Paneler {
		if panel, exists := panels[key]; exists {
			return panel
		}
		return nil
	}, panels
}

func layoutNodeOf(p unison.Paneler) *gurps.SheetLayoutNode {
	node, ok := p.AsPanel().ClientData()[sheetLayoutNodeKey].(*gurps.SheetLayoutNode)
	if !ok {
		return nil
	}
	return node
}

func flexDataOf(p unison.Paneler) *unison.FlexLayoutData {
	data, ok := p.AsPanel().LayoutData().(*unison.FlexLayoutData)
	if !ok {
		return nil
	}
	return data
}

// TestBuildLayoutBandsBuildsOneBandPerRootChild checks that each of the root's bands becomes a panel of its own, laid
// out the way its node calls for, with the weights of a row's surviving children handed to the row's layout.
func TestBuildLayoutBandsBuildsOneBandPerRootChild(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockTraitsKey, gurps.BlockSkillsKey, gurps.BlockNotesKey)
	root := testContainerNode(layoutnode.Column, fxp.One,
		testContainerNode(layoutnode.Row, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.Two),
			testBlockNode(gurps.BlockSkillsKey, fxp.Three),
		),
		testBlockNode(gurps.BlockNotesKey, fxp.One),
	)
	bands := buildLayoutBands(root, leaf)
	c.Equal(2, len(bands), "one band per child of the root")

	row, ok := bands[0].Layout().(*weightedRowLayout)
	c.True(ok, "a row band is laid out by the weighted row layout")
	c.Equal([]fxp.Int{fxp.Two, fxp.Three}, row.weights, "the weights of the surviving children are handed to the row")
	c.Equal(float32(1), row.hSpacing)
	c.Equal(2, len(bands[0].Children()))
	c.Equal(panels[gurps.BlockTraitsKey], bands[0].Children()[0])
	c.Equal(panels[gurps.BlockSkillsKey], bands[0].Children()[1])
	c.Equal(panels[gurps.BlockNotesKey], bands[1], "a band that is a single block is that block's panel")

	c.Nil(buildLayoutBands(nil, leaf), "there is nothing to build without a root")
	c.Nil(buildLayoutNode(nil, leaf))
}

// TestBuildLayoutColumnChildrenFillTheRow checks that the blocks stacked in a column standing beside something taller
// share out the height they are given, so the bottom of the last of them is the bottom of the row rather than the page
// showing through below it.
func TestBuildLayoutColumnChildrenFillTheRow(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockEncumbranceKey, gurps.BlockLiftingKey, gurps.BlockBodyKey)
	panels[gurps.BlockBodyKey] = newSizedPanel(10, 40, 200)
	built := buildLayoutNode(testContainerNode(layoutnode.Row, fxp.One,
		testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockEncumbranceKey, fxp.One),
			testBlockNode(gurps.BlockLiftingKey, fxp.One),
		),
		testBlockNode(gurps.BlockBodyKey, fxp.One),
	), leaf)
	c.NotNil(built)
	c.True(flexDataOf(panels[gurps.BlockEncumbranceKey]).VGrab, "a block stacked in a column must grab height")
	c.True(flexDataOf(panels[gurps.BlockLiftingKey]).VGrab)
	c.False(flexDataOf(built).VGrab, "the row itself must not, since a band would then fill an exported page")

	row := built.AsPanel()
	_, prefSize, _ := row.Sizes(geom.NewSize(200, 0))
	c.Equal(float32(200), prefSize.Height, "the row is as tall as the tallest thing in it")
	row.SetFrameRect(geom.NewRect(0, 0, 200, prefSize.Height))
	row.MarkForLayoutRecursively()
	row.ValidateLayout()

	column := row.Children()[0]
	c.Equal(prefSize.Height, column.FrameRect().Height, "the column is given the full height of the row")
	stacked := column.Children()
	c.Equal(2, len(stacked))
	c.Equal(column.FrameRect().Height, stacked[0].FrameRect().Height+stacked[1].FrameRect().Height+1,
		"the stacked blocks and the space between them fill the column")
	c.Equal(column.FrameRect().Bottom(), stacked[1].FrameRect().Bottom(),
		"the bottom of the last block is the bottom of the row")
}

// TestBuildLayoutNodePrunes checks that a block the leaf function has nothing for is left out, that a container all of
// whose children were left out goes with them, and that a container left holding a single child is replaced by it.
func TestBuildLayoutNodePrunes(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockTraitsKey)

	c.Nil(buildLayoutNode(testBlockNode(gurps.BlockSkillsKey, fxp.One), leaf),
		"a block with nothing to show is not built")
	c.Nil(buildLayoutNode(testContainerNode(layoutnode.Row, fxp.One,
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
		testBlockNode(gurps.BlockSpellsKey, fxp.One),
	), leaf), "a container with nothing left in it is not built")

	rowNode := testContainerNode(layoutnode.Row, fxp.Two,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	)
	rowNode.MinHeight = paper.Length{Length: 1, Units: paper.Inch}
	built := buildLayoutNode(rowNode, leaf)
	c.Equal(panels[gurps.BlockTraitsKey], built, "the only child left takes the container's place")
	c.Equal(rowNode, layoutNodeOf(built), "the child records the node that now governs its slot")
	c.Equal(float32(72), flexDataOf(built).MinSize.Height, "and takes on the container's minimum height")
}

// TestBuildLayoutNodeLayoutData checks that the builder is the sole authority on the layout data of the panels it
// returns: each of them fills its slot, carries the minimum height its node calls for, and records that node.
func TestBuildLayoutNodeLayoutData(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockTraitsKey, gurps.BlockSkillsKey)
	panels[gurps.BlockTraitsKey].SetLayoutData(&unison.FlexLayoutData{HSpan: 3, VSpan: 2, HAlign: align.End})
	traitsNode := testBlockNode(gurps.BlockTraitsKey, fxp.One)
	traitsNode.MinHeight = paper.Length{Length: 0.5, Units: paper.Inch}
	skillsNode := testBlockNode(gurps.BlockSkillsKey, fxp.One)
	rowNode := testContainerNode(layoutnode.Row, fxp.One, traitsNode, skillsNode)

	built := buildLayoutNode(rowNode, leaf)
	c.NotNil(built)
	c.Equal(rowNode, layoutNodeOf(built))
	c.Equal(&unison.FlexLayoutData{HAlign: align.Fill, VAlign: align.Fill, HGrab: true}, flexDataOf(built))

	traits := flexDataOf(panels[gurps.BlockTraitsKey])
	c.Equal(align.Fill, traits.HAlign, "the leaf's own layout data does not survive")
	c.Equal(align.Fill, traits.VAlign)
	c.True(traits.HGrab)
	c.Equal(0, traits.HSpan, "the spans a leaf used in the fixed grids are gone")
	c.Equal(0, traits.VSpan)
	c.Equal(geom.Size{Height: 36}, traits.MinSize, "the minimum height comes from the node")
	c.Equal(traitsNode, layoutNodeOf(panels[gurps.BlockTraitsKey]))
	c.Equal(geom.Size{}, flexDataOf(panels[gurps.BlockSkillsKey]).MinSize, "a block with no minimum height has none")
	c.Equal(skillsNode, layoutNodeOf(panels[gurps.BlockSkillsKey]))
}

// TestBuildLayoutNodeSquareFlags checks that a row is told which of its children take their width from its height, and
// that the square flag travels the way the layout node does: a container left holding a single block is dropped, and
// the block that takes its place keeps the flag of its own node rather than the container's.
func TestBuildLayoutNodeSquareFlags(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockPortraitKey, gurps.BlockIdentityKey, gurps.BlockNotesKey)
	portraitNode := testBlockNode(gurps.BlockPortraitKey, fxp.One)
	portraitNode.Square = true
	// The column holds the portrait and a block with nothing to show, so it is dropped and the portrait takes its
	// place.
	rowNode := testContainerNode(layoutnode.Row, fxp.One,
		testContainerNode(layoutnode.Column, fxp.One, portraitNode,
			testBlockNode(gurps.BlockSpellsKey, fxp.One)),
		testBlockNode(gurps.BlockIdentityKey, fxp.One),
		testBlockNode(gurps.BlockNotesKey, fxp.One),
	)

	built := buildLayoutNode(rowNode, leaf)
	c.NotNil(built)
	row, ok := built.AsPanel().Layout().(*weightedRowLayout)
	c.True(ok, "a row band is laid out by the weighted row layout")
	c.Equal([]bool{true, false, false}, row.square, "the surviving block must keep its own square flag")
	c.True(row.squareOf(0))
	c.False(row.squareOf(1))
	c.Equal(portraitNode, squareLayoutNodeOf(panels[gurps.BlockPortraitKey]),
		"the panel must name the Block node the flag lives on, not the container it stood in for")
	c.Nil(squareLayoutNodeOf(panels[gurps.BlockIdentityKey]), "a block that isn't square must name nothing")

	// The blocks that aren't lists are built once and used again by every rebuild, so the flag has to come off again.
	portraitNode.Square = false
	built = buildLayoutNode(rowNode, leaf)
	c.NotNil(built)
	row, ok = built.AsPanel().Layout().(*weightedRowLayout)
	c.True(ok)
	c.Equal([]bool{false, false, false}, row.square, "a flag taken off the model must come off the panel with it")
	c.Nil(squareLayoutNodeOf(panels[gurps.BlockPortraitKey]))
}

func isEmptyDerivedList(sheet *Sheet, key string) bool {
	switch key {
	case gurps.BlockReactionsKey:
		return sheet.Reactions.Table.RootRowCount() == 0
	case gurps.BlockConditionalModifiersKey:
		return sheet.ConditionalModifiers.Table.RootRowCount() == 0
	case gurps.BlockMeleeKey:
		return sheet.MeleeWeapons.Table.RootRowCount() == 0
	case gurps.BlockRangedKey:
		return sheet.RangedWeapons.Table.RootRowCount() == 0
	default:
		return false
	}
}

// TestBuildLayoutNodeRecordsTheContainerItWasBuiltFrom checks that a container panel records the Row or Column node its
// children were built from, and keeps doing so when an outer container left with only it is dropped onto it: the node
// governing the slot is then the outer container's, but the pairs of children the panel holds are still the inner
// one's.
func TestBuildLayoutNodeRecordsTheContainerItWasBuiltFrom(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockTraitsKey, gurps.BlockSkillsKey)
	row := testContainerNode(layoutnode.Row, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	)
	built := buildLayoutNode(row, leaf)
	c.NotNil(built)
	c.Equal(row, sheetLayoutContainerNodeOf(built.AsPanel()), "a container records the node it was built from")
	c.Equal(row, layoutNodeOf(built), "which also governs its slot when nothing was dropped onto it")
	c.Nil(sheetLayoutContainerNodeOf(panels[gurps.BlockTraitsKey]), "a block records no container")

	column := testContainerNode(layoutnode.Column, fxp.One, row, testBlockNode(gurps.BlockSpellsKey, fxp.One))
	built = buildLayoutNode(column, leaf)
	c.NotNil(built)
	c.Equal(column, layoutNodeOf(built), "the column left with only the row is dropped onto it and governs its slot")
	c.Equal(row, sheetLayoutContainerNodeOf(built.AsPanel()), "but the panel's children are still the row's")
	c.Equal(rowBandPanel, bandKindOf(built.AsPanel()), "and it is still laid out as a row")
}

// TestBuildLayoutSquareBlockInAStretchedRowStaysSquare checks, against the real builder, that a square block in a row a
// column stretches to the bottom of a taller sibling keeps its content square: the row is given more height than it
// asked for, so the block's width has to follow the height it is actually given rather than the one it was measured at.
func TestBuildLayoutSquareBlockInAStretchedRowStaysSquare(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockPortraitKey, gurps.BlockIdentityKey, gurps.BlockLiftingKey,
		gurps.BlockBodyKey)
	panels[gurps.BlockPortraitKey].SetBorder(unison.NewEmptyBorder(squareTestInsets))
	panels[gurps.BlockBodyKey] = newSizedPanel(10, 40, 200)
	portrait := testBlockNode(gurps.BlockPortraitKey, fxp.One)
	portrait.Square = true
	built := buildLayoutNode(testContainerNode(layoutnode.Row, fxp.One,
		testContainerNode(layoutnode.Column, fxp.One,
			testContainerNode(layoutnode.Row, fxp.One, portrait, testBlockNode(gurps.BlockIdentityKey, fxp.One)),
			testBlockNode(gurps.BlockLiftingKey, fxp.One),
		),
		testBlockNode(gurps.BlockBodyKey, fxp.One),
	), leaf)
	c.NotNil(built)
	band := built.AsPanel()
	_, prefSize, _ := band.Sizes(geom.NewSize(600, 0))
	c.Equal(float32(200), prefSize.Height, "the band is as tall as the body block")
	band.SetFrameRect(geom.NewRect(0, 0, 600, prefSize.Height))
	band.MarkForLayoutRecursively()
	band.ValidateLayout()

	portraitPanel := panels[gurps.BlockPortraitKey]
	inner := portraitPanel.Parent()
	c.True(inner.FrameRect().Height > 20,
		"the test requires the column to have stretched the row past the 20 its blocks asked for, but it is %v tall",
		inner.FrameRect().Height)
	c.Equal(inner.FrameRect().Height, portraitPanel.FrameRect().Height, "the portrait is given the full height of its row")
	content := portraitPanel.ContentRect(false).Size
	c.True(nearlySquare(content), "the portrait's content must be square at that height, but is %v", content)
}
