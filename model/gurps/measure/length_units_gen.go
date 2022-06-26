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

package measure

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

var (
	// AllLengthUnits holds all possible values.
	AllLengthUnits = []LengthUnits{
		FeetAndInches,
		Inch,
		Feet,
		Yard,
		Mile,
		Centimeter,
		Kilometer,
		Meter,
	}
	lengthUnitsData = []struct {
		key    string
		string string
	}{
		{
			key:    "ft_in",
			string: i18n.Text("Feet & Inches"),
		},
		{
			key:    "in",
			string: "in",
		},
		{
			key:    "ft",
			string: "ft",
		},
		{
			key:    "yd",
			string: "yd",
		},
		{
			key:    "mi",
			string: "mi",
		},
		{
			key:    "cm",
			string: "cm",
		},
		{
			key:    "km",
			string: "km",
		},
		{
			key:    "m",
			string: "m",
		},
	}
)

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
	return lengthUnitsData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum LengthUnits) String() string {
	return lengthUnitsData[enum.EnsureValid()].string
}

// ExtractLengthUnits extracts the value from a string.
func ExtractLengthUnits(str string) LengthUnits {
	for i, one := range lengthUnitsData {
		if strings.EqualFold(one.key, str) {
			return LengthUnits(i)
		}
	}
	return 0
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
