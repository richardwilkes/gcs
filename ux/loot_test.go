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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestTreasureGenPanelWithInRangeValues verifies that values already in range reach the fields unaltered.
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

// TestTreasureGenPanelWithOutOfRangeValues reproduces a crash when the stored loot generation values exceeded the
// fields' maximum. Creating a field clamps its value and reports it through the set callback while the dialog and the
// fields are still nil, so validating the OK button there panicked and the treasure generation dialog could never be
// opened.
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

// TestTreasureGenPanelValidateOKWithoutDialog verifies that bringing the OK button up to date before the dialog has
// been created, whether directly or through a field's validation, is a no-op rather than a panic.
func TestTreasureGenPanelValidateOKWithoutDialog(t *testing.T) {
	c := check.New(t)
	p := newTreasureGenPanel(fxp.Thousand, fxp.TenThousand)
	c.Nil(p.dialog, "no dialog yet")
	p.validateOK()
	p.minField.Validate()
	p.maxField.Validate()
}

// newTestLootSheet returns a loot sheet that can be built and rebuilt without a window, the way
// newTestSheetForTemplate does for a character sheet. Both the toolbar and the rebuild path reach for global state, so
// the key bindable actions are registered and a document dock is installed for the duration of the test.
func newTestLootSheet(t *testing.T) *LootSheet {
	t.Helper()
	registerKeyBindingsOnce.Do(func() { registerActions() })
	uxtest.SwapForTest(t, &Workspace.DocumentDock, NewDocumentDock())
	return NewLootSheet("test"+gurps.LootExt, gurps.NewLoot())
}

// TestLootSheetNewItemCommandUsesTheLiveList verifies that a loot sheet's "New Equipment" command adds its item to the
// list on screen rather than the one that existed when the sheet was created. LootSheet.createLists replaces any list
// whose columns no longer match what its provider asks for, and a command holding the original list would then create
// items in an orphan: not undoable, since an orphaned table can't find the undo manager, and neither selected nor
// scrolled into view on screen. Nothing in a loot sheet's data changes its columns today, so the mismatch is arranged
// here directly.
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

// TestLootSheetIgnoresAnotherEntitySheetSettings verifies that a loot sheet only responds to a change in the global
// sheet settings, which is what it reads from, and not to a change made in some character's per-sheet settings.
// Responding to the latter would rebuild the loot sheet for an edit that was never made to it.
func TestLootSheetIgnoresAnotherEntitySheetSettings(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	sheet.loot.ModifiedOn = jio.Time{}
	before := currentDisplayStamp()

	sheet.SheetSettingsUpdated(gurps.NewEntity(), gurps.SheetSettingsAllParts, true)
	c.Equal(jio.Time{}, sheet.loot.ModifiedOn, "another entity's settings change must not touch the loot sheet")
	c.Equal(before, currentDisplayStamp(), "another entity's settings change must not rebuild the loot sheet")
}

// TestLootSheetIgnoresTheDefaultsItDoesNotShow verifies that a change to the default block layout, attribute
// definitions or body type, none of which a loot sheet shows, leaves the loot sheet alone. Setting the default layout
// or applying the default attribute definitions used to fully rebuild every open loot sheet.
func TestLootSheetIgnoresTheDefaultsItDoesNotShow(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	before := currentDisplayStamp()

	sheet.SheetSettingsUpdated(nil, gurps.SheetSettingsLayoutPart|gurps.SheetSettingsAttributesPart|
		gurps.SheetSettingsBodyTypePart, true)
	c.Equal(before, currentDisplayStamp(), "a change to defaults the loot sheet doesn't show must not rebuild it")
}

// TestLootSheetShowsGlobalSheetSettingsWithoutBeingModified verifies that a change to the global sheet settings, which
// a loot sheet shows but doesn't save, rebuilds the loot sheet without bumping its modification timestamp. Applying the
// default attribute definitions or setting the default block layout used to leave every open loot sheet showing
// "Modified" as of that moment, and saving that date, for an edit never made to it.
func TestLootSheetShowsGlobalSheetSettingsWithoutBeingModified(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	sheet.loot.ModifiedOn = jio.Time{}

	for _, fullRebuild := range []bool{false, true} {
		before := currentDisplayStamp()
		sheet.SheetSettingsUpdated(nil, gurps.SheetSettingsGeneralPart, fullRebuild)
		c.Equal(jio.Time{}, sheet.loot.ModifiedOn,
			"a global settings change is no edit of the loot (fullRebuild=%v)", fullRebuild)
		c.True(sheet.Equipment.syncedAtStamp() > before,
			"the loot sheet's lists must be synced with the new settings (fullRebuild=%v)", fullRebuild)
		c.True(sheet.Notes.syncedAtStamp() > before,
			"the loot sheet's lists must be synced with the new settings (fullRebuild=%v)", fullRebuild)
	}
}

// TestLootCandidatesUseWhatGroupsHold verifies that treasure is generated from what a group holds, in nested groups
// too, rather than from the group itself, while a physical container is a candidate as it stands.
func TestLootCandidatesUseWhatGroupsHold(t *testing.T) {
	c := check.New(t)
	gems := gurps.NewEquipmentGroup(nil, nil)
	ruby := gurps.NewEquipment(nil, gems, false)
	ruby.Name = "Ruby"
	rare := gurps.NewEquipmentGroup(nil, gems)
	diamond := gurps.NewEquipment(nil, rare, false)
	diamond.Name = "Diamond"
	rare.Children = []*gurps.Equipment{diamond}
	gems.Children = []*gurps.Equipment{ruby, rare}
	chest := gurps.NewEquipment(nil, nil, true)
	chest.Name = "Chest"
	chest.Children = []*gurps.Equipment{gurps.NewEquipment(nil, chest, false)}
	candidates := lootCandidates([]*gurps.Equipment{gems, chest})
	names := make([]string, 0, len(candidates))
	for _, one := range candidates {
		names = append(names, one.Name)
	}
	c.Equal([]string{"Ruby", "Diamond", "Chest"}, names)
}

// TestTreasureGenPanelRecordsTheValidityOfAFieldBuiltOutOfRange checks that a field built from a stored value outside
// its range, which it shows flagged as invalid, is recorded as invalid as soon as its validity is gated, so that the OK
// button can't be enabled while the field shows a value other than the one generation would use.
func TestTreasureGenPanelRecordsTheValidityOfAFieldBuiltOutOfRange(t *testing.T) {
	c := check.New(t)
	p := newTreasureGenPanel(fxp.TenMillionMinusOne+fxp.One, fxp.TenThousand)
	c.True(p.minField.Invalid(), "the field shows the stored value flagged as invalid")
	c.True(p.invalid[p.minField], "and is recorded as invalid")
	c.False(p.maxField.Invalid())
	c.False(p.invalid[p.maxField])
	p = newTreasureGenPanel(fxp.Thousand, fxp.Half)
	c.True(p.maxField.Invalid())
	c.True(p.invalid[p.maxField], "a field built below its range is recorded as invalid too")
}
