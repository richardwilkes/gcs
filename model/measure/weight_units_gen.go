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
)

// Possible values.
const (
	Pound WeightUnits = iota
	PoundAlt
	Ounce
	Ton
	TonAlt
	Kilogram
	Gram
	LastWeightUnits = Gram
)

var (
	// AllWeightUnits holds all possible values.
	AllWeightUnits = []WeightUnits{
		Pound,
		PoundAlt,
		Ounce,
		Ton,
		TonAlt,
		Kilogram,
		Gram,
	}
	weightUnitsData = []struct {
		key    string
		string string
	}{
		{
			key:    "lb",
			string: "lb",
		},
		{
			key:    "#",
			string: "#",
		},
		{
			key:    "oz",
			string: "oz",
		},
		{
			key:    "tn",
			string: "tn",
		},
		{
			key:    "t",
			string: "t",
		},
		{
			key:    "kg",
			string: "kg",
		},
		{
			key:    "g",
			string: "g",
		},
	}
)

// WeightUnits holds the weight unit type. Note that conversions to/from metric are done using the simplified GURPS
// metric conversion of 1 lb = 0.5kg. For consistency, all metric weights are converted to kilograms, then to pounds,
// rather than the variations at different weights that the GURPS rules suggest.
type WeightUnits byte

// EnsureValid ensures this is of a known value.
func (enum WeightUnits) EnsureValid() WeightUnits {
	if enum <= LastWeightUnits {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum WeightUnits) Key() string {
	return weightUnitsData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum WeightUnits) String() string {
	return weightUnitsData[enum.EnsureValid()].string
}

// ExtractWeightUnits extracts the value from a string.
func ExtractWeightUnits(str string) WeightUnits {
	for i, one := range weightUnitsData {
		if strings.EqualFold(one.key, str) {
			return WeightUnits(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum WeightUnits) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *WeightUnits) UnmarshalText(text []byte) error {
	*enum = ExtractWeightUnits(string(text))
	return nil
}
