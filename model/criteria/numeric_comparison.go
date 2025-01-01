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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
)

// Possible NumericComparison values.
const (
	AnyNumber       = NumericComparison("")
	EqualsNumber    = NumericComparison("is")
	NotEqualsNumber = NumericComparison("is_not")
	AtLeastNumber   = NumericComparison("at_least")
	AtMostNumber    = NumericComparison("at_most")
)

// AllNumericComparisons is the complete set of NumericComparison values.
var AllNumericComparisons = []NumericComparison{
	AnyNumber,
	EqualsNumber,
	NotEqualsNumber,
	AtLeastNumber,
	AtMostNumber,
}

// NumericComparison holds the type for a numeric comparison.
type NumericComparison string

// EnsureValid ensures this is of a known value.
func (n NumericComparison) EnsureValid() NumericComparison {
	for _, one := range AllNumericComparisons {
		if one == n {
			return n
		}
	}
	return AllNumericComparisons[0]
}

// AltString returns an alternate string for this.
func (n NumericComparison) AltString() string {
	switch n {
	case AnyNumber:
		return i18n.Text("anything")
	case EqualsNumber:
		return ""
	case NotEqualsNumber:
		return i18n.Text("not")
	case AtLeastNumber:
		return i18n.Text("at least")
	case AtMostNumber:
		return i18n.Text("at most")
	default:
		return AnyNumber.String()
	}
}

// String implements fmt.Stringer.
func (n NumericComparison) String() string {
	switch n {
	case AnyNumber:
		return i18n.Text("is anything")
	case EqualsNumber:
		return i18n.Text("is")
	case NotEqualsNumber:
		return i18n.Text("is not")
	case AtLeastNumber:
		return i18n.Text("is at least")
	case AtMostNumber:
		return i18n.Text("is at most")
	default:
		return AnyNumber.String()
	}
}

// Describe returns a description of this NumericCompareType using a qualifier.
func (n NumericComparison) Describe(qualifier fxp.Int) string {
	v := n.EnsureValid()
	if v == AnyNumber {
		return v.String()
	}
	return v.String() + " " + qualifier.String()
}

// AltDescribe returns an alternate description of this NumericCompareType using a qualifier.
func (n NumericComparison) AltDescribe(qualifier fxp.Int) string {
	v := n.EnsureValid()
	result := v.AltString()
	if v == AnyNumber {
		return result
	}
	if result != "" {
		result += " "
	}
	return result + qualifier.String()
}

// Matches performs a comparison and returns true if the data matches.
func (n NumericComparison) Matches(qualifier, data fxp.Int) bool {
	switch n {
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

// ExtractNumericComparisonIndex extracts the index from a string.
func ExtractNumericComparisonIndex(str string) int {
	for i, one := range AllNumericComparisons {
		if strings.EqualFold(string(one), str) {
			return i
		}
	}
	return 0
}

// PrefixedNumericComparisonChoices returns the set of NumericComparison choices as strings with a prefix.
func PrefixedNumericComparisonChoices(prefix string) []string {
	choices := make([]string, len(AllNumericComparisons))
	for i, choice := range AllNumericComparisons {
		choices[i] = prefix + " " + choice.String()
	}
	return choices
}
