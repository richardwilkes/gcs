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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// menuKeySection is a menu as the Menu Keys settings present it: its title, the bindings of its own items in menu order
// and, after them, its submenus that hold any bindings.
type menuKeySection struct {
	title    string
	bindings []*gurps.Binding
	submenus []*menuKeySection
}

// menuKeySections returns the menus of the menu bar that hold key-bindable actions, in menu order, built from a bar of
// in-window menus so that no window is needed. A binding whose action is in no menu, such as one offered only by a
// context menu, goes in a final section of its own, so that it can still be given a key.
func menuKeySections() []*menuKeySection {
	registerKeyBindingsOnce.Do(registerActions)
	bindings := gurps.CurrentBindings()
	unplaced := make(map[int]*gurps.Binding, len(bindings))
	for _, b := range bindings {
		unplaced[b.Action.ID] = b
	}
	bar := unison.NewInWindowMenuFactory().NewMenu(unison.RootMenuID, "", nil)
	menuBarScope{}.populate(bar)
	var sections []*menuKeySection
	if root := newMenuKeySection(bar, unplaced); root != nil {
		sections = root.submenus
	}
	if len(unplaced) != 0 {
		other := &menuKeySection{title: i18n.Text("Other")}
		for _, b := range bindings { // Keeps them in title order
			if _, ok := unplaced[b.Action.ID]; ok {
				other.bindings = append(other.bindings, b)
			}
		}
		sections = append(sections, other)
	}
	return sections
}

// newMenuKeySection returns the section for the menu, or nil if neither it nor its submenus hold any of the bindings,
// removing the ones it places from unplaced.
func newMenuKeySection(menu unison.Menu, unplaced map[int]*gurps.Binding) *menuKeySection {
	section := &menuKeySection{title: menu.Title()}
	for i := range menu.Count() {
		item := menu.ItemAtIndex(i)
		if item.IsSeparator() {
			continue
		}
		if sub := item.SubMenu(); !xreflect.IsNil(sub) { // A typed nil for an item with no submenu
			if one := newMenuKeySection(sub, unplaced); one != nil {
				section.submenus = append(section.submenus, one)
			}
		} else if b, ok := unplaced[item.ID()]; ok {
			section.bindings = append(section.bindings, b)
			delete(unplaced, item.ID())
		}
	}
	if len(section.bindings) == 0 && len(section.submenus) == 0 {
		return nil
	}
	return section
}

// filtered returns the section with only the bindings whose title or key contains text, which must be lower case, and
// the submenus that keep any, or nil if nothing is left. Empty text keeps everything.
func (s *menuKeySection) filtered(text string) *menuKeySection {
	if text == "" {
		return s
	}
	result := &menuKeySection{title: s.title}
	for _, b := range s.bindings {
		if strings.Contains(strings.ToLower(b.Title), text) ||
			strings.Contains(strings.ToLower(b.KeyBinding.String()), text) {
			result.bindings = append(result.bindings, b)
		}
	}
	for _, sub := range s.submenus {
		if one := sub.filtered(text); one != nil {
			result.submenus = append(result.submenus, one)
		}
	}
	if len(result.bindings) == 0 && len(result.submenus) == 0 {
		return nil
	}
	return result
}
