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
	"os"
	"path/filepath"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// editingTableCmdIDs are the standard table commands that only an editable table offers.
var editingTableCmdIDs = []int{unison.DeleteItemID, DuplicateItemID, SyncWithSourceItemID, ClearSourceItemID}

// TestStandardTableCmdsReachableFromDockableFilterField verifies that the standard table commands of a list dockable
// are installed on the dockable rather than on its table, so that they still act on the selection while the filter
// field has the focus. Sync With Source and Clear Source used to be the odd ones out, installed on the table alone.
// Delete is left out because the field claims it for its own text.
func TestStandardTableCmdsReachableFromDockableFilterField(t *testing.T) {
	c := check.New(t)
	dockable := newTestTraitTableDockable()
	c.True(len(dockable.table.RootRows()) > 0, "the test list must have rows")
	ids := []int{OpenEditorItemID, DuplicateItemID, SyncWithSourceItemID, ClearSourceItemID}
	dockable.table.SelectByIndex(0)
	for _, id := range ids {
		c.True(dockable.filterField.CanPerformCmd(nil, id), "command %d must be reachable from the filter field", id)
	}
	dockable.table.ClearSelection()
	for _, id := range ids {
		c.False(dockable.filterField.CanPerformCmd(nil, id), "command %d must need a selection", id)
	}
}

// TestPageListEditingCmdsFollowOwner verifies that a page list with an owner offers the editing commands and that a
// read-only page list -- one derived from the rest of the sheet, which has no owner -- offers only the opening ones.
func TestPageListEditingCmdsFollowOwner(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	trait.Weapons = []*gurps.Weapon{gurps.NewWeapon(trait, true)}
	entity.Traits = []*gurps.Trait{trait}

	traits := NewTraitsPageList(sheet, entity)
	c.True(len(traits.Table.RootRows()) > 0, "the traits list must show the trait")
	traits.Table.SelectByIndex(0)
	c.True(traits.Table.CanPerformCmd(nil, OpenEditorItemID), "an owned list must offer Open Editor")
	for _, id := range editingTableCmdIDs {
		c.True(traits.Table.CanPerformCmd(nil, id), "an owned list must offer command %d", id)
	}

	weapons := NewMeleeWeaponsPageList(entity)
	c.True(len(weapons.Table.RootRows()) > 0, "the weapons list must show the trait's weapon")
	weapons.Table.SelectByIndex(0)
	c.True(weapons.Table.CanPerformCmd(nil, OpenEditorItemID), "a read-only list must still offer Open Editor")
	for _, id := range editingTableCmdIDs {
		c.False(weapons.Table.CanPerformCmd(nil, id), "a read-only list must not offer command %d", id)
	}
}

// watchedListOwner is a data owner without a source matcher, as an equipment list file is, that calls onSourceMatcher
// whenever it is asked for one.
type watchedListOwner struct {
	onSourceMatcher func()
}

func (w *watchedListOwner) OwningEntity() *gurps.Entity { return nil }

func (w *watchedListOwner) SourceMatcher() *gurps.SrcMatcher {
	w.onSourceMatcher()
	return nil
}

func (w *watchedListOwner) WeightUnit() fxp.WeightUnit { return fxp.Pound }

// TestSyncWithSourceForSelectionLoadsEachFileOnce verifies that syncing the selected rows of a library list loads the
// library file they are sourced from just once for the whole selection, rather than checking it again for each row, by
// taking the file away as soon as the first row has been synced, and that the file is checked again for a row matched
// once the sync is done.
func TestSyncWithSourceForSelectionLoadsEachFileOnce(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Test" + gurps.NotesExt}
	p := filepath.Join(user.Path(false), libFile.Path)
	libNotes := []*gurps.Note{gurps.NewNote(nil, nil, false), gurps.NewNote(nil, nil, false)}
	libNotes[0].MarkDown = "First"
	libNotes[1].MarkDown = "Second"
	c.NoError(gurps.SaveNotes(libNotes, p))
	owner := &watchedListOwner{}
	notes := make([]*gurps.Note, len(libNotes))
	for i, one := range libNotes {
		notes[i] = one.Clone(libFile, owner, nil, gurps.Reference)
		notes[i].MarkDown += " (old)"
	}
	// Each row asks its data owner for a matcher as it is synced, which is when a check of its file would be made.
	owner.onSourceMatcher = func() {
		if notes[0].MarkDown == "First" {
			if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
				t.Error(err)
			}
		}
	}
	dockable := NewNoteTableDockable("test"+gurps.NotesExt, notes)
	dockable.table.SelectAll()

	SyncWithSourceForSelection(dockable.table)
	_, err := os.Stat(p)
	c.True(os.IsNotExist(err), "precondition: the file was taken away during the sync")
	c.Equal("First", notes[0].MarkDown)
	c.Equal("Second", notes[1].MarkDown, "a row after the first is synced with the file as loaded for the selection")

	notes[0].MarkDown = "First (old)"
	state, _ := gurps.MatchSource(notes[0])
	c.Equal(srcstate.Missing, state, "the file is checked again for a row matched once the sync is done")
}
