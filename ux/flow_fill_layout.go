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
	"math"
	"slices"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var _ unison.Layout = &flowFillLayout{}

// unboundedWidth is the width flowFillLayout arranges children within when it has no width to fit them to.
const unboundedWidth = float32(math.MaxFloat32)

// flowFillLayout lays the children of its target out left to right at their preferred sizes, starting a new row when
// the next one doesn't fit, as unison.FlowLayout does, except for Fill, which is given whatever width is left on the
// row the children ahead of it end on, or a row of its own when less than its minimum width is left there. The children
// after Fill follow it on its row at their preferred widths, moving with it, as a status shown at the end of a toolbar
// does. A child is placed within the height of its row by the align.Enum it carries as layout data, as in a
// unison.FlowLayout; one without is placed at the top. The toolbars use it to keep their buttons and their search field
// on one row while they fit; see finishFlowToolbarLayout.
type flowFillLayout struct {
	// Fill is the child given what is left of its row: the last child when nil or not a child of the target.
	Fill     unison.Paneler
	HSpacing float32
	VSpacing float32
}

// LayoutSizes implements unison.Layout. The minimum width is the widest of the children's minimum widths, or of Fill's
// minimum width and the children that follow it, so that Fill is never left less than it needs, and the minimum height
// is that of the rows the hinted width calls for, so that none of them is cut off.
func (l *flowFillLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	var insets geom.Insets
	if b := target.Border(); b != nil {
		insets = b.Insets()
	}
	width := unboundedWidth
	if hint.Width >= 1 {
		width = max(hint.Width-insets.Width(), 0)
	}
	children := target.Children()
	_, prefSize, minSize.Width = l.arrange(children, width)
	for _, child := range children {
		childMin, _, _ := child.Sizes(geom.Size{})
		minSize.Width = max(minSize.Width, childMin.Width)
	}
	minSize.Height = prefSize.Height
	prefSize = prefSize.Add(insets.Size())
	return minSize.Add(insets.Size()), prefSize, unison.MaxSize(prefSize)
}

// PerformLayout implements unison.Layout.
func (l *flowFillLayout) PerformLayout(target *unison.Panel) {
	var insets geom.Insets
	if b := target.Border(); b != nil {
		insets = b.Insets()
	}
	children := target.Children()
	rects, _, _ := l.arrange(children, max(target.ContentRect(true).Width-insets.Width(), 0))
	for i, child := range children {
		rects[i].X += insets.Left
		rects[i].Y += insets.Top
		child.SetFrameRect(rects[i])
	}
}

// fillIndex returns the index of the child given what is left of its row; see Fill.
func (l *flowFillLayout) fillIndex(children []*unison.Panel) int {
	if l.Fill != nil {
		if i := slices.Index(children, l.Fill.AsPanel()); i >= 0 {
			return i
		}
	}
	return len(children) - 1
}

// arrange returns the frame of each child, relative to the top left of the target's content, the size that encloses
// them, and the least width the row holding Fill can be, within the given width. The children ahead of Fill are placed
// at their preferred sizes, wrapping as they go, with one wider than the given width narrowed to it, though never below
// its minimum width. Fill is placed after them on the row they end on, with the width left
// there once the children after it have been given their preferred widths, or on a row of its own when less than its
// minimum width would be left; within an unbounded width, it is given its preferred width.
func (l *flowFillLayout) arrange(children []*unison.Panel, width float32) (rects []geom.Rect, size geom.Size, fillRowMin float32) {
	rects = make([]geom.Rect, len(children))
	if len(children) == 0 {
		return rects, size, 0
	}
	var pt geom.Point
	var rowHeight float32
	rowStart := 0
	endRow := func(end int) {
		l.alignRow(children[rowStart:end], rects[rowStart:end], rowHeight)
		size.Height = pt.Y + rowHeight
		pt.X = 0
		pt.Y += rowHeight + l.VSpacing
		rowHeight = 0
		rowStart = end
	}
	fill := l.fillIndex(children)
	trailing := make([]geom.Size, 0, len(children)-fill-1)
	var trailingWidth float32
	for _, child := range children[fill+1:] {
		_, pref, _ := child.Sizes(geom.Size{})
		trailing = append(trailing, pref)
		trailingWidth += l.HSpacing + pref.Width
	}
	for i, child := range children[:fill] {
		childMin, pref, _ := child.Sizes(geom.Size{})
		if pt.X > 0 && pt.X+pref.Width > width {
			endRow(i)
		}
		if pref.Width > width {
			w := max(width, childMin.Width)
			_, pref, _ = child.Sizes(geom.NewSize(w, 0))
			pref.Width = w
		}
		rects[i] = geom.NewRect(pt.X, pt.Y, pref.Width, pref.Height)
		size.Width = max(size.Width, pt.X+pref.Width)
		rowHeight = max(rowHeight, pref.Height)
		pt.X += pref.Width + l.HSpacing
	}
	childMin, pref, _ := children[fill].Sizes(geom.Size{})
	fillRowMin = childMin.Width + trailingWidth
	fillWidth := pref.Width
	if width != unboundedWidth {
		if pt.X > 0 && width-pt.X < fillRowMin {
			endRow(fill)
		}
		fillWidth = max(width-pt.X-trailingWidth, 0)
	}
	rects[fill] = geom.NewRect(pt.X, pt.Y, fillWidth, pref.Height)
	rowHeight = max(rowHeight, pref.Height)
	pt.X += fillWidth + l.HSpacing
	for i, pref := range trailing {
		rects[fill+1+i] = geom.NewRect(pt.X, pt.Y, pref.Width, pref.Height)
		rowHeight = max(rowHeight, pref.Height)
		pt.X += pref.Width + l.HSpacing
	}
	size.Width = max(size.Width, pt.X-l.HSpacing)
	endRow(len(children))
	return rects, size, fillRowMin
}

// alignRow places the children of a completed row within its height by the align.Enum each carries as layout data, as
// a unison.FlowLayout does.
func (l *flowFillLayout) alignRow(children []*unison.Panel, rects []geom.Rect, rowHeight float32) {
	for i, child := range children {
		vAlign, ok := child.LayoutData().(align.Enum)
		if !ok {
			vAlign = align.Start
		}
		switch vAlign {
		case align.Middle:
			rects[i].Y += (rowHeight - rects[i].Height) / 2
		case align.End:
			rects[i].Y += rowHeight - rects[i].Height
		case align.Fill:
			rects[i].Height = rowHeight
		default: // align.Start
		}
	}
}
