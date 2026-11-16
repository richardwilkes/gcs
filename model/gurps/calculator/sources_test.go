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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/progression"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/rpgtools/dice"
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
// to its default from the attribute, or to 0 when the sheet lacks that attribute.
func TestSkillLevelOrDefault(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	c.Equal(4, SkillLevelOrDefault(e, "Acrobatics", gurps.DexterityID, -6), "without the skill, the default of DX-6 applies")
	c.Equal(0, SkillLevelOrDefault(e, "Acrobatics", "nope", -6), "a default from an attribute the sheet lacks is 0")
	addDXSkill(e, "Acrobatics", difficulty.Hard, fxp.Four)
	e.Recalculate()
	c.Equal(10, SkillLevelOrDefault(e, "Acrobatics", gurps.DexterityID, -6), "with the skill, its level applies")
}

// TestBasicLiftAndThrustFor verifies that an entity is answered under its own damage progression and a nil one under
// the global default sheet settings.
func TestBasicLiftAndThrustFor(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.Weight(fxp.Eighty), BasicLiftFor(nil, fxp.Twenty), "ST 20 has a Basic Lift of 80 under the default progression")
	c.Equal("2d-1", gurps.FormatDice(ThrustFor(nil, fxp.Twenty), false), "and a thrust of 2d-1")
	e := gurps.NewEntity()
	e.SheetSettings.DamageProgression = progression.KnowingYourOwnStrength
	c.Equal(fxp.Weight(fxp.FromInteger(200)), BasicLiftFor(e, fxp.Twenty),
		"under Knowing Your Own Strength, the entity's ST 20 has a Basic Lift of 200")
	c.Equal("3d", gurps.FormatDice(ThrustFor(e, fxp.Twenty), false), "and a thrust of 3d")
}

// TestTorsoDR verifies that a body without a torso, or a sheet without a body, has no torso DR. The split between armor
// and the rest is pinned through CollisionParticipantFromEntity.
func TestTorsoDR(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	arm := gurps.NewHitLocation(e, "")
	arm.LocID = "arm"
	arm.ChoiceName = "Arm"
	arm.TableName = "Arm"
	arm.Slots = 1
	arm.DRBonus = 4
	body := &gurps.Body{Roll: dice.Dice{Count: 3, Sides: 6}, Locations: []*gurps.HitLocation{arm}}
	body.Update(e)
	e.SheetSettings.BodyType = body
	e.Recalculate()
	total, armor := TorsoDR(e)
	c.Equal(0, total, "an armored arm does not stand in for a missing torso")
	c.Equal(0, armor, "nor does its armor")

	e.SheetSettings.BodyType = nil
	total, armor = TorsoDR(e)
	c.Equal(0, total, "without a body there is no torso DR")
	c.Equal(0, armor, "nor any armor")
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
