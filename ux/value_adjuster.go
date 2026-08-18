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
	if s.entity != nil {
		s.entity.Recalculate()
	}
	if s.owner != nil {
		MarkModified(s.owner)
		if s.rebuild {
			s.owner.Rebuild(true)
		}
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
// target via extract. When recalculate is true, the owning entity is recalculated after the change (and on undo/redo);
// when rebuild is true, the owner is rebuilt as well. The same extract should be used by the corresponding
// canAdjustSelection call so that the enable check and the action never diverge.
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
// undo/redo); when rebuild is true, the owner is rebuilt as well.
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
