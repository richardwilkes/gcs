// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package levelcap

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Group Type = iota
	Fixed
	Skill
	Spell
	Trait
	Attribute
	Unknown
)

// DefaultType is the default value.
const DefaultType Type = Unknown

// FirstType is the first valid value.
const FirstType Type = Group

// LastType is the last valid value.
const LastType Type = Unknown

// Types holds all possible values.
var Types = []Type{
	Group,
	Fixed,
	Skill,
	Spell,
	Trait,
	Attribute,
}

// Type holds the type of a LevelCap.
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
	case Group:
		return "group"
	case Fixed:
		return "fixed"
	case Skill:
		return "skill"
	case Spell:
		return "spell"
	case Trait:
		return "trait"
	case Attribute:
		return "attribute"
	default:
		return "unknown"
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Group:
		return i18n.Text(`the lowest or highest of`)
	case Fixed:
		return i18n.Text(`a fixed level`)
	case Skill:
		return i18n.Text(`a skill's level`)
	case Spell:
		return i18n.Text(`a spell's level`)
	case Trait:
		return i18n.Text(`a trait's level`)
	case Attribute:
		return i18n.Text(`an attribute's value`)
	case Unknown:
		return i18n.Text(`an unknown level cap type`)
	default:
		return i18n.Text(`an unknown level cap type`)
	}
}

// AltString returns the alternate string.
func (enum Type) AltString() string {
	switch enum {
	case Group:
		return i18n.Text(`Group`)
	case Fixed:
		return i18n.Text(`Fixed Level`)
	case Skill:
		return i18n.Text(`Skill Level`)
	case Spell:
		return i18n.Text(`Spell Level`)
	case Trait:
		return i18n.Text(`Trait Level`)
	case Attribute:
		return i18n.Text(`Attribute Value`)
	case Unknown:
		return i18n.Text(`Unknown`)
	default:
		return i18n.Text(`Unknown`)
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
