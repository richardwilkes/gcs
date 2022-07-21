/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package widget

import (
	"github.com/richardwilkes/unison"
)

// StringField holds the value for a string field.
type StringField struct {
	*unison.Field
	undoID    int64
	targetMgr *TargetMgr
	targetKey string
	undoTitle string
	last      string
	get       func() string
	set       func(string)
	useGet    bool
	inUndo    bool
}

// NewMultiLineStringField creates a new field for editing a string.
func NewMultiLineStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	f := newStringField(unison.NewMultiLineField(), targetMgr, targetKey, undoTitle, get, set)
	f.SetWrap(true)
	return f
}

// NewStringField creates a new field for editing a string.
func NewStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewField(), targetMgr, targetKey, undoTitle, get, set)
}

func newStringField(field *unison.Field, targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	f := &StringField{
		Field:     field,
		undoID:    unison.NextUndoID(),
		targetMgr: targetMgr,
		targetKey: targetKey,
		undoTitle: undoTitle,
		last:      get(),
		get:       get,
		set:       set,
		useGet:    true,
	}
	f.Self = f
	f.LostFocusCallback = f.lostFocus
	f.ModifiedCallback = f.modified
	f.Sync()
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: unison.FillAlignment,
		HGrab:  true,
	})
	if targetMgr != nil && targetKey != "" {
		f.ClientData()[TargetIDKey] = targetKey
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

func (f *StringField) modified() {
	text := f.Text()
	if !f.inUndo && f.undoID != unison.NoUndoID {
		if mgr := unison.UndoManagerFor(f); mgr != nil {
			undo := NewTargetUndo(f.targetMgr, f.targetKey, f.undoTitle, f.undoID, func(target *unison.Panel, data string) {
				self := f
				if target != nil {
					if field, ok := target.Self.(*StringField); ok {
						self = field
					}
				}
				self.setWithoutUndo(data, true)
			}, f.get())
			undo.AfterData = text
			mgr.Add(undo)
		}
	}
	if f.last != text {
		f.last = text
		f.set(text)
		MarkForLayoutWithinDockable(f)
		MarkModified(f)
	}
}

func (f *StringField) setWithoutUndo(text string, focus bool) {
	f.inUndo = true
	f.SetText(text)
	f.inUndo = false
	if focus {
		f.RequestFocus()
		f.SelectAll()
	}
}

// Sync the field to the current value.
func (f *StringField) Sync() {
	if !f.Focused() {
		f.useGet = true
	}
	f.setWithoutUndo(f.getData(), false)
}
