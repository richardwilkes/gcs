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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// The picker's page reference link is outside any list, so it is an ordinary tab stop, with or without a screen reader.
func TestPickerRowPageReferenceIsFollowedFromTheKeyboard(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	const ref = "https://example.com/ref"
	var holder, link *unison.Panel
	var box *unison.CheckBox
	var linkCount int
	screen.Do(func() {
		trait := gurps.NewTrait(nil, nil, false)
		trait.Name = "Alpha"
		trait.PageRef = ref
		holder = unison.NewPanel()
		boxes := addPickerRow(promptOperation{}, holder, trait, picker.Count, false, func() {}, nil)
		if len(boxes) == 1 {
			box = boxes[0]
		}
		links := panelsMatching(holder, func(p *unison.Panel) bool { return p.Accessibility.Role == role.Link })
		linkCount = len(links)
		if linkCount == 1 {
			link = links[0]
		}
		wnd.Content().AddChild(holder)
	})
	t.Cleanup(func() { screen.Do(holder.RemoveFromParent) })
	if box == nil {
		t.Fatal("the option must have a checkbox")
	}
	if link == nil {
		t.Fatalf("the option must end with its page reference as a link, but holds %d links", linkCount)
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	screen.Do(box.RequestFocus)
	c.Equal(box.AsPanel(), focus())
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(link, focus(), "Tab goes from the checkbox to the page reference")
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal([]string{ref}, screen.OpenedURLs(), "Return follows the page reference")
	screen.KeyPress(unison.KeySpace, mod.None)
	c.Equal([]string{ref}, screen.OpenedURLs(), "and so does Space")
	c.Equal(link, focus(), "the focus stays on the page reference")
}
