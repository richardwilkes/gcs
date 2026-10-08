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
// and ST, its Move as the suggested velocity, its skills or their defaults, and its torso DR split between armor and
// the rest.
func TestCollisionParticipantFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	p := CollisionParticipantFromEntity(e)
	c.True(p.HasHP, "the default sheet defines HP")
	c.True(p.HasST, "and ST")
	c.Equal(e.Attributes.Maximum(gurps.HitPointsID), p.HP, "the HP is the maximum")
	c.Equal(e.Attributes.Maximum(gurps.StrengthID), p.ST, "as is the ST")
	c.Equal(fxp.FromInteger(e.Move(e.EncumbranceLevel(false))), p.Velocity, "the Move is suggested as the velocity")
	c.Equal(e.Profile.AdjustedSizeModifier(), p.SM, "the SM is the adjusted one")
	c.Equal(4, p.Acrobatics, "without the skill, Acrobatics defaults to DX-6")
	c.Equal(6, p.Swimming, "without the skill, Swimming defaults to HT-4")
	c.Equal(0, p.ArmorDR, "a bare torso has no armor DR")
	c.Equal(0, p.InnateDR, "nor any other")

	gurpstest.AddCarriedEquipmentWithFeatures(e, "Mail Hauberk", gurpstest.NewDRBonus(fxp.Five, gurps.AllID, gurps.TorsoID))
	gurpstest.AddTraitWithFeatures(e, "Tough Skin", gurpstest.NewDRBonus(fxp.Two, gurps.AllID, gurps.TorsoID))
	e.Recalculate()
	p = CollisionParticipantFromEntity(e)
	c.Equal(5, p.ArmorDR, "the hauberk is armor")
	c.Equal(2, p.InnateDR, "the tough skin is not")
}
