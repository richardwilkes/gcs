// Code generated from "enum.go.tmpl" - DO NOT EDIT.

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

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	AdditionEquipmentModifierCostValueType EquipmentModifierCostValueType = iota
	PercentageEquipmentModifierCostValueType
	MultiplierEquipmentModifierCostValueType
	CostFactorEquipmentModifierCostValueType
	LastEquipmentModifierCostValueType = CostFactorEquipmentModifierCostValueType
)

// AllEquipmentModifierCostValueType holds all possible values.
var AllEquipmentModifierCostValueType = []EquipmentModifierCostValueType{
	AdditionEquipmentModifierCostValueType,
	PercentageEquipmentModifierCostValueType,
	MultiplierEquipmentModifierCostValueType,
	CostFactorEquipmentModifierCostValueType,
}

// EquipmentModifierCostValueType describes how an Equipment Modifier's cost value is applied.
type EquipmentModifierCostValueType byte

// EnsureValid ensures this is of a known value.
func (enum EquipmentModifierCostValueType) EnsureValid() EquipmentModifierCostValueType {
	if enum <= LastEquipmentModifierCostValueType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum EquipmentModifierCostValueType) Key() string {
	switch enum {
	case AdditionEquipmentModifierCostValueType:
		return "+"
	case PercentageEquipmentModifierCostValueType:
		return "%"
	case MultiplierEquipmentModifierCostValueType:
		return "x"
	case CostFactorEquipmentModifierCostValueType:
		return "cf"
	default:
		return EquipmentModifierCostValueType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum EquipmentModifierCostValueType) String() string {
	switch enum {
	case AdditionEquipmentModifierCostValueType:
		return i18n.Text("+")
	case PercentageEquipmentModifierCostValueType:
		return i18n.Text("%")
	case MultiplierEquipmentModifierCostValueType:
		return i18n.Text("x")
	case CostFactorEquipmentModifierCostValueType:
		return i18n.Text("CF")
	default:
		return EquipmentModifierCostValueType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum EquipmentModifierCostValueType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *EquipmentModifierCostValueType) UnmarshalText(text []byte) error {
	*enum = ExtractEquipmentModifierCostValueType(string(text))
	return nil
}

// ExtractEquipmentModifierCostValueType extracts the value from a string.
func ExtractEquipmentModifierCostValueType(str string) EquipmentModifierCostValueType {
	for _, enum := range AllEquipmentModifierCostValueType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
