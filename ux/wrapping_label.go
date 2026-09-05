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

// defaultWrappingLabelWidth is the width a wrappingLabel wraps to when its layout has not offered it one.
const defaultWrappingLabelWidth = 400

// wrappingLabel shows read-only text that wraps to the width it is given, which a unison.Label, being a single line,
// does not. Within a FlexLayout it must be given HAlign: align.Fill, since that is what makes the layout offer it the
// column's width to wrap to; until it has been offered one, it wraps to defaultWrappingLabelWidth.
type wrappingLabel struct {
	unison.Panel
	text string
	ink  unison.Ink
	font unison.Font
}

func newWrappingLabel() *wrappingLabel {
	l := &wrappingLabel{
		ink:  unison.DefaultLabelTheme.OnBackgroundInk,
		font: unison.DefaultLabelTheme.Font,
	}
	l.Self = l
	l.SetSizer(l.sizes)
	l.DrawCallback = l.draw
	return l
}

// setText replaces the text and the ink it is drawn in, and asks for the label and its ancestors to be laid out again,
// since the number of lines may have changed.
func (l *wrappingLabel) setText(text string, ink unison.Ink) {
	l.text = text
	l.ink = ink
	l.MarkForLayoutRecursivelyUpward()
	l.MarkForRedraw()
}

// lines returns the text wrapped to the width, or to the default width when no usable width is given.
func (l *wrappingLabel) lines(width float32) []*unison.Text {
	if width <= 0 {
		width = defaultWrappingLabelWidth
	}
	return unison.NewTextWrappedLines(l.text, &unison.TextDecoration{Font: l.font, OnBackgroundInk: l.ink}, width)
}

func (l *wrappingLabel) sizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	var insets geom.Insets
	if b := l.Border(); b != nil {
		insets = b.Insets()
	}
	for _, line := range l.lines(hint.Width - insets.Width()) {
		prefSize.Width = max(prefSize.Width, line.Width())
		prefSize.Height += line.Height()
	}
	prefSize = prefSize.Add(insets.Size()).Ceil()
	return prefSize, prefSize, unison.MaxSize(prefSize)
}

func (l *wrappingLabel) draw(gc *unison.Canvas, _ geom.Rect) {
	rect := l.ContentRect(false)
	y := rect.Y
	for _, line := range l.lines(rect.Width) {
		line.Draw(gc, geom.NewPoint(rect.X, y+line.Baseline()))
		y += line.Height()
	}
}
