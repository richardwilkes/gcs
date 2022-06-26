// Code generated from "enum.go.tmpl" - DO NOT EDIT.

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

package skill

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	SkillsWithName SelectionType = iota
	ThisWeapon
	WeaponsWithName
	LastSelectionType = WeaponsWithName
)

var (
	// AllSelectionType holds all possible values.
	AllSelectionType = []SelectionType{
		SkillsWithName,
		ThisWeapon,
		WeaponsWithName,
	}
	selectionTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "skills_with_name",
			string: i18n.Text("to skills whose name"),
		},
		{
			key:    "this_weapon",
			string: i18n.Text("to this weapon"),
		},
		{
			key:    "weapons_with_name",
			string: i18n.Text("to weapons whose name"),
		},
	}
)

// SelectionType holds the type of a selection.
type SelectionType byte

// EnsureValid ensures this is of a known value.
func (enum SelectionType) EnsureValid() SelectionType {
	if enum <= LastSelectionType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum SelectionType) Key() string {
	return selectionTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum SelectionType) String() string {
	return selectionTypeData[enum.EnsureValid()].string
}

// ExtractSelectionType extracts the value from a string.
func ExtractSelectionType(str string) SelectionType {
	for i, one := range selectionTypeData {
		if strings.EqualFold(one.key, str) {
			return SelectionType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum SelectionType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *SelectionType) UnmarshalText(text []byte) error {
	*enum = ExtractSelectionType(string(text))
	return nil
}
