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
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestListTagFilterHeadless drives the tag filter of a trait list inside a headless workspace, where the modifier keys
// are read from the window as they are for a person's clicks: a click chooses a tag, and a Shift-click on another adds
// it, so a trait has to carry both.
func TestListTagFilterHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &gurps.GlobalSettings().ListFilters, make(map[string][]*gurps.ListFilter))
	d := openListFilterTraitDockable(t, screen)
	var physical, exotic int
	screen.Do(func() {
		physical = d.tagFilter.popup.IndexOfItem("physical")
		exotic = d.tagFilter.popup.IndexOfItem("exotic")
	})
	read := func() (checked, names []string, text string) {
		screen.Do(func() { checked, names, text = checkedTags(d), visibleTraitNames(d), d.tagFilter.popup.Text() })
		return checked, names, text
	}

	uxtest.ChoosePopupItem(t, screen, wnd, d.tagFilter.popup, physical)
	checked, names, _ := read()
	c.Equal([]string{"Physical"}, checked, "the click chooses the tag")
	c.Equal([]string{"Acute Vision"}, names, "only the trait tagged Physical is shown")

	screen.Click(screen.PanelCenter(d.tagFilter.popup))
	var item *unison.Panel
	screen.Do(func() {
		if items := uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd)); exotic >= 0 && exotic < len(items) {
			item = items[exotic]
		}
	})
	if item == nil {
		t.Fatal("the tag filter's menu has no Exotic item")
	}
	screen.ClickWith(screen.PanelCenter(item), unison.ButtonLeft, mod.Shift)
	checked, names, text := read()
	c.Equal([]string{"Exotic", "Physical"}, checked, "the Shift-click adds the tag")
	c.Equal("Multiple", text, "the popup says that more than one tag is chosen")
	c.Equal(0, len(names), "no trait carries both tags")
}
