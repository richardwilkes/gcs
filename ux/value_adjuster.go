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
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// valueSnapshot captures a target and its value at a point in time, so the value can be restored on undo/redo.
type valueSnapshot[A, V any] struct {
	target A
	value  V
}

// snapshotList is the undo payload for a value adjustment applied across one or more targets.
type snapshotList[A, V any] struct {
	owner   Rebuildable
	entity  *gurps.Entity
	set     func(A, V)
	list    []valueSnapshot[A, V]
	rebuild bool
}

func (s *snapshotList[A, V]) apply() {
	for i := range s.list {
		s.set(s.list[i].target, s.list[i].value)
	}
	s.finish()
}

func (s *snapshotList[A, V]) finish() {
	if s.owner != nil && s.rebuild {
		// Rebuilding is a superset of marking the owner as modified -- it recalculates the entity, re-syncs every
		// table, refreshes the search results and restores the focus and scroll position -- so doing both would repeat
		// the whole update, and on a sheet holding hundreds of rows that update is the entire cost of the edit. The
		// one thing a rebuild doesn't do is bump the owner's modification timestamp, so that is done here instead, the
		// same way an undo that rebuilds does it (see restoredTables.report).
		if s.entity != nil && !ownerRebuildRecalculates(s.owner, s.entity) {
			s.entity.Recalculate()
		}
		if bumper, ok := s.owner.(modificationTimestampBumper); ok {
			// The bump has to happen before the rebuild for the panel showing it to pick up the new value.
			bumper.bumpModificationTimestamp()
		}
		s.owner.Rebuild(true)
		return
	}
	if s.entity != nil && !ownerRecalculates(s.owner, s.entity) {
		s.entity.Recalculate()
	}
	if s.owner != nil {
		MarkModified(s.owner)
	}
}

// sheetOwning returns the sheet that the given panel belongs to and that displays the given entity, or nil if there
// isn't one. The walk up the hierarchy stops at the first ModifiableRoot, since that is the document an edit reports
// itself to; anything above it belongs to something else.
func sheetOwning(owner unison.Paneler, entity *gurps.Entity) *Sheet {
	if xreflect.IsNil(owner) || entity == nil {
		return nil
	}
	for p := owner.AsPanel(); p != nil; p = p.Parent() {
		if _, ok := p.Self.(ModifiableRoot); ok {
			if sheet, ok2 := p.Self.(*Sheet); ok2 && sheet.entity == entity {
				return sheet
			}
			return nil
		}
	}
	return nil
}

// ownerRecalculates returns true if marking the owner as modified will recalculate the given entity on its own, so
// that a single edit doesn't pay for the recalculation twice. Only a Sheet does that, and only for its own entity:
// Sheet.MarkModified recalculates before updating anything, since everything it then touches reads the derived state.
// A sheet that is already in the middle of an update pass doesn't count, since MarkModified does nothing at all while
// awaitingUpdate is set. Skipping the recalculation in that case wouldn't defer it, it would drop it, leaving the
// derived state stale until some unrelated later edit. Reading the flag here is safe: it is only ever set and cleared
// within MarkModified, which, like everything else here, runs on the UI thread.
func ownerRecalculates(owner unison.Paneler, entity *gurps.Entity) bool {
	sheet := sheetOwning(owner, entity)
	return sheet != nil && !sheet.awaitingUpdate
}

// ownerRebuildRecalculates returns true if rebuilding the owner will recalculate the given entity on its own. As with
// marking as modified, only a Sheet does, and only for its own entity, since Sheet.Rebuild recalculates before it
// updates anything that reads the derived state. Unlike marking as modified, a rebuild is never suppressed, so there
// is no in-progress update pass to take into account here.
func ownerRebuildRecalculates(owner unison.Paneler, entity *gurps.Entity) bool {
	return sheetOwning(owner, entity) != nil
}

// recalculateEntityFor brings the entity that owns the given node up to date, unless marking the given owner as
// modified will do that on its own, so that a single edit doesn't pay for the recalculation twice. Callers are
// expected to mark the owner as modified afterwards.
func recalculateEntityFor[T gurps.Node[T]](node T, owner unison.Paneler) {
	entity := gurps.EntityFromNode(node)
	if !ownerRecalculates(owner, entity) {
		entity.Recalculate()
	}
}

// canAdjustSelection returns true if any selected row yields an adjustable target via extract.
func canAdjustSelection[T gurps.Node[T], A any](table *unison.Table[*Node[T]], extract func(T) (A, bool)) bool {
	for _, row := range table.SelectedRows(false) {
		if _, ok := extract(row.Data()); ok {
			return true
		}
	}
	return false
}

// adjustSelection snapshots, mutates, and registers an undoable edit for each selected row that yields an adjustable
// target via extract. When recalculate is true, the owning entity is recalculated after the change (and on undo/redo).
// Pass rebuild for a change that alters more of what the owner shows than the rows being adjusted -- which lists are
// on the page, which columns they hold -- so that the owner is rebuilt rather than just marked as modified. The same
// extract should be used by the corresponding canAdjustSelection call so that the enable check and the action never
// diverge.
func adjustSelection[T gurps.Node[T], A, V any](undoTitle string, owner Rebuildable, table *unison.Table[*Node[T]],
	extract func(T) (A, bool), get func(A) V, set func(A, V), mutate func(A), recalculate, rebuild bool,
) {
	rows := table.SelectedRows(false)
	targets := make([]A, 0, len(rows))
	for _, row := range rows {
		if target, ok := extract(row.Data()); ok {
			targets = append(targets, target)
		}
	}
	if len(targets) == 0 {
		return
	}
	var entity *gurps.Entity
	if recalculate {
		entity = gurps.EntityFromNode(rows[0].Data())
	}
	adjustTargets(undoTitle, owner, table, entity, targets, get, set, mutate, rebuild)
}

// adjustTargets snapshots, mutates, and registers an undoable edit for each of the given targets. undoSource is the
// panel used to locate the undo manager. When entity is non-nil, it is recalculated after the change (and on
// undo/redo); when rebuild is true, the owner is rebuilt instead of merely being marked as modified.
func adjustTargets[A, V any](undoTitle string, owner Rebuildable, undoSource unison.Paneler, entity *gurps.Entity,
	targets []A, get func(A) V, set func(A, V), mutate func(A), rebuild bool,
) {
	if len(targets) == 0 {
		return
	}
	before := &snapshotList[A, V]{owner: owner, entity: entity, set: set, rebuild: rebuild}
	after := &snapshotList[A, V]{owner: owner, entity: entity, set: set, rebuild: rebuild}
	for _, target := range targets {
		before.list = append(before.list, valueSnapshot[A, V]{target: target, value: get(target)})
		mutate(target)
		after.list = append(after.list, valueSnapshot[A, V]{target: target, value: get(target)})
	}
	if mgr := unison.UndoManagerFor(undoSource); mgr != nil {
		mgr.Add(&unison.UndoEdit[*snapshotList[A, V]]{
			ID:         unison.NextUndoID(),
			EditName:   undoTitle,
			UndoFunc:   func(edit *unison.UndoEdit[*snapshotList[A, V]]) { edit.BeforeData.apply() },
			RedoFunc:   func(edit *unison.UndoEdit[*snapshotList[A, V]]) { edit.AfterData.apply() },
			BeforeData: before,
			AfterData:  after,
		})
	}
	before.finish()
}
