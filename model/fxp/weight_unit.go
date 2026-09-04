// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"slices"
	"strings"
)

// weightUnitsBySuffixLen holds the weight units ordered by descending key length, so that suffix matching always
// prefers the most specific unit (e.g. "kg" before "g") regardless of the order in which the enum is declared.
var weightUnitsBySuffixLen = func() []WeightUnit {
	units := slices.Clone(WeightUnits)
	slices.SortStableFunc(units, func(a, b WeightUnit) int { return len(b.Key()) - len(a.Key()) })
	return units
}()

// TrailingWeightUnitFromString extracts a trailing WeightUnit from a string.
func TrailingWeightUnitFromString(s string, defUnits WeightUnit) WeightUnit {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, one := range weightUnitsBySuffixLen {
		if strings.HasSuffix(s, one.Key()) {
			return one
		}
	}
	return defUnits
}

// Format the weight for this WeightUnit, showing as many decimal places as the value has.
func (enum WeightUnit) Format(weight Weight) string {
	return enum.FormatWith(weight, NumberFormat{})
}

// FormatWith formats the weight for this WeightUnit, rendering the number according to the given NumberFormat. This is
// for display only: the rounding the NumberFormat may perform is lossy, so the result must never be parsed back into a
// stored value.
func (enum WeightUnit) FormatWith(weight Weight, format NumberFormat) string {
	return format.Format(enum.FromPounds(Int(weight))) + " " + enum.Key()
}

// FromPounds converts a weight in pounds to this WeightUnit.
func (enum WeightUnit) FromPounds(weight Int) Int {
	switch enum {
	case Pound, PoundAlt:
		return weight
	case Ounce:
		return weight.Mul(Sixteen)
	case Ton, TonAlt:
		return weight.Div(TwoThousand)
	case Kilogram:
		return weight.Div(Two)
	case Gram:
		return weight.Mul(FiveHundred)
	default:
		return Pound.FromPounds(weight)
	}
}

// ToPounds the weight for this WeightUnit.
func (enum WeightUnit) ToPounds(weight Int) Int {
	switch enum {
	case Pound, PoundAlt:
		return weight
	case Ounce:
		return weight.Div(Sixteen)
	case Ton, TonAlt:
		return weight.Mul(TwoThousand)
	case Kilogram:
		return weight.Mul(Two)
	case Gram:
		return weight.Div(FiveHundred)
	default:
		return Pound.ToPounds(weight)
	}
}
