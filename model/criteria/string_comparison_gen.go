// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	AnyText StringComparison = iota
	IsText
	IsNotText
	ContainsText
	DoesNotContainText
	StartsWithText
	DoesNotStartWithText
	EndsWithText
	DoesNotEndWithText
)

// DefaultStringComparison is the default value.
const DefaultStringComparison StringComparison = AnyText

// FirstStringComparison is the first valid value.
const FirstStringComparison StringComparison = AnyText

// LastStringComparison is the last valid value.
const LastStringComparison StringComparison = DoesNotEndWithText

// StringComparisons holds all possible values.
var StringComparisons = []StringComparison{
	AnyText,
	IsText,
	IsNotText,
	ContainsText,
	DoesNotContainText,
	StartsWithText,
	DoesNotStartWithText,
	EndsWithText,
	DoesNotEndWithText,
}

// StringComparison holds the type for a string comparison. The alternate string is the plural form, which only differs
// for the "not" cases.
type StringComparison byte

// EnsureValid ensures this is of a known value.
func (enum StringComparison) EnsureValid() StringComparison {
	if enum >= FirstStringComparison && enum <= LastStringComparison {
		return enum
	}
	return DefaultStringComparison
}

// Key returns the key used in serialization.
func (enum StringComparison) Key() string {
	switch enum {
	case AnyText:
		return ""
	case IsText:
		return "is"
	case IsNotText:
		return "is_not"
	case ContainsText:
		return "contains"
	case DoesNotContainText:
		return "does_not_contain"
	case StartsWithText:
		return "starts_with"
	case DoesNotStartWithText:
		return "does_not_start_with"
	case EndsWithText:
		return "ends_with"
	case DoesNotEndWithText:
		return "does_not_end_with"
	default:
		return DefaultStringComparison.Key()
	}
}

// String implements fmt.Stringer.
func (enum StringComparison) String() string {
	switch enum {
	case AnyText:
		return i18n.Text(`is anything`)
	case IsText:
		return i18n.Text(`is`)
	case IsNotText:
		return i18n.Text(`is not`)
	case ContainsText:
		return i18n.Text(`contains`)
	case DoesNotContainText:
		return i18n.Text(`does not contain`)
	case StartsWithText:
		return i18n.Text(`starts with`)
	case DoesNotStartWithText:
		return i18n.Text(`does not start with`)
	case EndsWithText:
		return i18n.Text(`ends with`)
	case DoesNotEndWithText:
		return i18n.Text(`does not end with`)
	default:
		return DefaultStringComparison.String()
	}
}

// AltString returns the alternate string.
func (enum StringComparison) AltString() string {
	switch enum {
	case AnyText:
		return i18n.Text(`is anything`)
	case IsText:
		return i18n.Text(`is`)
	case IsNotText:
		return i18n.Text(`are not`)
	case ContainsText:
		return i18n.Text(`contains`)
	case DoesNotContainText:
		return i18n.Text(`do not contain`)
	case StartsWithText:
		return i18n.Text(`starts with`)
	case DoesNotStartWithText:
		return i18n.Text(`do not start with`)
	case EndsWithText:
		return i18n.Text(`ends with`)
	case DoesNotEndWithText:
		return i18n.Text(`do not end with`)
	default:
		return DefaultStringComparison.AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum StringComparison) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *StringComparison) UnmarshalText(text []byte) error {
	*enum = ExtractStringComparison(string(text))
	return nil
}

// ExtractStringComparison extracts the value from a string.
func ExtractStringComparison(str string) StringComparison {
	for _, enum := range StringComparisons {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultStringComparison
}

// ExtractKnownStringComparison extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractStringComparison, which quietly maps anything it doesn't recognize onto the first value, this permits a
// caller that is dispatching on the type to detect unknown types.
func ExtractKnownStringComparison(str string) (value StringComparison, known bool) {
	for _, enum := range StringComparisons {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultStringComparison, false
}
