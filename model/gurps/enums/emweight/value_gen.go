// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emweight

import (
	"strings"
)

// Possible values.
const (
	Addition Value = iota
	PercentageAdder
	PercentageMultiplier
	Multiplier
)

// DefaultValue is the default value.
const DefaultValue Value = Addition

// FirstValue is the first valid value.
const FirstValue Value = Addition

// LastValue is the last valid value.
const LastValue Value = Multiplier

// Values holds all possible values.
var Values = []Value{
	Addition,
	PercentageAdder,
	PercentageMultiplier,
	Multiplier,
}

// Value describes how an Equipment Modifier's weight value is applied.
type Value byte

// EnsureValid returns this value if it is valid and DefaultValue otherwise.
func (enum Value) EnsureValid() Value {
	if enum >= FirstValue && enum <= LastValue {
		return enum
	}
	return DefaultValue
}

// Key returns the key used in serialization.
func (enum Value) Key() string {
	switch enum {
	case Addition:
		return "+"
	case PercentageAdder:
		return "%"
	case PercentageMultiplier:
		return "x%"
	case Multiplier:
		return "x"
	default:
		return DefaultValue.Key()
	}
}

// String implements fmt.Stringer.
func (enum Value) String() string {
	switch enum {
	case Addition:
		return `+`
	case PercentageAdder:
		return `%`
	case PercentageMultiplier:
		return `x%`
	case Multiplier:
		return `x`
	default:
		return DefaultValue.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Value) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Value) UnmarshalText(text []byte) error {
	*enum = ExtractValue(string(text))
	return nil
}

// ExtractValue returns the value whose key matches str, ignoring case, or DefaultValue if none does.
func ExtractValue(str string) Value {
	for _, enum := range Values {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultValue
}

// ExtractKnownValue is like ExtractValue, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownValue(str string) (value Value, known bool) {
	for _, enum := range Values {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultValue, false
}
