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
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	ListPrereqType PrereqType = iota
	TraitPrereqType
	AttributePrereqType
	ContainedQuantityPrereqType
	ContainedWeightPrereqType
	EquippedEquipmentPrereqType
	SkillPrereqType
	SpellPrereqType
	LastPrereqType = SpellPrereqType
)

// AllPrereqType holds all possible values.
var AllPrereqType = []PrereqType{
	ListPrereqType,
	TraitPrereqType,
	AttributePrereqType,
	ContainedQuantityPrereqType,
	ContainedWeightPrereqType,
	EquippedEquipmentPrereqType,
	SkillPrereqType,
	SpellPrereqType,
}

// PrereqType holds the type of a Prereq.
type PrereqType byte

// EnsureValid ensures this is of a known value.
func (enum PrereqType) EnsureValid() PrereqType {
	if enum <= LastPrereqType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum PrereqType) Key() string {
	switch enum {
	case ListPrereqType:
		return "prereq_list"
	case TraitPrereqType:
		return "trait_prereq"
	case AttributePrereqType:
		return "attribute_prereq"
	case ContainedQuantityPrereqType:
		return "contained_quantity_prereq"
	case ContainedWeightPrereqType:
		return "contained_weight_prereq"
	case EquippedEquipmentPrereqType:
		return "equipped_equipment"
	case SkillPrereqType:
		return "skill_prereq"
	case SpellPrereqType:
		return "spell_prereq"
	default:
		return PrereqType(0).Key()
	}
}

func (enum PrereqType) oldKeys() []string {
	switch enum {
	case ListPrereqType:
		return nil
	case TraitPrereqType:
		return []string{"advantage_prereq"}
	case AttributePrereqType:
		return nil
	case ContainedQuantityPrereqType:
		return nil
	case ContainedWeightPrereqType:
		return nil
	case EquippedEquipmentPrereqType:
		return nil
	case SkillPrereqType:
		return nil
	case SpellPrereqType:
		return nil
	default:
		return PrereqType(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum PrereqType) String() string {
	switch enum {
	case ListPrereqType:
		return i18n.Text("a list")
	case TraitPrereqType:
		return i18n.Text("a trait")
	case AttributePrereqType:
		return i18n.Text("the attribute")
	case ContainedQuantityPrereqType:
		return i18n.Text("a contained quantity of")
	case ContainedWeightPrereqType:
		return i18n.Text("a contained weight")
	case EquippedEquipmentPrereqType:
		return i18n.Text("has equipped equipment")
	case SkillPrereqType:
		return i18n.Text("a skill")
	case SpellPrereqType:
		return i18n.Text("spell(s)")
	default:
		return PrereqType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum PrereqType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *PrereqType) UnmarshalText(text []byte) error {
	*enum = ExtractPrereqType(string(text))
	return nil
}

// ExtractPrereqType extracts the value from a string.
func ExtractPrereqType(str string) PrereqType {
	for _, enum := range AllPrereqType {
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
