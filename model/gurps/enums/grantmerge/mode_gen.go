// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package grantmerge

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Ignore Mode = iota
	MergeLevels
	MergePoints
	KeepHighest
)

// DefaultMode is the default value.
const DefaultMode Mode = Ignore

// FirstMode is the first valid value.
const FirstMode Mode = Ignore

// LastMode is the last valid value.
const LastMode Mode = KeepHighest

// Modes holds all possible values.
var Modes = []Mode{
	Ignore,
	MergeLevels,
	MergePoints,
	KeepHighest,
}

// Mode holds what a grant does when the character already has the item it grants.
type Mode byte

// EnsureValid returns this value if it is valid and DefaultMode otherwise.
func (enum Mode) EnsureValid() Mode {
	if enum >= FirstMode && enum <= LastMode {
		return enum
	}
	return DefaultMode
}

// Key returns the key used in serialization.
func (enum Mode) Key() string {
	switch enum {
	case Ignore:
		return "ignore"
	case MergeLevels:
		return "merge_levels"
	case MergePoints:
		return "merge_points"
	case KeepHighest:
		return "keep_highest"
	default:
		return DefaultMode.Key()
	}
}

// String implements fmt.Stringer.
func (enum Mode) String() string {
	switch enum {
	case Ignore:
		return i18n.Text(`Ignore`)
	case MergeLevels:
		return i18n.Text(`Merge Levels`)
	case MergePoints:
		return i18n.Text(`Merge Points`)
	case KeepHighest:
		return i18n.Text(`Keep Highest`)
	default:
		return DefaultMode.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Mode) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Mode) UnmarshalText(text []byte) error {
	*enum = ExtractMode(string(text))
	return nil
}

// ExtractMode returns the value whose key matches str, ignoring case, or DefaultMode if none does.
func ExtractMode(str string) Mode {
	for _, enum := range Modes {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultMode
}

// ExtractKnownMode is like ExtractMode, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownMode(str string) (value Mode, known bool) {
	for _, enum := range Modes {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultMode, false
}
