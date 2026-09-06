// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package emcost

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Format returns a formatted version of the value.
func (enum Value) Format(value fxp.Int) string {
	switch enum {
	case Addition:
		return value.CommaWithSign()
	case Percentage:
		return value.CommaWithSign() + enum.String()
	case Multiplier:
		if value <= 0 {
			value = fxp.One
		}
		return enum.String() + value.Comma()
	case CostFactor:
		return value.CommaWithSign() + " " + enum.String()
	default:
		return Addition.Format(value)
	}
}

// ExtractValue extracts the numeric value from the string, interpreting it according to this Value. A non-positive
// multiplier is treated as 1.
func (enum Value) ExtractValue(s string) fxp.Int {
	return fxp.ExtractModifierValue(s, enum.EnsureValid() == Multiplier)
}

// ValueFromString examines a string to determine which Value it represents. A trailing "CF" indicates a cost factor,
// a trailing "%" a percentage, a leading or trailing "x" (or "×") a multiplier, and anything else is a plain addition.
func ValueFromString(s string) Value {
	switch {
	case strings.HasSuffix(strings.ToLower(strings.TrimSpace(s)), CostFactor.Key()):
		return CostFactor
	case fxp.HasPercentSuffix(s):
		return Percentage
	case fxp.HasMultiplierMarker(s):
		return Multiplier
	default:
		return Addition
	}
}
