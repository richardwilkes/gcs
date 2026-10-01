// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package layoutedge

import (
	"strings"
)

// Possible values.
const (
	Top Enum = iota
	Left
	Bottom
	Right
)

// DefaultEnum is the default value.
const DefaultEnum Enum = Top

// FirstEnum is the first valid value.
const FirstEnum Enum = Top

// LastEnum is the last valid value.
const LastEnum Enum = Right

// Enums holds all possible values.
var Enums = []Enum{
	Top,
	Left,
	Bottom,
	Right,
}

// Enum holds the edge of a sheet layout block that a drop targets.
type Enum byte

// EnsureValid returns this value if it is valid and DefaultEnum otherwise.
func (enum Enum) EnsureValid() Enum {
	if enum >= FirstEnum && enum <= LastEnum {
		return enum
	}
	return DefaultEnum
}

// Key returns the key used in serialization.
func (enum Enum) Key() string {
	switch enum {
	case Top:
		return "top"
	case Left:
		return "left"
	case Bottom:
		return "bottom"
	case Right:
		return "right"
	default:
		return DefaultEnum.Key()
	}
}

// String implements fmt.Stringer.
func (enum Enum) String() string {
	switch enum {
	case Top:
		return `Top`
	case Left:
		return `Left`
	case Bottom:
		return `Bottom`
	case Right:
		return `Right`
	default:
		return DefaultEnum.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Enum) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Enum) UnmarshalText(text []byte) error {
	*enum = ExtractEnum(string(text))
	return nil
}

// ExtractEnum returns the value whose key matches str, ignoring case, or DefaultEnum if none does.
func ExtractEnum(str string) Enum {
	for _, enum := range Enums {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultEnum
}

// ExtractKnownEnum is like ExtractEnum, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownEnum(str string) (value Enum, known bool) {
	for _, enum := range Enums {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultEnum, false
}
