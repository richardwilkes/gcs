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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// A template's equipment list follows the global sheet settings for its TL and LC columns, which the user can change
// while the template is open; a list that was merely synced would go on showing the old columns, and a command that
// captured the list when the template was created would be creating items in an orphan afterwards -- the model would
// gain the item, but it would be neither selected in the list on screen nor undoable, since an orphaned table can't
// find the undo manager.
func TestTemplateReplacesAListWhoseColumnsChanged(t *testing.T) {
	c := check.New(t)
	settings := gurps.GlobalSettings().SheetSettings()
	swapForTest(t, &settings.HideTLColumn, false)
	// Creating an item opens its editor, which looks for a dock to go into.
	swapForTest(t, &Workspace.DocumentDock, NewDocumentDock())

	data := gurps.NewTemplate()
	template := newTestTemplateDockable("My Template", data)
	stale := template.Equipment
	c.True(columnsMatchProvider(stale.Table), "the equipment list must start out with the columns its provider wants")
	hasTL := func(table *unison.Table[*Node[*gurps.Equipment]]) bool {
		for _, col := range table.Columns {
			if col.ID == gurps.EquipmentTLColumn {
				return true
			}
		}
		return false
	}
	c.True(hasTL(stale.Table), "the TL column must be shown while the setting says to show it")

	settings.HideTLColumn = true
	template.Rebuild(true)
	c.NotEqual(stale, template.Equipment, "hiding the TL column must have replaced the equipment list")
	c.True(columnsMatchProvider(template.Equipment.Table), "the replacement must have the columns its provider wants")
	c.False(hasTL(template.Equipment.Table), "the replacement must not have the TL column")

	mgr := template.UndoManager()
	c.NotNil(mgr, "the template must have an undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")
	template.AsPanel().PerformCmd(nil, NewCarriedEquipmentItemID)
	c.Equal(1, len(data.Equipment), "the command must have added an item to the template")
	c.Equal(1, template.Equipment.Table.RootRowCount(), "the new row must be in the list that is on screen")
	c.True(template.Equipment.Table.CopySelectionMap()[data.Equipment[0].ID()],
		"the new row must be selected in the list that is on screen")
	c.True(mgr.CanUndo(), "creating an item must be undoable")
	mgr.Undo()
	c.Equal(0, len(data.Equipment), "undo must take the new item back out of the template")
	c.Equal(0, template.Equipment.Table.RootRowCount(), "undo must take the row back out of the list on screen")
}

// The template used to note which list the focus was in and put it back by hand; the rebuild's ordinary focus
// restoration finds the replacement table by the reference key it shares with the one it replaced, so that is no
// longer needed.
func TestTemplateRebuildKeepsTheFocusInAReplacedList(t *testing.T) {
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

	var stale *PageList[*gurps.Equipment]
	var focused bool
	screen.Do(func() {
		stale = template.Equipment
		stale.Table.RequestFocus()
		focused = wnd.Focus() == stale.Table.AsPanel()
	})
	c.True(focused, "the equipment table takes the focus")

	var replaced, refocused bool
	screen.Do(func() {
		settings.HideTLColumn = true
		template.Rebuild(true)
		replaced = template.Equipment != stale
		refocused = wnd.Focus() == template.Equipment.Table.AsPanel()
	})
	c.True(replaced, "hiding the TL column must have replaced the equipment list")
	c.True(refocused, "the focus must have moved into the replacement")
}

// A template's search, like a sheet's, only looks in the lists that are on the page: a match in a list the layout
// hides could only be shown by scrolling a table nobody is looking at into view.
func TestTemplateSearchSkipsAListTheLayoutDoesNotShow(t *testing.T) {
	c := check.New(t)
	sheetSettings := gurps.GlobalSettings().Sheet
	swapForTest(t, &sheetSettings.Layout, sheetSettings.Layout.Clone())

	data := gurps.NewTemplate()
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Findable"
	data.Traits = []*gurps.Trait{trait}
	data.Skills = []*gurps.Skill{newTestSkill("Findable", fxp.One, nil)}
	template := newTestTemplateDockable("Search", data)
	var refs []*searchRef
	template.searchTracker.findMatches(&refs, "findable", true)
	c.Equal(2, len(refs), "with every list on the page, both rows are found")

	c.True(sheetSettings.Layout.Hide(gurps.BlockSkillsKey))
	template.Rebuild(true)
	c.Nil(template.Skills.AsPanel().Parent(), "the hidden skills list is not on the page")
	refs = nil
	template.searchTracker.findMatches(&refs, "findable", true)
	c.Equal(1, len(refs), "only the row in the list that is on the page is found")
	c.Equal(any(template.Traits.Table), refs[0].table, "and it is the trait")
}
