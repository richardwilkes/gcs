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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/role"
)

// A new sheet puts the focus on its name field, not on the portrait laid out ahead of it, where typing would go
// nowhere, nor on the Identity heading, which can take the focus while a screen reader is listening. The same choice is
// made whenever the sheet's focus has to be found afresh, as when the layout editor closes.
func TestSheetOpensWithTheFocusOnItsName(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	focusKey := func() (key string) {
		screen.Do(func() {
			if focus := wnd.Focus(); focus != nil {
				key = focus.RefKey
			}
		})
		return key
	}
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the sheet opens with the focus on its name")
	var portrait *PortraitPanel
	screen.Do(func() { portrait, _ = firstPanelOfType[*PortraitPanel](sheet.AsPanel()) })
	if portrait == nil {
		t.Fatal("the character sheet must show the portrait")
	}
	var portraitFirst bool
	screen.Do(func() {
		portraitFirst = firstFocusableInSubtree(sheet.scroll.Content().AsPanel(), nil) == portrait.AsPanel()
	})
	c.True(portraitFirst, "the portrait is the first tab stop on the page, which is what makes the choice one")
	screen.Do(func() { FocusFirstContent(sheet.toolbar, sheet.scroll.Content()) })
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the focus found afresh goes to the name as well")

	// Under Linux's rules, which a headless session follows, the headings take the focus while a screen reader is
	// listening and come ahead of the name; they are passed over too.
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	headingIndex, nameIndex := -1, -1
	screen.Do(func() {
		for i, one := range panelsMatching(sheet.scroll.Content().AsPanel(), (*unison.Panel).Focusable) {
			switch {
			case one.Accessibility.Role == role.Heading && headingIndex < 0:
				headingIndex = i
			case one.RefKey == identityPanelNameFieldRefKey:
				nameIndex = i
			}
		}
	})
	c.True(headingIndex >= 0 && headingIndex < nameIndex, "a heading is now a tab stop ahead of the name, but the "+
		"first heading is tab stop %d and the name is tab stop %d", headingIndex, nameIndex)
	screen.Do(func() { FocusFirstContent(sheet.toolbar, sheet.scroll.Content()) })
	c.Equal(identityPanelNameFieldRefKey, focusKey(), "the focus found afresh still goes to the name")
}

// A tab stop without a reference key, such as the portrait, a heading or a label, keeps the focus through a change that
// leaves it in place. When the change removes it, the focus goes to the name field, as for a keyed field that is gone.
func TestSheetKeepsTheFocusOnKeylessTabStops(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	var portrait *PortraitPanel
	screen.Do(func() { portrait, _ = firstPanelOfType[*PortraitPanel](sheet.AsPanel()) })
	if portrait == nil {
		t.Fatal("the character sheet must show the portrait")
	}
	screen.Do(portrait.RequestFocus)
	c.Equal(portrait.AsPanel(), focus())
	screen.Do(func() { sheet.updatePortrait(nil) })
	c.Equal(portrait.AsPanel(), focus(), "changing the portrait leaves the focus on it")
	screen.Do(func() { sheet.Rebuild(false) })
	c.Equal(portrait.AsPanel(), focus(), "a rebuild leaves the focus on it")

	// Under Linux's rules a heading takes the focus once asking for the tree has activated accessibility. The Identity
	// block never rebuilds its contents, so its heading survives Rebuild(false).
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	var heading *unison.Panel
	screen.Do(func() {
		for _, one := range panelsMatching(sheet.AsPanel(), func(p *unison.Panel) bool {
			return p.Accessibility.Role == role.Heading && p.Accessibility.Name == "Identity"
		}) {
			heading = one
		}
	})
	if heading == nil {
		t.Fatal("the character sheet must show the Identity heading")
	}
	screen.Do(heading.RequestFocus)
	c.Equal(heading, focus(), "the heading takes the focus while a screen reader is listening")
	screen.Do(func() { sheet.Rebuild(false) })
	c.Equal(heading, focus(), "a rebuild leaves the focus on the heading")

	// An attribute block's rebuild replaces its rows and their read-only points fields, which are keyless tab stops
	// with static text in the Tab order (the name label is not, being the value field's caption).
	focusForReadingSetter(t, screen, wnd)(true)
	var attrs *AttrPanel
	var points *NonEditablePageField
	screen.Do(func() {
		for _, one := range panelsOfType[*AttrPanel](sheet.AsPanel()) {
			if fields := panelsOfType[*NonEditablePageField](one.AsPanel()); len(fields) > 0 {
				attrs = one
				points = fields[0]
				break
			}
		}
	})
	if points == nil {
		t.Fatal("the character sheet must show an attribute block with a points field")
	}
	screen.Do(points.RequestFocus)
	c.Equal(points.AsPanel(), focus(), "the points field takes the focus with static text in the Tab order")
	screen.Do(func() { attrs.rebuild(gurps.SheetSettingsFor(sheet.Entity()).Attributes) })
	var key string
	var stillThere bool
	screen.Do(func() {
		stillThere = points.Window() == wnd
		if f := wnd.Focus(); f != nil {
			key = f.RefKey
		}
	})
	c.False(stillThere, "the points field is gone with the rows")
	c.Equal(identityPanelNameFieldRefKey, key, "the focus goes to the name field")
}
