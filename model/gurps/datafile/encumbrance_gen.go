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

package datafile

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	None Encumbrance = iota
	Light
	Medium
	Heavy
	ExtraHeavy
	LastEncumbrance = ExtraHeavy
)

var (
	// AllEncumbrance holds all possible values.
	AllEncumbrance = []Encumbrance{
		None,
		Light,
		Medium,
		Heavy,
		ExtraHeavy,
	}
	encumbranceData = []struct {
		key    string
		string string
	}{
		{
			key:    "none",
			string: i18n.Text("None"),
		},
		{
			key:    "light",
			string: i18n.Text("Light"),
		},
		{
			key:    "medium",
			string: i18n.Text("Medium"),
		},
		{
			key:    "heavy",
			string: i18n.Text("Heavy"),
		},
		{
			key:    "extra_heavy",
			string: i18n.Text("X-Heavy"),
		},
	}
)

// Encumbrance holds the encumbrance level.
type Encumbrance byte

// EnsureValid ensures this is of a known value.
func (enum Encumbrance) EnsureValid() Encumbrance {
	if enum <= LastEncumbrance {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Encumbrance) Key() string {
	return encumbranceData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Encumbrance) String() string {
	return encumbranceData[enum.EnsureValid()].string
}

// ExtractEncumbrance extracts the value from a string.
func ExtractEncumbrance(str string) Encumbrance {
	for i, one := range encumbranceData {
		if strings.EqualFold(one.key, str) {
			return Encumbrance(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Encumbrance) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Encumbrance) UnmarshalText(text []byte) error {
	*enum = ExtractEncumbrance(string(text))
	return nil
}
