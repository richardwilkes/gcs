// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package display

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NotShown Option = iota
	Inline
	Tooltip
	InlineAndTooltip
	LastOption = InlineAndTooltip
)

var (
	// AllOption holds all possible values.
	AllOption = []Option{
		NotShown,
		Inline,
		Tooltip,
		InlineAndTooltip,
	}
	optionData = []struct {
		key    string
		string string
	}{
		{
			key:    "not_shown",
			string: i18n.Text("Not Shown"),
		},
		{
			key:    "inline",
			string: i18n.Text("Inline"),
		},
		{
			key:    "tooltip",
			string: i18n.Text("Tooltip"),
		},
		{
			key:    "inline_and_tooltip",
			string: i18n.Text("Inline & Tooltip"),
		},
	}
)

// Option holds a display option.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= LastOption {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
	return optionData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Option) String() string {
	return optionData[enum.EnsureValid()].string
}

// ExtractOption extracts the value from a string.
func ExtractOption(str string) Option {
	for i, one := range optionData {
		if strings.EqualFold(one.key, str) {
			return Option(i)
		}
	}
	return 0
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
