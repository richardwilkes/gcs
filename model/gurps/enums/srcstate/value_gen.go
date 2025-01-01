// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package srcstate

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Custom Value = iota
	Matched
	Mismatched
	Missing
)

// LastValue is the last valid value.
const LastValue Value = Missing

// Values holds all possible values.
var Values = []Value{
	Custom,
	Matched,
	Mismatched,
	Missing,
}

// Value describes the state of a source compared to a piece of data.
type Value byte

// EnsureValid ensures this is of a known value.
func (enum Value) EnsureValid() Value {
	if enum <= Missing {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Value) Key() string {
	switch enum {
	case Custom:
		return "custom"
	case Matched:
		return "matched"
	case Mismatched:
		return "mismatched"
	case Missing:
		return "missing"
	default:
		return Value(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Value) String() string {
	switch enum {
	case Custom:
		return i18n.Text("Custom data which did not come from a library source")
	case Matched:
		return i18n.Text("Data matches library source data")
	case Mismatched:
		return i18n.Text("Data does NOT match library source data")
	case Missing:
		return i18n.Text("Unable to locate the library source data to compare against")
	default:
		return Value(0).String()
	}
}

// AltString returns the alternate string.
func (enum Value) AltString() string {
	switch enum {
	case Custom:
		return "—"
	case Matched:
		return ""
	case Mismatched:
		return "!"
	case Missing:
		return "?"
	default:
		return Value(0).AltString()
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
