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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
)

// CollisionParticipant is what a character sheet supplies about one of the objects in a collision: the one doing the
// moving, or the one it hits.
type CollisionParticipant struct {
	HP         fxp.Int // The maximum HP; meaningful only when HasHP.
	ST         fxp.Int // The maximum ST; meaningful only when HasST.
	Velocity   fxp.Int // The Move at the current encumbrance, suggested as the velocity in yards per second.
	SM         int
	Acrobatics int  // The Acrobatics skill level, or its default from DX.
	Swimming   int  // The Swimming skill level, or its default from HT.
	ArmorDR    int  // The torso DR that comes from armor, which counts as flexible against a fall (BX431).
	InnateDR   int  // The rest of the torso DR, which does not.
	HasHP      bool // Whether the sheet defines HP at all.
	HasST      bool // Whether the sheet defines ST at all.
}

// CollisionParticipantFromEntity reads what the sheet's character supplies about an object in a collision. The entity
// must not be nil.
func CollisionParticipantFromEntity(entity *gurps.Entity) CollisionParticipant {
	p := CollisionParticipant{
		Velocity:   fxp.FromInteger(entity.Move(entity.EncumbranceLevel(false))),
		SM:         entity.Profile.AdjustedSizeModifier(),
		Acrobatics: SkillLevelOrDefault(entity, "Acrobatics", gurps.DexterityID, -6),
		Swimming:   SkillLevelOrDefault(entity, "Swimming", gurps.HealthID, -4),
	}
	if entity.ResolveAttribute(gurps.HitPointsID) != nil {
		p.HP = entity.Attributes.Maximum(gurps.HitPointsID).Max(0)
		p.HasHP = true
	}
	if entity.ResolveAttribute(gurps.StrengthID) != nil {
		p.ST = entity.Attributes.Maximum(gurps.StrengthID).Max(0)
		p.HasST = true
	}
	total, armor := TorsoDR(entity)
	p.ArmorDR = armor
	p.InnateDR = total - armor
	return p
}
