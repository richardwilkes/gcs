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
	"github.com/richardwilkes/unison"
)

// tradeUndoEditData holds what a sale or purchase changes: the sheet's carried equipment and its shopping record. The
// carried equipment is restored into whichever table the sheet holds for it when the edit is applied rather than one
// captured when the edit was made, since the list may be off the page while the buy panel stands in for it, where
// liveTable can't find it, and a rebuild can replace it.
type tradeUndoEditData struct {
	sheet  *Sheet
	record []*gurps.ShoppingRecordEntry
	table  PreservedTableData[*gurps.Equipment]
}

// newTradeUndoEditData collects the undo edit data for the sheet's carried equipment and shopping record, or nil if the
// carried equipment couldn't be collected.
func newTradeUndoEditData(sheet *Sheet) *tradeUndoEditData {
	data := &tradeUndoEditData{
		sheet:  sheet,
		record: gurps.CloneShoppingRecordList(sheet.Entity().ShoppingRecord),
	}
	if err := data.table.Collect(sheet.CarriedEquipment.Table); err != nil {
		errs.Log(err)
		return nil
	}
	return data
}

// Apply the data. The record is put back before the carried equipment, since reporting the latter rebuilds the sheet,
// which tells the record's viewers about the change.
func (t *tradeUndoEditData) Apply() {
	t.sheet.Entity().SetShoppingRecord(t.record)
	table := t.sheet.CarriedEquipment.Table
	if err := t.table.Restore(table); err != nil {
		errs.Log(err)
	}
	var restored restoredTables
	restored.add(table, t.sheet)
	restored.report()
}

// recordTrade runs commit, which changes the sheet's carried equipment and shopping record through the entity, as a
// single undoable edit named editName, then reports the change by rebuilding the sheet. The undo manager is the sheet's
// own rather than the carried equipment table's, since the table has none while it is off the page (see
// tradeUndoEditData).
func recordTrade(sheet *Sheet, editName string, commit func()) {
	mgr := unison.UndoManagerFor(sheet)
	var before *tradeUndoEditData
	if mgr != nil {
		before = newTradeUndoEditData(sheet)
	}
	commit()
	sheet.CarriedEquipment.Table.SyncToModel()
	if before != nil {
		if after := newTradeUndoEditData(sheet); after != nil {
			mgr.Add(&unison.UndoEdit[*tradeUndoEditData]{
				ID:         unison.NextUndoID(),
				EditName:   editName,
				UndoFunc:   func(e *unison.UndoEdit[*tradeUndoEditData]) { e.BeforeData.Apply() },
				RedoFunc:   func(e *unison.UndoEdit[*tradeUndoEditData]) { e.AfterData.Apply() },
				AbsorbFunc: func(_ *unison.UndoEdit[*tradeUndoEditData], _ unison.Undoable) bool { return false },
				BeforeData: before,
				AfterData:  after,
			})
		}
	}
	rebuildAsModified(sheet, true)
}
