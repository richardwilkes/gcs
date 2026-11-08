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
	// reveal, when set, is asked by findRevealing to show the collapsed part of the content holding the widget with the
	// reference key, reporting whether it did.
	reveal func(refKey string) bool
}

// NewTargetMgr creates a new TargetMgr with the given root.
func NewTargetMgr(root unison.Paneler) *TargetMgr {
	return &TargetMgr{root: root.AsPanel()}
}

// NextPrefix returns the next unique prefix to use. It starts with "#", so that the keys it begins never match those of
// the sentence rows within the same panels, which start with a row's index, such as "3:name".
func (t *TargetMgr) NextPrefix() string {
	t.lastID++
	return "#" + strconv.Itoa(t.lastID) + ":"
}

// Find searches the tree of panels starting at the root, looking for a specific refKey.
func (t *TargetMgr) Find(refKey string) *unison.Panel {
	return t.root.FindRefKey(refKey)
}

// findRevealing is Find for the widget an undo or redo applies to, which a collapsed part of the content may hide: when
// Find comes up empty, reveal, if set, is asked to show the part holding it, and the search is made again.
func (t *TargetMgr) findRevealing(refKey string) *unison.Panel {
	if found := t.Find(refKey); found != nil || t.reveal == nil || !t.reveal(refKey) {
		return found
	}
	return t.Find(refKey)
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

// ReacquireFocus attempts to restore the focus previously obtained by a call to CurrentFocusRef(). A ref without a key,
// such as one for a heading or a label, leaves the focus alone while the window still has one, and otherwise focuses
// the first focusable part of the content. Looking up an empty key would find the root and focus its first focusable
// child, a toolbar button.
func (t *TargetMgr) ReacquireFocus(ref *FocusRef, toolbar, content unison.Paneler) {
	if ref == nil {
		return
	}
	if ref.Key == "" {
		if wnd := t.root.Window(); wnd != nil && wnd.CurrentFocus() != nil {
			return
		}
		FocusFirstContent(toolbar, content)
		return
	}
	if focus := t.Find(ref.Key); focus != nil {
		focus = tabStopOf(focus)
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

// tabStopOf returns the panel a panel found by its reference key gives the focus to. The views of a table share its
// reference key, and the first found is its primary view, which may be pages away from the row the person is on, so
// the focus goes to the view showing that row or, when there is none, to the view the table gives its tab stop to. The
// row comes first, since the view the focus was left on may no longer show it once the rows have been set out anew.
// Any other panel takes the focus itself.
func tabStopOf(p *unison.Panel) *unison.Panel {
	if holder, ok := p.Self.(unison.TabStopHolder); ok {
		if rowView, hasRowView := p.ClientData()[currentRowViewKey].(func() *unison.Panel); hasRowView {
			if view := rowView(); view != nil {
				return view
			}
		}
		if stop := holder.TabStopPanel(); stop != nil {
			return stop
		}
	}
	return p
}
