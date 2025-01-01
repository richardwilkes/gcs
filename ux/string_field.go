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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// StringField holds the value for a string field.
type StringField struct {
	*unison.Field
	targetMgr *TargetMgr
	targetKey string
	undoTitle string
	last      string
	get       func() string
	set       func(string)
	useGet    bool
}

// NewMultiLineStringField creates a new field for editing a string.
func NewMultiLineStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewMultiLineField(), targetMgr, targetKey, undoTitle, get, set)
}

// NewStringField creates a new field for editing a string.
func NewStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewField(), targetMgr, targetKey, undoTitle, get, set)
}

func newStringField(field *unison.Field, targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	f := &StringField{
		Field:     field,
		targetMgr: targetMgr,
		targetKey: targetKey,
		undoTitle: undoTitle,
		last:      get(),
		get:       get,
		set:       set,
		useGet:    true,
	}
	f.Self = f
	unison.UninstallFocusBorders(f, f)
	f.LostFocusCallback = f.lostFocus
	unison.InstallDefaultFieldBorder(f, f)
	f.ModifiedCallback = f.modified
	f.Sync()
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	if targetMgr != nil && targetKey != "" {
		f.RefKey = targetKey
	}
	return f
}

func (f *StringField) lostFocus() {
	f.useGet = true
	f.SetText(f.Text())
	f.DefaultFocusLost()
}

func (f *StringField) getData() string {
	if f.useGet {
		f.useGet = false
		return f.get()
	}
	return f.Text()
}

func (f *StringField) modified(before, after *unison.FieldState) {
	if f.CurrentUndoID() != unison.NoUndoID {
		if mgr := unison.UndoManagerFor(f); mgr != nil {
			undo := NewTargetUndo(f.targetMgr, f.targetKey, f.undoTitle, f.CurrentUndoID(),
				func(target *unison.Panel, data *unison.FieldState) {
					self := f
					if target != nil {
						if field, ok := target.Self.(*StringField); ok {
							self = field
						}
					}
					self.setWithoutUndo(data, true)
				}, before)
			undo.AfterData = after
			mgr.Add(undo)
		}
	}
	f.adjustForText()
}

func (f *StringField) adjustForText() {
	text := f.Text()
	if f.last != text {
		f.last = text
		f.set(text)
		MarkForLayoutWithinDockable(f)
		MarkModified(f)
	}
}

func (f *StringField) setWithoutUndo(state *unison.FieldState, focus bool) {
	f.ApplyFieldState(state)
	f.adjustForText()
	if focus {
		f.RequestFocus()
	}
	f.Validate()
}

// Sync the field to the current value.
func (f *StringField) Sync() {
	if !f.Focused() {
		f.useGet = true
	}
	state := f.GetFieldState()
	state.Text = f.getData()
	f.setWithoutUndo(state, false)
}
