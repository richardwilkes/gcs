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

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NameSkillSelectionType SkillSelectionType = iota
	ThisWeaponSkillSelectionType
	WeaponsWithNameSkillSelectionType
	LastSkillSelectionType = WeaponsWithNameSkillSelectionType
)

var (
	// AllSkillSelectionType holds all possible values.
	AllSkillSelectionType = []SkillSelectionType{
		NameSkillSelectionType,
		ThisWeaponSkillSelectionType,
		WeaponsWithNameSkillSelectionType,
	}
	skillSelectionTypeData = []struct {
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

// SkillSelectionType holds the type of a selection.
type SkillSelectionType byte

// EnsureValid ensures this is of a known value.
func (enum SkillSelectionType) EnsureValid() SkillSelectionType {
	if enum <= LastSkillSelectionType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum SkillSelectionType) Key() string {
	return skillSelectionTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum SkillSelectionType) String() string {
	return skillSelectionTypeData[enum.EnsureValid()].string
}

// ExtractSkillSelectionType extracts the value from a string.
func ExtractSkillSelectionType(str string) SkillSelectionType {
	for i, one := range skillSelectionTypeData {
		if strings.EqualFold(one.key, str) {
			return SkillSelectionType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum SkillSelectionType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *SkillSelectionType) UnmarshalText(text []byte) error {
	*enum = ExtractSkillSelectionType(string(text))
	return nil
}
