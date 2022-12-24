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
)

// Possible values.
const (
	AdditionEquipmentModifierWeightValueType EquipmentModifierWeightValueType = iota
	PercentageAdderEquipmentModifierWeightValueType
	PercentageMultiplierEquipmentModifierWeightValueType
	MultiplierEquipmentModifierWeightValueType
	LastEquipmentModifierWeightValueType = MultiplierEquipmentModifierWeightValueType
)

// AllEquipmentModifierWeightValueType holds all possible values.
var AllEquipmentModifierWeightValueType = []EquipmentModifierWeightValueType{
	AdditionEquipmentModifierWeightValueType,
	PercentageAdderEquipmentModifierWeightValueType,
	PercentageMultiplierEquipmentModifierWeightValueType,
	MultiplierEquipmentModifierWeightValueType,
}

// EquipmentModifierWeightValueType describes how an Equipment Modifier's weight value is applied.
type EquipmentModifierWeightValueType byte

// EnsureValid ensures this is of a known value.
func (enum EquipmentModifierWeightValueType) EnsureValid() EquipmentModifierWeightValueType {
	if enum <= LastEquipmentModifierWeightValueType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum EquipmentModifierWeightValueType) Key() string {
	switch enum {
	case AdditionEquipmentModifierWeightValueType:
		return "+"
	case PercentageAdderEquipmentModifierWeightValueType:
		return "%"
	case PercentageMultiplierEquipmentModifierWeightValueType:
		return "x%"
	case MultiplierEquipmentModifierWeightValueType:
		return "x"
	default:
		return EquipmentModifierWeightValueType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum EquipmentModifierWeightValueType) String() string {
	switch enum {
	case AdditionEquipmentModifierWeightValueType:
		return "+"
	case PercentageAdderEquipmentModifierWeightValueType:
		return "%"
	case PercentageMultiplierEquipmentModifierWeightValueType:
		return "x%"
	case MultiplierEquipmentModifierWeightValueType:
		return "x"
	default:
		return EquipmentModifierWeightValueType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum EquipmentModifierWeightValueType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *EquipmentModifierWeightValueType) UnmarshalText(text []byte) error {
	*enum = ExtractEquipmentModifierWeightValueType(string(text))
	return nil
}

// ExtractEquipmentModifierWeightValueType extracts the value from a string.
func ExtractEquipmentModifierWeightValueType(str string) EquipmentModifierWeightValueType {
	for _, enum := range AllEquipmentModifierWeightValueType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
