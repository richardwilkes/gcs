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
	AsNeeded DecimalPlace = iota
	ZeroPlaces
	OnePlace
	TwoPlaces
	ThreePlaces
	FourPlaces
)

// DefaultDecimalPlace is the default value.
const DefaultDecimalPlace DecimalPlace = AsNeeded

// FirstDecimalPlace is the first valid value.
const FirstDecimalPlace DecimalPlace = AsNeeded

// LastDecimalPlace is the last valid value.
const LastDecimalPlace DecimalPlace = FourPlaces

// DecimalPlaces holds all possible values.
var DecimalPlaces = []DecimalPlace{
	AsNeeded,
	ZeroPlaces,
	OnePlace,
	TwoPlaces,
	ThreePlaces,
	FourPlaces,
}

// DecimalPlace holds the number of decimal places to display for a value. AsNeeded shows as many as the value has (up
// to the fixed-point maximum) with trailing zeros removed, while the explicit choices round to that many places.
type DecimalPlace byte

// EnsureValid ensures this is of a known value.
func (enum DecimalPlace) EnsureValid() DecimalPlace {
	if enum >= FirstDecimalPlace && enum <= LastDecimalPlace {
		return enum
	}
	return DefaultDecimalPlace
}

// Key returns the key used in serialization.
func (enum DecimalPlace) Key() string {
	switch enum {
	case AsNeeded:
		return "as_needed"
	case ZeroPlaces:
		return "0"
	case OnePlace:
		return "1"
	case TwoPlaces:
		return "2"
	case ThreePlaces:
		return "3"
	case FourPlaces:
		return "4"
	default:
		return DefaultDecimalPlace.Key()
	}
}

// String implements fmt.Stringer.
func (enum DecimalPlace) String() string {
	switch enum {
	case AsNeeded:
		return i18n.Text(`As Needed`)
	case ZeroPlaces:
		return `0`
	case OnePlace:
		return `1`
	case TwoPlaces:
		return `2`
	case ThreePlaces:
		return `3`
	case FourPlaces:
		return `4`
	default:
		return DefaultDecimalPlace.String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum DecimalPlace) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *DecimalPlace) UnmarshalText(text []byte) error {
	*enum = ExtractDecimalPlace(string(text))
	return nil
}

// ExtractDecimalPlace extracts the value from a string.
func ExtractDecimalPlace(str string) DecimalPlace {
	for _, enum := range DecimalPlaces {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultDecimalPlace
}

// ExtractKnownDecimalPlace extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractDecimalPlace, which quietly maps anything it doesn't recognize onto the first value, this permits a
// caller that is dispatching on the type to detect unknown types.
func ExtractKnownDecimalPlace(str string) (value DecimalPlace, known bool) {
	for _, enum := range DecimalPlaces {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultDecimalPlace, false
}
