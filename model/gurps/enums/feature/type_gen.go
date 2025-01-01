// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package feature

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	AttributeBonus Type = iota
	ConditionalModifier
	DRBonus
	ReactionBonus
	SkillBonus
	SkillPointBonus
	SpellBonus
	SpellPointBonus
	WeaponBonus
	WeaponAccBonus
	WeaponScopeAccBonus
	WeaponDRDivisorBonus
	WeaponEffectiveSTBonus
	WeaponMinSTBonus
	WeaponMinReachBonus
	WeaponMaxReachBonus
	WeaponHalfDamageRangeBonus
	WeaponMinRangeBonus
	WeaponMaxRangeBonus
	WeaponRecoilBonus
	WeaponBulkBonus
	WeaponParryBonus
	WeaponBlockBonus
	WeaponRofMode1ShotsBonus
	WeaponRofMode1SecondaryBonus
	WeaponRofMode2ShotsBonus
	WeaponRofMode2SecondaryBonus
	WeaponNonChamberShotsBonus
	WeaponChamberShotsBonus
	WeaponShotDurationBonus
	WeaponReloadTimeBonus
	WeaponSwitch
	CostReduction
	ContainedWeightReduction
)

// LastType is the last valid value.
const LastType Type = ContainedWeightReduction

// Types holds all possible values.
var Types = []Type{
	AttributeBonus,
	ConditionalModifier,
	DRBonus,
	ReactionBonus,
	SkillBonus,
	SkillPointBonus,
	SpellBonus,
	SpellPointBonus,
	WeaponBonus,
	WeaponAccBonus,
	WeaponScopeAccBonus,
	WeaponDRDivisorBonus,
	WeaponEffectiveSTBonus,
	WeaponMinSTBonus,
	WeaponMinReachBonus,
	WeaponMaxReachBonus,
	WeaponHalfDamageRangeBonus,
	WeaponMinRangeBonus,
	WeaponMaxRangeBonus,
	WeaponRecoilBonus,
	WeaponBulkBonus,
	WeaponParryBonus,
	WeaponBlockBonus,
	WeaponRofMode1ShotsBonus,
	WeaponRofMode1SecondaryBonus,
	WeaponRofMode2ShotsBonus,
	WeaponRofMode2SecondaryBonus,
	WeaponNonChamberShotsBonus,
	WeaponChamberShotsBonus,
	WeaponShotDurationBonus,
	WeaponReloadTimeBonus,
	WeaponSwitch,
	CostReduction,
	ContainedWeightReduction,
}

// Type holds the type of a Feature.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= ContainedWeightReduction {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case AttributeBonus:
		return "attribute_bonus"
	case ConditionalModifier:
		return "conditional_modifier"
	case DRBonus:
		return "dr_bonus"
	case ReactionBonus:
		return "reaction_bonus"
	case SkillBonus:
		return "skill_bonus"
	case SkillPointBonus:
		return "skill_point_bonus"
	case SpellBonus:
		return "spell_bonus"
	case SpellPointBonus:
		return "spell_point_bonus"
	case WeaponBonus:
		return "weapon_bonus"
	case WeaponAccBonus:
		return "weapon_acc_bonus"
	case WeaponScopeAccBonus:
		return "weapon_scope_acc_bonus"
	case WeaponDRDivisorBonus:
		return "weapon_dr_divisor_bonus"
	case WeaponEffectiveSTBonus:
		return "weapon_effective_st_bonus"
	case WeaponMinSTBonus:
		return "weapon_min_st_bonus"
	case WeaponMinReachBonus:
		return "weapon_min_reach_bonus"
	case WeaponMaxReachBonus:
		return "weapon_max_reach_bonus"
	case WeaponHalfDamageRangeBonus:
		return "weapon_half_damage_range_bonus"
	case WeaponMinRangeBonus:
		return "weapon_min_range_bonus"
	case WeaponMaxRangeBonus:
		return "weapon_max_range_bonus"
	case WeaponRecoilBonus:
		return "weapon_recoil_bonus"
	case WeaponBulkBonus:
		return "weapon_bulk_bonus"
	case WeaponParryBonus:
		return "weapon_parry_bonus"
	case WeaponBlockBonus:
		return "weapon_block_bonus"
	case WeaponRofMode1ShotsBonus:
		return "weapon_rof_mode_1_shots_bonus"
	case WeaponRofMode1SecondaryBonus:
		return "weapon_rof_mode_1_secondary_bonus"
	case WeaponRofMode2ShotsBonus:
		return "weapon_rof_mode_2_shots_bonus"
	case WeaponRofMode2SecondaryBonus:
		return "weapon_rof_mode_2_secondary_bonus"
	case WeaponNonChamberShotsBonus:
		return "weapon_non_chamber_shots_bonus"
	case WeaponChamberShotsBonus:
		return "weapon_chamber_shots_bonus"
	case WeaponShotDurationBonus:
		return "weapon_shot_duration_bonus"
	case WeaponReloadTimeBonus:
		return "weapon_reload_time_bonus"
	case WeaponSwitch:
		return "weapon_switch"
	case CostReduction:
		return "cost_reduction"
	case ContainedWeightReduction:
		return "contained_weight_reduction"
	default:
		return Type(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case AttributeBonus:
		return i18n.Text("Gives an attribute modifier of")
	case ConditionalModifier:
		return i18n.Text("Gives a conditional modifier of")
	case DRBonus:
		return i18n.Text("Gives a DR bonus of")
	case ReactionBonus:
		return i18n.Text("Gives a reaction modifier of")
	case SkillBonus:
		return i18n.Text("Gives a skill level modifier of")
	case SkillPointBonus:
		return i18n.Text("Gives a skill point modifier of")
	case SpellBonus:
		return i18n.Text("Gives a spell level modifier of")
	case SpellPointBonus:
		return i18n.Text("Gives a spell point modifier of")
	case WeaponBonus:
		return i18n.Text("Gives a weapon damage modifier of")
	case WeaponAccBonus:
		return i18n.Text("Gives a weapon accuracy modifier of")
	case WeaponScopeAccBonus:
		return i18n.Text("Gives a weapon scope accuracy modifier of")
	case WeaponDRDivisorBonus:
		return i18n.Text("Gives a weapon DR divisor modifier of")
	case WeaponEffectiveSTBonus:
		return i18n.Text("Gives a weapon effective ST modifier of")
	case WeaponMinSTBonus:
		return i18n.Text("Gives a weapon minimum ST modifier of")
	case WeaponMinReachBonus:
		return i18n.Text("Gives a weapon minimum reach modifier of")
	case WeaponMaxReachBonus:
		return i18n.Text("Gives a weapon maximum reach modifier of")
	case WeaponHalfDamageRangeBonus:
		return i18n.Text("Gives a weapon half-damage range modifier of")
	case WeaponMinRangeBonus:
		return i18n.Text("Gives a weapon minimum range modifier of")
	case WeaponMaxRangeBonus:
		return i18n.Text("Gives a weapon maximum range modifier of")
	case WeaponRecoilBonus:
		return i18n.Text("Gives a weapon recoil modifier of")
	case WeaponBulkBonus:
		return i18n.Text("Gives a weapon bulk modifier of")
	case WeaponParryBonus:
		return i18n.Text("Gives a weapon parry modifier of")
	case WeaponBlockBonus:
		return i18n.Text("Gives a weapon block modifier of")
	case WeaponRofMode1ShotsBonus:
		return i18n.Text("Gives a weapon shots per attack (mode 1) modifier of")
	case WeaponRofMode1SecondaryBonus:
		return i18n.Text("Gives a weapon secondary projectiles (mode 1) modifier of")
	case WeaponRofMode2ShotsBonus:
		return i18n.Text("Gives a weapon shots per attack (mode 2) modifier of")
	case WeaponRofMode2SecondaryBonus:
		return i18n.Text("Gives a weapon secondary projectiles (mode 2) modifier of")
	case WeaponNonChamberShotsBonus:
		return i18n.Text("Gives a weapon non-chamber shots modifier of")
	case WeaponChamberShotsBonus:
		return i18n.Text("Gives a weapon chamber shots modifier of")
	case WeaponShotDurationBonus:
		return i18n.Text("Gives a weapon shot duration modifier of")
	case WeaponReloadTimeBonus:
		return i18n.Text("Gives a weapon reload time modifier of")
	case WeaponSwitch:
		return i18n.Text("Set the weapon flag")
	case CostReduction:
		return i18n.Text("Reduces the attribute cost of")
	case ContainedWeightReduction:
		return i18n.Text("Reduces the contained weight by")
	default:
		return Type(0).String()
	}
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

// ExtractType extracts the value from a string.
func ExtractType(str string) Type {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
