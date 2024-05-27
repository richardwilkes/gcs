// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

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
)

// LastSize is the last valid value.
const LastSize Size = A6

// Sizes holds all possible values.
var Sizes = []Size{
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

// Size holds a standard paper dimension.
type Size byte

// EnsureValid ensures this is of a known value.
func (enum Size) EnsureValid() Size {
	if enum <= A6 {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Size) Key() string {
	switch enum {
	case Letter:
		return "letter"
	case Legal:
		return "legal"
	case Tabloid:
		return "tabloid"
	case A0:
		return "a0"
	case A1:
		return "a1"
	case A2:
		return "a2"
	case A3:
		return "a3"
	case A4:
		return "a4"
	case A5:
		return "a5"
	case A6:
		return "a6"
	default:
		return Size(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Size) String() string {
	switch enum {
	case Letter:
		return i18n.Text("Letter")
	case Legal:
		return i18n.Text("Legal")
	case Tabloid:
		return i18n.Text("Tabloid")
	case A0:
		return i18n.Text("A0")
	case A1:
		return i18n.Text("A1")
	case A2:
		return i18n.Text("A2")
	case A3:
		return i18n.Text("A3")
	case A4:
		return i18n.Text("A4")
	case A5:
		return i18n.Text("A5")
	case A6:
		return i18n.Text("A6")
	default:
		return Size(0).String()
	}
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

// ExtractSize extracts the value from a string.
func ExtractSize(str string) Size {
	str = strings.TrimPrefix(strings.TrimPrefix(str, "na-"), "iso-") // For older files that had the Java prefixes
	for _, enum := range Sizes {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
