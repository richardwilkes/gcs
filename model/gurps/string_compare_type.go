/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible StringCompareType values.
const (
	AnyString              = StringCompareType("")
	IsString               = StringCompareType("is")
	IsNotString            = StringCompareType("is_not")
	ContainsString         = StringCompareType("contains")
	DoesNotContainString   = StringCompareType("does_not_contain")
	StartsWithString       = StringCompareType("starts_with")
	DoesNotStartWithString = StringCompareType("does_not_start_with")
	EndsWithString         = StringCompareType("ends_with")
	DoesNotEndWithString   = StringCompareType("does_not_end_with")
)

// AllStringCompareTypes is the complete set of StringCompareType values.
var AllStringCompareTypes = []StringCompareType{
	AnyString,
	IsString,
	IsNotString,
	ContainsString,
	DoesNotContainString,
	StartsWithString,
	DoesNotStartWithString,
	EndsWithString,
	DoesNotEndWithString,
}

// StringCompareType holds the type for a string comparison.
type StringCompareType string

// EnsureValid ensures this is of a known value.
func (s StringCompareType) EnsureValid() StringCompareType {
	for _, one := range AllStringCompareTypes {
		if one == s {
			return s
		}
	}
	return AllStringCompareTypes[0]
}

// String implements fmt.Stringer.
func (s StringCompareType) String() string {
	switch s {
	case AnyString:
		return i18n.Text("is anything")
	case IsString:
		return i18n.Text("is")
	case IsNotString:
		return i18n.Text("is not")
	case ContainsString:
		return i18n.Text("contains")
	case DoesNotContainString:
		return i18n.Text("does not contain")
	case StartsWithString:
		return i18n.Text("starts with")
	case DoesNotStartWithString:
		return i18n.Text("does not start with")
	case EndsWithString:
		return i18n.Text("ends with")
	case DoesNotEndWithString:
		return i18n.Text("does not end with")
	default:
		return AnyString.String()
	}
}

// AltString provides a variant of String() for the not cases.
func (s StringCompareType) AltString() string {
	switch s {
	case IsNotString:
		return i18n.Text("are not")
	case DoesNotContainString:
		return i18n.Text("do not contain")
	case DoesNotStartWithString:
		return i18n.Text("do not start with")
	case DoesNotEndWithString:
		return i18n.Text("do not end with")
	default:
		return s.String()
	}
}

// Describe returns a description of this StringCompareType using a qualifier.
func (s StringCompareType) Describe(qualifier string) string {
	v := s.EnsureValid()
	if v == AnyString {
		return v.String()
	}
	return v.String() + ` "` + qualifier + `"`
}

// Matches performs a comparison and returns true if the data matches.
func (s StringCompareType) Matches(qualifier, data string) bool {
	switch s {
	case AnyString:
		return true
	case IsString:
		return strings.EqualFold(data, qualifier)
	case IsNotString:
		return !strings.EqualFold(data, qualifier)
	case ContainsString:
		return strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotContainString:
		return !strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case StartsWithString:
		return strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotStartWithString:
		return !strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case EndsWithString:
		return strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotEndWithString:
		return !strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	default:
		return AnyString.Matches(qualifier, data)
	}
}

// ExtractStringCompareTypeIndex extracts the index from a string.
func ExtractStringCompareTypeIndex(str string) int {
	for i, one := range AllStringCompareTypes {
		if strings.EqualFold(string(one), str) {
			return i
		}
	}
	return 0
}

// PrefixedStringCompareTypeChoices returns the set of StringCompareType choices as strings with a prefix.
func PrefixedStringCompareTypeChoices(prefix, notPrefix string) []string {
	choices := make([]string, len(AllStringCompareTypes))
	for i, choice := range AllStringCompareTypes {
		if prefix == notPrefix || choice == AnyString || choice == IsString || choice == ContainsString || choice == StartsWithString || choice == EndsWithString {
			choices[i] = prefix + " " + choice.String()
		} else {
			choices[i] = notPrefix + " " + choice.AltString()
		}
	}
	return choices
}
