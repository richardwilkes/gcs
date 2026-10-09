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

	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/role"
)

// With a screen reader listening and the setting that puts static text in the Tab order on, the read-only fields (a hit
// location's penalty and DR, and the General Settings path rows) are tab stops, as the labels beside them are. With the
// setting off none of them is, while the notes field and the portrait, being controls, always are. The space shown for
// a location without a roll is never a tab stop, and a read-only field without text is left out of the description.
//
// On macOS and Windows the block headings follow the setting too, but a headless session follows Linux's rules, under
// which unison makes a heading a tab stop whenever a screen reader is listening, since that is how Orca reaches it.
func TestReadOnlyFieldsJoinTheTabOrderForScreenReaders(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusable := uxtest.FocusForReadingSetter(t, screen, wnd)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}

	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var body *BodyPanel
	var portrait *PortraitPanel
	var headings []*unison.Panel
	var readOnly []*NonEditablePageField
	var labels, blank []*unison.Label
	var notes []*StringField
	screen.Do(func() {
		portrait, _ = uxtest.FirstPanelOfType[*PortraitPanel](sheet.AsPanel())
		headings = uxtest.PanelsMatching(sheet.AsPanel(), func(p *unison.Panel) bool {
			return p.Accessibility.Role == role.Heading
		})
		body, ok = uxtest.FirstPanelOfType[*BodyPanel](sheet.AsPanel())
		if !ok {
			return
		}
		readOnly = uxtest.PanelsOfType[*NonEditablePageField](body.AsPanel())
		// The notes header has no text; TestBodyNotesHeaderIsReadAsTheOtherHeadersAre covers it.
		for _, one := range uxtest.PanelsOfType[*unison.Label](body.AsPanel()) {
			switch one.String() {
			case "":
			case " ":
				blank = append(blank, one)
			default:
				labels = append(labels, one)
			}
		}
		notes = uxtest.PanelsOfType[*StringField](body.AsPanel())
	})
	if body == nil {
		t.Fatal("the character sheet must show the body table")
	}
	c.NotNil(portrait, "the character sheet must show the portrait")
	c.True(len(headings) >= 3, "the character sheet's blocks must each carry a heading")
	// Each heading is its block's first child and holds the title strip, which the border leaves to the content, so the
	// block's rows must be laid out below it.
	screen.Do(func() {
		for _, heading := range headings {
			block := heading.Parent()
			children := block.Children()
			c.True(len(children) > 1 && children[0] == heading, "the %q heading comes first in its block", heading.Accessibility.Name)
			c.Equal(block.Accessibility.Name, heading.Accessibility.Name, "the block is named as its heading is")
			bottom := heading.FrameRect().Y + heading.FrameRect().Height
			for _, child := range children[1:] {
				c.True(child.FrameRect().Y >= bottom, "in %q, a %T at %v is laid out under the title strip",
					heading.Accessibility.Name, child.Self, child.FrameRect())
			}
		}
	})
	c.True(len(readOnly) >= 2, "the body table must show a penalty and a DR for each hit location")
	c.True(len(labels) >= 2, "the body table must show a roll and a location for each hit location")
	c.True(len(blank) >= 1, "the default body must have a location without a roll, shown as a space")
	c.True(len(notes) >= 1, "the body table must show a notes field for each hit location")
	for _, one := range readOnly {
		c.NotEqual("", one.Text.String(), "each penalty and DR must have text, or there would be nothing to read")
	}

	// With the setting off, only the controls and, under Linux's rules, the headings take the focus.
	setFocusable(false)
	for _, one := range readOnly {
		c.False(focusable(one), "%q must not be a tab stop while the setting is off", one.Text.String())
	}
	for _, one := range labels {
		c.False(focusable(one), "%q must not be a tab stop while the setting is off", one.String())
	}
	c.True(focusable(notes[0]), "the notes field is a control, so it is always a tab stop")
	c.True(focusable(portrait), "the portrait is a control, so it is always a tab stop")
	for _, one := range headings {
		c.True(focusable(one), "the %q heading is a tab stop under Linux's rules whatever the setting", one.Accessibility.Name)
	}

	setFocusable(true)
	for _, one := range headings {
		c.True(focusable(one), "the %q heading must be a tab stop while the setting is on", one.Accessibility.Name)
	}
	for _, one := range readOnly {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.Text.String())
	}
	for _, one := range labels {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.String())
	}
	for _, one := range blank {
		c.False(focusable(one), "a location without a roll must not offer a blank tab stop")
	}
	c.True(focusable(notes[0]))

	screen.Do(ShowGeneralSettings)
	var paths, empty []*NonEditableField
	screen.Do(func() {
		// The log path is empty here, since only main sets it.
		for _, one := range uxtest.PanelsOfType[*NonEditableField](wnd.Content()) {
			if one.Text.String() != "" {
				paths = append(paths, one)
			} else {
				empty = append(empty, one)
			}
		}
	})
	c.True(len(paths) >= 1, "the General Settings must show its path rows")
	c.True(len(empty) >= 1, "the log path must be empty here, or there is nothing to check the description of")
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	for _, one := range paths {
		c.True(focusable(one), "%q must be a tab stop while the setting is on", one.Text.String())
		node := screen.AccessibilityNodeFor(one)
		if node == nil {
			t.Fatalf("%q must be described", one.Text.String())
		}
		c.False(node.Ignored, "%q must be described", one.Text.String())
		c.Equal(one.Text.String(), node.Name)
	}
	for _, one := range empty {
		c.False(focusable(one), "a field without text must not be a tab stop")
		if node := screen.AccessibilityNodeFor(one); node != nil {
			c.True(node.Ignored, "a field without text must be left out of the description, but is %+v", node)
		}
	}

	setFocusable(false)
	for _, one := range readOnly {
		c.False(focusable(one), "%q must not be a tab stop once the setting is off again", one.Text.String())
	}
	for _, one := range headings {
		c.True(focusable(one), "the %q heading is a tab stop under Linux's rules whatever the setting", one.Accessibility.Name)
	}
	c.True(focusable(portrait))
	for _, one := range paths {
		c.False(focusable(one), "%q must not be a tab stop once the setting is off again", one.Text.String())
	}
}
