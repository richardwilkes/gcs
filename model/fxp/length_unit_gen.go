// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	FeetAndInches LengthUnit = iota
	Inch
	Feet
	Yard
	Mile
	Centimeter
	Kilometer
	Meter
)

// DefaultLengthUnit is the default value.
const DefaultLengthUnit LengthUnit = FeetAndInches

// FirstLengthUnit is the first valid value.
const FirstLengthUnit LengthUnit = FeetAndInches

// LastLengthUnit is the last valid value.
const LastLengthUnit LengthUnit = Meter

// LengthUnits holds all possible values.
var LengthUnits = []LengthUnit{
	FeetAndInches,
	Inch,
	Feet,
	Yard,
	Mile,
	Centimeter,
	Kilometer,
	Meter,
}

// LengthUnit holds the length unit type. Note that conversions to/from metric are done using the simplified GURPS
// metric conversion of 2.5 cm = 1 inch. For consistency, all metric lengths are converted to inches, rather than the
// variations at different lengths that the GURPS rules suggest.
type LengthUnit byte

// EnsureValid returns this value if it is valid and DefaultLengthUnit otherwise.
func (enum LengthUnit) EnsureValid() LengthUnit {
	if enum >= FirstLengthUnit && enum <= LastLengthUnit {
		return enum
	}
	return DefaultLengthUnit
}

// Key returns the key used in serialization.
func (enum LengthUnit) Key() string {
	switch enum {
	case FeetAndInches:
		return "ft_in"
	case Inch:
		return "in"
	case Feet:
		return "ft"
	case Yard:
		return "yd"
	case Mile:
		return "mi"
	case Centimeter:
		return "cm"
	case Kilometer:
		return "km"
	case Meter:
		return "m"
	default:
		return DefaultLengthUnit.Key()
	}
}

// String implements fmt.Stringer.
func (enum LengthUnit) String() string {
	switch enum {
	case FeetAndInches:
		return i18n.Text(`Feet & Inches`)
	case Inch:
		return `in`
	case Feet:
		return `ft`
	case Yard:
		return `yd`
	case Mile:
		return `mi`
	case Centimeter:
		return `cm`
	case Kilometer:
		return `km`
	case Meter:
		return `m`
	default:
		return DefaultLengthUnit.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum LengthUnit) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *LengthUnit) UnmarshalText(text []byte) error {
	*enum = ExtractLengthUnit(string(text))
	return nil
}

// ExtractLengthUnit returns the value whose key matches str, ignoring case, or DefaultLengthUnit if none does.
func ExtractLengthUnit(str string) LengthUnit {
	for _, enum := range LengthUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultLengthUnit
}

// ExtractKnownLengthUnit is like ExtractLengthUnit, but also reports whether str was recognized, so a caller can tell
// an unknown key from the default.
func ExtractKnownLengthUnit(str string) (value LengthUnit, known bool) {
	for _, enum := range LengthUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultLengthUnit, false
}
