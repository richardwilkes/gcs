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
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	NameSpellComparisonType SpellComparisonType = iota
	TagSpellComparisonType
	CollegeSpellComparisonType
	CollegeCountSpellComparisonType
	AnySpellComparisonType
	LastSpellComparisonType = AnySpellComparisonType
)

// AllSpellComparisonType holds all possible values.
var AllSpellComparisonType = []SpellComparisonType{
	NameSpellComparisonType,
	TagSpellComparisonType,
	CollegeSpellComparisonType,
	CollegeCountSpellComparisonType,
	AnySpellComparisonType,
}

// SpellComparisonType holds the type of a comparison.
type SpellComparisonType byte

// EnsureValid ensures this is of a known value.
func (enum SpellComparisonType) EnsureValid() SpellComparisonType {
	if enum <= LastSpellComparisonType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum SpellComparisonType) Key() string {
	switch enum {
	case NameSpellComparisonType:
		return "name"
	case TagSpellComparisonType:
		return "tag"
	case CollegeSpellComparisonType:
		return "college"
	case CollegeCountSpellComparisonType:
		return "college_count"
	case AnySpellComparisonType:
		return "any"
	default:
		return SpellComparisonType(0).Key()
	}
}

func (enum SpellComparisonType) oldKeys() []string {
	switch enum {
	case NameSpellComparisonType:
		return nil
	case TagSpellComparisonType:
		return []string{"category"}
	case CollegeSpellComparisonType:
		return nil
	case CollegeCountSpellComparisonType:
		return nil
	case AnySpellComparisonType:
		return nil
	default:
		return SpellComparisonType(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum SpellComparisonType) String() string {
	switch enum {
	case NameSpellComparisonType:
		return i18n.Text("whose name")
	case TagSpellComparisonType:
		return i18n.Text("with a tag which")
	case CollegeSpellComparisonType:
		return i18n.Text("whose college name")
	case CollegeCountSpellComparisonType:
		return i18n.Text("from different colleges")
	case AnySpellComparisonType:
		return i18n.Text("of any kind")
	default:
		return SpellComparisonType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum SpellComparisonType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *SpellComparisonType) UnmarshalText(text []byte) error {
	*enum = ExtractSpellComparisonType(string(text))
	return nil
}

// ExtractSpellComparisonType extracts the value from a string.
func ExtractSpellComparisonType(str string) SpellComparisonType {
	for _, enum := range AllSpellComparisonType {
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
