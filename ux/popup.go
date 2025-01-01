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
				if mgr := unison.UndoManagerFor(p); mgr != nil {
					undo := NewTargetUndo(p.targetMgr, p.targetKey, p.undoTitle, p.undoID, func(target *unison.Panel, data T) {
						self := p
						if target != nil {
							var field *Popup[T]
							if field, ok = target.Self.(*Popup[T]); ok {
								self = field
							}
						}
						self.set(data)
						MarkModified(self)
					}, p.get())
					undo.AfterData, _ = p.Selected()
					mgr.Add(undo)
				}
			}
			p.set(item)
			MarkModified(p)
		}
	}
	if targetMgr != nil && targetKey != "" {
		p.RefKey = targetKey
	}
	return p
}

// Sync the popup to the current value.
func (p *Popup[T]) Sync() {
	p.Select(p.get())
}
