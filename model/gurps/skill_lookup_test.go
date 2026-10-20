// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps_test

import (
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// referenceSkillNamed is Entity.SkillNamed as it was before its tests were reordered to put the cheap ones first.
func referenceSkillNamed(e *gurps.Entity, name, specialization string, requirePoints bool, excludes map[string]bool) []*gurps.Skill {
	var list []*gurps.Skill
	gurps.Traverse(func(sk *gurps.Skill) bool {
		if !excludes[sk.String()] {
			if !requirePoints || sk.IsTechnique() || sk.AdjustedPoints(nil) > 0 {
				if strings.EqualFold(sk.NameWithReplacements(), name) {
					if specialization == "" || strings.EqualFold(sk.SpecializationWithReplacements(), specialization) ||
						strings.EqualFold(sk.OptionalSpecializationWithReplacements(), specialization) {
						list = append(list, sk)
					}
				}
			}
		}
		return false
	}, true, true, e.Skills...)
	return list
}

// referenceSkillMatching is Entity.SkillMatching as it was before its tests were reordered and its criteria resolved
// once per walk.
func referenceSkillMatching(e *gurps.Entity, nameCriteria, specializationCriteria criteria.Text, replacements map[string]string, requirePoints bool, excludes map[string]bool) []*gurps.Skill {
	specializationMatches := func(required, optional string) bool {
		if required != "" {
			return specializationCriteria.Matches(replacements, required) ||
				(optional != "" && specializationCriteria.Matches(replacements, optional))
		}
		return specializationCriteria.Matches(replacements, optional)
	}
	var list []*gurps.Skill
	gurps.Traverse(func(sk *gurps.Skill) bool {
		if !excludes[sk.String()] {
			if !requirePoints || sk.IsTechnique() || sk.AdjustedPoints(nil) > 0 {
				if nameCriteria.Matches(replacements, sk.NameWithReplacements()) &&
					specializationMatches(sk.SpecializationWithReplacements(), sk.OptionalSpecializationWithReplacements()) {
					list = append(list, sk)
				}
			}
		}
		return false
	}, true, true, e.Skills...)
	return list
}

// Reordering the tests a skill lookup makes must not change which skills it finds, nor their order, for any name,
// specialization, points requirement or exclusion a sheet full of skills can present.
func TestSkillLookupOrderingEquivalence(t *testing.T) {
	c := check.New(t)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.1)
	e := gurpstest.NewLargeEntity(&size)
	var skills []*gurps.Skill
	gurps.Traverse(func(sk *gurps.Skill) bool {
		skills = append(skills, sk)
		return false
	}, false, true, e.Skills...)
	c.True(len(skills) > 0)
	// Give a few skills an optional specialization, so that path is exercised too.
	for i, sk := range skills {
		if i%7 == 0 && sk.Specialization == "" {
			sk.OptionalSpecialization = "Optional"
		}
	}
	e.Recalculate()
	excludeSets := []map[string]bool{nil, {}, {skills[0].String(): true, skills[len(skills)/2].String(): true}}
	for _, sk := range skills {
		for _, specialization := range []string{
			"", sk.SpecializationWithReplacements(),
			sk.OptionalSpecializationWithReplacements(), "No Such Specialization",
		} {
			for _, requirePoints := range []bool{false, true} {
				for _, excludes := range excludeSets {
					c.Equal(referenceSkillNamed(e, sk.NameWithReplacements(), specialization, requirePoints, excludes),
						e.SkillNamed(sk.NameWithReplacements(), specialization, requirePoints, excludes),
						"SkillNamed(%q, %q, %v, %v)", sk.NameWithReplacements(), specialization, requirePoints, excludes)
				}
			}
		}
	}
	replacements := map[string]string{"Specialty": "Fire", "Name": "Broadsword"}
	names := []criteria.Text{
		{Compare: criteria.AnyText},
		{Compare: criteria.IsText, Qualifier: "Ritual Magic"},
		{Compare: criteria.IsText, Qualifier: "@Name@"},
		{Compare: criteria.ContainsText, Qualifier: "ritual"},
		{Compare: criteria.StartsWithText, Qualifier: "Guns"},
		{Compare: criteria.IsNotText, Qualifier: "Broadsword"},
	}
	specializations := []criteria.Text{
		{Compare: criteria.AnyText},
		{Compare: criteria.IsText},
		{Compare: criteria.IsText, Qualifier: "Fire"},
		{Compare: criteria.IsText, Qualifier: "@Specialty@"},
		{Compare: criteria.IsText, Qualifier: "Optional"},
		{Compare: criteria.ContainsText, Qualifier: "i"},
		{Compare: criteria.IsNotText, Qualifier: "Pistol"},
	}
	for _, name := range names {
		for _, specialization := range specializations {
			for _, requirePoints := range []bool{false, true} {
				for _, excludes := range excludeSets {
					c.Equal(referenceSkillMatching(e, name, specialization, replacements, requirePoints, excludes),
						e.SkillMatching(name, specialization, replacements, requirePoints, excludes),
						"SkillMatching(%v, %v, %v, %v)", name, specialization, requirePoints, excludes)
				}
			}
		}
	}
}
