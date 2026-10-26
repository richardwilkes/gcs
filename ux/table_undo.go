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
	"github.com/richardwilkes/toolbox/v2/tid"
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

// beginTableUndo starts recording an undoable edit of a table's contents, capturing the current data as the "before"
// state; commitTableUndo captures the "after" state and records the edit. Returns nil when there is no undo manager or
// the "before" state could not be captured; commitTableUndo accepts nil, so an edit can be bracketed by the pair
// unconditionally. The optional beforeUndo and beforeRedo hooks run ahead of putting the "before" and "after" data
// back, for an edit that has to restore more than the table's contents (see MoveSelection).
func beginTableUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], title string, beforeUndo, beforeRedo func()) *unison.UndoEdit[*TableUndoEditData[T]] {
	notePriorModifierPicks(table)
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

// commitTableUndo finishes an edit begun with beginTableUndo: the table's current data becomes the edit's "after"
// state and the edit is handed to the table's undo manager. Both come from the table currently showing the data (see
// liveTable), since the edit may have replaced the table the "before" state was captured from, and an orphaned table
// has no manager above it any more. A nil edit is ignored.
func commitTableUndo[T gurps.Node[T]](table *unison.Table[*Node[T]], undo *unison.UndoEdit[*TableUndoEditData[T]]) {
	settleModifierChoices(table)
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

// restore puts the preserved data back into the table currently showing it and returns that table along with its
// owner, if any, leaving the reporting to restoredTables so an undo spanning several tables reports just once. It has
// to be the table on screen rather than the one the data was collected from: the model would come back either way,
// since an orphaned table's provider still points at it, but the selection would not, because the rebuild that follows
// re-applies the on-screen table's selection and would discard one restored into an orphan. Nothing is returned if
// there was nothing to restore or the restore failed. The table comes back as a plain Paneler, which lets an undo
// spanning tables of different row types hold their data behind one interface (see tableRestorer).
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

// grantRowList is what an edit removing items that carry grants needs of each list a grant can put rows in: undo data
// for its table and the table's reference key, which tells the list the edit is of from the others. Every *PageList[T]
// provides both.
type grantRowList interface {
	undoData() tableRestorer
	refKey() string
}

// grantRowListsProvider is the document whose lists a grant can put rows in, which is a sheet, since only an entity
// reconciles its grants.
type grantRowListsProvider interface {
	grantRowLists() []grantRowList
}

// otherGrantRowLists returns the undo data of the lists of the table's owner, other than the table's own, that a grant
// can put rows in, when any of the removed nodes, or anything within one, carries a grant, and nothing otherwise; see
// grantRowListsOf.
func otherGrantRowLists[T gurps.Node[T]](table *unison.Table[*Node[T]], removed []T) []tableRestorer {
	if !carriesGrants(removed) {
		return nil
	}
	return grantRowListsOf(table.ClientData()[TableOwnerClientKey], table.RefKey)
}

// carriesGrants returns true if any of the nodes, or anything within one, carries a grant.
func carriesGrants[T gurps.Node[T]](nodes []T) bool {
	carries := false
	gurps.Traverse(func(node T) bool {
		carries = gurps.CarriesGrants(node)
		return carries
	}, false, false, nodes...)
	return carries
}

// rowsCarryGrants returns true if any of the rows' nodes, or anything within one, carries a grant.
func rowsCarryGrants[T gurps.Node[T]](rows []*Node[T]) bool {
	var zero T
	nodes := make([]T, 0, len(rows))
	for _, row := range rows {
		if node := row.Data(); node != zero {
			nodes = append(nodes, node)
		}
	}
	return carriesGrants(nodes)
}

// grantRowListsOf returns the undo data of the owner's lists that a grant can put rows in, other than the one with the
// given reference key, or all of them for an empty key, when the owner is a sheet, and nothing otherwise. An edit that
// can take a carrier's grants away, by removing, disabling or moving the carrier or taking the grants out of it, takes
// the rows they put in those lists along with them, so an undo of the edit has to put those lists back as well, or the
// restored grants make fresh rows at the end of the lists in place of the ones the user had placed and switched.
func grantRowListsOf(owner any, except string) []tableRestorer {
	return grantRowListsSkipping(owner, func(refKey string) bool { return except != "" && refKey == except })
}

// grantRowListsSkipping is grantRowListsOf for the lists other than those whose reference keys skip returns true for.
func grantRowListsSkipping(owner any, skip func(refKey string) bool) []tableRestorer {
	provider, ok := owner.(grantRowListsProvider)
	if !ok || xreflect.IsNil(provider) {
		return nil
	}
	var restorers []tableRestorer
	for _, list := range provider.grantRowLists() {
		if xreflect.IsNil(list) || skip(list.refKey()) {
			continue
		}
		if restorer := list.undoData(); restorer != nil {
			restorers = append(restorers, restorer)
		}
	}
	return restorers
}

// restoreGrantRowLists puts the lists back without reporting the change, for an edit that reports it on its own once
// everything it changed is back.
func restoreGrantRowLists(restorers []tableRestorer) {
	for _, one := range restorers {
		one.restore()
	}
}

// beginTableUndoWith is beginTableUndo for an edit that alters other tables of the table's owner as well, whose undo
// data others holds (see otherGrantRowLists), so that undoing the edit puts every table back and updates the document
// once (see tablesUndoData.Apply). Returns nil when there is no undo manager or the table's "before" state could not be
// captured; commitTableUndoWith accepts nil.
func beginTableUndoWith[T gurps.Node[T]](table *unison.Table[*Node[T]], title string, others []tableRestorer) *unison.UndoEdit[*tablesUndoData] {
	notePriorModifierPicks(table)
	if unison.UndoManagerFor(table) == nil {
		return nil
	}
	before := NewTableUndoEditData(table)
	if before == nil {
		return nil
	}
	return &unison.UndoEdit[*tablesUndoData]{
		ID:         unison.NextUndoID(),
		EditName:   title,
		UndoFunc:   func(e *unison.UndoEdit[*tablesUndoData]) { e.BeforeData.Apply() },
		RedoFunc:   func(e *unison.UndoEdit[*tablesUndoData]) { e.AfterData.Apply() },
		AbsorbFunc: func(_ *unison.UndoEdit[*tablesUndoData], _ unison.Undoable) bool { return false },
		BeforeData: &tablesUndoData{restorers: append([]tableRestorer{before}, others...)},
	}
}

// commitTableUndoWith finishes an edit begun with beginTableUndoWith, as commitTableUndo does, with the "after" state
// of the other tables given by others, collected now that the edit is done. A nil edit is ignored.
func commitTableUndoWith[T gurps.Node[T]](table *unison.Table[*Node[T]], undo *unison.UndoEdit[*tablesUndoData], others func() []tableRestorer) {
	settleModifierChoices(table)
	if undo == nil {
		return
	}
	table = liveTable(table)
	if mgr := unison.UndoManagerFor(table); mgr != nil {
		restorers := others()
		if after := NewTableUndoEditData(table); after != nil {
			restorers = append([]tableRestorer{after}, restorers...)
		}
		undo.AfterData = &tablesUndoData{restorers: restorers}
		mgr.Add(undo)
	}
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
// contributes nothing.
func newTablesUndoData(lists []syncableList) *tablesUndoData {
	data := &tablesUndoData{restorers: make([]tableRestorer, 0, len(lists))}
	for _, list := range lists {
		if restorer := list.undoData(); restorer != nil {
			data.restorers = append(data.restorers, restorer)
		}
	}
	return data
}

// newTablesUndoDataForTables collects the undo edit data for each of the given tables, which must all belong to one
// owner, since restoring reports the change once for all of them. A table whose data can't be collected is skipped.
func newTablesUndoDataForTables[T gurps.Node[T]](tables []*unison.Table[*Node[T]]) *tablesUndoData {
	data := &tablesUndoData{restorers: make([]tableRestorer, 0, len(tables))}
	for _, table := range tables {
		if restorer := NewTableUndoEditData(table); restorer != nil {
			data.restorers = append(data.restorers, restorer)
		}
	}
	return data
}

// Apply the undo edit data to the tables.
func (d *tablesUndoData) Apply() {
	// Every list is put back before any of them is reported, so that the undo updates the document once rather than
	// once per table: reporting each list as it was restored would recalculate the entity and re-sync every table on
	// the sheet once per list. See restoredTables.
	var restored restoredTables
	for _, restorer := range d.restorers {
		restored.add(restorer.restore())
	}
	restored.report()
}

// syncWithAllSources syncs a document's model with the library sources its items came from and records the change as a
// single undoable edit spanning the given lists, which are the ones the sync can alter. One edit rather than one per
// list is what lets an undo put every list back and update the document just once (see tablesUndoData.Apply). The
// owner is rebuilt afterwards, which reports the change. Building the "before" data right before the sync is safe here
// because syncing alters items in place and moves nothing; see organizeTraits for an edit that has to take its
// snapshot with more care.
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

// add records the result of one restore call, keeping the first non-nil table and the first non-nil owner.
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
	// An owner is rebuilt rather than merely told that one of its tables changed, because much of what it shows is
	// derived from the data that was just put back and only rebuilding recomputes it. Marking as modified is left for
	// the tables that have no owner recorded, i.e. those in an editor or a library list, which nothing ever replaces or
	// reshapes. Either way it happens once for the whole undo rather than once per table, which is why restoring
	// doesn't report the change on its own.
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
// orphaned: applying data to it would update the model but leave the table on screen untouched, and a rebuild asked
// for through it would be skipped, since an orphan has no Rebuildable above it. The replacement is found through the
// owner recorded on the table, under the reference key the two share. A table with no owner recorded, or whose owner
// holds nothing under that key, is returned as-is, as is a nil table.
func liveTable[T gurps.Node[T]](table *unison.Table[*Node[T]]) *unison.Table[*Node[T]] {
	if table == nil || table.RefKey == "" {
		return table
	}
	rebuildable, found := table.ClientData()[TableOwnerClientKey].(Rebuildable)
	if !found || xreflect.IsNil(rebuildable) {
		return table
	}
	current := rebuildable.AsPanel().FindRefKey(table.RefKey)
	if current == nil {
		return table
	}
	if live, ok := current.Self.(*unison.Table[*Node[T]]); ok {
		return live
	}
	return table
}

// TableDragUndoEditData holds the undo edit data for a table drag: the table the rows landed in, the one they came from
// for a move between two tables, and the other lists the grants of the rows moved put rows in, when the move can take
// those grants away (see grantRowListsOf).
type TableDragUndoEditData[T gurps.Node[T]] struct {
	From   *TableUndoEditData[T]
	To     *TableUndoEditData[T]
	Others []tableRestorer
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
	// Both lists have to be back in place before either of them is reported: a row dragged from one list to the other
	// is in neither of them on undo, or in both of them on redo, while only the first has been restored, so reporting
	// there would re-sync the whole sheet against a state the user never had. One report covers both, because the two
	// tables always belong to the same owner: a drag only counts as a move -- the only case that leaves anything to put
	// back in the list the rows came from, which is why From is nil for the rest -- when both tables are in the same
	// dockable, and only equipment can move between two different tables at all (see the providers'
	// DropShouldMoveData). A nil From restores nothing. The lists the moved rows' grants put rows in come back along
	// with them, for the same reason.
	var restored restoredTables
	restored.add(t.To.restore())
	restored.add(t.From.restore())
	for _, other := range t.Others {
		restored.add(other.restore())
	}
	restored.report()
}

// priorModifierPicksKey holds, in a modifier table's client data, its choices' picks from when the edit began.
const priorModifierPicksKey = "prior-modifier-picks"

// notePriorModifierPicks records, as an edit of the table begins, the choice each enabled modifier is an option of.
func notePriorModifierPicks[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	provider, ok := modifierTableProvider(table)
	if !ok {
		return
	}
	picks := make(map[tid.TID]tid.TID)
	gurps.Traverse(func(node T) bool {
		if choice, inChoice := gurps.ModifierChoiceFor(node); inChoice && node.Enabled() {
			picks[node.ID()] = choice.ID()
		}
		return false
	}, false, false, provider.RootData()...)
	table.ClientData()[priorModifierPicksKey] = picks
}

// settleModifierChoices leaves each modifier choice in the table with no more than one enabled option once an edit is
// done. The choice's pick from before the edit stays on if it is still enabled; otherwise the first enabled one does.
func settleModifierChoices[T gurps.Node[T]](table *unison.Table[*Node[T]]) {
	live := liveTable(table)
	provider, ok := modifierTableProvider(live)
	if !ok {
		return
	}
	prior, found := table.ClientData()[priorModifierPicksKey].(map[tid.TID]tid.TID)
	delete(table.ClientData(), priorModifierPicksKey)
	if gurps.SettleModifierChoices(func(one T) bool {
		choice, _ := gurps.ModifierChoiceFor(one)
		return !found || prior[one.ID()] != choice.ID()
	}, provider.RootData()...) {
		live.MarkForRedraw()
	}
}

// modifierTableProvider returns the provider of the table if it is a table of modifiers.
func modifierTableProvider[T gurps.Node[T]](table *unison.Table[*Node[T]]) (TableProvider[T], bool) {
	var zero T
	if _, ok := any(zero).(gurps.ModifierChoiceProvider); !ok || table == nil {
		return nil, false
	}
	provider, ok := any(table.Model).(TableProvider[T])
	return provider, ok
}
