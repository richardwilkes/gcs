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
	"path/filepath"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
)

// PathField is a non-editable field showing a file system path, shortened in its middle whenever the field is given
// less width than the whole path needs, so that the view holding it can be narrowed without scrolling sideways. It
// asks for the room the whole path needs, so the path is only shortened once the layout has nothing more to give. The
// whole path stays the field's name, for screen readers, and is its tooltip while it is shortened, while the text they
// are given to read by character is the text actually drawn. SetProblem marks a path there is something wrong with,
// and SetClickCallback makes the field a button, for a path that can be changed by clicking it.
type PathField struct {
	NonEditableField
	path string
	full *unison.Text
	// problem is what is wrong with the path, if anything, which the tooltip says whether or not the path is shortened.
	problem       string
	clickCallback func()
	pressed       bool
	hovered       bool
	// The borders a clickable field shows: the plain one, the one the pointer over it calls for and the one the focus
	// does. They share their insets, so swapping them moves nothing.
	plainBorder, hoverBorder, focusBorder unison.Border
}

// NewPathField creates a new field showing the path, which may be empty.
func NewPathField(path string) *PathField {
	f := &PathField{path: path}
	f.init(f, func(field *NonEditableField) { field.SetTitle(path) }, align.Start)
	f.full = f.Text
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Middle,
		HGrab:  true,
	})
	f.Accessibility.Name = path
	describeWithHint(f, func() string { return f.problem })
	f.SetSizer(f.sizes)
	f.FrameChangeCallback = f.fit
	f.DrawCallback = f.draw
	return f
}

// SetProblem marks the path as one with the problem described, such as leading nowhere: a warning icon precedes the
// path, the field is washed with the error color as an invalid field is, and the tooltip says what the problem is
// whether or not the path is shortened, as does the field's description to a screen reader. An empty problem clears
// the mark.
func (f *PathField) SetProblem(problem string) {
	f.problem = problem
	if problem == "" {
		f.Drawable = nil
	} else {
		baseline := f.Font.Baseline()
		f.Drawable = &unison.DrawableSVG{
			SVG:  unison.TriangleExclamationSVG,
			Size: geom.NewSize(baseline, baseline),
		}
	}
	if f.FrameRect().Width > 0 {
		f.fit() // The icon takes room the text had, or gives it back
	}
	f.updateTooltip()
	f.MarkForLayoutRecursivelyUpward()
	f.MarkForRedraw()
}

// SetClickCallback makes the field a button that calls fn when it is clicked, or when Space is pressed while it has
// the focus, which it takes in its turn like any other field. Its border takes the focus color while the pointer is
// over it, and thickens while it has the focus, as an editable field's does. The path stays its name, so a caller
// should say what clicking does with describeWithHint.
func (f *PathField) SetClickCallback(fn func()) {
	f.clickCallback = fn
	f.SetFocusable(true)
	f.Accessibility.Role = role.Button
	f.plainBorder = f.Border()
	f.hoverBorder = nonEditableFieldBorder(unison.ThemeFocus)
	f.focusBorder = unison.NewDefaultFieldBorder(true)
	f.GainedFocusCallback = f.updateBorder
	f.LostFocusCallback = f.updateBorder
	f.MouseEnterCallback = func(_ geom.Point, _ mod.Modifiers) bool {
		f.hovered = true
		f.updateBorder()
		return true
	}
	f.MouseExitCallback = func() bool {
		f.hovered = false
		f.updateBorder()
		return true
	}
	f.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool {
		f.setPressed(true)
		return true
	}
	f.MouseDragCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		f.setPressed(where.In(f.ContentRect(true)))
		return true
	}
	f.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		f.setPressed(false)
		if where.In(f.ContentRect(true)) {
			fn()
		}
		return true
	}
	f.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if !unison.IsControlAction(keyCode, mods) {
			return false
		}
		// A held key acts once; its repeats are taken, so that they don't reach what holds the field.
		if !repeat {
			fn()
		}
		return true
	}
	f.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return unison.PointingCursor() }
	addAccessibilityCallback(f, func(node *accessibility.Node) {
		node.Actions = node.Actions.With(accessibility.Press)
	})
	f.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		if req.Action != accessibility.Press {
			return false
		}
		fn()
		return true
	}
}

// updateBorder gives a clickable field the border its state calls for.
func (f *PathField) updateBorder() {
	border := f.plainBorder
	switch {
	case f.Focused():
		border = f.focusBorder
	case f.hovered:
		border = f.hoverBorder
	}
	if f.Border() != border {
		f.SetBorder(border)
		f.MarkForRedraw()
	}
}

func (f *PathField) setPressed(pressed bool) {
	if f.pressed != pressed {
		f.pressed = pressed
		f.MarkForRedraw()
	}
}

// draw washes the field with the error color while there is a problem with the path and with the focus color while
// a click is being held on it, under the path.
func (f *PathField) draw(gc *unison.Canvas, rect geom.Rect) {
	r := f.ContentRect(true)
	if f.problem != "" {
		wash := &unison.ColorFilteredInk{OriginalInk: unison.ThemeError, ColorFilter: unison.Alpha30Filter()}
		gc.DrawRect(r, wash.Paint(gc, r, paintstyle.Fill))
	}
	if f.pressed {
		wash := &unison.ColorFilteredInk{OriginalInk: unison.ThemeFocus, ColorFilter: unison.Alpha30Filter()}
		gc.DrawRect(r, wash.Paint(gc, r, paintstyle.Fill))
	}
	f.DefaultDraw(gc, rect)
}

// sizes asks for the room the whole path needs, however much of it is showing.
func (f *PathField) sizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	prefSize, _ = unison.LabelContentSizes(f.full, f.Drawable, f.Font, f.Side, f.Gap)
	if b := f.Border(); b != nil {
		prefSize = prefSize.Add(b.Insets().Size())
	}
	prefSize = prefSize.Ceil().ConstrainForHint(hint)
	return prefSize, prefSize, prefSize
}

// fit shortens the path to the width the field has been given, less what an icon beside it takes.
func (f *PathField) fit() {
	width := f.ContentRect(false).Width
	if f.Drawable != nil {
		width -= f.Drawable.LogicalSize().Width + f.Gap
	}
	if text := elidePath(f.full, width, &f.TextDecoration); text.String() != f.Text.String() {
		f.Text = text
		f.MarkForRedraw()
	}
	f.updateTooltip()
}

// updateTooltip gives the field the tooltip its text calls for: the whole path while a shortened one is drawn, since
// the text itself says it all otherwise, with the problem, when there is one, under it or alone.
func (f *PathField) updateTooltip() {
	shortened := f.Text.String() != f.path
	switch {
	case shortened && f.problem != "":
		f.Tooltip = newWrappedTooltipWithSecondaryText(f.path, f.problem)
	case shortened:
		f.Tooltip = newWrappedTooltip(f.path)
	case f.problem != "":
		f.Tooltip = newWrappedTooltip(f.problem)
	default:
		f.Tooltip = nil
	}
}

// elidePath returns the text when it fits within the width, and otherwise shortens it as a path: its last element,
// which names what the path leads to, is kept whole with the middle of what leads to it replaced by an ellipsis, and
// only a last element that does not fit on its own has the whole path shortened in its middle instead.
func elidePath(text *unison.Text, width float32, decoration *unison.TextDecoration) *unison.Text {
	if text.Width() <= width {
		return text
	}
	dir, base := filepath.Split(text.String())
	if len(dir) < 2 {
		return elideMiddle(text, width, decoration)
	}
	// The separator goes with the last element, so that what leads to it is cut at an element's end.
	tail := unison.NewText(dir[len(dir)-1:]+base, decoration)
	ellipsis := unison.NewText("…", decoration)
	if tail.Width()+ellipsis.Width() > width {
		return elideMiddle(text, width, decoration)
	}
	lead := elideMiddle(unison.NewText(dir[:len(dir)-1], decoration), width-tail.Width(), decoration)
	// Kerning across the join may make the whole wider than its parts, so a result that does not fit after all is
	// given up for the plain shortening, which measures what it keeps whole.
	if result := unison.NewText(lead.String()+tail.String(), decoration); result.Width() <= width {
		return result
	}
	return elideMiddle(text, width, decoration)
}

// elideMiddle returns the text when it fits within the width or is empty, and otherwise the most of its start and end
// that fit around an ellipsis in their place, built with the decoration.
func elideMiddle(text *unison.Text, width float32, decoration *unison.TextDecoration) *unison.Text {
	if text.Width() <= width || text.Empty() {
		return text
	}
	runes := text.Runes()
	best := unison.NewText("…", decoration)
	// Binary search for the most runes that fit around the ellipsis: every count that fits is at most the answer, and
	// every count that doesn't is above it.
	lo, hi := 0, len(runes)-1
	for lo < hi {
		keep := (lo + hi + 1) / 2
		candidate := unison.NewTextFromRunes(elidedRunes(runes, keep), decoration)
		if candidate.Width() <= width {
			best = candidate
			lo = keep
		} else {
			hi = keep - 1
		}
	}
	return best
}

// elidedRunes returns keep of the runes, half from each end and the odd one from the start, around an ellipsis.
func elidedRunes(runes []rune, keep int) []rune {
	head := (keep + 1) / 2
	result := make([]rune, 0, keep+1)
	result = append(result, runes[:head]...)
	result = append(result, '…')
	return append(result, runes[len(runes)-(keep-head):]...)
}
