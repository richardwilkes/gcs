// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestRecalculateSettlesDefaultDrivenSkillLevel verifies that a recalculation leaves the stored level of a skill whose
// level is driven by its default in step with the level computed on demand, along with the level of a weapon that
// defaults to that skill. The stored level used to be computed against the default chosen by the previous pass, so
// when a leveled trait raised the attribute the default points at, the stored level -- and so the weapon -- showed the
// state before the change, and only caught up on the next edit.
func TestRecalculateSettlesDefaultDrivenSkillLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// An IQ-based skill with no points, whose level therefore comes entirely from its DX default.
	sk := addTestSkill(e, "Melee Weapon", "", "", 0)
	sk.Defaults = []*SkillDefault{{DefaultType: DexterityID}}

	// A leveled trait that raises DX by one per level.
	bonus := NewAttributeBonus(DexterityID)
	bonus.PerLevel = true
	trait := addTraitWithFeatures(e, "Melee Combat", bonus)
	trait.CanLevel = true

	// A weapon that defaults to the skill.
	eqp := addCarriedEquipmentWithFeatures(e, "Super Sledge")
	w := NewWeapon(eqp, true)
	w.Defaults = []*SkillDefault{{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Melee Weapon"}},
	}}
	eqp.Weapons = []*Weapon{w}
	e.Recalculate()

	dx := e.ResolveAttributeCurrent(DexterityID)
	c.Equal(dx, sk.LevelData.Level, "the skill starts at its DX default")
	c.Equal(dx, w.SkillLevel(nil), "the weapon starts at the skill's level")

	for _, levels := range []fxp.Int{fxp.One, fxp.Two, fxp.Three, fxp.Two, fxp.One, 0} {
		trait.Levels = levels
		e.Recalculate()
		want := dx + levels
		c.Equal(want, sk.CalculateLevel(nil).Level, "levels %v: the level computed on demand follows DX", levels)
		c.Equal(want, sk.LevelData.Level, "levels %v: the stored level must not lag behind", levels)
		c.Equal(want, w.SkillLevel(nil), "levels %v: the weapon's level must not lag behind", levels)
		c.False(sk.UpdateLevel(), "levels %v: a further update must find nothing left to change", levels)
	}
}
