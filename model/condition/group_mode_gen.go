// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	GroupAll GroupMode = iota
	GroupAny
	GroupAtLeast
)

// DefaultGroupMode is the default value.
const DefaultGroupMode GroupMode = GroupAll

// FirstGroupMode is the first valid value.
const FirstGroupMode GroupMode = GroupAll

// LastGroupMode is the last valid value.
const LastGroupMode GroupMode = GroupAtLeast

// GroupModes holds all possible values.
var GroupModes = []GroupMode{
	GroupAll,
	GroupAny,
	GroupAtLeast,
}

// GroupMode holds how many of a group's children must be met.
type GroupMode byte

// EnsureValid returns this value if it is valid and DefaultGroupMode otherwise.
func (enum GroupMode) EnsureValid() GroupMode {
	if enum >= FirstGroupMode && enum <= LastGroupMode {
		return enum
	}
	return DefaultGroupMode
}

// Key returns the key used in serialization.
func (enum GroupMode) Key() string {
	switch enum {
	case GroupAll:
		return "all"
	case GroupAny:
		return "any"
	case GroupAtLeast:
		return "at_least"
	default:
		return DefaultGroupMode.Key()
	}
}

// String implements fmt.Stringer.
func (enum GroupMode) String() string {
	switch enum {
	case GroupAll:
		return i18n.Text(`all`)
	case GroupAny:
		return i18n.Text(`any`)
	case GroupAtLeast:
		return i18n.Text(`at least`)
	default:
		return DefaultGroupMode.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum GroupMode) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *GroupMode) UnmarshalText(text []byte) error {
	*enum = ExtractGroupMode(string(text))
	return nil
}

// ExtractGroupMode returns the value whose key matches str, ignoring case, or DefaultGroupMode if none does.
func ExtractGroupMode(str string) GroupMode {
	for _, enum := range GroupModes {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultGroupMode
}

// ExtractKnownGroupMode is like ExtractGroupMode, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownGroupMode(str string) (value GroupMode, known bool) {
	for _, enum := range GroupModes {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultGroupMode, false
}
