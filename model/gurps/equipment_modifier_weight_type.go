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

package gurps

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/measure"
)

// StringWithExample returns an example along with the normal String() content.
func (enum EquipmentModifierWeightType) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted EquipmentModifierCostValueType values.
func (enum EquipmentModifierWeightType) Permitted() []EquipmentModifierWeightValueType {
	if enum.EnsureValid() == OriginalEquipmentModifierWeightType {
		return []EquipmentModifierWeightValueType{AdditionEquipmentModifierWeightValueType, PercentageAdderEquipmentModifierWeightValueType}
	}
	return []EquipmentModifierWeightValueType{AdditionEquipmentModifierWeightValueType, PercentageMultiplierEquipmentModifierWeightValueType, MultiplierEquipmentModifierWeightValueType}
}

// DetermineModifierWeightValueTypeFromString examines a string to determine what type it is, but restricts the result to
// those allowed for this EquipmentModifierWeightType.
func (enum EquipmentModifierWeightType) DetermineModifierWeightValueTypeFromString(s string) EquipmentModifierWeightValueType {
	mvt := LastEquipmentModifierWeightValueType.FromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == mvt {
			return mvt
		}
	}
	return permitted[0]
}

// ExtractFraction from the string.
func (enum EquipmentModifierWeightType) ExtractFraction(s string) fxp.Fraction {
	return enum.DetermineModifierWeightValueTypeFromString(s).ExtractFraction(s)
}

// Format returns a formatted version of the value.
func (enum EquipmentModifierWeightType) Format(s string, defUnits measure.WeightUnits) string {
	t := enum.DetermineModifierWeightValueTypeFromString(s)
	result := t.Format(t.ExtractFraction(s))
	if t == AdditionEquipmentModifierWeightValueType {
		result += " " + measure.TrailingWeightUnitsFromString(s, defUnits).String()
	}
	return result
}
