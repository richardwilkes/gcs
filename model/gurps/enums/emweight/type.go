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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// StringWithExample returns an example along with the normal String() content.
func (enum Type) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted values.
func (enum Type) Permitted() []Value {
	if enum.EnsureValid() == Original {
		return []Value{Addition, PercentageAdder}
	}
	return []Value{Addition, PercentageMultiplier, Multiplier}
}

// DetermineModifierWeightValueTypeFromString examines a string to determine what type it is, but restricts the result
// to those allowed for this Type.
func (enum Type) DetermineModifierWeightValueTypeFromString(s string) Value {
	mvt := LastValue.FromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == mvt {
			return mvt
		}
	}
	return permitted[0]
}

// ExtractFraction from the string.
func (enum Type) ExtractFraction(s string) fxp.Fraction {
	return enum.DetermineModifierWeightValueTypeFromString(s).ExtractFraction(s)
}

// Format returns a formatted version of the value.
func (enum Type) Format(s string, defUnits fxp.WeightUnit) string {
	t := enum.DetermineModifierWeightValueTypeFromString(s)
	result := t.Format(t.ExtractFraction(s))
	if t == Addition {
		result += " " + fxp.TrailingWeightUnitFromString(s, defUnits).String()
	}
	return result
}
