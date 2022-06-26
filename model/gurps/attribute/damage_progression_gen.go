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

package attribute

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	BasicSet DamageProgression = iota
	KnowingYourOwnStrength
	NoSchoolGrognardDamage
	ThrustEqualsSwingMinus2
	SwingEqualsThrustPlus2
	PhoenixFlameD3
	LastDamageProgression = PhoenixFlameD3
)

var (
	// AllDamageProgression holds all possible values.
	AllDamageProgression = []DamageProgression{
		BasicSet,
		KnowingYourOwnStrength,
		NoSchoolGrognardDamage,
		ThrustEqualsSwingMinus2,
		SwingEqualsThrustPlus2,
		PhoenixFlameD3,
	}
	damageProgressionData = []struct {
		key    string
		string string
		alt    string
	}{
		{
			key:    "basic_set",
			string: i18n.Text("Basic Set"),
		},
		{
			key:    "knowing_your_own_strength",
			string: i18n.Text("Knowing Your Own Strength"),
			alt:    i18n.Text("Pyramid 3-83, pages 16-19"),
		},
		{
			key:    "no_school_grognard_damage",
			string: i18n.Text("No School Grognard Damage"),
			alt:    i18n.Text("https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html"),
		},
		{
			key:    "thrust_equals_swing_minus_2",
			string: i18n.Text("Thrust = Swing-2"),
			alt:    i18n.Text("https://github.com/richardwilkes/gcs/issues/97"),
		},
		{
			key:    "swing_equals_thrust_plus_2",
			string: i18n.Text("Swing = Thrust+2"),
			alt:    i18n.Text("Houserule originating with Kevin Smyth. See https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/"),
		},
		{
			key:    "phoenix_flame_d3",
			string: i18n.Text("Phoenix Flame D3"),
			alt:    i18n.Text("Houserule that use d3s instead of d6s for Damage. See: https://github.com/richardwilkes/gcs/pull/393"),
		},
	}
)

// DamageProgression controls how Thrust and Swing are calculated.
type DamageProgression byte

// EnsureValid ensures this is of a known value.
func (enum DamageProgression) EnsureValid() DamageProgression {
	if enum <= LastDamageProgression {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum DamageProgression) Key() string {
	return damageProgressionData[enum.EnsureValid()].key
}

// String implements fmt.Stringer.
func (enum DamageProgression) String() string {
	return damageProgressionData[enum.EnsureValid()].string
}

// AltString returns the alternate string.
func (enum DamageProgression) AltString() string {
	return damageProgressionData[enum.EnsureValid()].alt
}

// ExtractDamageProgression extracts the value from a string.
func ExtractDamageProgression(str string) DamageProgression {
	for i, one := range damageProgressionData {
		if strings.EqualFold(one.key, str) {
			return DamageProgression(i)
		}
	}
	return 0
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum DamageProgression) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *DamageProgression) UnmarshalText(text []byte) error {
	*enum = ExtractDamageProgression(string(text))
	return nil
}
