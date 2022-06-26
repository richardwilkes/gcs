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
	undoTitle string
	last      string
	get       func() string
	set       func(string)
	useGet    bool
	inUndo    bool
}

// NewMultiLineStringField creates a new field for editing a string.
func NewMultiLineStringField(undoTitle string, get func() string, set func(string)) *StringField {
	f := newStringField(unison.NewMultiLineField(), undoTitle, get, set)
	f.SetWrap(true)
	return f
}

// NewStringField creates a new field for editing a string.
func NewStringField(undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewField(), undoTitle, get, set)
}

func newStringField(field *unison.Field, undoTitle string, get func() string, set func(string)) *StringField {
	f := &StringField{
		Field:     field,
		undoID:    unison.NextUndoID(),
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
			mgr.Add(&unison.UndoEdit[string]{
				ID:       f.undoID,
				EditName: f.undoTitle,
				UndoFunc: func(e *unison.UndoEdit[string]) { f.setWithoutUndo(e.BeforeData, true) },
				RedoFunc: func(e *unison.UndoEdit[string]) { f.setWithoutUndo(e.AfterData, true) },
				AbsorbFunc: func(e *unison.UndoEdit[string], other unison.Undoable) bool {
					if e2, ok := other.(*unison.UndoEdit[string]); ok && e2.ID == f.undoID {
						e.AfterData = e2.AfterData
						return true
					}
					return false
				},
				BeforeData: f.getData(),
				AfterData:  text,
			})
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
