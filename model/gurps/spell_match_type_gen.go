// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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

// AllSpellMatchType holds all possible values.
var AllSpellMatchType = []SpellMatchType{
	AllCollegesSpellMatchType,
	CollegeNameSpellMatchType,
	PowerSourceSpellMatchType,
	NameSpellMatchType,
}

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
	switch enum {
	case AllCollegesSpellMatchType:
		return "all_colleges"
	case CollegeNameSpellMatchType:
		return "college_name"
	case PowerSourceSpellMatchType:
		return "power_source_name"
	case NameSpellMatchType:
		return "spell_name"
	default:
		return SpellMatchType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum SpellMatchType) String() string {
	switch enum {
	case AllCollegesSpellMatchType:
		return i18n.Text("to all colleges")
	case CollegeNameSpellMatchType:
		return i18n.Text("to the college whose name")
	case PowerSourceSpellMatchType:
		return i18n.Text("to the power source whose name")
	case NameSpellMatchType:
		return i18n.Text("to the spell whose name")
	default:
		return SpellMatchType(0).String()
	}
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

// ExtractSpellMatchType extracts the value from a string.
func ExtractSpellMatchType(str string) SpellMatchType {
	for _, enum := range AllSpellMatchType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
