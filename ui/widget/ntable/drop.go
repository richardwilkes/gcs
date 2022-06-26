/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ntable

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

const tableProviderClientKey = "table-provider"

// InstallTableDropSupport installs our standard drop support on a table.
func InstallTableDropSupport[T gurps.NodeConstraint[T]](table *unison.Table[*Node[T]], provider TableProvider[T]) {
	table.ClientData()[tableProviderClientKey] = provider
	unison.InstallDropSupport[*Node[T], *tableDragUndoEditData[T]](table, provider.DragKey(),
		provider.DropShouldMoveData, willDropCallback[T], didDropCallback[T])
	table.DragRemovedRowsCallback = func() { widget.MarkModified(table) }
	table.DropOccurredCallback = func() { widget.MarkModified(table) }
}

func willDropCallback[T gurps.NodeConstraint[T]](from, to *unison.Table[*Node[T]], move bool) *unison.UndoEdit[*tableDragUndoEditData[T]] {
	mgr := unison.UndoManagerFor(from)
	if mgr == nil {
		return nil
	}
	data := newTableDragUndoEditData(from, to, move)
	if data == nil {
		return nil
	}
	return &unison.UndoEdit[*tableDragUndoEditData[T]]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Drag"),
		UndoFunc:   func(e *unison.UndoEdit[*tableDragUndoEditData[T]]) { e.BeforeData.apply() },
		RedoFunc:   func(e *unison.UndoEdit[*tableDragUndoEditData[T]]) { e.AfterData.apply() },
		AbsorbFunc: func(e *unison.UndoEdit[*tableDragUndoEditData[T]], other unison.Undoable) bool { return false },
		BeforeData: data,
	}
}

func didDropCallback[T gurps.NodeConstraint[T]](undo *unison.UndoEdit[*tableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
	if undo == nil {
		return
	}
	mgr := unison.UndoManagerFor(from)
	if mgr == nil {
		return
	}
	undo.AfterData = newTableDragUndoEditData(from, to, move)
	if undo.AfterData != nil {
		mgr.Add(undo)
	}
}
