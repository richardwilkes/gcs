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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestSyncOrRebuildList verifies that a page list is built when it doesn't exist, kept and synced while its columns
// still match what its provider wants, and built anew once they no longer do, with the caller's field naming the list
// that resulted each time.
func TestSyncOrRebuildList(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	built := 0
	var list *PageList[*gurps.Note]
	build := func() *PageList[*gurps.Note] {
		built++
		return NewNotesPageList(sheet, sheet.loot)
	}

	first := syncOrRebuildList(&list, build)
	c.Equal(1, built, "a list that doesn't exist is built")
	c.NotNil(first, "and returned")
	c.Equal(first, list, "and stored through the pointer")

	sheet.loot.Notes = []*gurps.Note{gurps.NewNote(nil, nil, false)}
	c.Equal(0, list.RowCount(), "the list doesn't show the note yet")
	c.Equal(first, syncOrRebuildList(&list, build), "a list whose columns still match is kept")
	c.Equal(1, built, "rather than built again")
	c.Equal(1, list.RowCount(), "and is synced")

	list.Table.Columns = append(list.Table.Columns, unison.ColumnInfo{ID: -1})
	second := syncOrRebuildList(&list, build)
	c.Equal(2, built, "a list whose columns no longer match is built anew")
	c.NotEqual(first, second, "so a different list results")
	c.Equal(second, list, "which is stored through the pointer")
	c.Equal(1, list.RowCount(), "and shows the model")
}

// TestPreserveSelectionsFollowsAReplacedList verifies that a selection recorded from a list is put back into whatever
// list stands in its place when the selections are restored, since a rebuild replaces a list whose columns changed
// and the selection belongs in the list that is on screen. A list that hasn't been built yet is tolerated in both
// directions.
func TestPreserveSelectionsFollowsAReplacedList(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	note := gurps.NewNote(nil, nil, false)
	sheet.loot.Notes = []*gurps.Note{note}
	sheet.Notes.Sync()
	sheet.Notes.Table.SetSelectionMap(map[tid.TID]bool{note.ID(): true})
	var unbuilt *PageList[*gurps.Equipment]
	lists := []sheetList{unbuilt, sheet.Notes}

	restore := preserveSelections(func() []sheetList { return lists })
	replacement := NewNotesPageList(sheet, sheet.loot)
	c.False(replacement.Table.HasSelection(), "a fresh list starts out with nothing selected")
	lists[1] = replacement
	restore()
	c.True(replacement.Table.CopySelectionMap()[note.ID()], "the selection is put back into the replacement")
	c.True(sheet.Notes.Table.CopySelectionMap()[note.ID()], "and the list it was recorded from is left as it was")
}

// TestListsForKeysFollowTheCanonicalBlockOrder verifies that the lists a dockable works through come in the canonical
// block order, filtered down to the keys the dockable's predicate accepts.
func TestListsForKeysFollowTheCanonicalBlockOrder(t *testing.T) {
	c := check.New(t)
	var keys []string
	lists := listsForKeys(func(key string) sheetList {
		keys = append(keys, key)
		return nil
	}, gurps.IsTemplateBlockKey)
	expected := []string{
		gurps.BlockTraitsKey,
		gurps.BlockSkillsKey,
		gurps.BlockSpellsKey,
		gurps.BlockEquipmentKey,
		gurps.BlockNotesKey,
	}
	c.Equal(expected, keys, "the keys the predicate accepts, in canonical order")
	c.Equal(len(expected), len(lists), "one list per accepted key")
}
