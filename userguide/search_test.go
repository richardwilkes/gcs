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
// order. A place is listed once, at its best rank, and a body match points at the heading of its section. A page with
// no level-1 heading is titled by its file name, and empty text matches nothing.
func TestSearchRanksTitlesThenHeadingsThenBody(t *testing.T) {
	c := check.New(t)
	fsys := fstest.MapFS{
		"Alpha.md":    {Data: []byte("# Alpha\n\nNothing here.\n\n## Lantern Care\n\nKeep the lantern lit.\n\n## Fuel\n\nA lantern needs oil.\n")},
		"Beta.md":     {Data: []byte("# The Lantern\n\nThe lantern is bright.\n\n## Wicks\n\nTrim the lantern wick.\n")},
		"Delta.md":    {Data: []byte("# Delta\n\nA lantern above any heading.\n\n## Other\n\nNothing.\n")},
		"Gamma.md":    {Data: []byte("# Gamma\n\n## Lantern\n\nThe lantern again.\n")},
		"Lanterns.md": {Data: []byte("Sold by the dozen.\n")},
	}
	idx := buildIndex(fsys, []string{"Beta.md", "Alpha.md"})
	want := []Match{
		{Page: "Beta.md", Text: "The Lantern", Kind: TitleMatch},
		{Page: "Lanterns.md", Text: "Lanterns", Kind: TitleMatch},
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
	c.Equal(0, len(idx.search("", 0)), "match count for empty text")
}

// A body match in a section longer than about 80 characters shows the text around the match, cut at spaces where it can
// be, with "…" in place of what was cut.
func TestSearchShowsTheTextAroundABodyMatch(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name  string
		text  string
		query string
		want  string
	}{
		{
			name:  "match near the start",
			text:  "The lantern hangs by the door of the old mill, where the keeper trims its wick each evening.",
			query: "lantern",
			want:  "The lantern hangs by the door of the old mill, where the keeper trims its wick…",
		},
		{
			name:  "match near the end",
			text:  "Each evening the keeper of the old mill climbs the stair and trims the wick of the lantern.",
			query: "lantern",
			want:  "…the keeper of the old mill climbs the stair and trims the wick of the lantern.",
		},
		{
			name:  "match in the middle",
			text:  "At dusk the keeper of the old mill lights the Lantern by the door and trims its wick as it turns.",
			query: "lantern",
			want:  "…keeper of the old mill lights the Lantern by the door and trims its wick as…",
		},
		{
			name:  "multi-byte characters ahead of the match",
			text:  "Más allá del étang helado y los piñones, a lantern marks the door of the old mill by the río.",
			query: "lantern",
			want:  "…del étang helado y los piñones, a lantern marks the door of the old mill by…",
		},
		{
			name:  "nowhere to break",
			text:  "Library/Lantern/Settings/Sheets/Templates/Campaigns/Characters/Equipment/Modifiers/Notes/Spells",
			query: "lantern",
			want:  "Library/Lantern/Settings/Sheets/Templates/Campaigns/Characters/Equipment/Modifie…",
		},
		{
			name:  "match longer than a snippet",
			text:  "Then the keeper of the old mill climbs the stair each evening to trim the wick of the lantern too.",
			query: "the keeper of the old mill climbs the stair each evening to trim the wick of the lantern",
			want:  "…the keeper of the old mill climbs the stair each evening to trim the wick of the lantern…",
		},
	} {
		var got string
		idx := buildIndex(fstest.MapFS{"Page.md": {Data: []byte(tc.text + "\n")}}, nil)
		if matches := idx.search(tc.query, 0); len(matches) == 1 {
			got = matches[0].Text
		}
		c.Equal(tc.want, got, tc.name)
	}
}

// Text is indexed as unison.Markdown shows it: a <br> or <hr> breaks the line, separating the words on either side, any
// other tag shows nothing, and code is text like any other, while an alert's marker and an image's alt text are not.
func TestSearchIndexesTextAsShown(t *testing.T) {
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
		{name: "fenced code", src: "Run:\n\n```\ngcs --help\n```\n", want: "Run: gcs --help"},
		{name: "alert marker", src: "> [!NOTE]\n> Save often.\n", want: "Save often."},
		{name: "image alt text", src: "See ![the dialog](images/dialog.png) here.\n", want: "See here."},
	} {
		var got string
		if p := indexPage(md, "Page.md", []byte(tc.src)); len(p.sections) != 0 {
			got = p.sections[0].text
		}
		c.Equal(tc.want, got, tc.name)
	}
}
