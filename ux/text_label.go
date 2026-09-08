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
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// defaultWrappingLabelWidth is the width a textLabel wraps to when its layout has not offered it one.
const defaultWrappingLabelWidth = 400

// pageRefPattern matches the Basic Set page references the calculators' notes cite: "B" and the page for the Characters
// book, "BX" and the page for the Campaigns book (page 338 onward), whether in parentheses like "(BX377)" or as each
// half of a range like "BX430-BX431". Only the Basic Set's keys are matched: every other page reference key is also a
// plausible piece of ordinary text (C4, TL6), and the Basic Set is the only book the notes cite.
var pageRefPattern = regexp.MustCompile(`\bBX?\d{1,3}\b`)

// textLabel shows read-only text whose page references, once linkPageRefs has been called, are drawn as links that open
// when clicked, which a unison.Label cannot do. Made with newWrappingLabel it wraps to the width it is given: within a
// FlexLayout it must then be given HAlign: align.Fill, since that is what makes the layout offer it the column's width
// to wrap to, and until it has been offered one it wraps to defaultWrappingLabelWidth. Made with newSingleLineLabel it
// stays on one line, as a unison.Label does, so it can stand in for one beside a field or a popup.
type textLabel struct {
	unison.Panel
	text        string
	ink         unison.Ink
	font        unison.Font
	linkHandler func(ref string)
	pressedRef  string
	wrap        bool
}

// pageRefLink is where one page reference lies within the label, so that clicks and the cursor can find it.
type pageRefLink struct {
	ref  string
	rect geom.Rect
}

// newWrappingLabel returns a label that wraps its text to the width it is given.
func newWrappingLabel() *textLabel {
	return newTextLabel(true)
}

// newSingleLineLabel returns a label that keeps its text on one line, however long it is.
func newSingleLineLabel() *textLabel {
	return newTextLabel(false)
}

func newTextLabel(wrap bool) *textLabel {
	l := &textLabel{
		ink:  unison.DefaultLabelTheme.OnBackgroundInk,
		font: unison.DefaultLabelTheme.Font,
		wrap: wrap,
	}
	l.Self = l
	l.SetSizer(l.sizes)
	l.DrawCallback = l.draw
	return l
}

// SetTitle replaces the text, keeping the ink, so that the label can be used where a unison.Label was.
func (l *textLabel) SetTitle(text string) {
	l.setText(text, l.ink)
}

// Title returns the text.
func (l *textLabel) Title() string {
	return l.text
}

func (l *textLabel) String() string {
	return l.text
}

// setText replaces the text and the ink it is drawn in, then asks for the label and its ancestors to be laid out again,
// since the number of lines may have changed.
func (l *textLabel) setText(text string, ink unison.Ink) {
	l.text = text
	l.ink = ink
	l.MarkForLayoutRecursivelyUpward()
	l.MarkForRedraw()
}

// linkPageRefs makes the page references in the text into links, drawn in the link theme, that pass the reference to
// the handler when clicked.
func (l *textLabel) linkPageRefs(handler func(ref string)) {
	l.linkHandler = handler
	l.MouseDownCallback = l.mouseDown
	l.MouseUpCallback = l.mouseUp
	l.UpdateCursorCallback = l.updateCursor
}

// paragraphs returns the text as one unison.Text per logical line, with the page references decorated as links when
// they are clickable.
func (l *textLabel) paragraphs() []*unison.Text {
	base := &unison.TextDecoration{Font: l.font, OnBackgroundInk: l.ink}
	link := &unison.TextDecoration{Font: l.font, OnBackgroundInk: unison.DefaultLinkTheme.OnBackgroundInk, Underline: true}
	split := strings.Split(l.text, "\n")
	paragraphs := make([]*unison.Text, 0, len(split))
	for _, para := range split {
		text := unison.NewText("", base)
		last := 0
		if l.linkHandler != nil {
			for _, m := range pageRefPattern.FindAllStringIndex(para, -1) {
				text.AddString(para[last:m[0]], base)
				text.AddString(para[m[0]:m[1]], link)
				last = m[1]
			}
		}
		text.AddString(para[last:], base)
		paragraphs = append(paragraphs, text)
	}
	return paragraphs
}

// lines returns the text wrapped to the width, or to the default width when no usable width is given. A single-line
// label's paragraphs are returned as they are, however wide.
func (l *textLabel) lines(width float32) []*unison.Text {
	paragraphs := l.paragraphs()
	if !l.wrap {
		return paragraphs
	}
	if width <= 0 {
		width = defaultWrappingLabelWidth
	}
	lines := make([]*unison.Text, 0, len(paragraphs))
	for _, para := range paragraphs {
		lines = append(lines, para.BreakToWidth(width)...)
	}
	return lines
}

func (l *textLabel) sizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
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

func (l *textLabel) draw(gc *unison.Canvas, _ geom.Rect) {
	rect := l.ContentRect(false)
	y := rect.Y
	for _, line := range l.lines(rect.Width) {
		line.Draw(gc, geom.NewPoint(rect.X, y+line.Baseline()))
		y += line.Height()
	}
}

// links returns the page references in the text as it is laid out right now, each with the rectangle it occupies. A
// reference is a single word, so wrapping never splits one across lines and each is found whole within its line.
func (l *textLabel) links() []pageRefLink {
	if l.linkHandler == nil {
		return nil
	}
	rect := l.ContentRect(false)
	var links []pageRefLink
	y := rect.Y
	for _, line := range l.lines(rect.Width) {
		s := line.String()
		for _, m := range pageRefPattern.FindAllStringIndex(s, -1) {
			start := utf8.RuneCountInString(s[:m[0]])
			end := start + utf8.RuneCountInString(s[m[0]:m[1]])
			x := rect.X + line.PositionForRuneIndex(start)
			links = append(links, pageRefLink{
				ref:  s[m[0]:m[1]],
				rect: geom.NewRect(x, y, rect.X+line.PositionForRuneIndex(end)-x, line.Height()),
			})
		}
		y += line.Height()
	}
	return links
}

// linkAt returns the page reference under the point, or an empty string when there is none.
func (l *textLabel) linkAt(where geom.Point) string {
	for _, link := range l.links() {
		if where.In(link.rect) {
			return link.ref
		}
	}
	return ""
}

// mouseDown remembers the link the press landed on, if any, so that mouseUp can tell a click from a drag off it.
func (l *textLabel) mouseDown(where geom.Point, button, _ int, _ mod.Modifiers) bool {
	if button != unison.ButtonLeft {
		return false
	}
	l.pressedRef = l.linkAt(where)
	return l.pressedRef != ""
}

// mouseUp opens the link the press landed on, provided the release is still over it, as a unison.Link does.
func (l *textLabel) mouseUp(where geom.Point, button int, _ mod.Modifiers) bool {
	if button != unison.ButtonLeft {
		return false
	}
	ref := l.pressedRef
	l.pressedRef = ""
	if ref == "" {
		return false
	}
	if l.linkAt(where) == ref {
		unison.SafeCall(func() { l.linkHandler(ref) })
	}
	return true
}

func (l *textLabel) updateCursor(where geom.Point) *unison.Cursor {
	if l.linkAt(where) != "" {
		return unison.PointingCursor()
	}
	return unison.ArrowCursor()
}
