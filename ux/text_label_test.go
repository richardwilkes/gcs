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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestWrappingLabelSizes verifies the sizing that makes the label wrap: a hint with no usable width wraps the text to
// the default width, a wider hint wraps it onto fewer lines and so asks for less height than a narrower one, the label
// never asks to be smaller than what it wrapped to, and a border's insets are added around the text.
func TestWrappingLabelSizes(t *testing.T) {
	c := check.New(t)
	l := newWrappingLabel()
	l.setText(strings.Repeat("lorem ipsum dolor sit amet ", 40), unison.DefaultLabelTheme.OnBackgroundInk)

	atDefault := len(l.lines(defaultWrappingLabelWidth))
	c.True(atDefault > 1, "the text is long enough to wrap at the default width, but takes %d line", atDefault)
	c.Equal(atDefault, len(l.lines(0)), "no width wraps to the default width")
	c.Equal(atDefault, len(l.lines(-50)), "as does a negative one")
	_, defaultPref, _ := l.sizes(geom.NewSize(defaultWrappingLabelWidth, 0))
	_, unhintedPref, _ := l.sizes(geom.Size{})
	c.Equal(defaultPref, unhintedPref, "a hint with no width sizes as the default width does")
	_, negativePref, _ := l.sizes(geom.NewSize(-50, 0))
	c.Equal(defaultPref, negativePref, "as does one with a negative width")
	c.True(defaultPref.Width <= defaultWrappingLabelWidth, "the wrapped text fits the default width")

	narrowMin, narrowPref, narrowMax := l.sizes(geom.NewSize(200, 0))
	_, widePref, _ := l.sizes(geom.NewSize(800, 0))
	c.True(len(l.lines(800)) < len(l.lines(200)), "a wider hint wraps onto fewer lines")
	c.True(widePref.Height < narrowPref.Height, "and so asks for less height")
	c.True(narrowPref.Width <= 200, "the wrapped text fits the hint")
	c.True(widePref.Width > narrowPref.Width, "and uses the width it is given")
	c.Equal(narrowPref, narrowMin, "the label is never shrunk below what it wrapped to")
	c.Equal(unison.MaxSize(narrowPref), narrowMax, "and may grow as far as any panel may")
	c.Equal(narrowPref, narrowPref.Ceil(), "the size is rounded up to whole pixels")

	// With a border, the text wraps to the width left inside the insets, so the hint that leaves 200 for the text must
	// wrap exactly as a bare hint of 200 did, and the insets are added back around the result.
	insets := geom.Insets{Top: 3, Left: 5, Bottom: 7, Right: 11}
	l.SetBorder(unison.NewEmptyBorder(insets))
	_, borderedPref, _ := l.sizes(geom.NewSize(200+insets.Width(), 0))
	c.Equal(narrowPref.Width+insets.Width(), borderedPref.Width, "the border's insets are added to the width")
	c.Equal(narrowPref.Height+insets.Height(), borderedPref.Height, "and to the height")
	_, borderedUnhintedPref, _ := l.sizes(geom.Size{})
	c.Equal(defaultPref.Add(insets.Size()), borderedUnhintedPref,
		"a hint with no width wraps to the default width regardless of the insets, which are still added")
}

// TestWrappingLabelSetText verifies that setText replaces the text and ink and asks the label and its ancestors to be
// laid out again, since the number of lines may have changed.
func TestWrappingLabelSetText(t *testing.T) {
	c := check.New(t)
	parent := unison.NewPanel()
	l := newWrappingLabel()
	parent.AddChild(l)
	c.Equal("", l.text, "a new label is empty")
	c.Equal(unison.DefaultLabelTheme.OnBackgroundInk, l.ink, "and drawn in the label theme's ink")
	parent.ValidateLayout()
	c.False(l.NeedsLayout, "laying the label out leaves it with no layout pending")
	c.False(parent.NeedsLayout)

	l.setText("Hello", unison.ThemeError)
	c.Equal("Hello", l.text)
	c.Equal(unison.ThemeError, l.ink)
	lines := l.lines(0)
	c.Equal(1, len(lines), "the new text is what the label wraps")
	c.Equal("Hello", lines[0].String())
	c.True(l.NeedsLayout, "the label asks to be laid out again")
	c.True(parent.NeedsLayout, "and so does its parent, whose size follows the label's line count")
}

// TestPageRefPattern pins which pieces of a note become links: Basic Set references, whether in parentheses, followed
// by punctuation or paired in a range, and nothing that merely resembles one, such as an explosive's name or a tech
// level, since a false link would open the wrong book.
func TestPageRefPattern(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		text string
		want []string
	}{
		{text: "see Dodge and Drop (BX377).", want: []string{"BX377"}},
		{text: "Collisions & Falls (BX430-BX431)", want: []string{"BX430", "BX431"}},
		{text: "Large-Area Injury (BX400): treat it as a torso hit", want: []string{"BX400"}},
		{text: "the Explosion modifier (B104) from the Characters book", want: []string{"B104"}},
		{text: "C4 at TL7, 6dx2 cr ex, SM +2 and HP 10 are not references"},
		{text: "B4000 is not a Basic Set page"},
		{text: "plain text without a reference"},
	} {
		c.Equal(tc.want, pageRefPattern.FindAllString(tc.text, -1), tc.text)
	}
}

// TestWrappingLabelLinks drives a note's label the way a click does: the reference must occupy a rectangle within the
// laid-out text, a press and release on it must open exactly that reference, a release elsewhere must not, and text
// outside a reference must not be a link at all. Wrapping must keep working as well, so the label is made narrow
// enough that the reference lands on a later line.
func TestWrappingLabelLinks(t *testing.T) {
	c := check.New(t)
	var opened []string
	label := newWrappingLabel()
	label.linkPageRefs(func(ref string) { opened = append(opened, ref) })
	label.setText("The only defense is Dodge and Drop (BX377): dive away. Work out DR as in Large-Area Injury (BX400).",
		unison.DefaultLabelTheme.OnBackgroundInk)
	_, prefSize, _ := label.Sizes(geom.NewSize(160, 0))
	c.True(len(label.lines(160)) > 1, "the label must wrap to the width it is offered")
	label.SetFrameRect(geom.NewRect(0, 0, 160, prefSize.Height))

	links := label.links()
	if len(links) != 2 {
		t.Fatalf("expected two links, got %d", len(links))
	}
	c.Equal("BX377", links[0].ref)
	c.Equal("BX400", links[1].ref)
	c.True(links[1].rect.Y > links[0].rect.Y, "the second reference must be on a later line")
	for _, link := range links {
		c.True(link.rect.Width > 0 && link.rect.Height > 0, "a link must occupy space")
	}

	second := links[1].rect.Center()
	c.Equal("BX400", label.linkAt(second))
	c.True(label.mouseDown(second, unison.ButtonLeft, 1, mod.None), "a press on a link is consumed")
	c.True(label.mouseUp(second, unison.ButtonLeft, mod.None))
	c.Equal([]string{"BX400"}, opened, "the release opens the reference the press landed on")

	first := links[0].rect.Center()
	c.True(label.mouseDown(first, unison.ButtonLeft, 1, mod.None))
	c.True(label.mouseUp(second, unison.ButtonLeft, mod.None), "the press is still consumed")
	c.Equal([]string{"BX400"}, opened, "a release over a different link opens nothing")

	origin := geom.NewPoint(1, 1)
	c.Equal("", label.linkAt(origin), "the first word is not a link")
	c.False(label.mouseDown(origin, unison.ButtonLeft, 1, mod.None), "a press off every link is not consumed")
	c.False(label.mouseUp(origin, unison.ButtonLeft, mod.None))
	c.False(label.mouseDown(first, unison.ButtonRight, 1, mod.None), "only the left button follows links")
	c.Equal([]string{"BX400"}, opened)
}

// TestSingleLineLabel verifies the mode that stands in for a unison.Label beside a field: the text stays on one line
// however narrow the hint, SetTitle and Title round-trip the text while keeping the ink, and a page reference in the
// title is still a link.
func TestSingleLineLabel(t *testing.T) {
	c := check.New(t)
	var opened []string
	label := newSingleLineLabel()
	label.linkPageRefs(func(ref string) { opened = append(opened, ref) })
	label.SetTitle("effective DR (Large-Area Injury, BX400)")
	c.Equal("effective DR (Large-Area Injury, BX400)", label.Title())
	c.Equal(unison.DefaultLabelTheme.OnBackgroundInk, label.ink, "SetTitle keeps the ink")
	c.Equal(1, len(label.lines(40)), "a single-line label never wraps")
	_, prefSize, _ := label.Sizes(geom.NewSize(40, 0))
	c.True(prefSize.Width > 40, "and asks for the width its text needs")
	label.SetFrameRect(geom.NewRect(0, 0, prefSize.Width, prefSize.Height))
	links := label.links()
	if len(links) != 1 {
		t.Fatalf("expected one link, got %d", len(links))
	}
	center := links[0].rect.Center()
	c.True(label.mouseDown(center, unison.ButtonLeft, 1, mod.None))
	c.True(label.mouseUp(center, unison.ButtonLeft, mod.None))
	c.Equal([]string{"BX400"}, opened, "the reference in a single-line label opens like one in a note")
}
