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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/check"
)

// stepLevelUpAndDown drives IncrementSkillLevel and DecrementSkillLevel through the expected point ladders, checking
// after each step that exactly one level was gained or lost and that the points landed on the expected value.
func stepLevelUpAndDown(t *testing.T, p SkillAdjustmentProvider, level func() Level, up, down []fxp.Int) {
	t.Helper()
	c := check.New(t)
	for _, expected := range up {
		before := level().Level
		p.IncrementSkillLevel()
		c.Equal(expected, p.RawPoints(), "points after incrementing to %v", expected)
		c.Equal(before+fxp.One, level().Level, "level after incrementing to %v points", expected)
	}
	for _, expected := range down {
		before := level().Level
		p.DecrementSkillLevel()
		c.Equal(expected, p.RawPoints(), "points after decrementing to %v", expected)
		c.Equal(before-fxp.One, level().Level, "level after decrementing to %v points", expected)
	}
}

func TestSkillLevelStepping(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Name = "Stepping"
	sk.Difficulty.Attribute = IntelligenceID
	sk.Difficulty.Difficulty = difficulty.Average
	e.Skills = append(e.Skills, sk)
	sk.SetRawPoints(fxp.One)
	cached := func() Level {
		c.Equal(sk.CalculateLevel(nil), sk.LevelData, "cached level must track the points")
		return sk.LevelData
	}
	stepLevelUpAndDown(t, sk, cached,
		[]fxp.Int{fxp.Two, fxp.Four, fxp.Eight, fxp.Twelve},
		[]fxp.Int{fxp.Eight, fxp.Four, fxp.Two, fxp.One})

	// Fractional points step from the whole-point boundary.
	sk.SetRawPoints(fxp.OneAndAHalf)
	sk.IncrementSkillLevel()
	c.Equal(fxp.Two, sk.Points)

	// Removing the last point leaves the skill with none.
	sk.SetRawPoints(fxp.One)
	sk.DecrementSkillLevel()
	c.Equal(fxp.Int(0), sk.Points)
	sk.DecrementSkillLevel()
	c.Equal(fxp.Int(0), sk.Points, "decrementing at zero points stays at zero")

	// Wildcard skills cost three times as much, so each level needs a wider search.
	sk.Difficulty.Difficulty = difficulty.Wildcard
	sk.SetRawPoints(fxp.Three)
	stepLevelUpAndDown(t, sk, cached,
		[]fxp.Int{fxp.Six, fxp.Twelve, fxp.TwentyFour, fxp.ThirtySix},
		[]fxp.Int{fxp.TwentyFour, fxp.Twelve, fxp.Six, fxp.Three})

	// Containers have no level of their own to adjust.
	container := NewSkill(e, nil, true)
	container.IncrementSkillLevel()
	c.Equal(fxp.Int(0), container.Points)

	// With no higher level to reach, as for a skill whose attribute is missing, the points are left alone.
	sk.Difficulty.Difficulty = difficulty.Average
	sk.Difficulty.Attribute = "missing"
	sk.SetRawPoints(fxp.Five)
	c.Equal(fxp.Min, sk.CalculateLevel(nil).Level, "precondition: the skill has no level")
	sk.IncrementSkillLevel()
	c.Equal(fxp.Five, sk.Points, "incrementing a skill with no higher level must leave its points alone")
}

func TestSpellLevelStepping(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sp := NewSpell(e, nil, false)
	sp.Name = "Stepping"
	e.Spells = append(e.Spells, sp)
	c.Equal(difficulty.Hard, sp.Difficulty.Difficulty)
	sp.SetRawPoints(fxp.One)
	cached := func() Level {
		c.Equal(sp.CalculateLevel(), sp.LevelData, "cached level must track the points")
		return sp.LevelData
	}
	stepLevelUpAndDown(t, sp, cached,
		[]fxp.Int{fxp.Two, fxp.Four, fxp.Eight, fxp.Twelve},
		[]fxp.Int{fxp.Eight, fxp.Four, fxp.Two, fxp.One})

	sp.DecrementSkillLevel()
	c.Equal(fxp.Int(0), sp.Points)

	container := NewSpell(e, nil, true)
	container.IncrementSkillLevel()
	c.Equal(fxp.Int(0), container.Points)
}

func TestSkillLevelSteppingPastKeptLevel(t *testing.T) {
	for _, tc := range []struct {
		name string
		// add puts on the sheet an item bought with 1 point and a trait whose grant sets the level 4 points buy, kept,
		// as the merge mode says, when higher or lower than the item's own, and returns the item along with what reads
		// its level.
		add  func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level)
		kept fxp.Int
	}{
		{name: "skill", add: func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level) {
			brawling := addSkill(e, "Brawling", fxp.One)
			addTraitWithFeatures(e, "Training",
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve, merge))
			return brawling, func() Level { return brawling.CalculateLevel(nil) }
		}, kept: fxp.FromInteger(12)},
		{name: "spell", add: func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			addTraitWithFeatures(e, "Magery",
				newSpellGrantOf(newHardSpell("Light", fxp.One), grantadjust.SetLevel, fxp.Ten, merge))
			return light, light.CalculateLevel
		}, kept: fxp.Ten},
	} {
		t.Run(tc.name+" kept highest", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepHighest)
			e.Recalculate()
			c.Equal(tc.kept, level().Level, "precondition: the grant holds the level above the one the point buys")
			// The next level above the kept one needs 8 points, in the second stride beyond the point bought, and
			// decrementing comes back to the fewest points that still show the kept level.
			stepLevelUpAndDown(t, p, level, []fxp.Int{fxp.Eight}, []fxp.Int{fxp.One})
			p.IncrementSkillLevel()
			c.Equal(fxp.Eight, p.RawPoints(), "the points decrementing took off come back")
		})
		t.Run(tc.name+" kept lowest", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepLowest)
			e.Recalculate()
			c.Equal(tc.kept-fxp.Two, level().Level, "precondition: the point bought gives a level below the one kept")
			// The levels up to the one kept are bought as without the grant, and none above it.
			stepLevelUpAndDown(t, p, level, []fxp.Int{fxp.Two, fxp.Four}, nil)
			p.IncrementSkillLevel()
			c.Equal(fxp.Four, p.RawPoints(), "the grant caps the level, so incrementing leaves the points alone")
			c.Equal(tc.kept, level().Level)
			stepLevelUpAndDown(t, p, level, nil, []fxp.Int{fxp.Two, fxp.One})
		})
	}
}

func TestDisabledSkillLevelStepping(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Name = "Stepping"
	sk.Difficulty.Attribute = IntelligenceID
	sk.Difficulty.Difficulty = difficulty.Average
	e.Skills = append(e.Skills, sk)
	sk.SetRawPoints(fxp.Eight)
	sk.Disabled = true
	sk.DecrementSkillLevel()
	c.Equal(fxp.Eight, sk.Points, "decrementing a disabled skill must leave its points alone")
	sk.IncrementSkillLevel()
	c.Equal(fxp.Eight, sk.Points, "incrementing a disabled skill must leave its points alone")

	// A disabled parent disables the skill, too.
	sk.Disabled = false
	container := NewSkill(e, nil, true)
	container.Disabled = true
	container.Children = append(container.Children, sk)
	sk.SetParent(container)
	sk.DecrementSkillLevel()
	c.Equal(fxp.Eight, sk.Points, "decrementing a skill under a disabled container must leave its points alone")

	sp := NewSpell(e, nil, false)
	sp.Name = "Stepping"
	e.Spells = append(e.Spells, sp)
	sp.SetRawPoints(fxp.Eight)
	sp.Disabled = true
	sp.DecrementSkillLevel()
	c.Equal(fxp.Eight, sp.Points, "decrementing a disabled spell must leave its points alone")
	sp.IncrementSkillLevel()
	c.Equal(fxp.Eight, sp.Points, "incrementing a disabled spell must leave its points alone")
}
