// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
		mgr:        targetMgr,
		key:        targetKey,
		callback:   applyCallback,
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

// recordTargetUndo adds an edit for a change to a widget to the undo manager the widget can find, if any. Undoing or
// redoing the edit hands apply the panel that holds the widget's target key at that moment, when the target manager
// can locate one, and the widget itself otherwise, so that a widget which has been swapped out by a rebuild in the
// meantime still routes the change to its replacement. An undo ID of unison.NoUndoID records nothing.
func recordTargetUndo[W unison.Paneler, D any](w W, targetMgr *TargetMgr, targetKey, title string, undoID int64, before, after D, apply func(self W, data D)) {
	if undoID == unison.NoUndoID {
		return
	}
	mgr := unison.UndoManagerFor(w)
	if mgr == nil {
		return
	}
	undo := NewTargetUndo(targetMgr, targetKey, title, undoID, func(target *unison.Panel, data D) {
		self := w
		if target != nil {
			if t, ok := target.Self.(W); ok {
				self = t
			}
		}
		apply(self, data)
	}, before)
	undo.AfterData = after
	mgr.Add(undo)
}

// setTargetRefKey makes the target key the panel's RefKey, so that the target manager can find the panel, when there
// is both a target manager and a key.
func setTargetRefKey(p unison.Paneler, targetMgr *TargetMgr, targetKey string) {
	if targetMgr != nil && targetKey != "" {
		p.AsPanel().RefKey = targetKey
	}
}
