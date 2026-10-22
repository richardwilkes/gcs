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
	"strings"

	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/mod"
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

// installSuggestions has the field offer, below it, the values suggest returns that contain its text as it is edited,
// as suggestedValues chooses them; suggest is called on each change. Choosing one replaces the text, as an undoable
// edit of its own. While they show, a bare Down moves the highlight, so this is installed after any dropdown of the
// field's (see unison.Field.InstallDropdown), which would otherwise open its menu over them.
func (f *StringField) installSuggestions(suggest func() []string) {
	f.suggestions = unison.NewFieldSuggestions[string](f.Field)
	f.suggestions.ChosenCallback = func(_ int, value string) { f.SetText(value) }
	prevModified := f.ModifiedCallback
	f.ModifiedCallback = func(before, after *unison.FieldState) {
		if prevModified != nil {
			prevModified(before, after)
		}
		f.suggestions.SetRows(suggestedValues(suggest(), after.Text))
	}
	prevKeyDown := f.KeyDownCallback
	f.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyDown && mods&mod.NonSticky == 0 && f.suggestions.Showing() {
			return f.DefaultKeyDown(keyCode, mods, repeat)
		}
		return prevKeyDown != nil && prevKeyDown(keyCode, mods, repeat)
	}
}

// suggestionsShowing returns true while the field's suggestions are showing, which is when they take the keys that
// drive them, Escape among them, ahead of the field.
func (f *StringField) suggestionsShowing() bool {
	return f.suggestions != nil && f.suggestions.Showing()
}

// suggestedValues returns the values that contain text, ignoring case and the space around text, in the order given,
// leaving out the one that is text itself, since it has already been found. Nothing is suggested for an empty text.
func suggestedValues(values []string, text string) []string {
	text = strings.ToLower(strings.TrimSpace(text))
	if text == "" {
		return nil
	}
	var matches []string
	for _, value := range values {
		if lower := strings.ToLower(value); lower != text && strings.Contains(lower, text) {
			matches = append(matches, value)
		}
	}
	return matches
}
