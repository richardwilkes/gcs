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
	Value
	Weight
)

// DefaultType is the default value.
const DefaultType Type = NotApplicable

// FirstType is the first valid value.
const FirstType Type = NotApplicable

// LastType is the last valid value.
const LastType Type = Weight

// Types holds all possible values.
var Types = []Type{
	NotApplicable,
	Count,
	Points,
	Value,
	Weight,
}

// TypesForEquipment holds the Types valid for the "equipment" group.
var TypesForEquipment = []Type{
	NotApplicable,
	Count,
	Value,
	Weight,
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

// EnsureValid returns this value if it is valid and DefaultType otherwise.
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
	case Value:
		return "value"
	case Weight:
		return "weight"
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
	case Value:
		return i18n.Text(`Value`)
	case Weight:
		return i18n.Text(`Weight`)
	default:
		return DefaultType.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}

// ExtractType returns the value whose key matches str, ignoring case, or DefaultType if none does.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultType
}

// ExtractKnownType is like ExtractType, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownType(str string) (value Type, known bool) {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultType, false
}
