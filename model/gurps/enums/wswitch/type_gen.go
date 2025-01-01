// Code generated from "enum.go.tmpl" - DO NOT EDIT.

// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package wswitch

import (
	"strings"

	"github.com/richardwilkes/toolbox/i18n"
)

// Possible values.
const (
	NotSwitched Type = iota
	CanBlock
	CanParry
	CloseCombat
	Fencing
	FullAuto1
	FullAuto2
	Bipod
	ControlledBursts1
	ControlledBursts2
	Jet
	Mounted
	MusclePowered
	RangeInMiles
	ReachChangeRequiresReady
	ReloadTimeIsPerShot
	RetractingStock
	TwoHanded
	Thrown
	Unbalanced
	TwoHandedAndUnreadyAfterAttack
	MusketRest
)

// LastType is the last valid value.
const LastType Type = MusketRest

// Types holds all possible values.
var Types = []Type{
	NotSwitched,
	CanBlock,
	CanParry,
	CloseCombat,
	Fencing,
	FullAuto1,
	FullAuto2,
	Bipod,
	ControlledBursts1,
	ControlledBursts2,
	Jet,
	Mounted,
	MusclePowered,
	RangeInMiles,
	ReachChangeRequiresReady,
	ReloadTimeIsPerShot,
	RetractingStock,
	TwoHanded,
	Thrown,
	Unbalanced,
	TwoHandedAndUnreadyAfterAttack,
	MusketRest,
}

// Type holds the type of a weapon switch.
type Type byte

// EnsureValid ensures this is of a known value.
func (enum Type) EnsureValid() Type {
	if enum <= MusketRest {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum Type) Key() string {
	switch enum {
	case NotSwitched:
		return "not_switched"
	case CanBlock:
		return "can_block"
	case CanParry:
		return "can_parry"
	case CloseCombat:
		return "close_combat"
	case Fencing:
		return "fencing"
	case FullAuto1:
		return "full_auto_1"
	case FullAuto2:
		return "full_auto_2"
	case Bipod:
		return "bipod"
	case ControlledBursts1:
		return "controlled_bursts_1"
	case ControlledBursts2:
		return "controlled_bursts_2"
	case Jet:
		return "jet"
	case Mounted:
		return "mounted"
	case MusclePowered:
		return "muscle_powered"
	case RangeInMiles:
		return "range_in_miles"
	case ReachChangeRequiresReady:
		return "reach_change_requires_ready"
	case ReloadTimeIsPerShot:
		return "reload_time_is_per_shot"
	case RetractingStock:
		return "retracting_stock"
	case TwoHanded:
		return "two-handed"
	case Thrown:
		return "thrown"
	case Unbalanced:
		return "unbalanced"
	case TwoHandedAndUnreadyAfterAttack:
		return "two-handed_and_unready_after_attack"
	case MusketRest:
		return "musket_rest"
	default:
		return Type(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum Type) String() string {
	switch enum {
	case NotSwitched:
		return ""
	case CanBlock:
		return i18n.Text("Can Block")
	case CanParry:
		return i18n.Text("Can Parry")
	case CloseCombat:
		return i18n.Text("Close Combat")
	case Fencing:
		return i18n.Text("Fencing")
	case FullAuto1:
		return i18n.Text("Fully Automatic (Mode 1)")
	case FullAuto2:
		return i18n.Text("Fully Automatic (Mode 2)")
	case Bipod:
		return i18n.Text("Has Bipod")
	case ControlledBursts1:
		return i18n.Text("High-cyclic Controlled Bursts (Mode 1)")
	case ControlledBursts2:
		return i18n.Text("High-cyclic Controlled Bursts (Mode 2)")
	case Jet:
		return i18n.Text("Jet")
	case Mounted:
		return i18n.Text("Mounted")
	case MusclePowered:
		return i18n.Text("Muscle Powered")
	case RangeInMiles:
		return i18n.Text("Range In Miles")
	case ReachChangeRequiresReady:
		return i18n.Text("Reach Change Requires Ready")
	case ReloadTimeIsPerShot:
		return i18n.Text("Reload Time Is Per Shot")
	case RetractingStock:
		return i18n.Text("Retracting Stock")
	case TwoHanded:
		return i18n.Text("Two-handed")
	case Thrown:
		return i18n.Text("Thrown")
	case Unbalanced:
		return i18n.Text("Unbalanced")
	case TwoHandedAndUnreadyAfterAttack:
		return i18n.Text("Two-handed and Unready After Attack")
	case MusketRest:
		return i18n.Text("Uses a Musket Rest")
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
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
