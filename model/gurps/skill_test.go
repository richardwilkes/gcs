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
	"encoding/json/v2"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestSkillHashIncludesOptionalSpecialization verifies that the optional specialization participates in the sync hash.
// Without it, SourceMatcher cannot tell a library skill from a local copy that differs only in that field, so the skill
// reports as matching its source while SyncWithSource would happily overwrite the field.
func TestSkillHashIncludesOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	newSkill := func(optionalSpecialization string) *Skill {
		sk := NewSkill(nil, nil, false)
		sk.Name = "Ritual Magic"
		sk.Specialization = "Fire"
		sk.OptionalSpecialization = optionalSpecialization
		return sk
	}
	c.NotEqual(Hash64(newSkill("")), Hash64(newSkill("Flame")),
		"skills differing only in optional specialization must hash differently")
	c.NotEqual(Hash64(newSkill("Flame")), Hash64(newSkill("Ember")),
		"skills with differing optional specializations must hash differently")
	c.Equal(Hash64(newSkill("Flame")), Hash64(newSkill("Flame")),
		"skills with identical optional specializations must hash the same")
}

// TestSkillPrereqTooltipPunctuation verifies the unsatisfied-prereq description doesn't emit doubled commas when more
// than one of the specialization, optional specialization and tech level sections is present.
func TestSkillPrereqTooltipPunctuation(t *testing.T) {
	c := check.New(t)
	e := NewEntity() // No skills, so nothing can satisfy the prereq.
	newPrereq := func(specialization, optionalSpecialization string) *SkillPrereq {
		p := NewSkillPrereq()
		p.NameCriteria.Qualifier = "Broadsword"
		if specialization != "" {
			p.SpecializationCriteria.Compare = criteria.IsText
			p.SpecializationCriteria.Qualifier = specialization
		}
		if optionalSpecialization != "" {
			p.OptionalSpecializationCriteria.Compare = criteria.IsText
			p.OptionalSpecializationCriteria.Qualifier = optionalSpecialization
		}
		p.LevelCriteria.Qualifier = fxp.Twelve
		return p
	}
	describe := func(p *SkillPrereq, exclude any) string {
		var tooltip xbytes.InsertBuffer
		c.False(p.Satisfied(e, exclude, &tooltip, "", nil), "prereq must be unsatisfied for the tooltip to be written")
		return tooltip.String()
	}
	tl := "3"
	withTL := NewSkill(e, nil, false)
	withTL.TechLevel = &tl

	for i, one := range []struct {
		prereq   *SkillPrereq
		exclude  any
		expected string
	}{
		{
			prereq:   newPrereq("", ""),
			expected: `Has a skill whose name is "Broadsword" and level is at least 12`,
		},
		{
			prereq:   newPrereq("Fencing", ""),
			expected: `Has a skill whose name is "Broadsword", specialization is "Fencing" and level is at least 12`,
		},
		{
			prereq: newPrereq("Fencing", "Rapier"),
			expected: `Has a skill whose name is "Broadsword", specialization is "Fencing", optional specialization ` +
				`is "Rapier" and level is at least 12`,
		},
		{
			prereq:  newPrereq("Fencing", "Rapier"),
			exclude: withTL,
			expected: `Has a skill whose name is "Broadsword", specialization is "Fencing", optional specialization ` +
				`is "Rapier", level is at least 12 and tech level matches`,
		},
		{
			prereq:   newPrereq("", ""),
			exclude:  withTL,
			expected: `Has a skill whose name is "Broadsword" level is at least 12 and tech level matches`,
		},
	} {
		actual := describe(one.prereq, one.exclude)
		c.False(strings.Contains(actual, ",,"), "case %d must not contain a doubled comma: %q", i, actual)
		c.Equal(one.expected, actual, "case %d", i)
	}
}

// TestTechniqueWithoutDefaultDoesNotPanic verifies that a technique loaded from data that lacks the "default" field is
// handled gracefully rather than panicking. UnmarshalJSONFrom and CalculateTechniqueLevel already tolerate a nil
// TechniqueDefault, so the rest of the code must as well.
func TestTechniqueWithoutDefaultDoesNotPanic(t *testing.T) {
	c := check.New(t)
	var sk Skill
	c.NoError(json.Unmarshal([]byte(`{"type":"technique","name":"Feint","difficulty":"h","points":2}`), &sk),
		"a technique without a default should load")
	c.True(sk.IsTechnique(), "the loaded skill must be a technique")
	c.True(sk.TechniqueDefault == nil, "the loaded technique must have no default")

	e := NewEntity()
	e.Skills = append(e.Skills, &sk)
	sk.SetDataOwner(e)
	c.NotPanics(e.Recalculate, "recalculating an entity with a defaultless technique must not panic")
	c.Equal("", sk.UnsatisfiedReason, "a technique without a default has nothing to be unsatisfied about")

	var tooltip xbytes.InsertBuffer
	c.True(sk.TechniqueSatisfied(&tooltip, ""), "a technique without a default must be considered satisfied")
	c.Equal("", tooltip.String(), "no tooltip should be generated")
	c.NotPanics(func() { sk.ModifierNotes() }, "ModifierNotes must not panic")
	c.NotPanics(func() { sk.AdjustedRelativeLevel() }, "AdjustedRelativeLevel must not panic")
	c.NotPanics(func() { sk.CellData(SkillDescriptionColumn, &CellData{}) }, "CellData must not panic")
	c.NotPanics(func() {
		clone := sk.Clone(LibraryFile{}, e, nil, Reference)
		c.True(clone.IsTechnique(), "the clone must still be a technique")
		c.True(clone.TechniqueDefault == nil, "the clone must still have no default")
	}, "cloning a technique without a default must not panic")
}

// TestSkillWithNullDefaultEntry verifies that a JSON null in a skill's "defaults" array is dropped rather than
// dereferenced. Such an entry decodes into a nil pointer without error, and everything that walks the defaults --
// level resolution, hashing, nameable extraction -- reads the elements without a nil check, so a data file holding one
// crashed GCS instead of the entry simply being skipped.
func TestSkillWithNullDefaultEntry(t *testing.T) {
	c := check.New(t)
	var sk Skill
	c.NoError(json.Unmarshal([]byte(`{"type":"skill","name":"Shortsword","difficulty":"dx/a","points":0,`+
		`"defaults":[null,{"type":"skill","name":{"compare":"is","qualifier":"Broadsword"},"modifier":-2}]}`), &sk),
		"a skill with a null default entry should load")
	c.Equal(1, len(sk.Defaults), "the null entry was dropped and the usable one kept")

	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.One)
	broadsword.Difficulty.Attribute = DexterityID
	e.Skills = append(e.Skills, &sk)
	sk.SetDataOwner(e)
	c.NotPanics(e.Recalculate, "recalculating with a null default entry must not panic")

	// The surviving default still resolves, so the null entry cost nothing but itself.
	c.NotNil(sk.DefaultedFrom, "the usable default still resolves")
	c.Equal(broadsword.LevelData.Level-fxp.Two, sk.LevelData.Level, "Shortsword defaults to Broadsword-2")

	// The other walkers over the defaults must be equally safe.
	c.NotPanics(func() { Hash64(&sk) }, "hashing must not panic")
	c.NotPanics(func() { sk.FillWithNameableKeys(make(map[string]string), nil) },
		"nameable extraction must not panic")
	c.NotPanics(func() { sk.Clone(LibraryFile{}, e, nil, Reference) }, "cloning must not panic")

	// A nil that reaches the list some other way must be skipped rather than handed back to callers that dereference
	// it, since resolveToSpecificDefaults' own nil guard used to append it anyway.
	sk.Defaults = append([]*SkillDefault{nil}, sk.Defaults...)
	c.Equal(1, len(sk.resolveToSpecificDefaults()), "a nil default is not passed along")
	c.NotPanics(e.Recalculate, "recalculating still must not panic")
}
