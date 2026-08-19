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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/drag"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/paintstyle"
)

// TableProviderClientKey is the key used to store the table provider with the table.
const TableProviderClientKey = "table-provider"

// AltDropSupport holds handlers for supporting an alternate drop type that drops onto a specific row, rather than
// moving or adding rows.
type AltDropSupport struct {
	DragKey *uti.DataType
	Drop    func(rowIndex int, data any)
}

// InstallTableDropSupport installs our standard drop support on a table.
func InstallTableDropSupport[T gurps.Node[T]](table *unison.Table[*Node[T]], provider TableProvider[T]) {
	table.ClientData()[TableProviderClientKey] = provider
	unison.InstallDropSupport(table, provider.DragKey(), provider.DropShouldMoveData, willDropCallback[T],
		didDropCallback[T])
	// No DragRemovedRowsCallback is installed. It would only ever fire for a move between two different tables, and
	// the providers only allow that within a single dockable (see their DropShouldMoveData), so the report made for
	// the table the rows landed in covers the one they left as well; a separate report for the source would just
	// update the same owner twice.
	table.DropOccurredCallback = func() {
		// unison notifies the table before it calls the did-drop callback, so this is deferred until the drop handlers
		// have had their turn, and it has to look up the table still on screen rather than assume it, since those
		// handlers may have caused the owner to replace the table. The handlers report the drop themselves when they
		// rebuild the owner (see dropRebuilder), in which case marking the table as modified on top of that would just
		// repeat the whole update; elsewhere the drop is reported here.
		unison.InvokeTaskAfter(func() {
			current := liveTable(table)
			if dropRebuilder(current) == nil {
				MarkModified(current)
			}
			current.RequestFocusWithoutScroll()
		}, 1)
	}
	if altDropSupport := provider.AltDropSupport(); altDropSupport != nil {
		altDataType := altDropSupport.DragKey
		originalDragCallbacks := table.Callbacks
		originalDrawOverCallback := table.DrawOverCallback
		altDropRowIndex := -1
		table.CanAcceptDropCallback = func(di drag.Info) bool {
			if table.Enabled() && !table.IsFiltered() && di.HasDataType(altDataType.UTI) {
				return true
			}
			return originalDragCallbacks.CanAcceptDropCallback(di)
		}
		dragEnterOrUpdate := func(di drag.Info, where geom.Point) (drag.Op, bool) {
			if di.HasDataType(altDataType.UTI) {
				op := drag.None
				if altDropRowIndex = table.OverRow(where.Y); altDropRowIndex != -1 {
					op = drag.Copy
				}
				// Redraw and flush whether over a row or not, so the highlight appears while a row is targeted and is
				// erased as soon as the drag moves off of the rows.
				flushDragFeedback(table.AsPanel())
				return op, true
			}
			return drag.None, false
		}
		table.DragEnteredCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) drag.Op {
			if op, ok := dragEnterOrUpdate(di, where); ok {
				return op
			}
			return originalDragCallbacks.DragEnteredCallback(di, where, mods)
		}
		table.DragUpdatedCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) drag.Op {
			if op, ok := dragEnterOrUpdate(di, where); ok {
				return op
			}
			return originalDragCallbacks.DragUpdatedCallback(di, where, mods)
		}
		table.DragExitedCallback = func() {
			if altDropRowIndex != -1 {
				altDropRowIndex = -1
				flushDragFeedback(table.AsPanel())
			}
			originalDragCallbacks.DragExitedCallback()
		}
		table.DropCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) bool {
			if di.HasDataType(altDataType.UTI) {
				handled := false
				if altDropRowIndex != -1 {
					undo := willDropCallback(nil, table, false)
					altDropSupport.Drop(altDropRowIndex, draggedTableData)
					// Notify the table the same way unison does for a normal drop. The providers' drop handlers only
					// rebuild when the data owner has an owning entity, so without this the dockable holding a
					// template or a traits/equipment list would go on showing itself as unmodified.
					unison.SafeCall(table.DropOccurredCallback)
					finishDidDrop(undo, nil, table, false)
					altDropRowIndex = -1
					flushDragFeedback(table.AsPanel())
					handled = true
				}
				return handled
			}
			return originalDragCallbacks.DropCallback(di, where, mods)
		}
		table.DrawOverCallback = func(gc *unison.Canvas, rect geom.Rect) {
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

func willDropCallback[T gurps.Node[T]](from, to *unison.Table[*Node[T]], move bool) *unison.UndoEdit[*TableDragUndoEditData[T]] {
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

func didDropCallback[T gurps.Node[T]](undo *unison.UndoEdit[*TableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
	if provider, ok := to.ClientData()[TableProviderClientKey]; ok {
		var tableProvider TableProvider[T]
		if tableProvider, ok = provider.(TableProvider[T]); ok {
			tableProvider.ProcessDropData(from, to)
		}
	}
	if rebuilder := dropRebuilder(to); rebuilder != nil {
		// This is also what reports the drop (see DropOccurredCallback in InstallTableDropSupport), which is why the
		// owner is rebuilt as modified rather than just rebuilt.
		rebuildAsModified(rebuilder, true)
		// The rebuild covers the whole owner, so it may have replaced either table: adding rows to one list and
		// removing them from another can change which columns each needs, and a list can only change its columns by
		// building a new table. Everything from here on -- the checks for where the rows came from and went, the
		// merging, the undo edit, even finding the undo manager -- has to work with the tables that took their place
		// rather than the orphaned ones the drag started and finished on. Refreshing both keeps them comparable: two
		// tables that were the same resolve to the same replacement, and two that were different have different
		// reference keys, so they stay different.
		from = liveTable(from)
		to = liveTable(to)
	}
	if isForCharacterOrLootSheet(to) || isForTemplate(to) {
		// Only process modifiers and nameables when the drop comes from something besides a character, loot sheet
		// or template; rows already on one of those have had these resolved.
		if !(isForCharacterOrLootSheet(from) || isForTemplate(from)) {
			ProcessModifiersForSelection(to)
			// Answering the modifier prompt rebuilds the owner all over again, and that rebuild can replace the tables
			// just as the one above did: only the modifiers that are enabled count toward a row having switchable
			// features, so turning one on or off can add or take away the switch column, and a list can only change its
			// columns by building a new table. An orphaned table has no Rebuildable above it, so a rebuild asked for
			// through it never happens, and the rows it reports as selected are its own rather than the ones the user
			// is now looking at -- both of which the steps below depend upon. Applying nameable substitutions rebuilds
			// as well, so refresh again afterwards. Refreshing both tables each time keeps them comparable, for the
			// same reason the rebuild above does.
			from = liveTable(from)
			to = liveTable(to)
			ProcessNameablesForSelection(to)
			from = liveTable(from)
			to = liveTable(to)
		}
		// Merge points into identical existing rows whenever rows are actually being added (from a different table),
		// including a drag from another sheet. A drag within the same table is only a reorder, so it is left alone.
		if from != to {
			MergeAddedRows(to)
		}
	}
	finishDidDrop(undo, from, to, move)
}

// dropRebuilder returns the owner that a drop into the given table is reported to by rebuilding it, or nil if the drop
// is instead reported by marking the table as modified. A drop into a table showing an entity's data is reported with
// a rebuild, because the rows that arrive can change more of what the owner shows than the list they landed in: the
// melee weapons, ranged weapons, reactions and conditional modifiers lists are only on the page while there is
// something to put in them, and the switch column comes and goes with the presence of switchable features, neither of
// which anything short of a rebuild re-creates. Everywhere else -- a template, a loot sheet, a library list, or a table
// in an editor whose item has no entity -- marking the table as modified covers everything a drop can change. The
// normal and the alternate drop handlers both decide whether to rebuild with this, and the table's drop notification
// uses it to tell whether they have already reported the drop, so the three always agree. The entity is taken from the
// table's own provider, which is what the alternate drop handlers clone the dropped rows for. Nil, typed or otherwise,
// is accepted, since the alternate drop path may have no source table at all.
func dropRebuilder(table unison.Paneler) Rebuildable {
	if xreflect.IsNil(table) {
		return nil
	}
	provider, ok := table.AsPanel().ClientData()[TableProviderClientKey].(gurps.DataOwnerProvider)
	if !ok || xreflect.IsNil(provider) {
		return nil
	}
	owner := provider.DataOwner()
	if xreflect.IsNil(owner) || owner.OwningEntity() == nil {
		return nil
	}
	return unison.Ancestor[Rebuildable](table)
}

func isForCharacterOrLootSheet(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return false
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet:
		return true
	default:
		return false
	}
}

func isForTemplate(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return false
	}

	_, ok := unison.AncestorOrSelf[unison.Dockable](panel).(*Template)
	return ok
}

func finishDidDrop[T gurps.Node[T]](undo *unison.UndoEdit[*TableDragUndoEditData[T]], from, to *unison.Table[*Node[T]], move bool) {
	if undo == nil {
		return
	}
	// A drop handler may have rebuilt the owner, replacing either table -- the alternate drop path rebuilds from within
	// the provider's handler, where there is no chance to refresh them beforehand -- and an orphaned table can't even
	// find the undo manager, so the edit would be silently discarded.
	from = liveTable(from)
	to = liveTable(to)
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
