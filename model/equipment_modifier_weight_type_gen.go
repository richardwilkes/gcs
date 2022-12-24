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
	OriginalEquipmentModifierWeightType EquipmentModifierWeightType = iota
	BaseEquipmentModifierWeightType
	FinalBaseEquipmentModifierWeightType
	FinalEquipmentModifierWeightType
	LastEquipmentModifierWeightType = FinalEquipmentModifierWeightType
)

// AllEquipmentModifierWeightType holds all possible values.
var AllEquipmentModifierWeightType = []EquipmentModifierWeightType{
	OriginalEquipmentModifierWeightType,
	BaseEquipmentModifierWeightType,
	FinalBaseEquipmentModifierWeightType,
	FinalEquipmentModifierWeightType,
}

// EquipmentModifierWeightType describes how an Equipment Modifier's weight is applied.
type EquipmentModifierWeightType byte

// EnsureValid ensures this is of a known value.
func (enum EquipmentModifierWeightType) EnsureValid() EquipmentModifierWeightType {
	if enum <= LastEquipmentModifierWeightType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum EquipmentModifierWeightType) Key() string {
	switch enum {
	case OriginalEquipmentModifierWeightType:
		return "to_original_weight"
	case BaseEquipmentModifierWeightType:
		return "to_base_weight"
	case FinalBaseEquipmentModifierWeightType:
		return "to_final_base_weight"
	case FinalEquipmentModifierWeightType:
		return "to_final_weight"
	default:
		return EquipmentModifierWeightType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum EquipmentModifierWeightType) String() string {
	switch enum {
	case OriginalEquipmentModifierWeightType:
		return i18n.Text("to original weight")
	case BaseEquipmentModifierWeightType:
		return i18n.Text("to base weight")
	case FinalBaseEquipmentModifierWeightType:
		return i18n.Text("to final base weight")
	case FinalEquipmentModifierWeightType:
		return i18n.Text("to final weight")
	default:
		return EquipmentModifierWeightType(0).String()
	}
}

// AltString returns the alternate string.
func (enum EquipmentModifierWeightType) AltString() string {
	switch enum {
	case OriginalEquipmentModifierWeightType:
		return i18n.Text("\"+5 lb\", \"-5 lb\", \"+10%\", \"-10%\"")
	case BaseEquipmentModifierWeightType:
		return i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\"")
	case FinalBaseEquipmentModifierWeightType:
		return i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\"")
	case FinalEquipmentModifierWeightType:
		return i18n.Text("\"+5 lb\", \"-5 lb\", \"x10%\", \"x3\", \"x2/3\"")
	default:
		return EquipmentModifierWeightType(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum EquipmentModifierWeightType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *EquipmentModifierWeightType) UnmarshalText(text []byte) error {
	*enum = ExtractEquipmentModifierWeightType(string(text))
	return nil
}

// ExtractEquipmentModifierWeightType extracts the value from a string.
func ExtractEquipmentModifierWeightType(str string) EquipmentModifierWeightType {
	for _, enum := range AllEquipmentModifierWeightType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
