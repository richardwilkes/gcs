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
	"image/color"
	"strings"
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

const samplePath = "/Users/someone/Library/Application Support/com.trollworks.gcs/gcs_prefs.json"

func TestElideMiddleKeepsTheMostThatFits(t *testing.T) {
	decoration := &unison.TextDecoration{Font: unison.FieldFont}
	full := unison.NewText(samplePath, decoration)
	for _, tc := range []struct {
		name  string
		width float32
	}{
		{name: "whole", width: full.Width()},
		{name: "half", width: full.Width() / 2},
		{name: "tenth", width: full.Width() / 10},
		{name: "none", width: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			got := elideMiddle(full, tc.width, decoration)
			if tc.width >= full.Width() {
				c.Equal(samplePath, got.String(), "the whole path fits, so it is left alone")
				return
			}
			s := got.String()
			head, tail, found := strings.Cut(s, "…")
			c.True(found, "the shortened path %q marks where it was cut", s)
			if tc.width == 0 {
				c.Equal("…", s, "the mark alone is as short as the path gets")
				return
			}
			c.True(strings.HasPrefix(samplePath, head), "the shortened path %q starts as the path does", s)
			c.True(strings.HasSuffix(samplePath, tail), "the shortened path %q ends as the path does", s)
			c.True(got.Width() <= tc.width, "the shortened path %q fits within %v", s, tc.width)
			kept := len([]rune(head)) + len([]rune(tail))
			if kept < len([]rune(samplePath))-1 {
				longer := unison.NewTextFromRunes(elidedRunes([]rune(samplePath), kept+1), decoration)
				c.True(longer.Width() > tc.width, "keeping one more rune, as %q does, would not fit", longer.String())
			}
			c.True(len([]rune(head))-len([]rune(tail)) <= 1, "the cut in %q is in the middle", s)
		})
	}
}

// TestElidePathKeepsTheLastElementWhole verifies that a path is shortened in what leads to its last element, which
// names what the path leads to, until the last element cannot fit on its own, when the whole path is shortened in its
// middle as any text would be.
func TestElidePathKeepsTheLastElementWhole(t *testing.T) {
	decoration := &unison.TextDecoration{Font: unison.FieldFont}
	full := unison.NewText(samplePath, decoration)
	last := unison.NewText("/gcs_prefs.json", decoration)
	ellipsis := unison.NewText("…", decoration)
	for _, tc := range []struct {
		name  string
		width float32
		whole bool
	}{
		{name: "whole", width: full.Width(), whole: true},
		{name: "half", width: full.Width() / 2, whole: true},
		{name: "last element and the mark", width: last.Width() + ellipsis.Width() + 2, whole: true},
		{name: "less than the last element", width: last.Width() / 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			got := elidePath(full, tc.width, decoration)
			s := got.String()
			if tc.width >= full.Width() {
				c.Equal(samplePath, s, "the whole path fits, so it is left alone")
				return
			}
			c.True(strings.Contains(s, "…"), "the shortened path %q marks where it was cut", s)
			c.True(got.Width() <= tc.width, "the shortened path %q fits within %v", s, tc.width)
			if tc.whole {
				c.True(strings.HasSuffix(s, "/gcs_prefs.json"), "the shortened path %q keeps the last element whole", s)
				c.True(strings.HasPrefix(samplePath, strings.SplitN(s, "…", 2)[0]), "the shortened path %q starts as the path does", s)
			} else {
				c.Equal(elideMiddle(full, tc.width, decoration).String(), s, "the last element cannot fit on its own, so the whole path is shortened in its middle")
			}
		})
	}
}

// The field asks the layout for the whole path, so that the path is only shortened when the layout has nothing more to
// give, and then shortens it to whatever width it was given, less what an icon beside it takes.
func TestPathFieldShortensThePathToItsWidth(t *testing.T) {
	c := check.New(t)
	f := NewPathField(samplePath)
	_, pref, _ := f.Sizes(geom.Size{})
	insets := f.Border().Insets()
	c.Equal(samplePath, f.Text.String(), "the whole path shows until the field is given less room")
	c.Equal(samplePath, f.Accessibility.Name, "the whole path is the field's name")
	c.Nil(f.Tooltip, "the whole path is in view, so there is nothing for a tooltip to add")

	f.SetFrameRect(geom.NewRect(0, 0, pref.Width/2, pref.Height))
	shortened := f.Text.String()
	c.True(strings.Contains(shortened, "…"), "the path is shortened to the width the field was given, but is %q", shortened)
	c.True(f.Text.Width() <= pref.Width/2-insets.Width(), "the shortened path %q fits the field", shortened)
	_, stillPref, _ := f.Sizes(geom.Size{})
	c.Equal(pref, stillPref, "the field still asks for the room the whole path needs")
	c.Equal(samplePath, f.Accessibility.Name, "the whole path is still the field's name")
	c.NotNil(f.Tooltip, "the tooltip shows the whole path while a shortened one is drawn")

	f.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	c.Equal(samplePath, f.Text.String(), "the whole path shows again once there is room for it")
	c.Nil(f.Tooltip, "and the tooltip goes with the shortening")

	f.SetProblem("not found")
	c.NotNil(f.Tooltip, "a problem with the path is told whether or not the path is shortened")
	if f.Drawable == nil {
		t.Fatal("a problem with the path is marked with an icon")
	}
	f.SetFrameRect(geom.NewRect(0, 0, pref.Width/2, pref.Height))
	shortened = f.Text.String()
	c.True(f.Text.Width() <= pref.Width/2-insets.Width()-f.Drawable.LogicalSize().Width-f.Gap,
		"the shortened path %q leaves room for the icon beside it", shortened)
	f.SetProblem("")
	c.Nil(f.Drawable, "the mark goes with the problem")
	f.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	c.Nil(f.Tooltip, "and nothing is left to tell once the whole path is in view again")
}

// TestPathFieldWashesAPathWithAProblem verifies that a field whose path has a problem is washed with the error color,
// as an invalid field is, whether or not it has a path to show.
func TestPathFieldWashesAPathWithAProblem(t *testing.T) {
	for _, tc := range []struct {
		name string
		path string
	}{
		{name: "a path", path: samplePath},
		{name: "no path", path: ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			// sample returns the color the field is drawn in, with the problem if one is given, at a point clear of the
			// path and the icon, in the room the field is given beyond what it asks for.
			sample := func(problem string) color.NRGBA {
				f := NewPathField(tc.path)
				f.SetProblem(problem)
				_, pref, _ := f.Sizes(geom.Size{})
				f.SetFrameRect(geom.NewRect(0, 0, pref.Width+40, pref.Height))
				r := f.ContentRect(true)
				img, err := unison.NewImageFromDrawing(int(r.Width), int(r.Height), 72, func(gc *unison.Canvas) {
					f.DrawCallback(gc, r)
				})
				if err != nil {
					t.Fatalf("unable to draw the field: %v", err)
				}
				defer img.Dispose()
				pixels, err := img.ToNRGBA()
				if err != nil {
					t.Fatalf("unable to read the drawing: %v", err)
				}
				return pixels.NRGBAAt(int(r.Width)-10, int(r.Height/2))
			}
			c.NotEqual(sample(""), sample("not found"), "the field is washed while its path has a problem")
		})
	}
}

// TestPathFieldShowsNothingWithoutAPath verifies that a field with no path shows nothing however little room it is
// given, even less than a problem's icon takes, rather than an ellipsis standing for a path that is not there.
func TestPathFieldShowsNothingWithoutAPath(t *testing.T) {
	c := check.New(t)
	f := NewPathField("")
	f.SetProblem("not found")
	_, pref, _ := f.Sizes(geom.Size{})
	f.SetFrameRect(geom.NewRect(0, 0, f.Border().Insets().Width(), pref.Height))
	c.Equal("", f.Text.String())
}

// TestPathFieldHighlightsUnderThePointer verifies that a clickable path field's border takes the focus color while
// the pointer is over it, so that it is seen to be clickable, and goes back to the plain field border when the pointer
// leaves, while a field that is only shown keeps its border whatever the pointer does.
func TestPathFieldHighlightsUnderThePointer(t *testing.T) {
	c := check.New(t)
	f := NewPathField(samplePath)
	plain := f.Border()
	c.Nil(f.MouseEnterCallback, "a path that is only shown has nothing to show the pointer")
	f.SetClickCallback(func() {})
	c.True(f.Border() == plain, "a clickable path starts with the plain field border")
	f.MouseEnterCallback(geom.Point{}, 0)
	c.True(f.Border() == f.hoverBorder, "the border is highlighted while the pointer is over the field")
	f.MouseExitCallback()
	c.True(f.Border() == plain, "and is plain again once the pointer leaves")
}

// TestPathFieldClickCallbackMakesItAButton verifies that a path field given a click callback calls it for a click
// that ends on the field, and for Space while it has the focus, which it must then be able to take, and not otherwise.
func TestPathFieldClickCallbackMakesItAButton(t *testing.T) {
	c := check.New(t)
	f := NewPathField(samplePath)
	c.False(f.Focusable(), "a path that is only shown is not a tab stop")
	clicks := 0
	f.SetClickCallback(func() { clicks++ })
	c.True(f.Focusable(), "a path that can be clicked is a tab stop")
	_, pref, _ := f.Sizes(geom.Size{})
	f.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	inside := geom.NewPoint(pref.Width/2, pref.Height/2)
	outside := geom.NewPoint(pref.Width*2, pref.Height/2)

	f.MouseDownCallback(inside, 1, 1, 0)
	f.MouseUpCallback(inside, 1, 0)
	c.Equal(1, clicks, "a click on the field calls the callback")
	f.MouseDownCallback(inside, 1, 1, 0)
	f.MouseDragCallback(outside, 1, 0)
	f.MouseUpCallback(outside, 1, 0)
	c.Equal(1, clicks, "a press dragged off the field before release does not")

	c.True(f.KeyDownCallback(unison.KeySpace, 0, false), "Space is taken")
	c.Equal(2, clicks, "and calls the callback")
	c.True(f.KeyDownCallback(unison.KeySpace, 0, true), "a held Space is still taken")
	c.Equal(2, clicks, "but acts only once")
	c.False(f.KeyDownCallback(unison.KeyReturn, 0, false), "other keys are left to what holds the field")
	c.Equal(2, clicks)
}
