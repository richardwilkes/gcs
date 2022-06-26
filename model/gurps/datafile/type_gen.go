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
	PC       Type = iota
	LastType      = PC
)

var (
	// AllType holds all possible values.
	AllType = []Type{
		PC,
	}
	typeData = []struct {
		key    string
		string string
	}{
		{
			key:    "character",
			string: i18n.Text("PC"),
		},
	}
)

// Type holds the type of an Entity.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= LastType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	return typeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	return typeData[enum.EnsureValid()].string
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for i, one := range typeData {
		if strings.EqualFold(one.key, str) {
			return Type(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}
