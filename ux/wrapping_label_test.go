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

// TestWrappingLabelSetText verifies that setText replaces the text and the ink it is drawn in, and asks for the label
// and its ancestors to be laid out again, since the number of lines may have changed.
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
