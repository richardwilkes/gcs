// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	OpIs CompareOp = iota
	OpIsNot
	OpAtLeast
	OpAtMost
	OpMoreThan
	OpLessThan
)

// DefaultCompareOp is the default value.
const DefaultCompareOp CompareOp = OpIs

// FirstCompareOp is the first valid value.
const FirstCompareOp CompareOp = OpIs

// LastCompareOp is the last valid value.
const LastCompareOp CompareOp = OpLessThan

// CompareOps holds all possible values.
var CompareOps = []CompareOp{
	OpIs,
	OpIsNot,
	OpAtLeast,
	OpAtMost,
	OpMoreThan,
	OpLessThan,
}

// CompareOp holds the operator a compare node applies to its operands.
type CompareOp byte

// EnsureValid returns this value if it is valid and DefaultCompareOp otherwise.
func (enum CompareOp) EnsureValid() CompareOp {
	if enum >= FirstCompareOp && enum <= LastCompareOp {
		return enum
	}
	return DefaultCompareOp
}

// Key returns the key used in serialization.
func (enum CompareOp) Key() string {
	switch enum {
	case OpIs:
		return "is"
	case OpIsNot:
		return "is_not"
	case OpAtLeast:
		return "at_least"
	case OpAtMost:
		return "at_most"
	case OpMoreThan:
		return "more_than"
	case OpLessThan:
		return "less_than"
	default:
		return DefaultCompareOp.Key()
	}
}

// String implements fmt.Stringer.
func (enum CompareOp) String() string {
	switch enum {
	case OpIs:
		return i18n.Text(`is`)
	case OpIsNot:
		return i18n.Text(`is not`)
	case OpAtLeast:
		return i18n.Text(`is at least`)
	case OpAtMost:
		return i18n.Text(`is at most`)
	case OpMoreThan:
		return i18n.Text(`is more than`)
	case OpLessThan:
		return i18n.Text(`is less than`)
	default:
		return DefaultCompareOp.String()
	}
}

// MarshalText implements encoding.TextMarshaler.
func (enum CompareOp) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (enum *CompareOp) UnmarshalText(text []byte) error {
	*enum = ExtractCompareOp(string(text))
	return nil
}

// ExtractCompareOp returns the value whose key matches str, ignoring case, or DefaultCompareOp if none does.
func ExtractCompareOp(str string) CompareOp {
	for _, enum := range CompareOps {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return DefaultCompareOp
}

// ExtractKnownCompareOp is like ExtractCompareOp, but also reports whether str was recognized, so a caller can tell an
// unknown key from the default.
func ExtractKnownCompareOp(str string) (value CompareOp, known bool) {
	for _, enum := range CompareOps {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
	}
	return DefaultCompareOp, false
}
