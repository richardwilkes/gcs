// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package spellmatch

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	AllColleges Type = iota
	CollegeName
	PowerSource
	Name
)

// DefaultType is the default value.
const DefaultType Type = AllColleges

// FirstType is the first valid value.
const FirstType Type = AllColleges

// LastType is the last valid value.
const LastType Type = Name

// Types holds all possible values.
var Types = []Type{
	AllColleges,
	CollegeName,
	PowerSource,
	Name,
}

// Type holds the type of a match.
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
	case AllColleges:
		return "all_colleges"
	case CollegeName:
		return "college_name"
	case PowerSource:
		return "power_source_name"
	case Name:
		return "spell_name"
	default:
		return DefaultType.Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case AllColleges:
		return i18n.Text(`to all colleges`)
	case CollegeName:
		return i18n.Text(`to the college whose name`)
	case PowerSource:
		return i18n.Text(`to the power source whose name`)
	case Name:
		return i18n.Text(`to the spell whose name`)
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
