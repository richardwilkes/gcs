// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package updatecheck

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	AtLaunch Option = iota
	Hourly
	Daily
	Never
)

// DefaultOption is the default value.
const DefaultOption Option = AtLaunch

// FirstOption is the first valid value.
const FirstOption Option = AtLaunch

// LastOption is the last valid value.
const LastOption Option = Never

// Options holds all possible values.
var Options = []Option{
	AtLaunch,
	Hourly,
	Daily,
	Never,
}

// Option holds how often to check for updates.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum >= FirstOption && enum <= LastOption {
		return enum
	}
	return DefaultOption
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
	switch enum {
	case AtLaunch:
		return "at_launch"
	case Hourly:
		return "hourly"
	case Daily:
		return "daily"
	case Never:
		return "never"
	default:
		return DefaultOption.Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case AtLaunch:
		return i18n.Text(`At launch`)
	case Hourly:
		return i18n.Text(`At launch and every hour`)
	case Daily:
		return i18n.Text(`At launch and every day`)
	case Never:
		return i18n.Text(`Never`)
	default:
		return DefaultOption.String()
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
	return DefaultOption
}

// ExtractKnownOption extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractOption, which quietly maps anything it doesn't recognize onto the first value, this permits a caller
// that is dispatching on the type to detect unknown types.
func ExtractKnownOption(str string) (value Option, known bool) {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultOption, false
}
