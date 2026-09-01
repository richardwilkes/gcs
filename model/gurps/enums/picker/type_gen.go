// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package picker

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	NotApplicable Type = iota
	Count
	Points
)

// DefaultType is the default value.
const DefaultType Type = NotApplicable

// FirstType is the first valid value.
const FirstType Type = NotApplicable

// LastType is the last valid value.
const LastType Type = Points

// Types holds all possible values.
var Types = []Type{
	NotApplicable,
	Count,
	Points,
}

// TypesForSkills holds the Types valid for the "skills" group.
var TypesForSkills = []Type{
	NotApplicable,
	Count,
	Points,
}

// TypesForSpells holds the Types valid for the "spells" group.
var TypesForSpells = []Type{
	NotApplicable,
	Count,
	Points,
}

// TypesForTraits holds the Types valid for the "traits" group.
var TypesForTraits = []Type{
	NotApplicable,
	Count,
	Points,
}

// Type holds the type of template picker.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum >= FirstType && enum <= LastType {
		return enum
	}
	return DefaultType
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case NotApplicable:
		return "not_applicable"
	case Count:
		return "count"
	case Points:
		return "points"
	default:
		return DefaultType.Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case NotApplicable:
		return i18n.Text(`Not Applicable`)
	case Count:
		return i18n.Text(`Count`)
	case Points:
		return i18n.Text(`Points`)
	default:
		return DefaultType.String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultType
}

// ExtractKnownType extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractType, which quietly maps anything it doesn't recognize onto the first value, this permits a caller that
// is dispatching on the type to detect unknown types.
func ExtractKnownType(str string) (value Type, known bool) {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultType, false
}
