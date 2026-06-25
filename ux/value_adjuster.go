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
	MarkModified(s.owner)
	if s.rebuild {
		s.owner.Rebuild(true)
	}
}

// canAdjustSelection returns true if any selected row yields an adjustable target via extract.
func canAdjustSelection[T gurps.NodeTypes, A any](table *unison.Table[*Node[T]], extract func(T) (A, bool)) bool {
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
func adjustSelection[T gurps.NodeTypes, A, V any](undoTitle string, owner Rebuildable, table *unison.Table[*Node[T]],
	extract func(T) (A, bool), get func(A) V, set func(A, V), mutate func(A), recalculate, rebuild bool,
) {
	rows := table.SelectedRows(false)
	before := &snapshotList[A, V]{owner: owner, set: set, rebuild: rebuild}
	after := &snapshotList[A, V]{owner: owner, set: set, rebuild: rebuild}
	for _, row := range rows {
		if target, ok := extract(row.Data()); ok {
			before.list = append(before.list, valueSnapshot[A, V]{target: target, value: get(target)})
			mutate(target)
			after.list = append(after.list, valueSnapshot[A, V]{target: target, value: get(target)})
		}
	}
	if len(before.list) == 0 {
		return
	}
	if recalculate {
		entity := gurps.EntityFromNode(gurps.AsNode(rows[0].Data()))
		before.entity = entity
		after.entity = entity
	}
	if mgr := unison.UndoManagerFor(table); mgr != nil {
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
