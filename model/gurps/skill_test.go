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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/toolbox/v2/xbytes"
)

// TestSkillOptionalSpecializationNotSynced verifies that the optional specialization is left out of the Source Sync
// mechanism entirely. It is a per-character choice layered on top of the library's skill, so a character that has
// picked one must still report as matching its source, and syncing with the source must never replace it with the
// library's value.
func TestSkillOptionalSpecializationNotSynced(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	newSkill := func(owner DataOwner, pageRef, optionalSpecialization string) *Skill {
		sk := NewSkill(owner, nil, false)
		sk.Name = "Ritual Magic"
		sk.Specialization = "Fire"
		sk.Difficulty.Attribute = IntelligenceID
		sk.PageRef = pageRef
		sk.OptionalSpecialization = optionalSpecialization
		return sk
	}

	// Stand in for the skill as it appears in a library file, and a character's copy of it that differs only in the
	// optional specialization the character chose. The library data is injected directly into the source matcher, which
	// is what a real sync gets from the library file it loaded.
	source := newSkill(nil, "B242", "Ember")
	local := newSkill(e, "B242", "Flame")
	e.Skills = append(e.Skills, local)
	libFile := LibraryFile{Library: "Test Library", Path: "Test" + SkillsExt}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	e.SourceMatcher().libHashes = map[LibraryFile]libSrcData{
		libFile: {dataHashes: map[tid.TID]HashAndData{source.TID: {Hash: Hash64(source), Data: source}}},
	}

	// Differing only in optional specialization is no difference at all as far as the source is concerned.
	c.Equal(Hash64(source), Hash64(local), "skills differing only in optional specialization must hash the same")
	state, _ := e.SourceMatcher().Match(local)
	c.Equal(srcstate.Matched, state, "a skill differing from its source only in optional specialization matches it")
	local.SyncWithSource()
	c.Equal("Flame", local.OptionalSpecialization, "syncing a matched skill leaves the optional specialization alone")

	// Once the local copy drifts on a field that *is* synced, the sync must bring that field back into line while still
	// leaving the optional specialization untouched.
	local.PageRef = ""
	state, _ = e.SourceMatcher().Match(local)
	c.Equal(srcstate.Mismatched, state, "precondition: a differing page reference is a mismatch")
	local.SyncWithSource()
	c.Equal("B242", local.PageRef, "the sync brings the source's page reference across")
	c.Equal("Flame", local.OptionalSpecialization, "the sync leaves the local optional specialization in place")
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
	c.NoError(jio.Unmarshal([]byte(`{"type":"technique","name":"Feint","difficulty":"h","points":2}`), &sk),
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
	c.NoError(jio.Unmarshal([]byte(`{"type":"skill","name":"Shortsword","difficulty":"dx/a","points":0,`+
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
