// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package display

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	NotShown Option = iota
	Inline
	Tooltip
	InlineAndTooltip
)

// DefaultOption is the default value.
const DefaultOption Option = NotShown

// FirstOption is the first valid value.
const FirstOption Option = NotShown

// LastOption is the last valid value.
const LastOption Option = InlineAndTooltip

// Options holds all possible values.
var Options = []Option{
	NotShown,
	Inline,
	Tooltip,
	InlineAndTooltip,
}

// Option holds a display option.
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
	case NotShown:
		return "not_shown"
	case Inline:
		return "inline"
	case Tooltip:
		return "tooltip"
	case InlineAndTooltip:
		return "inline_and_tooltip"
	default:
		return DefaultOption.Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	switch enum {
	case NotShown:
		return i18n.Text(`Not Shown`)
	case Inline:
		return i18n.Text(`Inline`)
	case Tooltip:
		return i18n.Text(`Tooltip`)
	case InlineAndTooltip:
		return i18n.Text(`Inline & Tooltip`)
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
