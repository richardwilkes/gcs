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

// Popup provides a popup menu that works with undo.
type Popup[T comparable] struct {
	*unison.PopupMenu[T]
	undoID    int64
	undoTitle string
	targetMgr *TargetMgr
	targetKey string
	get       func() T
	set       func(sel T)
	last      T
}

// NewPopup creates a new popup menu.
func NewPopup[T comparable](targetMgr *TargetMgr, targetKey, undoTitle string, get func() T, set func(T), items ...T) *Popup[T] {
	p := &Popup[T]{
		PopupMenu: unison.NewPopupMenu[T](),
		undoID:    unison.NextUndoID(),
		undoTitle: undoTitle,
		targetMgr: targetMgr,
		targetKey: targetKey,
		get:       get,
		set:       set,
		last:      get(),
	}
	p.Self = p
	for _, item := range items {
		p.AddItem(item)
	}
	p.Sync()
	p.SelectionChangedCallback = func(popup *unison.PopupMenu[T]) {
		if item, ok := popup.Selected(); ok {
			if p.last != item {
				p.last = item
				recordTargetUndo(p, p.targetMgr, p.targetKey, p.undoTitle, p.undoID, p.get(), item,
					func(self *Popup[T], data T) { self.setWithoutUndo(data) })
			}
			p.set(item)
			MarkModified(p)
		}
	}
	setTargetRefKey(p, targetMgr, targetKey)
	return p
}

func (p *Popup[T]) setWithoutUndo(item T) {
	callback := p.SelectionChangedCallback
	p.SelectionChangedCallback = nil
	p.Select(item)
	p.SelectionChangedCallback = callback
	p.last = item
	p.set(item)
	MarkModified(p)
}

// Sync the popup to the current value.
func (p *Popup[T]) Sync() {
	p.Select(p.get())
}
