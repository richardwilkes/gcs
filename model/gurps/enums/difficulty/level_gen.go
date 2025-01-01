// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package difficulty

import (
	"strings"
)

// Possible values.
const (
	Easy Level = iota
	Average
	Hard
	VeryHard
	Wildcard
)

// LastLevel is the last valid value.
const LastLevel Level = Wildcard

// Levels holds all possible values.
var Levels = []Level{
	Easy,
	Average,
	Hard,
	VeryHard,
	Wildcard,
}

// Level holds the difficulty level of a skill.
type Level byte

// EnsureValid ensures this is of a known value.
func (enum Level) EnsureValid() Level {
	if enum <= Wildcard {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Level) Key() string {
	switch enum {
	case Easy:
		return "e"
	case Average:
		return "a"
	case Hard:
		return "h"
	case VeryHard:
		return "vh"
	case Wildcard:
		return "w"
	default:
		return Level(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Level) String() string {
	switch enum {
	case Easy:
		return "E"
	case Average:
		return "A"
	case Hard:
		return "H"
	case VeryHard:
		return "VH"
	case Wildcard:
		return "W"
	default:
		return Level(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Level) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Level) UnmarshalText(text []byte) error {
	*enum = ExtractLevel(string(text))
	return nil
}

// ExtractLevel extracts the value from a string.
func ExtractLevel(str string) Level {
	for _, enum := range Levels {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
