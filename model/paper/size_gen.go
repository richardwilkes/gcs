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

package paper

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Letter Size = iota
	Legal
	Tabloid
	A0
	A1
	A2
	A3
	A4
	A5
	A6
	LastSize = A6
)

var (
	// AllSize holds all possible values.
	AllSize = []Size{
		Letter,
		Legal,
		Tabloid,
		A0,
		A1,
		A2,
		A3,
		A4,
		A5,
		A6,
	}
	sizeData = []struct {
		key    string
		string string
	}{
		{
			key:    "letter",
			string: i18n.Text("Letter"),
		},
		{
			key:    "legal",
			string: i18n.Text("Legal"),
		},
		{
			key:    "tabloid",
			string: i18n.Text("Tabloid"),
		},
		{
			key:    "a0",
			string: i18n.Text("A0"),
		},
		{
			key:    "a1",
			string: i18n.Text("A1"),
		},
		{
			key:    "a2",
			string: i18n.Text("A2"),
		},
		{
			key:    "a3",
			string: i18n.Text("A3"),
		},
		{
			key:    "a4",
			string: i18n.Text("A4"),
		},
		{
			key:    "a5",
			string: i18n.Text("A5"),
		},
		{
			key:    "a6",
			string: i18n.Text("A6"),
		},
	}
)

// Size holds a standard paper dimension.
type Size byte

// EnsureValid ensures this is of a known value.
func (enum Size) EnsureValid() Size {
	if enum <= LastSize {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Size) Key() string {
	return sizeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Size) String() string {
	return sizeData[enum.EnsureValid()].string
}

// ExtractSize extracts the value from a string.
func ExtractSize(str string) Size {
	str = strings.TrimPrefix(strings.TrimPrefix(str, "na-"), "iso-") // For older files that had the Java prefixes
	for i, one := range sizeData {
		if strings.EqualFold(one.key, str) {
			return Size(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Size) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Size) UnmarshalText(text []byte) error {
	*enum = ExtractSize(string(text))
	return nil
}
