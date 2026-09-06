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
	"fmt"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMarkModifiedDropsCallsWhileTheUpdateIsUnderWay verifies that a call to MarkModified arriving from inside the
// sync it performs -- a panel reporting a change as it is brought into line with the model -- is dropped rather than
// acted upon, so that the update doesn't repeat itself from inside itself, and that the guard is lifted once the update
// is over, so that the next edit is acted upon.
func TestMarkModifiedDropsCallsWhileTheUpdateIsUnderWay(t *testing.T) {
	c := check.New(t)
	template := newTestTemplateDockable("Reentrant", gurps.NewTemplate())
	counter := &syncCounter{}
	counter.Self = counter
	counter.onSync = func() { template.MarkModified(nil) }
	template.AsPanel().AddChild(counter)

	template.MarkModified(nil)
	c.Equal(1, counter.count, "the call from inside the sync must be dropped")
	c.False(template.awaitingUpdate, "the guard must be lifted once the update is over")

	template.MarkModified(nil)
	c.Equal(2, counter.count, "the next call must be acted upon")
}

// TestLootSheetMarkModifiedBumpsTheTimestamp verifies that marking a loot sheet as modified records when it was
// changed, which is the one thing its MarkModified does beyond what the other page dockables' do.
func TestLootSheetMarkModifiedBumpsTheTimestamp(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	sheet.loot.ModifiedOn = jio.Time{}
	sheet.MarkModified(nil)
	c.NotEqual(jio.Time{}, sheet.loot.ModifiedOn, "marking the loot sheet as modified must bump the timestamp")
}

// TestTemplateRebuildKeepsTheScrollPosition verifies that rebuilding a template puts the page back where the user had
// scrolled it. A rebuild that has to replace a list -- here by hiding the TL column of a long equipment list -- hands
// the focus the old table held to its replacement, and a table scrolls itself into view as it takes the focus: with
// the page scrolled well down into the list, that on its own would jump the page back up to the top of the list.
func TestTemplateRebuildKeepsTheScrollPosition(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	settings := gurps.GlobalSettings().SheetSettings()
	saved := settings.HideTLColumn
	t.Cleanup(func() { screen.Do(func() { settings.HideTLColumn = saved }) })
	screen.Do(func() { settings.HideTLColumn = false })
	template, ok := openedByAction(t, screen, newCharacterTemplateAction).(*Template)
	if !ok {
		t.Fatal("New Character Template must open a template")
	}

	screen.Do(func() {
		items := make([]*gurps.Equipment, 0, 200)
		for i := range cap(items) {
			item := gurps.NewEquipment(nil, nil, false)
			item.Name = fmt.Sprintf("Item %d", i)
			items = append(items, item)
		}
		template.template.Equipment = items
		template.Rebuild(true)
	})

	// The page has been laid out again by now, so the scroll panel knows how far down it can go.
	const wanted = 300
	var stale *PageList[*gurps.Equipment]
	var focused bool
	var before float32
	screen.Do(func() {
		stale = template.Equipment
		stale.Table.RequestFocus()
		focused = wnd.Focus() == stale.Table.AsPanel()
		template.scroll.SetPosition(0, wanted)
		_, before = template.scroll.Position()
	})
	c.True(focused, "the equipment table takes the focus")
	c.Equal(float32(wanted), before, "the page must be long enough to scroll to where the test wants it")

	var replaced, refocused bool
	var after float32
	screen.Do(func() {
		settings.HideTLColumn = true
		template.Rebuild(true)
		replaced = template.Equipment != stale
		refocused = wnd.Focus() == template.Equipment.Table.AsPanel()
		_, after = template.scroll.Position()
	})
	c.True(replaced, "hiding the TL column must have replaced the equipment list")
	c.True(refocused, "the rebuild must hand the focus to the replacement")
	c.Equal(float32(wanted), after, "the rebuild must put the scroll position back")
}
