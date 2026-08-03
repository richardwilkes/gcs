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

const (
	// multiplicationSign is the Unicode multiplication sign, which users may type in place of the ASCII "x".
	multiplicationSign = "×"
	// multiplierLeaders holds every rune that may lead a multiplier value. ValueFromString lowercases before
	// classifying and also accepts the Unicode multiplication sign, so extraction must strip all of these forms.
	multiplierLeaders = "xX" + multiplicationSign
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

// ExtractFraction from the string.
func (enum Value) ExtractFraction(s string) fxp.Fraction {
	s = strings.TrimLeft(strings.TrimSpace(s), multiplierLeaders)
	for s != "" && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		s = s[:len(s)-1]
	}
	fraction := fxp.NewFraction(s)
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

// ValueFromString examines a string to determine what type it is.
func ValueFromString(s string) Value {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, PercentageAdder.Key()):
		if strings.HasPrefix(s, Multiplier.Key()) || strings.HasPrefix(s, multiplicationSign) {
			return PercentageMultiplier
		}
		return PercentageAdder
	case strings.HasPrefix(s, Multiplier.Key()) || strings.HasPrefix(s, multiplicationSign) ||
		strings.HasSuffix(s, Multiplier.Key()) || strings.HasSuffix(s, multiplicationSign):
		return Multiplier
	default:
		return Addition
	}
}
