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
	"github.com/richardwilkes/unison/enums/mod"
)

// TestTreasureGenDialogGatesOKOnValidFields checks that the treasure generator's OK button follows the validity of the
// text the fields hold now. A field reports a changed value before it validates its new text, so an OK state taken from
// the fields' validity as the value changes is a step behind: it stayed enabled while a field was empty, and was
// disabled once the field held a valid value again.
func TestTreasureGenDialogGatesOKOnValidFields(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newLootSheetAction).(*LootSheet)
	if !ok {
		t.Fatal("New Loot Sheet must open a loot sheet")
	}
	done := false
	if !screen.Post(func() {
		sheet.generateTreasure()
		done = true
	}) {
		t.Fatal("the treasure generator could not be posted")
	}
	screen.Sync()
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	okButton := dialogButton(t, screen, dialog, unison.ModalResponseOK)
	cancel := dialogButton(t, screen, dialog, unison.ModalResponseCancel)
	var fields []*DecimalField
	screen.Do(func() { fields = uxtest.PanelsOfType[*DecimalField](dialogWnd.Content()) })
	if len(fields) != 2 {
		t.Fatalf("the treasure generator must have a minimum and a maximum field; it has %d fields", len(fields))
	}
	okEnabled := func() bool {
		var enabled bool
		screen.Do(func() { enabled = okButton.Enabled() })
		return enabled
	}
	replaceText := func(field *DecimalField, text string) {
		screen.Click(screen.PanelCenter(field))
		screen.KeyPress(unison.KeyA, mod.OSMenuCommand())
		if text == "" {
			screen.KeyPress(unison.KeyBackspace, 0)
			return
		}
		screen.Type(text)
	}
	c.True(okEnabled(), "the dialog opens with valid values")
	replaceText(fields[0], "")
	c.False(okEnabled(), "an empty minimum field holds no valid value")
	replaceText(fields[0], "5")
	c.True(okEnabled(), "a valid minimum allows generation again")
	replaceText(fields[0], "1")
	c.True(okEnabled())
	replaceText(fields[0], "")
	c.False(okEnabled(), "an empty minimum field is invalid even when its text parses to the value the field held")
	replaceText(fields[0], "5")
	c.True(okEnabled())
	replaceText(fields[1], "")
	c.False(okEnabled(), "an empty maximum field holds no valid value")
	replaceText(fields[1], "50")
	c.True(okEnabled(), "a valid maximum allows generation again")
	replaceText(fields[1], "4")
	c.False(okEnabled(), "a maximum below the minimum does not allow generation")
	replaceText(fields[1], "50")
	c.True(okEnabled())
	screen.Click(screen.PanelCenter(cancel))
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "the dialog has been dismissed")
	c.True(done, "the treasure generator has finished")
}
