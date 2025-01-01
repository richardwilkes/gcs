// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
)

// Possible values.
const (
	Pound WeightUnit = iota
	PoundAlt
	Ounce
	Ton
	TonAlt
	Kilogram
	Gram
)

// LastWeightUnit is the last valid value.
const LastWeightUnit WeightUnit = Gram

// WeightUnits holds all possible values.
var WeightUnits = []WeightUnit{
	Pound,
	PoundAlt,
	Ounce,
	Ton,
	TonAlt,
	Kilogram,
	Gram,
}

// WeightUnit holds the weight unit type. Note that conversions to/from metric are done using the simplified GURPS
// metric conversion of 1 lb = 0.5kg. For consistency, all metric weights are converted to kilograms, then to pounds,
// rather than the variations at different weights that the GURPS rules suggest.
type WeightUnit byte

// EnsureValid ensures this is of a known value.
func (enum WeightUnit) EnsureValid() WeightUnit {
	if enum <= Gram {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum WeightUnit) Key() string {
	switch enum {
	case Pound:
		return "lb"
	case PoundAlt:
		return "#"
	case Ounce:
		return "oz"
	case Ton:
		return "tn"
	case TonAlt:
		return "t"
	case Kilogram:
		return "kg"
	case Gram:
		return "g"
	default:
		return WeightUnit(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum WeightUnit) String() string {
	switch enum {
	case Pound:
		return "lb"
	case PoundAlt:
		return "#"
	case Ounce:
		return "oz"
	case Ton:
		return "tn"
	case TonAlt:
		return "t"
	case Kilogram:
		return "kg"
	case Gram:
		return "g"
	default:
		return WeightUnit(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum WeightUnit) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *WeightUnit) UnmarshalText(text []byte) error {
	*enum = ExtractWeightUnit(string(text))
	return nil
}

// ExtractWeightUnit extracts the value from a string.
func ExtractWeightUnit(str string) WeightUnit {
	for _, enum := range WeightUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
