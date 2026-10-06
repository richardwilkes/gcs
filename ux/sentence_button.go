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
	"strings"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
	"github.com/richardwilkes/unison/enums/weight"
)

// The private-use runes that open and close a bold span of a sentenceButton's text.
const (
	emStart = ''
	emEnd   = ''
)

// stripEm removes the markers emphasize adds.
var stripEm = strings.NewReplacer(string(emStart), "", string(emEnd), "")

// emphasize wraps s in the markers that have a sentenceButton draw it in bold.
func emphasize(s string) string {
	return string(emStart) + s + string(emEnd)
}

// sentenceButton draws a sentence whose spans wrapped by emphasize are bold, breaking it at its line breaks and wrapping
// it to the width it is given; within a FlexLayout it must be given HAlign: align.Fill for that. With a click handler it
// is a disclosure that Space, Enter or a click activates, always reported as collapsed, since what it discloses takes its
// place. Without one it is static text, a tab stop only while reading (see tabStopForReading).
type sentenceButton struct {
	unison.Panel
	text    string
	onClick func()
	hovered bool
	// The sentence's paragraphs as styled text, built with builtFont, and broken into lines for wrappedWidth, since
	// laying out and drawing ask for them often.
	built        []*unison.Text
	builtFont    unison.FontDescriptor
	wrapped      []*unison.Text
	wrappedWidth float32
}

// newSentenceButton returns a sentenceButton. onClick may be nil for static text.
func newSentenceButton(text string, onClick func()) *sentenceButton {
	b := &sentenceButton{onClick: onClick}
	b.Self = b
	b.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: 2, Left: 4, Bottom: 2, Right: 4}))
	b.SetSizer(b.sizes)
	b.DrawCallback = b.draw
	b.setText(text, "")
	if onClick == nil {
		b.Accessibility.Role = role.Label
		tabStopForReading(b)
		return b
	}
	b.SetFocusable(true)
	b.Accessibility.Role = role.DisclosureTriangle
	b.MouseDownCallback = func(_ geom.Point, _, _ int, _ mod.Modifiers) bool { return true }
	b.MouseUpCallback = func(where geom.Point, _ int, _ mod.Modifiers) bool {
		if where.In(b.ContentRect(true)) {
			b.onClick()
		}
		return true
	}
	b.MouseEnterCallback = func(_ geom.Point, _ mod.Modifiers) bool { return b.hover(true) }
	b.MouseExitCallback = func() bool { return b.hover(false) }
	b.UpdateCursorCallback = func(_ geom.Point) *unison.Cursor { return unison.PointingCursor() }
	b.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if !noModifiersDown(mods) ||
			(keyCode != unison.KeySpace && keyCode != unison.KeyReturn && keyCode != unison.KeyNumPadEnter) {
			return false
		}
		// A held key acts once; its repeats are taken, so that they don't reach what holds the sentence.
		if !repeat {
			b.onClick()
		}
		return true
	}
	addAccessibilityCallback(b, func(node *accessibility.Node) {
		node.Expandable = true
		node.Actions = node.Actions.With(accessibility.Press, accessibility.Expand, accessibility.Collapse)
	})
	b.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		switch req.Action {
		case accessibility.Press, accessibility.Expand:
			b.onClick()
			return true
		case accessibility.Collapse:
			return true
		default:
			return false
		}
	}
	return b
}

// setText replaces the sentence and the suffix a screen reader hears after it, such as a status.
func (b *sentenceButton) setText(text, suffix string) {
	b.text = text
	b.built = nil
	b.Accessibility.Name = b.plainText()
	if suffix != "" {
		b.Accessibility.Name += i18n.Text(", ") + suffix
	}
	b.MarkForLayoutRecursivelyUpward()
	b.MarkForRedraw()
}

// plainText returns the sentence without its bold markers.
func (b *sentenceButton) plainText() string {
	return stripEm.Replace(b.text)
}

func (b *sentenceButton) hover(on bool) bool {
	b.hovered = on
	b.MarkForRedraw()
	return true
}

// lines returns the sentence broken into lines no wider than width, rebuilding it only when the text, the font or the
// width has changed.
func (b *sentenceButton) lines(width float32) []*unison.Text {
	if width <= 0 {
		width = defaultWrappingLabelWidth
	}
	if font := unison.DefaultLabelTheme.Font.Descriptor(); b.built == nil || b.builtFont != font {
		b.built = b.buildText()
		b.builtFont = font
		b.wrapped = nil
	}
	if b.wrapped == nil || b.wrappedWidth != width {
		b.wrapped = nil
		for _, paragraph := range b.built {
			b.wrapped = append(b.wrapped, paragraph.BreakToWidth(width)...)
		}
		b.wrappedWidth = width
	}
	return b.wrapped
}

// buildText returns the paragraphs of the sentence, which line breaks separate, as styled text, its emphasized spans in
// bold.
func (b *sentenceButton) buildText() []*unison.Text {
	plain := &unison.TextDecoration{
		Font:            unison.DefaultLabelTheme.Font,
		OnBackgroundInk: unison.DefaultLabelTheme.OnBackgroundInk,
	}
	bold := *plain
	desc := plain.Font.Descriptor()
	desc.Weight = weight.Bold
	bold.Font = desc.Font()
	text := unison.NewText("", plain)
	paragraphs := []*unison.Text{text}
	decoration := plain
	for rest := b.text; rest != ""; {
		i := strings.IndexAny(rest, string([]rune{emStart, emEnd, '\n'}))
		if i < 0 {
			text.AddString(rest, decoration)
			break
		}
		text.AddString(rest[:i], decoration)
		r, size := utf8.DecodeRuneInString(rest[i:])
		switch r {
		case '\n':
			text = unison.NewText("", plain)
			paragraphs = append(paragraphs, text)
		case emStart:
			decoration = &bold
		default:
			decoration = plain
		}
		rest = rest[i+size:]
	}
	return paragraphs
}

// lineHeight returns the height of a line of the sentence, with the border above and below it.
func (b *sentenceButton) lineHeight() float32 {
	return unison.DefaultLabelTheme.Font.LineHeight() + b.Border().Insets().Height()
}

func (b *sentenceButton) sizes(hint geom.Size) (minSize, prefSize, maxSize geom.Size) {
	insets := b.Border().Insets()
	for _, line := range b.lines(hint.Width - insets.Width()) {
		prefSize.Width = max(prefSize.Width, line.Width())
		prefSize.Height += line.Height()
	}
	prefSize = prefSize.Add(insets.Size()).Ceil()
	return prefSize, prefSize, unison.MaxSize(prefSize)
}

func (b *sentenceButton) draw(gc *unison.Canvas, _ geom.Rect) {
	r := b.ContentRect(true)
	radius := geom.NewUniformSize(compactCornerRadius)
	if b.onClick != nil && b.hovered {
		gc.DrawRoundedRect(r, radius, unison.ThemeAboveSurface.Paint(gc, r, paintstyle.Fill))
	}
	if b.onClick != nil && b.Focused() {
		ring := r.Inset(geom.NewUniformInsets(0.5))
		gc.DrawRoundedRect(ring, radius, unison.ThemeFocus.Paint(gc, ring, paintstyle.Stroke))
	}
	rect := b.ContentRect(false)
	y := rect.Y
	for _, line := range b.lines(rect.Width) {
		line.Draw(gc, geom.NewPoint(rect.X, y+line.Baseline()))
		y += line.Height()
	}
}
