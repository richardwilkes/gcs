// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emweight

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Format returns a formatted version of the value.
func (enum Value) Format(fraction fxp.Fraction) string {
	switch enum {
	case Addition:
		return fraction.StringWithSign()
	case PercentageAdder:
		return fraction.StringWithSign() + enum.String()
	case PercentageMultiplier:
		if fraction.Numerator < 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
		return Multiplier.String() + fraction.String() + PercentageAdder.String()
	case Multiplier:
		if fraction.Numerator < 0 {
			fraction.Numerator = fxp.One
			fraction.Denominator = fxp.One
		}
		return enum.String() + fraction.String()
	default:
		return Addition.Format(fraction)
	}
}

// ExtractFraction extracts the fraction from the string, interpreting it according to this Value. Any multiplier
// marker, percent sign or trailing unit is ignored. A negative multiplier is treated as 1 (or 100%).
func (enum Value) ExtractFraction(s string) fxp.Fraction {
	fraction := fxp.NewFraction(strings.TrimRightFunc(fxp.StripMultiplierLeaders(s), isNotDigit))
	revised := enum.EnsureValid()
	switch revised {
	case PercentageMultiplier:
		if fraction.Numerator < 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
	case Multiplier:
		if fraction.Numerator < 0 {
			fraction.Numerator = fxp.One
			fraction.Denominator = fxp.One
		}
	default:
	}
	return fraction
}

func isNotDigit(r rune) bool {
	return r < '0' || r > '9'
}

// ValueFromString examines a string to determine which Value it represents. A trailing "%" indicates a percentage,
// which is a multiplier when it also has a leading "x" (or "×") and an adder otherwise; a leading or trailing "x" (or
// "×") without a percent sign indicates a multiplier, and anything else is a plain addition.
func ValueFromString(s string) Value {
	switch {
	case fxp.HasPercentSuffix(s):
		if fxp.HasMultiplierPrefix(s) {
			return PercentageMultiplier
		}
		return PercentageAdder
	case fxp.HasMultiplierMarker(s):
		return Multiplier
	default:
		return Addition
	}
}
