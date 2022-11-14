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

package model

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

var (
	// AllSpellComparisonType holds all possible values.
	AllSpellComparisonType = []SpellComparisonType{
		NameSpellComparisonType,
		TagSpellComparisonType,
		CollegeSpellComparisonType,
		CollegeCountSpellComparisonType,
		AnySpellComparisonType,
	}
	spellComparisonTypeData = []struct {
		key     string
		oldKeys []string
		string  string
	}{
		{
			key:    "name",
			string: i18n.Text("whose name"),
		},
		{
			key:     "tag",
			oldKeys: []string{"category"},
			string:  i18n.Text("with a tag which"),
		},
		{
			key:    "college",
			string: i18n.Text("whose college name"),
		},
		{
			key:    "college_count",
			string: i18n.Text("from different colleges"),
		},
		{
			key:    "any",
			string: i18n.Text("of any kind"),
		},
	}
)

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
	return spellComparisonTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum SpellComparisonType) String() string {
	return spellComparisonTypeData[enum.EnsureValid()].string
}

// ExtractSpellComparisonType extracts the value from a string.
func ExtractSpellComparisonType(str string) SpellComparisonType {
	for i, one := range spellComparisonTypeData {
		if strings.EqualFold(one.key, str) || txt.CaselessSliceContains(one.oldKeys, str) {
			return SpellComparisonType(i)
		}
	}
	return 0
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
