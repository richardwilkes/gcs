// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package study

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Self Type = iota
	Job
	Teacher
	Intensive
)

// DefaultType is the default value.
const DefaultType Type = Self

// FirstType is the first valid value.
const FirstType Type = Self

// LastType is the last valid value.
const LastType Type = Intensive

// Types holds all possible values.
var Types = []Type{
	Self,
	Job,
	Teacher,
	Intensive,
}

// Type holds the type of study.
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
	case Self:
		return "self"
	case Job:
		return "job"
	case Teacher:
		return "teacher"
	case Intensive:
		return "intensive"
	default:
		return DefaultType.Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case Self:
		return i18n.Text(`Self-Taught`)
	case Job:
		return i18n.Text(`On-the-Job Training`)
	case Teacher:
		return i18n.Text(`Professional Teacher`)
	case Intensive:
		return i18n.Text(`Intensive Training`)
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
