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
