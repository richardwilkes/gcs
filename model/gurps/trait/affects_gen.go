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

package trait

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Total Affects = iota
	BaseOnly
	LevelsOnly
	LastAffects = LevelsOnly
)

var (
	// AllAffects holds all possible values.
	AllAffects = []Affects{
		Total,
		BaseOnly,
		LevelsOnly,
	}
	affectsData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "total",
			string: i18n.Text("to cost"),
		},
		{
			key:    "base_only",
			string: i18n.Text("to base cost only"),
			alt:    i18n.Text("(base only)"),
		},
		{
			key:    "levels_only",
			string: i18n.Text("to leveled cost only"),
			alt:    i18n.Text("(levels only)"),
		},
	}
)

// Affects describes how a TraitModifier affects the point cost.
type Affects byte

// EnsureValid ensures this is of a known value.
func (enum Affects) EnsureValid() Affects {
	if enum <= LastAffects {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Affects) Key() string {
	return affectsData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Affects) String() string {
	return affectsData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum Affects) AltString() string {
	return affectsData[enum.EnsureValid()].alt
}

// ExtractAffects extracts the value from a string.
func ExtractAffects(str string) Affects {
	for i, one := range affectsData {
		if strings.EqualFold(one.key, str) {
			return Affects(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Affects) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Affects) UnmarshalText(text []byte) error {
	*enum = ExtractAffects(string(text))
	return nil
}
