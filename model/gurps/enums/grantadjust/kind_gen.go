// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package grantadjust

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	AddLevels Kind = iota
	AddPoints
	SetLevel
)

// DefaultKind is the default value.
const DefaultKind Kind = AddLevels

// FirstKind is the first valid value.
const FirstKind Kind = AddLevels

// LastKind is the last valid value.
const LastKind Kind = SetLevel

// Kinds holds all possible values.
var Kinds = []Kind{
	AddLevels,
	AddPoints,
	SetLevel,
}

// Kind holds how the amount of a grant adjusts the item it grants.
type Kind byte

// EnsureValid returns this value if it is valid and DefaultKind otherwise.
func (enum Kind) EnsureValid() Kind {
	if enum >= FirstKind && enum <= LastKind {
		return enum
	}
	return DefaultKind
}

// Key returns the key used in serialization.
func (enum Kind) Key() string {
	switch enum {
	case AddLevels:
		return "add_levels"
	case AddPoints:
		return "add_points"
	case SetLevel:
		return "set_level"
	default:
		return DefaultKind.Key()
	}
}

// String implements fmt.Stringer.
func (enum Kind) String() string {
	switch enum {
	case AddLevels:
		return i18n.Text(`Add Levels`)
	case AddPoints:
		return i18n.Text(`Add Points`)
	case SetLevel:
		return i18n.Text(`Set Level`)
	default:
		return DefaultKind.String()
	}
}

// AltString returns the alternate string.
func (enum Kind) AltString() string {
	switch enum {
	case AddLevels:
		return i18n.Text(`adjusting its level by`)
	case AddPoints:
		return i18n.Text(`adjusting its points by`)
	case SetLevel:
		return i18n.Text(`setting its level to`)
	default:
		return DefaultKind.AltString()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Kind) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Kind) UnmarshalText(text []byte) error {
	*enum = ExtractKind(string(text))
	return nil
}

// ExtractKind returns the value whose key matches str, ignoring case, or DefaultKind if none does.
func ExtractKind(str string) Kind {
	for _, enum := range Kinds {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultKind
}

// ExtractKnownKind is like ExtractKind, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownKind(str string) (value Kind, known bool) {
	for _, enum := range Kinds {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultKind, false
}
