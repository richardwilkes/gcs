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
	OriginalWeight ModifierWeightType = iota
	BaseWeight
	FinalBaseWeight
	FinalWeight
	LastModifierWeightType = FinalWeight
)

var (
	// AllModifierWeightType holds all possible values.
	AllModifierWeightType = []ModifierWeightType{
		OriginalWeight,
		BaseWeight,
		FinalBaseWeight,
		FinalWeight,
	}
	modifierWeightTypeData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "to_original_weight",
			string: i18n.Text("to original weight"),
			alt:    i18n.Text("\"+5 lb\", \"-5 lb\", \"+10%\", \"-10%\""),
		},
		{
			key:    "to_base_weight",
			string: i18n.Text("to base weight"),
			alt:    i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\""),
		},
		{
			key:    "to_final_base_weight",
			string: i18n.Text("to final base weight"),
			alt:    i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\""),
		},
		{
			key:    "to_final_weight",
			string: i18n.Text("to final weight"),
			alt:    i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\""),
		},
	}
)

// ModifierWeightType describes how an EquipmentModifier's weight is applied.
type ModifierWeightType byte

// EnsureValid ensures this is of a known value.
func (enum ModifierWeightType) EnsureValid() ModifierWeightType {
	if enum <= LastModifierWeightType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum ModifierWeightType) Key() string {
	return modifierWeightTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum ModifierWeightType) String() string {
	return modifierWeightTypeData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum ModifierWeightType) AltString() string {
	return modifierWeightTypeData[enum.EnsureValid()].alt
}

// ExtractModifierWeightType extracts the value from a string.
func ExtractModifierWeightType(str string) ModifierWeightType {
	for i, one := range modifierWeightTypeData {
		if strings.EqualFold(one.key, str) {
			return ModifierWeightType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum ModifierWeightType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *ModifierWeightType) UnmarshalText(text []byte) error {
	*enum = ExtractModifierWeightType(string(text))
	return nil
}
