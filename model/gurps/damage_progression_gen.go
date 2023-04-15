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
	BasicSet DamageProgression = iota
	KnowingYourOwnStrength
	NoSchoolGrognardDamage
	ThrustEqualsSwingMinus2
	SwingEqualsThrustPlus2
	PhoenixFlameD3
	LastDamageProgression = PhoenixFlameD3
)

// AllDamageProgression holds all possible values.
var AllDamageProgression = []DamageProgression{
	BasicSet,
	KnowingYourOwnStrength,
	NoSchoolGrognardDamage,
	ThrustEqualsSwingMinus2,
	SwingEqualsThrustPlus2,
	PhoenixFlameD3,
}

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
	switch enum {
	case BasicSet:
		return "basic_set"
	case KnowingYourOwnStrength:
		return "knowing_your_own_strength"
	case NoSchoolGrognardDamage:
		return "no_school_grognard_damage"
	case ThrustEqualsSwingMinus2:
		return "thrust_equals_swing_minus_2"
	case SwingEqualsThrustPlus2:
		return "swing_equals_thrust_plus_2"
	case PhoenixFlameD3:
		return "phoenix_flame_d3"
	default:
		return DamageProgression(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum DamageProgression) String() string {
	switch enum {
	case BasicSet:
		return i18n.Text("Basic Set")
	case KnowingYourOwnStrength:
		return i18n.Text("Knowing Your Own Strength")
	case NoSchoolGrognardDamage:
		return i18n.Text("No School Grognard Damage")
	case ThrustEqualsSwingMinus2:
		return i18n.Text("Thrust = Swing-2")
	case SwingEqualsThrustPlus2:
		return i18n.Text("Swing = Thrust+2")
	case PhoenixFlameD3:
		return i18n.Text("Phoenix Flame D3")
	default:
		return DamageProgression(0).String()
	}
}

// AltString returns the alternate string.
func (enum DamageProgression) AltString() string {
	switch enum {
	case BasicSet:
		return ""
	case KnowingYourOwnStrength:
		return i18n.Text("Pyramid 3-83, pages 16-19")
	case NoSchoolGrognardDamage:
		return i18n.Text("https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html")
	case ThrustEqualsSwingMinus2:
		return i18n.Text("https://github.com/richardwilkes/gcs/issues/97")
	case SwingEqualsThrustPlus2:
		return i18n.Text("Houserule originating with Kevin Smyth. See https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/")
	case PhoenixFlameD3:
		return i18n.Text("Houserule that use d3s instead of d6s for Damage. See: https://github.com/richardwilkes/gcs/pull/393")
	default:
		return DamageProgression(0).AltString()
	}
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

// ExtractDamageProgression extracts the value from a string.
func ExtractDamageProgression(str string) DamageProgression {
	for _, enum := range AllDamageProgression {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
