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
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

const tableProviderClientKey = "table-provider"

// InstallTableDropSupport installs our standard drop support on a table.
func InstallTableDropSupport[T gurps.NodeTypes](table *unison.Table[*Node[T]], provider TableProvider[T]) {
	table.ClientData()[tableProviderClientKey] = provider
	unison.InstallDropSupport[*Node[T], *TableDragUndoEditData[T]](table, provider.DragKey(),
		provider.DropShouldMoveData, willDropCallback[T], didDropCallback[T])
	table.DragRemovedRowsCallback = func() { widget.MarkModified(table) }
	table.DropOccurredCallback = func() {
		widget.MarkModified(table)
		table.RequestFocus()
	}
}

func willDropCallback[T gurps.NodeTypes](from, to *unison.Table[*Node[T]], move bool) *unison.UndoEdit[*TableDragUndoEditData[T]] {
	mgr := unison.UndoManagerFor(to)
	if mgr == nil {
		return nil
	}
	if from != nil && (!move || from == to) {
		from = nil
	}
	return &unison.UndoEdit[*TableDragUndoEditData[T]]{
		ID:         unison.NextUndoID(),
		EditName:   i18n.Text("Drag"),
		UndoFunc:   func(e *unison.UndoEdit[*TableDragUndoEditData[T]]) { e.BeforeData.Apply() },
		RedoFunc:   func(e *unison.UndoEdit[*TableDragUndoEditData[T]]) { e.AfterData.Apply() },
		AbsorbFunc: func(e *unison.UndoEdit[*TableDragUndoEditData[T]], other unison.Undoable) bool { return false },
		BeforeData: NewTableDragUndoEditData(from, to),
	}
}

func didDropCallback[T gurps.NodeTypes](undo *unison.UndoEdit[*TableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
	if provider, ok := to.ClientData()[tableProviderClientKey]; ok {
		var tableProvider TableProvider[T]
		if tableProvider, ok = provider.(TableProvider[T]); ok {
			tableProvider.ProcessDropData(from, to)
		}
	}
	entityProvider := unison.Ancestor[gurps.EntityProvider](to)
	if !toolbox.IsNil(entityProvider) && entityProvider.Entity() != nil {
		if entityProvider != unison.Ancestor[gurps.EntityProvider](from) {
			if rebuilder := unison.Ancestor[widget.Rebuildable](to); rebuilder != nil {
				rebuilder.Rebuild(true)
			}
			ProcessModifiersForSelection(to)
			ProcessNameablesForSelection(to)
		}
	}
	if undo == nil {
		return
	}
	mgr := unison.UndoManagerFor(to)
	if mgr == nil {
		return
	}
	if from != nil && (!move || from == to) {
		from = nil
	}
	undo.AfterData = NewTableDragUndoEditData(from, to)
	mgr.Add(undo)
}
