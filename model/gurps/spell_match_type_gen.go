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
	AllCollegesSpellMatchType SpellMatchType = iota
	CollegeNameSpellMatchType
	PowerSourceSpellMatchType
	NameSpellMatchType
	LastSpellMatchType = NameSpellMatchType
)

var (
	// AllSpellMatchType holds all possible values.
	AllSpellMatchType = []SpellMatchType{
		AllCollegesSpellMatchType,
		CollegeNameSpellMatchType,
		PowerSourceSpellMatchType,
		NameSpellMatchType,
	}
	spellMatchTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "all_colleges",
			string: i18n.Text("to all colleges"),
		},
		{
			key:    "college_name",
			string: i18n.Text("to the college whose name"),
		},
		{
			key:    "power_source_name",
			string: i18n.Text("to the power source whose name"),
		},
		{
			key:    "spell_name",
			string: i18n.Text("to the spell whose name"),
		},
	}
)

// SpellMatchType holds the type of a match.
type SpellMatchType byte

// EnsureValid ensures this is of a known value.
func (enum SpellMatchType) EnsureValid() SpellMatchType {
	if enum <= LastSpellMatchType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum SpellMatchType) Key() string {
	return spellMatchTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum SpellMatchType) String() string {
	return spellMatchTypeData[enum.EnsureValid()].string
}

// ExtractSpellMatchType extracts the value from a string.
func ExtractSpellMatchType(str string) SpellMatchType {
	for i, one := range spellMatchTypeData {
		if strings.EqualFold(one.key, str) {
			return SpellMatchType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum SpellMatchType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *SpellMatchType) UnmarshalText(text []byte) error {
	*enum = ExtractSpellMatchType(string(text))
	return nil
}
