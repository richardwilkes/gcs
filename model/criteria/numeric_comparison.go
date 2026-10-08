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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Describe returns a description of this NumericComparison using a qualifier.
func (enum NumericComparison) Describe(qualifier fxp.Int) string {
	return enum.DescribeWith(qualifier.String())
}

// DescribeWith returns a description of this NumericComparison using an already-formatted qualifier. Use this for a
// qualifier that carries more than a bare number, such as a weight with its units.
func (enum NumericComparison) DescribeWith(qualifier string) string {
	v := enum.EnsureValid()
	if v == AnyNumber {
		return v.String()
	}
	return v.String() + " " + qualifier
}

// AltDescribe returns an alternate description of this NumericComparison using a qualifier.
func (enum NumericComparison) AltDescribe(qualifier fxp.Int) string {
	return enum.AltDescribeWith(qualifier.String())
}

// AltDescribeWith returns an alternate description of this NumericComparison using an already-formatted qualifier.
// Use this for a qualifier that carries more than a bare number, such as a weight with its units.
func (enum NumericComparison) AltDescribeWith(qualifier string) string {
	v := enum.EnsureValid()
	result := v.AltString()
	if v == AnyNumber {
		return result
	}
	if result != "" {
		result += " "
	}
	return result + qualifier
}

// Matches performs a comparison and returns true if the data matches.
func (enum NumericComparison) Matches(qualifier, data fxp.Int) bool {
	switch enum {
	case AnyNumber:
		return true
	case EqualsNumber:
		return data == qualifier
	case NotEqualsNumber:
		return data != qualifier
	case AtLeastNumber:
		return data >= qualifier
	case AtMostNumber:
		return data <= qualifier
	default:
		return AnyNumber.Matches(qualifier, data)
	}
}

// PluralClause returns the comparison as a clause that follows something plural, such as "that are at least" in "points
// that are at least".
func (enum NumericComparison) PluralClause() string {
	switch enum.EnsureValid() {
	case EqualsNumber:
		return i18n.Text("that are")
	case NotEqualsNumber:
		return i18n.Text("that are not")
	case AtLeastNumber:
		return i18n.Text("that are at least")
	case AtMostNumber:
		return i18n.Text("that are at most")
	default:
		return i18n.Text("that are anything")
	}
}

// PrefixedNumericComparisonChoices returns the set of NumericComparison choices as strings with a prefix.
func PrefixedNumericComparisonChoices(prefix string) []string {
	choices := make([]string, len(NumericComparisons))
	for i, choice := range NumericComparisons {
		choices[i] = prefix + " " + choice.String()
	}
	return choices
}
