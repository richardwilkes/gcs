// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
		return Multiplier.String() + fraction.String() + PercentageAdder.String()
	case Multiplier:
		if fraction.Numerator <= 0 {
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
	s = strings.TrimLeft(strings.TrimSpace(s), Multiplier.Key())
	for s != "" && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		s = s[:len(s)-1]
	}
	fraction := fxp.NewFraction(s)
	revised := enum.EnsureValid()
	switch revised {
	case PercentageMultiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.Hundred
			fraction.Denominator = fxp.One
		}
	case Multiplier:
		if fraction.Numerator <= 0 {
			fraction.Numerator = fxp.One
			fraction.Denominator = fxp.One
		}
	default:
	}
	return fraction
}

// FromString examines a string to determine what type it is.
func (enum Value) FromString(s string) Value {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, "%"):
		if strings.HasPrefix(s, "x") {
			return PercentageMultiplier
		}
		return PercentageAdder
	case strings.HasPrefix(s, "x") || strings.HasSuffix(s, "x"):
		return Multiplier
	default:
		return Addition
	}
}
