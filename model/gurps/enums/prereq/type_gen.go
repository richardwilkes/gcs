// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package prereq

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
)

// Possible values.
const (
	List Type = iota
	Trait
	Attribute
	ContainedQuantity
	ContainedWeight
	EquippedEquipment
	Skill
	Spell
)

// LastType is the last valid value.
const LastType Type = Spell

// Types holds all possible values.
var Types = []Type{
	List,
	Trait,
	Attribute,
	ContainedQuantity,
	ContainedWeight,
	EquippedEquipment,
	Skill,
	Spell,
}

// Type holds the type of a Prereq.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= Spell {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case List:
		return "prereq_list"
	case Trait:
		return "trait_prereq"
	case Attribute:
		return "attribute_prereq"
	case ContainedQuantity:
		return "contained_quantity_prereq"
	case ContainedWeight:
		return "contained_weight_prereq"
	case EquippedEquipment:
		return "equipped_equipment"
	case Skill:
		return "skill_prereq"
	case Spell:
		return "spell_prereq"
	default:
		return Type(0).Key()
	}
}

func (enum Type) oldKeys() []string {
	switch enum {
	case List:
		return nil
	case Trait:
		return []string{"advantage_prereq"}
	case Attribute:
		return nil
	case ContainedQuantity:
		return nil
	case ContainedWeight:
		return nil
	case EquippedEquipment:
		return nil
	case Skill:
		return nil
	case Spell:
		return nil
	default:
		return Type(0).oldKeys()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case List:
		return i18n.Text("a list")
	case Trait:
		return i18n.Text("a trait")
	case Attribute:
		return i18n.Text("the attribute")
	case ContainedQuantity:
		return i18n.Text("a contained quantity of")
	case ContainedWeight:
		return i18n.Text("a contained weight")
	case EquippedEquipment:
		return i18n.Text("has equipped equipment")
	case Skill:
		return i18n.Text("a skill")
	case Spell:
		return i18n.Text("spell(s)")
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
		if strings.EqualFold(enum.Key(), str) || txt.CaselessSliceContains(enum.oldKeys(), str) {
			return enum
		}
	}
	return 0
}
