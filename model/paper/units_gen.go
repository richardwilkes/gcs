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
)

// Possible values.
const (
	Inch Units = iota
	Centimeter
	Millimeter
	LastUnits = Millimeter
)

var (
	// AllUnits holds all possible values.
	AllUnits = []Units{
		Inch,
		Centimeter,
		Millimeter,
	}
	unitsData = []struct {
		key    string
		string string
	}{
		{
			key:    "in",
			string: "in",
		},
		{
			key:    "cm",
			string: "cm",
		},
		{
			key:    "mm",
			string: "mm",
		},
	}
)

// Units holds the real-world length unit type.
type Units byte

// EnsureValid ensures this is of a known value.
func (enum Units) EnsureValid() Units {
	if enum <= LastUnits {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Units) Key() string {
	return unitsData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Units) String() string {
	return unitsData[enum.EnsureValid()].string
}

// ExtractUnits extracts the value from a string.
func ExtractUnits(str string) Units {
	for i, one := range unitsData {
		if strings.EqualFold(one.key, str) {
			return Units(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Units) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Units) UnmarshalText(text []byte) error {
	*enum = ExtractUnits(string(text))
	return nil
}
