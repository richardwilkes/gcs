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

// BlastTarget is what a character sheet supplies about the target of an explosion or area attack.
type BlastTarget struct {
	HP      fxp.Int // The maximum HP; meaningful only when HasHP.
	SM      int
	TorsoDR int  // The total torso DR.
	DR      int  // The DR against the attack, averaged over the exposed locations per Large-Area Injury (BX400).
	HasHP   bool // Whether the sheet defines HP at all.
}

// BlastTargetFromEntity reads what the sheet's character supplies about the target of an attack with the given damage
// type, of which only the base type matters to the DR, and the given test for the locations facing it, as ExposedFilter
// returns. The entity must not be nil.
func BlastTargetFromEntity(entity *gurps.Entity, damageType string, exposed func(*gurps.HitLocation) bool) BlastTarget {
	t := BlastTarget{
		SM: entity.Profile.AdjustedSizeModifier(),
		DR: LargeAreaDR(entity, BaseDamageType(damageType), exposed),
	}
	if entity.ResolveAttribute(gurps.HitPointsID) != nil {
		t.HP = entity.Attributes.Maximum(gurps.HitPointsID).Max(0)
		t.HasHP = true
	}
	t.TorsoDR, _ = TorsoDR(entity)
	return t
}
