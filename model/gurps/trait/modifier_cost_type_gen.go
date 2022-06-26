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

package trait

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	Percentage ModifierCostType = iota
	Points
	Multiplier
	LastModifierCostType = Multiplier
)

var (
	// AllModifierCostType holds all possible values.
	AllModifierCostType = []ModifierCostType{
		Percentage,
		Points,
		Multiplier,
	}
	modifierCostTypeData = []struct {
		key    string
		string string
	}{
		{
			key:    "percentage",
			string: i18n.Text("%"),
		},
		{
			key:    "points",
			string: i18n.Text("points"),
		},
		{
			key:    "multiplier",
			string: i18n.Text("×"),
		},
	}
)

// ModifierCostType describes how a TraitModifier's point cost is applied.
type ModifierCostType byte

// EnsureValid ensures this is of a known value.
func (enum ModifierCostType) EnsureValid() ModifierCostType {
	if enum <= LastModifierCostType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ModifierCostType) Key() string {
	return modifierCostTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ModifierCostType) String() string {
	return modifierCostTypeData[enum.EnsureValid()].string
}

// ExtractModifierCostType extracts the value from a string.
func ExtractModifierCostType(str string) ModifierCostType {
	for i, one := range modifierCostTypeData {
		if strings.EqualFold(one.key, str) {
			return ModifierCostType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ModifierCostType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ModifierCostType) UnmarshalText(text []byte) error {
	*enum = ExtractModifierCostType(string(text))
	return nil
}
