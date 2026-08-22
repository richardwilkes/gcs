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
