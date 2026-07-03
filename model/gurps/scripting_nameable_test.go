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
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestScriptSpellCollegeUsesReplacements verifies that a spell's college list is exposed to scripts with any nameable
// replacements applied, matching every other text field on the script spell object. It previously handed scripts the
// raw stored college (containing @placeholders@).
func TestScriptSpellCollegeUsesReplacements(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(nil, nil, false)
	spell.Name = "Create @element@"
	spell.College = CollegeList{"@element@ College", "Meta"}
	spell.Replacements = map[string]string{"element": "Fire"}
	c.Equal("Fire College|Meta", ResolveScript(nil, deferredNewScriptSpell(spell), "self.college.join('|')"))
}

// TestScriptSkillLevelResolutionExclusionUsesResolvedName verifies that reading a skill script object's level keys the
// self-reference recursion guard by the resolved (replacement-applied) name, so it agrees with the entity.skillLevel
// path (which keys by the resolved name). Previously the skill path keyed by the raw name, so for a nameable skill the
// two guards used different keys and the recursion guard was defeated.
func TestScriptSkillLevelResolutionExclusionUsesResolvedName(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sk := NewSkill(e, nil, false)
	sk.Name = "Guns (@type@)"
	sk.Specialization = "@type@"
	sk.Replacements = map[string]string{"type": "Rifle"}
	sk.Points = fxp.One
	e.Skills = append(e.Skills, sk)
	e.Recalculate()

	// Simulate a resolution already in progress that was started via the entity path (entity.skillLevel), which keys
	// the guard by the resolved name.
	name := sk.NameWithReplacements()
	specialization := sk.SpecializationWithReplacements()
	optionalSpecialization := sk.OptionalSpecializationWithReplacements()
	e.registerSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)
	defer e.unregisterSkillLevelResolutionExclusion(name, specialization, optionalSpecialization)

	// Because the guard is already registered under the resolved name, reading the skill's level via a script must
	// detect the exclusion and skip the recursive UpdateLevel, returning the already-stored level. The sentinel proves
	// UpdateLevel was skipped; without the fix the raw-name check would miss the exclusion and recompute over it.
	sk.LevelData.Level = fxp.FromInteger(999)
	c.Equal("999", ResolveScript(e, deferredNewScriptSkill(sk), "self.level"))
}
