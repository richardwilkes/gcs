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

package spell

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	AllColleges MatchType = iota
	CollegeName
	PowerSource
	Spell
	LastMatchType = Spell
)

var (
	// AllMatchType holds all possible values.
	AllMatchType = []MatchType{
		AllColleges,
		CollegeName,
		PowerSource,
		Spell,
	}
	matchTypeData = []struct {
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

// MatchType holds the type of a match.
type MatchType byte

// EnsureValid ensures this is of a known value.
func (enum MatchType) EnsureValid() MatchType {
	if enum <= LastMatchType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum MatchType) Key() string {
	return matchTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum MatchType) String() string {
	return matchTypeData[enum.EnsureValid()].string
}

// ExtractMatchType extracts the value from a string.
func ExtractMatchType(str string) MatchType {
	for i, one := range matchTypeData {
		if strings.EqualFold(one.key, str) {
			return MatchType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum MatchType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *MatchType) UnmarshalText(text []byte) error {
	*enum = ExtractMatchType(string(text))
	return nil
}
