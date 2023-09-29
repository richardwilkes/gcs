/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

// Code generated from "enum.go.tmpl" - DO NOT EDIT.

package fxp

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	FeetAndInches LengthUnits = iota
	Inch
	Feet
	Yard
	Mile
	Centimeter
	Kilometer
	Meter
	LastLengthUnits = Meter
)

// AllLengthUnits holds all possible values.
var AllLengthUnits = []LengthUnits{
	FeetAndInches,
	Inch,
	Feet,
	Yard,
	Mile,
	Centimeter,
	Kilometer,
	Meter,
}

// LengthUnits holds the length unit type. Note that conversions to/from metric are done using the simplified GURPS
// metric conversion of 1 yd = 1 meter. For consistency, all metric lengths are converted to meters, then to yards,
// rather than the variations at different lengths that the GURPS rules suggest.
type LengthUnits byte

// EnsureValid ensures this is of a known value.
func (enum LengthUnits) EnsureValid() LengthUnits {
	if enum <= LastLengthUnits {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum LengthUnits) Key() string {
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
		return LengthUnits(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum LengthUnits) String() string {
	switch enum {
	case FeetAndInches:
		return i18n.Text("Feet & Inches")
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
		return LengthUnits(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum LengthUnits) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *LengthUnits) UnmarshalText(text []byte) error {
	*enum = ExtractLengthUnits(string(text))
	return nil
}

// ExtractLengthUnits extracts the value from a string.
func ExtractLengthUnits(str string) LengthUnits {
	for _, enum := range AllLengthUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
