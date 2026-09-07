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

// undoableFieldSelf is what a field that embeds an undoableField provides to it: the steps the base dispatches to the
// embedding field, so that whatever the embedding field layers on top of the base's versions of them is not bypassed.
// Undo edits are applied through it too, since the field that has replaced this one after a rebuild is located by its
// Self.
type undoableFieldSelf interface {
	unison.Paneler
	gainedFocus()
	lostFocus()
	setWithoutUndo(state *unison.FieldState, focus bool)
}

// undoableField is the editing core shared by StringField and NumericField: a unison.Field whose text is parsed into a
// value that is handed to a setter, with each change recorded for undo against a target key, and which Sync keeps in
// step with the value it edits.
type undoableField[T comparable] struct {
	*unison.Field
	self      undoableFieldSelf
	targetMgr *TargetMgr
	targetKey string
	undoTitle string
	get       func() T
	set       func(T)
	// parse converts the field's text into a value and format renders a value as the field's text.
	parse  func(string) T
	format func(T) string
	last   T
	useGet bool
	// hasFocus records whether the field holds the keyboard focus, as reported by its own focus callbacks.
	// Panel.Focused() is not used for this because it also requires the window to be active, and a field keeps the
	// focus -- and receives the next keystroke -- across its window being deactivated and reactivated, with no focus
	// callbacks fired in between.
	hasFocus      bool
	marksModified bool
}

// init sets the field up to edit the value the accessors reach, on behalf of self, the field embedding it.
func (f *undoableField[T]) init(self undoableFieldSelf, field *unison.Field, targetMgr *TargetMgr, targetKey, undoTitle string, get func() T, set func(T), parse func(string) T, format func(T) string) {
	f.Field = field
	f.self = self
	f.targetMgr = targetMgr
	f.targetKey = targetKey
	f.undoTitle = undoTitle
	f.get = get
	f.set = set
	f.parse = parse
	f.format = format
	f.last = get()
	f.useGet = true
	f.marksModified = true
	f.Self = self
	unison.UninstallFocusBorders(f, f)
	f.GainedFocusCallback = self.gainedFocus
	f.LostFocusCallback = self.lostFocus
	unison.InstallDefaultFieldBorder(f, f)
	f.ModifiedCallback = f.modified
	setTargetRefKey(f, targetMgr, targetKey)
}

func (f *undoableField[T]) gainedFocus() {
	f.hasFocus = true
	f.DefaultFocusGained()
}

func (f *undoableField[T]) lostFocus() {
	f.hasFocus = false
	f.useGet = true
	f.SetText(f.format(f.parse(f.Text())))
	f.DefaultFocusLost()
}

func (f *undoableField[T]) getData() string {
	if f.useGet {
		f.useGet = false
		return f.format(f.get())
	}
	return f.Text()
}

// CurrentValue returns the current committed value, which may not be the same as the value showing.
func (f *undoableField[T]) CurrentValue() T {
	return f.get()
}

func (f *undoableField[T]) modified(before, after *unison.FieldState) {
	recordTargetUndo(f.self, f.targetMgr, f.targetKey, f.undoTitle, f.CurrentUndoID(), before, after,
		func(self undoableFieldSelf, data *unison.FieldState) { self.setWithoutUndo(data, true) })
	f.adjustForText()
}

func (f *undoableField[T]) adjustForText() {
	if v := f.parse(f.Text()); f.last != v {
		f.last = v
		f.set(v)
		MarkForLayoutWithinDockable(f)
		if f.marksModified {
			MarkModified(f.self)
		}
	}
}

func (f *undoableField[T]) setWithoutUndo(state *unison.FieldState, focus bool) {
	f.ApplyFieldState(state)
	f.adjustForText()
	if focus {
		f.RequestFocus()
	}
	f.Validate()
}

// Sync the field to the current value. While the field has the focus, the text in it is what the user is working on,
// so it is re-parsed rather than replaced.
func (f *undoableField[T]) Sync() {
	if !f.hasFocus {
		f.useGet = true
	}
	state := f.GetFieldState()
	state.Text = f.getData()
	f.self.setWithoutUndo(state, false)
}
