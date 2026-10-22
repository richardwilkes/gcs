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
	"github.com/richardwilkes/unison/enums/mod"
)

// installFieldSuggestions has the field offer, below it, the values suggest returns that contain its text as it is
// edited, as suggestedValues chooses them; suggest is called on each change. Choosing one replaces the text through
// SetText, as an undoable edit of its own, so the field's ModifiedCallback sees it as it sees any other edit, which is
// also how the rows learn of each change: the callback has to be in place already, since it is chained onto. While
// the rows show, a bare Down moves the highlight, so this comes after any dropdown of the field's (see
// unison.Field.InstallDropdown), which would otherwise open its menu over them. It returns the suggestions.
func installFieldSuggestions(field *unison.Field, suggest func() []string) *unison.FieldSuggestions[string] {
	suggestions := unison.NewFieldSuggestions[string](field)
	suggestions.ChosenCallback = func(_ int, value string) { field.SetText(value) }
	prevModified := field.ModifiedCallback
	field.ModifiedCallback = func(before, after *unison.FieldState) {
		if prevModified != nil {
			prevModified(before, after)
		}
		suggestions.SetRows(suggestedValues(suggest(), after.Text))
	}
	prevKeyDown := field.KeyDownCallback
	field.KeyDownCallback = func(keyCode unison.KeyCode, mods mod.Modifiers, repeat bool) bool {
		if keyCode == unison.KeyDown && mods&mod.NonSticky == 0 && suggestions.Showing() {
			return field.DefaultKeyDown(keyCode, mods, repeat)
		}
		return prevKeyDown != nil && prevKeyDown(keyCode, mods, repeat)
	}
	return suggestions
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
