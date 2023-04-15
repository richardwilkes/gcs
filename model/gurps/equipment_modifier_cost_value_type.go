/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// Format returns a formatted version of the value.
func (enum EquipmentModifierCostValueType) Format(value fxp.Int) string {
	switch enum {
	case AdditionEquipmentModifierCostValueType:
		return value.StringWithSign()
	case PercentageEquipmentModifierCostValueType:
		return value.StringWithSign() + enum.String()
	case MultiplierEquipmentModifierCostValueType:
		if value <= 0 {
			value = fxp.One
		}
		return enum.String() + value.String()
	case CostFactorEquipmentModifierCostValueType:
		return value.StringWithSign() + " " + enum.String()
	default:
		return AdditionEquipmentModifierCostValueType.Format(value)
	}
}

// ExtractValue from the string.
func (enum EquipmentModifierCostValueType) ExtractValue(s string) fxp.Int {
	v, _ := fxp.Extract(strings.TrimLeft(strings.TrimSpace(s), MultiplierEquipmentModifierCostValueType.Key()))
	if enum.EnsureValid() == MultiplierEquipmentModifierCostValueType && v <= 0 {
		v = fxp.One
	}
	return v
}

// FromString examines a string to determine what type it is.
func (enum EquipmentModifierCostValueType) FromString(s string) EquipmentModifierCostValueType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch {
	case strings.HasSuffix(s, CostFactorEquipmentModifierCostValueType.Key()):
		return CostFactorEquipmentModifierCostValueType
	case strings.HasSuffix(s, PercentageEquipmentModifierCostValueType.Key()):
		return PercentageEquipmentModifierCostValueType
	case strings.HasPrefix(s, MultiplierEquipmentModifierCostValueType.Key()) || strings.HasSuffix(s, MultiplierEquipmentModifierCostValueType.Key()):
		return MultiplierEquipmentModifierCostValueType
	default:
		return AdditionEquipmentModifierCostValueType
	}
}
