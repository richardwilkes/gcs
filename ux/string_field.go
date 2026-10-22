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
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

// StringField is a field that holds a string.
type StringField struct {
	undoableField[string]
	// suggestions offers completions of the text as it is typed, or is nil for a field that offers none; see
	// installSuggestions.
	suggestions *unison.FieldSuggestions[string]
}

// NewMultiLineStringField creates a new multi-line field for editing a string.
func NewMultiLineStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewMultiLineField(), targetMgr, targetKey, undoTitle, get, set)
}

// NewStringField creates a new field for editing a string.
func NewStringField(targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	return newStringField(unison.NewField(), targetMgr, targetKey, undoTitle, get, set)
}

func newStringField(field *unison.Field, targetMgr *TargetMgr, targetKey, undoTitle string, get func() string, set func(string)) *StringField {
	f := &StringField{}
	f.init(f, field, targetMgr, targetKey, undoTitle, get, set, textAsIs, textAsIs)
	f.Sync()
	f.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		HGrab:  true,
	})
	return f
}

// textAsIs is both the parse and the format function of a string field, whose text is its value.
func textAsIs(text string) string {
	return text
}

// installSuggestions has the field offer the values suggest returns that contain its text as it is edited, as
// installFieldSuggestions does, which says when to call this.
func (f *StringField) installSuggestions(suggest func() []string) {
	f.suggestions = installFieldSuggestions(f.Field, suggest)
}

// suggestionsShowing returns true while the field's suggestions are showing, which is when they take the keys that
// drive them, Escape among them, ahead of the field.
func (f *StringField) suggestionsShowing() bool {
	return f.suggestions != nil && f.suggestions.Showing()
}
