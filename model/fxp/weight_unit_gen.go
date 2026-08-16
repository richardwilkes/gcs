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

// DefaultWeightUnit is the default value.
const DefaultWeightUnit WeightUnit = Pound

// FirstWeightUnit is the first valid value.
const FirstWeightUnit WeightUnit = Pound

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
	if enum >= FirstWeightUnit && enum <= LastWeightUnit {
		return enum
	}
	return DefaultWeightUnit
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
		return DefaultWeightUnit.Key()
	}
}

// String implements fmt.Stringer.
func (enum WeightUnit) String() string {
	switch enum {
	case Pound:
		return `lb`
	case PoundAlt:
		return `#`
	case Ounce:
		return `oz`
	case Ton:
		return `tn`
	case TonAlt:
		return `t`
	case Kilogram:
		return `kg`
	case Gram:
		return `g`
	default:
		return DefaultWeightUnit.String()
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
	return DefaultWeightUnit
}

// ExtractKnownWeightUnit extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractWeightUnit, which quietly maps anything it doesn't recognize onto the first value, this permits a
// caller that is dispatching on the type to detect unknown types.
func ExtractKnownWeightUnit(str string) (value WeightUnit, known bool) {
	for _, enum := range WeightUnits {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultWeightUnit, false
}
