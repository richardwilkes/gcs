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

// PluralClause returns the comparison as a clause that follows something plural, such as "that contain" in "notes that
// contain".
func (enum StringComparison) PluralClause() string {
	switch enum.EnsureValid() {
	case IsText:
		return i18n.Text("that are")
	case IsNotText:
		return i18n.Text("that are not")
	case ContainsText:
		return i18n.Text("that contain")
	case DoesNotContainText:
		return i18n.Text("that do not contain")
	case StartsWithText:
		return i18n.Text("that start with")
	case DoesNotStartWithText:
		return i18n.Text("that do not start with")
	case EndsWithText:
		return i18n.Text("that end with")
	case DoesNotEndWithText:
		return i18n.Text("that do not end with")
	default:
		return i18n.Text("that are anything")
	}
}

// ListClause returns the comparison as a clause that follows a list of values, which it is matched against value by
// value, such as "where at least one contains" in "tags where at least one contains". A "not" comparison holds when
// no value matches, so it reads "where none contains". "is anything" reads as a plural clause does.
func (enum StringComparison) ListClause() string {
	switch enum.EnsureValid() {
	case IsText:
		return i18n.Text("where at least one is")
	case IsNotText:
		return i18n.Text("where none is")
	case ContainsText:
		return i18n.Text("where at least one contains")
	case DoesNotContainText:
		return i18n.Text("where none contains")
	case StartsWithText:
		return i18n.Text("where at least one starts with")
	case DoesNotStartWithText:
		return i18n.Text("where none starts with")
	case EndsWithText:
		return i18n.Text("where at least one ends with")
	case DoesNotEndWithText:
		return i18n.Text("where none ends with")
	default:
		return i18n.Text("that are anything")
	}
}

// Positive returns the comparison a "not" comparison negates, or the comparison itself when it isn't one.
func (enum StringComparison) Positive() StringComparison {
	v := enum.EnsureValid()
	switch v {
	case IsNotText:
		return IsText
	case DoesNotContainText:
		return ContainsText
	case DoesNotStartWithText:
		return StartsWithText
	case DoesNotEndWithText:
		return EndsWithText
	default:
		return v
	}
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
