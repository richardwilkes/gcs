// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	Total Option = iota
	BaseOnly
	LevelsOnly
)

// DefaultOption is the default value.
const DefaultOption Option = Total

// FirstOption is the first valid value.
const FirstOption Option = Total

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

// EnsureValid returns this value if it is valid and DefaultOption otherwise.
func (enum Option) EnsureValid() Option {
	if enum >= FirstOption && enum <= LastOption {
		return enum
	}
	return DefaultOption
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
		return DefaultOption.Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case Total:
		return i18n.Text(`to cost`)
	case BaseOnly:
		return i18n.Text(`to base cost only`)
	case LevelsOnly:
		return i18n.Text(`to leveled cost only`)
	default:
		return DefaultOption.String()
	}
}

// AltString returns the alternate string.
func (enum Option) AltString() string {
	switch enum {
	case Total:
		return ``
	case BaseOnly:
		return i18n.Text(`(base only)`)
	case LevelsOnly:
		return i18n.Text(`(levels only)`)
	default:
		return DefaultOption.AltString()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum Option) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *Option) UnmarshalText(text []byte) error {
	*enum = ExtractOption(string(text))
	return nil
}

// ExtractOption returns the value whose key matches str, ignoring case, or DefaultOption if none does.
func ExtractOption(str string) Option {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultOption
}

// ExtractKnownOption is like ExtractOption, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownOption(str string) (value Option, known bool) {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultOption, false
}
