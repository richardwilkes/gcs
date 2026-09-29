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
	"cmp"
	"regexp"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
	"github.com/richardwilkes/unison/enums/role"
)

// defaultWrappingLabelWidth is the width a textLabel wraps to when its layout has not offered it one.
const defaultWrappingLabelWidth = 400

// focusRingRoom is the horizontal gap between the focus ring and what it surrounds, where the label's bounds allow.
const focusRingRoom = 2

// pageRefPattern matches the Basic Set page references the calculators' notes cite: "B" and the page for the Characters
// book, "BX" and the page for the Campaigns book (page 338 onward), whether in parentheses like "(BX377)" or as each
// half of a range like "BX430-BX431". Only the Basic Set's keys are matched: every other page reference key is also a
// plausible piece of ordinary text (C4, TL6), and the Basic Set is the only book the notes cite.
var pageRefPattern = regexp.MustCompile(`\bBX?\d{1,3}\b`)

// textLabel shows read-only text whose page references, once linkPageRefs or linkRefs has been called, are links that
// can be followed with the mouse, the keyboard, or a screen reader, which a unison.Label cannot do. Made with
// newWrappingLabel it wraps to the width it is given: within a FlexLayout it must then be given HAlign: align.Fill,
// since that is what makes the layout offer it the column's width to wrap to, and until it has been offered one it
// wraps to defaultWrappingLabelWidth. Made with newSingleLineLabel it stays on one line, as a unison.Label does, so it
// can stand in for one beside a field or a popup.
type textLabel struct {
	unison.Panel
	text string
	ink  unison.Ink
	font unison.Font
	// linkFont is the font for the links, or nil for font.
	linkFont unison.Font
	// linkFinder returns the [start, end) byte offsets of the links in a piece of the text, in order. When it is nil,
	// the links are what pageRefPattern matches.
	linkFinder  func(text string) [][]int
	linkHandler func(ref string)
	pressedRef  string
	// current is the index of the link the keyboard is on, or -1 for the text as a whole.
	current int
	wrap    bool
	// hint marks a label a screen reader hears as the description of the control it follows (see
	// describeWithTrailingLabel).
	hint bool
}

// linkKey is the key of a link's virtual accessibility node: the link's index within the label.
type linkKey int

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
		ink:     unison.DefaultLabelTheme.OnBackgroundInk,
		font:    unison.DefaultLabelTheme.Font,
		current: -1,
		wrap:    wrap,
	}
	l.Self = l
	l.SetSizer(l.sizes)
	l.DrawCallback = l.draw
	// The text is drawn by hand, so a screen reader is told outright that this is a label and what it says; declaring
	// the role also lets it name the control laid out after it, as a unison.Label would.
	l.Accessibility.Role = role.Label
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
	l.Accessibility.Name = text
	l.syncForLinks()
	l.MarkForLayoutRecursivelyUpward()
	l.MarkForRedraw()
}

// syncForLinks makes the label a tab stop while its text holds links, with or without a screen reader, as a unison link
// is. A hint holding no links is left out of the accessibility tree, since it is heard with its control.
func (l *textLabel) syncForLinks() {
	hasLinks := l.linkHandler != nil && len(l.findLinks(l.text)) > 0
	l.SetFocusable(hasLinks)
	if l.hint {
		if hasLinks {
			l.Accessibility.Role = role.Label
		} else {
			l.Accessibility.Role = role.None
		}
	}
}

// describeWithTrailingLabel makes the label's text part of the control's accessibility description (see
// describeWithHint) and marks the label as a hint (see syncForLinks).
func describeWithTrailingLabel(control unison.Paneler, label *textLabel) {
	label.hint = true
	label.syncForLinks()
	describeWithHint(control, label.String)
}

// linkPageRefs makes the page references in the text into links, drawn in the link theme, that pass the reference to
// the handler when clicked.
//
// Each link is a keyboard stop of its own (see also textIsStop): Tab and Shift-Tab go from link to link and leave the
// label past either end, the arrows, Home and End move among the links, and Space or Enter follows the current one. A
// screen reader sees each link as an element within the text (see ProvideAccessibility).
func (l *textLabel) linkPageRefs(handler func(ref string)) {
	l.linkFinder = nil
	l.installLinks(handler)
}

// linkRefs is linkPageRefs for the given references, which may be to any book, rather than for the Basic Set references
// pageRefPattern finds.
func (l *textLabel) linkRefs(handler func(ref string), refs ...string) {
	l.linkFinder = func(text string) [][]int { return findRefs(text, refs) }
	l.installLinks(handler)
}

// findRefs returns the [start, end) byte offsets of the whole-word occurrences of the references in the text, in order.
// Of overlapping occurrences, the one starting first is kept, and the longest of those starting together.
func findRefs(text string, refs []string) [][]int {
	var found [][]int
	for _, ref := range refs {
		if ref == "" {
			continue
		}
		for from := 0; from < len(text); {
			i := strings.Index(text[from:], ref)
			if i < 0 {
				break
			}
			start := from + i
			end := start + len(ref)
			before, _ := utf8.DecodeLastRuneInString(text[:start])
			after, _ := utf8.DecodeRuneInString(text[end:])
			if !isWordRune(before) && !isWordRune(after) {
				found = append(found, []int{start, end})
			}
			from = end
		}
	}
	slices.SortFunc(found, func(a, b []int) int {
		if result := cmp.Compare(a[0], b[0]); result != 0 {
			return result
		}
		return cmp.Compare(b[1], a[1])
	})
	kept := found[:0]
	last := 0
	for _, one := range found {
		if one[0] >= last {
			kept = append(kept, one)
			last = one[1]
		}
	}
	return kept
}

// isWordRune reports whether r is a letter or digit; utf8.RuneError, returned at either end of the text, is not.
func isWordRune(r rune) bool {
	return r != utf8.RuneError && (unicode.IsLetter(r) || unicode.IsDigit(r))
}

// findLinks returns the [start, end) byte offsets of the links in a piece of the text, or nil when there is no handler.
func (l *textLabel) findLinks(text string) [][]int {
	switch {
	case l.linkHandler == nil:
		return nil
	case l.linkFinder != nil:
		return l.linkFinder(text)
	default:
		return pageRefPattern.FindAllStringIndex(text, -1)
	}
}

func (l *textLabel) installLinks(handler func(ref string)) {
	l.linkHandler = handler
	l.MouseDownCallback = l.mouseDown
	l.MouseUpCallback = l.mouseUp
	l.UpdateCursorCallback = l.updateCursor
	l.KeyDownCallback = l.keyDown
	l.GainedFocusCallback = l.gainedFocus
	l.LostFocusCallback = l.leaveLinks
	l.syncForLinks()
}

// textIsStop reports whether the text as a whole is a keyboard stop ahead of the links, which it is while accessibility
// is active and unison.FocusForReading is on, so that the text is heard before its links.
func (l *textLabel) textIsStop() bool {
	return unison.IsAccessibilityActive() && unison.FocusForReading()
}

func (l *textLabel) firstStop(links []pageRefLink) int {
	if len(links) == 0 || l.textIsStop() {
		return -1
	}
	return 0
}

// gainedFocus puts the keyboard on the first stop, or on the last link when Shift is down, as for Shift-Tab, so that
// the stops are met in order from either side.
func (l *textLabel) gainedFocus() {
	l.ScrollIntoView()
	links := l.links()
	l.current = l.firstStop(links)
	if wnd := l.Window(); wnd != nil && len(links) > 0 && wnd.LastKeyModifiers().ShiftDown() {
		l.current = len(links) - 1
	}
	l.MarkForRedraw()
}

func (l *textLabel) leaveLinks() {
	l.current = -1
	l.MarkForRedraw()
}

// keyDown moves among the stops and follows the current link. Tab and Shift-Tab past either end are left to the window,
// which moves the focus out of the label.
func (l *textLabel) keyDown(keyCode unison.KeyCode, mods mod.Modifiers, _ bool) bool {
	links := l.links()
	if len(links) == 0 {
		return false
	}
	first := l.firstStop(links)
	last := len(links) - 1
	current := min(max(l.current, first), last)
	if keyCode == unison.KeyTab {
		if mods&(mod.NonSticky&^mod.Shift) != 0 {
			return false
		}
		if mods.ShiftDown() {
			current--
		} else {
			current++
		}
		if current < first || current > last {
			return false
		}
		l.moveToLink(links, current)
		return true
	}
	if mods&mod.NonSticky != 0 {
		return false
	}
	switch keyCode {
	case unison.KeyRight, unison.KeyDown:
		l.moveToLink(links, min(current+1, last))
	case unison.KeyLeft, unison.KeyUp:
		l.moveToLink(links, max(current-1, first))
	case unison.KeyHome:
		l.moveToLink(links, 0)
	case unison.KeyEnd:
		l.moveToLink(links, last)
	case unison.KeySpace, unison.KeyReturn, unison.KeyNumPadEnter:
		switch {
		case current >= 0:
			l.follow(links[current].ref)
		case len(links) == 1:
			l.follow(links[0].ref)
		}
	default:
		return false
	}
	return true
}

// moveToLink makes the link at the index current, or the text as a whole for -1. The redraw also republishes the
// window's accessibility description, which is how a screen reader learns of the move.
func (l *textLabel) moveToLink(links []pageRefLink, index int) {
	if index >= 0 && index < len(links) {
		l.ScrollRectIntoView(links[index].rect)
	}
	if index != l.current {
		l.current = index
		l.MarkForRedraw()
	}
}

func (l *textLabel) follow(ref string) {
	unison.SafeCall(func() { l.linkHandler(ref) })
}

// ProvideAccessibility implements unison.AccessibilityProvider, describing each link as a child element that holds the
// focus while the keyboard is on it.
func (l *textLabel) ProvideAccessibility(b *unison.AccessibilityBuilder) {
	focusable := l.Focusable()
	for i, link := range l.links() {
		id := b.AddVirtualChild(linkKey(i), func(n *accessibility.Node) {
			n.Role = role.Link
			n.Name = link.ref
			n.Bounds = link.rect
			n.Actions = n.Actions.With(accessibility.Press)
			if focusable {
				n.Focusable = true
				n.Actions = n.Actions.With(accessibility.Focus)
			}
		})
		if i == l.current && b.Focused() {
			b.FocusChild(id)
		}
	}
}

// PerformAccessibilityAction implements unison.AccessibilityActor: Press follows a link, Focus puts the keyboard on it.
func (l *textLabel) PerformAccessibilityAction(req accessibility.ActionRequest) bool {
	index, ok := req.Key.(linkKey)
	if !ok {
		return false
	}
	links := l.links()
	if int(index) < 0 || int(index) >= len(links) {
		return false
	}
	switch req.Action {
	case accessibility.Press:
		l.follow(links[index].ref)
		return true
	case accessibility.Focus:
		if !l.Focusable() {
			return false
		}
		l.RequestFocus()
		if wnd := l.Window(); wnd == nil || wnd.Focus() != l.AsPanel() {
			return false
		}
		l.moveToLink(links, int(index))
		return true
	default:
		return false
	}
}

// paragraphs returns the text as one unison.Text per logical line, with the page references decorated as links when
// they are clickable.
func (l *textLabel) paragraphs() []*unison.Text {
	base := &unison.TextDecoration{Font: l.font, OnBackgroundInk: l.ink}
	linkFont := l.linkFont
	if linkFont == nil {
		linkFont = l.font
	}
	link := &unison.TextDecoration{Font: linkFont, OnBackgroundInk: unison.DefaultLinkTheme.OnBackgroundInk, Underline: true}
	split := strings.Split(l.text, "\n")
	paragraphs := make([]*unison.Text, 0, len(split))
	for _, para := range split {
		text := unison.NewText("", base)
		last := 0
		for _, m := range l.findLinks(para) {
			text.AddString(para[last:m[0]], base)
			text.AddString(para[m[0]:m[1]], link)
			last = m[1]
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
	var width float32
	for _, line := range l.lines(rect.Width) {
		line.Draw(gc, geom.NewPoint(rect.X, y+line.Baseline()))
		y += line.Height()
		width = max(width, line.Width())
	}
	// The focus ring surrounds the current link, else the text, which may be narrower than the label.
	if l.Focused() {
		ring := geom.NewRect(rect.X, rect.Y, width, y-rect.Y)
		if links := l.links(); l.current >= 0 && l.current < len(links) {
			ring = links[l.current].rect
		}
		ring.X -= focusRingRoom
		ring.Width += focusRingRoom * 2
		ring = ring.Intersect(l.ContentRect(true)).Inset(geom.NewUniformInsets(0.5))
		gc.DrawRect(ring, unison.ThemeFocus.Paint(gc, ring, paintstyle.Stroke))
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
		for _, m := range l.findLinks(s) {
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
		l.follow(ref)
	}
	return true
}

func (l *textLabel) updateCursor(where geom.Point) *unison.Cursor {
	if l.linkAt(where) != "" {
		return unison.PointingCursor()
	}
	return unison.ArrowCursor()
}
