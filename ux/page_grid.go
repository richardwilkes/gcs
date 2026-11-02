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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

var _ unison.Layout = &pageGrid{}

// pageGridGap is the space between the pages of a sheet shown as pages, through which the scroll panel's background
// shows.
const pageGridGap = float32(8)

// maxPaginationPasses caps how many times in a row the pages are filled while laying out what was just put on them
// keeps changing it, as a table sizing its columns for its new width can (see sizePageTableColumnsForFrame).
const maxPaginationPasses = 3

// pageGeometry is what filling a page depends on besides what is put on it.
type pageGeometry struct {
	size   geom.Size
	insets geom.Insets
}

// pageGrid is the content of a character sheet's scroll panel. Shown continuously, which a sheet always is while its
// block layout is being edited, it holds one page at its preferred size. Shown as pages, it holds as many pages of the
// paper's size as the content takes, filled band by band as the exporter fills them (see paginateBands) but with the
// sheet's own panels, an empty list that can be added to among them (see pageListLeaf) and a list going on past its
// page as a view of its table (see PageList.continuation), and sets them out in rows of as many as fit the scroll
// panel's view at the current scale. Any other child, such as the layout editor's overlay, is laid over the whole grid.
//
// The pages are filled from within the grid's own layout, so whatever changes the sheet -- a table synced, a row
// disclosed, a block that grew, a setting -- needs only to mark the layout, as it always has. Most layout passes only
// set the pages out again, for a resize of the window or a change of scale, so the pages are filled again only once
// one of these says their content changed since they were last filled:
//   - invalidate, which the sheet's rebuild and its change of mode call;
//   - a band's frame changing other than while the grid fills or sets out its pages, caught by the grid's
//     FrameChangeInChildHierarchyCallback. A page sets each band's frame once per layout, from the band's size, so it
//     changes only when something in the band grew or shrank enough to change the band, or a band above it did. Frames
//     deeper down are no guide: a titled block sets its heading's frame twice in every layout (see blockLayout);
//   - a list whose number of rows differs from what it had then, since a view keeps the range it was given (see
//     unison.Table.SetDrawRowRange), so a row added at the end of a list may change no frame at all;
//   - a change to the size or margins of the paper.
//
// While the grid needs layout, both LayoutSizes and PerformLayout first lay out any page that asked for it, which is
// what brings a frame change to light, and fill the pages if one of the above holds: in LayoutSizes when that is where
// the need is first seen, so that the number of pages feeds the size the scroll panel gives the grid, and otherwise in
// PerformLayout. Laying the pages out as they are filled can have a table size its columns anew and mark the layout
// once more, so they are laid out and checked again after each filling, up to maxPaginationPasses times, before what
// they hold is accepted.
type pageGrid struct {
	unison.Panel
	sheet *Sheet
	pages []*Page // The first is always the same page, which the sheet knows as its page
	// handedOut holds, by block key, how many of the list's continuations the filling under way has handed out.
	handedOut map[string]int
	// rowCounts holds, by block key, how many rows the list had when the pages were last filled.
	rowCounts map[string]int
	filledFor pageGeometry // The geometry of the pages when they were last filled
	// invalid is set when the pages are to be filled again whatever else says (see invalidate).
	invalid bool
	// contentChanged is set when a band's frame changed other than by the grid's own doing (see frameChangedBelow).
	contentChanged bool
	// arranging is set while the grid fills or sets out its pages, since the frame changes that makes are its own.
	arranging bool
	// settling is set while the grid brings its pages up to date (see settle).
	settling bool
}

// newPageGrid creates the page grid for the sheet, holding the page that is always its first.
func newPageGrid(s *Sheet) *pageGrid {
	g := &pageGrid{
		sheet:     s,
		handedOut: make(map[string]int),
		rowCounts: make(map[string]int),
		invalid:   true,
	}
	g.Self = g
	g.SetLayout(g)
	g.FrameChangeInChildHierarchyCallback = g.frameChangedBelow
	page := NewPage(s.entity)
	g.pages = []*Page{page}
	g.AddChild(page)
	return g
}

// paginated returns true if the grid shows pages rather than one continuous page.
func (g *pageGrid) paginated() bool {
	return g.sheet.paginated && !g.sheet.layoutEditing()
}

// invalidate has the pages filled again the next time the grid is laid out, or at once by relayout.
func (g *pageGrid) invalidate() {
	g.invalid = true
	g.MarkForLayoutRecursivelyUpward()
	g.MarkForRedraw()
}

// relayout lays the grid out at its preferred size at once, rather than at the next layout pass, so that whatever
// scrolls or moves the focus right after a rebuild finds everything where it will be drawn.
func (g *pageGrid) relayout() {
	_, pref, _ := g.Sizes(geom.Size{})
	r := g.FrameRect()
	r.Size = pref
	g.SetFrameRect(r)
	g.MarkForLayoutAndRedraw()
	g.ValidateLayout()
}

// frameChangedBelow is the grid's FrameChangeInChildHierarchyCallback.
func (g *pageGrid) frameChangedBelow(panel *unison.Panel) {
	if g.arranging || g.contentChanged || !g.paginated() {
		return
	}
	// Only a band's frame says anything about what is on the pages (see the type's comment), and every child of the
	// grid is a page while it shows pages.
	if parent := panel.Parent(); parent == nil || parent.Parent() != g.AsPanel() {
		return
	}
	g.contentChanged = true
	if !g.settling {
		// The page was laid out after the grid was, having been marked for layout without the grid, so the grid has to
		// be laid out again for the pages to be filled anew.
		g.MarkForLayoutRecursivelyUpward()
		g.MarkForRedraw()
	}
}

// LayoutSizes implements unison.Layout.
func (g *pageGrid) LayoutSizes(_ *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	if !g.paginated() {
		return g.pages[0].Sizes(hint)
	}
	if g.NeedsLayout || g.needsFilling() {
		g.settle()
	}
	paper := g.pages[0].paperSize()
	prefSize = g.gridSize(pagesAcross(g.viewportWidth(), paper.Width), paper)
	return prefSize, prefSize, prefSize
}

// PerformLayout implements unison.Layout.
func (g *pageGrid) PerformLayout(_ *unison.Panel) {
	var covered geom.Rect
	if g.paginated() {
		count := len(g.pages)
		g.settle()
		g.placePages()
		if len(g.pages) != count {
			// The pages were filled here rather than when the scroll panel asked for the grid's size, so it has to size
			// the grid again for the pages there are now.
			g.MarkForLayoutRecursivelyUpward()
			g.MarkForRedraw()
		}
		covered = g.ContentRect(false)
	} else {
		page := g.pages[0]
		_, pref, _ := page.Sizes(geom.Size{})
		page.SetFrameRect(geom.Rect{Size: pref})
		covered = page.FrameRect()
	}
	for _, child := range g.Children() {
		if _, ok := child.Self.(*Page); !ok {
			child.SetFrameRect(covered)
		}
	}
}

// viewportWidth returns the width of the scroll panel's view the grid is shown through, in the grid's own units, so
// at the current scale, or 0 when there is none. The scroll panel asks for the grid's size before it resizes the view,
// so the width is taken from the scroll panel itself.
func (g *pageGrid) viewportWidth() float32 {
	sp := unison.Ancestor[*unison.ScrollPanel](g)
	if sp == nil {
		return 0
	}
	width := sp.ContentRect(false).Width
	if border := sp.ContentView().Border(); border != nil {
		width -= border.Insets().Width()
	}
	return width / g.Scale().X
}

// pagesAcross returns how many pages of the given width fit side by side in the width, never fewer than one. A width
// worked out by gridSize for a number of pages yields that number again, give or take the rounding a scale introduces.
func pagesAcross(width, pageWidth float32) int {
	if pageWidth <= 0 {
		return 1
	}
	return max(1, int((width+pageGridGap)/(pageWidth+pageGridGap)+0.001))
}

// gridSize returns the size of the grid with its pages set out in rows of at most across.
func (g *pageGrid) gridSize(across int, paper geom.Size) geom.Size {
	across = max(min(across, len(g.pages)), 1)
	rows := (len(g.pages) + across - 1) / across
	return geom.NewSize(float32(across)*paper.Width+float32(across-1)*pageGridGap,
		float32(rows)*paper.Height+float32(rows-1)*pageGridGap)
}

// placePages sets the pages out in rows of as many as the grid's width holds.
func (g *pageGrid) placePages() {
	g.arranging = true
	defer func() { g.arranging = false }()
	paper := g.pages[0].paperSize()
	across := pagesAcross(g.ContentRect(false).Width, paper.Width)
	for i, page := range g.pages {
		page.SetFrameRect(geom.NewRect(float32(i%across)*(paper.Width+pageGridGap),
			float32(i/across)*(paper.Height+pageGridGap), paper.Width, paper.Height))
	}
}

// settle fills the pages again if their content changed since they were last filled, as the type's comment describes.
func (g *pageGrid) settle() {
	if g.settling || g.arranging {
		return
	}
	g.settling = true
	defer func() { g.settling = false }()
	for range maxPaginationPasses {
		if !g.needsFilling() {
			g.validatePages()
			if !g.needsFilling() {
				return
			}
		}
		g.fill()
	}
	// What the pages hold still changes as they are laid out, so the result is accepted rather than filling them again
	// on every pass.
	g.arranging = true
	g.validatePages()
	g.arranging = false
	g.contentChanged = false
}

// validatePages lays out whatever on the pages asked for it.
func (g *pageGrid) validatePages() {
	for _, page := range g.pages {
		page.ValidateLayout()
	}
}

// geometry returns the geometry of the pages as the page settings have it now.
func (g *pageGrid) geometry() pageGeometry {
	page := g.pages[0]
	return pageGeometry{size: page.paperSize(), insets: page.insets()}
}

// needsFilling returns true if the content of the pages changed since they were last filled.
func (g *pageGrid) needsFilling() bool {
	if g.invalid || g.contentChanged || g.filledFor != g.geometry() {
		return true
	}
	for key, list := range g.sheet.splittableLists() {
		if list.RowCount() != g.rowCounts[key] {
			return true
		}
	}
	return false
}

// fill fills the pages anew with the sheet's content, using the pages the grid has before adding any and dropping
// those left over.
func (g *pageGrid) fill() {
	g.arranging = true
	defer func() { g.arranging = false }()
	s := g.sheet
	// Everything starts out detached, so that a panel the layout no longer places, such as a derived list that ran out
	// of rows, isn't left in a band no page holds, where it would still seem to be on the sheet.
	for _, block := range s.blocks {
		block.AsPanel().RemoveFromParent()
	}
	for _, list := range s.splittableLists() {
		list.AsPanel().RemoveFromParent()
	}
	clear(g.handedOut)
	geometry := g.geometry()
	used := 0
	paginateBands(s.entity.SheetSettings.Layout.Root, s.pageLeaf, func() *Page {
		var page *Page
		if used < len(g.pages) {
			page = g.pages[used]
			page.RemoveAllChildren()
		} else {
			page = NewPage(s.entity)
			g.AddChild(page)
			g.pages = append(g.pages, page)
		}
		used++
		page.Force = false
		page.syncBorder()
		return page
	}, geometry.size)
	g.dropPagesAfter(used)
	for key, list := range s.splittableLists() {
		list.trimContinuations(g.handedOut[key])
		g.rowCounts[key] = list.RowCount()
	}
	g.filledFor = geometry
	g.invalid = false
	g.contentChanged = false
	// Every page was laid out at the grid's origin, which only the grid's own layout puts right (see placePages), so it
	// must follow even a filling LayoutSizes did while the grid wasn't marked for layout.
	g.NeedsLayout = true
}

// showContinuous puts the bands on the first page, which is then the one continuous page of the sheet, dropping every
// other page.
func (g *pageGrid) showContinuous(bands []*unison.Panel) {
	g.dropPagesAfter(1)
	page := g.pages[0]
	page.Force = false
	page.RemoveAllChildren()
	for _, band := range bands {
		page.AddChild(band)
	}
}

// joinLists gives each list all of its rows again, taking away its continuations.
func (g *pageGrid) joinLists() {
	for _, list := range g.sheet.splittableLists() {
		list.trimContinuations(0)
		list.ClearDrawRowRange()
	}
}

// dropPagesAfter takes the pages after the first n out of the grid.
func (g *pageGrid) dropPagesAfter(n int) {
	for _, page := range g.pages[n:] {
		page.RemoveAllChildren()
		page.RemoveFromParent()
	}
	clear(g.pages[n:])
	g.pages = g.pages[:n]
}
