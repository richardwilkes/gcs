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
	NotSwitchedWeaponSwitchType WeaponSwitchType = iota
	CanBlockWeaponSwitchType
	CanParryWeaponSwitchType
	CloseCombatWeaponSwitchType
	FencingWeaponSwitchType
	FullAuto1WeaponSwitchType
	FullAuto2WeaponSwitchType
	BipodWeaponSwitchType
	ControlledBursts1WeaponSwitchType
	ControlledBursts2WeaponSwitchType
	JetWeaponSwitchType
	MountedWeaponSwitchType
	ReachChangeRequiresReadyWeaponSwitchType
	RetractingStockWeaponSwitchType
	TwoHandedWeaponSwitchType
	UnbalancedWeaponSwitchType
	TwoHandedAndUnreadyAfterAttackWeaponSwitchType
	MusketRestWeaponSwitchType
	LastWeaponSwitchType = MusketRestWeaponSwitchType
)

// AllWeaponSwitchType holds all possible values.
var AllWeaponSwitchType = []WeaponSwitchType{
	NotSwitchedWeaponSwitchType,
	CanBlockWeaponSwitchType,
	CanParryWeaponSwitchType,
	CloseCombatWeaponSwitchType,
	FencingWeaponSwitchType,
	FullAuto1WeaponSwitchType,
	FullAuto2WeaponSwitchType,
	BipodWeaponSwitchType,
	ControlledBursts1WeaponSwitchType,
	ControlledBursts2WeaponSwitchType,
	JetWeaponSwitchType,
	MountedWeaponSwitchType,
	ReachChangeRequiresReadyWeaponSwitchType,
	RetractingStockWeaponSwitchType,
	TwoHandedWeaponSwitchType,
	UnbalancedWeaponSwitchType,
	TwoHandedAndUnreadyAfterAttackWeaponSwitchType,
	MusketRestWeaponSwitchType,
}

// WeaponSwitchType holds the type of a weapon switch.
type WeaponSwitchType byte

// EnsureValid ensures this is of a known value.
func (enum WeaponSwitchType) EnsureValid() WeaponSwitchType {
	if enum <= LastWeaponSwitchType {
		return enum
	}
	return 0
}

// Key returns the key used in serialization.
func (enum WeaponSwitchType) Key() string {
	switch enum {
	case NotSwitchedWeaponSwitchType:
		return "not_switched"
	case CanBlockWeaponSwitchType:
		return "can_block"
	case CanParryWeaponSwitchType:
		return "can_parry"
	case CloseCombatWeaponSwitchType:
		return "close_combat"
	case FencingWeaponSwitchType:
		return "fencing"
	case FullAuto1WeaponSwitchType:
		return "full_auto_1"
	case FullAuto2WeaponSwitchType:
		return "full_auto_2"
	case BipodWeaponSwitchType:
		return "bipod"
	case ControlledBursts1WeaponSwitchType:
		return "controlled_bursts_1"
	case ControlledBursts2WeaponSwitchType:
		return "controlled_bursts_2"
	case JetWeaponSwitchType:
		return "jet"
	case MountedWeaponSwitchType:
		return "mounted"
	case ReachChangeRequiresReadyWeaponSwitchType:
		return "reach_change_requires_ready"
	case RetractingStockWeaponSwitchType:
		return "retracting_stock"
	case TwoHandedWeaponSwitchType:
		return "two-handed"
	case UnbalancedWeaponSwitchType:
		return "unbalanced"
	case TwoHandedAndUnreadyAfterAttackWeaponSwitchType:
		return "two-handed_and_unready_after_attack"
	case MusketRestWeaponSwitchType:
		return "musket_rest"
	default:
		return WeaponSwitchType(0).Key()
	}
}

// String implements fmt.Stringer.
func (enum WeaponSwitchType) String() string {
	switch enum {
	case NotSwitchedWeaponSwitchType:
		return ""
	case CanBlockWeaponSwitchType:
		return i18n.Text("Can Block")
	case CanParryWeaponSwitchType:
		return i18n.Text("Can Parry")
	case CloseCombatWeaponSwitchType:
		return i18n.Text("Close Combat")
	case FencingWeaponSwitchType:
		return i18n.Text("Fencing")
	case FullAuto1WeaponSwitchType:
		return i18n.Text("Fully Automatic (Mode 1)")
	case FullAuto2WeaponSwitchType:
		return i18n.Text("Fully Automatic (Mode 2)")
	case BipodWeaponSwitchType:
		return i18n.Text("Has Bipod")
	case ControlledBursts1WeaponSwitchType:
		return i18n.Text("High-cyclic Controlled Bursts (Mode 1)")
	case ControlledBursts2WeaponSwitchType:
		return i18n.Text("High-cyclic Controlled Bursts (Mode 2)")
	case JetWeaponSwitchType:
		return i18n.Text("Jet")
	case MountedWeaponSwitchType:
		return i18n.Text("Mounted")
	case ReachChangeRequiresReadyWeaponSwitchType:
		return i18n.Text("Reach Change Requires Ready")
	case RetractingStockWeaponSwitchType:
		return i18n.Text("Retracting Stock")
	case TwoHandedWeaponSwitchType:
		return i18n.Text("Two-handed")
	case UnbalancedWeaponSwitchType:
		return i18n.Text("Unbalanced")
	case TwoHandedAndUnreadyAfterAttackWeaponSwitchType:
		return i18n.Text("Two-handed and Unready After Attack")
	case MusketRestWeaponSwitchType:
		return i18n.Text("Uses a Musket Rest")
	default:
		return WeaponSwitchType(0).String()
	}
}

// MarshalText implements the encoding.TextMarshaler interface.
func (enum WeaponSwitchType) MarshalText() (text []byte, err error) {
	return []byte(enum.Key()), nil
}

// UnmarshalText implements the encoding.TextUnmarshaler interface.
func (enum *WeaponSwitchType) UnmarshalText(text []byte) error {
	*enum = ExtractWeaponSwitchType(string(text))
	return nil
}

// ExtractWeaponSwitchType extracts the value from a string.
func ExtractWeaponSwitchType(str string) WeaponSwitchType {
	for _, enum := range AllWeaponSwitchType {
		if strings.EqualFold(enum.Key(), str) {
			return enum
		}
	}
	return 0
}
