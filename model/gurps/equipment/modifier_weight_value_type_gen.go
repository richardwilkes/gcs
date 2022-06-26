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

package equipment

import (
	"strings"
)

// Possible values.
const (
	WeightAddition ModifierWeightValueType = iota
	WeightPercentageAdder
	WeightPercentageMultiplier
	WeightMultiplier
	LastModifierWeightValueType = WeightMultiplier
)

var (
	// AllModifierWeightValueType holds all possible values.
	AllModifierWeightValueType = []ModifierWeightValueType{
		WeightAddition,
		WeightPercentageAdder,
		WeightPercentageMultiplier,
		WeightMultiplier,
	}
	modifierWeightValueTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "+",
			string: "+",
		},
		{
			key:    "%",
			string: "%",
		},
		{
			key:    "x%",
			string: "x%",
		},
		{
			key:    "x",
			string: "x",
		},
	}
)

// ModifierWeightValueType describes how an EquipmentModifier's weight is applied.
type ModifierWeightValueType byte

// EnsureValid ensures this is of a known value.
func (enum ModifierWeightValueType) EnsureValid() ModifierWeightValueType {
	if enum <= LastModifierWeightValueType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ModifierWeightValueType) Key() string {
	return modifierWeightValueTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ModifierWeightValueType) String() string {
	return modifierWeightValueTypeData[enum.EnsureValid()].string
}

// ExtractModifierWeightValueType extracts the value from a string.
func ExtractModifierWeightValueType(str string) ModifierWeightValueType {
	for i, one := range modifierWeightValueTypeData {
		if strings.EqualFold(one.key, str) {
			return ModifierWeightValueType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ModifierWeightValueType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ModifierWeightValueType) UnmarshalText(text []byte) error {
	*enum = ExtractModifierWeightValueType(string(text))
	return nil
}
