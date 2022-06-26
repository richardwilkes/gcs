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

package criteria

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible StringCompareType values.
const (
	Any              = StringCompareType("")
	Is               = StringCompareType("is")
	IsNot            = StringCompareType("is_not")
	Contains         = StringCompareType("contains")
	DoesNotContain   = StringCompareType("does_not_contain")
	StartsWith       = StringCompareType("starts_with")
	DoesNotStartWith = StringCompareType("does_not_start_with")
	EndsWith         = StringCompareType("ends_with")
	DoesNotEndWith   = StringCompareType("does_not_end_with")
)

// AllStringCompareTypes is the complete set of StringCompareType values.
var AllStringCompareTypes = []StringCompareType{
	Any,
	Is,
	IsNot,
	Contains,
	DoesNotContain,
	StartsWith,
	DoesNotStartWith,
	EndsWith,
	DoesNotEndWith,
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
	case Any:
		return i18n.Text("is anything")
	case Is:
		return i18n.Text("is")
	case IsNot:
		return i18n.Text("is not")
	case Contains:
		return i18n.Text("contains")
	case DoesNotContain:
		return i18n.Text("does not contain")
	case StartsWith:
		return i18n.Text("starts with")
	case DoesNotStartWith:
		return i18n.Text("does not start with")
	case EndsWith:
		return i18n.Text("ends with")
	case DoesNotEndWith:
		return i18n.Text("does not end with")
	default:
		return Any.String()
	}
}

// Describe returns a description of this StringCompareType using a qualifier.
func (s StringCompareType) Describe(qualifier string) string {
	v := s.EnsureValid()
	if v == Any {
		return v.String()
	}
	return v.String() + ` "` + qualifier + `"`
}

// Matches performs a comparison and returns true if the data matches.
func (s StringCompareType) Matches(qualifier, data string) bool {
	switch s {
	case Any:
		return true
	case Is:
		return strings.EqualFold(data, qualifier)
	case IsNot:
		return !strings.EqualFold(data, qualifier)
	case Contains:
		return strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotContain:
		return !strings.Contains(strings.ToLower(data), strings.ToLower(qualifier))
	case StartsWith:
		return strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotStartWith:
		return !strings.HasPrefix(strings.ToLower(data), strings.ToLower(qualifier))
	case EndsWith:
		return strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	case DoesNotEndWith:
		return !strings.HasSuffix(strings.ToLower(data), strings.ToLower(qualifier))
	default:
		return Any.Matches(qualifier, data)
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
func PrefixedStringCompareTypeChoices(prefix string) []string {
	choices := make([]string, len(AllStringCompareTypes))
	for i, choice := range AllStringCompareTypes {
		choices[i] = prefix + " " + choice.String()
	}
	return choices
}
