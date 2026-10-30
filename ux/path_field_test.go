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

// The field asks the layout for the whole path, so that the path is only shortened when the layout has nothing more to
// give, and then shortens it to whatever width it was given.
func TestPathFieldShortensThePathToItsWidth(t *testing.T) {
	c := check.New(t)
	f := NewPathField(samplePath)
	_, pref, _ := f.Sizes(geom.Size{})
	insets := f.Border().Insets()
	c.Equal(samplePath, f.Text.String(), "the whole path shows until the field is given less room")
	c.Equal(samplePath, f.Accessibility.Name, "the whole path is the field's name")

	f.SetFrameRect(geom.NewRect(0, 0, pref.Width/2, pref.Height))
	shortened := f.Text.String()
	c.True(strings.Contains(shortened, "…"), "the path is shortened to the width the field was given, but is %q", shortened)
	c.True(f.Text.Width() <= pref.Width/2-insets.Width(), "the shortened path %q fits the field", shortened)
	_, stillPref, _ := f.Sizes(geom.Size{})
	c.Equal(pref, stillPref, "the field still asks for the room the whole path needs")
	c.Equal(samplePath, f.Accessibility.Name, "the whole path is still the field's name")

	f.SetFrameRect(geom.NewRect(0, 0, pref.Width, pref.Height))
	c.Equal(samplePath, f.Text.String(), "the whole path shows again once there is room for it")
}
