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
	WithRequiredSkillWeaponSelectionType WeaponSelectionType = iota
	ThisWeaponWeaponSelectionType
	WithNameWeaponSelectionType
	LastWeaponSelectionType = WithNameWeaponSelectionType
)

// AllWeaponSelectionType holds all possible values.
var AllWeaponSelectionType = []WeaponSelectionType{
	WithRequiredSkillWeaponSelectionType,
	ThisWeaponWeaponSelectionType,
	WithNameWeaponSelectionType,
}

// WeaponSelectionType holds the type of a weapon selection.
type WeaponSelectionType byte

// EnsureValid ensures this is of a known value.
func (enum WeaponSelectionType) EnsureValid() WeaponSelectionType {
	if enum <= LastWeaponSelectionType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum WeaponSelectionType) Key() string {
	switch enum {
	case WithRequiredSkillWeaponSelectionType:
		return "weapons_with_required_skill"
	case ThisWeaponWeaponSelectionType:
		return "this_weapon"
	case WithNameWeaponSelectionType:
		return "weapons_with_name"
	default:
		return WeaponSelectionType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum WeaponSelectionType) String() string {
	switch enum {
	case WithRequiredSkillWeaponSelectionType:
		return i18n.Text("to weapons whose required skill name")
	case ThisWeaponWeaponSelectionType:
		return i18n.Text("to this weapon")
	case WithNameWeaponSelectionType:
		return i18n.Text("to weapons whose name")
	default:
		return WeaponSelectionType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum WeaponSelectionType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *WeaponSelectionType) UnmarshalText(text []byte) error {
	*enum = ExtractWeaponSelectionType(string(text))
	return nil
}

// ExtractWeaponSelectionType extracts the value from a string.
func ExtractWeaponSelectionType(str string) WeaponSelectionType {
	for _, enum := range AllWeaponSelectionType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
