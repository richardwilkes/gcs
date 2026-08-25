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
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// newTestTemplateDockable returns a template dockable for the given file path. Building the toolbar reaches for the
// bindable actions, so they are registered first.
func newTestTemplateDockable(fileName string, data *gurps.Template) *Template {
	registerKeyBindingsOnce.Do(func() { registerActions() })
	return NewTemplate(filepath.Join("some", "dir", fileName+gurps.TemplatesExt), data)
}

// TestPageInfoProviderForTemplate verifies that obtaining a template dockable's page info provider fills in the page
// title and clears the modification timestamp. A gurps.Template returns only these explicitly-set values, so a page
// building path that skipped this setup — printing used to — produced footers with a blank title, or with a stale
// title left behind by an earlier export in the same session.
func TestPageInfoProviderForTemplate(t *testing.T) {
	c := check.New(t)

	t.Run("the title comes from the dockable", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		c.Equal("", data.PageTitle(), "a template has no page title until one is supplied")

		provider := pageInfoProviderFor(dockable)
		c.Equal(data, provider, "the template itself is the page info provider")
		c.Equal("My Template", provider.PageTitle(), "the page title is the dockable's title")
		c.Equal("", provider.ModifiedOnString(), "a template has no modification timestamp to show")
	})

	t.Run("stale values from an earlier export are replaced", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		data.ExplicitPageTitle = "Some Other Template"
		data.ExplicitModifiedOn = "Jan 1, 1970"

		provider := pageInfoProviderFor(dockable)
		c.Equal("My Template", provider.PageTitle(), "the stale page title is replaced")
		c.Equal("", provider.ModifiedOnString(), "the stale modification timestamp is cleared")
	})

	t.Run("renaming the dockable's file is picked up", func(_ *testing.T) {
		data := gurps.NewTemplate()
		dockable := newTestTemplateDockable("My Template", data)
		c.Equal("My Template", pageInfoProviderFor(dockable).PageTitle())

		dockable.path = filepath.Join("elsewhere", "Renamed"+gurps.TemplatesExt)
		c.Equal("Renamed", pageInfoProviderFor(dockable).PageTitle(), "the current title is used, not the first one")
	})
}

// TestPageInfoProviderForSheet verifies that a character sheet's page info provider is handed back as-is, with its own
// title intact, since only templates need values supplied by their dockable.
func TestPageInfoProviderForSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Profile.Name = "Dai Blackthorn"

	provider := pageInfoProviderFor(sheet)
	c.Equal(entity, provider, "the entity itself is the page info provider")
	c.Equal("Dai Blackthorn", provider.PageTitle(), "the entity supplies its own page title")
}

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

// currentHeight returns the height the rows the list is currently set to draw take up.
func (f *fakeBandList) currentHeight() float32 {
	height := f.overhead
	for _, one := range f.heights[f.start:f.endBefore] {
		height += one + 1
	}
	return height
}

// OverheadHeight implements pageHelper.
func (f *fakeBandList) OverheadHeight() float32 { return f.overhead }

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

// TestPlaceBandSplitsAList verifies that a list gives up as many of its rows as fit and leaves the rest for the next
// page, and that applying that decision restricts what it draws and records where to pick it up from.
func TestPlaceBandSplitsAList(t *testing.T) {
	c := check.New(t)
	list := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20, 20, 20)
	band := buildTestBand(testBlockNode(gurps.BlockTraitsKey, fxp.One),
		map[string]unison.Paneler{gurps.BlockTraitsKey: list})
	c.Equal(list.AsPanel(), band, "a band that is a single block is that block's panel")

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(52), used, "the overhead plus the two rows that fit")
	c.False(done, "there are rows left over")

	c.True(placer.applyBand(band), "a list that placed rows stays on the page")
	c.Equal([][2]int{{0, 2}}, list.ranges, "only the rows that fit are drawn")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 2}, startAt, "the next page picks up at the first row left over")
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
	c.Equal(float32(31), used, "the row is placed anyway and allowed to run off the end of the page")
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
// that isn't finished is placed, and that applying the decision takes those blocks off the page.
func TestPlaceBandColumn(t *testing.T) {
	c := check.New(t)
	first := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20, 20, 20)
	second := newFakeBandList(gurps.BlockSkillsKey, 10, 20)
	band := buildTestBand(testContainerNode(layoutnode.Column, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), map[string]unison.Paneler{gurps.BlockTraitsKey: first, gurps.BlockSkillsKey: second})
	c.Equal(columnBandPanel, bandKindOf(band), "the builder's column is recognized as one")

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(52), used, "what the first block took, since the column stops there")
	c.False(done)

	c.True(placer.applyBand(band), "the column keeps what was placed")
	c.Equal([]*unison.Panel{first.AsPanel()}, band.Children(), "the block that got nothing is off the page")
	c.Equal([][2]int{{0, 2}}, first.ranges)
	c.Equal(0, len(second.ranges), "the block that got nothing is not told to draw anything")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 2}, startAt)
}

// TestPlaceBandColumnPlacesEverythingThatFits verifies that a column whose blocks all fit is finished in one pass.
func TestPlaceBandColumnPlacesEverythingThatFits(t *testing.T) {
	c := check.New(t)
	first := newFakeBandList(gurps.BlockTraitsKey, 10, 20)
	second := newFakeBandList(gurps.BlockSkillsKey, 10, 20)
	band := buildTestBand(testContainerNode(layoutnode.Column, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), map[string]unison.Paneler{gurps.BlockTraitsKey: first, gurps.BlockSkillsKey: second})

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 200, false)
	c.Equal(float32(63), used, "both blocks and the space between them")
	c.True(done)
	c.True(placer.applyBand(band))
	c.Equal(2, len(band.Children()), "nothing comes off the page")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 1, gurps.BlockSkillsKey: 1}, startAt)
}

// TestPlaceBandRow verifies that the blocks of a row are each given the same space, since they all start at the same
// height, and that the tallest of them is what the row costs.
func TestPlaceBandRow(t *testing.T) {
	c := check.New(t)
	left := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20, 20, 20)
	right := newFakeBandList(gurps.BlockSkillsKey, 10, 20, 20)
	band := buildTestBand(testContainerNode(layoutnode.Row, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), map[string]unison.Paneler{gurps.BlockTraitsKey: left, gurps.BlockSkillsKey: right})
	c.Equal(rowBandPanel, bandKindOf(band), "the builder's row is recognized as one")

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(52), used, "the tallest of the two")
	c.False(done, "the block on the left still has rows to place")
	c.True(placer.applyBand(band))
	c.Equal([][2]int{{0, 2}}, left.ranges)
	c.Equal([][2]int{{0, 2}}, right.ranges, "the block on the right was given the same space and fit in it")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 2, gurps.BlockSkillsKey: 2}, startAt)
}

// TestPlaceBandRowDefersAsAWhole verifies that a row none of whose blocks can place anything is left for the next page
// in its entirety, taking back what its other blocks could have placed rather than leaving a hole in the middle of it.
func TestPlaceBandRowDefersAsAWhole(t *testing.T) {
	c := check.New(t)
	left := newFakeBandList(gurps.BlockTraitsKey, 10, 20, 20)
	right := newFakeBandList(gurps.BlockSkillsKey, 10, 100)
	band := buildTestBand(testContainerNode(layoutnode.Row, fxp.One,
		testBlockNode(gurps.BlockTraitsKey, fxp.One),
		testBlockNode(gurps.BlockSkillsKey, fxp.One),
	), map[string]unison.Paneler{gurps.BlockTraitsKey: left, gurps.BlockSkillsKey: right})

	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 57, false)
	c.Equal(float32(0), used, "the row that can't be split up places nothing at all")
	c.False(done)
	c.False(placer.applyBand(band), "so the whole row comes off the page")
	c.Equal(0, len(left.ranges), "even the block that could have placed rows is left alone")
	c.Equal(0, len(right.ranges))
	c.Equal(map[string]int{}, startAt)
}

// TestPlaceBandHonorsAMinimumHeight verifies that a block's minimum height is weighed against what the page has left
// before the block is placed, so that a list whose rows would fit but whose floor would not is left for the next page
// rather than being placed and then stood at a floor that runs off the bottom of the page; that a list whose floor fits
// is placed and stands at it; and that a block which has the page to itself is placed at its floor regardless.
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

	// The long list takes 115 and the space below it 1, leaving 34: enough for the 31 the short list's rows take, but
	// not for the 72 it must stand at.
	band, long, short := build()
	startAt := make(map[string]int)
	placer := newBandPlacer(startAt)
	used, done := placer.placeBand(band, 150, false)
	c.Equal(float32(116), used, "the block whose floor doesn't fit must be left for the next page")
	c.False(done)
	c.True(placer.applyBand(band))
	c.Equal([]*unison.Panel{long.AsPanel()}, band.Children(), "so it comes off this page")
	c.Equal(0, len(short.ranges), "and is not told to draw anything")
	c.Equal(map[string]int{gurps.BlockTraitsKey: 5}, startAt)

	band, _, short = build()
	startAt = make(map[string]int)
	placer = newBandPlacer(startAt)
	used, done = placer.placeBand(band, 200, false)
	c.Equal(float32(188), used, "with room for it, the block stands at its floor rather than at the height of its rows")
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
}

// TestPageExporterKeepsEveryBandWithinThePage verifies, over a range of list lengths and for a list stacked below a
// long one both as a band of its own and within a column, that no band of an exported page runs past the bottom of
// the page's printable area. A list given a minimum height is what used to: the rows it had left fit in what remained
// of the page, so it was placed there, and only then was it stood at a floor deeper than that.
func TestPageExporterKeepsEveryBandWithinThePage(t *testing.T) {
	c := check.New(t)
	shapes := map[string]func() *gurps.SheetLayoutNode{
		"as bands": func() *gurps.SheetLayoutNode {
			return testContainerNode(layoutnode.Column, fxp.One,
				testBlockNode(gurps.BlockTraitsKey, fxp.One),
				testBlockNode(gurps.BlockSkillsKey, fxp.One),
			)
		},
		"in a column": func() *gurps.SheetLayoutNode {
			return testContainerNode(layoutnode.Column, fxp.One,
				testContainerNode(layoutnode.Row, fxp.One,
					testContainerNode(layoutnode.Column, fxp.One,
						testBlockNode(gurps.BlockTraitsKey, fxp.One),
						testBlockNode(gurps.BlockSkillsKey, fxp.One),
					),
					testBlockNode(gurps.BlockNotesKey, fxp.One),
				),
			)
		},
	}
	for name, shape := range shapes {
		for _, traitCount := range []int{0, 8, 16, 24, 32, 40, 48, 56} {
			entity := gurps.NewEntity()
			for i := range traitCount {
				trait := gurps.NewTrait(nil, nil, false)
				trait.Name = fmt.Sprintf("Trait %d", i+1)
				entity.Traits = append(entity.Traits, trait)
			}
			for i := range 2 {
				skill := gurps.NewSkill(nil, nil, false)
				skill.Name = fmt.Sprintf("Skill %d", i+1)
				entity.Skills = append(entity.Skills, skill)
				note := gurps.NewNote(nil, nil, false)
				note.MarkDown = fmt.Sprintf("Note %d", i+1)
				entity.Notes = append(entity.Notes, note)
			}
			layout := &gurps.SheetLayout{Root: shape()}
			keys := layout.VisibleKeys()
			for _, key := range gurps.AllBlockKeys {
				if !slices.Contains(keys, key) {
					layout.Hidden = append(layout.Hidden, key)
				}
			}
			c.True(layout.SetMinHeight(gurps.BlockSkillsKey, paper.Length{Length: 3, Units: paper.Inch}))
			layout.EnsureValidity()
			entity.SheetSettings.Layout = layout

			exporter := newPageExporter(entity)
			pageSize := exporter.PageSize()
			for i, page := range exporter.pages {
				bottom := pageSize.Height - page.insets().Bottom
				for _, band := range page.Children() {
					c.True(band.FrameRect().Bottom() <= bottom+0.5,
						"with %d traits %s, a band of page %d ends at %v, past the printable bottom of %v",
						traitCount, name, i+1, band.FrameRect().Bottom(), bottom)
				}
			}
		}
	}
}

// TestPageExporterFitsAnEmptyCharacterOnOnePage verifies that a character with no content still exports the way it
// always has: a single page whose every band came from the block layout.
func TestPageExporterFitsAnEmptyCharacterOnOnePage(t *testing.T) {
	c := check.New(t)
	exporter := newPageExporter(gurps.NewEntity())
	c.Equal(1, len(exporter.pages), "a character with no content takes a single page")
	children := exporter.pages[0].Children()
	c.True(len(children) > 1, "the page holds the bands of the layout")
	for i, child := range children {
		c.NotNil(layoutNodeOf(child), "band %d must record the layout node it was built from", i)
	}
}

// TestPageExporterSplitsAListAcrossPages verifies that a list too long for a single page is split across as many as it
// takes, with every row drawn exactly once.
func TestPageExporterSplitsAListAcrossPages(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	const traitCount = 200
	for i := range traitCount {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}

	rowCount := len(entity.Traits) // A new character may come with a trait of its own already
	exporter := newPageExporter(entity)
	c.True(len(exporter.pages) > 1, "%d traits don't fit on a single page", traitCount)
	ranges := drawnRowRanges(exporter, gurps.BlockTraitsKey)
	c.True(len(ranges) > 1, "the traits list must appear on more than one page")
	next := 0
	for i, one := range ranges {
		c.Equal(next, one[0], "the rows drawn on page %d must pick up where the page before it left off", i+1)
		c.True(one[1] > one[0], "page %d must draw at least one row", i+1)
		next = one[1]
	}
	c.Equal(rowCount, next, "every row must be drawn")
}

// TestPageExporterMovesABlockThatDoesNotFit verifies the user's requirement that a block which isn't a list is never
// split: one that can't fit in what is left of a page is moved to the next one whole, and only appears there.
func TestPageExporterMovesABlockThatDoesNotFit(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 5 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	// A layout of just the traits list and a body block too tall for any page to hold along with anything else.
	layout := &gurps.SheetLayout{
		Root: testContainerNode(layoutnode.Column, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockBodyKey, fxp.One),
		),
	}
	for _, key := range gurps.AllBlockKeys {
		if key != gurps.BlockTraitsKey && key != gurps.BlockBodyKey {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	c.True(layout.SetMinHeight(gurps.BlockBodyKey, paper.Length{Length: 20, Units: paper.Inch}))
	layout.EnsureValidity()
	entity.SheetSettings.Layout = layout

	exporter := newPageExporter(entity)
	c.Equal(2, len(exporter.pages), "the body block can't share the first page with the traits")
	c.Equal(1, countBlockPanels(exporter, gurps.BlockBodyKey), "and must appear exactly once")
	c.Equal(gurps.BlockTraitsKey, pageKeyOf(exporter.pages[0].Children()[0]))
	c.Equal(1, len(exporter.pages[1].Children()))
	c.Equal(gurps.BlockBodyKey, pageKeyOf(exporter.pages[1].Children()[0]),
		"the body block must be the first thing on the second page")
}

// TestPageExporterForTemplateAndLoot verifies that the kinds of sheet that have no character behind them still export.
func TestPageExporterForTemplateAndLoot(t *testing.T) {
	c := check.New(t)
	data := gurps.NewTemplate()
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Template Trait"
	data.Traits = []*gurps.Trait{trait}
	exporter := newPageExporter(pageInfoProviderFor(newTestTemplateDockable("My Template", data)))
	c.Equal(1, len(exporter.pages))
	c.Equal(1, countBlockPanels(exporter, gurps.BlockTraitsKey), "the template's traits are on the page")
	c.Equal(0, countBlockPanels(exporter, gurps.BlockBodyKey), "a block that only a character has is not")

	lootExporter := newPageExporter(gurps.NewLoot())
	c.Equal(1, len(lootExporter.pages))
	c.True(len(lootExporter.pages[0].Children()) > 0, "the loot sheet's top block is on the page")
	c.Equal(0, countBlockPanels(lootExporter, gurps.BlockBodyKey))
}

// drawnRowRanges returns the row range each page draws of the block with the given key, in page order.
func drawnRowRanges(p *pageExporter, key string) [][2]int {
	var ranges [][2]int
	for _, page := range p.pages {
		for _, panel := range blockPanels(page.AsPanel(), key) {
			if helper, ok := panel.Self.(pageHelper); ok {
				start, endBefore := helper.CurrentDrawRowRange()
				ranges = append(ranges, [2]int{start, endBefore})
			}
		}
	}
	return ranges
}

// countBlockPanels returns the number of panels showing the block with the given key across all of the pages.
func countBlockPanels(p *pageExporter, key string) int {
	count := 0
	for _, page := range p.pages {
		count += len(blockPanels(page.AsPanel(), key))
	}
	return count
}

// blockPanels returns the panels showing the block with the given key within the given panel.
func blockPanels(panel *unison.Panel, key string) []*unison.Panel {
	var found []*unison.Panel
	if pageKeyOf(panel) == key {
		found = append(found, panel)
	}
	for _, child := range panel.Children() {
		found = append(found, blockPanels(child, key)...)
	}
	return found
}

// TestPageExporterSplitsBothListsOfARow verifies that the two lists standing side by side in a band are each split at
// their own row, the way the character sheet has always paginated a row of two blocks.
func TestPageExporterSplitsBothListsOfARow(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 150 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
		skill := gurps.NewSkill(entity, nil, false)
		skill.Name = fmt.Sprintf("Skill %d", i+1)
		entity.Skills = append(entity.Skills, skill)
	}
	entity.Recalculate()
	// A new character may come with a trait of its own already, so the counts are taken from the character itself.
	rowCounts := map[string]int{
		gurps.BlockTraitsKey: len(entity.Traits),
		gurps.BlockSkillsKey: len(entity.Skills),
	}

	exporter := newPageExporter(entity)
	for _, key := range []string{gurps.BlockTraitsKey, gurps.BlockSkillsKey} {
		ranges := drawnRowRanges(exporter, key)
		c.True(len(ranges) > 1, "the %s list must be split across pages", key)
		next := 0
		for i, one := range ranges {
			c.Equal(next, one[0], "the %s rows drawn on page %d must pick up where the page before left off", key, i+1)
			c.True(one[1] > one[0], "page %d must draw at least one %s row", i+1, key)
			next = one[1]
		}
		c.Equal(rowCounts[key], next, "every %s row must be drawn", key)
	}
}

// TestPageExporterPaginatesABandGroup verifies that a band that is a Column of its own -- a band group -- paginates the
// way the two bands it was made from would have: the list at the top of it gives up as many rows as fit and the block
// below it follows on the page where the list finishes, since placeColumn works down a stack of blocks one at a time.
func TestPageExporterPaginatesABandGroup(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	for i := range 200 {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = fmt.Sprintf("Trait %d", i+1)
		entity.Traits = append(entity.Traits, trait)
	}
	for i := range 3 {
		note := gurps.NewNote(entity, nil, false)
		note.MarkDown = fmt.Sprintf("Note %d", i+1)
		entity.Notes = append(entity.Notes, note)
	}
	// A layout of a single band holding the traits list with the notes list stacked below it.
	layout := &gurps.SheetLayout{
		Root: testContainerNode(layoutnode.Column, fxp.One,
			testContainerNode(layoutnode.Column, fxp.One,
				testBlockNode(gurps.BlockTraitsKey, fxp.One),
				testBlockNode(gurps.BlockNotesKey, fxp.One),
			),
		),
	}
	for _, key := range gurps.AllBlockKeys {
		if key != gurps.BlockTraitsKey && key != gurps.BlockNotesKey {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	layout.EnsureValidity()
	c.Equal(1, len(layout.Root.Children), "the band group must have survived validation as a single band")
	c.Equal(layoutnode.Column, layout.Root.Children[0].Type)
	entity.SheetSettings.Layout = layout

	exporter := newPageExporter(entity)
	c.True(len(exporter.pages) > 1, "the traits don't fit on a single page")
	c.Equal(columnBandPanel, bandKindOf(exporter.pages[0].Children()[0]), "the band is placed as a column")
	ranges := drawnRowRanges(exporter, gurps.BlockTraitsKey)
	c.True(len(ranges) > 1, "the traits list must be split across pages")
	next := 0
	for i, one := range ranges {
		c.Equal(next, one[0], "the rows drawn on page %d must pick up where the page before it left off", i+1)
		c.True(one[1] > one[0], "page %d must draw at least one row", i+1)
		next = one[1]
	}
	c.Equal(len(entity.Traits), next, "every trait must be drawn")

	c.Equal(1, countBlockPanels(exporter, gurps.BlockNotesKey), "the notes list must appear exactly once")
	c.Equal([][2]int{{0, len(entity.Notes)}}, drawnRowRanges(exporter, gurps.BlockNotesKey),
		"the notes list must draw all of its rows")
	last := len(exporter.pages) - 1
	c.Equal(1, len(blockPanels(exporter.pages[last].AsPanel(), gurps.BlockNotesKey)),
		"the notes list must follow the traits on the page the traits finish on")
	c.Equal(1, len(blockPanels(exporter.pages[last].AsPanel(), gurps.BlockTraitsKey)),
		"the last page must hold the end of the traits list")
}
