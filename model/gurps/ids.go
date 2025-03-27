// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"strings"
)

const containerKeyPostfix = "_container"

// Various commonly used IDs
const (
	AllID              = "all"
	BasicMoveID        = "basic_move"
	BasicSpeedID       = "basic_speed"
	BlockID            = "block"
	DexterityID        = "dx"
	DodgeID            = "dodge"
	IntelligenceID     = "iq"
	LiftingStrengthID  = "lifting_st"
	MoveID             = "move"
	ParryID            = "parry"
	RitualMagicSpellID = "ritual_magic_spell"
	SizeModifierID     = "sm"
	SkillID            = "skill"
	SpellID            = "spell"
	StrengthID         = "st"
	StrikingStrengthID = "striking_st"
	TechniqueID        = "technique"
	ThrowingStrengthID = "throwing_st"
	TorsoID            = "torso"
)

// SanitizeID ensures the ID is not empty and consists of only lowercase alphanumeric characters. If permitLeadingDigits
// is false, then leading digits are stripped. A list of reserved values can be passed in to disallow specific IDs.
func SanitizeID(id string, permitLeadingDigits bool, reserved ...string) string {
	var buffer strings.Builder
	buffer.Grow(len(id))
	for _, ch := range id {
		if ch >= 'A' && ch <= 'Z' {
			ch += 'a' - 'A'
		}
		if ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9' && (permitLeadingDigits || buffer.Len() > 0)) {
			buffer.WriteRune(ch)
		}
	}
	if buffer.Len() == 0 {
		buffer.WriteByte('_')
	}
	for {
		ok := true
		id = buffer.String()
		for _, one := range reserved {
			if one == id {
				buffer.WriteByte('_')
				ok = false
				break
			}
		}
		if ok {
			return id
		}
	}
}
