// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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

// DefaultUnit is the default value.
const DefaultUnit Unit = Inch

// FirstUnit is the first valid value.
const FirstUnit Unit = Inch

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

// EnsureValid returns this value if it is valid and DefaultUnit otherwise.
func (enum Unit) EnsureValid() Unit {
	if enum >= FirstUnit && enum <= LastUnit {
		return enum
	}
	return DefaultUnit
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
		return DefaultUnit.Key()
	}
}

// String implements fmt.Stringer.
func (enum Unit) String() string {
	switch enum {
	case Inch:
		return `in`
	case Centimeter:
		return `cm`
	case Millimeter:
		return `mm`
	default:
		return DefaultUnit.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Unit) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Unit) UnmarshalText(text []byte) error {
	*enum = ExtractUnit(string(text))
	return nil
}

// ExtractUnit returns the value whose key matches str, ignoring case, or DefaultUnit if none does.
func ExtractUnit(str string) Unit {
	for _, enum := range Units {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultUnit
}

// ExtractKnownUnit is like ExtractUnit, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownUnit(str string) (value Unit, known bool) {
	for _, enum := range Units {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultUnit, false
}
