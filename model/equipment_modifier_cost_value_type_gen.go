// Code generated from "enum.go.tmpl" - DO NOT EDIT.

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

var (
	// AllEquipmentModifierCostValueType holds all possible values.
	AllEquipmentModifierCostValueType = []EquipmentModifierCostValueType{
		AdditionEquipmentModifierCostValueType,
		PercentageEquipmentModifierCostValueType,
		MultiplierEquipmentModifierCostValueType,
		CostFactorEquipmentModifierCostValueType,
	}
	equipmentModifierCostValueTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "+",
			string: i18n.Text("+"),
		},
		{
			key:    "%",
			string: i18n.Text("%"),
		},
		{
			key:    "x",
			string: i18n.Text("x"),
		},
		{
			key:    "cf",
			string: i18n.Text("CF"),
		},
	}
)

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
	return equipmentModifierCostValueTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum EquipmentModifierCostValueType) String() string {
	return equipmentModifierCostValueTypeData[enum.EnsureValid()].string
}

// ExtractEquipmentModifierCostValueType extracts the value from a string.
func ExtractEquipmentModifierCostValueType(str string) EquipmentModifierCostValueType {
	for i, one := range equipmentModifierCostValueTypeData {
		if strings.EqualFold(one.key, str) {
			return EquipmentModifierCostValueType(i)
		}
	}
	return 0
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
