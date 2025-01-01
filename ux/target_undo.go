// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import "github.com/richardwilkes/unison"

// TargetUndo provides undo support for fields that may be swapped out during updates by using a TargetMgr to locate the
// real target.
type TargetUndo[T any] struct {
	unison.UndoEdit[T]
	mgr      *TargetMgr
	key      string
	callback func(target *unison.Panel, data T)
}

// NewTargetUndo creates a new undo that supports having a revisable target.
func NewTargetUndo[T any](targetMgr *TargetMgr, targetKey, title string, undoID int64, applyCallback func(target *unison.Panel, data T), beforeData T) *TargetUndo[T] {
	t := &TargetUndo[T]{
		UndoEdit: unison.UndoEdit[T]{
			ID:       undoID,
			EditName: title,
			EditCost: 1,
			AbsorbFunc: func(e *unison.UndoEdit[T], other unison.Undoable) bool {
				if e2, ok := other.(*TargetUndo[T]); ok && e.ID == e2.ID {
					e.AfterData = e2.AfterData
					return true
				}
				return false
			},
			BeforeData: beforeData,
		},
		mgr:      targetMgr,
		key:      targetKey,
		callback: applyCallback,
	}
	t.UndoFunc = func(e *unison.UndoEdit[T]) { t.apply(e.BeforeData) }
	t.RedoFunc = func(e *unison.UndoEdit[T]) { t.apply(e.AfterData) }
	return t
}

func (t *TargetUndo[T]) apply(data T) {
	var target *unison.Panel
	if t.mgr != nil {
		target = t.mgr.Find(t.key)
	}
	t.callback(target, data)
}
