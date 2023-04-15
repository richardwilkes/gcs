// Code generated from "enum.go.tmpl" - DO NOT EDIT.

/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	NoneStrengthDamage StrengthDamage = iota
	ThrustStrengthDamage
	LeveledThrustStrengthDamage
	SwingStrengthDamage
	LeveledSwingStrengthDamage
	LastStrengthDamage = LeveledSwingStrengthDamage
)

// AllStrengthDamage holds all possible values.
var AllStrengthDamage = []StrengthDamage{
	NoneStrengthDamage,
	ThrustStrengthDamage,
	LeveledThrustStrengthDamage,
	SwingStrengthDamage,
	LeveledSwingStrengthDamage,
}

// StrengthDamage holds the type of strength dice to add to damage.
type StrengthDamage byte

// EnsureValid ensures this is of a known value.
func (enum StrengthDamage) EnsureValid() StrengthDamage {
	if enum <= LastStrengthDamage {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum StrengthDamage) Key() string {
	switch enum {
	case NoneStrengthDamage:
		return "none"
	case ThrustStrengthDamage:
		return "thr"
	case LeveledThrustStrengthDamage:
		return "thr_leveled"
	case SwingStrengthDamage:
		return "sw"
	case LeveledSwingStrengthDamage:
		return "sw_leveled"
	default:
		return StrengthDamage(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum StrengthDamage) String() string {
	switch enum {
	case NoneStrengthDamage:
		return i18n.Text("None")
	case ThrustStrengthDamage:
		return "thr"
	case LeveledThrustStrengthDamage:
		return i18n.Text("thr (leveled)")
	case SwingStrengthDamage:
		return "sw"
	case LeveledSwingStrengthDamage:
		return i18n.Text("sw (leveled)")
	default:
		return StrengthDamage(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum StrengthDamage) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *StrengthDamage) UnmarshalText(text []byte) error {
	*enum = ExtractStrengthDamage(string(text))
	return nil
}

// ExtractStrengthDamage extracts the value from a string.
func ExtractStrengthDamage(str string) StrengthDamage {
	for _, enum := range AllStrengthDamage {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
