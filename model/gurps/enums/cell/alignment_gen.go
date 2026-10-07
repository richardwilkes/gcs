// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package cell

import (
	"strings"
)

// Possible values.
const (
	AlignStart Alignment = iota
	AlignMiddle
	AlignEnd
)

// DefaultAlignment is the default value.
const DefaultAlignment Alignment = AlignStart

// FirstAlignment is the first valid value.
const FirstAlignment Alignment = AlignStart

// LastAlignment is the last valid value.
const LastAlignment Alignment = AlignEnd

// Alignments holds all possible values.
var Alignments = []Alignment{
	AlignStart,
	AlignMiddle,
	AlignEnd,
}

// Alignment holds the horizontal alignment of a table cell's content.
type Alignment byte

// EnsureValid returns this value if it is valid and DefaultAlignment otherwise.
func (enum Alignment) EnsureValid() Alignment {
	if enum >= FirstAlignment && enum <= LastAlignment {
		return enum
	}
	return DefaultAlignment
}

// Key returns the key used in serialization.
func (enum Alignment) Key() string {
	switch enum {
	case AlignStart:
		return "start"
	case AlignMiddle:
		return "middle"
	case AlignEnd:
		return "end"
	default:
		return DefaultAlignment.Key()
	}
}

// String implements fmt.Stringer.
func (enum Alignment) String() string {
	switch enum {
	case AlignStart:
		return `Start`
	case AlignMiddle:
		return `Middle`
	case AlignEnd:
		return `End`
	default:
		return DefaultAlignment.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Alignment) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Alignment) UnmarshalText(text []byte) error {
	*enum = ExtractAlignment(string(text))
	return nil
}

// ExtractAlignment returns the value whose key matches str, ignoring case, or DefaultAlignment if none does.
func ExtractAlignment(str string) Alignment {
	for _, enum := range Alignments {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultAlignment
}

// ExtractKnownAlignment is like ExtractAlignment, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownAlignment(str string) (value Alignment, known bool) {
	for _, enum := range Alignments {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultAlignment, false
}
