// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package stlimit

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	None Option = iota
	StrikingOnly
	LiftingOnly
	ThrowingOnly
)

// DefaultOption is the default value.
const DefaultOption Option = None

// FirstOption is the first valid value.
const FirstOption Option = None

// LastOption is the last valid value.
const LastOption Option = ThrowingOnly

// Options holds all possible values.
var Options = []Option{
	None,
	StrikingOnly,
	LiftingOnly,
	ThrowingOnly,
}

// Option holds a limitation for a Strength AttributeBonus.
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
	case None:
		return "none"
	case StrikingOnly:
		return "striking_only"
	case LiftingOnly:
		return "lifting_only"
	case ThrowingOnly:
		return "throwing_only"
	default:
		return DefaultOption.Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case None:
		return ``
	case StrikingOnly:
		return i18n.Text(`for striking only`)
	case LiftingOnly:
		return i18n.Text(`for lifting only`)
	case ThrowingOnly:
		return i18n.Text(`for throwing only`)
	default:
		return DefaultOption.String()
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
