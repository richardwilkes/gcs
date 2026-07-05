// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package selector

import (
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Possible values.
const (
	WeaponDamageStrengthBasis Field = iota
	WeaponDamageStrengthMultiplier
	WeaponBaseDamageDice
	WeaponBaseDamageDicePerLevel
	WeaponDamagePerDieModifier
	WeaponArmorDivisor
	WeaponDamageType
	WeaponFragmentationDice
	WeaponFragmentationArmorDivisor
	WeaponFragmentationType
	TraitSelfControlRoll
	TraitSelfControlAdjustment
	TraitFrequency
)

// LastField is the last valid value.
const LastField Field = TraitFrequency

// Fields holds all possible values.
var Fields = []Field{
	WeaponDamageStrengthBasis,
	WeaponDamageStrengthMultiplier,
	WeaponBaseDamageDice,
	WeaponBaseDamageDicePerLevel,
	WeaponDamagePerDieModifier,
	WeaponArmorDivisor,
	WeaponDamageType,
	WeaponFragmentationDice,
	WeaponFragmentationArmorDivisor,
	WeaponFragmentationType,
	TraitSelfControlRoll,
	TraitSelfControlAdjustment,
	TraitFrequency,
}

// Field identifies a multi-state field that a SelectorOverride can replace.
type Field byte

// EnsureValid ensures this is of a known value.
func (enum Field) EnsureValid() Field {
	if enum <= TraitFrequency {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Field) Key() string {
	switch enum {
	case WeaponDamageStrengthBasis:
		return "weapon_damage_strength_basis"
	case WeaponDamageStrengthMultiplier:
		return "weapon_damage_strength_multiplier"
	case WeaponBaseDamageDice:
		return "weapon_base_damage_dice"
	case WeaponBaseDamageDicePerLevel:
		return "weapon_base_damage_dice_per_level"
	case WeaponDamagePerDieModifier:
		return "weapon_damage_per_die_modifier"
	case WeaponArmorDivisor:
		return "weapon_armor_divisor"
	case WeaponDamageType:
		return "weapon_damage_type"
	case WeaponFragmentationDice:
		return "weapon_fragmentation_dice"
	case WeaponFragmentationArmorDivisor:
		return "weapon_fragmentation_armor_divisor"
	case WeaponFragmentationType:
		return "weapon_fragmentation_type"
	case TraitSelfControlRoll:
		return "trait_self_control_roll"
	case TraitSelfControlAdjustment:
		return "trait_self_control_adjustment"
	case TraitFrequency:
		return "trait_frequency"
	default:
		return Field(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Field) String() string {
	switch enum {
	case WeaponDamageStrengthBasis:
		return i18n.Text(`weapon damage strength basis`)
	case WeaponDamageStrengthMultiplier:
		return i18n.Text(`weapon damage strength multiplier`)
	case WeaponBaseDamageDice:
		return i18n.Text(`weapon base damage dice`)
	case WeaponBaseDamageDicePerLevel:
		return i18n.Text(`weapon base damage dice per level`)
	case WeaponDamagePerDieModifier:
		return i18n.Text(`weapon damage per-die modifier`)
	case WeaponArmorDivisor:
		return i18n.Text(`weapon armor divisor`)
	case WeaponDamageType:
		return i18n.Text(`weapon damage type`)
	case WeaponFragmentationDice:
		return i18n.Text(`weapon fragmentation dice`)
	case WeaponFragmentationArmorDivisor:
		return i18n.Text(`weapon fragmentation armor divisor`)
	case WeaponFragmentationType:
		return i18n.Text(`weapon fragmentation damage type`)
	case TraitSelfControlRoll:
		return i18n.Text(`trait self-control roll`)
	case TraitSelfControlAdjustment:
		return i18n.Text(`trait self-control adjustment`)
	case TraitFrequency:
		return i18n.Text(`trait frequency of appearance`)
	default:
		return Field(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Field) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Field) UnmarshalText(text []byte) error {
	*enum = ExtractField(string(text))
	return nil
}

// ExtractField extracts the value from a string.
func ExtractField(str string) Field {
	for _, enum := range Fields {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
