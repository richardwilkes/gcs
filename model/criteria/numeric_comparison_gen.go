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
	AnyNumber NumericComparison = iota
	EqualsNumber
	NotEqualsNumber
	AtLeastNumber
	AtMostNumber
)

// DefaultNumericComparison is the default value.
const DefaultNumericComparison NumericComparison = AnyNumber

// FirstNumericComparison is the first valid value.
const FirstNumericComparison NumericComparison = AnyNumber

// LastNumericComparison is the last valid value.
const LastNumericComparison NumericComparison = AtMostNumber

// NumericComparisons holds all possible values.
var NumericComparisons = []NumericComparison{
	AnyNumber,
	EqualsNumber,
	NotEqualsNumber,
	AtLeastNumber,
	AtMostNumber,
}

// NumericComparison holds the type for a numeric comparison.
type NumericComparison byte

// EnsureValid returns this value if it is valid and DefaultNumericComparison otherwise.
func (enum NumericComparison) EnsureValid() NumericComparison {
	if enum >= FirstNumericComparison && enum <= LastNumericComparison {
		return enum
	}
	return DefaultNumericComparison
}

// Key returns the key used in serialization.
func (enum NumericComparison) Key() string {
	switch enum {
	case AnyNumber:
		return ""
	case EqualsNumber:
		return "is"
	case NotEqualsNumber:
		return "is_not"
	case AtLeastNumber:
		return "at_least"
	case AtMostNumber:
		return "at_most"
	default:
		return DefaultNumericComparison.Key()
	}
}

// String implements fmt.Stringer.
func (enum NumericComparison) String() string {
	switch enum {
	case AnyNumber:
		return i18n.Text(`is anything`)
	case EqualsNumber:
		return i18n.Text(`is`)
	case NotEqualsNumber:
		return i18n.Text(`is not`)
	case AtLeastNumber:
		return i18n.Text(`is at least`)
	case AtMostNumber:
		return i18n.Text(`is at most`)
	default:
		return DefaultNumericComparison.String()
	}
}

// AltString returns the alternate string.
func (enum NumericComparison) AltString() string {
	switch enum {
	case AnyNumber:
		return i18n.Text(`anything`)
	case EqualsNumber:
		return ``
	case NotEqualsNumber:
		return i18n.Text(`not`)
	case AtLeastNumber:
		return i18n.Text(`at least`)
	case AtMostNumber:
		return i18n.Text(`at most`)
	default:
		return DefaultNumericComparison.AltString()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum NumericComparison) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *NumericComparison) UnmarshalText(text []byte) error {
	*enum = ExtractNumericComparison(string(text))
	return nil
}

// ExtractNumericComparison returns the value whose key matches str, ignoring case, or DefaultNumericComparison if none
// does.
func ExtractNumericComparison(str string) NumericComparison {
	for _, enum := range NumericComparisons {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultNumericComparison
}

// ExtractKnownNumericComparison is like ExtractNumericComparison, but also reports whether str was recognized, so a
// caller can tell an unknown key from the default.
func ExtractKnownNumericComparison(str string) (value NumericComparison, known bool) {
	for _, enum := range NumericComparisons {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultNumericComparison, false
}
