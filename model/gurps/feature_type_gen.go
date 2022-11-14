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
	AttributeBonusFeatureType FeatureType = iota
	ConditionalModifierFeatureType
	DRBonusFeatureType
	ReactionBonusFeatureType
	SkillBonusFeatureType
	SkillPointBonusFeatureType
	SpellBonusFeatureType
	SpellPointBonusFeatureType
	WeaponBonusFeatureType
	WeaponDRDivisorBonusFeatureType
	CostReductionFeatureType
	ContainedWeightReductionFeatureType
	LastFeatureType = ContainedWeightReductionFeatureType
)

var (
	// AllFeatureType holds all possible values.
	AllFeatureType = []FeatureType{
		AttributeBonusFeatureType,
		ConditionalModifierFeatureType,
		DRBonusFeatureType,
		ReactionBonusFeatureType,
		SkillBonusFeatureType,
		SkillPointBonusFeatureType,
		SpellBonusFeatureType,
		SpellPointBonusFeatureType,
		WeaponBonusFeatureType,
		WeaponDRDivisorBonusFeatureType,
		CostReductionFeatureType,
		ContainedWeightReductionFeatureType,
	}
	featureTypeData = []struct {
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
			key:    "weapon_dr_divisor_bonus",
			string: i18n.Text("Gives a weapon DR divisor modifier of"),
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

// FeatureType holds the type of a Feature.
type FeatureType byte

// EnsureValid ensures this is of a known value.
func (enum FeatureType) EnsureValid() FeatureType {
	if enum <= LastFeatureType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum FeatureType) Key() string {
	return featureTypeData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum FeatureType) String() string {
	return featureTypeData[enum.EnsureValid()].string
}

// ExtractFeatureType extracts the value from a string.
func ExtractFeatureType(str string) FeatureType {
	for i, one := range featureTypeData {
		if strings.EqualFold(one.key, str) {
			return FeatureType(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum FeatureType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *FeatureType) UnmarshalText(text []byte) error {
	*enum = ExtractFeatureType(string(text))
	return nil
}
