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

// TestLootSheetIgnoresTheDefaultsItDoesNotShow verifies that a change to the default block layout, attribute
// definitions or body type, none of which a loot sheet shows, leaves the loot sheet alone, as does any change made in
// some character's per-sheet settings, since a loot sheet reads only the global ones. Setting the default layout or
// applying the default attribute definitions used to fully rebuild every open loot sheet.
func TestLootSheetIgnoresTheDefaultsItDoesNotShow(t *testing.T) {
	for _, one := range []struct {
		name   string
		entity *gurps.Entity
		parts  gurps.SheetSettingsParts
	}{
		{
			name:  "defaults the loot sheet doesn't show",
			parts: gurps.SheetSettingsLayoutPart | gurps.SheetSettingsAttributesPart | gurps.SheetSettingsBodyTypePart,
		},
		{
			name:   "another entity's settings",
			entity: gurps.NewEntity(),
			parts:  gurps.SheetSettingsAllParts,
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestLootSheet(t)
			sheet.loot.ModifiedOn = jio.Time{}
			before := currentDisplayStamp()

			sheet.SheetSettingsUpdated(one.entity, one.parts, true)
			c.Equal(jio.Time{}, sheet.loot.ModifiedOn, "the change must not touch the loot sheet")
			c.Equal(before, currentDisplayStamp(), "the change must not rebuild the loot sheet")
		})
	}
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

// TestTreasureGenPanelRecordsTheValidityOfAFieldBuiltOutOfRange checks that a field built from a stored value in its
// range takes it unaltered, while one built from a value outside its range clamps the value into it, shows the stored
// value flagged as invalid, and is recorded as invalid as soon as its validity is gated, so that the OK button can't be
// enabled while the field shows a value other than the one generation would use. The clamped value is reported while
// the dialog and the fields are still nil, which once panicked, so that the treasure generation dialog could never be
// opened once the stored values exceeded the fields' maximum.
func TestTreasureGenPanelRecordsTheValidityOfAFieldBuiltOutOfRange(t *testing.T) {
	c := check.New(t)

	// A value above the field maximum survives loading the settings, so it can reach the dialog.
	settings := gurps.Settings{
		LootGenMinValue: fxp.TenMillionMinusOne + fxp.One,
		LootGenMaxValue: fxp.TenMillionMinusOne + fxp.Two,
	}
	settings.EnsureValidity()
	c.True(settings.LootGenMinValue > fxp.TenMillionMinusOne, "minimum value is left above the field maximum")
	c.True(settings.LootGenMaxValue > fxp.TenMillionMinusOne, "maximum value is left above the field maximum")

	for _, one := range []struct {
		name       string
		minValue   fxp.Int
		maxValue   fxp.Int
		wantMin    fxp.Int
		wantMax    fxp.Int
		minInvalid bool
		maxInvalid bool
	}{
		{
			name:     "in range",
			minValue: fxp.Thousand,
			maxValue: fxp.TenThousand,
			wantMin:  fxp.Thousand,
			wantMax:  fxp.TenThousand,
		},
		{
			name:       "minimum above range",
			minValue:   fxp.TenMillionMinusOne + fxp.One,
			maxValue:   fxp.TenThousand,
			wantMin:    fxp.TenMillionMinusOne,
			wantMax:    fxp.TenThousand,
			minInvalid: true,
		},
		{
			name:       "both above range, from the settings",
			minValue:   settings.LootGenMinValue,
			maxValue:   settings.LootGenMaxValue,
			wantMin:    fxp.TenMillionMinusOne,
			wantMax:    fxp.TenMillionMinusOne,
			minInvalid: true,
			maxInvalid: true,
		},
		{
			name:       "maximum below range",
			minValue:   fxp.Thousand,
			maxValue:   fxp.Half,
			wantMin:    fxp.Thousand,
			wantMax:    fxp.One,
			maxInvalid: true,
		},
		{
			name:       "both below range",
			minValue:   fxp.OneHundredth,
			maxValue:   fxp.OneHundredth,
			wantMin:    fxp.One,
			wantMax:    fxp.One,
			minInvalid: true,
			maxInvalid: true,
		},
	} {
		p := newTreasureGenPanel(one.minValue, one.maxValue)
		c.Equal(one.wantMin, p.minValue, "%s: minimum value clamped into range", one.name)
		c.Equal(one.wantMax, p.maxValue, "%s: maximum value clamped into range", one.name)
		c.Equal(one.minValue.String(), p.minField.Text(), "%s: the minimum field shows the stored value", one.name)
		c.Equal(one.maxValue.String(), p.maxField.Text(), "%s: the maximum field shows the stored value", one.name)
		c.Equal(one.minInvalid, p.minField.Invalid(), "%s: minimum field validity", one.name)
		c.Equal(one.minInvalid, p.invalid[p.minField], "%s: minimum field validity is recorded", one.name)
		c.Equal(one.maxInvalid, p.maxField.Invalid(), "%s: maximum field validity", one.name)
		c.Equal(one.maxInvalid, p.invalid[p.maxField], "%s: maximum field validity is recorded", one.name)
	}
}
