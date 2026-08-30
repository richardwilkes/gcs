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
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutedge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/side"
)

// newTestLayoutRegions returns a hand-built set of regions describing a page holding a two-block row above a
// full-width band, with a gap above the row, the seam of the page between the two bands and a gap below the band. Those
// two gaps are the only ones the editor ever builds: the space between two bands is a seam, since a block dropped there
// may be meant to straddle the two of them rather than to come between them.
func newTestLayoutRegions() *layoutRegions {
	pageRect := geom.NewRect(0, 0, 100, 200)
	band := &gurps.SheetLayoutNode{Type: layoutnode.Row}
	notes := &gurps.SheetLayoutNode{Type: layoutnode.Block, Key: gurps.BlockNotesKey}
	root := &gurps.SheetLayoutNode{Type: layoutnode.Column, Children: []*gurps.SheetLayoutNode{band, notes}}
	return &layoutRegions{
		pageRect: pageRect,
		leaves: []layoutLeafRegion{
			{key: gurps.BlockTraitsKey, rect: geom.NewRect(0, 10, 50, 40), bandIndex: -1},
			{key: gurps.BlockSkillsKey, rect: geom.NewRect(50, 10, 50, 40), bandIndex: -1},
			{key: gurps.BlockNotesKey, rect: geom.NewRect(0, 60, 100, 40), isRootBand: true, bandIndex: 1},
		},
		seams: []layoutSeamRegion{
			{
				parent:    root,
				first:     band,
				second:    notes,
				rect:      geom.NewRect(0, 55-layoutSeamThickness/2, 100, layoutSeamThickness),
				spanRect:  geom.NewRect(0, 10, 100, 90),
				depth:     -1,
				bandIndex: 1,
			},
		},
		gaps: []layoutGapRegion{
			newLayoutGapRegion(pageRect, 5, 0),
			newLayoutGapRegion(pageRect, 150, 2),
		},
	}
}

// TestResolveDropTarget verifies where a block being dragged would land for each of the places the pointer can be over
// the page. The drop is resolved from the regions and the point alone, so all of this can be checked without a window
// to drag in.
func TestResolveDropTarget(t *testing.T) {
	c := check.New(t)
	regions := newTestLayoutRegions()
	for _, one := range []struct {
		name       string
		draggedKey string
		pt         geom.Point
		want       dropTarget
	}{
		{
			name: "the left edge of a block",
			pt:   geom.NewPoint(3, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 10, 25, 40),
			},
		},
		{
			name: "the right edge of a block",
			pt:   geom.NewPoint(47, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Right,
				highlight: geom.NewRect(25, 10, 25, 40),
			},
		},
		{
			name: "the top edge of a block that isn't a band of its own",
			pt:   geom.NewPoint(25, 12),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Top,
				highlight: geom.NewRect(0, 10, 50, 20),
			},
		},
		{
			// Just above the seam of the page, which covers the last few pixels of the block.
			name: "the bottom edge of a block that isn't a band of its own",
			pt:   geom.NewPoint(25, 46),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Bottom,
				highlight: geom.NewRect(0, 30, 50, 20),
			},
		},
		{
			name: "a tie between the left and the top goes to the left",
			pt:   geom.NewPoint(5, 15),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 10, 25, 40),
			},
		},
		{
			name: "a tie between the right and the bottom goes to the right",
			pt:   geom.NewPoint(45, 45),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				edge:      layoutedge.Right,
				highlight: geom.NewRect(25, 10, 25, 40),
			},
		},
		{
			name:       "a block dropped on itself goes nowhere",
			draggedKey: gurps.BlockTraitsKey,
			pt:         geom.NewPoint(3, 30),
			want:       dropTarget{},
		},
		{
			// Within the outermost step of the edge only, and clear of the seam of the page above it; deeper in, the
			// band is meant as itself.
			name: "the top of a full-width band makes a band above it",
			pt:   geom.NewPoint(50, 65),
			want: dropTarget{kind: dropAsBand, bandIndex: 1, highlight: layoutBandBar(regions.pageRect, 60)},
		},
		{
			name: "the bottom of a full-width band makes a band below it",
			pt:   geom.NewPoint(50, 98),
			want: dropTarget{kind: dropAsBand, bandIndex: 2, highlight: layoutBandBar(regions.pageRect, 100)},
		},
		{
			name: "deeper inside the top of a full-width band, the band itself",
			pt:   geom.NewPoint(50, 72),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockNotesKey,
				edge:      layoutedge.Top,
				highlight: geom.NewRect(0, 60, 100, 20),
			},
		},
		{
			name: "deeper inside the bottom of a full-width band, the band itself",
			pt:   geom.NewPoint(50, 88),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockNotesKey,
				edge:      layoutedge.Bottom,
				highlight: geom.NewRect(0, 80, 100, 20),
			},
		},
		{
			name: "the side of a full-width band still shares its row",
			pt:   geom.NewPoint(2, 80),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockNotesKey,
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 60, 50, 40),
			},
		},
		{
			name: "the middle of the seam between two bands makes a band there",
			pt:   geom.NewPoint(50, 54),
			want: dropTarget{kind: dropAsBand, bandIndex: 1, highlight: layoutBandBar(regions.pageRect, 55)},
		},
		{
			name: "the space above the first band prepends a band",
			pt:   geom.NewPoint(50, 2),
			want: dropTarget{kind: dropAsBand, bandIndex: 0, highlight: regions.gaps[0].rect},
		},
		{
			name: "the space below the last band appends a band",
			pt:   geom.NewPoint(50, 180),
			want: dropTarget{kind: dropAsBand, bandIndex: 2, highlight: regions.gaps[1].rect},
		},
		{
			name: "a point off the side of the page goes nowhere",
			pt:   geom.NewPoint(-1, 30),
			want: dropTarget{},
		},
		{
			name: "a point below the page goes nowhere",
			pt:   geom.NewPoint(50, 250),
			want: dropTarget{},
		},
	} {
		c.Equal(one.want, resolveDropTarget(regions, one.pt, one.draggedKey), one.name)
	}
}

// newTestLadderRegions returns a hand-built set of regions describing a band that is a row holding a column of two
// blocks beside a third block, along with the nodes of the things in it. The left and top edges of the first block are
// shared with both the column and the band, its right edge with the column alone and its bottom edge with neither,
// which is what the ladder that picks the thing a drop is meant for walks.
func newTestLadderRegions() (regions *layoutRegions, nodesByName map[string]*gurps.SheetLayoutNode) {
	nodes := map[string]*gurps.SheetLayoutNode{
		"band":   {Type: layoutnode.Row},
		"column": {Type: layoutnode.Column},
		"traits": {Type: layoutnode.Block, Key: gurps.BlockTraitsKey},
		"skills": {Type: layoutnode.Block, Key: gurps.BlockSkillsKey},
		"notes":  {Type: layoutnode.Block, Key: gurps.BlockNotesKey},
	}
	pageRect := geom.NewRect(0, 0, 100, 200)
	return &layoutRegions{
		pageRect: pageRect,
		containers: []layoutContainerRegion{
			{node: nodes["band"], rect: geom.NewRect(0, 0, 100, 100), isRootBand: true},
			{node: nodes["column"], rect: geom.NewRect(0, 0, 50, 100), depth: 1, bandIndex: -1},
		},
		leaves: []layoutLeafRegion{
			{
				key:       gurps.BlockTraitsKey,
				node:      nodes["traits"],
				ancestors: []int{0, 1},
				rect:      geom.NewRect(0, 0, 50, 60),
				bandIndex: -1,
			},
			{
				key:       gurps.BlockSkillsKey,
				node:      nodes["skills"],
				ancestors: []int{0, 1},
				rect:      geom.NewRect(0, 61, 50, 39),
				bandIndex: -1,
			},
			{
				key:       gurps.BlockNotesKey,
				node:      nodes["notes"],
				ancestors: []int{0},
				rect:      geom.NewRect(50, 0, 50, 100),
				bandIndex: -1,
			},
		},
		gaps: []layoutGapRegion{newLayoutGapRegion(pageRect, 150, 1)},
	}, nodes
}

// TestResolveDropTargetEdgeLadder verifies that how far the pointer is from the edge of a block says which of the
// things sharing that edge the drop is meant for: the outermost container within the first step of it, the next one in
// within the second, and the block itself past the last of them or on an edge no container shares.
func TestResolveDropTargetEdgeLadder(t *testing.T) {
	c := check.New(t)
	regions, nodes := newTestLadderRegions()
	for _, one := range []struct {
		name string
		pt   geom.Point
		want dropTarget
	}{
		{
			name: "against an edge two containers share, the outermost of them",
			pt:   geom.NewPoint(2, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["band"],
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 0, 50, 100),
			},
		},
		{
			name: "a step further in, the next container in",
			pt:   geom.NewPoint(10, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["column"],
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 0, 25, 100),
			},
		},
		{
			name: "past every container sharing the edge, the block itself",
			pt:   geom.NewPoint(20, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				node:      nodes["traits"],
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 0, 25, 60),
			},
		},
		{
			name: "an edge no container shares belongs to the block however near the pointer is",
			pt:   geom.NewPoint(25, 58),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				node:      nodes["traits"],
				edge:      layoutedge.Bottom,
				highlight: geom.NewRect(0, 30, 50, 30),
			},
		},
		{
			name: "above a container that is a band of the page, a new band",
			pt:   geom.NewPoint(25, 2),
			want: dropTarget{kind: dropAsBand, bandIndex: 0, highlight: layoutBandBar(regions.pageRect, 0)},
		},
		{
			name: "a step further down, the top edge of the column within the band",
			pt:   geom.NewPoint(25, 10),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["column"],
				edge:      layoutedge.Top,
				highlight: geom.NewRect(0, 0, 50, 50),
			},
		},
		{
			name: "the far side of the band, which only the band and the block it holds share",
			pt:   geom.NewPoint(98, 50),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["band"],
				edge:      layoutedge.Right,
				highlight: geom.NewRect(50, 0, 50, 100),
			},
		},
	} {
		c.Equal(one.want, resolveDropTarget(regions, one.pt, gurps.BlockSpellsKey), one.name)
	}
}

// newTestSeamRegions returns a hand-built set of regions describing a page holding a band that is a row of two blocks
// above a full-width band, along with the nodes of the things in it. The seam between the two blocks of the row runs up
// and down between them and the seam between the two bands runs across the page, and the two cross one another where
// the row's seam reaches the bottom of its band.
func newTestSeamRegions() (regions *layoutRegions, nodesByName map[string]*gurps.SheetLayoutNode) {
	nodes := map[string]*gurps.SheetLayoutNode{
		"root":   {Type: layoutnode.Column},
		"band":   {Type: layoutnode.Row},
		"traits": {Type: layoutnode.Block, Key: gurps.BlockTraitsKey},
		"skills": {Type: layoutnode.Block, Key: gurps.BlockSkillsKey},
		"notes":  {Type: layoutnode.Block, Key: gurps.BlockNotesKey},
	}
	nodes["band"].Children = []*gurps.SheetLayoutNode{nodes["traits"], nodes["skills"]}
	nodes["root"].Children = []*gurps.SheetLayoutNode{nodes["band"], nodes["notes"]}
	pageRect := geom.NewRect(0, 0, 100, 200)
	return &layoutRegions{
		pageRect: pageRect,
		containers: []layoutContainerRegion{
			{node: nodes["band"], rect: geom.NewRect(0, 0, 100, 60), isRootBand: true},
		},
		leaves: []layoutLeafRegion{
			{
				key:       gurps.BlockTraitsKey,
				node:      nodes["traits"],
				ancestors: []int{0},
				rect:      geom.NewRect(0, 0, 50, 60),
				bandIndex: -1,
			},
			{
				key:       gurps.BlockSkillsKey,
				node:      nodes["skills"],
				ancestors: []int{0},
				rect:      geom.NewRect(50, 0, 50, 60),
				bandIndex: -1,
			},
			{
				key:        gurps.BlockNotesKey,
				node:       nodes["notes"],
				rect:       geom.NewRect(0, 61, 100, 39),
				isRootBand: true,
				bandIndex:  1,
			},
		},
		seams: []layoutSeamRegion{
			{
				parent:    nodes["root"],
				first:     nodes["band"],
				second:    nodes["notes"],
				rect:      geom.NewRect(0, 60.5-layoutSeamThickness/2, 100, layoutSeamThickness),
				spanRect:  geom.NewRect(0, 0, 100, 100),
				depth:     -1,
				bandIndex: 1,
			},
			{
				parent:    nodes["band"],
				first:     nodes["traits"],
				second:    nodes["skills"],
				rect:      geom.NewRect(50-layoutSeamThickness/2, 0, layoutSeamThickness, 60),
				spanRect:  geom.NewRect(0, 0, 100, 60),
				bandIndex: -1,
				vertical:  true,
			},
		},
		gaps: []layoutGapRegion{newLayoutGapRegion(pageRect, 0, 0), newLayoutGapRegion(pageRect, 150, 2)},
	}, nodes
}

// TestResolveDropTargetSeams verifies what each third of a seam means, that the strip only reaches as far as its
// thickness, that the seam of a container wins over the page's own where the two cross, and that the seam a block being
// dragged is one of the two sides of is passed over.
func TestResolveDropTargetSeams(t *testing.T) {
	c := check.New(t)
	regions, nodes := newTestSeamRegions()
	pageSeamBar := layoutBandBar(regions.pageRect, 60.5)
	for _, one := range []struct {
		name       string
		draggedKey string
		pt         geom.Point
		want       dropTarget
	}{
		{
			name: "the left third of a seam between two bands",
			pt:   geom.NewPoint(10, 55),
			want: dropTarget{
				kind:      dropStraddle,
				parent:    nodes["root"],
				first:     nodes["band"],
				second:    nodes["notes"],
				edge:      layoutedge.Left,
				highlight: geom.NewRect(0, 0, 50, 100),
			},
		},
		{
			name: "the right third of a seam between two bands",
			pt:   geom.NewPoint(90, 55),
			want: dropTarget{
				kind:      dropStraddle,
				parent:    nodes["root"],
				first:     nodes["band"],
				second:    nodes["notes"],
				edge:      layoutedge.Right,
				highlight: geom.NewRect(50, 0, 50, 100),
			},
		},
		{
			name: "the middle third of a seam between two bands makes a band there",
			pt:   geom.NewPoint(35, 55),
			want: dropTarget{kind: dropAsBand, bandIndex: 1, highlight: pageSeamBar},
		},
		{
			name: "the top third of a seam between two blocks side by side",
			pt:   geom.NewPoint(50, 10),
			want: dropTarget{
				kind:      dropStraddle,
				parent:    nodes["band"],
				first:     nodes["traits"],
				second:    nodes["skills"],
				edge:      layoutedge.Top,
				highlight: geom.NewRect(0, 0, 100, 30),
			},
		},
		{
			name: "the bottom third of a seam between two blocks side by side",
			pt:   geom.NewPoint(50, 45),
			want: dropTarget{
				kind:      dropStraddle,
				parent:    nodes["band"],
				first:     nodes["traits"],
				second:    nodes["skills"],
				edge:      layoutedge.Bottom,
				highlight: geom.NewRect(0, 30, 100, 30),
			},
		},
		{
			name: "the middle third of a seam between two blocks side by side comes between them",
			pt:   geom.NewPoint(50, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["traits"],
				edge:      layoutedge.Right,
				highlight: geom.NewRect(50-layoutBandBarHeight/2, 0, layoutBandBarHeight, 60),
				bar:       true,
			},
		},
		{
			name: "just within the near edge of the strip, the seam still",
			pt:   geom.NewPoint(50-layoutSeamThickness/2, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				node:      nodes["traits"],
				edge:      layoutedge.Right,
				highlight: geom.NewRect(50-layoutBandBarHeight/2, 0, layoutBandBarHeight, 60),
				bar:       true,
			},
		},
		{
			name: "just past the near edge of the strip, the block behind it",
			pt:   geom.NewPoint(50-layoutSeamThickness/2-1, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockTraitsKey,
				node:      nodes["traits"],
				edge:      layoutedge.Right,
				highlight: geom.NewRect(25, 0, 25, 60),
			},
		},
		{
			// The two strips cross here, and the seam of the band is the one deeper in.
			name: "where two seams cross, the one deeper in",
			pt:   geom.NewPoint(50, 55),
			want: dropTarget{
				kind:      dropStraddle,
				parent:    nodes["band"],
				first:     nodes["traits"],
				second:    nodes["skills"],
				edge:      layoutedge.Bottom,
				highlight: geom.NewRect(0, 30, 100, 30),
			},
		},
		{
			name:       "the seam beside the block being dragged is passed over",
			draggedKey: gurps.BlockTraitsKey,
			pt:         geom.NewPoint(50, 30),
			want: dropTarget{
				kind:      dropOnEdge,
				key:       gurps.BlockSkillsKey,
				node:      nodes["skills"],
				edge:      layoutedge.Left,
				highlight: geom.NewRect(50, 0, 25, 60),
			},
		},
		{
			name:       "the seam beside the band being dragged is passed over",
			draggedKey: gurps.BlockNotesKey,
			pt:         geom.NewPoint(10, 60.5),
			want:       dropTarget{kind: dropAsBand, bandIndex: 0, highlight: regions.gaps[0].rect},
		},
	} {
		draggedKey := one.draggedKey
		if draggedKey == "" {
			draggedKey = gurps.BlockSpellsKey
		}
		c.Equal(one.want, resolveDropTarget(regions, one.pt, draggedKey), one.name)
	}
}

// TestResolveDropTargetWithoutRegions verifies that a drop resolved before anything has been worked out goes nowhere
// rather than panicking.
func TestResolveDropTargetWithoutRegions(t *testing.T) {
	c := check.New(t)
	c.Equal(dropTarget{}, resolveDropTarget(nil, geom.NewPoint(5, 5), gurps.BlockNotesKey),
		"a drop resolved without regions must go nowhere")
}

// newTestSheetForLayoutEditing returns a sheet that is in block layout editing mode, leaving that mode when the test
// finishes.
func newTestSheetForLayoutEditing(t *testing.T) (*Sheet, *sheetLayoutEditor) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	sheet.toggleLayoutEditing()
	t.Cleanup(func() {
		if sheet.layoutEditing() {
			sheet.toggleLayoutEditing()
		}
	})
	return sheet, sheet.layoutEditor
}

// undoEditCount returns the number of edits the undo manager is holding, leaving it exactly as it was found.
func undoEditCount(mgr *unison.UndoManager) int {
	count := 0
	for mgr.CanUndo() {
		mgr.Undo()
		count++
	}
	for range count {
		mgr.Redo()
	}
	return count
}

// findTestDivider returns the divider whose left-hand block is the one with the given key.
func findTestDivider(regions *layoutRegions, key string) *layoutDividerRegion {
	for i := range regions.dividers {
		if regions.dividers[i].left.Type == layoutnode.Block && regions.dividers[i].left.Key == key {
			return &regions.dividers[i]
		}
	}
	return nil
}

// findTestSeam returns the seam between the two given nodes, whichever way round they are, or nil if there is none.
func findTestSeam(regions *layoutRegions, first, second *gurps.SheetLayoutNode) *layoutSeamRegion {
	for i := range regions.seams {
		seam := &regions.seams[i]
		if (seam.first == first && seam.second == second) || (seam.first == second && seam.second == first) {
			return seam
		}
	}
	return nil
}

// TestLayoutRegionsMatchThePage verifies that the regions describe exactly the blocks that are on the page, that each
// full-width band is tied to the band of the model it was built from, and that a row of two blocks has a divider and a
// seam between them.
func TestLayoutRegionsMatchThePage(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	layout := sheet.Entity().SheetSettings.Layout
	regions := editor.ensureRegions()

	expected := make([]string, 0, len(gurps.AllBlockKeys))
	for _, key := range layout.VisibleKeys() {
		if !isEmptyDerivedList(sheet, key) {
			expected = append(expected, key)
		}
	}
	c.True(isEmptyDerivedList(sheet, gurps.BlockReactionsKey),
		"the test requires a block that the builder leaves off the page")
	actual := make([]string, 0, len(regions.leaves))
	for i := range regions.leaves {
		actual = append(actual, regions.leaves[i].key)
	}
	slices.Sort(expected)
	slices.Sort(actual)
	c.Equal(expected, actual, "there must be one region per block that is actually on the page")

	for i := range regions.leaves {
		leaf := &regions.leaves[i]
		c.True(leaf.rect.Width > 0 && leaf.rect.Height > 0, "the region of %q must have a size", leaf.key)
		if leaf.isRootBand {
			c.True(leaf.bandIndex >= 0 && leaf.bandIndex < len(layout.Root.Children),
				"the band index of %q must name one of the model's bands", leaf.key)
			c.True(leaf.node == layout.Root.Children[leaf.bandIndex],
				"the band index of %q must point at the band it was built from", leaf.key)
		}
	}

	for i := range regions.containers {
		container := &regions.containers[i]
		c.True(container.rect.Width > 0 && container.rect.Height > 0, "every container region must have a size")
		c.True(container.isRootBand == (container.depth == 0), "only the outermost containers are bands")
	}
	for i := range regions.leaves {
		leaf := &regions.leaves[i]
		for depth, index := range leaf.ancestors {
			c.Equal(depth, regions.containers[index].depth,
				"the containers around %q must be listed outermost first", leaf.key)
		}
		if len(leaf.ancestors) == 0 {
			c.True(leaf.isRootBand, "a block with no container around it is a band of its own")
		} else {
			c.True(regions.containers[leaf.ancestors[0]].isRootBand,
				"the outermost container around %q must be a band", leaf.key)
		}
	}
	identity := regions.leafFor(gurps.BlockIdentityKey)
	c.NotNil(identity, "the identity block must be on the page")
	c.Equal(3, len(identity.ancestors),
		"the identity block sits inside the band, the column beside the portrait and the row it shares")

	spells := regions.leafFor(gurps.BlockSpellsKey)
	c.NotNil(spells, "the spells block must be on the page")
	c.True(spells.isRootBand, "the spells block is a band of its own in the default layout")
	c.Equal(0, len(spells.ancestors), "a block that is a band of its own sits inside nothing")

	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	c.False(traits.isRootBand, "the traits block shares a band with the skills block in the default layout")

	divider := findTestDivider(regions, gurps.BlockTraitsKey)
	c.NotNil(divider, "the traits and skills blocks must have a divider between them")
	c.Equal(gurps.BlockSkillsKey, divider.right.Key, "the divider must sit between the traits and skills blocks")
	c.True(divider.rect.Width > 0 && divider.rect.Height > 0, "the divider must have a size")

	seam := findTestSeam(regions, traits.node, regions.leafFor(gurps.BlockSkillsKey).node)
	c.NotNil(seam, "the traits and skills blocks must have a seam between them")
	c.True(seam.vertical, "two blocks side by side have a seam that runs up and down")
	c.Equal(layoutSeamThickness, seam.rect.Width, "the seam must be a strip of the standard thickness")
	c.Equal(-1, seam.bandIndex, "only a seam between two of the page's bands names a band")
	for i := range regions.seams {
		s := &regions.seams[i]
		c.True(s.rect.Width > 0 && s.rect.Height > 0, "every seam must have a size")
		c.True(s.spanRect.Width >= s.rect.Width && s.spanRect.Height >= s.rect.Height,
			"a seam must reach no further than the pair it lies between")
		if s.depth < 0 {
			c.True(s.bandIndex > 0 && s.bandIndex < len(layout.Root.Children),
				"a seam of the page must name the band below it")
			c.True(s.parent == layout.Root, "a seam of the page belongs to the root")
		}
	}

	c.Equal(2, len(regions.gaps),
		"the only gaps are the one above the first band and the one below the last, since between two bands is a seam")
	c.Equal(0, regions.gaps[0].index, "the first gap must prepend a band")
	c.Equal(len(layout.Root.Children), regions.gaps[len(regions.gaps)-1].index,
		"the last gap must append a band past the end")
}

// TestPageFooterNumberingIgnoresTheOverlay verifies that the overlay the layout editor puts beside the sheet's page
// doesn't make the page's footer call it the second of two -- which, the page then being even-numbered, would also
// swap the two halves of the footer around.
func TestPageFooterNumberingIgnoresTheOverlay(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	c.Equal(2, len(sheet.content.Children()), "the test requires the overlay to sit beside the page")
	c.Equal(editor.overlay, sheet.content.Children()[0], "and in front of it")
	number, count := sheet.page.pageNumbering()
	c.Equal(1, number, "the page must still be the first")
	c.Equal(1, count, "of one")
	sheet.toggleLayoutEditing()
	number, count = sheet.page.pageNumbering()
	c.Equal(1, number, "and remain so once editing ends")
	c.Equal(1, count)
}

// TestSecondButtonDuringAGestureIsIgnored verifies that a press of another button while a divider drag is under way
// neither abandons the drag -- which would leave the weights it had already written in place with nothing recorded to
// undo them -- nor commits it on that button's release, and that the release of the button that began the drag still
// commits it as one undoable edit. The reverse holds too: a press while a context menu is pending starts nothing.
func TestSecondButtonDuringAGestureIsIgnored(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	mgr := sheet.UndoManager()
	divider := findTestDivider(editor.ensureRegions(), gurps.BlockTraitsKey)
	c.NotNil(divider, "the traits and skills blocks must have a divider between them")
	start := divider.rect.Center()
	moved := geom.NewPoint(start.X+40, start.Y)

	editor.mouseDown(start, unison.ButtonLeft)
	c.Equal(layoutDraggingDivider, editor.mode, "a press on the divider must begin a drag")
	editor.mouseDrag(moved)
	c.NotNil(editor.beforeLayout, "the drag must be holding the layout it started from")
	editor.mouseDown(moved, unison.ButtonRight)
	c.Equal(layoutDraggingDivider, editor.mode, "a second button must not abandon the drag")
	c.NotNil(editor.beforeLayout, "nor throw away the layout it started from")
	editor.mouseUp(moved, unison.ButtonRight)
	c.Equal(layoutDraggingDivider, editor.mode, "nor may its release end the drag")
	c.False(mgr.CanUndo(), "nothing may be recorded before the button that began the drag is released")
	editor.mouseUp(moved, unison.ButtonLeft)
	c.Equal(layoutIdle, editor.mode, "the release of the button that began the drag must end it")
	c.Nil(editor.beforeLayout, "and leave nothing of it behind")
	c.Equal(1, undoEditCount(mgr), "the resize must be recorded as one undoable edit")
	traits, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockTraitsKey)
	c.True(traits.Weight > fxp.One, "the resize must have been kept")

	editor.mouseDown(start, unison.ButtonRight)
	c.Equal(layoutContextPending, editor.mode, "a right press must wait for its release to show the context menu")
	editor.mouseDown(start, unison.ButtonLeft)
	c.Equal(layoutContextPending, editor.mode, "a second button must not start a gesture on top of that")
	editor.mouseDrag(moved)
	c.Equal(layoutContextPending, editor.mode, "nor may dragging it")
	editor.mouseUp(moved, unison.ButtonLeft)
	c.Equal(layoutContextPending, editor.mode, "nor may its release end what the right button began")
	editor.mouseUp(start, unison.ButtonRight)
	c.Equal(layoutIdle, editor.mode, "the release of the right button must end it")
	c.Equal(1, undoEditCount(mgr), "and nothing may have been resized along the way")
}

// TestSeamBelongsToTheContainerItLiesIn verifies that the seam between two blocks names, as its parent, the container
// the model holds the two of them in, even when that container's panel has taken over the slot of an outer container
// the builder dropped -- here a column whose other child, an empty derived list, was left off the page -- so that the
// node the slot is governed by and the node the pair belongs to differ.
func TestSeamBelongsToTheContainerItLiesIn(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t, testContainerNode(layoutnode.Column, fxp.One,
		testContainerNode(layoutnode.Row, fxp.One,
			testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One),
		),
		testBlockNode(gurps.BlockReactionsKey, fxp.One),
	))
	c.True(isEmptyDerivedList(sheet, gurps.BlockReactionsKey),
		"the test requires a block that the builder leaves off the page")
	layout := sheet.Entity().SheetSettings.Layout
	traitsNode, row, _ := layout.Find(gurps.BlockTraitsKey)
	c.NotNil(row, "the traits block must be in the tree")
	c.Equal(layoutnode.Row, row.Type, "held by the row")
	column, _ := layout.Parent(row)
	c.NotNil(column, "which the column must hold")
	c.Equal(layoutnode.Column, column.Type)
	skillsNode, _, _ := layout.Find(gurps.BlockSkillsKey)

	regions := editor.ensureRegions()
	c.Equal(1, len(regions.containers), "the column dropped onto the row leaves one container on the page")
	c.Equal(column, regions.containers[0].node, "whose slot the column's node governs")
	seam := findTestSeam(regions, traitsNode, skillsNode)
	c.NotNil(seam, "the two blocks must have a seam between them")
	c.Equal(row, seam.parent, "which belongs to the row the model holds them in, not to the column that governs the slot")
	target := resolveDropTarget(regions, geom.NewPoint(seam.rect.CenterX(), seam.rect.Y+1), gurps.BlockNotesKey)
	c.Equal(dropStraddle, target.kind, "the end of the seam must offer to straddle the pair")
	c.Equal(row, target.parent, "in the container the model holds them in")
}

// TestHideAndShowBlock verifies that taking a block off the sheet and putting it back are each one undoable edit that
// updates the sheet exactly once, and that both can still be undone once editing mode has been left.
func TestHideAndShowBlock(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	layout := sheet.Entity().SheetSettings.Layout
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	before := layout.VisibleKeys()

	counter.count = 0
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.Equal(1, counter.count, "hiding a block must sync the sheet exactly once")
	c.False(layout.Contains(gurps.BlockNotesKey), "the hidden block must be out of the tree")
	c.Equal([]string{gurps.BlockNotesKey}, layout.HiddenKeys(), "the hidden block must be recorded as hidden")
	c.Nil(sheet.blockPanel(gurps.BlockNotesKey).AsPanel().Parent(), "the hidden block must be off the page")
	c.Nil(editor.ensureRegions().leafFor(gurps.BlockNotesKey), "the hidden block must have no region")

	counter.count = 0
	sheet.showLayoutBlock(gurps.BlockNotesKey)
	c.Equal(1, counter.count, "showing a block must sync the sheet exactly once")
	c.True(layout.Contains(gurps.BlockNotesKey), "the shown block must be back in the tree")
	c.Equal(0, len(layout.HiddenKeys()), "nothing must be left hidden")
	c.Equal(2, undoEditCount(mgr), "hiding and showing a block must be two undoable edits")

	// An undo can be asked for from the Edit menu long after editing mode was left, so it has to work from there.
	sheet.toggleLayoutEditing()
	c.False(sheet.layoutEditing(), "editing mode must have been left")

	mgr.Undo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"undoing the show must take the block back off the sheet")
	mgr.Undo()
	c.Equal(before, sheet.Entity().SheetSettings.Layout.VisibleKeys(),
		"undoing the hide must put the sheet back the way it was")
	mgr.Redo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"redoing the hide must take the block off the sheet again")
	mgr.Redo()
	c.Equal(before, sheet.Entity().SheetSettings.Layout.VisibleKeys(), "redoing the show must put the block back")
}

// TestHideBlockThatIsAlreadyHiddenDoesNothing verifies that an edit that changes nothing is neither recorded nor shown.
func TestHideBlockThatIsAlreadyHiddenDoesNothing(t *testing.T) {
	c := check.New(t)
	sheet, _ := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)

	counter.count = 0
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.Equal(0, counter.count, "hiding a block that is already hidden must not sync the sheet")
	c.Equal(1, undoEditCount(sheet.UndoManager()), "hiding a block that is already hidden must not be recorded")
}

// TestMoveBlockByDragging verifies that dropping a block against the edge of another one moves it there as a single
// undoable edit that updates the sheet exactly once, and that the move round trips through undo and redo.
func TestMoveBlockByDragging(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	before := sheet.Entity().SheetSettings.Layout.VisibleKeys()
	regions := editor.ensureRegions()
	spells := regions.leafFor(gurps.BlockSpellsKey)
	c.NotNil(spells, "the spells block must be on the page")
	notes := regions.leafFor(gurps.BlockNotesKey)
	c.NotNil(notes, "the notes block must be on the page")

	counter.count = 0
	editor.beginBlockDrag(gurps.BlockSpellsKey, spells.rect.Center())
	editor.endBlockDrag(geom.NewPoint(notes.rect.X+1, notes.rect.CenterY()))

	c.Equal(1, counter.count, "moving a block must sync the sheet exactly once")
	node, parent, index := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockSpellsKey)
	c.NotNil(node, "the moved block must still be in the tree")
	c.Equal(layoutnode.Row, parent.Type, "the moved block must have joined a row")
	c.Equal(0, index, "the moved block must sit to the left of the block it was dropped on")
	c.Equal(gurps.BlockNotesKey, parent.Children[1].Key, "the moved block must sit beside the notes block")

	mgr.Undo()
	c.Equal(before, sheet.Entity().SheetSettings.Layout.VisibleKeys(), "undo must put the block back where it was")
	mgr.Redo()
	_, parent, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockSpellsKey)
	c.Equal(layoutnode.Row, parent.Type, "redo must move the block back beside the notes block")
	c.Equal(1, undoEditCount(mgr), "moving a block must be one undoable edit")
}

// layoutTreeStringOf renders a layout node and everything below it as a compact string, so that a test can state the
// shape it expects in one readable line.
func layoutTreeStringOf(node *gurps.SheetLayoutNode) string {
	if node == nil {
		return "<nil>"
	}
	if node.Type == layoutnode.Block {
		return node.Key
	}
	parts := make([]string, 0, len(node.Children))
	for _, child := range node.Children {
		parts = append(parts, layoutTreeStringOf(child))
	}
	return node.Type.Key() + "[" + strings.Join(parts, " ") + "]"
}

// TestMoveBlockBesideAColumn verifies that a block dropped a step inside the left edge of the primary attributes block
// lands beside the whole of the column that block is in, spanning everything in it, since that column's left edge is
// the edge the pointer is on. The band's own left edge is in the same place, so the column is the second rung of the
// ladder rather than the first; the first step of an edge two things share this way is what a drop right against it
// means.
func TestMoveBlockBesideAColumn(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	layout := sheet.Entity().SheetSettings.Layout
	regions := editor.ensureRegions()
	primary := regions.leafFor(gurps.BlockPrimaryAttributesKey)
	c.NotNil(primary, "the primary attributes block must be on the page")
	body := regions.leafFor(gurps.BlockBodyKey)
	c.NotNil(body, "the body type block must be on the page")
	column := layout.Root.Children[1].Children[0]
	c.Equal(layoutnode.Column, column.Type,
		"the primary attributes block's column must be the first thing in the second band")

	where := geom.NewPoint(primary.rect.X+3*layoutEdgeLadderStep/2, primary.rect.CenterY())
	target := resolveDropTarget(regions, where, gurps.BlockBodyKey)
	c.Equal(dropOnEdge, target.kind)
	c.Equal(layoutedge.Left, target.edge)
	c.True(column == target.node, "the drop must be against the column rather than against the block within it")

	editor.beginBlockDrag(gurps.BlockBodyKey, body.rect.Center())
	editor.endBlockDrag(where)
	c.Equal("row[body column[row[column[primary_attributes damage] secondary_attributes] point_pools] "+
		"column[encumbrance lifting]]",
		layoutTreeStringOf(sheet.Entity().SheetSettings.Layout.Root.Children[1]),
		"the block must stand beside the whole column")
}

// TestMoveBlockBesideARow verifies that a block dropped on the middle of the seam below the identity block lands below
// the whole of the row that block is in, spanning it, since coming between the two things a seam lies between is
// joining their container between them.
func TestMoveBlockBesideARow(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	layout := sheet.Entity().SheetSettings.Layout
	regions := editor.ensureRegions()
	identity := regions.leafFor(gurps.BlockIdentityKey)
	c.NotNil(identity, "the identity block must be on the page")
	description := regions.leafFor(gurps.BlockDescriptionKey)
	c.NotNil(description, "the description block must be on the page")
	lifting := regions.leafFor(gurps.BlockLiftingKey)
	c.NotNil(lifting, "the lifting block must be on the page")
	row := layout.Root.Children[0].Children[1].Children[0]
	c.Equal(layoutnode.Row, row.Type, "the identity block's row must be the first thing in its column")

	where := geom.NewPoint(description.rect.CenterX(), identity.rect.Bottom()-2)
	target := resolveDropTarget(regions, where, gurps.BlockLiftingKey)
	c.Equal(dropOnEdge, target.kind)
	c.Equal(layoutedge.Bottom, target.edge)
	c.True(row == target.node, "the drop must be against the row rather than against the identity block")
	c.True(target.bar, "coming between two things is shown as a bar rather than as an area")

	editor.beginBlockDrag(gurps.BlockLiftingKey, lifting.rect.Center())
	editor.endBlockDrag(where)
	c.Equal("column[row[identity miscellaneous] lifting description]",
		layoutTreeStringOf(sheet.Entity().SheetSettings.Layout.Root.Children[0].Children[1]),
		"the block must span the full width of the row it was dropped below")
	c.Equal("row[column[row[column[primary_attributes damage] secondary_attributes] point_pools] body encumbrance]",
		layoutTreeStringOf(sheet.Entity().SheetSettings.Layout.Root.Children[1]),
		"the column the block came out of must collapse onto what is left of it")
}

// TestBlockDroppedWhereItStartedChangesNothing verifies that picking a block up and putting it back down without
// moving it neither records an edit nor rebuilds the sheet.
func TestBlockDroppedWhereItStartedChangesNothing(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	spells := editor.ensureRegions().leafFor(gurps.BlockSpellsKey)
	c.NotNil(spells, "the spells block must be on the page")

	counter.count = 0
	editor.beginBlockDrag(gurps.BlockSpellsKey, spells.rect.Center())
	editor.endBlockDrag(spells.rect.Center())
	c.Equal(0, counter.count, "a block dropped on itself must not sync the sheet")
	c.False(sheet.UndoManager().CanUndo(), "a block dropped on itself must not be recorded")
}

// TestDividerDragTransfersWidth verifies that dragging the divider between two blocks moves width from one of them to
// the other without altering how much of the row the pair takes up, and that a divider put back where it started
// leaves nothing to undo.
func TestDividerDragTransfersWidth(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	mgr := sheet.UndoManager()
	divider := findTestDivider(editor.ensureRegions(), gurps.BlockTraitsKey)
	c.NotNil(divider, "the traits and skills blocks must have a divider between them")
	left := divider.left
	right := divider.right
	total := left.Weight + right.Weight
	start := geom.NewPoint(divider.rect.CenterX(), divider.rect.CenterY())

	editor.beginDividerDrag(divider, start)
	editor.updateDividerDrag(geom.NewPoint(start.X+40, start.Y))
	c.Equal(total, left.Weight+right.Weight, "a divider drag must not change how much of the row the pair takes up")
	c.True(left.Weight > right.Weight, "dragging the divider to the right must widen the left-hand block")

	editor.updateDividerDrag(start)
	editor.endDividerDrag(start)
	c.Equal(total, left.Weight+right.Weight, "the weights must still add up to what they did")
	c.False(mgr.CanUndo(), "a divider put back where it started must not be recorded")

	counter := installSyncCounter(sheet)
	counter.count = 0
	editor.beginDividerDrag(divider, start)
	editor.endDividerDrag(geom.NewPoint(start.X+40, start.Y))
	c.Equal(1, counter.count, "resizing a block must sync the sheet exactly once")
	c.Equal(total, left.Weight+right.Weight, "the weights must still add up to what they did")

	mgr.Undo()
	traits, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockTraitsKey)
	c.Equal(fxp.One, traits.Weight, "undo must put the original weight back")

	mgr.Redo()
	traits, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockTraitsKey)
	c.True(traits.Weight > fxp.One, "redo must widen the block again")
	c.Equal(1, undoEditCount(mgr), "resizing a block must be one undoable edit")
}

// TestBottomEdgeDragSetsMinimumHeight verifies that dragging a block's bottom edge past its natural height gives it a
// minimum height, that leaving it within a few pixels of its natural height gives it that height back, and that the
// whole gesture is a single undoable edit.
func TestBottomEdgeDragSetsMinimumHeight(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	notes := editor.ensureRegions().leafFor(gurps.BlockNotesKey)
	c.NotNil(notes, "the notes block must be on the page")
	natural := notes.naturalHeight
	node, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockNotesKey)
	c.NotNil(node, "the notes block must be in the tree")
	c.Equal(float64(0), node.MinHeight.Length, "the notes block must start out at its natural height")

	counter.count = 0
	editor.beginBottomDrag(notes, geom.NewPoint(notes.rect.CenterX(), notes.rect.Bottom()))
	editor.updateBottomDrag(geom.NewPoint(notes.rect.CenterX(), notes.rect.Y+natural+layoutMinHeightSlop-1))
	c.Equal(float64(0), node.MinHeight.Length, "a bottom edge left within the slop must leave the natural height")

	wanted := notes.rect.Y + natural + 100
	editor.updateBottomDrag(geom.NewPoint(notes.rect.CenterX(), wanted))
	c.True(xmath.Abs(node.MinHeight.Pixels()-(natural+100)) < 0.01,
		"a bottom edge dragged past the slop must set the minimum height it was left at")

	editor.endBottomDrag(geom.NewPoint(notes.rect.CenterX(), wanted))
	c.Equal(1, counter.count, "setting a block's height must sync the sheet exactly once")

	mgr.Undo()
	node, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockNotesKey)
	c.Equal(float64(0), node.MinHeight.Length, "undo must give the block back its natural height")
	mgr.Redo()
	node, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockNotesKey)
	c.True(node.MinHeight.Pixels() > natural, "redo must give the block its minimum height back")
	c.Equal(1, undoEditCount(mgr), "setting a block's height must be one undoable edit")
}

// TestResetLayoutRestoresTheDefault verifies that resetting returns the sheet to the layout new sheets are given, as
// one undoable edit that updates the sheet exactly once.
func TestResetLayoutRestoresTheDefault(t *testing.T) {
	c := check.New(t)
	sheet, _ := newTestSheetForLayoutEditing(t)
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	original := gurps.Hash64(sheet.Entity().SheetSettings.Layout)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	c.NotEqual(original, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "the layout must have been altered")

	counter.count = 0
	sheet.resetLayout()
	c.Equal(1, counter.count, "resetting the layout must sync the sheet exactly once")
	c.Equal(gurps.Hash64(gurps.GlobalSettings().Sheet.Layout), gurps.Hash64(sheet.Entity().SheetSettings.Layout),
		"resetting must give the sheet the layout new sheets are given")

	mgr.Undo()
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"undoing the reset must take the block back off the sheet")
	mgr.Redo()
	c.Equal(original, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "redoing the reset must restore the default")
	c.Equal(2, undoEditCount(mgr), "resetting the layout must be one more undoable edit")
}

// TestEditingModeSurvivesARebuild verifies that a rebuild leaves the overlay on top of the page and the size of it,
// since everything on the page is replaced by one and the page's own size can change along with it.
func TestEditingModeSurvivesARebuild(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	overlay := editor.overlay
	c.NotNil(overlay, "the overlay must be in place")
	c.Equal(0, sheet.content.IndexOfChild(overlay), "the overlay must be on top of the page")

	sheet.Rebuild(true)

	c.True(sheet.layoutEditing(), "a rebuild must not leave editing mode")
	c.True(overlay == editor.overlay, "a rebuild must not replace the overlay")
	c.Equal(0, sheet.content.IndexOfChild(overlay), "the overlay must still be on top of the page")
	c.Equal(sheet.page.FrameRect(), overlay.FrameRect(), "the overlay must still cover the page exactly")
	c.NotNil(editor.ensureRegions().leafFor(gurps.BlockNotesKey), "the regions must describe the rebuilt page")
}

// TestLeavingEditingModeTakesTheOverlayAway verifies that the sheet goes back to being an ordinary sheet.
func TestLeavingEditingModeTakesTheOverlayAway(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	overlay := editor.overlay
	c.NotNil(overlay, "the overlay must be in place")

	sheet.toggleLayoutEditing()

	c.False(sheet.layoutEditing(), "editing mode must have been left")
	c.Nil(overlay.Parent(), "the overlay must have been taken off the content")
	c.Nil(sheet.contentLayout.overlay, "the content's layout must no longer be laying the overlay out")
	c.Equal(1, len(sheet.content.Children()), "only the page must be left")
}

// sheetLayoutGivenToNewEntity returns the block layout an entity with no embedded sheet settings is given when it is
// unmarshaled. It comes from the published snapshot of the global sheet settings rather than the live settings, so it
// is what shows whether a change to the default layout was published.
func sheetLayoutGivenToNewEntity(t *testing.T) *gurps.SheetLayout {
	t.Helper()
	var e gurps.Entity
	e.DiscardCaches()
	check.New(t).NoError(jio.Unmarshal(fmt.Appendf(nil, `{"version":%d}`, jio.CurrentDataVersion), &e))
	return e.SheetSettings.Layout
}

// restoreDefaultSheetLayout puts the given layout back as the default for new sheets when the test finishes,
// republishing it, since the default the test installed was published as well.
func restoreDefaultSheetLayout(t *testing.T, layout *gurps.SheetLayout) {
	t.Helper()
	t.Cleanup(func() {
		gurps.GlobalSettings().Sheet.Layout = layout
		gurps.SyncGlobalSheetSettings()
	})
}

// TestUseLayoutAsDefaultStoresACopy verifies that making a sheet's layout the default for new sheets stores a copy of
// it rather than the sheet's own, publishes it for the sheets opened from then on, and tells everything that lays
// itself out from the defaults about the change.
func TestUseLayoutAsDefaultStoresACopy(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	restoreDefaultSheetLayout(t, global.Sheet.Layout)

	sheet, _ := newTestSheetForLayoutEditing(t)
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	recorder.entities = nil
	recorder.updates = nil
	c.True(sheetLayoutGivenToNewEntity(t).Contains(gurps.BlockNotesKey),
		"the published default must start out showing the block the sheet hides, or the test proves nothing")

	sheet.useLayoutAsDefault()

	layout := sheet.Entity().SheetSettings.Layout
	c.Equal(gurps.Hash64(layout), gurps.Hash64(global.Sheet.Layout), "the default must match the sheet's layout")
	c.True(layout != global.Sheet.Layout, "the default must be a copy, not the sheet's own layout")
	c.Equal(gurps.Hash64(layout), gurps.Hash64(sheetLayoutGivenToNewEntity(t)),
		"a sheet without embedded settings opened from now on must be given the new default")
	c.Equal(1, len(recorder.updates), "everything that lays itself out from the defaults must be told once")
	c.True(recorder.updates[0], "the notification must ask for a full rebuild")
	c.Nil(recorder.entities[0], "a nil entity is what says the change was to the defaults")

	// Neither of the two can be changed by way of the other.
	global.Sheet.Layout.Hide(gurps.BlockSpellsKey)
	c.True(layout.Contains(gurps.BlockSpellsKey), "changing the default must not change the sheet")
}

// TestResetDefaultLayoutRestoresTheFactoryLayout verifies that resetting the default layout puts the factory layout in
// place for new sheets and publishes it for the sheets opened from then on, leaves the sheet's own layout alone, and
// tells everything that lays itself out from the defaults about the change.
func TestResetDefaultLayoutRestoresTheFactoryLayout(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings()
	restoreDefaultSheetLayout(t, global.Sheet.Layout)

	sheet, _ := newTestSheetForLayoutEditing(t)
	recorder := &sheetSettingsRecorder{}
	recorder.Self = recorder
	Workspace.DocumentDock.DockTo(recorder, nil, side.Left)
	sheet.hideLayoutBlock(gurps.BlockNotesKey)
	sheet.useLayoutAsDefault()
	c.False(global.Sheet.Layout.Contains(gurps.BlockNotesKey), "the default must start out altered")
	c.False(sheetLayoutGivenToNewEntity(t).Contains(gurps.BlockNotesKey),
		"the published default must start out altered too, or the test proves nothing")
	recorder.entities = nil
	recorder.updates = nil

	sheet.resetDefaultLayout()

	c.Equal(gurps.Hash64(gurps.FactorySheetLayout()), gurps.Hash64(global.Sheet.Layout),
		"the default must be the factory layout again")
	c.Equal(gurps.Hash64(gurps.FactorySheetLayout()), gurps.Hash64(sheetLayoutGivenToNewEntity(t)),
		"a sheet without embedded settings opened from now on must be given the factory layout again")
	c.False(sheet.Entity().SheetSettings.Layout.Contains(gurps.BlockNotesKey),
		"the sheet's own layout must be left alone")
	c.Equal(1, len(recorder.updates), "everything that lays itself out from the defaults must be told once")
	c.True(recorder.updates[0], "the notification must ask for a full rebuild")
	c.Nil(recorder.entities[0], "a nil entity is what says the change was to the defaults")
}

// TestCancelDragPutsTheLayoutBack verifies that abandoning a gesture partway through restores the layout it started
// from, which is what the Escape key does.
func TestCancelDragPutsTheLayoutBack(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	before := gurps.Hash64(sheet.Entity().SheetSettings.Layout)
	// A date no rebuild could arrive at on its own, so that one that bumps it can't be mistaken for one that didn't.
	modified := jio.Time(time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC))
	sheet.Entity().ModifiedOn = modified
	divider := findTestDivider(editor.ensureRegions(), gurps.BlockTraitsKey)
	c.NotNil(divider, "the traits and skills blocks must have a divider between them")
	start := geom.NewPoint(divider.rect.CenterX(), divider.rect.CenterY())

	editor.beginDividerDrag(divider, start)
	editor.updateDividerDrag(geom.NewPoint(start.X+40, start.Y))
	c.NotEqual(before, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "the drag must have altered the layout")

	editor.cancelDrag()
	c.Equal(before, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "canceling must put the layout back")
	c.False(sheet.UndoManager().CanUndo(), "a canceled drag must not be recorded")
	c.Equal(layoutIdle, editor.mode, "canceling must leave nothing in progress")
	c.Equal(modified, sheet.Entity().ModifiedOn,
		"a canceled drag must leave the modification date alone, since nothing about the sheet changed")
}

// newTestSheetWithLayout returns a sheet in block layout editing mode whose layout is the given bands, with every block
// they don't hold hidden and every block they do an inch tall, so that a test points at a page whose shape it knows and
// every band is deep enough to have an inside as well as an edge.
func newTestSheetWithLayout(t *testing.T, bands ...*gurps.SheetLayoutNode) (*Sheet, *sheetLayoutEditor) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	layout := &gurps.SheetLayout{Root: testContainerNode(layoutnode.Column, fxp.One, bands...)}
	keys := layout.VisibleKeys()
	for _, key := range gurps.AllBlockKeys {
		if !slices.Contains(keys, key) {
			layout.Hidden = append(layout.Hidden, key)
		}
	}
	layout.EnsureValidity()
	for _, key := range keys {
		layout.SetMinHeight(key, paper.Length{Length: 1, Units: paper.Inch})
	}
	sheet.Entity().SheetSettings.Layout = layout
	sheet.Rebuild(true)
	sheet.toggleLayoutEditing()
	t.Cleanup(func() {
		if sheet.layoutEditing() {
			sheet.toggleLayoutEditing()
		}
	})
	return sheet, sheet.layoutEditor
}

// newTestSheetWithLeafBands returns a sheet in block layout editing mode whose layout holds nothing but the given
// blocks, each a band of its own.
func newTestSheetWithLeafBands(t *testing.T, keys ...string) (*Sheet, *sheetLayoutEditor) {
	t.Helper()
	bands := make([]*gurps.SheetLayoutNode, 0, len(keys))
	for _, key := range keys {
		bands = append(bands, testBlockNode(key, fxp.One))
	}
	return newTestSheetWithLayout(t, bands...)
}

// dragBlockTo picks the block with the given key up from the middle of wherever it is and drops it at the given point.
func dragBlockTo(t *testing.T, editor *sheetLayoutEditor, key string, where geom.Point) {
	t.Helper()
	leaf := editor.ensureRegions().leafFor(key)
	if leaf == nil {
		t.Fatalf("the %q block must be on the page to be dragged", key)
	}
	editor.beginBlockDrag(key, leaf.rect.Center())
	editor.endBlockDrag(where)
}

// layoutTreeOf returns the sheet's block layout tree in the form layoutTreeStringOf renders.
func layoutTreeOf(sheet *Sheet) string {
	return layoutTreeStringOf(sheet.Entity().SheetSettings.Layout.Root)
}

// TestDropDeepInsideABandStacksTheTwoIntoOneBand verifies that a block dropped on the half of a full-width band that is
// deeper than the outermost step of the edge ladder stacks the two of them into a band group, and that a third block
// dropped against the edge of that group then spans both of them. The block being dragged is the band just below the
// one being dropped on, so the seam between the two of them is passed over and the ladder is what the drop lands on:
// the top and bottom edges of a band belong to the page, so a drop within the outermost step of one means a new band.
func TestDropDeepInsideABandStacksTheTwoIntoOneBand(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLeafBands(t, gurps.BlockTraitsKey, gurps.BlockSkillsKey, gurps.BlockNotesKey)
	c.Equal("column[traits skills notes]", layoutTreeOf(sheet), "each block must start out as a band of its own")
	regions := editor.ensureRegions()
	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	c.True(traits.isRootBand, "the traits block must be a band of its own")
	c.True(traits.rect.Height > 2*layoutEdgeLadderStep, "the band must be deep enough to have an inside")

	deep := geom.NewPoint(traits.rect.CenterX(), traits.rect.Bottom()-3*layoutEdgeLadderStep/2)
	shallow := geom.NewPoint(traits.rect.CenterX(), traits.rect.Bottom()-layoutEdgeLadderStep/4)
	deepTarget := resolveDropTarget(regions, deep, gurps.BlockSkillsKey)
	shallowTarget := resolveDropTarget(regions, shallow, gurps.BlockSkillsKey)
	c.Equal(dropOnEdge, deepTarget.kind, "deep inside its edge, the band is meant as itself")
	c.Equal(layoutedge.Bottom, deepTarget.edge)
	c.True(traits.node == deepTarget.node, "the drop must be against the block that was pointed at")
	c.Equal(edgeHalf(traits.rect, layoutedge.Bottom), deepTarget.highlight, "the bottom half of the block is lit up")
	c.Equal(dropAsBand, shallowTarget.kind, "within the outermost step, the edge still belongs to the page")
	c.Equal(1, shallowTarget.bandIndex)
	c.Equal(layoutBandBar(regions.pageRect, traits.rect.Bottom()), shallowTarget.highlight,
		"a new band is shown as a bar across the page")
	c.NotEqual(deepTarget.highlight, shallowTarget.highlight, "the two drops must not look alike")

	counter := installSyncCounter(sheet)
	counter.count = 0
	dragBlockTo(t, editor, gurps.BlockSkillsKey, deep)
	c.Equal(1, counter.count, "grouping two bands must sync the sheet exactly once")
	c.Equal("column[column[traits skills] notes]", layoutTreeOf(sheet),
		"the two blocks must have become one band of the page")

	// The group is a container like any other, so its edge can be dropped against, which is the point of making one.
	regions = editor.ensureRegions()
	traits = regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must still be on the page")
	c.False(traits.isRootBand, "the traits block is now inside the band group rather than being a band itself")
	group := sheet.Entity().SheetSettings.Layout.Root.Children[0]
	beside := geom.NewPoint(traits.rect.X+layoutEdgeLadderStep/4, traits.rect.CenterY())
	target := resolveDropTarget(regions, beside, gurps.BlockNotesKey)
	c.Equal(dropOnEdge, target.kind)
	c.Equal(layoutedge.Left, target.edge)
	c.True(group == target.node, "the drop must be against the whole group rather than against the traits block")

	dragBlockTo(t, editor, gurps.BlockNotesKey, beside)
	c.Equal("column[row[notes column[traits skills]]]", layoutTreeOf(sheet),
		"the block must stand beside the group, spanning both of the blocks stacked in it")

	mgr := sheet.UndoManager()
	c.Equal(2, undoEditCount(mgr), "the two drops must be two undoable edits")
	mgr.Undo()
	c.Equal("column[column[traits skills] notes]", layoutTreeOf(sheet), "undo must take the block back out of the row")
	mgr.Undo()
	c.Equal("column[traits skills notes]", layoutTreeOf(sheet), "undo must take the two blocks back apart")
	mgr.Redo()
	c.Equal("column[column[traits skills] notes]", layoutTreeOf(sheet), "redo must put the group back")
	mgr.Redo()
	c.Equal("column[row[notes column[traits skills]]]", layoutTreeOf(sheet), "redo must put the block back beside it")
}

// TestDropAtTheEdgeOfABandGroupStillMakesANewBand verifies that the bottom edge of a band group means a new band of the
// page, exactly as it does for a block that is a band of its own. The strip on that edge belongs to the seam the group
// shares with the band below it, and the middle of a seam of the page is a new band between the two of them, so a band
// group never turns into a target that would nest a block inside it.
func TestDropAtTheEdgeOfABandGroupStillMakesANewBand(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLeafBands(t, gurps.BlockTraitsKey, gurps.BlockSkillsKey, gurps.BlockSpellsKey,
		gurps.BlockNotesKey)
	traits := editor.ensureRegions().leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	dragBlockTo(t, editor, gurps.BlockSkillsKey,
		geom.NewPoint(traits.rect.CenterX(), traits.rect.Bottom()-3*layoutEdgeLadderStep/2))
	c.Equal("column[column[traits skills] spells notes]", layoutTreeOf(sheet), "the first two blocks must be a group")

	regions := editor.ensureRegions()
	skills := regions.leafFor(gurps.BlockSkillsKey)
	c.NotNil(skills, "the skills block must be on the page")
	c.Equal(1, len(skills.ancestors), "the skills block sits inside the group alone")
	group := &regions.containers[skills.ancestors[0]]
	c.True(group.isRootBand, "the group must be a band of the page")

	spells := regions.leafFor(gurps.BlockSpellsKey)
	c.NotNil(spells, "the spells block must be on the page")
	where := geom.NewPoint(skills.rect.CenterX(), skills.rect.Bottom()-layoutEdgeLadderStep/4)
	target := resolveDropTarget(regions, where, gurps.BlockNotesKey)
	c.Equal(dropAsBand, target.kind, "the bottom of the group is the bottom of the last block stacked in it")
	c.Equal(1, target.bandIndex, "the new band goes after the group")
	c.Equal(layoutBandBar(regions.pageRect, (group.rect.Bottom()+spells.rect.Y)/2), target.highlight,
		"the bar sits on the seam between the group and the band below it")

	dragBlockTo(t, editor, gurps.BlockNotesKey, where)
	c.Equal("column[column[traits skills] notes spells]", layoutTreeOf(sheet),
		"the block must become a band of its own after the group rather than joining it")
}

// TestStraddleTwoBandsFromTheSeamBetweenThem verifies that dropping a block on one end of the seam between two of the
// page's bands makes it stand beside both of them, spanning the pair. There is no other gesture that asks for it: the
// edge of either band alone can only ask to stand beside that one band.
func TestStraddleTwoBandsFromTheSeamBetweenThem(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLeafBands(t, gurps.BlockTraitsKey, gurps.BlockSkillsKey, gurps.BlockNotesKey,
		gurps.BlockSpellsKey)
	c.Equal("column[traits skills notes spells]", layoutTreeOf(sheet), "each block must start out as a band of its own")
	regions := editor.ensureRegions()
	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	skills := regions.leafFor(gurps.BlockSkillsKey)
	c.NotNil(skills, "the skills block must be on the page")
	seam := findTestSeam(regions, traits.node, skills.node)
	c.NotNil(seam, "the traits and skills bands must have a seam between them")
	span := traits.rect.Union(skills.rect)
	c.Equal(span, seam.spanRect, "the seam must span both of the bands it lies between")

	where := geom.NewPoint(seam.rect.X+seam.rect.Width/6, seam.rect.CenterY())
	target := resolveDropTarget(regions, where, gurps.BlockNotesKey)
	c.Equal(dropStraddle, target.kind, "the end of a seam asks for the block to span both sides of it")
	c.Equal(layoutedge.Left, target.edge)
	c.True(traits.node == target.first && skills.node == target.second,
		"the pair must be the bands the seam lies between")
	c.Equal(edgeHalf(span, layoutedge.Left), target.highlight, "the left half of the pair together is lit up")
	c.False(target.bar, "a block that spans a pair is shown as an area rather than as a bar")

	counter := installSyncCounter(sheet)
	counter.count = 0
	dragBlockTo(t, editor, gurps.BlockNotesKey, where)
	c.Equal(1, counter.count, "straddling two bands must sync the sheet exactly once")
	c.Equal("column[row[notes column[traits skills]] spells]", layoutTreeOf(sheet),
		"the two bands must have become one group with the block beside it")

	mgr := sheet.UndoManager()
	c.Equal(1, undoEditCount(mgr), "straddling two bands must be one undoable edit")
	mgr.Undo()
	c.Equal("column[traits skills notes spells]", layoutTreeOf(sheet), "undo must take the two bands back apart")
	mgr.Redo()
	c.Equal("column[row[notes column[traits skills]] spells]", layoutTreeOf(sheet),
		"redo must put the block back beside the pair")
}

// TestDropBetweenTwoBandsFromTheSeamBetweenThem verifies that dropping a block on the middle of the seam between two of
// the page's bands makes it a band of its own between them.
func TestDropBetweenTwoBandsFromTheSeamBetweenThem(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLeafBands(t, gurps.BlockTraitsKey, gurps.BlockSkillsKey, gurps.BlockNotesKey,
		gurps.BlockSpellsKey)
	regions := editor.ensureRegions()
	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	skills := regions.leafFor(gurps.BlockSkillsKey)
	c.NotNil(skills, "the skills block must be on the page")
	seam := findTestSeam(regions, traits.node, skills.node)
	c.NotNil(seam, "the traits and skills bands must have a seam between them")

	where := geom.NewPoint(seam.rect.CenterX(), seam.rect.CenterY())
	target := resolveDropTarget(regions, where, gurps.BlockSpellsKey)
	c.Equal(dropAsBand, target.kind, "the middle of a seam of the page asks for a band between the two of them")
	c.Equal(1, target.bandIndex, "the new band takes the place of the second of the pair")
	c.Equal(layoutBandBar(regions.pageRect, seam.rect.CenterY()), target.highlight,
		"a new band is shown as a bar across the page")

	counter := installSyncCounter(sheet)
	counter.count = 0
	dragBlockTo(t, editor, gurps.BlockSpellsKey, where)
	c.Equal(1, counter.count, "coming between two bands must sync the sheet exactly once")
	c.Equal("column[traits spells skills notes]", layoutTreeOf(sheet),
		"the block must become a plain band between the two of them")

	mgr := sheet.UndoManager()
	c.Equal(1, undoEditCount(mgr), "coming between two bands must be one undoable edit")
	mgr.Undo()
	c.Equal("column[traits skills notes spells]", layoutTreeOf(sheet), "undo must put the block back where it was")
	mgr.Redo()
	c.Equal("column[traits spells skills notes]", layoutTreeOf(sheet), "redo must put the block back between them")
}

// TestStraddleTwoBlocksSideBySideFromTheSeamBetweenThem verifies that dropping a block on one end of the seam between
// two blocks that sit side by side puts it above or below the pair, spanning the width of both.
func TestStraddleTwoBlocksSideBySideFromTheSeamBetweenThem(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t,
		testContainerNode(layoutnode.Row, fxp.One, testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One)),
		testBlockNode(gurps.BlockNotesKey, fxp.One))
	c.Equal("column[row[traits skills] notes]", layoutTreeOf(sheet), "the first band must be a row of two blocks")
	regions := editor.ensureRegions()
	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	skills := regions.leafFor(gurps.BlockSkillsKey)
	c.NotNil(skills, "the skills block must be on the page")
	seam := findTestSeam(regions, traits.node, skills.node)
	c.NotNil(seam, "the traits and skills blocks must have a seam between them")
	c.True(seam.vertical, "two blocks side by side have a seam that runs up and down")
	span := traits.rect.Union(skills.rect)

	where := geom.NewPoint(seam.rect.CenterX(), seam.rect.Y+seam.rect.Height/6)
	target := resolveDropTarget(regions, where, gurps.BlockNotesKey)
	c.Equal(dropStraddle, target.kind, "the end of a seam asks for the block to span both sides of it")
	c.Equal(layoutedge.Top, target.edge)
	c.Equal(edgeHalf(span, layoutedge.Top), target.highlight, "the top half of the pair together is lit up")

	counter := installSyncCounter(sheet)
	counter.count = 0
	dragBlockTo(t, editor, gurps.BlockNotesKey, where)
	c.Equal(1, counter.count, "straddling two blocks must sync the sheet exactly once")
	c.Equal("column[column[notes row[traits skills]]]", layoutTreeOf(sheet),
		"the block must span the width of both, which the band becomes a group to hold")

	mgr := sheet.UndoManager()
	c.Equal(1, undoEditCount(mgr), "straddling two blocks must be one undoable edit")
	mgr.Undo()
	c.Equal("column[row[traits skills] notes]", layoutTreeOf(sheet), "undo must put the block back where it was")
	mgr.Redo()
	c.Equal("column[column[notes row[traits skills]]]", layoutTreeOf(sheet), "redo must put the block back above them")
}

// TestDropBetweenTwoBlocksSideBySideFromTheSeamBetweenThem verifies that dropping a block on the middle of the seam
// between two blocks that sit side by side puts it into their row between the two of them.
func TestDropBetweenTwoBlocksSideBySideFromTheSeamBetweenThem(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t,
		testContainerNode(layoutnode.Row, fxp.One, testBlockNode(gurps.BlockTraitsKey, fxp.One),
			testBlockNode(gurps.BlockSkillsKey, fxp.One)),
		testBlockNode(gurps.BlockNotesKey, fxp.One))
	regions := editor.ensureRegions()
	traits := regions.leafFor(gurps.BlockTraitsKey)
	c.NotNil(traits, "the traits block must be on the page")
	skills := regions.leafFor(gurps.BlockSkillsKey)
	c.NotNil(skills, "the skills block must be on the page")
	seam := findTestSeam(regions, traits.node, skills.node)
	c.NotNil(seam, "the traits and skills blocks must have a seam between them")

	where := geom.NewPoint(seam.rect.CenterX(), seam.rect.CenterY())
	target := resolveDropTarget(regions, where, gurps.BlockNotesKey)
	c.Equal(dropOnEdge, target.kind, "the middle of a seam asks for the block to come between the two of them")
	c.Equal(layoutedge.Right, target.edge)
	c.True(traits.node == target.node, "coming between two things side by side is joining their row after the first")
	c.True(target.bar, "coming between two things is shown as a bar rather than as an area")
	c.Equal(layoutSeamBar(seam), target.highlight, "the bar sits on the seam and reaches no further than the pair")

	counter := installSyncCounter(sheet)
	counter.count = 0
	dragBlockTo(t, editor, gurps.BlockNotesKey, where)
	c.Equal(1, counter.count, "coming between two blocks must sync the sheet exactly once")
	c.Equal("column[row[traits notes skills]]", layoutTreeOf(sheet), "the block must have joined the row between them")

	mgr := sheet.UndoManager()
	c.Equal(1, undoEditCount(mgr), "coming between two blocks must be one undoable edit")
	mgr.Undo()
	c.Equal("column[row[traits skills] notes]", layoutTreeOf(sheet), "undo must put the block back where it was")
	mgr.Redo()
	c.Equal("column[row[traits notes skills]]", layoutTreeOf(sheet), "redo must put the block back between them")
}

// portraitContentSize returns the size of the portrait's picture area, which is the block's content rect: its frame
// less the insets of its titled border.
func portraitContentSize(t *testing.T, editor *sheetLayoutEditor) geom.Size {
	t.Helper()
	leaf := editor.ensureRegions().leafFor(gurps.BlockPortraitKey)
	if leaf == nil {
		t.Fatal("the portrait block must be on the page")
	}
	return leaf.panel.ContentRect(false).Size
}

// nearlySquare returns true if the given picture area is square to within a page pixel, which is as near as the
// rounding of a weight into the fixed-point number the model keeps allows.
func nearlySquare(size geom.Size) bool {
	return xmath.Abs(size.Width-size.Height) < 1
}

// TestSquarePortraitNarrowsTheWiderBlock verifies that a portrait that is far wider than it is tall is squared by
// taking width away from it, which is much the smaller change of the two, rather than by making its whole band as tall
// as the portrait is wide.
func TestSquarePortraitNarrowsTheWiderBlock(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t,
		testContainerNode(layoutnode.Row, fxp.One, testBlockNode(gurps.BlockPortraitKey, fxp.Three),
			testBlockNode(gurps.BlockIdentityKey, fxp.One)))
	counter := installSyncCounter(sheet)
	mgr := sheet.UndoManager()
	before := portraitContentSize(t, editor)
	c.True(before.Width > before.Height+1, "the portrait must start out wider than it is tall, but was %v", before)
	node, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.NotNil(node, "the portrait must be in the tree")
	minHeight := node.MinHeight

	counter.count = 0
	editor.squarePortrait()
	syncs := counter.count

	after := portraitContentSize(t, editor)
	c.True(nearlySquare(after), "the picture area must have been made square, but was %v", after)
	c.Equal(1, syncs, "squaring the portrait must sync the sheet exactly once")
	node, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.True(node.Weight > 0 && node.Weight < fxp.Three, "the block's share of its row must have been narrowed")
	c.Equal(minHeight, node.MinHeight, "the block's minimum height must have been left alone")
	c.True(xmath.Abs(after.Height-before.Height) < 1, "the height must have been left alone, but went from %v to %v",
		before.Height, after.Height)
	c.Equal(1, undoEditCount(mgr), "squaring the portrait must be one undoable edit")

	mgr.Undo()
	restored := portraitContentSize(t, editor)
	c.True(xmath.Abs(restored.Width-before.Width) < 0.01 && xmath.Abs(restored.Height-before.Height) < 0.01,
		"undo must give the portrait the shape it had, but %v became %v", before, restored)
	mgr.Redo()
	c.True(nearlySquare(portraitContentSize(t, editor)), "redo must make the picture area square again")
}

// TestSquarePortraitWidensAgainstATallerNeighbor verifies that a portrait kept tall by what stands beside it is squared
// by widening it. Shortening it is not on offer at all: the row is as tall as the column beside it, so the minimum
// height that would square the block is swallowed and the block is left exactly as it was.
func TestSquarePortraitWidensAgainstATallerNeighbor(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t,
		testContainerNode(layoutnode.Row, fxp.One, testBlockNode(gurps.BlockPortraitKey, fxp.One),
			testContainerNode(layoutnode.Column, fxp.Four, testBlockNode(gurps.BlockIdentityKey, fxp.One),
				testBlockNode(gurps.BlockDescriptionKey, fxp.One), testBlockNode(gurps.BlockPointsKey, fxp.One))))
	counter := installSyncCounter(sheet)
	before := portraitContentSize(t, editor)
	c.True(before.Height > before.Width+1, "the portrait must start out taller than it is wide, but was %v", before)

	counter.count = 0
	editor.squarePortrait()
	syncs := counter.count

	after := portraitContentSize(t, editor)
	c.True(nearlySquare(after), "the picture area must have been made square, but was %v", after)
	c.Equal(1, syncs, "squaring the portrait must sync the sheet exactly once")
	c.True(after.Width > before.Width, "the block must have been widened, but went from %v to %v", before.Width,
		after.Width)
	node, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.True(node.Weight > fxp.One, "the block's share of its row must have been widened")
	c.Equal(1, undoEditCount(sheet.UndoManager()), "squaring the portrait must be one undoable edit")
}

// TestSquarePortraitAloneInABandSetsItsHeight verifies that a portrait that has the page to itself, which leaves
// nothing to take width from, is squared by giving the block the minimum height that makes its picture area as tall as
// it is wide.
func TestSquarePortraitAloneInABandSetsItsHeight(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t, testBlockNode(gurps.BlockPortraitKey, fxp.One),
		testBlockNode(gurps.BlockNotesKey, fxp.One))
	counter := installSyncCounter(sheet)
	before := portraitContentSize(t, editor)
	c.True(before.Width > before.Height+1, "the portrait must start out wider than it is tall, but was %v", before)
	leaf := editor.ensureRegions().leafFor(gurps.BlockPortraitKey)
	c.NotNil(leaf, "the portrait block must be on the page")
	insets := leaf.panel.Border().Insets()

	counter.count = 0
	editor.squarePortrait()
	syncs := counter.count

	after := portraitContentSize(t, editor)
	c.True(nearlySquare(after), "the picture area must have been made square, but was %v", after)
	c.Equal(1, syncs, "squaring the portrait must sync the sheet exactly once")
	c.True(xmath.Abs(after.Width-before.Width) < 0.01, "the width must have been left alone, since nothing shares the "+
		"band with the portrait")
	node, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.True(xmath.Abs(node.MinHeight.Pixels()-(before.Width+insets.Height())) < 1,
		"the minimum height must be the picture area's width plus the border, but was %v", node.MinHeight.Pixels())
	c.Equal(1, undoEditCount(sheet.UndoManager()), "squaring the portrait must be one undoable edit")
}

// TestSquarePortraitThatIsAlreadySquareDoesNothing verifies that asking for a square a second time neither records an
// edit nor rebuilds the sheet.
func TestSquarePortraitThatIsAlreadySquareDoesNothing(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t, testBlockNode(gurps.BlockPortraitKey, fxp.One),
		testBlockNode(gurps.BlockNotesKey, fxp.One))
	counter := installSyncCounter(sheet)
	editor.squarePortrait()
	c.True(nearlySquare(portraitContentSize(t, editor)), "the picture area must have been made square")
	before := gurps.Hash64(sheet.Entity().SheetSettings.Layout)

	counter.count = 0
	editor.squarePortrait()
	syncs := counter.count

	c.Equal(0, syncs, "squaring a portrait that is already square must not sync the sheet")
	c.Equal(before, gurps.Hash64(sheet.Entity().SheetSettings.Layout), "the layout must have been left alone")
	c.Equal(1, undoEditCount(sheet.UndoManager()), "squaring a portrait that is already square must not be recorded")
}

// TestDividerDragBesideTheSquarePortraitTakesOverTheWidth verifies that dragging the divider beside the portrait of a
// sheet laid out the factory way takes its width out of the height's hands, that nothing jumps as that happens, and
// that undoing the drag gives the portrait its square back.
func TestDividerDragBesideTheSquarePortraitTakesOverTheWidth(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	mgr := sheet.UndoManager()
	layout := sheet.Entity().SheetSettings.Layout
	node, _, _ := layout.Find(gurps.BlockPortraitKey)
	c.NotNil(node, "the portrait must be in the tree")
	c.True(node.Square, "the factory portrait takes its width from the height of the band it is in")
	regions := editor.ensureRegions()
	divider := findTestDivider(regions, gurps.BlockPortraitKey)
	c.NotNil(divider, "the portrait must have a divider beside it")
	before := portraitContentSize(t, editor)
	c.True(nearlySquare(before), "the picture area must start out square, but was %v", before)
	start := geom.NewPoint(divider.rect.CenterX(), divider.rect.CenterY())

	editor.beginDividerDrag(divider, start)
	c.False(node.Square, "a divider drag beside the portrait must hand its width over to the weights")
	atStart := portraitContentSize(t, editor)
	c.True(xmath.Abs(atStart.Width-before.Width) < 1,
		"the width must not jump as the drag takes it over, but %v became %v", before.Width, atStart.Width)

	editor.endDividerDrag(geom.NewPoint(start.X+40, start.Y))
	after := portraitContentSize(t, editor)
	c.True(xmath.Abs(after.Width-(before.Width+40)) < 2,
		"the block must follow the divider, but %v became %v", before.Width, after.Width)
	node, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.False(node.Square, "the width must still be the user's to set")
	c.Equal(1, undoEditCount(mgr), "the drag must be one undoable edit")

	mgr.Undo()
	node, _, _ = sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.True(node.Square, "undo must give the portrait its square back")
	restored := portraitContentSize(t, editor)
	c.True(xmath.Abs(restored.Width-before.Width) < 0.01 && nearlySquare(restored),
		"undo must give the picture area the shape it had, but %v became %v", before, restored)
}

// TestBottomEdgeDragKeepsTheSquarePortraitSquare verifies that dragging the portrait's bottom edge grows its picture
// area without taking it off its square: the minimum height raises the band, and the width follows the height.
func TestBottomEdgeDragKeepsTheSquarePortraitSquare(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetForLayoutEditing(t)
	leaf := editor.ensureRegions().leafFor(gurps.BlockPortraitKey)
	c.NotNil(leaf, "the portrait block must be on the page")
	before := portraitContentSize(t, editor)
	c.True(nearlySquare(before), "the picture area must start out square, but was %v", before)
	// A copy is taken, since the regions are thrown away the moment the drag moves anything.
	region := *leaf
	where := geom.NewPoint(region.rect.CenterX(), region.rect.Bottom()+60)

	editor.beginBottomDrag(&region, geom.NewPoint(region.rect.CenterX(), region.rect.Bottom()))
	editor.endBottomDrag(where)

	after := portraitContentSize(t, editor)
	c.True(nearlySquare(after), "the picture area must still be square, but was %v", after)
	c.True(after.Height > before.Height+50, "the block must be taller, but %v became %v", before.Height, after.Height)
	c.True(after.Width > before.Width+50, "and just as much wider, but %v became %v", before.Width, after.Width)
	node, _, _ := sheet.Entity().SheetSettings.Layout.Find(gurps.BlockPortraitKey)
	c.True(node.Square, "the bottom edge must leave the width in the height's hands")
	c.True(node.MinHeight.Pixels() > 0, "the drag must have set a minimum height")
}

// TestSquarePortraitButton verifies that only the portrait carries the square button, that it sits just to the left of
// the close button, and that pressing and releasing on it squares the picture area.
func TestSquarePortraitButton(t *testing.T) {
	c := check.New(t)
	sheet, editor := newTestSheetWithLayout(t,
		testContainerNode(layoutnode.Row, fxp.One, testBlockNode(gurps.BlockPortraitKey, fxp.Three),
			testBlockNode(gurps.BlockIdentityKey, fxp.One)))
	counter := installSyncCounter(sheet)
	regions := editor.ensureRegions()
	portrait := regions.leafFor(gurps.BlockPortraitKey)
	c.NotNil(portrait, "the portrait block must be on the page")
	for i := range regions.leaves {
		leaf := &regions.leaves[i]
		c.Equal(leaf.key == gurps.BlockPortraitKey, !leaf.squareRect.Empty(),
			"only the portrait may have a square button, but %q %v", leaf.key, leaf.squareRect)
	}
	c.Equal(portrait.closeRect.Size, portrait.squareRect.Size, "the two buttons must be the same size")
	c.Equal(portrait.closeRect.Y, portrait.squareRect.Y, "the two buttons must sit in the same strip")
	c.Equal(portrait.closeRect.X-layoutButtonSize-layoutButtonGap, portrait.squareRect.X,
		"the square button must sit just to the left of the close button")

	where := portrait.squareRect.Center()
	c.Nil(editor.closeAt(where), "the square button must not overlap the close button")
	c.True(editor.squareAt(where) != nil, "the square button must be found where it is drawn")
	c.Equal(unison.PointingCursor(), editor.cursorAt(where), "the pointer must say the button can be pressed")

	counter.count = 0
	editor.mouseDown(where, unison.ButtonLeft)
	c.Equal(gurps.BlockPortraitKey, editor.squareKey, "the press must be on the portrait's square button")
	editor.mouseUp(where, unison.ButtonLeft)
	syncs := counter.count

	c.Equal(1, syncs, "pressing the square button must sync the sheet exactly once")
	c.Equal("", editor.squareKey, "the press must have been let go of")
	c.Equal(layoutIdle, editor.mode, "the gesture must be over")
	c.True(nearlySquare(portraitContentSize(t, editor)), "pressing the square button must square the picture area")
	c.Equal(1, undoEditCount(sheet.UndoManager()), "pressing the square button must be one undoable edit")
}
