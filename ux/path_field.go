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
	"github.com/richardwilkes/unison/enums/align"
)

// PathField is a non-editable field showing a file system path, shortened in its middle whenever the field is given
// less width than the whole path needs, so that the view holding it can be narrowed without scrolling sideways. It
// asks for the room the whole path needs, so the path is only shortened once the layout has nothing more to give. The
// whole path stays the field's name, for screen readers, and its tooltip, while the text they are given to read by
// character is the text actually drawn.
type PathField struct {
	NonEditableField
	path string
	full *unison.Text
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
	if path == "" {
		return f
	}
	f.Accessibility.Name = path
	f.Tooltip = newWrappedTooltip(path)
	f.SetSizer(f.sizes)
	f.FrameChangeCallback = f.fit
	return f
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

// fit shortens the path to the width the field has been given.
func (f *PathField) fit() {
	if text := elideMiddle(f.full, f.ContentRect(false).Width, &f.TextDecoration); text.String() != f.Text.String() {
		f.Text = text
		f.MarkForRedraw()
	}
}

// elideMiddle returns the text when it fits within the width, and otherwise the most of its start and end that fit
// around an ellipsis in their place, built with the decoration.
func elideMiddle(text *unison.Text, width float32, decoration *unison.TextDecoration) *unison.Text {
	if text.Width() <= width {
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
