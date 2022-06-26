/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package editors

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/theme"
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

// NewHeader creates a new list header.
func NewHeader[T gurps.NodeConstraint[T]](title, tooltip string, forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	if forPage {
		return NewPageTableColumnHeader[T](title, tooltip)
	}
	return unison.NewTableColumnHeader[*ntable.Node[T]](title, tooltip)
}

// NewSVGHeader creates a new list header with an SVG image as its content rather than text.
func NewSVGHeader[T gurps.NodeConstraint[T]](svg *unison.SVG, tooltip string, forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip)
		baseline := header.Font.Baseline()
		header.Drawable = &unison.DrawableSVG{
			SVG:  svg,
			Size: unison.NewSize(baseline, baseline),
		}
		return header
	}
	header := unison.NewTableColumnHeader[*ntable.Node[T]]("", tooltip)
	baseline := header.Font.Baseline()
	header.Drawable = &unison.DrawableSVG{
		SVG:  svg,
		Size: unison.NewSize(baseline, baseline),
	}
	return header
}

// NewSVGPairHeader creates a new list header with a pair of SVG images as its content rather than text.
func NewSVGPairHeader[T gurps.NodeConstraint[T]](leftSVG, rightSVG *unison.SVG, tooltip string, forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip)
		baseline := header.Font.Baseline()
		header.Drawable = &widget.DrawableSVGPair{
			Left:  leftSVG,
			Right: rightSVG,
			Size:  unison.NewSize(baseline*2+4, baseline),
		}
		return header
	}
	header := unison.NewTableColumnHeader[*ntable.Node[T]]("", tooltip)
	baseline := header.Font.Baseline()
	header.Drawable = &widget.DrawableSVGPair{
		Left:  leftSVG,
		Right: rightSVG,
		Size:  unison.NewSize(baseline*2+4, baseline),
	}
	return header
}

// NewPageRefHeader creates a new page reference header.
func NewPageRefHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGHeader[T](res.BookmarkSVG, gurps.PageRefTooltipText, forPage)
}

// NewEquippedHeader creates a new equipped header.
func NewEquippedHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGHeader[T](res.CheckmarkSVG,
		i18n.Text(`Whether this piece of equipment is equipped or just carried. Items that are not equipped do not apply any features they may normally contribute to the character.`),
		forPage)
}

// NewEnabledHeader creates a new enabled header.
func NewEnabledHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGHeader[T](res.CheckmarkSVG,
		i18n.Text(`Whether this item is enabled. Items that are not enabled do not apply any features they may normally contribute to the character.`),
		forPage)
}

// NewMoneyHeader creates a new money header.
func NewMoneyHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGHeader[T](res.CoinsSVG,
		i18n.Text(`The value of one of these pieces of equipment`),
		forPage)
}

// NewExtendedMoneyHeader creates a new extended money page header.
func NewExtendedMoneyHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGPairHeader[T](res.StackSVG, res.CoinsSVG,
		i18n.Text(`The value of all of these pieces of equipment, plus the value of any contained equipment`), forPage)
}

// NewWeightHeader creates a new weight page header.
func NewWeightHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGHeader[T](res.WeightSVG,
		i18n.Text(`The weight of one of these pieces of equipment`),
		forPage)
}

// NewExtendedWeightHeader creates a new extended weight page header.
func NewExtendedWeightHeader[T gurps.NodeConstraint[T]](forPage bool) unison.TableColumnHeader[*ntable.Node[T]] {
	return NewSVGPairHeader[T](res.StackSVG, res.WeightSVG,
		i18n.Text(`The weight of all of these pieces of equipment, plus the weight of any contained equipment`), forPage)
}

// PageTableColumnHeaderTheme holds the theme values for PageTableColumnHeaders. Modifying this data will not alter
// existing PageTableColumnHeaders, but will alter any PageTableColumnHeaders created in the future.
var PageTableColumnHeaderTheme = unison.LabelTheme{
	Font:            theme.PageLabelPrimaryFont,
	OnBackgroundInk: theme.OnHeaderColor,
	Gap:             3,
	HAlign:          unison.MiddleAlignment,
	VAlign:          unison.MiddleAlignment,
	Side:            unison.LeftSide,
}

var _ unison.TableColumnHeader[*ntable.Node[*gurps.Trait]] = &PageTableColumnHeader[*gurps.Trait]{}

// PageTableColumnHeader provides a default page table column header panel.
type PageTableColumnHeader[T gurps.NodeConstraint[T]] struct {
	unison.Label
	sortState unison.SortState
}

// NewPageTableColumnHeader creates a new page table column header panel with the given title.
func NewPageTableColumnHeader[T gurps.NodeConstraint[T]](title, tooltip string) *PageTableColumnHeader[T] {
	h := &PageTableColumnHeader[T]{
		Label: unison.Label{
			LabelTheme: PageTableColumnHeaderTheme,
			Text:       title,
		},
		sortState: unison.SortState{
			Order:     -1,
			Ascending: true,
			Sortable:  true,
		},
	}

	h.Self = h
	h.SetSizer(h.DefaultSizes)
	h.DrawCallback = h.DefaultDraw
	h.MouseUpCallback = h.DefaultMouseUp
	if tooltip != "" {
		h.Tooltip = unison.NewTooltipWithText(tooltip)
	}
	return h
}

// DefaultSizes provides the default sizing.
func (h *PageTableColumnHeader[T]) DefaultSizes(hint unison.Size) (min, pref, max unison.Size) {
	_, pref, _ = h.Label.DefaultSizes(hint)
	if b := h.Border(); b != nil {
		pref.AddInsets(b.Insets())
	}
	pref.GrowToInteger()
	pref.ConstrainForHint(hint)
	return pref, pref, pref
}

// DefaultDraw provides the default drawing.
func (h *PageTableColumnHeader[T]) DefaultDraw(canvas *unison.Canvas, dirty unison.Rect) {
	if h.sortState.Order == 0 {
		canvas.DrawRect(dirty, theme.MarkerColor.Paint(canvas, dirty, unison.Fill))
		save := h.OnBackgroundInk
		h.OnBackgroundInk = theme.OnMarkerColor
		h.Label.DefaultDraw(canvas, dirty)
		h.OnBackgroundInk = save
	} else {
		h.Label.DefaultDraw(canvas, dirty)
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
	if h.sortState.Sortable && h.ContentRect(false).ContainsPoint(where) {
		if header, ok := h.Parent().Self.(*unison.TableHeader[*ntable.Node[T]]); ok {
			header.SortOn(h)
			header.ApplySort()
		}
	}
	return true
}
