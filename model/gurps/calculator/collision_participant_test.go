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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestCollisionParticipantFromEntity verifies what a sheet supplies about an object in a collision: its maximum HP
// and ST, its Move at the current encumbrance as the suggested velocity, its SM, its skills or their defaults, and
// its torso DR split between armor and the rest.
func TestCollisionParticipantFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	e.Profile.SizeModifier = 1
	hp := e.Attributes.Set[gurps.HitPointsID]
	hp.Adjustment = fxp.Two
	hp.Damage = fxp.Three
	e.Recalculate()
	p := CollisionParticipantFromEntity(e)
	c.True(p.HasHP, "the default sheet defines HP")
	c.True(p.HasST, "and ST")
	c.Equal(fxp.Twelve, p.HP, "the HP is the maximum of 12, not the 9 left after the injury")
	c.Equal(fxp.Ten, p.ST, "the ST is its own 10, not the HP")
	c.Equal(fxp.Five, p.Velocity, "the Move is suggested as the velocity")
	c.Equal(1, p.SM, "the SM is the sheet's")
	c.Equal(4, p.Acrobatics, "without the skill, Acrobatics defaults to DX-6")
	c.Equal(6, p.Swimming, "without the skill, Swimming defaults to HT-4")
	c.Equal(0, p.ArmorDR, "a bare torso has no armor DR")
	c.Equal(0, p.InnateDR, "nor any other")

	hauberk := gurpstest.AddCarriedEquipmentWithFeatures(e, "Mail Hauberk",
		gurpstest.NewDRBonus(fxp.Five, gurps.AllID, gurps.TorsoID))
	hauberk.BaseWeight = "30 lb"
	gurpstest.AddTraitWithFeatures(e, "Tough Skin", gurpstest.NewDRBonus(fxp.Two, gurps.AllID, gurps.TorsoID))
	e.Recalculate()
	p = CollisionParticipantFromEntity(e)
	c.Equal(5, p.ArmorDR, "the hauberk is armor")
	c.Equal(2, p.InnateDR, "the tough skin is not")
	c.Equal(fxp.Four, p.Velocity, "the 30 lb hauberk is a light load, which takes the Move down to 4")
}
