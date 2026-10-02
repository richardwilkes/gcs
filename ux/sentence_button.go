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

// stripEm removes the markers em adds.
var stripEm = strings.NewReplacer(string(emStart), "", string(emEnd), "")

// em wraps s in the markers that have a sentenceButton draw it in bold.
func em(s string) string {
	return string(emStart) + s + string(emEnd)
}

// sentenceButton draws a sentence whose spans wrapped by em are bold, wrapping it to the width it is given; within a
// FlexLayout it must be given HAlign: align.Fill for that. With a click handler it is a disclosure that Space, Enter or a
// click activates. Without one it is static text, a tab stop only while reading (see tabStopForReading).
type sentenceButton struct {
	unison.Panel
	text     string
	suffix   string
	onClick  func()
	expanded func() bool
	hovered  bool
}

// newSentenceButton returns a sentenceButton. onClick may be nil for static text; expanded reports the state a screen
// reader is told of when it is not.
func newSentenceButton(text string, onClick func(), expanded func() bool) *sentenceButton {
	b := &sentenceButton{onClick: onClick, expanded: expanded}
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
	b.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
		if mods&mod.NonSticky != 0 ||
			(keyCode != unison.KeySpace && keyCode != unison.KeyReturn && keyCode != unison.KeyNumPadEnter) {
			return false
		}
		b.onClick()
		return true
	}
	addAccessibilityCallback(b, func(node *accessibility.Node) {
		node.Expandable = true
		node.Expanded = b.expanded != nil && b.expanded()
		node.Actions = node.Actions.With(accessibility.Press, accessibility.Expand, accessibility.Collapse)
	})
	b.Accessibility.ActionCallback = func(req accessibility.ActionRequest) bool {
		switch req.Action {
		case accessibility.Press:
		case accessibility.Expand, accessibility.Collapse:
			if (b.expanded != nil && b.expanded()) == (req.Action == accessibility.Expand) {
				return true
			}
		default:
			return false
		}
		b.onClick()
		return true
	}
	return b
}

// setText replaces the sentence and the suffix a screen reader hears after it, such as a status.
func (b *sentenceButton) setText(text, suffix string) {
	b.text = text
	b.suffix = suffix
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

func (b *sentenceButton) lines(width float32) []*unison.Text {
	plain := &unison.TextDecoration{
		Font:            unison.DefaultLabelTheme.Font,
		OnBackgroundInk: unison.DefaultLabelTheme.OnBackgroundInk,
	}
	bold := *plain
	desc := plain.Font.Descriptor()
	desc.Weight = weight.Bold
	bold.Font = desc.Font()
	text := unison.NewText("", plain)
	decoration := plain
	for rest := b.text; rest != ""; {
		i := strings.IndexAny(rest, string([]rune{emStart, emEnd}))
		if i < 0 {
			text.AddString(rest, decoration)
			break
		}
		text.AddString(rest[:i], decoration)
		r, size := utf8.DecodeRuneInString(rest[i:])
		decoration = plain
		if r == emStart {
			decoration = &bold
		}
		rest = rest[i+size:]
	}
	if width <= 0 {
		width = defaultWrappingLabelWidth
	}
	return text.BreakToWidth(width)
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
	radius := geom.NewUniformSize(4)
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
