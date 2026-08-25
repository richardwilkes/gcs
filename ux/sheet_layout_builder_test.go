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
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// testBlockNode returns a Block node for the given key and weight.
func testBlockNode(key string, weight fxp.Int) *gurps.SheetLayoutNode {
	return &gurps.SheetLayoutNode{Type: layoutnode.Block, Key: key, Weight: weight}
}

// testContainerNode returns a Row or Column node holding the given children.
func testContainerNode(nodeType layoutnode.Type, weight fxp.Int,
	children ...*gurps.SheetLayoutNode,
) *gurps.SheetLayoutNode {
	return &gurps.SheetLayoutNode{Type: nodeType, Weight: weight, Children: children}
}

// testLeafFunc returns a leaf function that hands back a fresh panel for each of the given keys and nil for anything
// else, along with the map of the panels it made, so that a test can tell which panel is which.
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

// layoutNodeOf returns the layout node the builder recorded on the given panel.
func layoutNodeOf(p unison.Paneler) *gurps.SheetLayoutNode {
	node, ok := p.AsPanel().ClientData()[sheetLayoutNodeKey].(*gurps.SheetLayoutNode)
	if !ok {
		return nil
	}
	return node
}

// flexDataOf returns the layout data the builder gave the panel.
func flexDataOf(p unison.Paneler) *unison.FlexLayoutData {
	data, ok := p.AsPanel().LayoutData().(*unison.FlexLayoutData)
	if !ok {
		return nil
	}
	return data
}

// TestBuildLayoutBandsBuildsOneBandPerRootChild verifies that each of the root's bands becomes a panel of its own, laid
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

// TestBuildLayoutNodeColumn verifies that a column becomes a panel that stacks its children in a single column.
func TestBuildLayoutNodeColumn(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockTraitsKey, gurps.BlockSkillsKey)
	built := buildLayoutNode(testContainerNode(layoutnode.Column, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), leaf)
	c.NotNil(built)
	flex, ok := built.AsPanel().Layout().(*unison.FlexLayout)
	c.True(ok, "a column is laid out by a single-column flex layout")
	c.Equal(1, flex.Columns)
	c.Equal(float32(1), flex.VSpacing)
	c.Equal([]*unison.Panel{panels[gurps.BlockTraitsKey], panels[gurps.BlockSkillsKey]}, built.AsPanel().Children())
}

// TestBuildLayoutColumnChildrenFillTheRow verifies that the blocks stacked in a column that stands beside something
// taller than they are share out the height they are given, so that the bottom of the last of them is the bottom of
// the row rather than the page showing through below it.
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

// TestBuildLayoutNodePrunes verifies that a block the leaf function has nothing for is left out, that a container all
// of whose children were left out goes with them, and that a container left holding a single child is replaced by that
// child.
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

// TestBuildLayoutNodeLayoutData verifies that the builder is the sole authority on the layout data of the panels it
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

// TestBuildLayoutNodeSquareFlags verifies that a row is told which of its children take their width from its height,
// and that the square flag travels the same way the layout node does: a container left holding a single block is
// dropped, and the block that takes its place keeps the flag of its own node rather than the container's.
func TestBuildLayoutNodeSquareFlags(t *testing.T) {
	c := check.New(t)
	leaf, panels := testLeafFunc(gurps.BlockPortraitKey, gurps.BlockIdentityKey, gurps.BlockNotesKey)
	portraitNode := testBlockNode(gurps.BlockPortraitKey, fxp.One)
	portraitNode.Square = true
	// The column holds the portrait and a block with nothing to show, so it is dropped and the portrait takes its place.
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

// TestFactoryPortraitPictureAreaIsSquare verifies that the portrait of a sheet laid out the factory way has a square
// picture area whatever the page and the rest of the layout come to, since its width is taken from the height of the
// band it is in rather than from a share of the page width, and that the blocks beside it still fill the rest of that
// band.
func TestFactoryPortraitPictureAreaIsSquare(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	node, _, _ := entity.SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.NotNil(node, "the portrait must be in the layout")
	c.True(node.Square, "the factory portrait takes its width from the height of the band it is in")
	portrait := sheet.blocks[gurps.BlockPortraitKey].AsPanel()

	measure := func(what string) geom.Size {
		band := sheet.page.Children()[0]
		children := band.Children()
		c.Equal(0, band.IndexOfChild(portrait), "the portrait must be the first block of the first band (%s)", what)
		used := float32(len(children) - 1) // The single pixel between each pair of blocks
		for _, child := range children {
			used += child.FrameRect().Width
		}
		c.True(xmath.Abs(used-band.FrameRect().Width) < 0.01,
			"the blocks of the first band must fill it (%s), but %v of %v was used", what, used, band.FrameRect().Width)
		size := portrait.ContentRect(false).Size
		c.True(xmath.Abs(size.Width-size.Height) < 0.5, "the picture area must be square (%s), but was %v", what, size)
		return size
	}

	letter := measure("letter")

	_, _, valid := gurps.ParsePageSize("a4")
	c.True(valid, "a4 must be a page size the sheet can be laid out on")
	wasSize := entity.SheetSettings.Page.Size
	entity.SheetSettings.Page.Size = "a4"
	sheet.Rebuild(true)
	a4 := measure("a4")
	c.True(xmath.Abs(a4.Width-letter.Width) < 0.5,
		"a narrower page must leave the picture area alone, but %v became %v", letter.Width, a4.Width)

	entity.SheetSettings.Page.Size = wasSize
	c.True(entity.SheetSettings.Layout.SetMinHeight(gurps.BlockIdentityKey,
		paper.Length{Length: 2, Units: paper.Inch}))
	sheet.Rebuild(true)
	taller := measure("a two inch tall identity block")
	c.True(taller.Width > letter.Width+50,
		"a taller band must make the picture area wider, but %v became %v", letter.Width, taller.Width)
}

// TestSheetPageMatchesTheLayout verifies that the sheet's page holds one panel per band of the layout that has anything
// to show, and that the blocks that aren't lists are the panels the sheet keeps rather than new ones.
func TestSheetPageMatchesTheLayout(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	layout := sheet.Entity().SheetSettings.Layout

	c.Equal(nonEmptyBandCount(sheet, layout), len(sheet.page.Children()),
		"every band that has something to show, and only those, must be on the page")
	c.True(len(sheet.page.Children()) < len(layout.Root.Children),
		"a character with no content has bands with nothing to show")
	c.Equal(12, len(sheet.blocks), "every block that isn't a list is built and kept")
	for key, block := range sheet.blocks {
		c.NotNil(block.AsPanel().Parent(), "the %s block must be on the page", key)
	}
	c.Nil(sheet.Reactions.AsPanel().Parent(), "an empty reactions list must not be placed")
	c.NotNil(sheet.MeleeWeapons.AsPanel().Parent(), "a list with rows must be placed")
}

// nonEmptyBandCount returns the number of the layout's bands that have something to show for the given sheet, which for
// a character with no content means every band but the two whose lists are empty.
func nonEmptyBandCount(sheet *Sheet, layout *gurps.SheetLayout) int {
	count := 0
	for _, band := range layout.Root.Children {
		if bandHasContent(sheet, band) {
			count++
		}
	}
	return count
}

func bandHasContent(sheet *Sheet, node *gurps.SheetLayoutNode) bool {
	if node.Type == layoutnode.Block {
		return !isEmptyDerivedList(sheet, node.Key)
	}
	for _, child := range node.Children {
		if bandHasContent(sheet, child) {
			return true
		}
	}
	return false
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

// TestSheetHidingAndShowingABlock verifies that hiding a block takes it off the page without taking it away from the
// sheet, that nothing that reaches for it afterwards is handed something nobody can see, and that showing it again
// brings it back.
func TestSheetHidingAndShowingABlock(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	layout := sheet.Entity().SheetSettings.Layout
	notes := sheet.Notes
	c.NotNil(notes.AsPanel().Parent(), "the notes list starts out on the page")
	c.NotNil(sheet.keyToPanel(noteDragKey), "a note can be dropped onto the sheet")

	c.True(layout.Hide(gurps.BlockNotesKey))
	sheet.Rebuild(true)
	c.NotNil(sheet.Notes, "the sheet keeps its notes list even when the layout doesn't show it")
	c.Nil(sheet.Notes.AsPanel().Parent(), "a list that isn't on the page has no parent")
	c.Nil(sheet.keyToPanel(noteDragKey), "a list that isn't on the page is not a drop target")
	c.NotNil(sheet.keyToPanel(traitDragKey), "the lists that are still on the page remain drop targets")
	for _, child := range sheet.page.Children() {
		c.NotEqual(sheet.Notes.AsPanel(), child, "the notes list must not be on the page")
	}

	c.True(layout.Show(gurps.BlockNotesKey))
	sheet.Rebuild(true)
	c.NotNil(sheet.Notes.AsPanel().Parent(), "showing the block puts the list back on the page")
	c.NotNil(sheet.keyToPanel(noteDragKey), "and makes it a drop target again")
}

// TestSheetHidingABlockThatIsNotAList verifies that a block the sheet keeps a panel for is taken off the page when it
// is hidden, and that the panel itself survives to be put back.
func TestSheetHidingABlockThatIsNotAList(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	layout := sheet.Entity().SheetSettings.Layout
	portrait := sheet.blocks[gurps.BlockPortraitKey]
	c.NotNil(portrait)

	c.True(layout.Hide(gurps.BlockPortraitKey))
	sheet.Rebuild(true)
	c.Equal(portrait, sheet.blocks[gurps.BlockPortraitKey], "the panel is kept")
	c.Nil(portrait.AsPanel().Parent(), "but it isn't on the page")

	c.True(layout.Show(gurps.BlockPortraitKey))
	sheet.Rebuild(true)
	c.NotNil(portrait.AsPanel().Parent(), "showing it puts the very same panel back")
	c.Equal(portrait, sheet.blocks[gurps.BlockPortraitKey])
}

// TestTemplateBuildsFromTheDefaultLayout verifies that a template lays its content out from the default layout, showing
// exactly the five blocks a template can have.
func TestTemplateBuildsFromTheDefaultLayout(t *testing.T) {
	c := check.New(t)
	template := newTestTemplateWithBodyType("Humanoid")
	lists := []unison.Paneler{template.Traits, template.Skills, template.Spells, template.Equipment, template.Notes}
	for i, list := range lists {
		c.NotNil(list.AsPanel().Parent(), "list %d must be placed", i)
	}
	found := 0
	for _, child := range template.content.Children() {
		found += countPlacedLists(child, lists)
	}
	c.Equal(len(lists), found, "the template shows exactly its five lists")
}

// countPlacedLists returns the number of the given lists that are the panel or one of its descendants.
func countPlacedLists(panel *unison.Panel, lists []unison.Paneler) int {
	count := 0
	for _, list := range lists {
		if list.AsPanel() == panel {
			count++
		}
	}
	for _, child := range panel.Children() {
		count += countPlacedLists(child, lists)
	}
	return count
}

// TestBuildLayoutNodeRecordsTheContainerItWasBuiltFrom verifies that a container panel records the Row or Column node
// its children were built from, and keeps doing so when an outer container that was left with only it is dropped onto
// it: the node that governs the slot is then the outer container's, but the pairs of children the panel holds are
// still the inner one's.
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

// TestBuildLayoutSquareBlockInAStretchedRowStaysSquare verifies, against the real builder, that a square block in a row
// that a column stretches to the bottom of a taller sibling keeps its content square: the row is given more height
// than it asked for, so the block's width has to follow the height it is actually given rather than the one the row
// was measured at.
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
