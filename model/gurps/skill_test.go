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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantadjust"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/check"
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
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)

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
			expected: `Has skill Broadsword at level at least 12`,
		},
		{
			prereq:   newPrereq("Fencing", ""),
			expected: `Has skill Broadsword (Fencing) at level at least 12`,
		},
		{
			prereq:   newPrereq("Fencing", "Rapier"),
			expected: `Has skill Broadsword (Fencing) with an optional specialization that is Rapier at level at least 12`,
		},
		{
			prereq:  newPrereq("Fencing", "Rapier"),
			exclude: withTL,
			expected: `Has skill Broadsword (Fencing) with an optional specialization that is Rapier at level at least ` +
				`12 with a matching tech level`,
		},
		{
			prereq:   newPrereq("", ""),
			exclude:  withTL,
			expected: `Has skill Broadsword at level at least 12 with a matching tech level`,
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
// dereferenced. Such an entry decodes into a nil pointer without error, and the walkers over the defaults -- level
// resolution, hashing, nameable extraction -- dereferenced it, so a data file holding one crashed GCS instead of the
// entry simply being skipped.
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

// TestSkillMarshalCalc verifies the "calc" object Skill writes through the marshaling shared with Spell: a container
// with nothing to report omits it, a skill without a resolvable level records only the reason and resolved notes, a
// leveled skill adds the level and relative skill level, and the data fields keep preceding "calc". A spell records its
// reason in both branches as well.
func TestSkillMarshalCalc(t *testing.T) {
	c := check.New(t)
	marshaled := func(s any) string {
		data, err := jio.Marshal(s)
		c.NoError(err)
		return string(data)
	}

	container := NewSkill(nil, nil, true)
	container.Name = "Melee"
	c.False(strings.Contains(marshaled(container), `"calc"`), "a container with nothing to report has no calc")

	noLevel := NewSkill(nil, nil, false)
	noLevel.Name = "Shortsword"
	noLevel.LocalNotes = "Favors @weapon@"
	noLevel.Replacements = map[string]string{"weapon": "the sabre"}
	c.True(noLevel.LevelData.Level <= 0, "precondition: the skill must have no resolvable level")
	c.False(strings.Contains(marshaled(noLevel), `"unsatisfied_reason"`), "no reason means no field")
	noLevel.UnsatisfiedReason = "Requires Broadsword"
	out := marshaled(noLevel)
	c.True(strings.Contains(out, `"calc":{"resolved_notes":"Favors the sabre","unsatisfied_reason":"Requires Broadsword"}`),
		"an unleveled skill records the resolved notes and reason, and nothing else: %s", out)
	c.True(strings.Index(out, `"name"`) < strings.Index(out, `"calc"`), "the data fields precede calc: %s", out)
	var restored Skill
	c.NoError(jio.Unmarshal([]byte(out), &restored))
	c.Equal(noLevel.Name, restored.Name, "the data fields survive a round-trip")
	c.Equal(noLevel.Replacements, restored.Replacements, "the replacements survive a round-trip")
	noLevel.Replacements = nil
	noLevel.LocalNotes = "Plain"
	c.False(strings.Contains(marshaled(noLevel), `"resolved_notes"`), "notes that resolve to themselves are not repeated")

	e := NewEntity()
	leveled := addTestSkill(e, "Shortsword", "", "", fxp.Four)
	e.Recalculate()
	c.True(leveled.LevelData.Level > 0, "precondition: the skill must have a positive level")
	leveled.UnsatisfiedReason = "Requires Broadsword"
	out = marshaled(leveled)
	c.True(strings.Contains(out, `"calc":{"level":`+leveled.LevelData.Level.String()+`,"rsl":"`+leveled.RelativeLevel()+
		`","unsatisfied_reason":"Requires Broadsword"}`), "a leveled skill records its level, rsl and reason: %s", out)

	const reason = "Requires Magery 1"
	noLevelSpell := NewSpell(nil, nil, false)
	noLevelSpell.Name = "Fireball"
	c.True(noLevelSpell.LevelData.Level <= 0, "precondition: the spell must have no resolvable level")
	c.False(strings.Contains(marshaled(noLevelSpell), "unsatisfied_reason"), "no reason means no field for a spell")
	noLevelSpell.UnsatisfiedReason = reason
	c.True(strings.Contains(marshaled(noLevelSpell), `"unsatisfied_reason":"`+reason+`"`),
		"an unsatisfied spell with no resolvable level must still record the reason")
	leveledSpell := addTestSpell(e, "Fireball", fxp.Four)
	e.Recalculate()
	c.True(leveledSpell.LevelData.Level > 0, "precondition: the spell must have a positive level")
	leveledSpell.UnsatisfiedReason = reason
	c.True(strings.Contains(marshaled(leveledSpell), `"unsatisfied_reason":"`+reason+`"`),
		"an unsatisfied spell with a level must record the reason")
}

// TestSkillPrereqIgnoresSkillWithNoPoints verifies that a skill with no points spent on it doesn't satisfy a skill
// prereq, even when it has a level from a default, while a purchased skill does.
func TestSkillPrereqIgnoresSkillWithNoPoints(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", 0)
	shortsword := addTestSkill(e, "Shortsword", "", "", 0)
	shortsword.Defaults = []*SkillDefault{{DefaultType: IntelligenceID, Modifier: -fxp.Five}}
	axe := addTestSkill(e, "Axe/Mace", "", "", fxp.One)
	e.Recalculate()
	c.Equal("-", broadsword.LevelData.LevelAsString(false), "precondition: Broadsword has no level")
	c.NotEqual("-", shortsword.LevelData.LevelAsString(false), "precondition: Shortsword has a level from its default")
	c.NotEqual("-", axe.LevelData.LevelAsString(false), "precondition: Axe/Mace has a level from its points")

	for _, compare := range []criteria.NumericComparison{criteria.AnyNumber, criteria.AtLeastNumber} {
		for _, one := range []struct {
			skill  *Skill
			counts bool
		}{
			{skill: broadsword, counts: false},
			{skill: shortsword, counts: false},
			{skill: axe, counts: true},
		} {
			p := NewSkillPrereq()
			p.LevelCriteria.Compare = compare // Qualifier is 0, so "at least 0"
			p.NameCriteria.Qualifier = one.skill.Name
			c.Equal(one.counts, p.Satisfied(e, nil, nil, "", nil), "%s: has %s", compare, one.skill.Name)
			p.Has = false
			c.Equal(!one.counts, p.Satisfied(e, nil, nil, "", nil), "%s: doesn't have %s", compare, one.skill.Name)
		}
	}
}

func TestDisabledSkillsAndSpellsShowNoLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	_, skill, spell, _ := newSwitchableItemSet(e)
	skill.Points = fxp.Four
	spell.Points = fxp.Four
	e.Recalculate()
	var data CellData
	skill.CellData(SkillLevelColumn, &data)
	c.NotEqual("-", data.Primary)
	spell.CellData(SpellLevelColumn, &data)
	c.NotEqual("-", data.Primary)

	skill.SetDisabled(true)
	spell.SetDisabled(true)
	e.Recalculate()
	data = CellData{}
	skill.CellData(SkillLevelColumn, &data)
	c.Equal("-", data.Primary)
	c.Equal("", data.Tooltip)
	data = CellData{}
	spell.CellData(SpellLevelColumn, &data)
	c.Equal("-", data.Primary)
	c.Equal("", data.Tooltip)
}

func TestDisabledSkillsAndSpellsCostNothing(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	skills := NewSkill(e, nil, true)
	brawling := NewSkill(e, skills, false)
	brawling.Points = fxp.Four
	karate := NewSkill(e, skills, false)
	karate.Points = fxp.FromInteger(8)
	skills.Children = []*Skill{brawling, karate}
	e.Skills = append(e.Skills, skills)
	spells := NewSpell(e, nil, true)
	light := NewSpell(e, spells, false)
	light.Points = fxp.One
	fireball := NewSpell(e, spells, false)
	fireball.Points = fxp.Two
	spells.Children = []*Spell{light, fireball}
	e.Spells = append(e.Spells, spells)
	e.Recalculate()
	c.Equal(fxp.FromInteger(12), skills.AdjustedPoints(nil))
	c.Equal(fxp.Three, spells.AdjustedPoints(nil))

	karate.SetDisabled(true)
	fireball.SetDisabled(true)
	e.Recalculate()
	c.Equal(fxp.Int(0), karate.AdjustedPoints(nil))
	c.Equal(fxp.Int(0), fireball.AdjustedPoints(nil))
	c.Equal(fxp.Four, skills.AdjustedPoints(nil))
	c.Equal(fxp.One, spells.AdjustedPoints(nil))
	c.Equal(NumericRangeOf(fxp.Four), skills.PointsRange(nil))
	c.Equal(e.PointsBreakdown().Skills, skills.AdjustedPoints(nil))
	c.Equal(e.PointsBreakdown().Spells, spells.AdjustedPoints(nil))
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

func TestSkillLevelSteppingPastKeptLevel(t *testing.T) {
	for _, tc := range []struct {
		name string
		// add puts on the sheet an item bought with 1 point and a trait whose grant sets the level 4 points buy, kept,
		// as the merge mode says, when higher or lower than the item's own, and returns the item along with what reads
		// its level.
		add func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level)
		// curse puts on the sheet a trait whose grant takes 2 levels from the item.
		curse func(e *Entity)
		// bless puts on the sheet an item with no points and a trait whose grant adds 2 levels to it, and returns the
		// item along with what reads its level. It is nil for a technique, which has a level of its own with no points,
		// at its default, so that the subtests of an item with none leave it out.
		bless func(e *Entity) (SkillAdjustmentProvider, func() Level)
		kept  fxp.Int
		// next is the fewest points that buy the level above the one kept, and far those that buy the level 3 above
		// it, which the curse brings down to the one above it. ladder holds the points that buy each level from the one
		// above the level the point bought gives up to the one kept.
		next, far fxp.Int
		ladder    []fxp.Int
	}{
		{name: "skill", add: func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level) {
			brawling := addSkill(e, "Brawling", fxp.One)
			addTraitWithFeatures(e, "Training",
				newSkillGrantOf(newEasySkill("Brawling", fxp.One), grantadjust.SetLevel, fxp.Twelve, merge))
			return brawling, func() Level { return brawling.CalculateLevel(nil) }
		}, curse: func(e *Entity) {
			addTraitWithFeatures(e, "Curse",
				newSkillGrant(newEasySkill("Brawling", fxp.One), -fxp.Two, grantmerge.Ignore))
		}, bless: func(e *Entity) (SkillAdjustmentProvider, func() Level) {
			brawling := addSkill(e, "Brawling", 0)
			addTraitWithFeatures(e, "Blessing",
				newSkillGrant(newEasySkill("Brawling", fxp.One), fxp.Two, grantmerge.Ignore))
			return brawling, func() Level { return brawling.CalculateLevel(nil) }
		}, kept: fxp.FromInteger(12), next: fxp.Eight, far: fxp.Sixteen, ladder: []fxp.Int{fxp.Two, fxp.Four}},
		{name: "spell", add: func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level) {
			light := newHardSpell("Light", fxp.One)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			addTraitWithFeatures(e, "Magery",
				newSpellGrantOf(newHardSpell("Light", fxp.One), grantadjust.SetLevel, fxp.Ten, merge))
			return light, light.CalculateLevel
		}, curse: func(e *Entity) {
			addTraitWithFeatures(e, "Curse", newSpellGrant(newHardSpell("Light", fxp.One), -fxp.Two, grantmerge.Ignore))
		}, bless: func(e *Entity) (SkillAdjustmentProvider, func() Level) {
			light := newHardSpell("Light", 0)
			light.SetDataOwner(e)
			e.Spells = append(e.Spells, light)
			addTraitWithFeatures(e, "Blessing", newSpellGrant(newHardSpell("Light", fxp.One), fxp.Two, grantmerge.Ignore))
			return light, light.CalculateLevel
		}, kept: fxp.Ten, next: fxp.Eight, far: fxp.Sixteen, ladder: []fxp.Int{fxp.Two, fxp.Four}},
		{name: "technique", add: func(e *Entity, merge grantmerge.Mode) (SkillAdjustmentProvider, func() Level) {
			// An Average technique on Karate at 10 goes a level higher with each point.
			addHardSkill(e, "Karate", fxp.Four)
			disarming := newTechniqueOf("Karate", "Disarming", fxp.One)
			disarming.SetDataOwner(e)
			e.Skills = append(e.Skills, disarming)
			addTraitWithFeatures(e, "Training", newSkillGrantOf(newTechniqueOf("Karate", "Disarming", fxp.One),
				grantadjust.SetLevel, fxp.FromInteger(14), merge))
			return disarming, func() Level { return disarming.CalculateLevel(nil) }
		}, curse: func(e *Entity) {
			addTraitWithFeatures(e, "Curse",
				newSkillGrant(newTechniqueOf("Karate", "Disarming", fxp.One), -fxp.Two, grantmerge.Ignore))
		}, kept: fxp.FromInteger(14), next: fxp.Five, far: fxp.Seven, ladder: []fxp.Int{fxp.Two, fxp.Three, fxp.Four}},
	} {
		top := tc.ladder[len(tc.ladder)-1]
		// down holds the points that buy each level from the one below the one kept down to the one the point buys.
		down := []fxp.Int{fxp.One}
		for _, points := range tc.ladder[:len(tc.ladder)-1] {
			down = append([]fxp.Int{points}, down...)
		}
		t.Run(tc.name+" kept highest", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepHighest)
			e.Recalculate()
			c.Equal(tc.kept, level().Level, "precondition: the grant holds the level above the one the point buys")
			// The next level above the kept one needs next points, for a skill or spell in the second stride beyond the
			// point bought, and decrementing comes back to the fewest points that still show the kept level, none,
			// since the grant keeps the level of an item with no points too.
			stepLevelUpAndDown(t, p, level, []fxp.Int{tc.next}, []fxp.Int{0})
			p.IncrementSkillLevel()
			c.Equal(tc.next, p.RawPoints(), "the points decrementing took off come back")
		})
		t.Run(tc.name+" kept highest with levels taken away", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepHighest)
			tc.curse(e)
			e.Recalculate()
			c.Equal(tc.kept, level().Level, "precondition: the grant holds the level above the one the point buys")
			// The level shown rises once the level bought, less the 2 taken away, passes the one kept, which takes
			// points beyond the first stride after the point bought.
			stepLevelUpAndDown(t, p, level, []fxp.Int{tc.far}, []fxp.Int{0})
		})
		t.Run(tc.name+" kept lowest", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepLowest)
			e.Recalculate()
			c.Equal(tc.kept-fxp.FromInteger(len(tc.ladder)), level().Level,
				"precondition: the point bought gives a level below the one kept")
			// The levels up to the one kept are bought as without the grant, and none above it.
			stepLevelUpAndDown(t, p, level, tc.ladder, nil)
			p.IncrementSkillLevel()
			c.Equal(top, p.RawPoints(), "the grant caps the level, so incrementing leaves the points alone")
			c.Equal(tc.kept, level().Level)
			stepLevelUpAndDown(t, p, level, nil, down)
			// With points that buy a level several strides above the one kept, decrementing comes down to the fewest
			// points that show the level below it.
			p.SetRawPoints(fxp.Sixteen)
			c.Equal(tc.kept, level().Level)
			stepLevelUpAndDown(t, p, level, nil, down[:1])
		})
		if tc.bless == nil {
			continue
		}
		t.Run(tc.name+" kept highest with no level of its own", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.add(e, grantmerge.KeepHighest)
			p.SetRawPoints(0)
			level = recalculatedLevel(e, level)
			c.Equal(tc.kept, level().Level, "precondition: the grant gives the item the level kept")
			// The points that buy the level kept stand in for the item's own only while it has none, so the next level
			// still needs as many points of its own.
			stepLevelUpAndDown(t, p, level, []fxp.Int{tc.next}, []fxp.Int{0})
		})
		t.Run(tc.name+" levels added to one with no level of its own", func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			p, level := tc.bless(e)
			level = recalculatedLevel(e, level)
			c.Equal(tc.kept-fxp.One, level().Level, "precondition: the grant gives the item the points of its levels")
			// The first point gives the item a level of its own, which the grant then adds its levels to in place of
			// the points it gave.
			stepLevelUpAndDown(t, p, level, []fxp.Int{fxp.One}, []fxp.Int{0})
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

// TestSkillFillWithNameableKeys verifies that nameable substitution keys are extracted from all of a skill's text
// fields, including the optional specialization (see issue #1054).
func TestSkillFillWithNameableKeys(t *testing.T) {
	c := check.New(t)
	var s Skill
	s.Name = "Skill @who@"
	s.Specialization = "@spec@"
	s.OptionalSpecialization = "@opt@"
	s.LocalNotes = "@note@"
	m := make(map[string]string)
	s.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{
		"who":  nameable.Unset,
		"spec": nameable.Unset,
		"opt":  nameable.Unset,
		"note": nameable.Unset,
	}, m)
}

// TestHasDefaultToOptionalSpecialization verifies the default-relationship test used by the swap feature: a skill never
// has a default to itself, and a specialized skill that defaults to the unspecialized skill ("is empty") does not
// register a default to a sibling specialization, whether a required one or, consistent with skill matching, an
// optional one.
func TestHasDefaultToOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, "Fire", fxp.Four, "base") // defaults to the unspecialized skill
	water := addRitualMagic(e, "Water", fxp.Four, "base")
	earth := addRitualMagic(e, "Earth", fxp.Four, "")
	earth.Specialization = ""
	earth.OptionalSpecialization = "Earth"
	e.Recalculate()

	c.True(fire.HasDefaultTo(base), "Fire's empty-specialization default targets the unspecialized skill")
	c.True(water.HasDefaultTo(base), "Water defaults to the unspecialized Ritual Magic")
	c.False(water.HasDefaultTo(water), "a skill must never default to itself")
	c.False(fire.HasDefaultTo(water), "a specialized skill must not register a default to a sibling specialization")
	c.False(water.HasDefaultTo(fire), "a specialized skill must not register a default to a sibling specialization")
	c.False(fire.HasDefaultTo(earth),
		"an empty-specialization default must not be treated as a default to an optionally-specialized sibling")
}

// TestSwapNotStuckWithOptionalSpecializationSibling reproduces the reported issue: with a base skill carrying the
// automatic -2 default to an optionally-specialized sibling (Earth), swapping Earth's defaults must actually cycle them
// rather than staying stuck.
func TestSwapNotStuckWithOptionalSpecializationSibling(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "") // no declared defaults; only the synthetic optional-spec default
	addRitualMagic(e, "Air", fxp.Four, "any")
	addRitualMagic(e, "Fire", fxp.Four, "any")
	earth := addRitualMagic(e, "Earth", fxp.Four, "any")
	earth.Specialization = ""
	earth.OptionalSpecialization = "Earth"
	e.Recalculate()

	// Earth must be able to cycle through its possible defaults rather than being stuck.
	targets := map[string]bool{}
	visitedBase := false
	for range 5 {
		earth.SwapToNextDefault()
		if ds := earth.DefaultSkill(); ds != nil {
			targets[ds.String()] = true
			if ds == base {
				visitedBase = true
			}
		} else {
			targets["<none>"] = true
		}
	}
	c.True(len(targets) > 1, "swapping Earth must actually cycle its default, not stay stuck on one choice")
	c.True(visitedBase, "Earth must be able to default to the unspecialized base skill")
}

// TestDefaultSkillDoesNotFlipToHigherSpecialized verifies that specialized skills defaulting to the unspecialized skill
// record that default, and resolve it (for display and swapping) to the unspecialized skill, whichever sibling
// specialization's level climbs above it.
func TestDefaultSkillDoesNotFlipToHigherSpecialized(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name        string
		firePoints  fxp.Int
		waterPoints fxp.Int
	}{
		{"Water above the base", fxp.One, fxp.FromInteger(20)},
		{"Fire above the base", fxp.Twelve, fxp.One},
	} {
		e := NewEntity()
		base := addRitualMagic(e, "", fxp.Eight, "")
		fire := addRitualMagic(e, "Fire", one.firePoints, "base")
		water := addRitualMagic(e, "Water", one.waterPoints, "base")
		e.Recalculate()

		c.True(max(fire.LevelData.Level, water.LevelData.Level) > base.LevelData.Level,
			"%s: the test requires a sibling's level to exceed the base skill's", one.name)
		for _, sk := range []*Skill{fire, water} {
			spec := sk.SpecializationWithReplacements()
			c.NotNil(sk.DefaultedFrom, "%s: %s should have a resolved default", one.name, spec)
			c.Equal("", sk.DefaultedFrom.SpecializationWithReplacements(nil),
				"%s: %s must default to the unspecialized skill, not a higher-level sibling", one.name, spec)
			c.True(sk.DefaultSkill() == base, "%s: %s must resolve to the unspecialized base skill", one.name, spec)
		}
	}
}

// TestDefaultPersistsAfterCreation verifies that a skill which defaults to any sibling picks the highest-level default
// when first resolved, then keeps that choice rather than auto-jumping to whichever sibling later becomes the highest.
func TestDefaultPersistsAfterCreation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, "Fire", fxp.Twelve, "") // highest at creation
	water := addRitualMagic(e, "Water", fxp.One, "")
	air := addRitualMagic(e, "Air", fxp.One, "any")
	e.Recalculate()

	c.NotNil(air.DefaultSkill(), "Air should resolve a default")
	c.Equal("Fire", air.DefaultSkill().SpecializationWithReplacements(),
		"Air should initially default to the highest-level sibling")

	// Raise Water far above Fire; Air must keep its existing choice instead of jumping to the new highest.
	water.SetRawPoints(fxp.FromInteger(40))
	e.Recalculate()
	c.Equal("Fire", air.DefaultSkill().SpecializationWithReplacements(),
		"Air must keep its chosen default after creation, not auto-jump to the new highest sibling")
}

// TestAlternateDefaultsAvailable verifies that only a skill with more than one resolvable default reports that
// alternates are available (which is what enables the Swap Defaults menu for non-mutual defaults).
func TestAlternateDefaultsAvailable(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, "Fire", fxp.Twelve, "base") // single resolvable default
	addRitualMagic(e, "Water", fxp.One, "")
	air := addRitualMagic(e, "Air", fxp.One, "any") // three resolvable defaults
	e.Recalculate()

	c.True(air.AlternateDefaultsAvailable(), "Air defaults to any sibling, so it has multiple choices")
	c.False(fire.AlternateDefaultsAvailable(), "Fire has a single resolvable default, so there is nothing to swap among")
}

// TestSwapToNextDefaultCycles verifies that swapping cycles through every resolvable default and wraps back around.
func TestSwapToNextDefaultCycles(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, "Fire", fxp.Twelve, "")
	addRitualMagic(e, "Water", fxp.One, "")
	air := addRitualMagic(e, "Air", fxp.One, "any")
	e.Recalculate()

	start := air.DefaultSkill().SpecializationWithReplacements()
	seen := map[string]bool{start: true}
	for range 2 {
		air.SwapToNextDefault()
		e.Recalculate()
		seen[air.DefaultSkill().SpecializationWithReplacements()] = true
	}
	c.Equal(3, len(seen), "swapping should visit all three siblings (unspecialized, Fire, Water)")

	air.SwapToNextDefault()
	e.Recalculate()
	c.Equal(start, air.DefaultSkill().SpecializationWithReplacements(),
		"cycling should wrap back to the starting default")
}

// TestSyntheticOptionalSpecDefaultTracksBest verifies that a skill whose only default is the automatic -2 default to a
// sibling's optional specialization tracks the best such sibling dynamically rather than locking onto the first one. It
// must not be treated as a user-managed (swappable, sticky) default.
func TestSyntheticOptionalSpecDefaultTracksBest(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "") // no declared defaults of its own
	earth := addRitualMagic(e, "Earth", fxp.Four, "")
	fire := addRitualMagic(e, "Fire", fxp.Four, "")
	// Make both Earth and Fire use an optional specialization, so the base skill gains synthetic -2 defaults to them.
	earth.Specialization = ""
	earth.OptionalSpecialization = "Earth"
	fire.Specialization = ""
	fire.OptionalSpecialization = "Fire"
	e.Recalculate()

	c.NotNil(base.DefaultedFrom, "base should pick up a synthetic default to an optionally-specialized sibling")
	c.False(base.AlternateDefaultsAvailable(),
		"a skill with no declared defaults is not user-swappable, even with multiple synthetic defaults")

	// Raise Fire well above Earth; the synthetic default must follow the best sibling instead of staying on Earth.
	fire.SetRawPoints(fxp.FromInteger(40))
	e.Recalculate()
	c.Equal("Fire", base.DefaultedFrom.SpecializationWithReplacements(nil),
		"base's synthetic optional-spec default must track the best sibling (Fire), not stay locked on Earth")
}

// TestTechniqueDefaultAnySpecializationIgnoresQualifier reproduces issue #1061: a technique whose skill default was set
// to a specific specialization ("is Demolition") and then switched to "whose specialization is anything" must resolve
// against any matching specialization rather than staying locked on the leftover qualifier text, and of the skills it
// matches must use the highest-level one rather than whichever is encountered first.
func TestTechniqueDefaultAnySpecializationIgnoresQualifier(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	const explosives = "Explosives"
	demolition := addTestSkill(e, explosives, "Demolition", "", fxp.Four)
	underwater := addTestSkill(e, explosives, "Underwater Demolition", "", fxp.FromInteger(20))
	e.Recalculate()
	c.True(underwater.LevelData.Level > demolition.LevelData.Level,
		"precondition: Underwater Demolition must be the higher-level skill")

	// Restricted to "is Demolition", the default resolves only to the lower-level Demolition skill.
	isDef := &SkillDefault{
		DefaultType:    SkillID,
		Name:           textCriteria(criteria.IsText, explosives),
		Specialization: textCriteria(criteria.IsText, "Demolition"),
	}
	isResult := CalculateTechniqueLevel(e, nil, "Technique", "", nil, isDef, difficulty.Average, fxp.One, false, nil, nil)

	// The same default still carrying the "Demolition" qualifier but with the comparison changed to "anything" must
	// match any specialization, so the higher-level Underwater Demolition wins.
	anyDef := &SkillDefault{
		DefaultType:    SkillID,
		Name:           textCriteria(criteria.IsText, explosives),
		Specialization: textCriteria(criteria.AnyText, "Demolition"),
	}
	anyResult := CalculateTechniqueLevel(e, nil, "Technique", "", nil, anyDef, difficulty.Average, fxp.One, false, nil, nil)

	c.True(anyResult.Level > isResult.Level, "an 'anything' specialization must resolve to the higher-level "+
		"Underwater Demolition, not stay locked on the Demolition qualifier")
	c.Equal(underwater.LevelData.Level+fxp.One, anyResult.Level,
		"the default must resolve to the highest-level matching skill, plus the technique's 1 point")
}

// TestTechniqueDefaultAnySpecializationResolvesSkill exercises the full technique node path for issue #1061: the
// resolved default skill, the technique's level, and the satisfied check must all honor the "anything" specialization
// rather than the leftover qualifier text.
func TestTechniqueDefaultAnySpecializationResolvesSkill(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	const explosives = "Explosives"
	addTestSkill(e, explosives, "Demolition", "", fxp.Four)
	underwater := addTestSkill(e, explosives, "Underwater Demolition", "", fxp.FromInteger(20))

	tech := NewTechnique(e, nil, explosives)
	// Leftover qualifier from a prior "is Demolition" selection, now switched to "anything".
	tech.TechniqueDefault.Specialization = textCriteria(criteria.AnyText, "Demolition")
	tech.TechniqueDefault.Modifier = -fxp.Two
	e.Skills = append(e.Skills, tech)
	e.Recalculate()

	def := tech.DefaultSkill()
	c.NotNil(def, "the technique must resolve a default skill")
	c.True(def == underwater, "the technique must default to the higher-level Underwater Demolition, not Demolition")
	c.True(tech.TechniqueSatisfied(nil, ""), "the technique must be satisfied by any matching specialization")
	// The "Default:" note must name the skill actually matched, not the stale qualifier text.
	c.Equal("Default: "+explosives+" (Underwater Demolition)-2", tech.ModifierNotes(),
		"the default note must show the resolved skill, not the leftover Demolition qualifier")
}

// TestSyntheticOptionalSpecDefaultSameRequiredSpec verifies the automatic optional-specialty default only relates
// skills that share the same required specialization. A fully-unspecialized skill must not pick up a default to a
// sibling that has a different required specialization, even if that sibling also carries an optional specialization.
func TestSyntheticOptionalSpecDefaultSameRequiredSpec(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "") // truly unspecialized
	// Different required specialization (different college) that also carries an optional specialization.
	fire := addTestSkill(e, "Ritual Magic", "Fire", "", fxp.Twelve)
	fire.OptionalSpecialization = "Flame"
	// Same (empty) required specialization, with an optional specialization - the legitimate optional-specialty case.
	earth := addTestSkill(e, "Ritual Magic", "", "", fxp.Four)
	earth.OptionalSpecialization = "Earth"
	e.Recalculate()

	specs := map[string]bool{}
	for _, d := range base.resolveToSpecificDefaults() {
		specs[d.Specialization.Qualifier] = true
	}
	c.True(specs["Earth"], "base should default to the same-required-spec optional-Earth sibling")
	c.False(specs["Fire"], "base must not default to a sibling with a different required specialization")
	c.False(specs["Flame"], "base must not default to a sibling with a different required specialization")
}

// TestSwapToSiblingReassignsOrphan reproduces the reported issue: with Air defaulting to Earth, cycling Earth's
// defaults eventually selects Air. Air can no longer default to Earth (that would be a cycle), so it must fall back to
// its best remaining default rather than ending up with no default at all.
func TestSwapToSiblingReassignsOrphan(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // base (best non-cyclic default)
	addRitualMagic(e, "Fire", fxp.Four, "base")
	addRitualMagic(e, "Water", fxp.Four, "base")
	air := addRitualMagic(e, "Air", fxp.Four, "any")
	earth := addRitualMagic(e, "Earth", fxp.Four, "any")
	e.Recalculate()

	// Establish Air -> Earth.
	air.DefaultedFrom = newSkillDefaultTo("Ritual Magic", "Earth", false, -fxp.Six)
	e.Recalculate()
	c.Equal("Earth", air.DefaultSkill().SpecializationWithReplacements(), "precondition: Air defaults to Earth")

	// Cycle Earth's defaults until it selects Air (Air is one of Earth's valid choices).
	reached := false
	for range 8 {
		earth.SwapToNextDefault()
		if ds := earth.DefaultSkill(); ds != nil && ds.SpecializationWithReplacements() == "Air" {
			reached = true
			break
		}
	}
	c.True(reached, "cycling Earth's defaults should be able to select Air")

	// Air must not be orphaned: it keeps a default, resolving to the best remaining choice (the base skill).
	c.NotNil(air.DefaultedFrom, "Air must keep a default when other choices remain")
	c.NotNil(air.DefaultSkill(), "Air must resolve to a real skill")
	c.Equal("", air.DefaultSkill().SpecializationWithReplacements(),
		"Air should fall back to the best remaining default (the base skill)")

	// The two must not default to each other.
	c.True(air.DefaultSkill() != earth, "Air must not default back to Earth")
	c.Equal("Air", earth.DefaultSkill().SpecializationWithReplacements(), "Earth should keep its chosen default of Air")
}

// TestSwapNeverCreatesMutualCycle verifies that no sequence of swaps (nor a hand-forced cycle) can leave two skills
// defaulting to each other, which would be invalid.
func TestSwapNeverCreatesMutualCycle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // base, no default
	air := addRitualMagic(e, "Air", fxp.Four, "any")
	earth := addRitualMagic(e, "Earth", fxp.Four, "any")
	e.Recalculate()

	hasCycle := func() bool {
		for _, a := range e.Skills {
			if da := a.DefaultSkill(); da != nil && da.DefaultSkill() == a {
				return true
			}
		}
		return false
	}

	c.False(hasCycle(), "no cycle should exist initially")
	for range 6 {
		air.SwapToNextDefault()
		e.Recalculate()
		earth.SwapToNextDefault()
		e.Recalculate()
		c.False(hasCycle(), "swapping must never make two skills default to each other")
	}

	// A directly-forced cycle must be broken on the next recalculation.
	air.DefaultedFrom = newSkillDefaultTo("Ritual Magic", "Earth", false, -fxp.Six)
	earth.DefaultedFrom = newSkillDefaultTo("Ritual Magic", "Air", false, -fxp.Six)
	e.Recalculate()
	c.False(hasCycle(), "a forced mutual cycle must be broken on recalculation")
}

// TestMutualSwapDefaults verifies that the classic pairwise swap (two skills that default to each other) still flips
// the primary skill and stays stable, using the same routing the UI uses.
func TestMutualSwapDefaults(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	broadsword := addTestSkill(e, "Broadsword", "", "", fxp.Four)
	broadsword.Defaults = []*SkillDefault{newSkillDefaultTo("Rapier", "", true, -fxp.Four)}
	rapier := addTestSkill(e, "Rapier", "", "", fxp.Eight) // more points, so it is the primary skill
	rapier.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Four)}
	e.Recalculate()

	c.NotNil(broadsword.DefaultedFrom, "Broadsword should default to Rapier initially")
	c.Nil(rapier.DefaultedFrom, "Rapier should stand on its own initially")

	// Drive the swap through the same decision the sheet uses.
	swapViaUI(broadsword)
	e.Recalculate()

	c.Nil(broadsword.DefaultedFrom, "after the swap, Broadsword should stand on its own")
	c.NotNil(rapier.DefaultedFrom, "after the swap, Rapier should default to Broadsword")

	// The relationship must remain stable across further recalculations.
	for range 3 {
		e.Recalculate()
	}
	c.Nil(broadsword.DefaultedFrom, "the swapped relationship should be stable")
	c.NotNil(rapier.DefaultedFrom, "the swapped relationship should be stable")
}

// swapViaUI mirrors the routing in the sheet's swapDefaults handler: prefer a mutual swap, then a best-swappable
// partner, and finally cycling the skill's own alternate defaults.
func swapViaUI(skill *Skill) {
	if !skill.CanSwapDefaults() {
		return
	}
	if swap := skill.DefaultSkill(); skill.CanSwapDefaultsWith(swap) {
		skill.DefaultedFrom = nil
		swap.SwapDefaults()
	} else if other := skill.BestSwappableSkill(); other != nil {
		skill.DefaultedFrom = nil
		other.SwapDefaults()
	} else {
		skill.SwapToNextDefault()
	}
}

// TestSwapDefaultsLeavesGrantedSkillsAlone verifies that a skill a grant put on the sheet, which changes only with its
// grant, neither swaps its own default nor serves as the partner of a swap.
func TestSwapDefaultsLeavesGrantedSkillsAlone(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	shortsword := addTestSkill(e, "Shortsword", "", "", fxp.One)
	shortsword.Defaults = []*SkillDefault{newSkillDefaultTo("Broadsword", "", true, -fxp.Two)}
	broadsword := newEasySkill("Broadsword", fxp.One)
	broadsword.Difficulty.Difficulty = difficulty.Average
	broadsword.Defaults = []*SkillDefault{newSkillDefaultTo("Shortsword", "", true, -fxp.Two)}
	addTraitWithFeatures(e, "Training", newSkillGrant(broadsword, fxp.Four, grantmerge.Ignore))
	e.Recalculate()
	granted := grantedOf(e.Skills)[0]
	c.NotNil(shortsword.DefaultedFrom, "precondition: Shortsword defaults to the granted Broadsword")
	c.False(granted.CanSwapDefaults(), "a granted skill's default is its grant's business")
	c.False(shortsword.CanSwapDefaultsWith(granted), "nor can another skill swap with it")
	c.Nil(shortsword.BestSwappableSkill())
	swapViaUI(shortsword)
	e.Recalculate()
	c.NotNil(shortsword.DefaultedFrom, "so the swap changes nothing")
	c.Nil(granted.DefaultedFrom)
}

// TestRecalculateSettlesDefaultDrivenSkillLevel verifies that a recalculation leaves the stored level of a skill whose
// level is driven by its default in step with the level computed on demand, along with the level of a weapon that
// defaults to that skill. The stored level used to be computed against the default chosen by the previous pass, so
// when a leveled trait raised the attribute the default points at, the stored level -- and so the weapon -- showed the
// state before the change, and only caught up on the next edit.
func TestRecalculateSettlesDefaultDrivenSkillLevel(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// An IQ-based skill with no points, whose level therefore comes entirely from its DX default.
	sk := addTestSkill(e, "Melee Weapon", "", "", 0)
	sk.Defaults = []*SkillDefault{{DefaultType: DexterityID}}

	// A leveled trait that raises DX by one per level.
	bonus := NewAttributeBonus(DexterityID)
	bonus.PerLevel = true
	trait := addTraitWithFeatures(e, "Melee Combat", bonus)
	trait.CanLevel = true

	// A weapon that defaults to the skill.
	eqp := addCarriedEquipmentWithFeatures(e, "Super Sledge")
	w := NewWeapon(eqp, true)
	w.Defaults = []*SkillDefault{{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: "Melee Weapon"}},
	}}
	eqp.Weapons = []*Weapon{w}
	e.Recalculate()

	dx := e.ResolveAttributeCurrent(DexterityID)
	c.Equal(dx, sk.LevelData.Level, "the skill starts at its DX default")
	c.Equal(dx, w.SkillLevel(nil), "the weapon starts at the skill's level")

	for _, levels := range []fxp.Int{fxp.One, fxp.Two, fxp.Three, fxp.Two, fxp.One, 0} {
		trait.Levels = levels
		e.Recalculate()
		want := dx + levels
		c.Equal(want, sk.CalculateLevel(nil).Level, "levels %v: the level computed on demand follows DX", levels)
		c.Equal(want, sk.LevelData.Level, "levels %v: the stored level must not lag behind", levels)
		c.Equal(want, w.SkillLevel(nil), "levels %v: the weapon's level must not lag behind", levels)
		c.False(sk.UpdateLevel(), "levels %v: a further update must find nothing left to change", levels)
	}
}
