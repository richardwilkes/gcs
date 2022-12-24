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

// AllDisplayOption holds all possible values.
var AllDisplayOption = []DisplayOption{
	NotShownDisplayOption,
	InlineDisplayOption,
	TooltipDisplayOption,
	InlineAndTooltipDisplayOption,
}

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
	switch enum {
	case NotShownDisplayOption:
		return "not_shown"
	case InlineDisplayOption:
		return "inline"
	case TooltipDisplayOption:
		return "tooltip"
	case InlineAndTooltipDisplayOption:
		return "inline_and_tooltip"
	default:
		return DisplayOption(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum DisplayOption) String() string {
	switch enum {
	case NotShownDisplayOption:
		return i18n.Text("Not Shown")
	case InlineDisplayOption:
		return i18n.Text("Inline")
	case TooltipDisplayOption:
		return i18n.Text("Tooltip")
	case InlineAndTooltipDisplayOption:
		return i18n.Text("Inline & Tooltip")
	default:
		return DisplayOption(0).String()
	}
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

// ExtractDisplayOption extracts the value from a string.
func ExtractDisplayOption(str string) DisplayOption {
	for _, enum := range AllDisplayOption {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
