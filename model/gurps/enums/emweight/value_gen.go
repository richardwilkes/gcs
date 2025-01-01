// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

// EnsureValid ensures this is of a known value.
func (enum Value) EnsureValid() Value {
	if enum <= Multiplier {
		return enum
	}
	return 0
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
		return Value(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Value) String() string {
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
		return Value(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Value) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Value) UnmarshalText(text []byte) error {
	*enum = ExtractValue(string(text))
	return nil
}

// ExtractValue extracts the value from a string.
func ExtractValue(str string) Value {
	for _, enum := range Values {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
