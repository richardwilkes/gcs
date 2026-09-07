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

// beginTableUndo starts recording an undoable edit of a table's contents, capturing the data as it stands as the edit's
// "before" state; commitTableUndo captures the "after" state once the edit has been made and records the edit. Nil is
// returned when there is nothing to record with, i.e. the table has no undo manager, or when the "before" state could
// not be captured, and commitTableUndo accepts nil, so an edit can be bracketed by the pair unconditionally. The
// optional beforeUndo and beforeRedo hooks run ahead of putting the "before" and "after" data back, respectively, for
// an edit that has to restore more than the table's contents (see MoveSelection).
func beginTableUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], title string, beforeUndo, beforeRedo func()) *unison.UndoEdit[*TableUndoEditData[T]] {
	if unison.UndoManagerFor(table) == nil {
		return nil
	}
	before := NewTableUndoEditData(table)
	if before == nil {
		return nil
	}
	return &unison.UndoEdit[*TableUndoEditData[T]]{
		ID:       unison.NextUndoID(),
		EditName: title,
		UndoFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]]) {
			if beforeUndo != nil {
				beforeUndo()
			}
			e.BeforeData.Apply()
		},
		RedoFunc: func(e *unison.UndoEdit[*TableUndoEditData[T]]) {
			if beforeRedo != nil {
				beforeRedo()
			}
			e.AfterData.Apply()
		},
		AbsorbFunc: func(_ *unison.UndoEdit[*TableUndoEditData[T]], _ unison.Undoable) bool { return false },
		BeforeData: before,
	}
}

// commitTableUndo finishes an edit begun with beginTableUndo: the table's data as it now stands becomes the edit's
// "after" state and the edit is handed to the table's undo manager. Both are taken from the table currently showing the
// data (see liveTable), since the edit itself may have replaced the table the "before" state was captured from, and an
// orphaned table has no manager above it any more. A nil edit, which is what beginTableUndo hands back when there is
// nothing to record with, is ignored.
func commitTableUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], undo *unison.UndoEdit[*TableUndoEditData[T]]) {
	if undo == nil {
		return
	}
	table = liveTable(table)
	if mgr := unison.UndoManagerFor(table); mgr != nil {
		undo.AfterData = NewTableUndoEditData(table)
		mgr.Add(undo)
	}
}

// Apply the undo edit data to a table.
func (t *TableUndoEditData[T]) Apply() {
	var restored restoredTables
	restored.add(t.restore())
	restored.report()
}

// restore puts the preserved data back into the table that is currently showing it, without reporting the change to
// anything, and returns that table along with the owner it belongs to, if any. It has to be the table on screen rather
// than the one the data was collected from: the model would come back either way, since an orphaned table's provider
// still points at it, but the selection would not, because the rebuild that follows records and re-applies the
// selection of the table on screen and would discard one restored into an orphan. Nothing is returned if there was
// nothing to restore or the restore failed. The reporting is left to restoredTables, so that an undo spanning several
// tables can do it just once for all of them. The table comes back as a plain Paneler, which is all restoredTables
// needs and what lets an undo spanning tables of different row types hold their data behind one interface (see
// tableRestorer).
func (t *TableUndoEditData[T]) restore() (unison.Paneler, Rebuildable) {
	if t == nil {
		return nil, nil
	}
	table := liveTable(t.Table)
	if err := t.Data.Restore(table); err != nil {
		errs.Log(err)
		return nil, nil
	}
	owner, found := table.ClientData()[TableOwnerClientKey].(Rebuildable)
	if !found {
		return table, nil
	}
	return table, owner
}

// tableRestorer is what tablesUndoData needs of each table's undo data: TableUndoEditData.restore(), stripped of the
// row type so that the data for tables of different row types can be held together.
type tableRestorer interface {
	restore() (unison.Paneler, Rebuildable)
}

// syncableList is what "Sync With All Sources" needs of a page list, independent of the list's row type: undo data for
// its table and a way to bring the table up to date with its model. Every *PageList[T] provides both.
type syncableList interface {
	undoData() tableRestorer
	syncToModel()
}

// librarySyncer is the model behind a sheet, template or loot sheet, which knows how to sync everything it holds with
// the library sources the items came from.
type librarySyncer interface {
	SyncWithLibrarySources()
}

// tablesUndoData holds the undo data for an edit that spans several page lists at once, such as "Sync With All
// Sources", which can alter every sourced list a document has.
type tablesUndoData struct {
	restorers []tableRestorer
}

// newTablesUndoData collects the undo edit data for each of the given lists. A list whose data can't be collected
// contributes nothing, just as it would to a single-table edit (see NewTableUndoEditData).
func newTablesUndoData(lists []syncableList) *tablesUndoData {
	data := &tablesUndoData{restorers: make([]tableRestorer, 0, len(lists))}
	for _, list := range lists {
		if restorer := list.undoData(); restorer != nil {
			data.restorers = append(data.restorers, restorer)
		}
	}
	return data
}

// Apply the undo edit data to the tables.
func (d *tablesUndoData) Apply() {
	// Every list is put back before any of them is reported, so that the undo updates the document once rather than
	// once per table: a single rebuild brings all of the lists back into line, while reporting each one as it was
	// restored would recalculate the entity and re-sync every table on the sheet once per list for the one undo. See
	// restoredTables for the rest of the reasoning.
	var restored restoredTables
	for _, restorer := range d.restorers {
		restored.add(restorer.restore())
	}
	restored.report()
}

// syncWithAllSources syncs a document's model with the library sources its items came from and records the change as
// a single undoable edit spanning the given lists, which are the ones the sync can alter. One edit rather than one per
// list is what lets an undo put every list back and update the document just once (see tablesUndoData.Apply). The
// owner is rebuilt afterwards, which is what reports the change (see rebuildAsModified). Building the "before" data
// as part of the edit, right before the sync, is safe here because syncing alters items in place and moves nothing;
// see organizeTraits for an edit that has to take its snapshot with more care.
func syncWithAllSources(owner Rebuildable, model librarySyncer, lists ...syncableList) {
	var undo *unison.UndoEdit[*tablesUndoData]
	mgr := unison.UndoManagerFor(owner)
	if mgr != nil {
		undo = &unison.UndoEdit[*tablesUndoData]{
			ID:         unison.NextUndoID(),
			EditName:   syncWithSourceAction.Title,
			UndoFunc:   func(e *unison.UndoEdit[*tablesUndoData]) { e.BeforeData.Apply() },
			RedoFunc:   func(e *unison.UndoEdit[*tablesUndoData]) { e.AfterData.Apply() },
			AbsorbFunc: func(_ *unison.UndoEdit[*tablesUndoData], _ unison.Undoable) bool { return false },
			BeforeData: newTablesUndoData(lists),
		}
	}
	model.SyncWithLibrarySources()
	for _, list := range lists {
		list.syncToModel()
	}
	if undo != nil {
		undo.AfterData = newTablesUndoData(lists)
		mgr.Add(undo)
	}
	rebuildAsModified(owner, true)
}

// restoredTables accumulates the results of restoring one or more tables, so that the undo they belong to reports the
// change exactly once no matter how many tables it had to put back.
type restoredTables struct {
	table unison.Paneler
	owner Rebuildable
}

// add records what restoring one table produced, i.e. the results of a TableUndoEditData.restore() call.
func (r *restoredTables) add(table unison.Paneler, owner Rebuildable) {
	if xreflect.IsNil(r.table) && !xreflect.IsNil(table) {
		r.table = table
	}
	if xreflect.IsNil(r.owner) && !xreflect.IsNil(owner) {
		r.owner = owner
	}
}

// report tells the rest of the app about the restored data.
func (r *restoredTables) report() {
	// An owner is rebuilt rather than merely told that one of its tables changed, because far more of what it shows
	// than the rows themselves is derived from the data that was just put back, and only rebuilding recomputes any of
	// it (see rebuildAsModified). Marking as modified is left for the tables that have no owner recorded, i.e. those
	// in an editor or a library list, which nothing ever replaces or reshapes.
	//
	// Whichever of the two applies happens once for the whole undo rather than once per table, which is why restoring
	// doesn't report the change on its own: on a sheet that may hold hundreds of rows all of that work is the entire
	// cost of the edit, and an undo spanning six tables would otherwise pay it six times over.
	if !xreflect.IsNil(r.owner) {
		rebuildAsModified(r.owner, true)
		return
	}
	if !xreflect.IsNil(r.table) {
		MarkModified(r.table)
	}
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
