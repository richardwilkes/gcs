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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// addDXSkill adds a DX-based skill of the given difficulty and points to the entity's skills and returns it. The
// entity is not recalculated.
func addDXSkill(e *gurps.Entity, name string, diff difficulty.Level, points fxp.Int) *gurps.Skill {
	sk := gurps.NewSkill(e, nil, false)
	sk.Name = name
	sk.Difficulty.Attribute = gurps.DexterityID
	sk.Difficulty.Difficulty = diff
	sk.Points = points
	e.Skills = append(e.Skills, sk)
	return sk
}

// addLeveledTrait adds a leveled trait with the given name and levels to the entity's traits and returns it. The
// entity is not recalculated.
func addLeveledTrait(e *gurps.Entity, name string, levels fxp.Int) *gurps.Trait {
	trait := gurpstest.AddTraitWithFeatures(e, name)
	trait.CanLevel = true
	trait.Levels = levels
	return trait
}

// addSpell adds a spell with the given name and points to the entity's spells and returns it. The entity is not
// recalculated.
func addSpell(e *gurps.Entity, name string, points fxp.Int) *gurps.Spell {
	sp := gurps.NewSpell(e, nil, false)
	sp.Name = name
	sp.Points = points
	e.Spells = append(e.Spells, sp)
	return sp
}

// TestSkillLevelOrDefault verifies that a skill on the sheet supplies its level, and that one that is not falls back
// to its default from the attribute.
func TestSkillLevelOrDefault(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	c.Equal(4, SkillLevelOrDefault(e, "Acrobatics", gurps.DexterityID, -6), "without the skill, the default of DX-6 applies")
	addDXSkill(e, "Acrobatics", difficulty.Hard, fxp.Four)
	e.Recalculate()
	c.Equal(10, SkillLevelOrDefault(e, "Acrobatics", gurps.DexterityID, -6), "with the skill, its level applies")
}

// TestTorsoDR verifies that the torso DR is split into the part that comes from armor and the rest.
func TestTorsoDR(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	total, armor := TorsoDR(e)
	c.Equal(0, total, "a bare torso has no DR")
	c.Equal(0, armor, "nor any armor")
	gurpstest.AddCarriedEquipmentWithFeatures(e, "Mail Hauberk", gurpstest.NewDRBonus(fxp.Five, gurps.AllID, gurps.TorsoID))
	gurpstest.AddTraitWithFeatures(e, "Tough Skin", gurpstest.NewDRBonus(fxp.Two, gurps.AllID, gurps.TorsoID))
	e.Recalculate()
	total, armor = TorsoDR(e)
	c.Equal(7, total, "the hauberk and the tough skin both count toward the total")
	c.Equal(5, armor, "only the hauberk counts as armor")
}

// TestBasicLiftAndThrustFor verifies that a nil entity is answered from the global default sheet settings.
func TestBasicLiftAndThrustFor(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.Weight(fxp.Twenty), BasicLiftFor(nil, fxp.Ten), "ST 10 has a Basic Lift of 20 under the default progression")
	c.Equal("1d-2", gurps.FormatDice(ThrustFor(nil, fxp.Ten), false), "ST 10 has a thrust of 1d-2 under the default progression")
	e := gurps.NewEntity()
	c.Equal(e.BasicLiftForST(fxp.Ten), BasicLiftFor(e, fxp.Ten), "an entity answers for itself")
}

// TestSpellLevel verifies that the highest level of the named spell is found, or 0 when it is not known.
func TestSpellLevel(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	c.Equal(0, spellLevel(e, "Recover Energy"), "an unknown spell has no level")
	sp := addSpell(e, "Recover Energy", fxp.FromInteger(28))
	e.Recalculate()
	want := sp.CalculateLevel().Level.AsInteger[int]()
	c.True(want > 0, "the spell must have a level")
	c.Equal(want, spellLevel(e, "recover energy"), "the name is matched regardless of case")
}
