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

// TestBlastTargetFromEntity verifies what a sheet supplies about the target of an explosion: its maximum HP, its SM,
// its torso DR and the Large-Area Injury DR against the base damage type over the exposed locations.
func TestBlastTargetFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	gurpstest.AddCarriedEquipmentWithFeatures(e, "Mail Hauberk", gurpstest.NewDRBonus(fxp.Five, gurps.AllID, gurps.TorsoID))
	gurpstest.AddCarriedEquipmentWithFeatures(e, "Fire Cloak", gurpstest.NewDRBonus(fxp.Four, "burn", gurps.TorsoID))
	e.Recalculate()

	target := BlastTargetFromEntity(e, "cr ex", nil)
	c.True(target.HasHP, "the default sheet defines HP")
	c.Equal(e.Attributes.Maximum(gurps.HitPointsID), target.HP, "the HP is the maximum")
	c.Equal(e.Profile.AdjustedSizeModifier(), target.SM, "the SM is the adjusted one")
	c.Equal(5, target.TorsoDR, "the torso DR is the general DR")
	c.Equal(3, target.DR, "with everything exposed, the torso's 5 and a bare location's 0 average to 3, rounded up")

	torsoOnly := ExposedFilter(e.SheetSettings.BodyType.LookupLocationByID(e, gurps.TorsoID), false)
	c.Equal(9, BlastTargetFromEntity(e, "burn ex", torsoOnly).DR,
		"only the base type is looked up, so the burn-specialized DR counts for a burning explosion")
	c.Equal(5, BlastTargetFromEntity(e, "cr ex", torsoOnly).DR, "but not for a crushing one")
}
