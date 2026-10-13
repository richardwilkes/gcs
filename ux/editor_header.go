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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/colors"
	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/side"
)

// The spacing of the buttons at the end of a page table column header (see PageTableColumnHeader.addTrailingButtons):
// the gap that keeps the first clear of the title and the gap between neighbors.
const (
	headerButtonLeadingGap = 4
	headerButtonGap        = 2
)

// NewEditorListHeader creates a new list header for an editor.
func NewEditorListHeader[T gurps.Node[T]](title, tooltip string, less func(a, b string) bool, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		return NewPageTableColumnHeader[T](title, tooltip, less)
	}
	return NewTableColumnHeader[T](title, tooltip, less)
}

// NewEditorListSVGHeader creates a new list header with an SVG image as its content rather than text.
func NewEditorListSVGHeader[T gurps.Node[T]](icon *unison.SVG, tooltip string, less func(a, b string) bool, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip, less)
		baseline := header.Font.Baseline()
		header.Drawable = &unison.DrawableSVG{
			SVG:  icon,
			Size: geom.NewSize(baseline, baseline),
		}
		return header
	}
	header := NewTableColumnHeader[T]("", tooltip, less)
	baseline := header.Font.Baseline()
	header.Drawable = &unison.DrawableSVG{
		SVG:  icon,
		Size: geom.NewSize(baseline, baseline),
	}
	return header
}

// NewEditorListSVGPairHeader creates a new list header with a pair of SVG images as its content rather than text.
func NewEditorListSVGPairHeader[T gurps.Node[T]](leftSVG, rightSVG *unison.SVG, tooltip string, less func(a, b string) bool, forPage bool) unison.TableColumnHeader[*Node[T]] {
	if forPage {
		header := NewPageTableColumnHeader[T]("", tooltip, less)
		baseline := header.Font.Baseline()
		header.Drawable = &DrawableSVGPair{
			Left:  leftSVG,
			Right: rightSVG,
			Size:  geom.NewSize(baseline*2+4, baseline),
		}
		return header
	}
	header := NewTableColumnHeader[T]("", tooltip, less)
	baseline := header.Font.Baseline()
	header.Drawable = &DrawableSVGPair{
		Left:  leftSVG,
		Right: rightSVG,
		Size:  geom.NewSize(baseline*2+4, baseline),
	}
	return header
}

// groupContainersFirst wraps a column's own comparison so that it sees the value a cell shows rather than the marker
// CellDataForSort puts ahead of a container's while the "group containers on sort" preference is on. Without this, a
// comparison that reads a number out of the text -- which every numeric column's does -- reads the marker instead and
// reports every container as worth the same. The table's default comparison (flexibleLess) already does this for the
// columns that supply no comparison of their own.
func groupContainersFirst(less func(a, b string) bool) func(a, b string) bool {
	if less == nil {
		return nil
	}
	return func(a, b string) bool {
		aContainer := strings.HasPrefix(a, containerMarker)
		bContainer := strings.HasPrefix(b, containerMarker)
		if aContainer != bContainer {
			return aContainer
		}
		return less(strings.TrimPrefix(a, containerMarker), strings.TrimPrefix(b, containerMarker))
	}
}

func headerFromData[T gurps.Node[T]](data gurps.HeaderData, forPage bool) unison.TableColumnHeader[*Node[T]] {
	data.Less = groupContainersFirst(data.Less)
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
		case gurps.HeaderDatabase:
			img1 = svg.Database
		case gurps.HeaderSwitch:
			img1 = svg.ToggleSwitch
		case gurps.HeaderStackedCoins:
			img1 = svg.Stack
			img2 = svg.Coins
		case gurps.HeaderStackedWeight:
			img1 = svg.Stack
			img2 = svg.Weight
		}
		// Without a name, an icon header is called by its tooltip, which unison would read in front of the column's
		// value in every row, so it is given data.Name.
		if img2 != nil {
			header := NewEditorListSVGPairHeader[T](img1, img2, data.Detail, data.Less, forPage)
			header.AsPanel().Accessibility.Name = data.Name
			return header
		}
		if img1 != nil {
			header := NewEditorListSVGHeader[T](img1, data.Detail, data.Less, forPage)
			header.AsPanel().Accessibility.Name = data.Name
			return header
		}
	}
	header := NewEditorListHeader[T](data.Title, data.Detail, data.Less, forPage)
	// data.Name spells out an abbreviated title such as "Pts"; most text titles have none and are read as drawn.
	header.AsPanel().Accessibility.Name = data.Name
	return header
}

// NewTableColumnHeader creates a new table column header panel with the given title in small caps.
func NewTableColumnHeader[T gurps.Node[T]](title, tooltip string, less func(a, b string) bool) *unison.DefaultTableColumnHeader[*Node[T]] {
	header := unison.NewTableColumnHeader[*Node[T]](title, tooltip, less)
	header.Text = unison.NewSmallCapsText(title, &header.TextDecoration)
	return header
}

// PageTableColumnHeaderTheme holds the theme values for PageTableColumnHeaders. Modifying this data will not alter
// existing PageTableColumnHeaders, but will alter any PageTableColumnHeaders created in the future.
var PageTableColumnHeaderTheme = unison.LabelTheme{
	Font:            fonts.PageLabelPrimary,
	OnBackgroundInk: colors.OnHeader,
	Gap:             3,
	HAlign:          align.Middle,
	VAlign:          align.Middle,
	Side:            side.Left,
}

var _ unison.TableColumnHeader[*Node[*gurps.Trait]] = &PageTableColumnHeader[*gurps.Trait]{}

// PageTableColumnHeader provides a default page table column header panel.
type PageTableColumnHeader[T gurps.Node[T]] struct {
	*unison.Label
	less func(a, b string) bool
	// textTooltip is the tooltip for the header's text. Tooltip is normally the same, but holds a trailing button's
	// tooltip while the pointer is over that button (see buttonTooltip).
	textTooltip *unison.Panel
	buttons     []*unison.Button
	// pressed is the trailing button that took the last mouse down, which the matching mouse up goes to.
	pressed     *unison.Button
	sortState   unison.SortState
	tooltipText string
}

// headerLabel implements labelBackedHeader.
func (h *PageTableColumnHeader[T]) headerLabel() *unison.Label {
	return h.Label
}

// NewPageTableColumnHeader creates a new page table column header panel with the given title.
func NewPageTableColumnHeader[T gurps.Node[T]](title, tooltip string, less func(a, b string) bool) *PageTableColumnHeader[T] {
	h := &PageTableColumnHeader[T]{
		Label: unison.NewLabel(),
		less:  less,
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
	h.SetTooltipText(tooltip)
	return h
}

// TooltipText returns the text of the header's tooltip, or an empty string if it has none.
func (h *PageTableColumnHeader[T]) TooltipText() string {
	return h.tooltipText
}

// SetTooltipText sets the text of the header's tooltip, removing the tooltip entirely when the text is empty.
func (h *PageTableColumnHeader[T]) SetTooltipText(text string) {
	if h.tooltipText == text {
		return
	}
	h.tooltipText = text
	if text == "" {
		h.textTooltip = nil
	} else {
		h.textTooltip = newWrappedTooltip(text)
	}
	h.Tooltip = h.textTooltip
}

// DefaultSizes provides the default sizing, which includes the border, and with it the room reserved for any trailing
// buttons.
func (h *PageTableColumnHeader[T]) DefaultSizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	return h.Label.DefaultSizes(hint)
}

// DefaultDraw provides the default drawing.
func (h *PageTableColumnHeader[T]) DefaultDraw(gc *unison.Canvas, dirty geom.Rect) {
	if len(h.buttons) != 0 {
		// A button's tooltip lent to the column must not stay behind as the column's description.
		h.Tooltip = h.textTooltip
	}
	if h.sortState.Order == 0 {
		r := h.ContentRect(false)
		y := r.Y
		if h.sortState.Ascending {
			y = r.Bottom() - 1
		}
		gc.DrawLine(geom.NewPoint(r.X, y), geom.NewPoint(r.Right(), y),
			unison.ThemeFocus.Paint(gc, r, paintstyle.Stroke))
		save := h.OnBackgroundInk
		h.OnBackgroundInk = unison.ThemeFocus
		h.Label.DefaultDraw(gc, dirty)
		h.OnBackgroundInk = save
	} else {
		h.Label.DefaultDraw(gc, dirty)
	}
}

// Less returns the current less function.
func (h *PageTableColumnHeader[T]) Less() func(a, b string) bool {
	return h.less
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
func (h *PageTableColumnHeader[T]) DefaultMouseUp(where geom.Point, _ int, _ mod.Modifiers) bool {
	if h.sortState.Sortable && where.In(h.ContentRect(false)) {
		if header, ok := h.Parent().Self.(*unison.TableHeader[*Node[T]]); ok {
			header.SortOn(h)
			header.ApplySort()
		}
	}
	return true
}

// newHeaderButton returns an icon button sized to sit at the end of a page table column header without making it any
// taller. See PageTableColumnHeader.addTrailingButtons.
func newHeaderButton(icon *unison.SVG, tooltip string, onClick func()) *unison.Button {
	b := unison.NewSVGButton(icon)
	b.OnBackgroundInk = colors.OnHeader
	b.OnSelectionInk = colors.OnHeader
	b.Font = fonts.PageLabelPrimary
	b.DrawableOnlyVMargin = 0
	b.DrawableOnlyHMargin = 1
	if drawable, ok := b.Drawable.(*unison.DrawableSVG); ok {
		size := fonts.PageLabelPrimary.Baseline() - 2
		drawable.Size = geom.NewSize(size, size)
	}
	b.Tooltip = newWrappedTooltip(tooltip)
	b.Accessibility.Name = tooltip
	// The header is attached to the window only while it is drawn or handed an event, so the keyboard can't reach the
	// button, and a click that focused it would leave the focus on a panel about to be detached again.
	b.SetFocusable(false)
	b.ClickCallback = onClick
	return b
}

// addTrailingButtons adds the buttons to the end of the header, in a strip its border reserves, which keeps the title,
// the sort indicator and a click to sort clear of them. The header is not one of its table header's children but is
// installed only long enough to be drawn or handed an event (see unison.TableHeader), so the buttons never receive
// events of their own: the header passes on to them the mouse presses, cursor and tooltip requests it is given. It is
// described as a group, so that assistive technologies reach the buttons inside it.
func (h *PageTableColumnHeader[T]) addTrailingButtons(buttons ...*unison.Button) {
	for _, b := range buttons {
		h.AddChild(b)
	}
	h.buttons = append(h.buttons, buttons...)
	width := float32(headerButtonLeadingGap)
	for i, b := range h.buttons {
		if i != 0 {
			width += headerButtonGap
		}
		_, pref, _ := b.Sizes(geom.Size{})
		width += pref.Width
	}
	h.SetBorder(unison.NewEmptyBorder(geom.Insets{Right: width}))
	h.SetLayout(&trailingButtonsLayout[T]{header: h})
	h.Accessibility.Role = role.Group
	h.MouseDownCallback = h.buttonMouseDown
	h.MouseUpCallback = h.buttonMouseUp
	h.UpdateCursorCallback = h.buttonCursor
	h.UpdateTooltipCallback = h.buttonTooltip
}

// buttonAt returns the visible trailing button under the point, which is in the header's coordinates, or nil.
func (h *PageTableColumnHeader[T]) buttonAt(where geom.Point) *unison.Button {
	for _, b := range h.buttons {
		if !b.Hidden && where.In(b.FrameRect()) {
			return b
		}
	}
	return nil
}

func (h *PageTableColumnHeader[T]) buttonMouseDown(where geom.Point, button, clickCount int, mods mod.Modifiers) bool {
	h.pressed = nil
	if b := h.buttonAt(where); b != nil && b.Enabled() {
		h.pressed = b
		return b.MouseDownCallback(where.Sub(b.FrameRect().Point), button, clickCount, mods)
	}
	return true
}

func (h *PageTableColumnHeader[T]) buttonMouseUp(where geom.Point, button int, mods mod.Modifiers) bool {
	if b := h.pressed; b != nil {
		h.pressed = nil
		return b.MouseUpCallback(where.Sub(b.FrameRect().Point), button, mods)
	}
	return h.DefaultMouseUp(where, button, mods)
}

func (h *PageTableColumnHeader[T]) buttonCursor(where geom.Point) *unison.Cursor {
	if b := h.buttonAt(where); b != nil && b.UpdateCursorCallback != nil {
		return b.UpdateCursorCallback(where.Sub(b.FrameRect().Point))
	}
	return nil
}

// buttonTooltip lends the column the tooltip of the trailing button under the point, or else the header's own. The
// table header takes whatever Tooltip holds once this returns.
func (h *PageTableColumnHeader[T]) buttonTooltip(where geom.Point, suggestedAvoidInRoot geom.Rect) geom.Rect {
	if b := h.buttonAt(where); b != nil {
		h.Tooltip = b.Tooltip
		return b.RectToRoot(b.ContentRect(true)).Align()
	}
	h.Tooltip = h.textTooltip
	return suggestedAvoidInRoot
}

// trailingButtonsLayout sizes a page table column header just as its sizer would and lines up its trailing buttons at
// its right edge, centered vertically.
type trailingButtonsLayout[T gurps.Node[T]] struct {
	header *PageTableColumnHeader[T]
}

// LayoutSizes implements unison.Layout.
func (l *trailingButtonsLayout[T]) LayoutSizes(_ *unison.Panel, hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	return l.header.DefaultSizes(hint)
}

// PerformLayout implements unison.Layout.
func (l *trailingButtonsLayout[T]) PerformLayout(target *unison.Panel) {
	bounds := target.ContentRect(true)
	x := bounds.Right()
	for _, b := range slices.Backward(l.header.buttons) {
		_, pref, _ := b.Sizes(geom.Size{})
		x -= pref.Width
		b.SetFrameRect(geom.NewRect(x, bounds.Y+(bounds.Height-pref.Height)/2, pref.Width, pref.Height))
		x -= headerButtonGap
	}
}
