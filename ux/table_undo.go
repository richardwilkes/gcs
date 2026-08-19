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
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// TableUndoEditData holds the data necessary to provide undo for a table.
type TableUndoEditData[T gurps.Node[T]] struct {
	Table *unison.Table[*Node[T]]
	Data  PreservedTableData[T]
}

// NewTableUndoEditData collects the undo edit data for a table.
func NewTableUndoEditData[T gurps.Node[T]](table *unison.Table[*Node[T]]) *TableUndoEditData[T] {
	if table == nil {
		return nil
	}
	table = liveTable(table)
	undo := &TableUndoEditData[T]{Table: table}
	if err := undo.Data.Collect(table); err != nil {
		errs.Log(err)
		return nil
	}
	return undo
}

// Apply the undo edit data to a table.
func (t *TableUndoEditData[T]) Apply() {
	var restored restoredTables
	restored.add(t.restore())
	restored.report()
}

// restore puts the preserved data back into the table that is currently showing it, without reporting the change to
// anything, and returns that table along with the owner that has to rebuild its lists for the table's columns to match
// what its provider now calls for, or nil if the columns already match. Nothing is returned if there was nothing to
// restore or the restore failed. The reporting is left to restoredTables, so that an undo spanning several tables can
// do it just once for all of them.
func (t *TableUndoEditData[T]) restore() (*unison.Table[*Node[T]], Rebuildable) {
	if t == nil {
		return nil, nil
	}
	table := liveTable(t.Table)
	if err := t.Data.Restore(table); err != nil {
		errs.Log(err)
		return nil, nil
	}
	return table, ownerNeedingRebuildFor(table)
}

// restoredTables accumulates the results of restoring one or more tables, so that the undo they belong to reports the
// change exactly once no matter how many tables it had to put back.
type restoredTables struct {
	table unison.Paneler
	owner Rebuildable
}

// add records what restoring one table produced, i.e. the results of a TableUndoEditData.restore() call.
func (r *restoredTables) add(table unison.Paneler, ownerNeedingRebuild Rebuildable) {
	if xreflect.IsNil(r.table) && !xreflect.IsNil(table) {
		r.table = table
	}
	if xreflect.IsNil(r.owner) && !xreflect.IsNil(ownerNeedingRebuild) {
		r.owner = ownerNeedingRebuild
	}
}

// report tells the rest of the app about the restored data.
func (r *restoredTables) report() {
	// Some columns are only present when the data calls for them (the switch column, for one), so putting the old data
	// back may require the owner to rebuild its lists to bring such a column into or out of view. Which of the two is
	// needed can't be known until the data is back in place, which is why restoring doesn't report the change on its
	// own: a rebuild already recalculates the entity, re-syncs every table, refreshes the search results and restores
	// the focus and scroll position, i.e. all that marking a table as modified would do apart from bumping the owner's
	// modification timestamp, which is therefore done here as well. So exactly one of the two is done, once for the
	// whole undo rather than once per table, since on a sheet that may hold hundreds of rows all of that work is the
	// entire cost of the edit and an undo spanning six tables would otherwise pay it six times over.
	if !xreflect.IsNil(r.owner) {
		if bumper, ok := r.owner.(modificationTimestampBumper); ok {
			// The bump has to happen before the rebuild for the panel showing it to pick up the new value.
			bumper.bumpModificationTimestamp()
		}
		r.owner.Rebuild(true)
		return
	}
	if !xreflect.IsNil(r.table) {
		MarkModified(r.table)
	}
}

// modificationTimestampBumper is implemented by those owners that record when their data was last changed. Marking such
// an owner as modified bumps that timestamp, but rebuilding it doesn't, so anything that rebuilds in place of marking
// as modified has to ask for the bump itself. Not every owner has one -- a template's modification time is whatever the
// file system says it is -- which is why this is an optional interface rather than part of Rebuildable.
type modificationTimestampBumper interface {
	bumpModificationTimestamp()
}

var (
	_ modificationTimestampBumper = &Sheet{}
	_ modificationTimestampBumper = &LootSheet{}
)

// ownerNeedingRebuildFor returns the owner that has to rebuild its lists for the given table to show the columns its
// provider now calls for, or nil if the columns already match or there is no owner to ask.
func ownerNeedingRebuildFor[T gurps.Node[T]](table *unison.Table[*Node[T]]) Rebuildable {
	provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T])
	if !ok || !columnsOutOfSync(provider.ColumnIDs(), table.Columns) {
		return nil
	}
	owner, found := table.ClientData()[TableOwnerClientKey].(Rebuildable)
	if !found || xreflect.IsNil(owner) {
		return nil
	}
	return owner
}

// liveTable returns the table that is currently showing the data the given table was created for. An owner that has to
// alter its set of columns can only do so by replacing the table entirely, which leaves any table captured earlier
// orphaned: applying data to it would update the model but leave the table the user is looking at untouched. A nil
// table is returned as-is, since callers that may not have a source table at all (the alternate drop path, for one)
// pass one through here. This is the typed counterpart of liveOwner, for callers that need the table itself rather
// than just something to ask for a rebuild through.
func liveTable[T gurps.Node[T]](table *unison.Table[*Node[T]]) *unison.Table[*Node[T]] {
	if table == nil {
		return nil
	}
	if current, ok := liveOwner(table).(*unison.Table[*Node[T]]); ok {
		return current
	}
	return table
}

// TableDragUndoEditData holds the undo edit data for a table drag.
type TableDragUndoEditData[T gurps.Node[T]] struct {
	From *TableUndoEditData[T]
	To   *TableUndoEditData[T]
}

// NewTableDragUndoEditData collects the undo edit data for a table drag.
func NewTableDragUndoEditData[T gurps.Node[T]](from, to *unison.Table[*Node[T]]) *TableDragUndoEditData[T] {
	return &TableDragUndoEditData[T]{
		From: NewTableUndoEditData(from),
		To:   NewTableUndoEditData(to),
	}
}

// Apply the undo edit data to a table.
func (t *TableDragUndoEditData[T]) Apply() {
	// Both lists have to be back in place before either of them is reported: a row that was dragged from one list to
	// the other is in neither of them while only the first has been restored, so reporting there would recalculate the
	// entity and re-sync the whole sheet against a state the user never had, and the second restore would then pay for
	// the same update all over again. One report covers both, because the two tables always belong to the same owner.
	// A drag only counts as a move -- the only case that leaves anything to put back in the list the rows came from,
	// which is why From is nil for the rest -- when both tables are in the same dockable, and only equipment can move
	// between two different tables at all (see the providers' DropShouldMoveData). So a drag between two sheets is a
	// copy that leaves the source sheet untouched, and "Move to Other/Carried Equipment" works within one sheet by
	// construction. A nil From restores nothing and contributes nothing to report.
	var restored restoredTables
	restored.add(t.To.restore())
	restored.add(t.From.restore())
	restored.report()
}
