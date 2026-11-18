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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// bindingTitles returns the titles of the section's own bindings, in order, or nil if it has none.
func bindingTitles(section *menuKeySection) []string {
	if len(section.bindings) == 0 {
		return nil
	}
	titles := make([]string, len(section.bindings))
	for i, b := range section.bindings {
		titles[i] = b.Title
	}
	return titles
}

// submenuTitles returns the titles of the section's submenus, in order, or nil if it has none.
func submenuTitles(section *menuKeySection) []string {
	if len(section.submenus) == 0 {
		return nil
	}
	titles := make([]string, len(section.submenus))
	for i, sub := range section.submenus {
		titles[i] = sub.title
	}
	return titles
}

// TestMenuKeySectionsFollowTheMenuBar verifies that the sections are the menus of the menu bar, in its order, that each
// holds its bindings in the menu's order rather than the title order the bindings come in, that every binding is in
// exactly one of them, so that none is left for a catch-all section, and that the Export submenu's bindings are there
// without the menu having been shown, which is when its templates are filled in. It also pins Undo and Redo to the
// titles they are registered under, since the titles of their actions say "Cannot Undo" until there is something to
// undo.
func TestMenuKeySectionsFollowTheMenuBar(t *testing.T) {
	c := check.New(t)
	sections := menuKeySections()
	titles := make([]string, len(sections))
	for i, section := range sections {
		titles[i] = section.title
	}
	c.Equal([]string{"File", "Edit", "Insert", "Item", "Sheet", "Settings", "View", "Window"}, titles,
		"one section per menu that holds a bindable action, in menu order, with nothing left for an Other section")
	placed := make(map[string]int)
	var count func(section *menuKeySection)
	count = func(section *menuKeySection) {
		for _, b := range section.bindings {
			placed[b.ID]++
		}
		for _, sub := range section.submenus {
			count(sub)
		}
	}
	for _, section := range sections {
		count(section)
	}
	bindings := gurps.CurrentBindings()
	c.Equal(len(bindings), len(placed), "the sections hold nothing but the bindings")
	for _, b := range bindings {
		c.Equal(1, placed[b.ID], "the binding %q must be in exactly one section", b.ID)
	}
	c.Equal([]string{
		"Undo", "Redo", "Cut", "Copy", "Paste", "Duplicate", "Delete", "Select All",
		"Jump to Search/Filter Field",
	}, bindingTitles(sections[1]), "the Edit menu's bindings, in its order")
	c.Equal([]string{"Open…", "Close", "Save", "Save As…", "Print…"}, bindingTitles(sections[0]),
		"the File menu's own bindings, in its order")
	c.Equal([]string{"New", "Export"}, submenuTitles(sections[0]),
		"the File menu's submenus that hold bindings; Open Recent holds none")
	c.Equal([]string{"PDF", "WEBP", "PNG", "JPEG", "Export Portrait"}, bindingTitles(sections[0].submenus[1]),
		"the Export menu's bindings must be present before the menu has ever been shown")
}

// TestMenuKeySectionFiltered verifies that filtering keeps the bindings whose title or key holds the text, dropping the
// submenus and sections left with none, and that empty text keeps everything.
func TestMenuKeySectionFiltered(t *testing.T) {
	c := check.New(t)
	ctrlE := unison.KeyBinding{KeyCode: unison.KeyE, Modifiers: mod.Control}
	ctrlUp := unison.KeyBinding{KeyCode: unison.KeyUp, Modifiers: mod.Control}
	section := &menuKeySection{
		title: "Item",
		bindings: []*gurps.Binding{
			{ID: "open.editor", Title: "Open Editor", KeyBinding: ctrlE},
			{ID: "toggle.state", Title: "Toggle State"},
		},
		submenus: []*menuKeySection{
			{
				title:    "Move",
				bindings: []*gurps.Binding{{ID: "move.up", Title: "Move Up", KeyBinding: ctrlUp}},
			},
		},
	}
	for _, one := range []struct {
		text     string
		bindings []string
		submenus []string
	}{
		{text: "", bindings: []string{"Open Editor", "Toggle State"}, submenus: []string{"Move"}},
		{text: "editor", bindings: []string{"Open Editor"}},
		{text: "move", submenus: []string{"Move"}},
		{text: strings.ToLower(ctrlE.Modifiers.String()), bindings: []string{"Open Editor"}, submenus: []string{"Move"}},
		{text: "nothing"},
	} {
		got := section.filtered(one.text)
		if one.bindings == nil && one.submenus == nil {
			c.Nil(got, "%q: a section left with nothing is dropped", one.text)
			continue
		}
		if got == nil {
			t.Errorf("%q: something must be kept", one.text)
			continue
		}
		c.Equal(one.bindings, bindingTitles(got), "%q: the bindings kept", one.text)
		c.Equal(one.submenus, submenuTitles(got), "%q: the submenus kept", one.text)
	}
	c.True(section.filtered("") == section, "empty text keeps the section as it is")
}
