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
	PercentageTraitModifierCostType TraitModifierCostType = iota
	PointsTraitModifierCostType
	MultiplierTraitModifierCostType
	LastTraitModifierCostType = MultiplierTraitModifierCostType
)

// AllTraitModifierCostType holds all possible values.
var AllTraitModifierCostType = []TraitModifierCostType{
	PercentageTraitModifierCostType,
	PointsTraitModifierCostType,
	MultiplierTraitModifierCostType,
}

// TraitModifierCostType describes how a TraitModifier's point cost is applied.
type TraitModifierCostType byte

// EnsureValid ensures this is of a known value.
func (enum TraitModifierCostType) EnsureValid() TraitModifierCostType {
	if enum <= LastTraitModifierCostType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum TraitModifierCostType) Key() string {
	switch enum {
	case PercentageTraitModifierCostType:
		return "percentage"
	case PointsTraitModifierCostType:
		return "points"
	case MultiplierTraitModifierCostType:
		return "multiplier"
	default:
		return TraitModifierCostType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum TraitModifierCostType) String() string {
	switch enum {
	case PercentageTraitModifierCostType:
		return i18n.Text("%")
	case PointsTraitModifierCostType:
		return i18n.Text("points")
	case MultiplierTraitModifierCostType:
		return i18n.Text("×")
	default:
		return TraitModifierCostType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum TraitModifierCostType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *TraitModifierCostType) UnmarshalText(text []byte) error {
	*enum = ExtractTraitModifierCostType(string(text))
	return nil
}

// ExtractTraitModifierCostType extracts the value from a string.
func ExtractTraitModifierCostType(str string) TraitModifierCostType {
	for _, enum := range AllTraitModifierCostType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
