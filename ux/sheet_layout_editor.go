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

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutedge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/layoutnode"
	"github.com/richardwilkes/gcs/v5/model/paper"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

var _ unison.Layout = &overlayStackLayout{}

// The sizes, in page pixels, of the things the layout editor draws and reacts to. They are float32 so that halving one
// of them yields the half rather than the integer division of it.
const (
	layoutButtonSize       float32 = 12 // The side of each of the square buttons in a block's top right corner
	layoutButtonGap        float32 = 2  // The space around and between the buttons in a block's top right corner
	layoutBottomEdgeHeight float32 = 5  // The height of the strip along a block's bottom edge that resizes it
	layoutDividerWidth     float32 = 9  // The width of the strip on the gap between two blocks that resizes them
	layoutBandBarHeight    float32 = 3  // The height of the bar drawn where a block would become a band of its own
	layoutMinHeightSlop    float32 = 4  // How far past its natural height a block's bottom edge must go to set one
	layoutAutoScrollAmount float32 = 24 // How far ahead of the pointer the page is scrolled during a drag
	layoutEdgeLadderStep   float32 = 8  // How wide the band of each container that shares the edge a drop is near is
	layoutSharedEdgeSlop   float32 = 1  // How far apart two edges may be and still count as the same edge
	layoutSeamThickness    float32 = 16 // How thick the strip on the seam between two things a drop can straddle is
	// How far from square the portrait's picture area may be and still be left alone, and how near to square a way of
	// squaring it has to land for it to count as having got there.
	layoutSquareSlop      float32 = 0.5
	layoutSquareTolerance float32 = 1
)

// overlayStackLayout lays a sheet's content out as the page at its preferred size with, while the block layout is
// being edited, the editor's overlay covering it exactly. The content has only ever held the one page, so this behaves
// the way the single-column layout it replaced did whenever there is no overlay.
type overlayStackLayout struct {
	page    *Page
	overlay *unison.Panel
}

// LayoutSizes implements unison.Layout.
func (l *overlayStackLayout) LayoutSizes(_ *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	return l.page.Sizes(hint)
}

// PerformLayout implements unison.Layout.
func (l *overlayStackLayout) PerformLayout(_ *unison.Panel) {
	_, prefSize, _ := l.page.Sizes(geom.Size{})
	l.page.SetFrameRect(geom.Rect{Size: prefSize})
	if l.overlay != nil {
		l.overlay.SetFrameRect(l.page.FrameRect())
	}
}

// layoutEditMode is what the layout editor is in the middle of doing.
type layoutEditMode byte

const (
	layoutIdle layoutEditMode = iota
	layoutPressed
	layoutDraggingBlock
	layoutDraggingDivider
	layoutDraggingBottom
	layoutContextPending
)

// layoutLeafRegion is one block of the layout as it appears on the page, in the overlay's coordinates. The ancestors
// are the indexes, within the regions' containers, of the rows and columns the block sits inside, outermost first.
type layoutLeafRegion struct {
	key       string
	node      *gurps.SheetLayoutNode
	panel     *unison.Panel
	ancestors []int
	rect      geom.Rect
	closeRect geom.Rect
	// The button that makes the portrait's picture area square, which sits just to the left of the close button. Only
	// the portrait has one, so this is the zero rect for every other block.
	squareRect    geom.Rect
	bottomRect    geom.Rect
	naturalHeight float32
	isRootBand    bool
	bandIndex     int
}

// layoutContainerRegion is one of the rows and columns the layout is built from, as it appears on the page, in the
// overlay's coordinates. The depth is how many containers it sits inside, so a band of the page is at zero. Dropping a
// block against the edge of one of these is how it is made to span everything the container holds.
type layoutContainerRegion struct {
	node       *gurps.SheetLayoutNode
	rect       geom.Rect
	depth      int
	isRootBand bool
	bandIndex  int
}

// layoutDividerRegion is the gap between two adjacent children of a row, dragging which moves width from one of them
// to the other. The nodes are the ones that govern the slots the two panels occupy, so they are the ones whose weights
// the row's layout is dividing the width by.
type layoutDividerRegion struct {
	rowPanel  *unison.Panel
	left      *gurps.SheetLayoutNode
	right     *gurps.SheetLayoutNode
	leftRect  geom.Rect
	rightRect geom.Rect
	rect      geom.Rect
	index     int
}

// layoutSeamRegion is the seam between two consecutive children of one of the rows and columns the layout is built
// from, the page's list of bands included. Dropping a block on one end of it makes the block straddle the two of them,
// spanning both, and dropping it in the middle puts it between them; see resolveDropTarget for the zones the strip is
// divided into. The nodes are the ones that govern the slots the two panels occupy, which are the ones the model has
// to be told to group.
type layoutSeamRegion struct {
	parent   *gurps.SheetLayoutNode
	first    *gurps.SheetLayoutNode
	second   *gurps.SheetLayoutNode
	rect     geom.Rect // The strip on the seam itself, which is what a drop has to be within
	spanRect geom.Rect // The two children together, which is what a block dropped here would span
	// How many containers the seam's own container sits inside. The page's list of bands is at -1, so that the seam of
	// a container always wins over the page's own at the same point.
	depth int
	// The position among the model's bands that a block dropped in the middle of the seam takes; -1 unless the seam is
	// between two of the page's bands.
	bandIndex int
	vertical  bool // Whether the two children sit side by side, so that the seam between them runs up and down
}

// layoutGapRegion is the space above the page's first band or below its last, dropping a block into which makes it a
// band of its own. The index is the position among the model's bands that a block dropped here takes. The space
// between two bands is a seam rather than a gap, since there is more than one thing that can be meant there.
type layoutGapRegion struct {
	rect  geom.Rect
	y     float32
	index int
}

// layoutRegions is everything the layout editor can point at, worked out from the panels that are actually on the page
// rather than from the tree, so that a block the builder left out can't be pointed at.
type layoutRegions struct {
	leaves     []layoutLeafRegion
	containers []layoutContainerRegion
	dividers   []layoutDividerRegion
	seams      []layoutSeamRegion
	gaps       []layoutGapRegion
	pageRect   geom.Rect
}

// leafFor returns the region of the block with the given key, or nil if it isn't on the page.
func (r *layoutRegions) leafFor(key string) *layoutLeafRegion {
	for i := range r.leaves {
		if r.leaves[i].key == key {
			return &r.leaves[i]
		}
	}
	return nil
}

// dropTargetKind is what dropping the block being dragged where the pointer is would do.
type dropTargetKind byte

const (
	dropNothing dropTargetKind = iota
	dropOnEdge
	dropAsBand
	dropStraddle
)

// dropTarget is where the block being dragged would land.
type dropTarget struct {
	key       string                 // The block it would be placed against, if that is a block; dropOnEdge only
	node      *gurps.SheetLayoutNode // The block or container it would be placed against; dropOnEdge only
	parent    *gurps.SheetLayoutNode // The container holding the pair it would span; dropStraddle only
	first     *gurps.SheetLayoutNode // The first of the pair it would span; dropStraddle only
	second    *gurps.SheetLayoutNode // The second of the pair it would span; dropStraddle only
	edge      layoutedge.Enum        // The side of the node or pair it would be placed on; not dropAsBand
	highlight geom.Rect              // What to draw to show where it would go
	bandIndex int                    // The band it would become; dropAsBand only
	kind      dropTargetKind
	bar       bool // Whether the highlight is a bar, which is drawn solid, rather than an area to be filled
}

// sheetLayoutEditor is the block layout editor of one sheet. While it is in place, a transparent overlay sits on top of
// the page, taking every mouse event that would otherwise have reached a field or a table, drawing the affordances and
// carrying out the gestures made on them. Nothing it draws can reach a printed or exported page, since neither builds
// its pages from this sheet's panels.
type sheetLayoutEditor struct {
	focusRef     *FocusRef // The focus to hand back when editing ends
	sheet        *Sheet
	overlay      *unison.Panel
	regions      *layoutRegions
	beforeLayout *gurps.SheetLayout
	divider      layoutDividerRegion
	bottom       layoutLeafRegion
	target       dropTarget
	hoverPt      geom.Point
	dragPt       geom.Point
	pressPt      geom.Point
	grabOffset   geom.Point
	dragKey      string
	pressKey     string
	closeKey     string
	squareKey    string
	button       int // The mouse button that began the gesture under way, if any
	mode         layoutEditMode
	hovering     bool
}

// newSheetLayoutEditor creates a layout editor for the given sheet. It does nothing until it is started.
func newSheetLayoutEditor(sheet *Sheet) *sheetLayoutEditor {
	return &sheetLayoutEditor{sheet: sheet}
}

// start puts the overlay in place and takes the keyboard focus, which pulls it out of whatever field held it.
func (e *sheetLayoutEditor) start() {
	overlay := unison.NewPanel()
	overlay.SetFocusable(true)
	// The sheet reroutes an item drop to whichever list it belongs in. Editing the layout suspends everything else the
	// page can do, so it suspends that as well rather than letting a drop land on a block that is being moved.
	overlay.CanAcceptDropCallback = func(_ drag.Info) bool { return false }
	overlay.DrawCallback = e.draw
	overlay.UpdateCursorCallback = e.cursorAt
	overlay.MouseDownCallback = func(where geom.Point, button, _ int, _ mod.Modifiers) bool {
		e.mouseDown(where, button)
		// Always claiming the press is what makes the window route the rest of the gesture here.
		return true
	}
	overlay.MouseDragCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		e.mouseDrag(where)
		return true
	}
	overlay.MouseUpCallback = func(where geom.Point, button int, _ mod.Modifiers) bool {
		e.mouseUp(where, button)
		return true
	}
	overlay.MouseMoveCallback = func(where geom.Point, _ mod.Modifiers) bool {
		e.mouseMove(where)
		return true
	}
	overlay.MouseExitCallback = func() bool {
		e.mouseExit()
		return true
	}
	overlay.KeyDownCallback = e.keyDown
	e.overlay = overlay
	// Index 0 is drawn last and hit-tested first, so the overlay covers the page no matter what else is added to the
	// content later.
	e.sheet.content.AddChildAtIndex(overlay, 0)
	e.sheet.contentLayout.overlay = overlay
	// A frame change anywhere below the page means the regions no longer describe where anything is. They are only
	// thrown away here, not recomputed, since this fires once per panel per layout pass.
	e.sheet.page.FrameChangeInChildHierarchyCallback = func(_ *unison.Panel) { e.regions = nil }
	e.sheet.content.MarkForLayoutAndRedraw()
	e.syncFrame()
	// Whatever had the focus is remembered so that it can be handed back when editing ends: the Edit menu's undo and
	// redo go to the focused panel's undo manager, so leaving the sheet without a focus would make the layout edits
	// look as though they had vanished from the undo stack.
	e.focusRef = e.sheet.targetMgr.CurrentFocusRef()
	overlay.RequestFocus()
}

// stop takes the overlay away, abandoning anything that was in the middle of being dragged.
func (e *sheetLayoutEditor) stop() {
	e.cancelDrag()
	e.sheet.page.FrameChangeInChildHierarchyCallback = nil
	e.sheet.contentLayout.overlay = nil
	if e.overlay != nil {
		e.overlay.RemoveFromParent()
		e.overlay = nil
	}
	e.regions = nil
	e.sheet.content.MarkForLayoutAndRedraw()
	e.restoreFocus()
	e.sheet.UpdateCursorNow()
}

// restoreFocus hands the focus back to whatever had it before editing began, or to the sheet's first focusable content
// when that is gone, so that the sheet -- and with it the undo manager the Edit menu consults -- stays focused.
func (e *sheetLayoutEditor) restoreFocus() {
	wnd := e.sheet.Window()
	if wnd == nil {
		return
	}
	focusRef := e.focusRef
	e.focusRef = nil
	if focusRef != nil {
		e.sheet.targetMgr.ReacquireFocus(focusRef, e.sheet.toolbar, e.sheet.scroll.Content())
	}
	// With nothing remembered, the focus is wherever removing the overlay left it -- which may be a toolbar button
	// unison picked on its own -- so the sheet's content is focused instead, as it is when a sheet is first shown.
	if focus := wnd.Focus(); focusRef == nil || focus == nil || !unison.AncestorIsOrSelf(focus, e.sheet) {
		FocusFirstContent(e.sheet.toolbar, e.sheet.scroll.Content())
	}
}

// syncFrame brings the overlay back over the page, which it has to be told about whenever the page changes size.
func (e *sheetLayoutEditor) syncFrame() {
	if e.overlay == nil {
		return
	}
	if r := e.sheet.page.FrameRect(); r != e.overlay.FrameRect() {
		e.overlay.SetFrameRect(r)
	}
	e.invalidateRegions()
}

// invalidateRegions discards what the editor knows about where everything is, so that the next thing that needs it
// works it out again.
func (e *sheetLayoutEditor) invalidateRegions() {
	e.regions = nil
}

// markForRedraw asks for the overlay to be drawn again.
func (e *sheetLayoutEditor) markForRedraw() {
	if e.overlay != nil {
		e.overlay.MarkForRedraw()
	}
}

// ensureRegions returns the regions, working them out first if they are stale.
func (e *sheetLayoutEditor) ensureRegions() *layoutRegions {
	if e.regions == nil {
		e.regions = e.buildRegions()
	}
	return e.regions
}

// rectOf returns the given panel's frame in the overlay's coordinates.
func (e *sheetLayoutEditor) rectOf(p *unison.Panel) geom.Rect {
	return e.overlay.RectFromRoot(p.RectToRoot(p.ContentRect(true)))
}

// sheetLayoutNodeOf returns the layout node that governs the slot the given panel occupies, or nil if the panel isn't
// one the layout builder made.
func sheetLayoutNodeOf(p *unison.Panel) *gurps.SheetLayoutNode {
	node, ok := p.ClientData()[sheetLayoutNodeKey].(*gurps.SheetLayoutNode)
	if !ok {
		return nil
	}
	return node
}

// panelKeys returns the block key of every panel the sheet could have put on the page. A panel's key can only be found
// this way: the node recorded on it is the one that governs its slot, which for a container that was left with a
// single child is the container's node rather than the surviving block's.
func (e *sheetLayoutEditor) panelKeys() map[*unison.Panel]string {
	keys := make(map[*unison.Panel]string, len(gurps.AllBlockKeys))
	for _, key := range gurps.AllBlockKeys {
		if p := e.sheet.blockPanel(key); !xreflect.IsNil(p) {
			keys[p.AsPanel()] = key
		}
	}
	return keys
}

// buildRegions works out where everything the editor can point at is, from the panels that are on the page: the blocks,
// the rows and columns they are gathered into, the dividers between the children of a row and the gaps between the
// bands of the page. The containers are there so that a block can be dropped against the edge of a whole row or column
// rather than only against the edge of a single block; see resolveDropTarget for how one of them is picked.
func (e *sheetLayoutEditor) buildRegions() *layoutRegions {
	regions := &layoutRegions{}
	if e.overlay == nil {
		return regions
	}
	page := e.sheet.page.AsPanel()
	regions.pageRect = e.rectOf(page)
	keys := e.panelKeys()
	root := e.sheet.entity.SheetSettings.Layout.Root
	bandIndexes := make(map[*gurps.SheetLayoutNode]int, len(root.Children))
	for i, child := range root.Children {
		bandIndexes[child] = i
	}
	bands := page.Children()
	bandRects := make([]geom.Rect, 0, len(bands))
	bandModelIndexes := make([]int, 0, len(bands))
	for _, band := range bands {
		modelIndex := -1
		if node := sheetLayoutNodeOf(band); node != nil {
			if i, exists := bandIndexes[node]; exists {
				modelIndex = i
			}
		}
		bandRects = append(bandRects, e.rectOf(band))
		bandModelIndexes = append(bandModelIndexes, modelIndex)
		e.collectRegions(regions, band, keys, nil, true, modelIndex)
	}
	// The page's list of bands is a column like any other, so the space between two bands is a seam like any other. It
	// sits at a depth of -1, since a band is at zero and a seam deeper in always wins.
	e.collectSeams(regions, root, bands, false, -1)
	e.collectGaps(regions, bandRects, bandModelIndexes, len(root.Children))
	return regions
}

// collectRegions adds the regions of the given panel and everything below it. The ancestors are the containers the
// panel sits inside, outermost first, which is the order the edge ladder resolveDropTarget uses walks them in.
func (e *sheetLayoutEditor) collectRegions(regions *layoutRegions, panel *unison.Panel, keys map[*unison.Panel]string,
	ancestors []int, isRootBand bool, bandIndex int,
) {
	node := sheetLayoutNodeOf(panel)
	if node == nil {
		return
	}
	if container := sheetLayoutContainerNodeOf(panel); container != nil {
		index := len(regions.containers)
		regions.containers = append(regions.containers, layoutContainerRegion{
			node:       node,
			rect:       e.rectOf(panel),
			depth:      len(ancestors),
			isRootBand: isRootBand,
			bandIndex:  bandIndex,
		})
		children := panel.Children()
		if container.Type == layoutnode.Row {
			e.collectDividers(regions, panel, children)
		}
		// The seams lie between this container's own children, so they belong to the node the panel was built from,
		// which isn't the node that governs the panel's slot when an outer container was dropped onto it.
		e.collectSeams(regions, container, children, container.Type == layoutnode.Row, len(ancestors))
		childAncestors := append(slices.Clip(ancestors), index)
		for _, child := range children {
			e.collectRegions(regions, child, keys, childAncestors, false, -1)
		}
		return
	}
	key := keys[panel]
	if key == "" {
		return
	}
	rect := e.rectOf(panel)
	_, prefSize, _ := panel.Sizes(geom.Size{Width: rect.Width})
	closeRect := geom.NewRect(rect.Right()-layoutButtonSize-layoutButtonGap, rect.Y+layoutButtonGap, layoutButtonSize,
		layoutButtonSize)
	var squareRect geom.Rect
	if key == gurps.BlockPortraitKey {
		squareRect = closeRect
		squareRect.X -= layoutButtonSize + layoutButtonGap
	}
	regions.leaves = append(regions.leaves, layoutLeafRegion{
		key:           key,
		node:          node,
		panel:         panel,
		ancestors:     ancestors,
		rect:          rect,
		closeRect:     closeRect,
		squareRect:    squareRect,
		bottomRect:    geom.NewRect(rect.X, rect.Bottom()-layoutBottomEdgeHeight, rect.Width, layoutBottomEdgeHeight),
		naturalHeight: prefSize.Height,
		isRootBand:    isRootBand,
		bandIndex:     bandIndex,
	})
}

// collectDividers adds the region of each gap between the adjacent children of a row.
func (e *sheetLayoutEditor) collectDividers(regions *layoutRegions, rowPanel *unison.Panel, children []*unison.Panel) {
	for i := 0; i+1 < len(children); i++ {
		left := sheetLayoutNodeOf(children[i])
		right := sheetLayoutNodeOf(children[i+1])
		if left == nil || right == nil {
			continue
		}
		leftRect := e.rectOf(children[i])
		rightRect := e.rectOf(children[i+1])
		center := (leftRect.Right() + rightRect.X) / 2
		top := min(leftRect.Y, rightRect.Y)
		regions.dividers = append(regions.dividers, layoutDividerRegion{
			rowPanel:  rowPanel,
			left:      left,
			right:     right,
			leftRect:  leftRect,
			rightRect: rightRect,
			rect: geom.NewRect(center-layoutDividerWidth/2, top, layoutDividerWidth,
				max(leftRect.Bottom(), rightRect.Bottom())-top),
			index: i,
		})
	}
}

// collectSeams adds the region of the seam between each pair of consecutive children of a container. The parent is the
// node the children were built from -- the one the model holds them in, which is the one a drop that straddles a pair
// of them concerns -- and the children are the panels the builder actually made, so a child that was pruned leaves the
// two neighbors it still has adjoining one another, with the seam between them. The depth is how many containers the
// container itself sits inside, or -1 for the page's list of bands, and vertical says whether the children sit side by
// side rather than stacked.
func (e *sheetLayoutEditor) collectSeams(regions *layoutRegions, parent *gurps.SheetLayoutNode,
	children []*unison.Panel, vertical bool, depth int,
) {
	if parent == nil {
		return
	}
	for i := 0; i+1 < len(children); i++ {
		first := sheetLayoutNodeOf(children[i])
		second := sheetLayoutNodeOf(children[i+1])
		if first == nil || second == nil {
			continue
		}
		bandIndex := -1
		if depth < 0 {
			// A block dropped in the middle of a seam of the page becomes the band the second of the pair is now, so
			// the model has to be able to say which that is.
			if bandIndex = slices.Index(parent.Children, second); bandIndex < 0 {
				continue
			}
		}
		firstRect := e.rectOf(children[i])
		secondRect := e.rectOf(children[i+1])
		spanRect := firstRect.Union(secondRect)
		var rect geom.Rect
		if vertical {
			center := (firstRect.Right() + secondRect.X) / 2
			rect = geom.NewRect(center-layoutSeamThickness/2, spanRect.Y, layoutSeamThickness, spanRect.Height)
		} else {
			center := (firstRect.Bottom() + secondRect.Y) / 2
			rect = geom.NewRect(spanRect.X, center-layoutSeamThickness/2, spanRect.Width, layoutSeamThickness)
		}
		regions.seams = append(regions.seams, layoutSeamRegion{
			parent:    parent,
			first:     first,
			second:    second,
			rect:      rect,
			spanRect:  spanRect,
			depth:     depth,
			bandIndex: bandIndex,
			vertical:  vertical,
		})
	}
}

// collectGaps adds the region above the page's first band and the one below its last. A gap takes the model index of
// the band below it, so that dropping a block into it puts it there; the one after the last band takes the index past
// the end. The space between two bands is a seam rather than a gap, since a block dropped there may be meant to
// straddle the two of them rather than to come between them.
func (e *sheetLayoutEditor) collectGaps(regions *layoutRegions, bandRects []geom.Rect, bandModelIndexes []int,
	bandCount int,
) {
	pageRect := regions.pageRect
	if len(bandRects) == 0 {
		regions.gaps = append(regions.gaps, newLayoutGapRegion(pageRect, pageRect.CenterY(), bandCount))
		return
	}
	if index := bandModelIndexes[0]; index >= 0 {
		regions.gaps = append(regions.gaps, newLayoutGapRegion(pageRect, (pageRect.Y+bandRects[0].Y)/2, index))
	}
	regions.gaps = append(regions.gaps, newLayoutGapRegion(pageRect,
		(bandRects[len(bandRects)-1].Bottom()+pageRect.Bottom())/2, bandCount))
}

// newLayoutGapRegion creates the region of the gap centered on the given y.
func newLayoutGapRegion(pageRect geom.Rect, y float32, index int) layoutGapRegion {
	return layoutGapRegion{
		rect:  layoutBandBar(pageRect, y),
		y:     y,
		index: index,
	}
}

// layoutBandBar returns the bar drawn across the page where a block would become a band of its own.
func layoutBandBar(pageRect geom.Rect, y float32) geom.Rect {
	return geom.NewRect(pageRect.X, y-layoutBandBarHeight/2, pageRect.Width, layoutBandBarHeight)
}

// layoutSeamBar returns the bar drawn along a seam where a block would come between the two things it lies between.
// It is the band bar of a seam that reaches no further than the pair of them does, and it runs up and down rather than
// across when the two of them sit side by side.
func layoutSeamBar(seam *layoutSeamRegion) geom.Rect {
	if seam.vertical {
		return geom.NewRect(seam.rect.CenterX()-layoutBandBarHeight/2, seam.spanRect.Y, layoutBandBarHeight,
			seam.spanRect.Height)
	}
	return geom.NewRect(seam.spanRect.X, seam.rect.CenterY()-layoutBandBarHeight/2, seam.spanRect.Width,
		layoutBandBarHeight)
}

// resolveDropTarget works out what dropping the block with the given key at the given point would do. It is a function
// of the regions and the point alone, so that what the editor does with a gesture can be checked without one. Three
// kinds of zone are tried, in this order: the seams between the children of a container, then the edge ladder of the
// block the pointer is over, then the gaps above and below the page's bands.
//
// A seam is the strip layoutSeamThickness thick that lies on the join between two consecutive children of a row or
// column, the page's list of bands included, and runs the length the two of them share. It is divided into thirds
// along that length. Either end asks for the block to straddle the pair: to stand against that end of the two of them
// taken together, spanning both, which is what nothing else can ask for and what a seam is for. The middle asks for
// the block to come between them, which for a seam of the page means a new band there and anywhere else means the
// block joining the pair's own container between the two of them. A seam whose own container sits deeper wins over one
// higher up at the same point, and a seam is passed over altogether when the block being dragged is one of the pair,
// since neither thing it offers would be a move.
//
// The block under the pointer says which edge is meant: whichever of its four sides the pointer is nearest. How far
// from that edge the pointer is then says which of the nested things that share the edge is meant, so that a block can
// be placed against a whole row or column rather than only against the one block it touches. The containers the block
// sits inside whose own edge is in the same place, outermost first, each own a band layoutEdgeLadderStep wide: within
// the first step of the edge the outermost of them is meant, within the second the next one in, and so on, until past
// the last of them the block itself is meant. An edge none of them share belongs to the block alone, however near the
// pointer is to it.
//
// The Top and Bottom edges of a band of the page are shared with the page itself, so the outermost rung of the ladder
// there -- the pointer less than layoutEdgeLadderStep from the edge, which whatever the ladder picked is within
// layoutSharedEdgeSlop of -- means a new band. Deeper than that, the band is meant as itself, and dropping a block
// against it stacks the two of them into a band group. A container that is a band of the page can only ever be the
// outermost thing sharing an edge, so ladderContainer only picks one within that first step and a container always
// means a new band; only a block that is a band of its own is reached deeper in. Away from every block, the nearer of
// the gaps above the first band and below the last means a new band there.
func resolveDropTarget(regions *layoutRegions, pt geom.Point, draggedKey string) dropTarget {
	if regions == nil || !pt.In(regions.pageRect) {
		return dropTarget{}
	}
	if seam := seamAt(regions, pt, draggedKey); seam != nil {
		return seamDropTarget(regions, seam, pt)
	}
	for i := range regions.leaves {
		leaf := &regions.leaves[i]
		if !pt.In(leaf.rect) {
			continue
		}
		if leaf.key == draggedKey {
			// Dropping a block on itself is not a move.
			return dropTarget{}
		}
		edge, distance := nearestEdge(leaf.rect, pt)
		key := leaf.key
		node := leaf.node
		rect := leaf.rect
		isRootBand := leaf.isRootBand
		bandIndex := leaf.bandIndex
		if container := ladderContainer(regions, leaf, edge, distance); container != nil {
			key = ""
			node = container.node
			rect = container.rect
			isRootBand = container.isRootBand
			bandIndex = container.bandIndex
		}
		if isRootBand && bandIndex >= 0 && distance < layoutEdgeLadderStep &&
			(edge == layoutedge.Top || edge == layoutedge.Bottom) {
			// The outermost step of the top or bottom edge of a band belongs to the page, so it means a new band there.
			// Deeper in, the band is meant as itself and the drop groups the two of them into a band of their own.
			index := bandIndex
			y := rect.Y
			if edge == layoutedge.Bottom {
				index++
				y = rect.Bottom()
			}
			return dropTarget{
				kind:      dropAsBand,
				bandIndex: index,
				highlight: layoutBandBar(regions.pageRect, y),
			}
		}
		return dropTarget{
			kind:      dropOnEdge,
			key:       key,
			node:      node,
			edge:      edge,
			highlight: edgeHalf(rect, edge),
		}
	}
	best := -1
	var bestDistance float32
	for i := range regions.gaps {
		if distance := xmath.Abs(pt.Y - regions.gaps[i].y); best == -1 || distance < bestDistance {
			best = i
			bestDistance = distance
		}
	}
	if best == -1 {
		return dropTarget{}
	}
	return dropTarget{
		kind:      dropAsBand,
		bandIndex: regions.gaps[best].index,
		highlight: regions.gaps[best].rect,
	}
}

// seamAt returns the seam the given point is on, or nil if it is on none of them. The innermost of the seams the point
// is on wins, so that the seam between two blocks within a band is reachable where it crosses the seam between that
// band and the next. A seam the block being dragged is one of the two sides of is passed over: standing beside itself
// and coming between itself and its neighbor are both where it already is.
func seamAt(regions *layoutRegions, pt geom.Point, draggedKey string) *layoutSeamRegion {
	var dragged *gurps.SheetLayoutNode
	if leaf := regions.leafFor(draggedKey); leaf != nil {
		dragged = leaf.node
	}
	var best *layoutSeamRegion
	for i := range regions.seams {
		seam := &regions.seams[i]
		if !pt.In(seam.rect) {
			continue
		}
		if dragged != nil && (seam.first == dragged || seam.second == dragged) {
			continue
		}
		if best == nil || seam.depth > best.depth {
			best = seam
		}
	}
	return best
}

// seamDropTarget returns what dropping the block on the given seam at the given point would do. See resolveDropTarget
// for the thirds the seam is divided into.
func seamDropTarget(regions *layoutRegions, seam *layoutSeamRegion, pt geom.Point) dropTarget {
	if seam.vertical {
		third := seam.rect.Height / 3
		switch {
		case pt.Y < seam.rect.Y+third:
			return straddleDropTarget(seam, layoutedge.Top)
		case pt.Y >= seam.rect.Bottom()-third:
			return straddleDropTarget(seam, layoutedge.Bottom)
		default:
			// Coming between two things that sit side by side means joining their row to the right of the first.
			return dropTarget{
				kind:      dropOnEdge,
				node:      seam.first,
				edge:      layoutedge.Right,
				highlight: layoutSeamBar(seam),
				bar:       true,
			}
		}
	}
	third := seam.rect.Width / 3
	switch {
	case pt.X < seam.rect.X+third:
		return straddleDropTarget(seam, layoutedge.Left)
	case pt.X >= seam.rect.Right()-third:
		return straddleDropTarget(seam, layoutedge.Right)
	case seam.bandIndex >= 0:
		// Coming between two of the page's bands means a band of the page rather than a group of the two of them.
		return dropTarget{
			kind:      dropAsBand,
			bandIndex: seam.bandIndex,
			highlight: layoutBandBar(regions.pageRect, seam.rect.CenterY()),
		}
	default:
		// Coming between two things that are stacked means joining their column below the first.
		return dropTarget{
			kind:      dropOnEdge,
			node:      seam.first,
			edge:      layoutedge.Bottom,
			highlight: layoutSeamBar(seam),
			bar:       true,
		}
	}
}

// straddleDropTarget returns the target that makes the block being dragged span both sides of the given seam, standing
// against the given edge of the two of them taken together.
func straddleDropTarget(seam *layoutSeamRegion, edge layoutedge.Enum) dropTarget {
	return dropTarget{
		kind:      dropStraddle,
		parent:    seam.parent,
		first:     seam.first,
		second:    seam.second,
		edge:      edge,
		highlight: edgeHalf(seam.spanRect, edge),
	}
}

// ladderContainer returns the container that the given distance from the given edge of the given block names, or nil
// if that distance is past every container sharing the edge and so names the block itself. See resolveDropTarget for
// what the ladder is for.
func ladderContainer(regions *layoutRegions, leaf *layoutLeafRegion, edge layoutedge.Enum,
	distance float32,
) *layoutContainerRegion {
	step := int(max(distance, 0) / layoutEdgeLadderStep)
	rung := 0
	for _, index := range leaf.ancestors {
		container := &regions.containers[index]
		if !sharesEdge(container.rect, leaf.rect, edge) {
			continue
		}
		if rung == step {
			return container
		}
		rung++
	}
	return nil
}

// sharesEdge returns true if the given edge of the two rects is in the same place, near enough that the pointer can't
// be pointing at one of them rather than the other.
func sharesEdge(a, b geom.Rect, edge layoutedge.Enum) bool {
	switch edge {
	case layoutedge.Left:
		return xmath.Abs(a.X-b.X) <= layoutSharedEdgeSlop
	case layoutedge.Right:
		return xmath.Abs(a.Right()-b.Right()) <= layoutSharedEdgeSlop
	case layoutedge.Bottom:
		return xmath.Abs(a.Bottom()-b.Bottom()) <= layoutSharedEdgeSlop
	default:
		return xmath.Abs(a.Y-b.Y) <= layoutSharedEdgeSlop
	}
}

// nearestEdge returns the side of the rect the point is closest to, along with how far from it the point is. Ties go to
// the left, then the right, then the top, so that the two sides a block is most often placed against win them.
func nearestEdge(r geom.Rect, pt geom.Point) (edge layoutedge.Enum, distance float32) {
	edge = layoutedge.Left
	distance = pt.X - r.X
	if d := r.Right() - pt.X; d < distance {
		edge = layoutedge.Right
		distance = d
	}
	if d := pt.Y - r.Y; d < distance {
		edge = layoutedge.Top
		distance = d
	}
	if d := r.Bottom() - pt.Y; d < distance {
		edge = layoutedge.Bottom
		distance = d
	}
	return edge, distance
}

// edgeHalf returns the half of the rect that lies against the given edge.
func edgeHalf(r geom.Rect, edge layoutedge.Enum) geom.Rect {
	switch edge {
	case layoutedge.Left:
		return geom.NewRect(r.X, r.Y, r.Width/2, r.Height)
	case layoutedge.Right:
		return geom.NewRect(r.X+r.Width/2, r.Y, r.Width/2, r.Height)
	case layoutedge.Bottom:
		return geom.NewRect(r.X, r.Y+r.Height/2, r.Width, r.Height/2)
	default:
		return geom.NewRect(r.X, r.Y, r.Width, r.Height/2)
	}
}

// closeAt returns the block whose close button is at the given point, if any.
func (e *sheetLayoutEditor) closeAt(where geom.Point) *layoutLeafRegion {
	regions := e.ensureRegions()
	for i := range regions.leaves {
		if where.In(regions.leaves[i].closeRect) {
			return &regions.leaves[i]
		}
	}
	return nil
}

// squareAt returns the block whose square button is at the given point, if any. Only the portrait has one, so this
// only ever finds that block.
func (e *sheetLayoutEditor) squareAt(where geom.Point) *layoutLeafRegion {
	regions := e.ensureRegions()
	for i := range regions.leaves {
		if !regions.leaves[i].squareRect.Empty() && where.In(regions.leaves[i].squareRect) {
			return &regions.leaves[i]
		}
	}
	return nil
}

// dividerAt returns the divider at the given point, if any.
func (e *sheetLayoutEditor) dividerAt(where geom.Point) *layoutDividerRegion {
	regions := e.ensureRegions()
	for i := range regions.dividers {
		if where.In(regions.dividers[i].rect) {
			return &regions.dividers[i]
		}
	}
	return nil
}

// bottomEdgeAt returns the block whose bottom edge is at the given point, if any.
func (e *sheetLayoutEditor) bottomEdgeAt(where geom.Point) *layoutLeafRegion {
	regions := e.ensureRegions()
	for i := range regions.leaves {
		if where.In(regions.leaves[i].bottomRect) {
			return &regions.leaves[i]
		}
	}
	return nil
}

// leafAt returns the block at the given point, if any.
func (e *sheetLayoutEditor) leafAt(where geom.Point) *layoutLeafRegion {
	regions := e.ensureRegions()
	for i := range regions.leaves {
		if where.In(regions.leaves[i].rect) {
			return &regions.leaves[i]
		}
	}
	return nil
}

// mouseDown starts whatever gesture the press begins. The close buttons, the dividers and the bottom edges are all
// small and sit on top of a block, so they are asked about first; a press on a block itself only becomes a move once
// the pointer has traveled far enough for it to be a drag rather than a click.
//
// The window delivers every press to the panel under the pointer, whichever buttons are already down, so a second
// button pressed while a gesture is under way arrives here too. It is ignored: it may neither abandon the gesture --
// which would leave whatever a drag had already written into the layout in place, with nothing recorded to undo it --
// nor start another on top of it. The gesture belongs to the button that began it, and only that button's release
// ends it.
func (e *sheetLayoutEditor) mouseDown(where geom.Point, button int) {
	if e.mode != layoutIdle {
		return
	}
	e.button = button
	if button == unison.ButtonRight {
		e.mode = layoutContextPending
		return
	}
	if button != unison.ButtonLeft {
		return
	}
	e.closeKey = ""
	e.squareKey = ""
	e.pressKey = ""
	e.pressPt = where
	if leaf := e.closeAt(where); leaf != nil {
		e.mode = layoutPressed
		e.closeKey = leaf.key
		return
	}
	if leaf := e.squareAt(where); leaf != nil {
		e.mode = layoutPressed
		e.squareKey = leaf.key
		return
	}
	if divider := e.dividerAt(where); divider != nil {
		e.beginDividerDrag(divider, where)
		return
	}
	if leaf := e.bottomEdgeAt(where); leaf != nil {
		e.beginBottomDrag(leaf, where)
		return
	}
	if leaf := e.leafAt(where); leaf != nil {
		e.mode = layoutPressed
		e.pressKey = leaf.key
		return
	}
	e.mode = layoutPressed
}

// mouseDrag carries whatever gesture is under way forward.
func (e *sheetLayoutEditor) mouseDrag(where geom.Point) {
	switch e.mode {
	case layoutPressed:
		if e.pressKey != "" && e.overlay != nil && e.overlay.IsDragGesture(e.pressPt) {
			e.beginBlockDrag(e.pressKey, e.pressPt)
			e.updateBlockDrag(where)
		}
	case layoutDraggingBlock:
		e.updateBlockDrag(where)
	case layoutDraggingDivider:
		e.updateDividerDrag(where)
	case layoutDraggingBottom:
		e.updateBottomDrag(where)
	default:
	}
}

// mouseUp finishes whatever gesture was under way, provided it is the button that began it that was released; see
// mouseDown for why the release of any other button is ignored.
func (e *sheetLayoutEditor) mouseUp(where geom.Point, button int) {
	if button != e.button {
		return
	}
	mode := e.mode
	e.mode = layoutIdle
	switch mode {
	case layoutContextPending:
		e.showContextMenu(where)
	case layoutPressed:
		closeKey := e.closeKey
		squareKey := e.squareKey
		e.closeKey = ""
		e.squareKey = ""
		e.pressKey = ""
		switch {
		case closeKey != "":
			if leaf := e.closeAt(where); leaf != nil && leaf.key == closeKey {
				e.sheet.hideLayoutBlock(closeKey)
			}
		case squareKey != "":
			if leaf := e.squareAt(where); leaf != nil && leaf.key == squareKey {
				e.squarePortrait()
			}
		default:
		}
	case layoutDraggingBlock:
		e.endBlockDrag(where)
	case layoutDraggingDivider:
		e.endDividerDrag(where)
	case layoutDraggingBottom:
		e.endBottomDrag(where)
	default:
	}
}

// mouseMove tracks the pointer so that what is under it can be pointed out.
func (e *sheetLayoutEditor) mouseMove(where geom.Point) {
	e.hovering = true
	e.hoverPt = where
	e.markForRedraw()
}

// mouseExit stops pointing anything out once the pointer has left the page.
func (e *sheetLayoutEditor) mouseExit() {
	if e.hovering {
		e.hovering = false
		e.markForRedraw()
	}
}

// keyDown handles the Escape key: the first one abandons a gesture that is under way, and the next leaves editing mode.
func (e *sheetLayoutEditor) keyDown(keyCode unison.KeyCode, _ mod.Modifiers, _ bool) bool {
	if keyCode != unison.KeyEscape {
		return false
	}
	if e.mode != layoutIdle {
		e.cancelDrag()
		return true
	}
	e.sheet.toggleLayoutEditing()
	return true
}

// beginBlockDrag starts moving the block with the given key. The layout as it stands is remembered here rather than
// at the press, since a press that never becomes a drag changes nothing and has nothing to undo.
func (e *sheetLayoutEditor) beginBlockDrag(key string, where geom.Point) {
	e.mode = layoutDraggingBlock
	e.beforeLayout = e.sheet.entity.SheetSettings.Layout.Clone()
	e.dragKey = key
	e.grabOffset = geom.Point{}
	if leaf := e.ensureRegions().leafFor(key); leaf != nil {
		e.grabOffset = where.Sub(leaf.rect.Point)
	}
	e.updateBlockDrag(where)
}

// updateBlockDrag follows the pointer, scrolling the page along with it when it nears an edge of the view.
func (e *sheetLayoutEditor) updateBlockDrag(where geom.Point) {
	e.dragPt = e.autoScroll(where)
	e.target = resolveDropTarget(e.ensureRegions(), e.dragPt, e.dragKey)
	e.markForRedraw()
}

// endBlockDrag drops the block being moved wherever the pointer is.
func (e *sheetLayoutEditor) endBlockDrag(where geom.Point) {
	key := e.dragKey
	target := resolveDropTarget(e.ensureRegions(), where, key)
	e.mode = layoutIdle
	e.dragKey = ""
	e.pressKey = ""
	e.target = dropTarget{}
	layout := e.sheet.entity.SheetSettings.Layout
	switch target.kind {
	case dropOnEdge:
		layout.MoveBeside(key, target.node, target.edge)
	case dropAsBand:
		layout.MoveToBand(key, target.bandIndex)
	case dropStraddle:
		layout.Straddle(key, target.first, target.second, target.edge)
	default:
		e.beforeLayout = nil
		e.invalidateRegions()
		e.markForRedraw()
		return
	}
	e.finish(i18n.Text("Move Block"))
}

// autoScroll brings the area just ahead of and just behind the pointer into view, then converts the point back, since
// scrolling has moved the page underneath it.
func (e *sheetLayoutEditor) autoScroll(where geom.Point) geom.Point {
	if e.overlay == nil {
		return where
	}
	root := e.overlay.PointToRoot(where)
	e.overlay.ScrollRectIntoView(geom.NewRect(where.X, where.Y-layoutAutoScrollAmount, 1, 1))
	e.overlay.ScrollRectIntoView(geom.NewRect(where.X, where.Y+layoutAutoScrollAmount, 1, 1))
	return e.overlay.PointFromRoot(root)
}

// beginDividerDrag starts moving width from one of a row's children to the next. The geometry the drag works from is
// the geometry at the moment it started, so that the blocks moving as the width is transferred can't feed back into
// it.
func (e *sheetLayoutEditor) beginDividerDrag(divider *layoutDividerRegion, where geom.Point) {
	e.mode = layoutDraggingDivider
	e.beforeLayout = e.sheet.entity.SheetSettings.Layout.Clone()
	// A copy is taken, since the regions are thrown away the moment the drag moves anything.
	e.divider = *divider
	e.updateDividerDrag(where)
}

// updateDividerDrag gives the two blocks either side of the divider the shares of the row the pointer calls for. The
// two weights always add up to what they added up to before, so nothing else in the row changes width.
func (e *sheetLayoutEditor) updateDividerDrag(where geom.Point) {
	divider := &e.divider
	span := divider.rightRect.Right() - divider.leftRect.X
	if span <= 0 {
		return
	}
	e.releaseSquareDivider(divider)
	total := divider.left.Weight + divider.right.Weight
	if total <= 0 {
		total = fxp.Two
	}
	lowest := layoutMinBlockWidth
	highest := span - layoutMinBlockWidth
	if highest < lowest {
		// The row is too narrow to hold two blocks at the narrowest a block may be, so the best that can be done is to
		// split it.
		lowest = span / 2
		highest = lowest
	}
	leftWidth := min(max(where.X-divider.leftRect.X, lowest), highest)
	leftWeight := total.Mul(fxp.FromFloat(leftWidth / span))
	rightWeight := total - leftWeight
	if leftWeight <= 0 || rightWeight <= 0 {
		return
	}
	divider.left.Weight = leftWeight
	divider.right.Weight = rightWeight
	// The row's layout was handed the weights as values when it was built, so it has to be told as well; nothing is
	// rebuilt until the drag is over.
	if row, ok := divider.rowPanel.Layout().(*weightedRowLayout); ok && divider.index+1 < len(row.weights) {
		row.weights[divider.index] = leftWeight
		row.weights[divider.index+1] = rightWeight
	}
	e.relayoutLive(divider.rowPanel)
}

// releaseSquareDivider hands a square block's width over to the weights so that the drag can move it. A block that
// takes its width from the height of its row ignores its weight altogether, so a divider drag that touches one -- on
// either side of the divider -- has to take the flag off it, in the model and in the row's live layout alike, or the
// weights the drag writes would change nothing. Every child of the row is first given the weight that reproduces the
// width it has on screen at that moment, so that nothing jumps as the flag comes off: the drag then moves the divider
// from where it stands rather than from wherever the old weights alone would have put it.
//
// This does nothing at all once there is no square block left beside the divider, so the drag can call it as often as
// it likes.
func (e *sheetLayoutEditor) releaseSquareDivider(divider *layoutDividerRegion) {
	row, ok := divider.rowPanel.Layout().(*weightedRowLayout)
	if !ok || (!row.squareOf(divider.index) && !row.squareOf(divider.index+1)) {
		return
	}
	children := divider.rowPanel.Children()
	for i, child := range children {
		weight := fxp.FromFloat(child.FrameRect().Width)
		if weight <= 0 {
			continue
		}
		if node := sheetLayoutNodeOf(child); node != nil {
			node.Weight = weight
		}
		if i < len(row.weights) {
			row.weights[i] = weight
		}
	}
	for _, index := range []int{divider.index, divider.index + 1} {
		if index >= len(children) || !row.squareOf(index) {
			continue
		}
		if node := squareLayoutNodeOf(children[index]); node != nil {
			node.Square = false
		}
		row.setSquare(index, false)
	}
}

// endDividerDrag finishes a width transfer.
func (e *sheetLayoutEditor) endDividerDrag(where geom.Point) {
	e.updateDividerDrag(where)
	e.mode = layoutIdle
	e.finish(i18n.Text("Resize Block"))
}

// beginBottomDrag starts setting the minimum height of a block.
func (e *sheetLayoutEditor) beginBottomDrag(leaf *layoutLeafRegion, where geom.Point) {
	e.mode = layoutDraggingBottom
	e.beforeLayout = e.sheet.entity.SheetSettings.Layout.Clone()
	// A copy is taken, since the regions are thrown away the moment the drag moves anything.
	e.bottom = *leaf
	e.updateBottomDrag(where)
}

// updateBottomDrag gives the block the minimum height the pointer calls for. A height that is no more than a little
// past what the block wanted anyway is taken as a request for its natural height, so that a block can be put back to
// growing with its content without having to land exactly on its current bottom edge.
//
// A block that takes its width from the height of its row keeps doing so: the minimum height set here raises the height
// of the row, which widens the block to match, so dragging the bottom edge of the portrait grows its picture area
// without taking it off its square. That is what the gesture asks for, so the flag is left alone -- it is the divider,
// which asks for a width the height can't give, that takes it off.
func (e *sheetLayoutEditor) updateBottomDrag(where geom.Point) {
	leaf := &e.bottom
	height := max(where.Y-leaf.rect.Y, 0)
	if height <= leaf.naturalHeight+layoutMinHeightSlop {
		height = 0
	}
	minHeight := paper.Length{Length: float64(height) / 72, Units: paper.Inch}
	if !e.sheet.entity.SheetSettings.Layout.SetMinHeight(leaf.key, minHeight) {
		return
	}
	// The builder is the sole authority on a panel's layout data, so the live value goes in the one place it keeps the
	// minimum height, whether the block's parent is a row or a column.
	if data, ok := leaf.panel.LayoutData().(*unison.FlexLayoutData); ok {
		data.MinSize.Height = minHeight.Pixels()
	}
	e.relayoutLive(leaf.panel)
}

// endBottomDrag finishes setting a minimum height.
func (e *sheetLayoutEditor) endBottomDrag(where geom.Point) {
	e.updateBottomDrag(where)
	e.mode = layoutIdle
	e.finish(i18n.Text("Set Block Height"))
}

// layoutSquareCandidate is what one way of making a block's content square turned out to do: how far from square the
// content was left and how far the dimension that moved had to travel to get there, as a fraction of what it was. The
// apply function makes the change the candidate stands for in the model, which is all that is needed once the winner
// has been picked, since applying it rebuilds the page from the model.
type layoutSquareCandidate struct {
	apply    func()
	err      float32
	relative float32
}

// squarePortrait makes the portrait's picture area square. That area is the block's content rect, so the shape the
// layout has to be brought to is that square plus the block's border insets.
//
// There are two ways to arrive at one, and each of them is tried on the page itself rather than reasoned about, since
// only the page can say what the rest of the layout leaves room for. The vertical candidate gives the block the
// minimum height that would make its content as tall as it is wide, exactly as dragging its bottom edge does. The
// horizontal candidate, available only when the block sits in a row, gives it the share of that row's width that would
// make its content as wide as it is tall, exactly as dragging the divider beside it does: with pool the width the row
// has to divide up and W the weights of its other children, the share that yields a frame width of target is
// W*target/(pool-target), which is out of reach when what is left over would push the other blocks below the narrowest
// a block may be. Each candidate is applied, the page is laid out again, the block is measured and what was there
// before is put back, so a height a taller neighbor swallows or a width the row refuses is seen for what it is.
//
// The winner is the candidate that actually lands on a square and moves the block the least, measured as a fraction of
// the dimension it changes, so that the portrait is nudged rather than being turned into a page-wide slab whenever the
// other way round would do just as well. If neither lands on one, the nearer miss wins, and if that is no better than
// the shape the block already has, nothing is done at all. Only the winner is applied for real, as a single undoable
// edit that rebuilds the sheet once.
func (e *sheetLayoutEditor) squarePortrait() {
	found := e.ensureRegions().leafFor(gurps.BlockPortraitKey)
	if found == nil || found.panel == nil {
		return
	}
	// A copy is taken, since the regions are thrown away the moment a candidate moves anything.
	leaf := *found
	var insets geom.Insets
	if border := leaf.panel.Border(); border != nil {
		insets = border.Insets()
	}
	content := leaf.panel.ContentRect(false).Size
	current := xmath.Abs(content.Width - content.Height)
	if current <= layoutSquareSlop {
		return
	}
	var best *layoutSquareCandidate
	for _, candidate := range []*layoutSquareCandidate{
		e.probeSquareByHeight(&leaf, content, insets),
		e.probeSquareByWidth(&leaf, content, insets),
	} {
		if candidate == nil {
			continue
		}
		if best == nil || betterSquareCandidate(candidate, best) {
			best = candidate
		}
	}
	if best == nil || best.err >= current {
		return
	}
	e.beforeLayout = e.sheet.entity.SheetSettings.Layout.Clone()
	best.apply()
	e.finish(i18n.Text("Square Portrait"))
}

// betterSquareCandidate returns true if the first candidate is the one to take. A candidate that lands on a square
// always beats one that doesn't; between two that land on one, the one that moves the block least wins, and between
// two that don't, the nearer miss does.
func betterSquareCandidate(candidate, best *layoutSquareCandidate) bool {
	candidateIsSquare := candidate.err < layoutSquareTolerance
	bestIsSquare := best.err < layoutSquareTolerance
	if candidateIsSquare != bestIsSquare {
		return candidateIsSquare
	}
	if candidateIsSquare {
		return candidate.relative < best.relative
	}
	return candidate.err < best.err
}

// probeSquareByHeight measures what giving the block the minimum height that would make its content as tall as it is
// wide actually does, then puts back the height it had. Returns nil if the block isn't in the tree.
func (e *sheetLayoutEditor) probeSquareByHeight(leaf *layoutLeafRegion, content geom.Size,
	insets geom.Insets,
) *layoutSquareCandidate {
	layout := e.sheet.entity.SheetSettings.Layout
	node, _, _ := layout.Find(leaf.key)
	if node == nil {
		return nil
	}
	key := leaf.key
	was := node.MinHeight
	target := paper.Length{Length: float64(content.Width+insets.Height()) / 72, Units: paper.Inch}
	e.applyLayoutMinHeight(leaf, target)
	result := leaf.panel.ContentRect(false).Size
	e.applyLayoutMinHeight(leaf, was)
	return &layoutSquareCandidate{
		apply:    func() { layout.SetMinHeight(key, target) },
		err:      xmath.Abs(result.Width - result.Height),
		relative: relativeLayoutChange(result.Height, content.Height),
	}
}

// applyLayoutMinHeight gives the block the given minimum height in the model and on the live panel alike, exactly as
// dragging its bottom edge does, and lays the page out again so that the result can be measured.
func (e *sheetLayoutEditor) applyLayoutMinHeight(leaf *layoutLeafRegion, minHeight paper.Length) {
	e.sheet.entity.SheetSettings.Layout.SetMinHeight(leaf.key, minHeight)
	if data, ok := leaf.panel.LayoutData().(*unison.FlexLayoutData); ok {
		data.MinSize.Height = minHeight.Pixels()
	}
	e.relayoutLive(leaf.panel)
}

// probeSquareByWidth measures what giving the block the share of its row that would make its content as wide as it is
// tall actually does, then puts back the share it had. A block whose width already comes from the height of its row
// ignores its weight, so the probe takes that flag off for the measurement and the change it stands for takes it off
// for good: asking for a width the height doesn't give is asking for the width to be the user's to set. Returns nil if
// the block isn't in a row, or if the row has no room for a block that wide.
func (e *sheetLayoutEditor) probeSquareByWidth(leaf *layoutLeafRegion, content geom.Size,
	insets geom.Insets,
) *layoutSquareCandidate {
	node := leaf.node
	if node == nil {
		return nil
	}
	if parent, _ := e.sheet.entity.SheetSettings.Layout.Parent(node); parent == nil || parent.Type != layoutnode.Row {
		return nil
	}
	rowPanel := leaf.panel.Parent()
	if rowPanel == nil {
		return nil
	}
	row, ok := rowPanel.Layout().(*weightedRowLayout)
	if !ok {
		return nil
	}
	count := len(rowPanel.Children())
	index := rowPanel.IndexOfChild(leaf.panel)
	if count < 2 || index < 0 || index >= len(row.weights) {
		return nil
	}
	target := content.Height + insets.Width()
	remainder := rowPanel.ContentRect(false).Width - row.hSpacing*float32(count-1) - target
	if target < layoutMinBlockWidth || remainder < float32(count-1)*layoutMinBlockWidth {
		// Squaring the block this way would leave the rest of the row below the narrowest a block may be, so the row
		// wouldn't give the block the width it asked for anyway.
		return nil
	}
	var others float32
	for i := range count {
		if i != index {
			others += row.weightOf(i)
		}
	}
	weight := fxp.FromFloat(others * target / remainder)
	if others <= 0 || weight <= 0 {
		return nil
	}
	wasNode := node.Weight
	wasLive := row.weights[index]
	wasSquare := row.squareOf(index)
	squareNode := squareLayoutNodeOf(leaf.panel)
	setLayoutRowSquare(row, index, squareNode, false)
	e.applyLayoutRowWeight(rowPanel, row, index, node, weight, weight)
	result := leaf.panel.ContentRect(false).Size
	setLayoutRowSquare(row, index, squareNode, wasSquare)
	e.applyLayoutRowWeight(rowPanel, row, index, node, wasNode, wasLive)
	return &layoutSquareCandidate{
		apply: func() {
			node.Weight = weight
			if squareNode != nil {
				squareNode.Square = false
			}
		},
		err:      xmath.Abs(result.Width - result.Height),
		relative: relativeLayoutChange(result.Width, content.Width),
	}
}

// applyLayoutRowWeight gives the block's slot the given share of its row in the model and in the row's live layout
// alike, exactly as dragging the divider beside it does, and lays the page out again so that the result can be
// measured. The two weights are always the same save when the pair the probe found is being put back, which is done
// value for value so that nothing is left changed.
func (e *sheetLayoutEditor) applyLayoutRowWeight(rowPanel *unison.Panel, row *weightedRowLayout, index int,
	node *gurps.SheetLayoutNode, nodeWeight, liveWeight fxp.Int,
) {
	node.Weight = nodeWeight
	row.weights[index] = liveWeight
	e.relayoutLive(rowPanel)
}

// setLayoutRowSquare says whether the block's slot takes its width from the height of its row, in the model and in the
// row's live layout alike. Nothing is laid out again here, since the caller has a weight to set as well and one pass
// serves for both.
func setLayoutRowSquare(row *weightedRowLayout, index int, node *gurps.SheetLayoutNode, square bool) {
	row.setSquare(index, square)
	if node != nil {
		node.Square = square
	}
}

// relativeLayoutChange returns how far a dimension moved as a fraction of what it was, which is what makes the two
// ways of squaring a block comparable even though one of them moves its width and the other its height.
func relativeLayoutChange(now, was float32) float32 {
	if was <= 0 {
		return 0
	}
	return xmath.Abs(now-was) / was
}

// relayoutLive shows the effect of a drag without rebuilding anything, so that the page follows the pointer.
func (e *sheetLayoutEditor) relayoutLive(p *unison.Panel) {
	p.MarkForLayoutRecursively()
	p.MarkForLayoutRecursivelyUpward()
	e.sheet.page.ApplyPreferredSize()
	e.syncFrame()
	e.markForRedraw()
}

// finish records the layout the gesture arrived at as a single undoable edit and shows it. A gesture that put the
// layout back the way it was -- a block dropped where it came from, a divider dragged back to where it started -- adds
// nothing to the undo stack and costs nothing to show, since nothing changed.
func (e *sheetLayoutEditor) finish(name string) {
	before := e.beforeLayout
	e.beforeLayout = nil
	if before == nil {
		return
	}
	layout := e.sheet.entity.SheetSettings.Layout
	if gurps.Hash64(before) == gurps.Hash64(layout) {
		e.invalidateRegions()
		e.markForRedraw()
		return
	}
	e.sheet.recordLayoutUndo(name, before, layout.Clone())
	rebuildAsModified(e.sheet, true)
}

// cancelDrag abandons whatever gesture is under way, putting back the layout it started from.
func (e *sheetLayoutEditor) cancelDrag() {
	if e.mode == layoutIdle {
		e.beforeLayout = nil
		return
	}
	e.mode = layoutIdle
	e.dragKey = ""
	e.pressKey = ""
	e.closeKey = ""
	e.squareKey = ""
	e.target = dropTarget{}
	before := e.beforeLayout
	e.beforeLayout = nil
	if before != nil && gurps.Hash64(before) != gurps.Hash64(e.sheet.entity.SheetSettings.Layout) {
		e.sheet.entity.SheetSettings.Layout = before
		// The layout is being put back exactly as it was, so nothing about the sheet has changed: the page is rebuilt
		// to show that, but the modification date is left alone, just as finish leaves it for a gesture that ended
		// where it began.
		e.sheet.Rebuild(true)
		return
	}
	e.invalidateRegions()
	e.markForRedraw()
}

// showContextMenu pops up the block layout menu, with the block that was clicked on as the one that can be hidden.
func (e *sheetLayoutEditor) showContextMenu(where geom.Point) {
	if e.overlay == nil || e.overlay.Window() == nil {
		return
	}
	hideKey := ""
	if leaf := e.leafAt(where); leaf != nil {
		hideKey = leaf.key
	}
	f := unison.DefaultMenuFactory()
	m := f.NewMenu(unison.PopupMenuTemporaryBaseID|unison.ContextMenuIDFlag, "", nil)
	id := 1
	e.sheet.appendLayoutMenuItems(f, m, &id, hideKey)
	e.overlay.FlushDrawing()
	m.Popup(geom.Rect{Point: e.overlay.PointToRoot(where), Width: 1, Height: 1}, 0)
	m.Dispose()
}

// cursorAt returns the cursor that says what the pointer would do where it is.
func (e *sheetLayoutEditor) cursorAt(where geom.Point) *unison.Cursor {
	switch e.mode {
	case layoutDraggingBlock:
		return unison.ClosedHandCursor()
	case layoutDraggingDivider:
		return unison.ResizeHorizontalCursor()
	case layoutDraggingBottom:
		return unison.ResizeVerticalCursor()
	default:
	}
	if e.closeAt(where) != nil || e.squareAt(where) != nil {
		return unison.PointingCursor()
	}
	if e.dividerAt(where) != nil {
		return unison.ResizeHorizontalCursor()
	}
	if e.bottomEdgeAt(where) != nil {
		return unison.ResizeVerticalCursor()
	}
	if e.leafAt(where) != nil {
		return unison.OpenHandCursor()
	}
	return unison.ArrowCursor()
}

// draw paints everything the editor shows on top of the page.
func (e *sheetLayoutEditor) draw(gc *unison.Canvas, _ geom.Rect) {
	regions := e.ensureRegions()
	dash := unison.NewDashPathEffect([]float32{3, 3}, 0)
	for i := range regions.leaves {
		leaf := &regions.leaves[i]
		outline := leaf.rect.Inset(geom.NewUniformInsets(0.5))
		paint := unison.ThemeFocus.Paint(gc, outline, paintstyle.Stroke)
		paint.SetStrokeWidth(1)
		paint.SetPathEffect(dash)
		gc.DrawRect(outline, paint)
		e.drawCloseButton(gc, leaf)
		e.drawSquareButton(gc, leaf)
	}
	if e.hovering && e.mode != layoutDraggingBlock {
		e.drawHoverHints(gc)
	}
	if e.mode == layoutDraggingBlock {
		e.drawDropFeedback(gc)
		e.drawGhost(gc)
	}
}

// drawCloseButton paints the button that takes a block off the sheet.
func (e *sheetLayoutEditor) drawCloseButton(gc *unison.Canvas, leaf *layoutLeafRegion) {
	e.drawLayoutButton(gc, leaf.closeRect, unison.ThemeError, unison.ThemeOnError, 3.5,
		func(mark geom.Rect, paint *unison.Paint) {
			gc.DrawLine(mark.Point, mark.BottomRight(), paint)
			gc.DrawLine(mark.TopRight(), mark.BottomLeft(), paint)
		})
}

// drawSquareButton paints the button that makes the portrait's picture area square. Only the portrait has one.
func (e *sheetLayoutEditor) drawSquareButton(gc *unison.Canvas, leaf *layoutLeafRegion) {
	if leaf.squareRect.Empty() {
		return
	}
	e.drawLayoutButton(gc, leaf.squareRect, unison.ThemeWarning, unison.ThemeOnWarning, 3,
		func(mark geom.Rect, paint *unison.Paint) { gc.DrawRect(mark, paint) })
}

// drawLayoutButton paints one of the small buttons in a block's title strip: a filled square with the glyph the given
// function draws on it, inset by the amount given. The pointer resting on it swaps the two inks for the pair given, so
// that what the button would do is shown before it is pressed.
func (e *sheetLayoutEditor) drawLayoutButton(gc *unison.Canvas, r geom.Rect,
	hoverBackground, hoverForeground unison.Ink, inset float32, glyph func(mark geom.Rect, paint *unison.Paint),
) {
	background := unison.Ink(unison.ThemeFocus)
	foreground := unison.Ink(unison.ThemeOnFocus)
	if e.hovering && e.mode != layoutDraggingBlock && e.hoverPt.In(r) {
		background = hoverBackground
		foreground = hoverForeground
	}
	gc.DrawRoundedRect(r, geom.NewSize(3, 3), background.Paint(gc, r, paintstyle.Fill))
	mark := r.Inset(geom.NewUniformInsets(inset))
	paint := foreground.Paint(gc, mark, paintstyle.Stroke)
	paint.SetStrokeWidth(1.5)
	glyph(mark, paint)
}

// drawHoverHints points out the divider or the bottom edge the pointer is on.
func (e *sheetLayoutEditor) drawHoverHints(gc *unison.Canvas) {
	if e.closeAt(e.hoverPt) != nil || e.squareAt(e.hoverPt) != nil {
		return
	}
	if divider := e.dividerAt(e.hoverPt); divider != nil {
		paint := unison.ThemeFocus.Paint(gc, divider.rect, paintstyle.Fill)
		paint.SetColorFilter(unison.Alpha30Filter())
		gc.DrawRect(divider.rect, paint)
		return
	}
	if leaf := e.bottomEdgeAt(e.hoverPt); leaf != nil {
		bar := geom.NewRect(leaf.rect.X, leaf.rect.Bottom()-layoutBandBarHeight, leaf.rect.Width, layoutBandBarHeight)
		gc.DrawRect(bar, unison.ThemeFocus.Paint(gc, bar, paintstyle.Fill))
	}
}

// drawDropFeedback shows where the block being dragged would land.
func (e *sheetLayoutEditor) drawDropFeedback(gc *unison.Canvas) {
	if e.target.kind == dropNothing {
		return
	}
	r := e.target.highlight
	if e.target.kind == dropAsBand || e.target.bar {
		// A block that comes between two things is shown as a line rather than an area, so it is drawn solid to be seen
		// at all, and so that it can't be taken for the filled half of the area a block would span.
		gc.DrawRect(r, unison.ThemeWarning.Paint(gc, r, paintstyle.Fill))
		return
	}
	fill := unison.ThemeWarning.Paint(gc, r, paintstyle.Fill)
	fill.SetColorFilter(unison.Alpha30Filter())
	gc.DrawRect(r, fill)
	stroke := unison.ThemeWarning.Paint(gc, r, paintstyle.Stroke)
	stroke.SetStrokeWidth(2)
	gc.DrawRect(r.Inset(geom.NewUniformInsets(1)), stroke)
}

// drawGhost paints the label of the block being dragged at the pointer.
func (e *sheetLayoutEditor) drawGhost(gc *unison.Canvas) {
	title := gurps.BlockTitle(e.dragKey)
	if title == "" {
		return
	}
	text := unison.NewText(title, &unison.TextDecoration{
		Font:            fonts.PageLabelPrimary,
		OnBackgroundInk: unison.ThemeOnFocus,
	})
	size := geom.NewSize(text.Width()+12, text.Height()+8)
	// The ghost is carried by the point within the block that was grabbed, but never so far from the pointer that the
	// pointer ends up outside it.
	offset := geom.NewPoint(min(e.grabOffset.X, size.Width-1), min(e.grabOffset.Y, size.Height-1))
	r := geom.Rect{Point: e.dragPt.Sub(offset), Size: size}
	gc.DrawRoundedRect(r, geom.NewSize(4, 4), unison.ThemeFocus.Paint(gc, r, paintstyle.Fill))
	text.Draw(gc, geom.NewPoint(r.X+6, r.Y+4+text.Baseline()))
}
