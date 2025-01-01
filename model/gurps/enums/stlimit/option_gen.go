// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	None Option = iota
	StrikingOnly
	LiftingOnly
	ThrowingOnly
)

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

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= ThrowingOnly {
		return enum
	}
	return 0
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
		return Option(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case None:
		return ""
	case StrikingOnly:
		return i18n.Text("for striking only")
	case LiftingOnly:
		return i18n.Text("for lifting only")
	case ThrowingOnly:
		return i18n.Text("for throwing only")
	default:
		return Option(0).String()
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
