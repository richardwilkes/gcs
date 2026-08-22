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
	"strconv"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptSpellLevelWithoutEntity verifies that a spell resolved without an owning entity (as happens for library and
// template files) reports 0 for its level and relative level rather than leaking the fxp.Min sentinel that
// CalculateSpellLevel leaves behind.
func TestScriptSpellLevelWithoutEntity(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(nil, nil, false)
	spell.Name = "Ignite Fire"
	c.Equal(fxp.Min, spell.LevelData.Level) // precondition: the stored level really is the sentinel
	c.Equal("0", ResolveScript(nil, deferredNewScriptSpell(spell), "self.level"))
	c.Equal("0", ResolveScript(nil, deferredNewScriptSpell(spell), "self.relativeLevel"))
}

// TestScriptSpellLevelWithTooFewPoints verifies that a spell with fewer than one point, which CalculateSpellLevel marks
// uncomputable with the fxp.Min sentinel, reports 0 to scripts instead.
func TestScriptSpellLevelWithTooFewPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	spell := NewSpell(e, nil, false)
	spell.Name = "Ignite Fire"
	spell.Points = 0
	e.Spells = append(e.Spells, spell)
	e.Recalculate()
	c.Equal(fxp.Min, spell.LevelData.Level) // precondition: the stored level really is the sentinel
	c.Equal("0", ResolveScript(e, deferredNewScriptSpell(spell), "self.level"))
	c.Equal("0", ResolveScript(e, deferredNewScriptSpell(spell), "self.relativeLevel"))
}

// TestScriptSpellLevelWithPoints verifies that a normally-resolvable spell still reports its real level and relative
// level, i.e. that the sentinel guard doesn't disturb the usual case. A relative level may legitimately be negative,
// so it must not be clamped.
func TestScriptSpellLevelWithPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	spell := NewSpell(e, nil, false)
	spell.Name = "Ignite Fire"
	spell.Points = fxp.One
	e.Spells = append(e.Spells, spell)
	e.Recalculate()
	c.Equal(strconv.Itoa(spell.LevelData.Level.AsInteger[int]()),
		ResolveScript(e, deferredNewScriptSpell(spell), "self.level"))
	c.True(spell.LevelData.RelativeLevel < 0) // precondition: a Hard spell with 1 point is below its attribute
	c.Equal(strconv.Itoa(spell.LevelData.RelativeLevel.AsInteger[int]()),
		ResolveScript(e, deferredNewScriptSpell(spell), "self.relativeLevel"))
}

// TestScriptSpellTechLevelMatchesSkill verifies that a spell without a tech level hands scripts the same empty string
// the skill wrapper does, rather than undefined, so the common "no TL" idiom behaves the same for both node types.
func TestScriptSpellTechLevelMatchesSkill(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(nil, nil, false)
	skill := NewSkill(nil, nil, false)
	c.Equal("string", ResolveScript(nil, deferredNewScriptSpell(spell), "typeof self.techLevel"))
	c.Equal("", ResolveScript(nil, deferredNewScriptSpell(spell), "self.techLevel"))
	c.Equal(ResolveScript(nil, deferredNewScriptSkill(skill), "typeof self.techLevel"),
		ResolveScript(nil, deferredNewScriptSpell(spell), "typeof self.techLevel"))

	// A fresh spell is needed here because ResolveScript caches by (node ID, script text).
	withTL := NewSpell(nil, nil, false)
	tl := "3"
	withTL.TechLevel = &tl
	c.Equal("3", ResolveScript(nil, deferredNewScriptSpell(withTL), "self.techLevel"))
}

// TestScriptSpellDifficultyMatchesSkill verifies that the spell wrapper reports its difficulty in the same
// serialization form the skill wrapper uses, so case-sensitive JS comparisons work across both node types.
func TestScriptSpellDifficultyMatchesSkill(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(nil, nil, false)
	skill := NewSkill(nil, nil, false)
	skill.Difficulty.Difficulty = spell.Difficulty.Difficulty
	c.Equal(spell.Difficulty.Difficulty.Key(), ResolveScript(nil, deferredNewScriptSpell(spell), "self.difficulty"))
	c.Equal(ResolveScript(nil, deferredNewScriptSkill(skill), "self.difficulty"),
		ResolveScript(nil, deferredNewScriptSpell(spell), "self.difficulty"))
}
