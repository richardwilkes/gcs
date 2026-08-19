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

// TestTreasureGenPanelWithInRangeValues verifies that the fields are primed with the values they were given and that
// nothing is altered when those values are already within the range the fields allow.
func TestTreasureGenPanelWithInRangeValues(t *testing.T) {
	c := check.New(t)
	p := newTreasureGenPanel(fxp.Thousand, fxp.TenThousand)
	c.Equal(fxp.Thousand, p.minValue, "minimum value")
	c.Equal(fxp.TenThousand, p.maxValue, "maximum value")
	c.Equal(fxp.Thousand.String(), p.minField.Text(), "minimum field text")
	c.Equal(fxp.TenThousand.String(), p.maxField.Text(), "maximum field text")
	c.False(p.minField.Invalid(), "minimum field validity")
	c.False(p.maxField.Invalid(), "maximum field validity")
}

// TestTreasureGenPanelWithOutOfRangeValues reproduces the crash that occurred when the stored loot generation values
// were larger than the maximum the fields accept. Creating the field clamps the value into range and reports it through
// the set callback, which runs while the dialog and the fields themselves are still nil, so validating the OK button at
// that point panicked with a nil pointer dereference and the treasure generation dialog could never be opened.
func TestTreasureGenPanelWithOutOfRangeValues(t *testing.T) {
	c := check.New(t)

	// A value above the field maximum survives loading the settings, so it can reach the dialog.
	settings := gurps.Settings{
		LootGenMinValue: fxp.TenMillionMinusOne + fxp.One,
		LootGenMaxValue: fxp.TenMillionMinusOne + fxp.Two,
	}
	settings.EnsureValidity()
	c.True(settings.LootGenMinValue > fxp.TenMillionMinusOne, "minimum value is left above the field maximum")
	c.True(settings.LootGenMaxValue > fxp.TenMillionMinusOne, "maximum value is left above the field maximum")

	p := newTreasureGenPanel(settings.LootGenMinValue, settings.LootGenMaxValue)
	c.Equal(fxp.TenMillionMinusOne, p.minValue, "minimum value clamped into range")
	c.Equal(fxp.TenMillionMinusOne, p.maxValue, "maximum value clamped into range")

	// The same clamping happens for values below the field minimum.
	p = newTreasureGenPanel(fxp.OneHundredth, fxp.OneHundredth)
	c.Equal(fxp.One, p.minValue, "minimum value clamped up into range")
	c.Equal(fxp.One, p.maxValue, "maximum value clamped up into range")
}

// TestTreasureGenPanelValidateOKWithoutDialog verifies that asking for the OK button state before the dialog has been
// created is a no-op rather than a panic, which is the state the fields are in while they are being created.
func TestTreasureGenPanelValidateOKWithoutDialog(t *testing.T) {
	c := check.New(t)
	p := newTreasureGenPanel(fxp.Thousand, fxp.TenThousand)
	c.Nil(p.dialog, "no dialog yet")
	p.validateOK()
	p.minField = nil
	p.maxField = nil
	p.validateOK()
}

// newTestLootSheet returns a loot sheet that can be built and rebuilt without a window, the way
// newTestSheetForTemplate does for a character sheet. Both the toolbar and the rebuild path reach for global state, so
// the key bindable actions are registered and a document dock is installed for the duration of the test.
func newTestLootSheet(t *testing.T) *LootSheet {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	saved := Workspace.DocumentDock
	t.Cleanup(func() { Workspace.DocumentDock = saved })
	Workspace.DocumentDock = NewDocumentDock()
	return NewLootSheet("test"+gurps.LootExt, gurps.NewLoot())
}

// TestLootSheetNewItemCommandUsesTheLiveList verifies that a loot sheet's "New Equipment" command adds its item to the
// list the user is looking at rather than to whichever list existed when the sheet was created. LootSheet.createLists
// replaces any list whose columns no longer match what its provider asks for, exactly as a character sheet does, and a
// command holding on to the list it was handed at construction would afterwards be creating items in an orphan: the
// insertion wouldn't be undoable, since an orphaned table can't find the undo manager, and the new row would be
// neither selected nor scrolled into view in the list that is on screen. Nothing in a loot sheet's data calls for a
// different set of columns today -- the switch column is reserved for character sheets -- so the mismatch that drives
// the replacement is arranged here directly.
func TestLootSheetNewItemCommandUsesTheLiveList(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	stale := sheet.Equipment
	mgr := sheet.UndoManager()
	c.NotNil(mgr, "the loot sheet must have an undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")

	sheet.Equipment.Table.Columns = append(sheet.Equipment.Table.Columns, unison.ColumnInfo{ID: -1})
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Equipment, "a column set that no longer matches must have replaced the equipment list")

	sheet.AsPanel().PerformCmd(nil, NewOtherEquipmentItemID)
	c.Equal(1, len(sheet.loot.Equipment), "the command must have added a piece of equipment to the loot")
	added := sheet.loot.Equipment[0]
	c.Equal(1, sheet.Equipment.Table.RootRowCount(), "the new row must be in the list that is on screen")
	c.True(sheet.Equipment.Table.CopySelectionMap()[added.ID()],
		"the new row must be selected in the list that is on screen")

	c.True(mgr.CanUndo(), "creating an item must be undoable")
	mgr.Undo()
	c.Equal(0, len(sheet.loot.Equipment), "undo must take the new equipment back out of the loot")
	c.Equal(0, sheet.Equipment.Table.RootRowCount(), "undo must take the row back out of the list that is on screen")
}
