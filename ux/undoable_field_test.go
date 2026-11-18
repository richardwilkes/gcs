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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestStringFieldSyncWhileFocusedKeepsEdit verifies that a Sync() arriving while the field has the focus leaves the
// text the user is working on alone, even when the focus has left and come back with no sync in between, while one
// arriving once the focus is gone replaces it with the stored value's rendering. The field decides this from its own
// focus callbacks, not from Panel.Focused(), which is false whenever the window is inactive even though the field
// still holds the focus: a sync in that state used to replace a half-typed tag list with its normalized rendering,
// moving the caret out from under the user.
func TestStringFieldSyncWhileFocusedKeepsEdit(t *testing.T) {
	c := check.New(t)
	var tags []string
	f := NewStringField(nil, "", "Tags",
		func() string { return gurps.CombineTags(tags) },
		func(text string) { tags = gurps.ExtractTags(text) })
	c.Equal("", f.Text())

	f.gainedFocus()
	f.SetText("one,tw")
	c.Equal([]string{"one", "tw"}, tags, "the edit is applied as it is typed")
	f.Sync()
	c.Equal("one,tw", f.Text(), "syncing while focused must not replace the text being edited")
	c.Equal([]string{"one", "tw"}, tags, "syncing while focused must not alter the value")

	f.lostFocus()
	f.Sync()
	c.Equal("one, tw", f.Text(), "syncing once the focus is gone shows the stored value's rendering")
	c.Equal([]string{"one", "tw"}, tags, "the rendering is what the field parses back to")

	f.gainedFocus()
	f.lostFocus()
	f.gainedFocus()
	f.SetText("one,tw,th")
	f.Sync()
	c.Equal("one,tw,th", f.Text(), "syncing once the focus is back must not either, with no sync since it left")
}

// TestDisabledFieldIsLeftAloneByTheFocus verifies that a field focused while disabled, which happens only for a screen
// reader, is not treated as being edited: Sync replaces its text and losing the focus does not reformat it. A field
// that could have been edited is still brought into range.
func TestDisabledFieldIsLeftAloneByTheFocus(t *testing.T) {
	c := check.New(t)

	f, value := newTestIntegerField(5, 0, 10)
	f.SetEnabled(false)
	f.gainedFocus()
	*value = 7
	f.Sync()
	c.Equal("7", f.Text(), "a disabled field holding the focus shows the value it is given")
	c.Equal(7, *value)
	*value = 50 // Out of the field's range, as a value from the model may be.
	f.Sync()
	c.Equal("50", f.Text(), "whatever the value is")
	f.lostFocus()
	c.Equal("50", f.Text(), "the focus leaving a field that was only read leaves its text alone")

	// A field enabled by the time the focus leaves it may have been typed into.
	f.gainedFocus()
	f.SetEnabled(true)
	f.lostFocus()
	c.Equal("10", f.Text(), "a field that could be typed into is brought into range")

	// So may one that gained the focus while enabled, even if disabled since.
	f, value = newTestIntegerField(5, 0, 10)
	f.gainedFocus()
	f.Field.SetText("50")
	f.SetEnabled(false)
	f.lostFocus()
	c.Equal("10", f.Text(), "a field that was being edited is brought into range")
	c.Equal(10, *value)
}
