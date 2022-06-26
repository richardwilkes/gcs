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
)

// StringWithExample returns an example along with the normal String() content.
func (enum ModifierCostType) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted ModifierCostValueType values.
func (enum ModifierCostType) Permitted() []ModifierCostValueType {
	if enum.EnsureValid() == BaseCost {
		return []ModifierCostValueType{CostFactor, Multiplier}
	}
	return []ModifierCostValueType{Addition, Percentage, Multiplier}
}

// DetermineModifierCostValueTypeFromString examines a string to determine what type it is, but restricts the result to
// those allowed for this ModifierCostType.
func (enum ModifierCostType) DetermineModifierCostValueTypeFromString(s string) ModifierCostValueType {
	cvt := DetermineModifierCostValueTypeFromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == cvt {
			return cvt
		}
	}
	return permitted[0]
}

// ExtractValue from the string.
func (enum ModifierCostType) ExtractValue(s string) fxp.Int {
	return enum.DetermineModifierCostValueTypeFromString(s).ExtractValue(s)
}

// Format returns a formatted version of the value.
func (enum ModifierCostType) Format(s string) string {
	cvt := enum.DetermineModifierCostValueTypeFromString(s)
	return cvt.Format(cvt.ExtractValue(s))
}
