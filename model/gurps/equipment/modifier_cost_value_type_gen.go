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

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Addition ModifierCostValueType = iota
	Percentage
	Multiplier
	CostFactor
	LastModifierCostValueType = CostFactor
)

var (
	// AllModifierCostValueType holds all possible values.
	AllModifierCostValueType = []ModifierCostValueType{
		Addition,
		Percentage,
		Multiplier,
		CostFactor,
	}
	modifierCostValueTypeData = []struct {
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

// ModifierCostValueType describes how an EquipmentModifier's cost is applied.
type ModifierCostValueType byte

// EnsureValid ensures this is of a known value.
func (enum ModifierCostValueType) EnsureValid() ModifierCostValueType {
	if enum <= LastModifierCostValueType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ModifierCostValueType) Key() string {
	return modifierCostValueTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ModifierCostValueType) String() string {
	return modifierCostValueTypeData[enum.EnsureValid()].string
}

// ExtractModifierCostValueType extracts the value from a string.
func ExtractModifierCostValueType(str string) ModifierCostValueType {
	for i, one := range modifierCostValueTypeData {
		if strings.EqualFold(one.key, str) {
			return ModifierCostValueType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ModifierCostValueType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ModifierCostValueType) UnmarshalText(text []byte) error {
	*enum = ExtractModifierCostValueType(string(text))
	return nil
}
