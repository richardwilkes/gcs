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
	if t == nil {
		return
	}
	table := liveTable(t.Table)
	if err := t.Data.Apply(table); err != nil {
		errs.Log(err)
		return
	}
	// Some columns are only present when the data calls for them (the switch column, for one), so putting the old data
	// back may require the owner to rebuild its lists to bring such a column into or out of view.
	if provider, ok := table.ClientData()[TableProviderClientKey].(TableProvider[T]); ok &&
		columnsOutOfSync(provider.ColumnIDs(), table.Columns) {
		if owner, ok2 := table.ClientData()[TableOwnerClientKey].(Rebuildable); ok2 && !xreflect.IsNil(owner) {
			owner.Rebuild(true)
		}
	}
}

// liveTable returns the table that is currently showing the data the given table was created for. An owner that has to
// alter its set of columns can only do so by replacing the table entirely, which leaves any table captured earlier
// orphaned: applying data to it would update the model but leave the table the user is looking at untouched.
func liveTable[T gurps.Node[T]](table *unison.Table[*Node[T]]) *unison.Table[*Node[T]] {
	if table.RefKey == "" {
		return table
	}
	owner, ok := table.ClientData()[TableOwnerClientKey].(Rebuildable)
	if !ok || xreflect.IsNil(owner) {
		return table
	}
	if panel := owner.AsPanel().FindRefKey(table.RefKey); panel != nil {
		if current, ok2 := panel.Self.(*unison.Table[*Node[T]]); ok2 {
			return current
		}
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
	t.To.Apply()
	if t.From != nil {
		t.From.Apply()
	}
}
