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
	// The "before" snapshot has to be taken first. Organizing reparents traits into containers that are themselves
	// still in the old top-level slice, so a snapshot taken afterwards would walk each moved trait twice, once where it
	// still sits at the top level and once inside the container it was just filed into, and the undo would put two
	// copies of every moved row back.
	//
	// Filing a trait into a disabled container takes away the grants it carries, along with the rows they put in the
	// other lists, so those lists are part of the edit; see otherGrantRowLists. Both checks are made on the old
	// top-level slice, which still holds every trait afterwards, so those lists are captured both times or neither.
	list := provider.RootData()
	undo := beginTableUndoWith(table, organizeTraitsAction.Title, otherGrantRowLists(table, list))
	organized, changed := gurps.OrganizeTraits(provider.DataOwner(), list)
	if !changed {
		// The list was already organized, so there is nothing to record and nothing to report: an edit here would
		// leave the user with an undo that does nothing and a document marked as modified for no change.
		return
	}
	provider.SetRootData(organized)
	table.SyncToModel()
	commitTableUndoWith(table, undo, func() []tableRestorer { return otherGrantRowLists(table, list) })
	// A rebuild is the report, rather than also marking the owner as modified, since it is a superset of that and
	// doing both would repeat the whole update (see rebuildAsModified).
	rebuildAsModified(owner, true)
}
