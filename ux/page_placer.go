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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// pageKey is the client data key under which a leaf function records which block a panel shows (see pageKeyOf).
const pageKey = "pageKey"

// pageSpacerKey is the client data key that marks a panel made by newPageSpacer.
const pageSpacerKey = "pageSpacer"

// maxBandPlacementPasses caps the number of pages a single band may spread over. Each pass places at least one row or
// block, so the cap only keeps a band that somehow fails to make progress from paginating forever.
const maxBandPlacementPasses = 1000

// placedBlockMarker is what the startAt map holds for a block that isn't a list once it has been placed. Such a block
// is placed whole or not at all, so there is no row to pick up from, only the fact that it has had its say. A list with
// no rows has nothing to pick up from either, and is given the marker too once placed.
const placedBlockMarker = 1

// paginateBands fills pages with the bands of the given layout root, in order, asking leaf for each page's view of a
// block; newPage hands back the next page to fill (the exporter creates one, the sheet reuses its own). Returns the
// pages it filled, in order.
//
// The first page is asked for up front, since it may hold content of its own and there is never less than one page;
// each later one only once there is something to put on it, so a band ending exactly at the bottom of a page can't
// leave a blank one. Each band is placed on its own, one page at a time: a band that doesn't fit in what is left of its
// page is built again for the next page, with leaf picking each of its blocks up where the band's startAt map says it
// left off, and a band that builds to nothing, or to nothing but a page spacer, has been placed in full. Every page is
// laid out at pageSize, and forced to that full size once all of the bands are placed.
func paginateBands(root *gurps.SheetLayoutNode, leaf func(startAt map[string]int) layoutLeafFunc,
	newPage func() *Page, pageSize geom.Size,
) []*Page {
	r := geom.Rect{Size: pageSize}
	page := newPage()
	pages := []*Page{page}
	nextPage := func() {
		page = newPage()
		pages = append(pages, page)
	}
	var bands []*gurps.SheetLayoutNode
	if root != nil {
		bands = root.Children
	}
	needNewPage := false
	for i, node := range bands {
		startAt := make(map[string]int)
		finished := false
		for range maxBandPlacementPasses {
			built := buildLayoutNode(node, leaf(startAt))
			if xreflect.IsNil(built) || isPageSpacer(built.AsPanel()) {
				finished = true // Nothing is left of the band that hasn't already been placed
				break
			}
			if needNewPage {
				nextPage()
				needNewPage = false
			}
			band := built.AsPanel()
			page.AddChild(band)
			layoutPage(page, r)
			// A band that has the page to itself is allowed to overflow it, since moving it on wouldn't give it any
			// more room than it already has and it would never be placed at all.
			alone := len(page.Children()) == 1
			placer := newBandPlacer(startAt)
			used, done := placer.placeBand(band, availableHeightFor(page, band, pageSize), alone)
			if used == 0 && !done && !alone {
				// Nothing at all fit in what was left of this page, so the band starts over on a page of its own.
				page.RemoveChild(band)
				nextPage()
				page.AddChild(band)
				layoutPage(page, r)
				placer = newBandPlacer(startAt)
				_, done = placer.placeBand(band, availableHeightFor(page, band, pageSize), true)
			}
			if !placer.applyBand(band) {
				page.RemoveChild(band)
			}
			layoutPage(page, r)
			if done {
				finished = true
				break
			}
			needNewPage = true
		}
		if !finished {
			errs.Log(errs.New("gave up trying to fit a block layout band onto the pages"), "band", i)
		}
	}
	for _, one := range pages {
		one.Force = true
		layoutPage(one, r)
	}
	return pages
}

// splittableList is what filling pages asks of a list, whatever the type of its rows (see PageList).
type splittableList interface {
	pagePanel
	RowCount() int
	continuationPanel(n int) pagePanel
	trimContinuations(used int)
}

// pageListLeaf returns the panel to show the rows of the list from start on, as a leaf function filling pages hands it
// out (see paginateBands): the list itself for its first rows, the next of its continuations for the rows of each
// later page, counted under the block's key in handedOut, and a page spacer once every row has been placed. The panel
// is taken out of wherever it was, marked with the block's key under pageKey and given its rows. A list with no rows
// gives nothing when omitEmpty is set: the exporter leaves such a list off the printed pages, while the sheet shows it
// unless it is one derived from the character, so that there is somewhere to add to.
func pageListLeaf(key string, list splittableList, start int, handedOut map[string]int, omitEmpty bool) unison.Paneler {
	count := list.RowCount()
	if count == 0 && omitEmpty {
		return nil
	}
	var panel pagePanel
	switch {
	case start == 0:
		handedOut[key] = 0
		panel = list
	case start >= count:
		return newPageSpacer()
	default:
		panel = list.continuationPanel(handedOut[key])
		handedOut[key]++
	}
	panel.AsPanel().RemoveFromParent()
	panel.AsPanel().ClientData()[pageKey] = key
	panel.SetDrawRowRange(start, count)
	return panel
}

// newPageSpacer returns a panel that stands in for a block placed on an earlier page, keeping the block's slot in a Row
// empty rather than giving it up, so that a list continuing beside it keeps its width and column widths. It draws
// nothing, asks for no space and carries no pageKey, so pageKeyOf reports none for it.
func newPageSpacer() *unison.Panel {
	p := unison.NewPanel()
	p.ClientData()[pageSpacerKey] = true
	return p
}

// isPageSpacer returns true if the panel was made by newPageSpacer.
func isPageSpacer(p *unison.Panel) bool {
	_, ok := p.ClientData()[pageSpacerKey]
	return ok
}

// availableHeightFor returns the height still available on the page for the given band, i.e. everything from the top of
// the band down to the bottom of the page's printable area.
func availableHeightFor(page *Page, band *unison.Panel, pageSize geom.Size) float32 {
	return (pageSize.Height - page.insets().Bottom) - band.FrameRect().Y
}

// layoutPage puts the page at the given rect and lays out everything on it.
func layoutPage(page *Page, r geom.Rect) {
	page.SetFrameRect(r)
	page.MarkForLayoutRecursively()
	page.ValidateLayout()
}

// pageKeyOf returns the block key the leaf function stamped on the panel, or an empty string if it has none.
func pageKeyOf(panel *unison.Panel) string {
	if key, ok := panel.ClientData()[pageKey].(string); ok {
		return key
	}
	return ""
}

// panelMinHeight returns the minimum height the builder recorded for the slot the panel occupies.
func panelMinHeight(panel *unison.Panel) float32 {
	if data, ok := panel.LayoutData().(*unison.FlexLayoutData); ok {
		return data.MinSize.Height
	}
	return 0
}

type pageHelper interface {
	OverheadHeight() float32
	// MinimumHeight returns the height the list stands at however few rows it draws, none included.
	MinimumHeight() float32
	RowHeights() []float32
	CurrentDrawRowRange() (start, endBefore int)
	SetDrawRowRange(start, endBefore int)
	ClearDrawRowRange()
}

// bandPanelKind is the kind of thing one of the panels of a band is, as far as fitting it onto a page is concerned.
type bandPanelKind byte

const (
	// atomicBandPanel is a block that is placed whole or not at all.
	atomicBandPanel bandPanelKind = iota
	// listBandPanel is a block whose rows may be split across a page boundary.
	listBandPanel
	// rowBandPanel is a container whose children stand side by side.
	rowBandPanel
	// columnBandPanel is a container whose children are stacked.
	columnBandPanel
)

// bandKindOf returns what kind of thing the given panel of a band is. It inspects the panel rather than the node under
// sheetLayoutNodeKey, since a child that took a dropped single-child container's place records that container's node.
func bandKindOf(panel *unison.Panel) bandPanelKind {
	if _, ok := panel.Self.(pageHelper); ok {
		return listBandPanel
	}
	if _, ok := panel.Layout().(*weightedRowLayout); ok {
		return rowBandPanel
	}
	if node := sheetLayoutContainerNodeOf(panel); node != nil && node.Type == layoutnode.Column {
		return columnBandPanel
	}
	return atomicBandPanel
}

// bandLeafPlacement records what came of trying to place one of the blocks of a band.
type bandLeafPlacement struct {
	helper        pageHelper // nil unless the block is a list
	start         int        // The first row of a list to be drawn
	firstUnplaced int        // The first row of a list that didn't fit
	placed        bool       // Whether anything at all was placed
}

// bandPlacer works out how much of a band fits in the space a page has left for it and then carries that decision out.
// One placer is used for one attempt at placing one band on one page.
type bandPlacer struct {
	startAt  map[string]int
	results  map[*unison.Panel]*bandLeafPlacement
	vSpacing float32
}

// newBandPlacer creates a placer that picks each block of the band up from the row the given map says it left off at,
// and that records where it leaves off in turn.
func newBandPlacer(startAt map[string]int) *bandPlacer {
	return &bandPlacer{
		startAt: startAt,
		results: make(map[*unison.Panel]*bandLeafPlacement),
		// The space a column of blocks leaves between them, matching the layout the builder gives a Column node.
		vSpacing: 1,
	}
}

// placeBand works out how much of the given panel of a band fits in avail, the height the page has left for it,
// returning the height that takes and whether everything in it was placed. Nothing is altered: the decisions are
// recorded and applyBand carries them out.
//
// A list gives up as many of its rows as fit and leaves the rest for the next page; anything else is placed whole or
// left for the next page. mayOverflow means the panel has the page to itself and must place something: a list then
// takes at least its first row and any other block is placed anyway, losing whatever runs past the bottom of the page.
//
// A minimum height is a floor the panel stands at however little of it is placed, so a panel whose floor is deeper than
// avail is left whole for the next page, unless mayOverflow is set, since no page has more room to offer it. A page
// spacer is the exception: it holds an empty slot, so it is always placed, taking no height at all.
func (b *bandPlacer) placeBand(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	if isPageSpacer(panel) {
		b.results[panel] = &bandLeafPlacement{placed: true}
		return 0, true
	}
	minHeight := panelMinHeight(panel)
	if minHeight > avail && !mayOverflow {
		b.deferSubtree([]*unison.Panel{panel})
		return 0, false
	}
	switch bandKindOf(panel) {
	case listBandPanel:
		used, done = b.placeList(panel, avail, mayOverflow)
	case rowBandPanel:
		used, done = b.placeRow(panel, avail, mayOverflow)
	case columnBandPanel:
		used, done = b.placeColumn(panel, avail, mayOverflow)
	default:
		used, done = b.placeAtomic(panel, avail, mayOverflow)
	}
	if used > 0 {
		used = max(used, minHeight)
	}
	return used, done
}

// placeList works out how many of the rows of a list fit. The list stands at its minimum height however few of them
// that is, so one whose minimum doesn't fit is left whole for the next page, rows or none, unless mayOverflow is set.
func (b *bandPlacer) placeList(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	helper, ok := panel.Self.(pageHelper)
	if !ok {
		return 0, true
	}
	start, endBefore := helper.CurrentDrawRowRange()
	heights := helper.RowHeights()
	endBefore = min(endBefore, len(heights))
	result := &bandLeafPlacement{
		helper:        helper,
		start:         start,
		firstUnplaced: endBefore,
		placed:        true,
	}
	b.results[panel] = result
	floor := helper.MinimumHeight()
	if floor > avail && !mayOverflow {
		result.placed = false
		result.firstUnplaced = start
		return 0, false
	}
	used = helper.OverheadHeight()
	remaining := avail - used
	for i := start; i < endBefore; i++ {
		height := heights[i]
		if height > remaining {
			if i != start {
				result.firstUnplaced = i
				return max(used, floor), false
			}
			if !mayOverflow {
				result.placed = false
				result.firstUnplaced = start
				return 0, false
			}
			// A row too tall to fit on a page of its own is placed anyway and allowed to run off the end of the page,
			// dropping whatever doesn't fit, since there is nowhere else for it to go.
			//
			// TODO: In the future, see if I can do sub-row partitioning.
			result.firstUnplaced = start + 1
			return max(used+height, floor), start+1 >= endBefore
		}
		remaining -= height
		used += height
	}
	return max(used, floor), true
}

// placeAtomic works out whether a block that can't be split fits.
func (b *bandPlacer) placeAtomic(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	result := &bandLeafPlacement{placed: true}
	b.results[panel] = result
	_, pref, _ := panel.Sizes(geom.Size{Width: panel.FrameRect().Width})
	height := max(pref.Height, panelMinHeight(panel))
	if height <= avail {
		return height, true
	}
	if mayOverflow {
		return avail, true
	}
	result.placed = false
	return 0, false
}

// placeColumn works out how much of a stack of blocks fits, from the top down.
func (b *bandPlacer) placeColumn(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	children := panel.Children()
	remaining := avail
	placedAny := false
	for i, child := range children {
		if i != 0 {
			used += b.vSpacing
			remaining -= b.vSpacing
		}
		// Only the first thing to be placed may overflow the page, since anything below it could still be moved on to
		// a page where there is more room for it.
		childUsed, childDone := b.placeBand(child, remaining, mayOverflow && !placedAny)
		used += childUsed
		remaining -= childUsed
		if childUsed > 0 {
			placedAny = true
		}
		if !childDone {
			// Nothing below a block that isn't finished may be placed, or the page after this one would have the
			// blocks of the column out of order.
			b.deferSubtree(children[i+1:])
			return used, false
		}
	}
	return used, true
}

// placeRow works out how much of a set of blocks standing side by side fits. They all start at the same height, so
// they are each given the same space, and if any of them can't take even a part of what it holds, the whole row is
// left for the next page rather than leaving a hole in the middle of it. A page spacer holds nothing and is always
// finished, so it never holds the row back.
func (b *bandPlacer) placeRow(panel *unison.Panel, avail float32, mayOverflow bool) (used float32, done bool) {
	children := panel.Children()
	done = true
	deferred := false
	for _, child := range children {
		childUsed, childDone := b.placeBand(child, avail, mayOverflow)
		used = max(used, childUsed)
		if !childDone {
			done = false
			if childUsed == 0 {
				deferred = true
			}
		}
	}
	if deferred {
		b.deferSubtree(children)
		return 0, false
	}
	return used, done
}

// deferSubtree marks every block within the given panels as having been given nothing, taking back anything that was
// placed within them, so that applyBand leaves them for the next page.
func (b *bandPlacer) deferSubtree(panels []*unison.Panel) {
	for _, panel := range panels {
		switch bandKindOf(panel) {
		case rowBandPanel, columnBandPanel:
			b.deferSubtree(panel.Children())
		default:
			b.results[panel] = &bandLeafPlacement{}
		}
	}
}

// applyBand carries out what placeBand decided: a block that was given nothing is taken off the page, as is any
// container that leaves empty; a list that only partly fit is told to draw just the rows that did; and the point each
// block is to be picked up from on the next page is recorded. A page spacer stays, holding its slot, unless the row it
// is in was left for the next page. Returns false if nothing of the panel is left.
func (b *bandPlacer) applyBand(panel *unison.Panel) bool {
	switch bandKindOf(panel) {
	case rowBandPanel, columnBandPanel:
		for _, child := range slices.Clone(panel.Children()) {
			if !b.applyBand(child) {
				panel.RemoveChild(child)
			}
		}
		return len(panel.Children()) != 0
	default:
		result := b.results[panel]
		if result == nil || !result.placed {
			return false
		}
		key := pageKeyOf(panel)
		if result.helper == nil {
			if key != "" {
				b.startAt[key] = placedBlockMarker
			}
			return true
		}
		result.helper.SetDrawRowRange(result.start, result.firstUnplaced)
		if key != "" {
			// A list that placed a row moves past it, so only one with no rows at all needs the marker to be seen as
			// placed rather than handed out again for the next page.
			b.startAt[key] = max(result.firstUnplaced, placedBlockMarker)
		}
		return true
	}
}
