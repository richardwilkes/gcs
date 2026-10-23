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

// newUnlearnableSkill returns a skill attached to the given entity that has no points and no usable default, so its
// computed level is the fxp.Min sentinel.
func newUnlearnableSkill(e *Entity, name string) *Skill {
	sk := NewSkill(e, nil, false)
	sk.Name = name
	sk.Specialization = "None"
	sk.Points = 0
	e.Skills = append(e.Skills, sk)
	e.Recalculate()
	return sk
}

// entity.skillLevel() applies the same sentinel guard as the skill wrapper when it matches an unlearnable skill.
func TestScriptEntitySkillLevelSentinel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newUnlearnableSkill(e, "Unlearnable Via Entity")
	c.Equal("0", ResolveScript(e, ScriptSelfProvider{},
		`entity.skillLevel("Unlearnable Via Entity", "None", false, "")`))
	c.Equal("0", ResolveScript(e, ScriptSelfProvider{},
		`entity.skillLevel("Unlearnable Via Entity", "None", true, "")`))
}

// A skill whose level cannot be computed (no points and no usable default) reports 0 to scripts rather than leaking the
// fxp.Min sentinel as -922337203685477, while a normally-resolvable skill still reports its real level and relative
// level, so the sentinel guard doesn't disturb the usual case. A relative level may legitimately be negative, so it
// must not be clamped.
func TestScriptSkillLevelWithPoints(t *testing.T) {
	c := check.New(t)
	for _, tc := range []struct {
		name     string
		points   fxp.Int
		sentinel bool
	}{
		{name: "Learnable", points: fxp.One},
		{name: "Unlearnable", points: 0, sentinel: true},
	} {
		e := NewEntity()
		sk := NewSkill(e, nil, false)
		sk.Name = tc.name
		sk.Specialization = "None"
		sk.Points = tc.points
		e.Skills = append(e.Skills, sk)
		e.Recalculate()
		level, relativeLevel := "0", "0"
		if tc.sentinel {
			c.Equal(fxp.Min, sk.LevelData.Level, "%s: precondition: the stored level really is the sentinel", tc.name)
		} else {
			c.NotEqual(fxp.Min, sk.LevelData.Level, tc.name)
			c.True(sk.LevelData.RelativeLevel < 0,
				"%s: precondition: an Average skill with 1 point is below its attribute", tc.name)
			level = strconv.Itoa(sk.LevelData.Level.AsInteger[int]())
			relativeLevel = strconv.Itoa(sk.LevelData.RelativeLevel.AsInteger[int]())
		}
		c.Equal(level, ResolveScript(e, deferredNewScriptSkill(sk), "self.level"), tc.name)
		c.Equal(relativeLevel, ResolveScript(e, deferredNewScriptSkill(sk), "self.relativeLevel"), tc.name)
	}
}

// A panic while recalculating a skill's level for a script must not leave the skill-level resolution exclusion
// registered on the entity. Leaving it registered would freeze the skill's script-visible level and produce spurious
// "attempt to resolve skill level via itself" errors until the next full Recalculate.
func TestScriptSkillLevelExclusionReleasedOnPanic(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Name = "Exploding"
	sk.Specialization = "None"
	sk.Points = fxp.One
	e.Skills = append(e.Skills, sk)
	e.Recalculate()

	// A nil default makes UpdateLevel panic when it tries to determine the best default to use.
	sk.Defaults = []*SkillDefault{nil}

	// ResolveScript recovers the panic and reports it as the script's result, so the read completes but the exclusion
	// registered around UpdateLevel must have been removed by the deferred unregister.
	c.NotEqual("0", ResolveScript(e, deferredNewScriptSkill(sk), "self.level"))
	c.Equal(0, len(e.skillResolverExclusions))
}
