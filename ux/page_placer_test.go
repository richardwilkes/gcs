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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

var _ pageHelper = &fakeBandList{}

// fakeBandList stands in for a page list when exercising the placement of a band, so that the rows a band is made to
// give up can be dictated exactly rather than being whatever a real table happens to measure.
type fakeBandList struct {
	unison.Panel
	overhead  float32
	heights   []float32
	start     int
	endBefore int
	ranges    [][2]int
}

// newFakeBandList returns a list of the given key whose rows are the given heights, and which costs the given overhead
// to show at all.
func newFakeBandList(key string, overhead float32, heights ...float32) *fakeBandList {
	f := &fakeBandList{
		overhead:  overhead,
		heights:   heights,
		endBefore: len(heights),
	}
	f.Self = f
	f.ClientData()[pageKey] = key
	f.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		width := hint.Width
		if width <= 0 {
			width = 100
		}
		size := geom.NewSize(width, f.currentHeight())
		return size, size, size
	})
	return f
}

// currentHeight returns the list's overhead plus the height of the rows it is currently set to draw.
func (f *fakeBandList) currentHeight() float32 {
	height := f.overhead
	for _, one := range f.heights[f.start:f.endBefore] {
		height += one
	}
	return height
}

// OverheadHeight implements pageHelper.
func (f *fakeBandList) OverheadHeight() float32 { return f.overhead }

// MinimumHeight implements pageHelper. The fake stands at no height beyond its overhead.
func (f *fakeBandList) MinimumHeight() float32 { return f.overhead }

// RowHeights implements pageHelper.
func (f *fakeBandList) RowHeights() []float32 { return f.heights }

// CurrentDrawRowRange implements pageHelper.
func (f *fakeBandList) CurrentDrawRowRange() (start, endBefore int) { return f.start, f.endBefore }

// SetDrawRowRange implements pageHelper, recording every range it is handed.
func (f *fakeBandList) SetDrawRowRange(start, endBefore int) {
	f.ranges = append(f.ranges, [2]int{start, endBefore})
	f.start = start
	f.endBefore = endBefore
}

// ClearDrawRowRange implements pageHelper, giving the list every row again.
func (f *fakeBandList) ClearDrawRowRange() {
	f.start = 0
	f.endBefore = len(f.heights)
}

// buildTestBand builds a band the way the exporter does, from the given node and a leaf function that hands back the
// given panels by key.
func buildTestBand(node *gurps.SheetLayoutNode, panels map[string]unison.Paneler) *unison.Panel {
	built := buildLayoutNode(node, func(key string) unison.Paneler {
		if panel, exists := panels[key]; exists {
			return panel
		}
		return nil
	})
	if xreflect.IsNil(built) {
		return nil
	}
	return built.AsPanel()
}

// TestPlaceBandListTakesAtLeastOneRow verifies that a list that can't fit even its first row is left for the next page,
// unless it already has the page to itself, in which case it takes one row anyway and lets it run off the end.
func TestPlaceBandListTakesAtLeastOneRow(t *testing.T) {
	c := check.New(t)
	panels := map[string]unison.Paneler{}
	node := testBlockNode(gurps.BlockTraitsKey, fxp.One)

	list := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20)
	panels[gurps.BlockTraitsKey] = list
	band := buildTestBand(node, panels)
	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 25, false)
	c.Equal(float32(0), used, "not even the first row fits")
	c.False(done)
	c.False(placer.applyBand(band), "a list that placed nothing comes off the page")
	c.Equal(0, len(list.ranges), "and is not told to draw anything")
	c.Equal(map[string]int{}, startAt, "nor does it move on from where it starts")

	list = newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20)
	panels[gurps.BlockTraitsKey] = list
	band = buildTestBand(node, panels)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 25, true)
	c.Equal(float32(30), used, "the row is placed anyway and allowed to run off the end of the page")
	c.False(done, "the second row is still to come")
	c.True(placer.applyBand(band))
	c.Equal([][2]int{{0, 1}}, list.ranges)
	c.Equal(map[string]int{gurps.BlockTraitsKey: 1}, startAt)
}

// TestPlaceBandAtomicBlock verifies that a block that isn't a list is placed whole when it fits, left for the next page
// when it doesn't, and placed anyway when it has the page to itself.
func TestPlaceBandAtomicBlock(t *testing.T) {
	c := check.New(t)
	node := testBlockNode(gurps.BlockBodyKey, fxp.One)
	newBand := func() *unison.Panel {
		panel := newSizedPanel(10, 100, 50)
		panel.ClientData()[pageKey] = gurps.BlockBodyKey
		return buildTestBand(node, map[string]unison.Paneler{gurps.BlockBodyKey: panel})
	}

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(newBand(), 60, false)
	c.Equal(float32(50), used, "a block that fits takes the height it asks for")
	c.True(done)

	band := newBand()
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 40, false)
	c.Equal(float32(0), used, "a block that doesn't fit is left for the next page")
	c.False(done)
	c.False(placer.applyBand(band), "and comes off this one")
	c.Equal(map[string]int{}, startAt)

	band = newBand()
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 40, true)
	c.Equal(float32(40), used, "a block with the page to itself is placed anyway and clipped")
	c.True(done)
	c.True(placer.applyBand(band))
	c.Equal(map[string]int{gurps.BlockBodyKey: placedBlockMarker}, startAt,
		"a block that isn't a list is only ever placed once")
}

// TestPlaceBandColumn verifies that a column of blocks is placed from the top down, that nothing below the first block
// that isn't finished is placed, and that applying the decision takes those blocks off the page, while a column whose
// blocks all fit is finished in one pass.
func TestPlaceBandColumn(t *testing.T) {
	c := check.New(t)
	build := func(firstHeights ...float32) (band *unison.Panel, first, second *fakeBandList) {
		first = newFakeBandList(gurps.BlockTraitsKey, 10, firstHeights...)
		second = newFakeBandList(gurps.BlockSkillsKey, 10, 20)
		band = buildTestBand(testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One),
		), map[string]unison.Paneler{gurps.BlockTraitsKey: first, gurps.BlockSkillsKey: second})
		return band, first, second
	}

	band, first, second := build(20, 20, 20, 20)
	c.Equal(columnBandPanel, bandKindOf(band), "the builder's column is recognized as one")

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(50), used, "what the first block took, since the column stops there")
	c.False(done)

	c.True(placer.applyBand(band), "the column keeps what was placed")
	c.Equal([]*unison.Panel{first.AsPanel()}, band.Children(), "the block that got nothing is off the page")
	c.Equal([][2]int{{0, 2}}, first.ranges)
	c.Equal(0, len(second.ranges), "the block that got nothing is not told to draw anything")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 2}, startAt)

	band, _, _ = build(20)
	startAt = make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 200, false)
	c.Equal(float32(61), used, "a column whose blocks all fit takes both blocks and the space between them")
	c.True(done, "and is finished in one pass")
	c.True(placer.applyBand(band))
	c.Equal(2, len(band.Children()), "nothing comes off the page")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 1, gurps.BlockSkillsKey: 1}, startAt)
}

// TestPlaceBandRow verifies that the blocks of a row are each given the same space, since they all start at the same
// height, and that the tallest of them is what the row costs, while a row with a block that can place nothing at all
// is left for the next page in its entirety, taking back what its other blocks, those within a column of it included,
// could have placed rather than leaving a hole in the middle of it.
func TestPlaceBandRow(t *testing.T) {
	c := check.New(t)
	build := func(leftHeights, rightHeights []float32) (band *unison.Panel, left, right *fakeBandList) {
		left = newFakeBandList(gurps.BlockTraitsKey, 10, leftHeights...)
		right = newFakeBandList(gurps.BlockSkillsKey, 10, rightHeights...)
		band = buildTestBand(testContainerNode(layoutnode.Row, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One),
		), map[string]unison.Paneler{gurps.BlockTraitsKey: left, gurps.BlockSkillsKey: right})
		return band, left, right
	}

	band, left, right := build([]float32{20, 20, 20, 20}, []float32{20, 20})
	c.Equal(rowBandPanel, bandKindOf(band), "the builder's row is recognized as one")
	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(50), used, "the tallest of the two")
	c.False(done, "the block on the left still has rows to place")
	c.True(placer.applyBand(band))
	c.Equal([][2]int{{0, 2}}, left.ranges)
	c.Equal([][2]int{{0, 2}}, right.ranges, "the block on the right was given the same space and fit in it")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 2, gurps.BlockSkillsKey: 2}, startAt)

	band, left, right = build([]float32{20, 20}, []float32{100})
	startAt = make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 57, false)
	c.Equal(float32(0), used, "the row that can't be split up places nothing at all")
	c.False(done)
	c.False(placer.applyBand(band), "so the whole row comes off the page")
	c.Equal(0, len(left.ranges), "even the block that could have placed rows is left alone")
	c.Equal(0, len(right.ranges))
	c.Equal(map[string]int{}, startAt)

	nested := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20)
	band = buildTestBand(testContainerNode(layoutnode.Row, fxp.One,
		testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One),
		),
		testBlockNode(gurps.BlockNotesKey, fxp.One),
	), map[string]unison.Paneler{
		gurps.BlockTraitsKey: nested,
		gurps.BlockSkillsKey: newFakeBandList(gurps.BlockSkillsKey, 10, 20),
		gurps.BlockNotesKey:  newFakeBandList(gurps.BlockNotesKey, 10, 100),
	})
	startAt = make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 57, false)
	c.Equal(float32(0), used, "a row holding a column places nothing at all when a block beside the column can't")
	c.False(done)
	c.False(placer.applyBand(band), "so the whole row comes off the page")
	c.Equal(0, len(nested.ranges), "even a block within the column that could have placed rows is left alone")
	c.Equal(map[string]int{}, startAt)
}

// TestPlaceBandHonorsAMinimumHeight verifies that a block's minimum height is checked against what the page has left: a
// list whose rows fit but whose floor doesn't is left for the next page rather than running off the bottom, one whose
// floor fits is placed and stands at it, and one with the page to itself is placed at its floor regardless. A page
// spacer keeping the slot of a block with a floor has none of its own: neither the placer nor the row's layout gives
// the empty slot any height. Building the spacer with its block's floor fails it.
func TestPlaceBandHonorsAMinimumHeight(t *testing.T) {
	c := check.New(t)
	inch := paper.Length{Length: 1, Units: paper.Inch}
	build := func() (band *unison.Panel, long, short *fakeBandList) {
		long = newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20, 20, 20, 20)
		short = newFakeBandList(gurps.BlockSkillsKey, 10, 20)
		skills := testBlockNode(gurps.BlockSkillsKey, fxp.One)
		skills.MinHeight = inch
		band = buildTestBand(testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			skills,
		), map[string]unison.Paneler{gurps.BlockTraitsKey: long, gurps.BlockSkillsKey: short})
		return band, long, short
	}

	// The long list takes 110 and the space below it 1, leaving 39: enough for the 30 the short list's rows take, but
	// not for the 72 it must stand at.
	band, long, short := build()
	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 150, false)
	c.Equal(float32(111), used, "the block whose floor doesn't fit must be left for the next page")
	c.False(done)
	c.True(placer.applyBand(band))
	c.Equal([]*unison.Panel{long.AsPanel()}, band.Children(), "so it comes off this page")
	c.Equal(0, len(short.ranges), "and is not told to draw anything")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 5}, startAt)

	band, _, short = build()
	startAt = make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 200, false)
	c.Equal(float32(183), used, "with room for it, the block stands at its floor rather than at the height of its rows")
	c.True(done)
	c.True(placer.applyBand(band))
	c.Equal(2, len(band.Children()), "and nothing comes off the page")
	c.Equal([][2]int{{0, 1}}, short.ranges)
	c.Equal(map[string]int{gurps.BlockTraitsKey: 5, gurps.BlockSkillsKey: 1}, startAt)

	alone := func() *unison.Panel {
		skills := testBlockNode(gurps.BlockSkillsKey, fxp.One)
		skills.MinHeight = inch
		return buildTestBand(skills, map[string]unison.Paneler{
			gurps.BlockSkillsKey: newFakeBandList(gurps.BlockSkillsKey, 10, 20),
		})
	}
	band = alone()
	placer = newBandPlacer(make(map[string]int))
	used, done = placer.placeBand(band, 50, false)
	c.Equal(float32(0), used, "a list whose floor is deeper than the page has left is deferred, whole")
	c.False(done)
	c.False(placer.applyBand(band))

	band = alone()
	placer = newBandPlacer(make(map[string]int))
	used, done = placer.placeBand(band, 50, true)
	c.Equal(float32(72), used, "unless it has the page to itself, when it stands at its floor regardless")
	c.True(done)
	c.True(placer.applyBand(band))

	traits := testBlockNode(gurps.BlockTraitsKey, fxp.One)
	traits.MinHeight = inch
	short = newFakeBandList(gurps.BlockSkillsKey, 10, 20)
	band = buildTestBand(testContainerNode(layoutnode.Row, fxp.One,
		traits,
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), map[string]unison.Paneler{gurps.BlockTraitsKey: newPageSpacer(), gurps.BlockSkillsKey: short})
	placer = newBandPlacer(make(map[string]int))
	used, done = placer.placeBand(band, 200, false)
	c.Equal(float32(30), used, "a spacer reserves none of the height its block's floor called for")
	c.True(done)
	_, pref, _ := band.Sizes(geom.Size{Width: 400})
	c.Equal(used, pref.Height, "nor does the row stand at that floor")
	c.True(placer.applyBand(band))
	c.Equal(2, len(band.Children()), "the spacer keeps its slot")

	// A real list stands at its view's minimum height however few rows it draws, which for one with no rows, or only a
	// short one, is more than its border, header and rows take.
	loot := newTestLootSheet(t)
	for _, count := range []int{0, 1} {
		loot.loot.Notes = nil
		for range count {
			loot.loot.Notes = append(loot.loot.Notes, gurps.NewNote(nil, nil, false))
		}
		loot.Notes.Sync()
		listBand := func() *unison.Panel {
			return buildTestBand(testBlockNode(gurps.BlockNotesKey, fxp.One), map[string]unison.Paneler{
				gurps.BlockNotesKey: pageListLeaf(gurps.BlockNotesKey, loot.Notes, 0, make(map[string]int), false),
			})
		}
		band = listBand()
		_, pref, _ = band.Sizes(geom.Size{})
		placer = newBandPlacer(make(map[string]int))
		used, done = placer.placeBand(band, pref.Height, false)
		c.Equal(pref.Height, used, "a list with %d row(s) takes the height it stands at", count)
		c.True(done)

		band = listBand()
		placer = newBandPlacer(make(map[string]int))
		used, done = placer.placeBand(band, pref.Height-1, false)
		c.Equal(float32(0), used, "a list with %d row(s) is left for the next page when that height doesn't fit", count)
		c.False(done)
		c.False(placer.applyBand(band))

		band = listBand()
		placer = newBandPlacer(make(map[string]int))
		used, done = placer.placeBand(band, pref.Height-1, true)
		c.Equal(pref.Height, used, "unless a list with %d row(s) has the page to itself", count)
		c.True(done)
	}
}

// TestPlaceBandCountsAListsRowsAtTheHeightTheyMeasure verifies that the placer counts the rows of a real list at the
// height the list stands at when it draws them: the whole list fits in exactly that height, and a pixel less leaves
// only its last row for the next page, with the rows placed standing at the height the placer counted for them.
// Counting a row divider the table doesn't draw fails it.
func TestPlaceBandCountsAListsRowsAtTheHeightTheyMeasure(t *testing.T) {
	c := check.New(t)
	const count = 30
	loot := newTestLootSheet(t)
	for i := range count {
		note := gurps.NewNote(nil, nil, false)
		note.MarkDown = fmt.Sprintf("Note %d", i+1)
		loot.loot.Notes = append(loot.loot.Notes, note)
	}
	loot.Notes.Sync()
	listBand := func() *unison.Panel {
		return buildTestBand(testBlockNode(gurps.BlockNotesKey, fxp.One), map[string]unison.Paneler{
			gurps.BlockNotesKey: pageListLeaf(gurps.BlockNotesKey, loot.Notes, 0, make(map[string]int), false),
		})
	}

	band := listBand()
	_, pref, _ := band.Sizes(geom.Size{})
	placer := newBandPlacer(make(map[string]int))
	used, done := placer.placeBand(band, pref.Height, false)
	c.Equal(pref.Height, used, "the list takes the height it stands at")
	c.True(done, "with every row placed")

	band = listBand()
	startAt := make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, pref.Height-1, false)
	c.False(done)
	c.True(placer.applyBand(band))
	c.Equal(map[string]int{gurps.BlockNotesKey: count - 1}, startAt, "a pixel less leaves only the last row")
	_, pref, _ = band.Sizes(geom.Size{})
	c.Equal(used, pref.Height, "and the rows placed stand at the height the placer counted for them")
}

// TestPaginateBandsFinishesABandWithNothingLeft verifies that a band that builds to nothing but a page spacer, every
// block of it having been placed already, is finished, whether it comes to that before any of it is placed or after a
// pass that wasn't done: it is neither left on a page nor given a new page of its own. Treating only a nil band as
// finished fails it, putting the spacer on the page or opening a page for nothing.
func TestPaginateBandsFinishesABandWithNothingLeft(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	pageSize := (&pageExporter{entity: entity}).PageSize()
	tall := pageSize.Height * 0.6
	for _, one := range []struct {
		name     string
		root     *gurps.SheetLayoutNode
		rows     map[string][]float32 // The rows of each list the leaf hands back
		placedAt map[string]int       // The row from which the leaf reports a list as placed on an earlier page
		keys     []string             // The keys of the blocks the only page must hold, in order
	}{
		{
			name: "before any of it is placed",
			root: testContainerNode(layoutnode.Column, fxp.One,
				testContainerNode(layoutnode.Row, fxp.One,
					testBlockNode(gurps.BlockTraitsKey, fxp.One),
					testBlockNode(gurps.BlockSkillsKey, fxp.One),
				),
				testBlockNode(gurps.BlockNotesKey, fxp.One),
			),
			rows: map[string][]float32{
				gurps.BlockTraitsKey: {20},
				gurps.BlockSkillsKey: {20},
				gurps.BlockNotesKey:  {20, 20},
			},
			placedAt: map[string]int{gurps.BlockNotesKey: 2},
			keys:     []string{gurps.BlockNotesKey},
		},
		{
			name: "after a pass that wasn't done",
			root: testContainerNode(layoutnode.Column, fxp.One,
				testBlockNode(gurps.BlockTraitsKey, fxp.One),
			),
			rows:     map[string][]float32{gurps.BlockTraitsKey: {tall, tall}},
			placedAt: map[string]int{gurps.BlockTraitsKey: 1},
			keys:     []string{gurps.BlockTraitsKey},
		},
	} {
		leaf := func(startAt map[string]int) layoutLeafFunc {
			return func(key string) unison.Paneler {
				rows, exists := one.rows[key]
				if !exists {
					return nil
				}
				if startAt[key] >= one.placedAt[key] {
					return newPageSpacer()
				}
				list := newFakeBandList(key, 10, rows...)
				list.start = startAt[key]
				return list
			}
		}
		created := 0
		pages := paginateBands(one.root, leaf, func() *Page {
			created++
			return NewPage(entity)
		}, pageSize)
		c.Equal(1, created, "%s: no page may be opened for the band with nothing left", one.name)
		c.Equal(1, len(pages), "%s: everything fits on a single page", one.name)
		for i, page := range pages {
			keys := make([]string, 0, len(page.Children()))
			for _, band := range page.Children() {
				keys = append(keys, pageKeyOf(band))
			}
			c.Equal(one.keys, keys, "%s: page %d must hold just the blocks that had something to place", one.name,
				i+1)
			c.Equal(0, len(uxtest.PanelsMatching(page.AsPanel(), isPageSpacer)),
				"%s: page %d must hold no spacer", one.name, i+1)
		}
	}
}

// TestPaginatedSheetKeepsAnEmptyListBesideTheStartOfALongOne verifies that an empty list sharing a band with a list
// going on over several pages is shown once, on the page the band starts on, rather than handed out afresh for each
// page the band goes on to. Taking a list with no rows for one not yet placed fails it.
func TestPaginatedSheetKeepsAnEmptyListBesideTheStartOfALongOne(t *testing.T) {
	c := check.New(t)
	sheet := newPaginatedTestSheetWithTraits(t, 150)
	pages := sheet.grid.pages
	c.True(len(drawnRowRanges(pages, gurps.BlockTraitsKey)) > 1, "the traits must go on over several pages")
	c.Equal(0, sheet.Skills.RowCount(), "the skills list must be empty")
	c.True(unison.AncestorIs(sheet.Skills.Table, pages[0]), "the skills list must be on the first page")
	count := 0
	for _, page := range pages {
		count += len(blockPanels(page.AsPanel(), gurps.BlockSkillsKey))
	}
	c.Equal(1, count, "the skills list must be shown once")
}
