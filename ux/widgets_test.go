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
	"bytes"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

// renderTooltipMarkdown renders the given text using the same Markdown configuration that unison uses for tooltips
// (goldmark with the GFM extension), so the tests exercise the real rendering behavior rather than an approximation.
func renderTooltipMarkdown(c check.Checker, text string) string {
	var buffer bytes.Buffer
	c.NoError(goldmark.New(goldmark.WithExtensions(extension.GFM)).Convert([]byte(text), &buffer))
	return buffer.String()
}

func TestMarkdownHardLineBreaks(t *testing.T) {
	c := check.New(t)
	for _, d := range []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "single line", in: "+2 Foo", want: "+2 Foo"},
		{name: "two lines", in: "+2 Foo\n+3 Bar", want: "+2 Foo  \n+3 Bar"},
		{name: "three lines", in: "+2 Foo\n+3 Bar\n+1 Baz", want: "+2 Foo  \n+3 Bar  \n+1 Baz"},
		{name: "paragraph break", in: "a\n\nb", want: "a  \n  \nb"},
	} {
		c.Equal(d.want, markdownHardLineBreaks(d.in), d.name)
	}
}

// TestMarkdownHardLineBreaksRendering reproduces the conditional modifiers amount column tooltip bug: without the fix,
// the newline-separated bonuses collapse onto a single line because the Markdown renderer treats a single newline as a
// soft break. After the fix, each bonus renders on its own line.
func TestMarkdownHardLineBreaksRendering(t *testing.T) {
	c := check.New(t)
	tooltip := "+2 Beauty\n+3 Charisma\n+1 Voice"

	// Without the fix, the renderer produces no line breaks (all bonuses on one line).
	raw := renderTooltipMarkdown(c, tooltip)
	c.Equal(0, strings.Count(raw, "<br>"), "unconverted tooltip should collapse onto one line")

	// With the fix, each newline becomes a hard line break, so there is one break between each of the three bonuses.
	fixed := renderTooltipMarkdown(c, markdownHardLineBreaks(tooltip))
	c.Equal(2, strings.Count(fixed, "<br>"), "converted tooltip should render one bonus per line")
}

// TestMarkdownHardLineBreaksPreservesParagraphs verifies that blank lines (paragraph breaks) and other block constructs
// continue to render correctly after conversion, so the fix does not disturb tooltips that already rely on Markdown.
func TestMarkdownHardLineBreaksPreservesParagraphs(t *testing.T) {
	c := check.New(t)
	rendered := renderTooltipMarkdown(c, markdownHardLineBreaks("first\n\nsecond"))
	c.Equal(2, strings.Count(rendered, "<p>"), "paragraph break should be preserved as two paragraphs")
	c.Equal(0, strings.Count(rendered, "<br>"), "a lone paragraph break should not introduce a hard line break")
}

// TestNewApplyCancelButtons verifies the behavior every editor with pending changes relies on: both buttons start out
// disabled, apply closes the editor only when it reports success, cancel closes it without applying, and the keyboard
// shortcuts are shown only when asked for.
func TestNewApplyCancelButtons(t *testing.T) {
	c := check.New(t)
	toolbar := unison.NewPanel()
	applied, closed := 0, 0
	succeed := true
	applyButton, cancelButton := newApplyCancelButtons(toolbar, false,
		func() bool {
			applied++
			return succeed
		},
		func() { closed++ })
	c.Equal([]*unison.Panel{applyButton.AsPanel(), cancelButton.AsPanel()}, toolbar.Children(),
		"apply then cancel are added to the toolbar")
	c.False(applyButton.Enabled(), "apply starts out disabled")
	c.False(cancelButton.Enabled(), "cancel starts out disabled")

	applyButton.ClickCallback()
	c.Equal(1, applied)
	c.Equal(1, closed, "a successful apply closes the editor")

	succeed = false
	applyButton.ClickCallback()
	c.Equal(2, applied)
	c.Equal(1, closed, "an apply that fails leaves the editor open")

	cancelButton.ClickCallback()
	c.Equal(2, applied, "cancel applies nothing")
	c.Equal(2, closed, "and closes the editor")

	c.Equal(1, len(applyButton.Tooltip.Children()), "without shortcuts, the tooltip is a single line")
	applyButton, cancelButton = newApplyCancelButtons(unison.NewPanel(), true, func() bool { return true }, func() {})
	c.Equal(2, len(applyButton.Tooltip.Children()), "with shortcuts, the tooltip gains a second line")
	c.Equal(2, len(cancelButton.Tooltip.Children()))
}

// TestNewPopupMenuOffersItemsAndReportsChoices verifies that newPopupMenu offers the items in order, starts out showing
// the current value without reporting it as a choice, and hands each later selection to onSelect. Every settings popup
// is built through this helper, so a callback that fired during construction would clobber settings on open.
func TestNewPopupMenuOffersItemsAndReportsChoices(t *testing.T) {
	c := check.New(t)
	var chosen []string
	popup := newPopupMenu([]string{"one", "two", "three"}, "two", func(item string) { chosen = append(chosen, item) })
	c.Equal(3, popup.ItemCount(), "all items should be offered")
	for i, want := range []string{"one", "two", "three"} {
		item, ok := popup.ItemAt(i)
		c.True(ok, "item %d should exist", i)
		c.Equal(want, item, "items should keep their order")
	}
	selected, ok := popup.Selected()
	c.True(ok, "the current value should be selected")
	c.Equal("two", selected, "the current value should be selected")
	c.Nil(chosen, "building the popup must not report the initial selection")

	popup.Select("three")
	c.Equal([]string{"three"}, chosen, "selecting an item should report it")
	popup.Select("three")
	c.Equal([]string{"three"}, chosen, "re-selecting the same item must not report it again")
	popup.Select("one")
	c.Equal([]string{"three", "one"}, chosen, "each new selection should be reported")
}

// TestNewPopupMenuWithUnknownCurrentValue verifies that a current value that is not among the items leaves the popup
// with no selection rather than picking something arbitrary, and that onSelect is not called until the user chooses.
func TestNewPopupMenuWithUnknownCurrentValue(t *testing.T) {
	c := check.New(t)
	calls := 0
	popup := newPopupMenu([]int{1, 2, 3}, 42, func(int) { calls++ })
	_, ok := popup.Selected()
	c.False(ok, "an unknown current value should leave nothing selected")
	c.Equal(0, calls, "an unknown current value must not be reported")
	popup.SelectIndex(1)
	c.Equal(1, calls, "choosing an item should be reported")
}

// TestInstallPopupSelectionOnHandBuiltMenu verifies the wiring the calendar popup uses on a menu it fills itself, with
// disabled library headings among the items: the current value is selected and later choices are reported.
func TestInstallPopupSelectionOnHandBuiltMenu(t *testing.T) {
	c := check.New(t)
	popup := unison.NewPopupMenu[string]()
	popup.AddDisabledItem("Library")
	popup.AddItem("Gregorian", "Imperial")
	var got string
	installPopupSelection(popup, "Imperial", func(item string) { got = item })
	selected, ok := popup.Selected()
	c.True(ok, "the current value should be selected")
	c.Equal("Imperial", selected, "the current value should be selected")
	c.Equal("", got, "installing the selection must not report it")

	popup.Select("Gregorian")
	c.Equal("Gregorian", got, "choosing an enabled item should be reported")
}
