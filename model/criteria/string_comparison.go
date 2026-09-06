// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria

import "strings"

// Describe returns a description of this StringComparison using a qualifier.
func (enum StringComparison) Describe(qualifier string) string {
	v := enum.EnsureValid()
	if v == AnyText {
		return v.String()
	}
	return v.String() + ` "` + qualifier + `"`
}

// DescribeWithPrefix returns a description of this StringComparison using a qualifier and prefix.
func (enum StringComparison) DescribeWithPrefix(prefix, notPrefix, qualifier string) string {
	v := enum.EnsureValid()
	var info string
	if prefix == notPrefix || !v.IsNotType() {
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
func (enum StringComparison) Matches(qualifier, data string) bool {
	switch enum {
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
func (enum StringComparison) IsNotType() bool {
	return enum == IsNotText || enum == DoesNotContainText || enum == DoesNotStartWithText || enum == DoesNotEndWithText
}

// PrefixedStringComparisonChoices returns the set of StringComparison choices as strings with a prefix.
func PrefixedStringComparisonChoices(prefix, notPrefix string) []string {
	choices := make([]string, len(StringComparisons))
	for i, choice := range StringComparisons {
		if prefix == notPrefix || !choice.IsNotType() {
			choices[i] = prefix + " " + choice.String()
		} else {
			choices[i] = notPrefix + " " + choice.AltString()
		}
	}
	return choices
}
