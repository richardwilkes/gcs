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
	MeleeWeaponType WeaponType = iota
	RangedWeaponType
	LastWeaponType = RangedWeaponType
)

// AllWeaponType holds all possible values.
var AllWeaponType = []WeaponType{
	MeleeWeaponType,
	RangedWeaponType,
}

// WeaponType holds the type of an weapon definition.
type WeaponType byte

// EnsureValid ensures this is of a known value.
func (enum WeaponType) EnsureValid() WeaponType {
	if enum <= LastWeaponType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum WeaponType) Key() string {
	switch enum {
	case MeleeWeaponType:
		return "melee_weapon"
	case RangedWeaponType:
		return "ranged_weapon"
	default:
		return WeaponType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum WeaponType) String() string {
	switch enum {
	case MeleeWeaponType:
		return i18n.Text("Melee Weapon")
	case RangedWeaponType:
		return i18n.Text("Ranged Weapon")
	default:
		return WeaponType(0).String()
	}
}

// AltString returns the alternate string.
func (enum WeaponType) AltString() string {
	switch enum {
	case MeleeWeaponType:
		return i18n.Text("Melee Weapons")
	case RangedWeaponType:
		return i18n.Text("Ranged Weapons")
	default:
		return WeaponType(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum WeaponType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *WeaponType) UnmarshalText(text []byte) error {
	*enum = ExtractWeaponType(string(text))
	return nil
}

// ExtractWeaponType extracts the value from a string.
func ExtractWeaponType(str string) WeaponType {
	for _, enum := range AllWeaponType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
