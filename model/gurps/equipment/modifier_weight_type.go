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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
)

// StringWithExample returns an example along with the normal String() content.
func (enum ModifierWeightType) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted ModifierCostValueType values.
func (enum ModifierWeightType) Permitted() []ModifierWeightValueType {
	if enum.EnsureValid() == OriginalWeight {
		return []ModifierWeightValueType{WeightAddition, WeightPercentageAdder}
	}
	return []ModifierWeightValueType{WeightAddition, WeightPercentageMultiplier, WeightMultiplier}
}

// DetermineModifierWeightValueTypeFromString examines a string to determine what type it is, but restricts the result to
// those allowed for this ModifierWeightType.
func (enum ModifierWeightType) DetermineModifierWeightValueTypeFromString(s string) ModifierWeightValueType {
	mvt := DetermineModifierWeightValueTypeFromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == mvt {
			return mvt
		}
	}
	return permitted[0]
}

// ExtractFraction from the string.
func (enum ModifierWeightType) ExtractFraction(s string) fxp.Fraction {
	return enum.DetermineModifierWeightValueTypeFromString(s).ExtractFraction(s)
}

// Format returns a formatted version of the value.
func (enum ModifierWeightType) Format(s string, defUnits measure.WeightUnits) string {
	t := enum.DetermineModifierWeightValueTypeFromString(s)
	result := t.Format(t.ExtractFraction(s))
	if t == WeightAddition {
		result += " " + measure.TrailingWeightUnitsFromString(s, defUnits).String()
	}
	return result
}
