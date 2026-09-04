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
	"slices"

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

// AltDropSupport holds handlers for supporting an alternate drop type that drops onto specific rows, rather than moving
// or adding rows. The row indexes handed to Drop are those of the rows the drop is to be applied to, in the order they
// appear in the table (see altDropTargets). They are raw indexes into the table the drop landed in, and it is up to
// each Drop implementation to resolve them to rows before it changes anything about that table -- in particular before
// rebuilding the table's owner, which replaces the table and leaves the indexes pointing into an orphan.
type AltDropSupport struct {
	DragKey *uti.DataType
	Drop    func(rowIndexes []int, data any)
}

// altDropTargets returns the rows that an alternate drop released over the row at the given index applies to. Releasing
// over one of several selected rows applies the drop to all of them, so that a modifier can be attached to a whole
// batch of traits or equipment in one go; releasing anywhere else -- over a row that isn't part of the selection, or
// with at most one row selected -- applies it to just the row it landed on, as it always has.
//
// Only rows that are explicitly selected count. unison paints the descendants of a selected container as indirectly
// selected rather than putting them into the selection itself, and dropping onto a container attaches the modifiers to
// the container alone, so a container's children are no more a target here than they are when one is dropped onto
// directly. The scan runs over the disclosed rows rather than the selection map, which is also what leaves out any
// selected row currently tucked inside a closed container: a row that can't be highlighted while dragging shouldn't be
// quietly modified by the drop either.
func altDropTargets[T gurps.Node[T]](table *unison.Table[*Node[T]], hovered int) []int {
	if hovered == -1 {
		return nil
	}
	var selected []int
	rowCount := table.LastRowIndex() + 1
	for i := range rowCount {
		if table.IsRowSelected(i) {
			selected = append(selected, i)
		}
	}
	if len(selected) > 1 && slices.Contains(selected, hovered) {
		return selected
	}
	return []int{hovered}
}

// InstallTableDropSupport installs our standard drop support on a table.
func InstallTableDropSupport[T gurps.Node[T]](table *unison.Table[*Node[T]], provider TableProvider[T]) {
	table.ClientData()[TableProviderClientKey] = provider
	table.InstallDropSupport(provider.DragKey(), provider.DropShouldMoveData, willDropCallback[T], didDropCallback[T])
	// The keyboard repositioning commands are the equivalents of a drag within the table, so they belong on exactly
	// the tables that accept one.
	InstallMoveSelectionHandlers(table)
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
		var altDropTargetIndexes []int
		table.CanAcceptDropCallback = func(di drag.Info) bool {
			if table.Enabled() && !table.IsFiltered() && di.HasDataType(altDataType.UTI) {
				return true
			}
			return originalDragCallbacks.CanAcceptDropCallback(di)
		}
		dragEnterOrUpdate := func(di drag.Info, where geom.Point) (drag.Op, bool) {
			if di.HasDataType(altDataType.UTI) {
				op := drag.None
				if altDropTargetIndexes = altDropTargets(table, table.OverRow(where.Y)); len(altDropTargetIndexes) != 0 {
					op = drag.Copy
				}
				// Redraw and flush whether over a row or not, so the highlight appears while rows are targeted and is
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
			if len(altDropTargetIndexes) != 0 {
				altDropTargetIndexes = nil
				flushDragFeedback(table.AsPanel())
			}
			originalDragCallbacks.DragExitedCallback()
		}
		table.DropCallback = func(di drag.Info, where geom.Point, mods mod.Modifiers) bool {
			if di.HasDataType(altDataType.UTI) {
				handled := false
				if len(altDropTargetIndexes) != 0 {
					undo := willDropCallback(nil, table, false)
					altDropSupport.Drop(altDropTargetIndexes, draggedTableData)
					// Notify the table the same way unison does for a normal drop. The providers' drop handlers only
					// rebuild when the data owner has an owning entity, so without this the dockable holding a
					// template or a traits/equipment list would go on showing itself as unmodified.
					unison.SafeCall(table.DropOccurredCallback)
					finishDidDrop(undo, nil, table, false)
					altDropTargetIndexes = nil
					flushDragFeedback(table.AsPanel())
					handled = true
				}
				return handled
			}
			return originalDragCallbacks.DropCallback(di, where, mods)
		}
		table.DrawOverCallback = func(gc *unison.Canvas, rect geom.Rect) {
			originalDrawOverCallback(gc, rect)
			if len(altDropTargetIndexes) == 0 {
				return
			}
			// Every row the drop will be applied to is highlighted, so that the user can see the whole extent of what
			// is about to happen before letting go. This is drawn afresh on every drag-update event, and the targets
			// can be a selection of hundreds of rows with most of them scrolled out of view, so the work is confined
			// to the rows the dirty rect reaches: RowFrame sums the heights of every row above the one asked for, so
			// it is only called for the targets among those rows -- the indexes are in table order (see
			// altDropTargets), so the scan stops at the first one past them -- and a single paint serves them all.
			first := max(table.OverRow(rect.Y), 0)
			last := table.OverRow(rect.Bottom())
			if last == -1 {
				last = table.LastRowIndex()
			}
			paint := unison.ThemeWarning.Paint(gc, rect, paintstyle.Fill)
			paint.SetColorFilter(unison.Alpha30Filter())
			for _, rowIndex := range altDropTargetIndexes {
				if rowIndex < first {
					continue
				}
				if rowIndex > last {
					break
				}
				if frame := table.RowFrame(rowIndex); frame.Intersects(rect) {
					gc.DrawRect(frame, paint)
				}
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
	if shouldProcessModifiersAndNameablesTo(to) {
		if shouldProcessModifiersAndNameablesFrom(from) {
			// Answering the modifier prompt rebuilds the owner all over again, and that rebuild can replace the tables
			// just as the one above did: only the modifiers that are enabled count toward a row having switchable
			// features, so turning one on or off can add or take away the switch column, and a list can only change its
			// columns by building a new table. An orphaned table has no Rebuildable above it, so a rebuild asked for
			// through it never happens, and the rows it reports as selected are its own rather than the ones the user
			// is now looking at -- both of which the steps below depend upon. Applying nameable substitutions rebuilds
			// as well, so refresh again afterwards. Refreshing both tables each time keeps them comparable, for the
			// same reason the rebuild above does.
			ProcessModifiersForSelection(to)
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
	if clearPreconfiguredFlag(to, nil) {
		to = liveTable(to)
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

// shouldProcessModifiersAndNameablesFrom checks if the copy source should process modifiers and nameables
func shouldProcessModifiersAndNameablesFrom(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return false
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet:
		return false
	default:
		return true
	}
}

// shouldProcessModifiersAndNameablesFrom checks if the copy destination should process modifiers and nameables
func shouldProcessModifiersAndNameablesTo(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return false
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet, *Template:
		return true
	default:
		return false
	}
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
