// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
)

// Possible values.
const (
	Inch Unit = iota
	Centimeter
	Millimeter
)

// LastUnit is the last valid value.
const LastUnit Unit = Millimeter

// Units holds all possible values.
var Units = []Unit{
	Inch,
	Centimeter,
	Millimeter,
}

// Unit holds the real-world length unit type.
type Unit byte

// EnsureValid ensures this is of a known value.
func (enum Unit) EnsureValid() Unit {
	if enum <= Millimeter {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Unit) Key() string {
	switch enum {
	case Inch:
		return "in"
	case Centimeter:
		return "cm"
	case Millimeter:
		return "mm"
	default:
		return Unit(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Unit) String() string {
	switch enum {
	case Inch:
		return "in"
	case Centimeter:
		return "cm"
	case Millimeter:
		return "mm"
	default:
		return Unit(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Unit) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Unit) UnmarshalText(text []byte) error {
	*enum = ExtractUnit(string(text))
	return nil
}

// ExtractUnit extracts the value from a string.
func ExtractUnit(str string) Unit {
	for _, enum := range Units {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
