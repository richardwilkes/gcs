// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package userguide

import (
	"fmt"
	"testing"
	"testing/fstest"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// Title matches come first, then heading matches, then body matches, each in table of contents order and then document
// order. A place is listed once, at its best rank, and a body match points at the heading of its section.
func TestSearchRanksTitlesThenHeadingsThenBody(t *testing.T) {
	c := check.New(t)
	fsys := fstest.MapFS{
		"Alpha.md": {Data: []byte("# Alpha\n\nNothing here.\n\n## Lantern Care\n\nKeep the lantern lit.\n\n## Fuel\n\nA lantern needs oil.\n")},
		"Beta.md":  {Data: []byte("# The Lantern\n\nThe lantern is bright.\n\n## Wicks\n\nTrim the lantern wick.\n")},
		"Delta.md": {Data: []byte("# Delta\n\nA lantern above any heading.\n\n## Other\n\nNothing.\n")},
		"Gamma.md": {Data: []byte("# Gamma\n\n## Lantern\n\nThe lantern again.\n")},
	}
	idx := buildIndex(fsys, []string{"Beta.md", "Alpha.md"})
	want := []Match{
		{Page: "Beta.md", Text: "The Lantern", Kind: TitleMatch},
		{Page: "Alpha.md", Anchor: "lantern-care", Text: "Lantern Care", Section: "Lantern Care", Kind: HeadingMatch},
		{Page: "Gamma.md", Anchor: "lantern", Text: "Lantern", Section: "Lantern", Kind: HeadingMatch},
		{Page: "Beta.md", Anchor: "wicks", Text: "Trim the lantern wick.", Section: "Wicks", Kind: BodyMatch},
		{Page: "Alpha.md", Anchor: "fuel", Text: "A lantern needs oil.", Section: "Fuel", Kind: BodyMatch},
		{Page: "Delta.md", Text: "A lantern above any heading.", Kind: BodyMatch},
	}
	for _, limit := range []int{0, len(want) + 1, 3} {
		expected := want
		if limit > 0 && limit < len(want) {
			expected = want[:limit]
		}
		got := idx.search("lantern", limit)
		c.Equal(len(expected), len(got), "match count for limit %d", limit)
		for i := range min(len(expected), len(got)) {
			c.Equal(fmt.Sprintf("%+v", expected[i]), fmt.Sprintf("%+v", got[i]), "match %d for limit %d", i, limit)
		}
	}
}

// Inline HTML is indexed as unison.Markdown shows it: a <br> or <hr> breaks the line, separating the words on either
// side, and any other tag shows nothing.
func TestSearchIndexesInlineHTMLAsShown(t *testing.T) {
	c := check.New(t)
	md := unison.NewMarkdownParser()
	for _, tc := range []struct {
		name string
		src  string
		want string
	}{
		{name: "br", src: "| Syntax |\n|---|\n| `- item`<br>`- item` |\n", want: "Syntax - item - item"},
		{name: "hr", src: "one<HR />two\n", want: "one two"},
		{name: "other tag", src: "one<span>two</span>\n", want: "onetwo"},
	} {
		var got string
		if p := indexPage(md, "Page.md", []byte(tc.src)); len(p.sections) != 0 {
			got = p.sections[0].text
		}
		c.Equal(tc.want, got, tc.name)
	}
}
