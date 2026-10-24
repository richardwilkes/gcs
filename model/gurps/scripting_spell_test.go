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

// A spell whose level CalculateSpellLevel cannot compute, because it has no owning entity (as happens for library and
// template files) or fewer than one point, reports 0 to scripts for its level and relative level rather than leaking
// the fxp.Min sentinel left behind. A normally-resolvable spell still reports its real level and relative level, so
// the sentinel guard doesn't disturb the usual case. A relative level may legitimately be negative, so it must not be
// clamped.
func TestScriptSpellLevelWithPoints(t *testing.T) {
	for _, tc := range []struct {
		name     string
		owned    bool
		points   fxp.Int
		sentinel bool
	}{
		{name: "without entity", sentinel: true},
		{name: "too few points", owned: true, points: 0, sentinel: true},
		{name: "with points", owned: true, points: fxp.One},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			var e *Entity
			var owner DataOwner
			if tc.owned {
				e = NewEntity()
				owner = e
			}
			spell := NewSpell(owner, nil, false)
			spell.Name = "Ignite Fire"
			if tc.owned {
				spell.Points = tc.points
				e.Spells = append(e.Spells, spell)
				e.Recalculate()
			}
			if tc.sentinel {
				c.Equal(fxp.Min, spell.LevelData.Level, "precondition: the stored level really is the sentinel")
				c.Equal("0", ResolveScript(e, deferredNewScriptSpell(spell), "self.level"))
				c.Equal("0", ResolveScript(e, deferredNewScriptSpell(spell), "self.relativeLevel"))
				return
			}
			c.Equal(strconv.Itoa(spell.LevelData.Level.AsInteger[int]()),
				ResolveScript(e, deferredNewScriptSpell(spell), "self.level"))
			c.True(spell.LevelData.RelativeLevel < 0, "precondition: a Hard spell with 1 point is below its attribute")
			c.Equal(strconv.Itoa(spell.LevelData.RelativeLevel.AsInteger[int]()),
				ResolveScript(e, deferredNewScriptSpell(spell), "self.relativeLevel"))
		})
	}
}

// A spell without a tech level hands scripts the same empty string the skill wrapper does, rather than undefined, so
// the common "no TL" idiom behaves the same for both node types.
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

// The spell wrapper reports its difficulty in the same serialization form the skill wrapper uses, so case-sensitive JS
// comparisons work across both node types.
func TestScriptSpellDifficultyMatchesSkill(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(nil, nil, false)
	skill := NewSkill(nil, nil, false)
	skill.Difficulty.Difficulty = spell.Difficulty.Difficulty
	c.Equal(spell.Difficulty.Difficulty.Key(), ResolveScript(nil, deferredNewScriptSpell(spell), "self.difficulty"))
	c.Equal(ResolveScript(nil, deferredNewScriptSkill(skill), "self.difficulty"),
		ResolveScript(nil, deferredNewScriptSpell(spell), "self.difficulty"))
}
