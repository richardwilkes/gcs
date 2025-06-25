// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// TableProviderClientKey is the key used to store the table provider with the table.
const TableProviderClientKey = "table-provider"

// AltDropSupport holds handlers for supporting an alternate drop type that drops onto a specific row, rather than
// moving or adding rows.
type AltDropSupport struct {
	DragKey string
	Drop    func(rowIndex int, data any)
}

// InstallTableDropSupport installs our standard drop support on a table.
func InstallTableDropSupport[T gurps.NodeTypes](table *unison.Table[*Node[T]], provider TableProvider[T]) {
	table.ClientData()[TableProviderClientKey] = provider
	unison.InstallDropSupport(table, provider.DragKey(), provider.DropShouldMoveData, willDropCallback[T],
		didDropCallback[T])
	table.DragRemovedRowsCallback = func() { MarkModified(table) }
	table.DropOccurredCallback = func() {
		// We need to defer this to give newly added skills a chance to choose their defaults first, before the normal
		// top-to-bottom sweep fills in the defaults.
		unison.InvokeTaskAfter(func() {
			MarkModified(table)
			table.RequestFocus()
		}, 1)
	}
	if altDropSupport := provider.AltDropSupport(); altDropSupport != nil {
		originalDataDragOverCallback := table.DataDragOverCallback
		originalDataDragExitCallback := table.DataDragExitCallback
		originalDataDragDropCallback := table.DataDragDropCallback
		originalDrawOverCallback := table.DrawOverCallback
		altDropRowIndex := -1
		table.DataDragOverCallback = func(where unison.Point, data map[string]any) bool {
			if _, ok := data[altDropSupport.DragKey]; ok {
				altDropRowIndex = table.OverRow(where.Y)
				return altDropRowIndex != -1
			}
			return originalDataDragOverCallback(where, data)
		}
		table.DataDragExitCallback = func() {
			altDropRowIndex = -1
			originalDataDragExitCallback()
		}
		table.DataDragDropCallback = func(where unison.Point, data map[string]any) {
			if altDropRowIndex != -1 {
				if dd, ok := data[altDropSupport.DragKey]; ok {
					undo := willDropCallback(nil, table, false)
					altDropSupport.Drop(altDropRowIndex, dd)
					finishDidDrop(undo, nil, table, false)
				}
				altDropRowIndex = -1
				table.MarkForRedraw()
			} else {
				originalDataDragDropCallback(where, data)
			}
		}
		table.DrawOverCallback = func(gc *unison.Canvas, rect unison.Rect) {
			originalDrawOverCallback(gc, rect)
			if altDropRowIndex > -1 && altDropRowIndex <= table.LastRowIndex() {
				frame := table.RowFrame(altDropRowIndex)
				paint := unison.ThemeWarning.Paint(gc, frame, paintstyle.Fill)
				paint.SetColorFilter(unison.Alpha30Filter())
				gc.DrawRect(frame, paint)
			}
		}
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
		AbsorbFunc: func(_ *unison.UndoEdit[*TableDragUndoEditData[T]], _ unison.Undoable) bool { return false },
		BeforeData: NewTableDragUndoEditData(from, to),
	}
}

func didDropCallback[T gurps.NodeTypes](undo *unison.UndoEdit[*TableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
	if provider, ok := to.ClientData()[TableProviderClientKey]; ok {
		var tableProvider TableProvider[T]
		if tableProvider, ok = provider.(TableProvider[T]); ok {
			tableProvider.ProcessDropData(from, to)
		}
	}
	if toEntityProvider := unison.Ancestor[gurps.DataOwnerProvider](to); !toolbox.IsNil(toEntityProvider) {
		if owner := toEntityProvider.DataOwner(); !toolbox.IsNil(owner) && !toolbox.IsNil(owner.OwningEntity()) {
			if rebuilder := unison.Ancestor[Rebuildable](to); rebuilder != nil {
				rebuilder.Rebuild(true)
			}
		}
	}
	// Only process modifiers and nameables if this is a drop into a character or loot sheet from something
	// besides a character or loot sheet.
	if isForCharacterOrLootSheet(to) && !isForCharacterOrLootSheet(from) {
		ProcessModifiersForSelection(to)
		ProcessNameablesForSelection(to)
	}
	finishDidDrop(undo, from, to, move)
}

func isForCharacterOrLootSheet(panel unison.Paneler) bool {
	if toolbox.IsNil(panel) {
		return false
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet:
		return true
	default:
		return false
	}
}

func finishDidDrop[T gurps.NodeTypes](undo *unison.UndoEdit[*TableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
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
