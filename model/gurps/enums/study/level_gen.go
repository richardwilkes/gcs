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
	Standard Level = iota
	Level1
	Level2
	Level3
	Level4
)

// DefaultLevel is the default value.
const DefaultLevel Level = Standard

// FirstLevel is the first valid value.
const FirstLevel Level = Standard

// LastLevel is the last valid value.
const LastLevel Level = Level4

// Levels holds all possible values.
var Levels = []Level{
	Standard,
	Level1,
	Level2,
	Level3,
	Level4,
}

// Level holds the number of study hours required per point.
type Level byte

// EnsureValid returns this value if it is valid and DefaultLevel otherwise.
func (enum Level) EnsureValid() Level {
	if enum >= FirstLevel && enum <= LastLevel {
		return enum
	}
	return DefaultLevel
}

// Key returns the key used in serialization.
func (enum Level) Key() string {
	switch enum {
	case Standard:
		return ""
	case Level1:
		return "180"
	case Level2:
		return "160"
	case Level3:
		return "140"
	case Level4:
		return "120"
	default:
		return DefaultLevel.Key()
	}
}

// String implements fmt.Stringer.
func (enum Level) String() string {
	switch enum {
	case Standard:
		return i18n.Text(`Standard`)
	case Level1:
		return i18n.Text(`Reduction for Talent level 1`)
	case Level2:
		return i18n.Text(`Reduction for Talent level 2`)
	case Level3:
		return i18n.Text(`Reduction for Talent level 3`)
	case Level4:
		return i18n.Text(`Reduction for Talent level 4`)
	default:
		return DefaultLevel.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Level) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Level) UnmarshalText(text []byte) error {
	*enum = ExtractLevel(string(text))
	return nil
}

// ExtractLevel returns the value whose key matches str, ignoring case, or DefaultLevel if none does.
func ExtractLevel(str string) Level {
	for _, enum := range Levels {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultLevel
}

// ExtractKnownLevel is like ExtractLevel, but also reports whether str was recognized, so a caller can tell an unknown
// key from the default.
func ExtractKnownLevel(str string) (value Level, known bool) {
	for _, enum := range Levels {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultLevel, false
}
