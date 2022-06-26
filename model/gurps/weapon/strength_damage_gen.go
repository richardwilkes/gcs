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

package weapon

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	None StrengthDamage = iota
	Thrust
	LeveledThrust
	Swing
	LeveledSwing
	LastStrengthDamage = LeveledSwing
)

var (
	// AllStrengthDamage holds all possible values.
	AllStrengthDamage = []StrengthDamage{
		None,
		Thrust,
		LeveledThrust,
		Swing,
		LeveledSwing,
	}
	strengthDamageData = []struct {
		key    string
		string string
	}{
		{
			key:    "none",
			string: i18n.Text("None"),
		},
		{
			key:    "thr",
			string: "thr",
		},
		{
			key:    "thr_leveled",
			string: i18n.Text("thr (leveled)"),
		},
		{
			key:    "sw",
			string: "sw",
		},
		{
			key:    "sw_leveled",
			string: i18n.Text("sw (leveled)"),
		},
	}
)

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
	return strengthDamageData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum StrengthDamage) String() string {
	return strengthDamageData[enum.EnsureValid()].string
}

// ExtractStrengthDamage extracts the value from a string.
func ExtractStrengthDamage(str string) StrengthDamage {
	for i, one := range strengthDamageData {
		if strings.EqualFold(one.key, str) {
			return StrengthDamage(i)
		}
	}
	return 0
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
