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

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/xmath"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// maxTradeRowsShown is the most rows a tradeList shows at once, beyond which it scrolls.
const maxTradeRowsShown = 10

// tradeList lists the items of a trade in a dialog, in a bordered box: a header naming its columns, drawn as the
// sheet's tables draw theirs, the rows, banded alternately, which scroll once there are more than maxTradeRowsShown of
// them, and a footer for the totals. The three are panels of their own, since only the rows scroll, each laid out by a
// tradeColumnsLayout in the columns they share, so that their columns line up.
type tradeList struct {
	box     *unison.Panel
	header  *unison.Panel
	rows    *unison.Panel
	footer  *unison.Panel
	columns *tradeColumns
}

// newTradeList returns an empty list with the given number of columns.
func newTradeList(columns int) *tradeList {
	l := &tradeList{columns: &tradeColumns{widths: make([]float32, columns)}}
	l.header = newTradeColumnsPanel(&tradeColumnsLayout{columns: l.columns}, unison.StdVSpacing)
	l.header.DrawCallback = func(gc *unison.Canvas, rect geom.Rect) {
		gc.DrawRect(rect, colors.Header.Paint(gc, rect, paintstyle.Fill))
	}
	rowsLayout := &tradeColumnsLayout{columns: l.columns}
	l.rows = newTradeColumnsPanel(rowsLayout, unison.StdVSpacing/2)
	l.rows.DrawCallback = func(gc *unison.Canvas, _ geom.Rect) {
		for i, band := range rowsLayout.bands {
			if i&1 == 1 {
				gc.DrawRect(band, unison.ThemeBanding.Paint(gc, band, paintstyle.Fill))
			}
		}
	}
	l.footer = newTradeColumnsPanel(&tradeColumnsLayout{columns: l.columns}, unison.StdVSpacing)
	l.footer.SetBorder(unison.NewCompoundBorder(
		unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.Insets{Top: 1}, false), l.footer.Border(),
	))
	l.columns.panels = []*unison.Panel{l.header, l.rows, l.footer}
	scroll := unison.NewScrollPanel()
	scroll.SetContent(l.rows, behavior.Fill, behavior.Fill)
	scroll.BackgroundInk = unison.ThemeSurface
	scroll.SetLayout(&tradeRowsLayout{Layout: scroll, rows: l.rows, layout: rowsLayout})
	scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	l.box = unison.NewPanel()
	l.box.SetBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{}, geom.NewUniformInsets(1), false))
	l.box.SetLayout(&unison.FlexLayout{Columns: 1})
	l.box.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})
	l.box.AddChild(l.header)
	l.box.AddChild(scroll)
	l.box.AddChild(l.footer)
	return l
}

// addHeader adds a column header spanning span columns, aligned within them as given.
func (l *tradeList) addHeader(title string, span int, hAlign align.Enum) *unison.Label {
	label := unison.NewLabel()
	label.OnBackgroundInk = colors.OnHeader
	label.SetTitle(title)
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: span, HAlign: hAlign, VAlign: align.Middle})
	l.header.AddChild(label)
	return label
}

// addTotal adds a line to the footer, right-aligned under the last column.
func (l *tradeList) addTotal() *unison.Label {
	label := unison.NewLabel()
	label.SetLayoutData(&unison.FlexLayoutData{HSpan: len(l.columns.widths), HAlign: align.End, VAlign: align.Middle})
	l.footer.AddChild(label)
	return label
}

// newTradeColumnsPanel returns a panel laid out by layout, inset from its sides as the header, the rows and the footer
// all are, so that their columns line up, and from its top and bottom by vertical. The right inset leaves room for the
// scroll bar over the rows, so that it never covers the last column.
func newTradeColumnsPanel(layout *tradeColumnsLayout, vertical float32) *unison.Panel {
	panel := unison.NewPanel()
	panel.SetLayout(layout)
	panel.SetBorder(unison.NewEmptyBorder(geom.Insets{
		Top:    vertical,
		Left:   unison.StdHSpacing,
		Bottom: vertical,
		Right:  max(unison.StdHSpacing, unison.DefaultScrollBarTheme.MinimumThickness),
	}))
	panel.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return panel
}

// tradeRowsLayout sizes the scroll panel holding a tradeList's rows to show them all, up to maxTradeRowsShown of them,
// beyond which it scrolls.
type tradeRowsLayout struct {
	unison.Layout
	rows   *unison.Panel
	layout *tradeColumnsLayout
}

// LayoutSizes implements unison.Layout.
func (l *tradeRowsLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	_, prefSize, maxSize = l.Layout.LayoutSizes(target, hint)
	most := l.layout.heightShowing(l.rows, maxTradeRowsShown)
	if border := target.Border(); border != nil {
		most += border.Insets().Height()
	}
	prefSize.Height = min(prefSize.Height, most)
	maxSize = geom.NewSize(max(maxSize.Width, prefSize.Width), max(maxSize.Height, prefSize.Height))
	return prefSize, prefSize, maxSize
}

// tradeColumns are the widths of the columns of a tradeList, shared by its header, rows and footer. Each width is that
// of the widest cell in the column across all three; the first column also takes whatever width is left over when a
// panel is laid out.
type tradeColumns struct {
	panels []*unison.Panel
	widths []float32
}

// measure finds the widths from the cells of the panels as they are now.
func (c *tradeColumns) measure() {
	clear(c.widths)
	var spanning []tradeCell
	for _, panel := range c.panels {
		for _, row := range tradeCellsOf(panel, len(c.widths)) {
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
		if short := cell.pref.Width - tradeSpanWidth(c.widths, cell.col, cell.span); short > 0 {
			c.widths[cell.col+cell.span-1] += short
		}
	}
}

// tradeSpanWidth returns the width of span columns starting at col, with the spacing between them.
func tradeSpanWidth(widths []float32, col, span int) float32 {
	var width float32
	for i := col; i < col+span; i++ {
		width += widths[i]
	}
	return width + unison.StdHSpacing*float32(span-1)
}

// tradeCell is a child of a panel laid out in a tradeList's columns, with the columns it spans and its preferred size.
type tradeCell struct {
	panel *unison.Panel
	data  *unison.FlexLayoutData
	col   int
	span  int
	pref  geom.Size
}

// tradeCellsOf returns the panel's children grouped into rows of columns cells. Each child fills the columns its
// FlexLayoutData spans, one by default, and a new row begins once a row's columns are full.
func tradeCellsOf(panel *unison.Panel, columns int) [][]tradeCell {
	var rows [][]tradeCell
	row := make([]tradeCell, 0, columns)
	col := 0
	for _, child := range panel.Children() {
		data, ok := child.LayoutData().(*unison.FlexLayoutData)
		if !ok {
			data = &unison.FlexLayoutData{}
		}
		span := min(max(data.HSpan, 1), columns-col)
		_, pref, _ := child.Sizes(geom.Size{})
		row = append(row, tradeCell{panel: child, data: data, col: col, span: span, pref: pref.Ceil()})
		if col += span; col >= columns {
			rows = append(rows, row)
			row = make([]tradeCell, 0, columns)
			col = 0
		}
	}
	if len(row) != 0 {
		rows = append(rows, row)
	}
	return rows
}

// tradeRowHeight returns the height of the row's tallest cell.
func tradeRowHeight(row []tradeCell) float32 {
	var height float32
	for _, cell := range row {
		height = max(height, cell.pref.Height)
	}
	return height
}

// tradeRowsHeight returns the height of the first count rows, with the spacing between them.
func tradeRowsHeight(rows [][]tradeCell, count int) float32 {
	var height float32
	for i, row := range rows[:count] {
		if i > 0 {
			height += unison.StdVSpacing
		}
		height += tradeRowHeight(row)
	}
	return height
}

// tradeColumnsLayout lays a panel's children out in rows of a tradeList's columns, within the panel's border, with
// unison.StdHSpacing between the columns and unison.StdVSpacing between the rows. A cell is placed within its columns
// and its row by the HAlign and VAlign of its FlexLayoutData.
type tradeColumnsLayout struct {
	columns *tradeColumns
	// bands are the rows as last laid out, each widened to the panel's edges and extended halfway to its neighbors, or
	// to the panel's edge, for banding them.
	bands []geom.Rect
}

// LayoutSizes implements unison.Layout.
func (l *tradeColumnsLayout) LayoutSizes(target *unison.Panel, _ geom.Size) (minSize, prefSize, maxSize geom.Size) {
	l.columns.measure()
	rows := tradeCellsOf(target, len(l.columns.widths))
	prefSize = geom.NewSize(tradeSpanWidth(l.columns.widths, 0, len(l.columns.widths)), tradeRowsHeight(rows, len(rows)))
	if border := target.Border(); border != nil {
		prefSize = prefSize.Add(border.Insets().Size())
	}
	prefSize = prefSize.Ceil()
	return prefSize, prefSize, unison.MaxSize(prefSize)
}

// heightShowing returns the height the panel needs to show its first count rows, or all of them if it has fewer.
func (l *tradeColumnsLayout) heightShowing(target *unison.Panel, count int) float32 {
	rows := tradeCellsOf(target, len(l.columns.widths))
	height := tradeRowsHeight(rows, min(count, len(rows)))
	if border := target.Border(); border != nil {
		height += border.Insets().Height()
	}
	return xmath.Ceil(height)
}

// PerformLayout implements unison.Layout.
func (l *tradeColumnsLayout) PerformLayout(target *unison.Panel) {
	l.columns.measure()
	r := target.ContentRect(false)
	widths := slices.Clone(l.columns.widths)
	widths[0] += max(r.Width-tradeSpanWidth(widths, 0, len(widths)), 0)
	xs := make([]float32, len(widths))
	x := r.X
	for i, width := range widths {
		xs[i] = x
		x += width + unison.StdHSpacing
	}
	rows := tradeCellsOf(target, len(widths))
	l.bands = l.bands[:0]
	full := target.ContentRect(true)
	y := r.Y
	for i, row := range rows {
		height := tradeRowHeight(row)
		for _, cell := range row {
			width := tradeSpanWidth(widths, cell.col, cell.span)
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

// wrapToWidthOf has a panel whose contents wrap to the width they are offered, such as a TextLabel, size itself, when
// it is offered none, for the width it has, or the width of another panel until it has been given one, whichever is
// wider. It never asks for more width than that other panel, so it never widens a dialog past it, while its height is
// that of its contents wrapped to the width it has, so that the dialog is sized for that.
func wrapToWidthOf(panel, other *unison.Panel) {
	wrapped := func(hint geom.Size, sizes unison.Sizer) (minSize, prefSize, maxSize geom.Size) {
		if hint.Width > 0 {
			return sizes(hint)
		}
		_, pref, _ := other.Sizes(geom.Size{})
		hint.Width = max(pref.Width, panel.FrameRect().Width)
		minSize, prefSize, maxSize = sizes(hint)
		minSize.Width = min(minSize.Width, pref.Width)
		prefSize.Width = min(prefSize.Width, pref.Width)
		return minSize, prefSize, maxSize
	}
	if layout := panel.Layout(); layout != nil {
		panel.SetLayout(&wrappedLayout{Layout: layout, wrapped: wrapped})
	} else if sizer := panel.Sizer(); sizer != nil {
		panel.SetSizer(func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) { return wrapped(hint, sizer) })
	}
}

// wrappedLayout is a layout whose sizes come from wrapped, given the hint and the layout's own sizing.
type wrappedLayout struct {
	unison.Layout
	wrapped func(hint geom.Size, sizes unison.Sizer) (minSize, prefSize, maxSize geom.Size)
}

// LayoutSizes implements unison.Layout.
func (l *wrappedLayout) LayoutSizes(target *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	return l.wrapped(hint, func(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
		return l.Layout.LayoutSizes(target, hint)
	})
}

// setLabelTitle replaces the label's text, if it differs, and has it laid out again for its new width.
func setLabelTitle(label *unison.Label, text string) {
	if label.String() != text {
		label.SetTitle(text)
		label.MarkForLayoutRecursivelyUpward()
		label.MarkForRedraw()
	}
}
