// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package spellcmp

import (
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Name Type = iota
	Tag
	College
	CollegeCount
	Any
)

// DefaultType is the default value.
const DefaultType Type = Name

// FirstType is the first valid value.
const FirstType Type = Name

// LastType is the last valid value.
const LastType Type = Any

// Types holds all possible values.
var Types = []Type{
	Name,
	Tag,
	College,
	CollegeCount,
	Any,
}

// Type holds the type of a comparison.
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
	case Name:
		return "name"
	case Tag:
		return "tag"
	case College:
		return "college"
	case CollegeCount:
		return "college_count"
	case Any:
		return "any"
	default:
		return DefaultType.Key()
	}
}

func (enum Type) oldKeys() []string {
	switch enum {
	case Name:
		return nil
	case Tag:
		return []string{"category"}
	case College:
		return nil
	case CollegeCount:
		return nil
	case Any:
		return nil
	default:
		return DefaultType.oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Name:
		return i18n.Text(`whose name`)
	case Tag:
		return i18n.Text(`with a tag which`)
	case College:
		return i18n.Text(`whose college name`)
	case CollegeCount:
		return i18n.Text(`from different colleges`)
	case Any:
		return i18n.Text(`of any kind`)
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
		if slices.ContainsFunc(enum.oldKeys(), func(s string) bool { return strings.EqualFold(s, str) }) {
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
		if slices.ContainsFunc(enum.oldKeys(), func(s string) bool { return strings.EqualFold(s, str) }) {
			return enum, true
		}
	}
	return DefaultType, false
}
