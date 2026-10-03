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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/accessibility"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

func TestSentenceButton(t *testing.T) {
	c := check.New(t)
	screen, _ := startHeadlessWorkspace(t, c)
	text := "Has trait " + emphasize("Magery") + " at level at least 1, along with enough other words to need wrapping"
	clicks := 0
	var button, static *sentenceButton
	screen.Do(func() {
		button = newSentenceButton(text, func() { clicks++ })
		static = newSentenceButton("Summary", nil)
		_, wide, _ := button.Sizes(geom.Size{Width: 2000})
		_, narrow, _ := button.Sizes(geom.Size{Width: 150})
		c.True(narrow.Width <= 150, "the text wraps to the width it is given: %v", narrow)
		c.True(narrow.Height > wide.Height, "wrapping adds lines: %v versus %v", narrow, wide)
	})
	wnd := showInTestWindow(t, screen, 300, button, static)

	screen.Click(screen.PanelCenter(button))
	c.Equal(1, clicks, "a click activates it")
	screen.Do(button.RequestFocus)
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.KeyPress(unison.KeyReturn, mod.None)
	c.Equal(3, clicks, "Space and Enter activate it")
	screen.KeyPress(unison.KeySpace, mod.Command)
	c.Equal(3, clicks, "a modified Space does not")

	screen.AccessibilityTree(wnd)
	node := screen.AccessibilityNodeFor(button)
	if node == nil {
		t.Fatal("the sentence must be described")
	}
	c.Equal(role.DisclosureTriangle, node.Role)
	c.True(node.Expandable)
	c.False(node.Expanded, "what it discloses takes its place, so it is never shown expanded")
	c.True(node.Focusable)
	c.True(node.Actions.Has(accessibility.Press))
	c.Equal("Has trait Magery at level at least 1, along with enough other words to need wrapping", node.Name)
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.ID, Action: accessibility.Collapse}))
	c.Equal(3, clicks, "collapsing what is already collapsed does nothing")
	c.True(screen.PerformAccessibilityAction(accessibility.ActionRequest{Node: node.ID, Action: accessibility.Expand}))
	c.Equal(4, clicks)

	screen.Do(func() { button.setText("Plain "+emphasize("bold"), "met") })
	screen.AccessibilityTree(wnd)
	c.Equal("Plain bold, met", screen.AccessibilityNodeFor(button).Name, "the suffix follows the sentence")

	node = screen.AccessibilityNodeFor(static)
	if node == nil {
		t.Fatal("static text must be described")
	}
	c.Equal(role.Label, node.Role)
	c.Equal("Summary", node.Name)
	c.False(node.Focusable, "static text is no tab stop while not reading")
}
