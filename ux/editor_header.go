// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/side"
)

// NewEditorListHeader creates a new list header for an editor.
func NewEditorListHeader[T gurps.NodeTypes](title, tooltip string, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		return NewPageTableColumnHeader[T](title, tooltip)
	}
	return NewTableColumnHeader[T](title, tooltip)
}

// NewEditorListSVGHeader creates a new list header with an SVG image as its content rather than text.
func NewEditorListSVGHeader[T gurps.NodeTypes](svg *unison.SVG, tooltip string, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip)
		baseline := header.Font.Baseline()
		header.Drawable = &unison.DrawableSVG{
			SVG:  svg,
			Size: unison.NewSize(baseline, baseline),
		}
		return header
	}
	header := NewTableColumnHeader[T]("", tooltip)
	baseline := header.Font.Baseline()
	header.Drawable = &unison.DrawableSVG{
		SVG:  svg,
		Size: unison.NewSize(baseline, baseline),
	}
	return header
}

// NewEditorListSVGPairHeader creates a new list header with a pair of SVG images as its content rather than text.
func NewEditorListSVGPairHeader[T gurps.NodeTypes](leftSVG, rightSVG *unison.SVG, tooltip string, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip)
		baseline := header.Font.Baseline()
		header.Drawable = &DrawableSVGPair{
			Left:  leftSVG,
			Right: rightSVG,
			Size:  unison.NewSize(baseline*2+4, baseline),
		}
		return header
	}
	header := NewTableColumnHeader[T]("", tooltip)
	baseline := header.Font.Baseline()
	header.Drawable = &DrawableSVGPair{
		Left:  leftSVG,
		Right: rightSVG,
		Size:  unison.NewSize(baseline*2+4, baseline),
	}
	return header
}

// NewEditorPageRefHeader creates a new page reference header.
func NewEditorPageRefHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGHeader[T](svg.Bookmark, gurps.PageRefTooltipText(), forPage)
}

// NewEditorEquippedHeader creates a new equipped header.
func NewEditorEquippedHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGHeader[T](unison.CheckmarkSVG,
		i18n.Text(`Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.`),
		forPage)
}

// NewEnabledHeader creates a new enabled header.
func NewEnabledHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGHeader[T](unison.CheckmarkSVG,
		i18n.Text(`Whether this item is enabled. Items that are not enabled do not apply any features they may normally contribute to the character.`),
		forPage)
}

// NewMoneyHeader creates a new money header.
func NewMoneyHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGHeader[T](svg.Coins,
		i18n.Text(`The value of one of these pieces of equipment`),
		forPage)
}

// NewExtendedMoneyHeader creates a new extended money page header.
func NewExtendedMoneyHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGPairHeader[T](svg.Stack, svg.Coins,
		i18n.Text(`The value of all of these pieces of equipment, plus the value of any contained equipment`), forPage)
}

// NewWeightHeader creates a new weight page header.
func NewWeightHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGHeader[T](svg.Weight,
		i18n.Text(`The weight of one of these pieces of equipment`),
		forPage)
}

// NewEditorExtendedWeightHeader creates a new extended weight page header.
func NewEditorExtendedWeightHeader[T gurps.NodeTypes](forPage bool) unison.TableColumnHeader[*Node[T]] {
	return NewEditorListSVGPairHeader[T](svg.Stack, svg.Weight,
		i18n.Text(`The weight of all of these pieces of equipment, plus the weight of any contained equipment`), forPage)
}

func headerFromData[T gurps.NodeTypes](data gurps.HeaderData, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if data.TitleIsImageKey {
		var img1, img2 *unison.SVG
		switch data.Title {
		case gurps.HeaderCheckmark:
			img1 = unison.CheckmarkSVG
		case gurps.HeaderCoins:
			img1 = svg.Coins
		case gurps.HeaderWeight:
			img1 = svg.Weight
		case gurps.HeaderBookmark:
			img1 = svg.Bookmark
		case gurps.HeaderStackedCoins:
			img1 = svg.Stack
			img2 = svg.Coins
		case gurps.HeaderStackedWeight:
			img1 = svg.Stack
			img2 = svg.Weight
		}
		if img2 != nil {
			return NewEditorListSVGPairHeader[T](img1, img2, data.Detail, forPage)
		}
		if img1 != nil {
			return NewEditorListSVGHeader[T](img1, data.Detail, forPage)
		}
	}
	return NewEditorListHeader[T](data.Title, data.Detail, forPage)
}

// NewTableColumnHeader creates a new table column header panel with the given title in small caps.
func NewTableColumnHeader[T gurps.NodeTypes](title, tooltip string) *unison.DefaultTableColumnHeader[*Node[T]] {
	header := unison.NewTableColumnHeader[*Node[T]](title, tooltip)
	header.Text = unison.NewSmallCapsText(title, &header.TextDecoration)
	return header
}

// PageTableColumnHeaderTheme holds the theme values for PageTableColumnHeaders. Modifying this data will not alter
// existing PageTableColumnHeaders, but will alter any PageTableColumnHeaders created in the future.
var PageTableColumnHeaderTheme = unison.LabelTheme{
	TextDecoration: unison.TextDecoration{
		Font:            gurps.PageLabelPrimaryFont,
		OnBackgroundInk: gurps.OnThemeHeader,
	},
	Gap:    3,
	HAlign: align.Middle,
	VAlign: align.Middle,
	Side:   side.Left,
}

var _ unison.TableColumnHeader[*Node[*gurps.Trait]] = &PageTableColumnHeader[*gurps.Trait]{}

// PageTableColumnHeader provides a default page table column header panel.
type PageTableColumnHeader[T gurps.NodeTypes] struct {
	*unison.Label
	sortState unison.SortState
}

// NewPageTableColumnHeader creates a new page table column header panel with the given title.
func NewPageTableColumnHeader[T gurps.NodeTypes](title, tooltip string) *PageTableColumnHeader[T] {
	h := &PageTableColumnHeader[T]{
		Label: unison.NewLabel(),
		sortState: unison.SortState{
			Order:     -1,
			Ascending: true,
			Sortable:  true,
		},
	}
	h.LabelTheme = PageTableColumnHeaderTheme
	h.Text = unison.NewSmallCapsText(title, &h.TextDecoration)
	h.Self = h
	h.SetSizer(h.DefaultSizes)
	h.DrawCallback = h.DefaultDraw
	h.MouseUpCallback = h.DefaultMouseUp
	if tooltip != "" {
		h.Tooltip = newWrappedTooltip(tooltip)
	}
	return h
}

// DefaultSizes provides the default sizing.
func (h *PageTableColumnHeader[T]) DefaultSizes(hint unison.Size) (minSize, prefSize, maxSize unison.Size) {
	_, prefSize, _ = h.Label.DefaultSizes(hint)
	if b := h.Border(); b != nil {
		prefSize = prefSize.Add(b.Insets().Size())
	}
	prefSize = prefSize.Ceil().ConstrainForHint(hint)
	return prefSize, prefSize, prefSize
}

// DefaultDraw provides the default drawing.
func (h *PageTableColumnHeader[T]) DefaultDraw(gc *unison.Canvas, dirty unison.Rect) {
	if h.sortState.Order == 0 {
		r := h.ContentRect(false)
		y := r.Y
		if h.sortState.Ascending {
			y = r.Bottom() - 1
		}
		gc.DrawLine(r.X, y, r.Right(), y, unison.ThemeFocus.Paint(gc, r, paintstyle.Stroke))
		save := h.OnBackgroundInk
		h.OnBackgroundInk = unison.ThemeFocus
		h.Label.DefaultDraw(gc, dirty)
		h.OnBackgroundInk = save
	} else {
		h.Label.DefaultDraw(gc, dirty)
	}
}

// SortState returns the current SortState.
func (h *PageTableColumnHeader[T]) SortState() unison.SortState {
	return h.sortState
}

// SetSortState sets the SortState.
func (h *PageTableColumnHeader[T]) SetSortState(state unison.SortState) {
	if h.sortState != state {
		h.sortState = state
		h.MarkForRedraw()
	}
}

// DefaultMouseUp provides the default mouse up handling.
func (h *PageTableColumnHeader[T]) DefaultMouseUp(where unison.Point, _ int, _ unison.Modifiers) bool {
	if h.sortState.Sortable && where.In(h.ContentRect(false)) {
		if header, ok := h.Parent().Self.(*unison.TableHeader[*Node[T]]); ok {
			header.SortOn(h)
			header.ApplySort()
		}
	}
	return true
}
