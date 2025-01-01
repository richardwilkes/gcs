// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package fxp

import (
	"strings"
)

// TrailingWeightUnitFromString extracts a trailing WeightUnit from a string.
func TrailingWeightUnitFromString(s string, defUnits WeightUnit) WeightUnit {
	s = strings.ToLower(strings.TrimSpace(s))
	for _, one := range WeightUnits {
		if strings.HasSuffix(s, one.Key()) {
			return one
		}
	}
	return defUnits
}

// Format the weight for this WeightUnit.
func (enum WeightUnit) Format(weight Weight) string {
	switch enum {
	case Pound, PoundAlt:
		return Int(weight).Comma() + " " + enum.Key()
	case Ounce:
		return Int(weight).Mul(Sixteen).Comma() + " " + enum.Key()
	case Ton, TonAlt:
		return Int(weight).Div(TwoThousand).Comma() + " " + enum.Key()
	case Kilogram:
		return Int(weight).Div(Two).Comma() + " " + enum.Key()
	case Gram:
		return Int(weight).Mul(FiveHundred).Comma() + " " + enum.Key()
	default:
		return Pound.Format(weight)
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
