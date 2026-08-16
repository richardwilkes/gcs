// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package prereq

import (
	"slices"
	"strings"

	"github.com/richardwilkes/toolbox/v2/i18n"
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
	Script
	Unknown
)

// DefaultType is the default value.
const DefaultType Type = Unknown

// FirstType is the first valid value.
const FirstType Type = List

// LastType is the last valid value.
const LastType Type = Unknown

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
	Script,
}

// Type holds the type of a Prereq.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum >= FirstType && enum <= LastType {
		return enum
	}
	return DefaultType
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
	case Script:
		return "script_prereq"
	default:
		return "unknown_prereq"
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
	case Script:
		return nil
	default:
		return nil
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case List:
		return i18n.Text(`a list`)
	case Trait:
		return i18n.Text(`a trait`)
	case Attribute:
		return i18n.Text(`the attribute`)
	case ContainedQuantity:
		return i18n.Text(`a contained quantity of`)
	case ContainedWeight:
		return i18n.Text(`a contained weight`)
	case EquippedEquipment:
		return i18n.Text(`has equipped equipment`)
	case Skill:
		return i18n.Text(`a skill`)
	case Spell:
		return i18n.Text(`spell(s)`)
	case Script:
		return i18n.Text(`has script`)
	case Unknown:
		return i18n.Text(`an unknown prerequisite type`)
	default:
		return i18n.Text(`an unknown prerequisite type`)
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
		if slices.ContainsFunc(enum.oldKeys(), func(s string) bool { return strings.EqualFold(s, str) }) {
			return enum
		}
	}
	return DefaultType
}

// ExtractKnownType extracts the value from a string, reporting whether the string was actually recognized.
//
// Unlike ExtractType, which quietly maps anything it doesn't recognize onto the first value, this permits a caller that
// is dispatching on the type to detect unknown types.
func ExtractKnownType(str string) (value Type, known bool) {
	for _, enum := range Types {
		if strings.EqualFold(enum.Key(), str) {
			return enum, true
		}
		if slices.ContainsFunc(enum.oldKeys(), func(s string) bool { return strings.EqualFold(s, str) }) {
			return enum, true
		}
	}
	return DefaultType, false
}
