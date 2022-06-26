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

package equipment

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Format returns a formatted version of the value.
func (enum ModifierWeightValueType) Format(fraction fxp.Fraction) string {
	switch enum {
	case WeightAddition:
		return fraction.StringWithSign()
	case WeightPercentageAdder:
		return fraction.StringWithSign() + enum.String()
	case WeightPercentageMultiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
		return WeightMultiplier.String() + fraction.String() + WeightPercentageAdder.String()
	case WeightMultiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.One
			fraction.Denominator = fxp.One
		}
		return enum.String() + fraction.String()
	default:
		return WeightAddition.Format(fraction)
	}
}

// ExtractFraction from the string.
func (enum ModifierWeightValueType) ExtractFraction(s string) fxp.Fraction {
	s = strings.TrimLeft(strings.TrimSpace(s), WeightMultiplier.Key())
	for len(s) > 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		s = s[:len(s)-1]
	}
	fraction := fxp.NewFraction(s)
	revised := enum.EnsureValid()
	switch revised {
	case WeightPercentageMultiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
	case WeightMultiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.One
			fraction.Denominator = fxp.One
		}
	default:
	}
	return fraction
}

// DetermineModifierWeightValueTypeFromString examines a string to determine what type it is.
func DetermineModifierWeightValueTypeFromString(s string) ModifierWeightValueType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, "%"):
		if strings.HasPrefix(s, "x") {
			return WeightPercentageMultiplier
		}
		return WeightPercentageAdder
	case strings.HasPrefix(s, "x") || strings.HasSuffix(s, "x"):
		return WeightMultiplier
	default:
		return WeightAddition
	}
}
