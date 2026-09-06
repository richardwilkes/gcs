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
	"testing"

	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// TestNonEditablePageFieldSetTitleIfChanged verifies that a field built for a text function shows that text from the
// start, that a sync which yields the same text neither replaces it nor asks for a layout, and that one which yields
// different text does both, marking the field's ancestors along with it.
func TestNonEditablePageFieldSetTitleIfChanged(t *testing.T) {
	c := check.New(t)
	text := "one"
	field := NewNonEditablePageFieldEndFor(func() string { return text })
	c.Equal("one", field.Text.String(), "the field is synced when created")
	c.Equal(align.End, field.HAlign)
	c.Equal(align.Start, NewNonEditablePageFieldFor(func() string { return text }).HAlign)

	parent := unison.NewPanel()
	parent.AddChild(field)
	parent.NeedsLayout = false
	field.NeedsLayout = false
	c.False(field.SetTitleIfChanged("one"), "the same text is not a change")
	c.False(field.NeedsLayout, "unchanged text asks for no layout")
	c.False(parent.NeedsLayout, "unchanged text asks for no layout")

	text = "two"
	field.Sync()
	c.Equal("two", field.Text.String(), "a sync picks up the new text")
	c.True(field.NeedsLayout, "changed text asks for a layout")
	c.True(parent.NeedsLayout, "changed text asks its ancestors for a layout")

	field.NeedsLayout = false
	c.True(field.SetTitleIfChanged("three"), "different text is a change")
	c.Equal("three", field.Text.String())
	c.True(field.NeedsLayout)
}
