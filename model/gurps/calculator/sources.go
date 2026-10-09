// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
)

// SkillLevelOrDefault returns the entity's level in the named skill, falling back to its default from the attribute
// when the skill is not on the sheet. It is zero when even the default cannot be worked out.
func SkillLevelOrDefault(entity *gurps.Entity, name, defaultAttrID string, modifier int) int {
	if sk := entity.BestSkillNamed(name, "", false, nil); sk != nil {
		return sk.CalculateLevel(nil).Level.AsInteger[int]()
	}
	def := &gurps.SkillDefault{DefaultType: defaultAttrID, Modifier: fxp.FromInteger(modifier)}
	level := def.SkillLevelFast(entity, nil, false, nil, true)
	if level == fxp.Min {
		return 0
	}
	return level.AsInteger[int]()
}

// BasicLiftFor returns the Basic Lift the entity has with the given ST, or the one a character typed in has with it
// when the entity is nil, worked out under the global default sheet settings.
func BasicLiftFor(entity *gurps.Entity, st fxp.Int) fxp.Weight {
	if entity != nil {
		return entity.BasicLiftForST(st)
	}
	return gurps.BasicLiftForST(st, gurps.SheetSettingsFor(nil).DamageProgression)
}

// ThrustFor returns the thrust damage for the given ST under the entity's damage progression, or the global default one
// when the entity is nil.
func ThrustFor(entity *gurps.Entity, st fxp.Int) dice.Dice {
	return gurps.SheetSettingsFor(entity).DamageProgression.Thrust(st.AsInteger[int]())
}

// TorsoDR returns the entity's total DR on the torso and the part of it that comes from armor.
func TorsoDR(entity *gurps.Entity) (total, armor int) {
	body := entity.SheetSettings.BodyType
	if body == nil {
		return 0, 0
	}
	torso := body.LookupLocationByID(entity, gurps.TorsoID)
	if torso == nil {
		return 0, 0
	}
	return torso.DR(entity, nil, nil)[gurps.AllID], torso.ArmorDR(entity, nil)[gurps.AllID]
}

// spellLevel returns the highest level the entity knows the named spell at, or 0 when it does not know it.
func spellLevel(entity *gurps.Entity, name string) int {
	var level fxp.Int
	gurps.Traverse(func(sp *gurps.Spell) bool {
		if strings.EqualFold(sp.NameWithReplacements(), name) {
			level = level.Max(sp.CalculateLevel().Level)
		}
		return false
	}, true, true, entity.Spells...)
	return level.AsInteger[int]()
}
