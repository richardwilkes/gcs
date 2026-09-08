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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// tabBar is a row of buttons, one per tab, of which exactly one is selected at a time. What the tabs switch between is
// the caller's business: SelectionChangedCallback runs with the index of the tab that has just been chosen, whether by
// a click or by selectTab, and only when the choice actually changed. The buttons wrap onto further rows when the bar
// is too narrow to hold them all on one.
type tabBar struct {
	unison.Panel
	SelectionChangedCallback func(index int)
	group                    *unison.Group
	buttons                  []*unison.Button
	selected                 int
}

// newTabBar returns an empty tab bar with a line along its bottom to set it off from what the tabs show.
func newTabBar() *tabBar {
	b := &tabBar{group: unison.NewGroup(), selected: -1}
	b.Self = b
	b.SetBorder(unison.NewCompoundBorder(unison.NewLineBorder(unison.ThemeSurfaceEdge, geom.Size{},
		geom.Insets{Bottom: 1}, false), unison.NewEmptyBorder(unison.StdInsets())))
	b.SetLayout(&unison.FlowLayout{HSpacing: unison.StdHSpacing, VSpacing: unison.StdVSpacing})
	b.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	return b
}

// addTab adds a tab with the given title to the end of the bar and returns its index. No tab is selected until
// selectTab is called.
func (b *tabBar) addTab(title string) int {
	index := len(b.buttons)
	button := unison.NewButton()
	button.Sticky = true
	button.SetTitle(title)
	button.ClickCallback = func() { b.selectTab(index) }
	b.group.Add(button)
	b.buttons = append(b.buttons, button)
	b.AddChild(button)
	return index
}

// selectedIndex returns the index of the selected tab, or -1 when none has been selected yet.
func (b *tabBar) selectedIndex() int {
	return b.selected
}

// selectTab selects the tab at the given index, running SelectionChangedCallback if that is a change. An index that
// names no tab is ignored.
func (b *tabBar) selectTab(index int) {
	if index < 0 || index >= len(b.buttons) {
		return
	}
	b.group.Select(b.buttons[index])
	b.MarkForRedraw()
	if b.selected == index {
		return
	}
	b.selected = index
	if b.SelectionChangedCallback != nil {
		b.SelectionChangedCallback(index)
	}
}
