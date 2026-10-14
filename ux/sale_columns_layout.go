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
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// saleColumnCount is the number of columns in the list of items for sale: the item, the quantity to sell, the quantity
// carried, the full price of one and what the quantity sells for.
const saleColumnCount = 5

// saleColumns are the widths of the columns of the list of items for sale. The header, the rows and the footer of the
// list are separate panels, since only the rows scroll, so they share the widths to keep their columns lined up. Each
// width is that of the widest cell in the column across all of the panels; the first column, the item, also takes
// whatever width is left over when a panel is laid out.
type saleColumns struct {
	panels []*unison.Panel
	widths [saleColumnCount]float32
}

// measure finds the widths from the cells of the panels as they are now.
func (c *saleColumns) measure() {
	c.widths = [saleColumnCount]float32{}
	var spanning []saleCell
	for _, panel := range c.panels {
		for _, row := range saleCellsOf(panel) {
			for _, cell := range row {
				if cell.span > 1 {
					spanning = append(spanning, cell)
				} else {
					c.widths[cell.col] = max(c.widths[cell.col], cell.pref.Width)
				}
			}
		}
	}
	// A cell spanning several columns widens the last of them, should they not already hold it between them.
	for _, cell := range spanning {
		if short := cell.pref.Width - c.spanWidth(cell.col, cell.span); short > 0 {
			c.widths[cell.col+cell.span-1] += short
		}
	}
}

// spanWidth returns the width of span columns starting at col, with the spacing between them.
func (c *saleColumns) spanWidth(col, span int) float32 {
	return saleSpanWidth(c.widths, col, span)
}

func saleSpanWidth(widths [saleColumnCount]float32, col, span int) float32 {
	var width float32
	for i := col; i < col+span; i++ {
		width += widths[i]
	}
	return width + unison.StdHSpacing*float32(span-1)
}

// saleCell is a child of a panel laid out in the list's columns, with the columns it spans and its preferred size.
type saleCell struct {
	panel *unison.Panel
	data  *unison.FlexLayoutData
	col   int
	span  int
	pref  geom.Size
}

// saleCellsOf returns the panel's children grouped into rows. Each child fills the columns its FlexLayoutData spans, one
// by default, and a new row begins once a row's columns are full.
func saleCellsOf(panel *unison.Panel) [][]saleCell {
	var rows [][]saleCell
	row := make([]saleCell, 0, saleColumnCount)
	col := 0
	for _, child := range panel.Children() {
		data, ok := child.LayoutData().(*unison.FlexLayoutData)
		if !ok {
			data = &unison.FlexLayoutData{}
		}
		span := min(max(data.HSpan, 1), saleColumnCount-col)
		_, pref, _ := child.Sizes(geom.Size{})
		row = append(row, saleCell{panel: child, data: data, col: col, span: span, pref: pref.Ceil()})
		if col += span; col >= saleColumnCount {
			rows = append(rows, row)
			row = make([]saleCell, 0, saleColumnCount)
			col = 0
		}
	}
	if len(row) != 0 {
		rows = append(rows, row)
	}
	return rows
}

// saleRowHeight returns the height of the row's tallest cell.
func saleRowHeight(row []saleCell) float32 {
	var height float32
	for _, cell := range row {
		height = max(height, cell.pref.Height)
	}
	return height
}

// saleColumnsLayout lays a panel's children out in rows of the list's columns, within the panel's border, with
// unison.StdHSpacing between the columns and unison.StdVSpacing between the rows. A cell is placed within its columns
// and its row by the HAlign and VAlign of its FlexLayoutData.
type saleColumnsLayout struct {
	columns *saleColumns
	// bands are the rows as last laid out, each widened to the panel's edges and extended halfway to its neighbors, or
	// to the panel's edge, for banding them.
	bands []geom.Rect
}

// LayoutSizes implements unison.Layout.
func (l *saleColumnsLayout) LayoutSizes(target *unison.Panel, _ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	l.columns.measure()
	rows := saleCellsOf(target)
	prefSize = geom.NewSize(l.columns.spanWidth(0, saleColumnCount), saleRowsHeight(rows, len(rows)))
	if border := target.Border(); border != nil {
		prefSize = prefSize.Add(border.Insets().Size())
	}
	prefSize = prefSize.Ceil()
	return prefSize, prefSize, unison.MaxSize(prefSize)
}

// saleRowsHeight returns the height of the first count rows, with the spacing between them.
func saleRowsHeight(rows [][]saleCell, count int) float32 {
	var height float32
	for i, row := range rows[:count] {
		if i > 0 {
			height += unison.StdVSpacing
		}
		height += saleRowHeight(row)
	}
	return height
}

// heightShowing returns the height the panel needs to show its first count rows, or all of them if it has fewer.
func (l *saleColumnsLayout) heightShowing(target *unison.Panel, count int) float32 {
	rows := saleCellsOf(target)
	height := saleRowsHeight(rows, min(count, len(rows)))
	if border := target.Border(); border != nil {
		height += border.Insets().Height()
	}
	return xmath.Ceil(height)
}

// PerformLayout implements unison.Layout.
func (l *saleColumnsLayout) PerformLayout(target *unison.Panel) {
	l.columns.measure()
	r := target.ContentRect(false)
	widths := l.columns.widths
	widths[0] += max(r.Width-saleSpanWidth(widths, 0, saleColumnCount), 0)
	var xs [saleColumnCount]float32
	x := r.X
	for i, width := range widths {
		xs[i] = x
		x += width + unison.StdHSpacing
	}
	rows := saleCellsOf(target)
	l.bands = l.bands[:0]
	full := target.ContentRect(true)
	y := r.Y
	for i, row := range rows {
		height := saleRowHeight(row)
		for _, cell := range row {
			width := saleSpanWidth(widths, cell.col, cell.span)
			size := geom.NewSize(min(cell.pref.Width, width), min(cell.pref.Height, height))
			pt := geom.NewPoint(xs[cell.col], y)
			switch cell.data.HAlign {
			case align.Middle:
				pt.X += (width - size.Width) / 2
			case align.End:
				pt.X += width - size.Width
			case align.Fill:
				size.Width = width
			default:
			}
			switch cell.data.VAlign {
			case align.Middle:
				pt.Y += (height - size.Height) / 2
			case align.End:
				pt.Y += height - size.Height
			case align.Fill:
				size.Height = height
			default:
			}
			cell.panel.SetFrameRect(geom.Rect{Point: pt, Size: size})
		}
		band := geom.NewRect(full.X, y-unison.StdVSpacing/2, full.Width, height+unison.StdVSpacing)
		if i == 0 {
			band.Height += band.Y - full.Y
			band.Y = full.Y
		}
		if i == len(rows)-1 {
			band.Height = full.Bottom() - band.Y
		}
		l.bands = append(l.bands, band)
		y += height + unison.StdVSpacing
	}
}
