// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/role"
)

// A label that follows a field, such as its units, is the field's description rather than an element of its own, so it
// is no tab stop with static text in the Tab order.
func TestTrailingLabelsDescribeTheirFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	setFocusForReading := uxtest.FocusForReadingSetter(t, screen, wnd)
	sheet := openNewCharacterSheet(t, screen)
	screen.Do(func() { Display(sheet) })
	calc := uxtest.SoleEditor[*Dockable](t, screen, func(d unison.Dockable) bool {
		_, isCalculator := d.AsPanel().Self.(*Dockable)
		return isCalculator
	})
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)
	setFocusForReading(true)
	focusable := func(p unison.Paneler) (result bool) {
		screen.Do(func() { result = p.AsPanel().Focusable() })
		return result
	}

	var trailing *ux.TextLabel
	screen.Do(func() {
		children := hiking.restField.Parent().Children()
		if len(children) == 2 {
			if label, isLabel := children[1].Self.(*ux.TextLabel); isLabel {
				trailing = label
			}
		}
	})
	if trailing == nil {
		t.Fatal("the rest field must be followed by its label")
	}
	node := screen.AccessibilityNodeFor(hiking.restField)
	if node == nil {
		t.Fatal("the rest field must be described")
	}
	c.Equal("minutes of rest halfway through the day (0 for none)", node.Description)
	c.False(focusable(trailing), "the label is heard with the field, so it is no tab stop")
	if label := screen.AccessibilityNodeFor(trailing); label != nil {
		c.True(label.Ignored, "the label is heard with the field, so it is not described, but is %+v", label)
	}
	// The description follows the label's text as it changes.
	screen.Do(func() { trailing.SetTitle("minutes of rest") })
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	node = screen.AccessibilityNodeFor(hiking.restField)
	c.Equal("minutes of rest", node.Description)
	// A label holding a link is described, and is a tab stop so that the link can be reached.
	screen.Do(func() { trailing.SetTitle("minutes of rest (B426)") })
	if screen.AccessibilityTree(wnd) == nil {
		t.Fatal("the window must be described")
	}
	c.True(focusable(trailing), "a label holding a link is a tab stop")
	label := screen.AccessibilityNodeFor(trailing)
	if label == nil {
		t.Fatal("a label holding a link must be described")
	}
	c.False(label.Ignored)
	c.Equal(role.Label, label.Role)
	c.Equal(1, len(label.Children), "the link is an element within the label")
	c.Equal("minutes of rest (B426)", screen.AccessibilityNodeFor(hiking.restField).Description)
}
