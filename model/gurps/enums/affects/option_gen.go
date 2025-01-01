// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package affects

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Total Option = iota
	BaseOnly
	LevelsOnly
)

// LastOption is the last valid value.
const LastOption Option = LevelsOnly

// Options holds all possible values.
var Options = []Option{
	Total,
	BaseOnly,
	LevelsOnly,
}

// Option describes how a TraitModifier affects the point cost.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= LevelsOnly {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
	switch enum {
	case Total:
		return "total"
	case BaseOnly:
		return "base_only"
	case LevelsOnly:
		return "levels_only"
	default:
		return Option(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case Total:
		return i18n.Text("to cost")
	case BaseOnly:
		return i18n.Text("to base cost only")
	case LevelsOnly:
		return i18n.Text("to leveled cost only")
	default:
		return Option(0).String()
	}
}

// AltString returns the alternate string.
func (enum Option) AltString() string {
	switch enum {
	case Total:
		return ""
	case BaseOnly:
		return i18n.Text("(base only)")
	case LevelsOnly:
		return i18n.Text("(levels only)")
	default:
		return Option(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Option) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Option) UnmarshalText(text []byte) error {
	*enum = ExtractOption(string(text))
	return nil
}

// ExtractOption extracts the value from a string.
func ExtractOption(str string) Option {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
