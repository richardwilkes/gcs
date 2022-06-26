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

package feature

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	AttributeBonusType Type = iota
	ConditionalModifierType
	DRBonusType
	ReactionBonusType
	SkillBonusType
	SkillPointBonusType
	SpellBonusType
	SpellPointBonusType
	WeaponBonusType
	CostReductionType
	ContainedWeightReductionType
	LastType = ContainedWeightReductionType
)

var (
	// AllType holds all possible values.
	AllType = []Type{
		AttributeBonusType,
		ConditionalModifierType,
		DRBonusType,
		ReactionBonusType,
		SkillBonusType,
		SkillPointBonusType,
		SpellBonusType,
		SpellPointBonusType,
		WeaponBonusType,
		CostReductionType,
		ContainedWeightReductionType,
	}
	typeData = []struct {
		key    string
		string string
	}{
		{
			key:    "attribute_bonus",
			string: i18n.Text("Gives an attribute modifier of"),
		},
		{
			key:    "conditional_modifier",
			string: i18n.Text("Gives a conditional modifier of"),
		},
		{
			key:    "dr_bonus",
			string: i18n.Text("Gives a DR bonus of"),
		},
		{
			key:    "reaction_bonus",
			string: i18n.Text("Gives a reaction modifier of"),
		},
		{
			key:    "skill_bonus",
			string: i18n.Text("Gives a skill level modifier of"),
		},
		{
			key:    "skill_point_bonus",
			string: i18n.Text("Gives a skill point modifier of"),
		},
		{
			key:    "spell_bonus",
			string: i18n.Text("Gives a spell level modifier of"),
		},
		{
			key:    "spell_point_bonus",
			string: i18n.Text("Gives a spell point modifier of"),
		},
		{
			key:    "weapon_bonus",
			string: i18n.Text("Gives a weapon damage modifier of"),
		},
		{
			key:    "cost_reduction",
			string: i18n.Text("Reduces the attribute cost of"),
		},
		{
			key:    "contained_weight_reduction",
			string: i18n.Text("Reduces the contained weight by"),
		},
	}
)

// Type holds the type of a Feature.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= LastType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	return typeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	return typeData[enum.EnsureValid()].string
}

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for i, one := range typeData {
		if strings.EqualFold(one.key, str) {
			return Type(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Type) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Type) UnmarshalText(text []byte) error {
	*enum = ExtractType(string(text))
	return nil
}
