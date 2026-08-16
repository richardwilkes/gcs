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
// metric conversion of 1 yd = 1 meter. For consistency, all metric lengths are converted to meters, then to yards,
// rather than the variations at different lengths that the GURPS rules suggest.
type LengthUnit byte

// EnsureValid ensures this is of a known value.
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

// MarshalText implements the encoding.TextMarshaler interface.
func (enum LengthUnit) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *LengthUnit) UnmarshalText(text []byte) error {
	*enum = ExtractLengthUnit(string(text))
	return nil
}

// ExtractLengthUnit extracts the value from a string.
func ExtractLengthUnit(str string) LengthUnit {
	for _, enum := range LengthUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultLengthUnit
}

// ExtractKnownLengthUnit extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractLengthUnit, which quietly maps anything it doesn't recognize onto the first value, this permits a
// caller that is dispatching on the type to detect unknown types.
func ExtractKnownLengthUnit(str string) (value LengthUnit, known bool) {
	for _, enum := range LengthUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultLengthUnit, false
}
