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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/attribute"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// The buttons a sheet keeps out of the Tab order (the description's randomize buttons and the buttons that show and
// hide an attribute section or a hit location's sub-table) are tab stops that Space presses while a screen reader is
// listening and the setting that puts static text and disabled controls in the Tab order is on. Otherwise they take no
// focus, and one that is a tab stop when the screen reader goes drops out as the focus leaves it.
func TestSheetButtonsJoinTheTabOrderForReading(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	setFocusForReading := focusForReadingSetter(t, screen, wnd)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var buttons []*unison.Button
	var randomize *unison.Button
	var gender *StringField
	var inDescription, inBody, inAttributes int
	screen.Do(func() {
		// A new sheet has neither a hit location with a sub-table nor a section among its attributes.
		entity := sheet.Entity()
		settings := gurps.SheetSettingsFor(entity)
		body := settings.BodyType
		subTable := &gurps.Body{Roll: gurps.Roller.Parse("1d")}
		subTable.SetOwningLocation(body.Locations[0])
		body.Locations[0].SubTable = subTable
		loc := gurps.NewHitLocation(entity, "")
		loc.LocID = "within"
		loc.ChoiceName = "Within"
		loc.TableName = "Within"
		loc.Slots = 6
		subTable.AddLocation(loc)
		body.Update(entity)
		settings.Attributes.Set["section"] = &gurps.AttributeDef{
			DefID: "section",
			Type:  attribute.SecondarySeparator,
			Name:  "Section",
			Order: len(settings.Attributes.Set) + 1,
		}
		entity.Recalculate()
		sheet.Rebuild(true)
		desc, _ := firstPanelOfType[*DescriptionPanel](sheet.AsPanel())
		if desc == nil {
			return
		}
		for _, one := range panelsOfType[*unison.Button](desc.AsPanel()) {
			buttons = append(buttons, one)
			if randomize == nil {
				randomize = one
			}
		}
		for _, one := range panelsOfType[*StringField](desc.AsPanel()) {
			if one.RefKey == descriptionPanelGenderFieldRefKey {
				gender = one
			}
		}
		inDescription = len(buttons)
		if body, found := firstPanelOfType[*BodyPanel](sheet.AsPanel()); found {
			buttons = append(buttons, panelsOfType[*unison.Button](body.AsPanel())...)
		}
		inBody = len(buttons) - inDescription
		for _, one := range panelsOfType[*AttrPanel](sheet.AsPanel()) {
			buttons = append(buttons, panelsOfType[*unison.Button](one.AsPanel())...)
		}
		inAttributes = len(buttons) - inDescription - inBody
	})
	if randomize == nil || gender == nil {
		t.Fatal("the character sheet must show the description, with a Gender field and its randomize button")
	}
	c.Equal(9, inDescription, "the description must have its randomize buttons")
	c.Equal(1, inBody, "the hit location with a sub-table must have the button that shows and hides it")
	c.Equal(1, inAttributes, "the section of the attributes must have the button that shows and hides it")
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	focus := func() (focus *unison.Panel) {
		screen.Do(func() { focus = wnd.Focus() })
		return focus
	}
	none := func(why string) {
		t.Helper()
		for _, one := range buttons {
			c.False(focusable(one), "%s, but a button took the focus", why)
		}
	}
	none("a button kept out of the Tab order takes no focus while no screen reader is listening")

	setFocusForReading(false)
	none("a button kept out of the Tab order takes no focus while the setting is off")
	node := screen.AccessibilityNodeFor(randomize)
	if node == nil {
		t.Fatal("the randomize button must be described")
	}
	c.Equal(role.Button, node.Role)
	c.False(node.Focusable, "the button is described as it is")
	c.False(node.Actions.Has(accessibility.Focus))
	c.True(node.Actions.Has(accessibility.Press), "a screen reader can press the button even so")

	setFocusForReading(true)
	for _, one := range buttons {
		c.True(focusable(one), "a button kept out of the Tab order is a tab stop while the setting is on")
	}
	node = screen.AccessibilityNodeFor(randomize)
	if node == nil {
		t.Fatal("the randomize button must be described")
	}
	c.True(node.Focusable, "the button is described as able to take the focus")
	c.True(node.Actions.Has(accessibility.Focus))
	c.Equal("Randomize the gender using the current ancestry", node.Name)

	// Tab goes from the tab stop ahead of the button to the button, and Space presses it.
	var ahead *unison.Panel
	screen.Do(func() {
		all := panelsMatching(sheet.scroll.Content().AsPanel(), (*unison.Panel).Focusable)
		for i, one := range all {
			if one == randomize.AsPanel() && i > 0 {
				ahead = all[i-1]
			}
		}
	})
	if ahead == nil {
		t.Fatal("the randomize button must follow another tab stop")
	}
	screen.Do(ahead.RequestFocus)
	c.Equal(ahead, focus())
	screen.KeyPress(unison.KeyTab, mod.None)
	c.Equal(randomize.AsPanel(), focus(), "Tab reaches the randomize button")
	screen.Do(func() {
		gender.SetText("")
		sheet.Entity().Profile.Gender = ""
	})
	screen.KeyPress(unison.KeySpace, mod.None)
	var text string
	screen.Do(func() { text = gender.Text() })
	c.NotEqual("", text, "Space on the randomize button randomizes the gender")
	c.Equal(gender.AsPanel(), focus(), "and puts the focus on the field, where a screen reader reads what it now holds")

	// A screen reader's Focus action on the button moves the focus there.
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.ID, Action: accessibility.Focus}))
	c.Equal(randomize.AsPanel(), focus())

	setFocusForReading(false)
	none("a button kept out of the Tab order takes no focus once the setting is off again")

	// Nothing describes the window once the screen reader goes, so the button stays a tab stop until it loses the
	// focus.
	setFocusForReading(true)
	screen.Do(randomize.RequestFocus)
	c.Equal(randomize.AsPanel(), focus())
	t.Cleanup(func() { unison.SetAccessibilityEnabled(true) })
	screen.Do(func() { unison.SetAccessibilityEnabled(false) })
	var active bool
	screen.Do(func() { active = unison.IsAccessibilityActive() })
	c.False(active, "the screen reader must have gone")
	screen.Do(gender.RequestFocus)
	c.Equal(gender.AsPanel(), focus())
	c.False(focusable(randomize), "a button the focus has left is no tab stop once the screen reader has gone")
}

// The hit locations' notes header is only an icon, so it is a label named Notes, described by its tooltip, and a tab
// stop with static text in the Tab order, as the Roll, Location and DR headers are.
func TestBodyNotesHeaderIsReadAsTheOtherHeadersAre(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	setFocusForReading := focusForReadingSetter(t, screen, wnd)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var headers []*unison.Label
	screen.Do(func() {
		body, found := firstPanelOfType[*BodyPanel](sheet.AsPanel())
		if !found {
			return
		}
		// The first seven rows are the four headers and the empty panels in the separator columns between them.
		if rows := blockRows(body.AsPanel()); len(rows) > 7 {
			for _, child := range rows[:7] {
				if label, isLabel := child.Self.(*unison.Label); isLabel {
					headers = append(headers, label)
				}
			}
		}
	})
	if len(headers) != 4 {
		t.Fatalf("the body table must have four headers, but has %d", len(headers))
	}
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	setFocusForReading(true)
	names := make([]string, 0, len(headers))
	for _, one := range headers {
		node := screen.AccessibilityNodeFor(one)
		if node == nil {
			t.Fatalf("the %q header must be described", one.String())
		}
		names = append(names, node.Name)
		c.Equal(role.Label, node.Role, "the %q header is a label", node.Name)
		c.False(node.Ignored, "the %q header must be described", node.Name)
		c.True(node.Focusable, "the %q header is described as a tab stop", node.Name)
		c.True(focusable(one), "the %q header is a tab stop with static text in the Tab order", node.Name)
	}
	c.Equal([]string{"Roll", "Location", "DR", "Notes"}, names)
	notes := screen.AccessibilityNodeFor(headers[3])
	c.Equal("Notes for the hit location", notes.Description, "the tooltip says the rest")

	setFocusForReading(false)
	for _, one := range headers {
		c.False(focusable(one), "the %q header is no tab stop while the setting is off", one.String())
	}
}

// A label that follows a field, such as its units, is the field's description rather than an element of its own, so it
// is no tab stop with static text in the Tab order. (The calculators' link-aware labels are covered by the same test
// in ux/calculators.)
func TestTrailingLabelsDescribeTheirFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	setFocusForReading := focusForReadingSetter(t, screen, wnd)
	setFocusForReading(true)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}
	screen.Do(ShowGeneralSettings)
	var settings *generalSettingsDockable
	screen.Do(func() { settings, _ = firstPanelOfType[*generalSettingsDockable](wnd.Content()) })
	if settings == nil {
		t.Fatal("the General Settings must be showing")
	}
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	for _, one := range []struct {
		field *unison.Panel
		name  string
		hint  string
	}{
		{field: settings.tooltipDelayField.AsPanel(), name: "Tooltip Delay", hint: "seconds"},
		{field: settings.tooltipDismissalField.AsPanel(), name: "Tooltip Dismissal", hint: "seconds"},
		{field: settings.cursorSizeField.AsPanel(), name: "Cursor Size", hint: "points"},
		{field: settings.exportResolutionField.AsPanel(), name: "Image Export Resolution", hint: "ppi"},
	} {
		node := screen.AccessibilityNodeFor(one.field)
		if node == nil {
			t.Fatalf("the %s field must be described", one.name)
		}
		c.Equal(one.name, node.Name)
		c.Equal(one.hint, node.Description, "the %s field is described by the label that follows it", one.name)
		var units *unison.Label
		screen.Do(func() {
			siblings := one.field.Parent().Children()
			if len(siblings) == 2 {
				if found, isLabel := siblings[1].Self.(*unison.Label); isLabel {
					units = found
				}
			}
		})
		if units == nil {
			t.Fatalf("the %s field must be followed by its units", one.name)
		}
		c.Equal(one.hint, units.String())
		c.False(focusable(units), "the units of the %s field are heard with it, so they are no tab stop", one.name)
		if label := screen.AccessibilityNodeFor(units); label != nil {
			c.True(label.Ignored, "the units of the %s field are not described, but are %+v", one.name, label)
		}
	}
}
