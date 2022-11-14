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

package model

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NotShownDisplayOption DisplayOption = iota
	InlineDisplayOption
	TooltipDisplayOption
	InlineAndTooltipDisplayOption
	LastDisplayOption = InlineAndTooltipDisplayOption
)

var (
	// AllDisplayOption holds all possible values.
	AllDisplayOption = []DisplayOption{
		NotShownDisplayOption,
		InlineDisplayOption,
		TooltipDisplayOption,
		InlineAndTooltipDisplayOption,
	}
	displayOptionData = []struct {
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

// DisplayOption holds a display option.
type DisplayOption byte

// EnsureValid ensures this is of a known value.
func (enum DisplayOption) EnsureValid() DisplayOption {
	if enum <= LastDisplayOption {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum DisplayOption) Key() string {
	return displayOptionData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum DisplayOption) String() string {
	return displayOptionData[enum.EnsureValid()].string
}

// ExtractDisplayOption extracts the value from a string.
func ExtractDisplayOption(str string) DisplayOption {
	for i, one := range displayOptionData {
		if strings.EqualFold(one.key, str) {
			return DisplayOption(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum DisplayOption) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *DisplayOption) UnmarshalText(text []byte) error {
	*enum = ExtractDisplayOption(string(text))
	return nil
}
