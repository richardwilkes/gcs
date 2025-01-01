// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible StringComparison values.
const (
	AnyText              = StringComparison("")
	IsText               = StringComparison("is")
	IsNotText            = StringComparison("is_not")
	ContainsText         = StringComparison("contains")
	DoesNotContainText   = StringComparison("does_not_contain")
	StartsWithText       = StringComparison("starts_with")
	DoesNotStartWithText = StringComparison("does_not_start_with")
	EndsWithText         = StringComparison("ends_with")
	DoesNotEndWithText   = StringComparison("does_not_end_with")
)

// AllStringComparisons is the complete set of StringComparison values.
var AllStringComparisons = []StringComparison{
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

// StringComparison holds the type for a string comparison.
type StringComparison string

// EnsureValid ensures this is of a known value.
func (s StringComparison) EnsureValid() StringComparison {
	for _, one := range AllStringComparisons {
		if one == s {
			return s
		}
	}
	return AllStringComparisons[0]
}

// String implements fmt.Stringer.
func (s StringComparison) String() string {
	switch s {
	case AnyText:
		return i18n.Text("is anything")
	case IsText:
		return i18n.Text("is")
	case IsNotText:
		return i18n.Text("is not")
	case ContainsText:
		return i18n.Text("contains")
	case DoesNotContainText:
		return i18n.Text("does not contain")
	case StartsWithText:
		return i18n.Text("starts with")
	case DoesNotStartWithText:
		return i18n.Text("does not start with")
	case EndsWithText:
		return i18n.Text("ends with")
	case DoesNotEndWithText:
		return i18n.Text("does not end with")
	default:
		return AnyText.String()
	}
}

// AltString provides a variant of String() for the not cases.
func (s StringComparison) AltString() string {
	switch s {
	case IsNotText:
		return i18n.Text("are not")
	case DoesNotContainText:
		return i18n.Text("do not contain")
	case DoesNotStartWithText:
		return i18n.Text("do not start with")
	case DoesNotEndWithText:
		return i18n.Text("do not end with")
	default:
		return s.String()
	}
}

// Describe returns a description of this StringCompareType using a qualifier.
func (s StringComparison) Describe(qualifier string) string {
	v := s.EnsureValid()
	if v == AnyText {
		return v.String()
	}
	return v.String() + ` "` + qualifier + `"`
}

// DescribeWithPrefix returns a description of this StringCompareType using a qualifier and prefix.
func (s StringComparison) DescribeWithPrefix(prefix, notPrefix, qualifier string) string {
	v := s.EnsureValid()
	var info string
	if prefix == notPrefix || !s.IsNotType() {
		info = prefix + " " + v.String()
	} else {
		info = notPrefix + " " + v.AltString()
	}
	if v == AnyText {
		return info
	}
	return info + ` "` + qualifier + `"`
}

// Matches performs a comparison and returns true if the data matches.
func (s StringComparison) Matches(qualifier, data string) bool {
	switch s {
	case AnyText:
		return true
	case IsText:
		return strings.EqualFold(data, qualifier)
	case IsNotText:
		return !strings.EqualFold(data, qualifier)
	case ContainsText:
		return strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotContainText:
		return !strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case StartsWithText:
		return strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotStartWithText:
		return !strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case EndsWithText:
		return strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotEndWithText:
		return !strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	default:
		return AnyText.Matches(qualifier, data)
	}
}

// IsNotType returns true if this is a "not" type.
func (s StringComparison) IsNotType() bool {
	return s == IsNotText || s == DoesNotContainText || s == DoesNotStartWithText || s == DoesNotEndWithText
}

// ExtractStringComparisonIndex extracts the index from a string.
func ExtractStringComparisonIndex(str string) int {
	for i, one := range AllStringComparisons {
		if strings.EqualFold(string(one), str) {
			return i
		}
	}
	return 0
}

// PrefixedStringComparisonChoices returns the set of StringCompareType choices as strings with a prefix.
func PrefixedStringComparisonChoices(prefix, notPrefix string) []string {
	choices := make([]string, len(AllStringComparisons))
	for i, choice := range AllStringComparisons {
		if prefix == notPrefix || !choice.IsNotType() {
			choices[i] = prefix + " " + choice.String()
		} else {
			choices[i] = notPrefix + " " + choice.AltString()
		}
	}
	return choices
}
