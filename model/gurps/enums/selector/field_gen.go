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
	WeaponDamageType Field = iota
)

// LastField is the last valid value.
const LastField Field = WeaponDamageType

// Fields holds all possible values.
var Fields = []Field{
	WeaponDamageType,
}

// Field identifies a multi-state field that a SelectorOverride can replace.
type Field byte

// EnsureValid ensures this is of a known value.
func (enum Field) EnsureValid() Field {
	if enum <= WeaponDamageType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Field) Key() string {
	switch enum {
	case WeaponDamageType:
		return "weapon_damage_type"
	default:
		return Field(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Field) String() string {
	switch enum {
	case WeaponDamageType:
		return i18n.Text(`weapon damage type`)
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
