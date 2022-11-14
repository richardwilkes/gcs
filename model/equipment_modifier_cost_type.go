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

package model

import (
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// StringWithExample returns an example along with the normal String() content.
func (enum EquipmentModifierCostType) StringWithExample() string {
	return fmt.Sprintf("%s (e.g. %s)", enum.String(), enum.AltString())
}

// Permitted returns the permitted EquipmentModifierCostValueType values.
func (enum EquipmentModifierCostType) Permitted() []EquipmentModifierCostValueType {
	if enum.EnsureValid() == BaseEquipmentModifierCostType {
		return []EquipmentModifierCostValueType{CostFactorEquipmentModifierCostValueType, MultiplierEquipmentModifierCostValueType}
	}
	return []EquipmentModifierCostValueType{AdditionEquipmentModifierCostValueType, PercentageEquipmentModifierCostValueType, MultiplierEquipmentModifierCostValueType}
}

// FromString examines a string to determine what type it is, but restricts the result to those allowed for this
// EquipmentModifierCostValueType.
func (enum EquipmentModifierCostType) FromString(s string) EquipmentModifierCostValueType {
	cvt := AdditionEquipmentModifierCostValueType.FromString(s)
	permitted := enum.Permitted()
	for _, one := range permitted {
		if one == cvt {
			return cvt
		}
	}
	return permitted[0]
}

// ExtractValue from the string.
func (enum EquipmentModifierCostType) ExtractValue(s string) fxp.Int {
	return enum.FromString(s).ExtractValue(s)
}

// Format returns a formatted version of the value.
func (enum EquipmentModifierCostType) Format(s string) string {
	cvt := enum.FromString(s)
	return cvt.Format(cvt.ExtractValue(s))
}
