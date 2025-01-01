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
	"strconv"

	"github.com/richardwilkes/unison"
)

// Selectable panels can have their selection queried and set.
type Selectable interface {
	Selection() (start, end int)
	SetSelection(start, end int)
}

// FocusRef holds a focus reference.
type FocusRef struct {
	Key        string
	SelStart   int
	SelEnd     int
	Selectable bool
}

// TargetMgr provides management of target panels.
type TargetMgr struct {
	root   *unison.Panel
	lastID int
}

// NewTargetMgr creates a new TargetMgr with the given root.
func NewTargetMgr(root unison.Paneler) *TargetMgr {
	return &TargetMgr{root: root.AsPanel()}
}

// NextPrefix returns the next unique prefix to use.
func (t *TargetMgr) NextPrefix() string {
	t.lastID++
	return strconv.Itoa(t.lastID) + ":"
}

// Find searches the tree of panels starting at the root, looking for a specific refKey.
func (t *TargetMgr) Find(refKey string) *unison.Panel {
	return t.root.FindRefKey(refKey)
}

// CurrentFocusRef returns the current FocusRef, if any.
func (t *TargetMgr) CurrentFocusRef() *FocusRef {
	wnd := t.root.Window()
	if wnd == nil {
		return nil
	}
	focus := wnd.Focus()
	if focus == nil {
		return nil
	}
	if unison.AncestorIsOrSelf(focus, t.root) {
		ref := &FocusRef{Key: focus.RefKey}
		if s, ok := focus.Self.(Selectable); ok {
			ref.Selectable = true
			ref.SelStart, ref.SelEnd = s.Selection()
		}
		return ref
	}
	return nil
}

// ReacquireFocus attempts to restore the focus previously obtained by a call to CurrentFocusRef().
func (t *TargetMgr) ReacquireFocus(ref *FocusRef, toolbar, content unison.Paneler) {
	if ref != nil {
		if focus := t.Find(ref.Key); focus != nil {
			focus.RequestFocus()
			if ref.Selectable {
				if s, ok := focus.Self.(Selectable); ok {
					s.SetSelection(ref.SelStart, ref.SelEnd)
				}
			}
		} else {
			FocusFirstContent(toolbar, content)
		}
	}
}
