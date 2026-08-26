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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// organizeTraits files the table's top-level traits into the standard set of containers as one undoable step. It works
// for any traits table -- a sheet's or template's page list as well as a trait library list -- since everything it
// needs comes from the table's own provider. Nothing is recorded and nothing is marked as modified when the traits are
// already organized.
func organizeTraits(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	provider, ok := any(table.Model).(TableProvider[*gurps.Trait])
	if !ok {
		return
	}
	// The "before" snapshot has to be taken while the list is still the one the user is looking at. Organizing
	// reparents traits into containers that are themselves still in the old top-level slice, so a snapshot taken
	// afterwards -- the way Sheet.syncWithAllSources builds its edit inline (ux/sheet.go) -- would walk each moved
	// trait twice, once where it still sits at the top level and once inside the container it was just filed into,
	// and the undo would put two copies of every moved row back. syncWithAllSources gets away with building its
	// "before" data after the fact only because syncing alters traits in place and moves nothing.
	mgr := unison.UndoManagerFor(table)
	var before *TableUndoEditData[*gurps.Trait]
	if mgr != nil {
		before = NewTableUndoEditData(table)
	}
	organized, changed := gurps.OrganizeTraits(provider.DataOwner(), provider.RootData())
	if !changed {
		// The list was already organized, so there is nothing to record and nothing to report: an edit here would
		// leave the user with an undo that does nothing and a document marked as modified for no change.
		return
	}
	provider.SetRootData(organized)
	table.SyncToModel()
	if mgr != nil && before != nil {
		mgr.Add(&unison.UndoEdit[*TableUndoEditData[*gurps.Trait]]{
			ID:         unison.NextUndoID(),
			EditName:   organizeTraitsAction.Title,
			UndoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Trait]]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*TableUndoEditData[*gurps.Trait]]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[*gurps.Trait]], _ unison.Undoable) bool { return false },
			BeforeData: before,
			AfterData:  NewTableUndoEditData(table),
		})
	}
	// A rebuild is the report, rather than also marking the owner as modified, since it is a superset of that and
	// doing both would repeat the whole update (see rebuildAsModified).
	rebuildAsModified(owner, true)
}
