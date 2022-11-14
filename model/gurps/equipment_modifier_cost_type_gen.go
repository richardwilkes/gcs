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

package gurps

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	OriginalEquipmentModifierCostType EquipmentModifierCostType = iota
	BaseEquipmentModifierCostType
	FinalBaseEquipmentModifierCostType
	FinalEquipmentModifierCostType
	LastEquipmentModifierCostType = FinalEquipmentModifierCostType
)

var (
	// AllEquipmentModifierCostType holds all possible values.
	AllEquipmentModifierCostType = []EquipmentModifierCostType{
		OriginalEquipmentModifierCostType,
		BaseEquipmentModifierCostType,
		FinalBaseEquipmentModifierCostType,
		FinalEquipmentModifierCostType,
	}
	equipmentModifierCostTypeData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "to_original_cost",
			string: i18n.Text("to original cost"),
			alt:    i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\""),
		},
		{
			key:    "to_base_cost",
			string: i18n.Text("to base cost"),
			alt:    i18n.Text("\"x2\", \"+2 CF\", \"-0.2 CF\""),
		},
		{
			key:    "to_final_base_cost",
			string: i18n.Text("to final base cost"),
			alt:    i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\""),
		},
		{
			key:    "to_final_cost",
			string: i18n.Text("to final cost"),
			alt:    i18n.Text("\"+5\", \"-5\", \"+10%\", \"-10%\", \"x3.2\""),
		},
	}
)

// EquipmentModifierCostType describes how an Equipment Modifier's cost is applied.
type EquipmentModifierCostType byte

// EnsureValid ensures this is of a known value.
func (enum EquipmentModifierCostType) EnsureValid() EquipmentModifierCostType {
	if enum <= LastEquipmentModifierCostType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum EquipmentModifierCostType) Key() string {
	return equipmentModifierCostTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum EquipmentModifierCostType) String() string {
	return equipmentModifierCostTypeData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum EquipmentModifierCostType) AltString() string {
	return equipmentModifierCostTypeData[enum.EnsureValid()].alt
}

// ExtractEquipmentModifierCostType extracts the value from a string.
func ExtractEquipmentModifierCostType(str string) EquipmentModifierCostType {
	for i, one := range equipmentModifierCostTypeData {
		if strings.EqualFold(one.key, str) {
			return EquipmentModifierCostType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum EquipmentModifierCostType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *EquipmentModifierCostType) UnmarshalText(text []byte) error {
	*enum = ExtractEquipmentModifierCostType(string(text))
	return nil
}
