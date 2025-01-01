// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package progression

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	BasicSet Option = iota
	KnowingYourOwnStrength
	NoSchoolGrognardDamage
	ThrustEqualsSwingMinus2
	SwingEqualsThrustPlus2
	Tbone1
	Tbone1Clean
	Tbone2
	Tbone2Clean
	PhoenixFlameD3
)

// LastOption is the last valid value.
const LastOption Option = PhoenixFlameD3

// Options holds all possible values.
var Options = []Option{
	BasicSet,
	KnowingYourOwnStrength,
	NoSchoolGrognardDamage,
	ThrustEqualsSwingMinus2,
	SwingEqualsThrustPlus2,
	Tbone1,
	Tbone1Clean,
	Tbone2,
	Tbone2Clean,
	PhoenixFlameD3,
}

// Option controls how Thrust and Swing are calculated.
type Option byte

// EnsureValid ensures this is of a known value.
func (enum Option) EnsureValid() Option {
	if enum <= PhoenixFlameD3 {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Option) Key() string {
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
	case Tbone1:
		return "tbone_1"
	case Tbone1Clean:
		return "tbone_1_clean"
	case Tbone2:
		return "tbone_2"
	case Tbone2Clean:
		return "tbone_2_clean"
	case PhoenixFlameD3:
		return "phoenix_flame_d3"
	default:
		return Option(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Option) String() string {
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
	case Tbone1:
		return i18n.Text("T Bone's New Damage for ST (option 1)")
	case Tbone1Clean:
		return i18n.Text("T Bone's New Damage for ST (option 1, cleaned)")
	case Tbone2:
		return i18n.Text("T Bone's New Damage for ST (option 2)")
	case Tbone2Clean:
		return i18n.Text("T Bone's New Damage for ST (option 2, cleaned)")
	case PhoenixFlameD3:
		return i18n.Text("Phoenix Flame D3")
	default:
		return Option(0).String()
	}
}

// AltString returns the alternate string.
func (enum Option) AltString() string {
	switch enum {
	case BasicSet:
		return i18n.Text("*The standard damage progression*")
	case KnowingYourOwnStrength:
		return i18n.Text("*From [Pyramid 3-83, pages 16-19](PY83:16)*")
	case NoSchoolGrognardDamage:
		return i18n.Text("*From [Adjusting Swing Damage in Dungeon Fantasy](https://noschoolgrognard.blogspot.com/2013/04/adjusting-swing-damage-in-dungeon.html)*")
	case ThrustEqualsSwingMinus2:
		return i18n.Text("*From [Alternate Damage Scheme (Thr = Sw-2)](https://github.com/richardwilkes/gcs/issues/97)*")
	case SwingEqualsThrustPlus2:
		return i18n.Text("*From a [house rule](https://gamingballistic.com/2020/12/04/df-eastmarch-boss-fight-and-house-rules/) originating with Kevin Smyth*")
	case Tbone1:
		return i18n.Text("*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*")
	case Tbone1Clean:
		return i18n.Text("*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*")
	case Tbone2:
		return i18n.Text("*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*")
	case Tbone2Clean:
		return i18n.Text("*From [T Bone's Games Diner](https://www.gamesdiner.com/rules-nugget-gurps-new-damage-for-st/)*")
	case PhoenixFlameD3:
		return i18n.Text("*From a [house rule](https://github.com/richardwilkes/gcs/pull/393) that uses d3s instead of d6s for damage*")
	default:
		return Option(0).AltString()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum Option) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *Option) UnmarshalText(text []byte) error {
	*enum = ExtractOption(string(text))
	return nil
}

// ExtractOption extracts the value from a string.
func ExtractOption(str string) Option {
	for _, enum := range Options {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
