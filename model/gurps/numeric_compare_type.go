// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
)

// Possible NumericCompareType values.
const (
	AnyNumber       = NumericCompareType("")
	EqualsNumber    = NumericCompareType("is")
	NotEqualsNumber = NumericCompareType("is_not")
	AtLeastNumber   = NumericCompareType("at_least")
	AtMostNumber    = NumericCompareType("at_most")
)

// AllNumericCompareTypes is the complete set of NumericCompareType values.
var AllNumericCompareTypes = []NumericCompareType{
	AnyNumber,
	EqualsNumber,
	NotEqualsNumber,
	AtLeastNumber,
	AtMostNumber,
}

// NumericCompareType holds the type for a numeric comparison.
type NumericCompareType string

// EnsureValid ensures this is of a known value.
func (n NumericCompareType) EnsureValid() NumericCompareType {
	for _, one := range AllNumericCompareTypes {
		if one == n {
			return n
		}
	}
	return AllNumericCompareTypes[0]
}

// AltString returns an alternate string for this.
func (n NumericCompareType) AltString() string {
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
func (n NumericCompareType) String() string {
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
func (n NumericCompareType) Describe(qualifier fxp.Int) string {
	v := n.EnsureValid()
	if v == AnyNumber {
		return v.String()
	}
	return v.String() + " " + qualifier.String()
}

// AltDescribe returns an alternate description of this NumericCompareType using a qualifier.
func (n NumericCompareType) AltDescribe(qualifier fxp.Int) string {
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
func (n NumericCompareType) Matches(qualifier, data fxp.Int) bool {
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

// ExtractNumericCompareTypeIndex extracts the index from a string.
func ExtractNumericCompareTypeIndex(str string) int {
	for i, one := range AllNumericCompareTypes {
		if strings.EqualFold(string(one), str) {
			return i
		}
	}
	return 0
}

// PrefixedNumericCompareTypeChoices returns the set of NumericCompareType choices as strings with a prefix.
func PrefixedNumericCompareTypeChoices(prefix string) []string {
	choices := make([]string, len(AllNumericCompareTypes))
	for i, choice := range AllNumericCompareTypes {
		choices[i] = prefix + " " + choice.String()
	}
	return choices
}
