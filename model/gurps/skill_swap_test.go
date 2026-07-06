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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

const (
	broadswordSkill  = "Broadsword"
	rapierSkill      = "Rapier"
	ritualMagicSkill = "Ritual Magic"
	airSpec          = "Air"
	earthSpec        = "Earth"
	fireSpec         = "Fire"
	waterSpec        = "Water"
	darkSpec         = "Dark"
	lightSpec        = "Light"
	baseDef          = "base"
	anyDef           = "any"
)

// newSkillDefaultTo builds a skill-based default. When anySpecialization is true, the default carries no specialization
// criteria (matching any specialization); otherwise it requires an exact ("is") match of the given specialization,
// including the empty string for an unspecialized skill.
func newSkillDefaultTo(name, specialization string, anySpecialization bool, modifier fxp.Int) *SkillDefault {
	def := &SkillDefault{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: name}},
		Modifier:    modifier,
	}
	if !anySpecialization {
		def.Specialization = criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: specialization}}
	}
	return def
}

func textCriteria(compare criteria.StringComparison, qualifier string) criteria.Text {
	return criteria.Text{TextData: criteria.TextData{Compare: compare, Qualifier: qualifier}}
}

// addRitualMagic creates a Ritual Magic skill with the given specialization and points. When defaultMode is non-empty
// it also installs a default: "base" defaults to the unspecialized Ritual Magic (an exact, empty-specialization match);
// "any" defaults to any Ritual Magic skill.
func addRitualMagic(e *Entity, specialization string, points fxp.Int, defaultMode string) *Skill {
	sk := addTestSkill(e, ritualMagicSkill, specialization, "", points)
	switch defaultMode {
	case baseDef:
		sk.Defaults = []*SkillDefault{newSkillDefaultTo(ritualMagicSkill, "", false, -fxp.Six)}
	case anyDef:
		sk.Defaults = []*SkillDefault{newSkillDefaultTo(ritualMagicSkill, "", true, -fxp.Six)}
	}
	return sk
}

// TestSkillMatchingEmptySpecialization verifies that an exact ("is") match against an empty specialization matches only
// the unspecialized skill, not specialized skills that merely have an empty *optional* specialization. This is the bug
// that caused a specialized Ritual Magic to default to a sibling specialization instead of the unspecialized one.
func TestSkillMatchingEmptySpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, fireSpec, fxp.Four, "")
	addRitualMagic(e, waterSpec, fxp.Four, "")
	e.Recalculate()

	nameIs := textCriteria(criteria.IsText, ritualMagicSkill)

	matches := e.SkillMatching(nameIs, textCriteria(criteria.IsText, ""), nil, false, nil)
	c.Equal(1, len(matches), "an empty 'is' specialization must match only the unspecialized skill")
	if len(matches) == 1 {
		c.Equal("", matches[0].SpecializationWithReplacements(), "the single match must be the unspecialized skill")
	}

	matches = e.SkillMatching(nameIs, textCriteria(criteria.IsText, fireSpec), nil, false, nil)
	c.Equal(1, len(matches), "an 'is "+fireSpec+"' specialization must match only the "+fireSpec+" skill")
	if len(matches) == 1 {
		c.Equal(fireSpec, matches[0].SpecializationWithReplacements(),
			"the single match must be the "+fireSpec+" skill")
	}

	matches = e.SkillMatching(nameIs, criteria.Text{}, nil, false, nil)
	c.Equal(3, len(matches), "an empty (any) specialization criteria must match all three skills")
}

// TestSkillMatchingOptionalSpecialization verifies that a skill with no required specialization but a non-empty
// optional specialization is matched by that optional specialization, and is kept distinct from a fully unspecialized
// sibling. An "is empty" criteria must match only the truly unspecialized skill.
func TestSkillMatchingOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // truly unspecialized
	earth := addRitualMagic(e, earthSpec, fxp.Four, "")
	earth.Specialization = ""
	earth.OptionalSpecialization = earthSpec
	e.Recalculate()

	nameIs := textCriteria(criteria.IsText, ritualMagicSkill)

	matches := e.SkillMatching(nameIs, textCriteria(criteria.IsText, ""), nil, false, nil)
	c.Equal(1, len(matches), "an empty specialization must not match a skill carrying an optional specialization")
	if len(matches) == 1 {
		c.Equal("", matches[0].OptionalSpecializationWithReplacements(),
			"the match must be the truly unspecialized skill")
	}

	matches = e.SkillMatching(nameIs, textCriteria(criteria.IsText, earthSpec), nil, false, nil)
	c.Equal(1, len(matches), "an 'is "+earthSpec+"' specialization must match the optionally-specialized skill")
	if len(matches) == 1 {
		c.Equal(earthSpec, matches[0].OptionalSpecializationWithReplacements(),
			"the match must be the optional-"+earthSpec+" skill")
	}
}

// TestSkillMatchingBothSpecializations verifies a skill that carries both a required and an optional specialization is
// matched by either one, and is not matched by an empty specialization.
func TestSkillMatchingBothSpecializations(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // truly unspecialized
	light := addTestSkill(e, ritualMagicSkill, lightSpec, "", fxp.Four)
	light.OptionalSpecialization = darkSpec
	e.Recalculate()

	name := textCriteria(criteria.IsText, ritualMagicSkill)
	c.Equal(1, len(e.SkillMatching(name, textCriteria(criteria.IsText, lightSpec), nil, false, nil)),
		"the required specialization should match")
	c.Equal(1, len(e.SkillMatching(name, textCriteria(criteria.IsText, darkSpec), nil, false, nil)),
		"the optional specialization should match")
	matches := e.SkillMatching(name, textCriteria(criteria.IsText, ""), nil, false, nil)
	c.Equal(1, len(matches), "an empty specialization should match only the truly unspecialized skill")
	if len(matches) == 1 {
		c.Equal("", matches[0].SpecializationWithReplacements(), "the empty match must be the unspecialized skill")
		c.Equal("", matches[0].OptionalSpecializationWithReplacements(),
			"the empty match must be the unspecialized skill")
	}
}

// TestHasDefaultToOptionalSpecialization verifies the swap-relationship test is consistent with skill matching for a
// skill that has an optional specialization: a default to the unspecialized skill ("is empty") must not be reported as
// a default to the optionally-specialized sibling.
func TestHasDefaultToOptionalSpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, fireSpec, fxp.Four, baseDef) // defaults to the unspecialized skill
	earth := addRitualMagic(e, earthSpec, fxp.Four, "")
	earth.Specialization = ""
	earth.OptionalSpecialization = earthSpec
	e.Recalculate()

	c.True(fire.HasDefaultTo(base), fireSpec+"'s empty-specialization default targets the unspecialized skill")
	c.False(fire.HasDefaultTo(earth),
		"an empty-specialization default must not be treated as a default to an optionally-specialized sibling")
}

// TestSwapNotStuckWithOptionalSpecializationSibling reproduces the reported issue: with a base skill carrying the
// automatic -2 default to an optionally-specialized sibling (Earth), swapping defaults on either must actually change
// things rather than being permanently stuck.
func TestSwapNotStuckWithOptionalSpecializationSibling(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "") // no declared defaults; only the synthetic optional-spec default
	addRitualMagic(e, airSpec, fxp.Four, anyDef)
	addRitualMagic(e, fireSpec, fxp.Four, anyDef)
	earth := addRitualMagic(e, earthSpec, fxp.Four, anyDef)
	earth.Specialization = ""
	earth.OptionalSpecialization = earthSpec
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
	c.True(len(targets) > 1, "swapping "+earthSpec+" must actually cycle its default, not stay stuck on one choice")
	c.True(visitedBase, earthSpec+" must be able to default to the unspecialized base skill")
}

// TestSpecializedDefaultsToUnspecialized verifies that specialized skills which default to the unspecialized skill
// resolve to it, regardless of the relative levels of the sibling specializations.
func TestSpecializedDefaultsToUnspecialized(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")                     // unspecialized base
	fire := addRitualMagic(e, fireSpec, fxp.Twelve, baseDef) // higher level than the base
	water := addRitualMagic(e, waterSpec, fxp.One, baseDef)
	e.Recalculate()

	c.NotNil(fire.DefaultedFrom, fireSpec+" should have a resolved default")
	c.Equal("", fire.DefaultedFrom.SpecializationWithReplacements(nil),
		fireSpec+" must default to the unspecialized skill")
	c.NotNil(water.DefaultedFrom, waterSpec+" should have a resolved default")
	c.Equal("", water.DefaultedFrom.SpecializationWithReplacements(nil),
		waterSpec+" must default to the unspecialized skill, not the higher-level "+fireSpec+" specialization")
}

// TestHasDefaultToExactMatch verifies the default-relationship test used by the swap feature: a skill never has a
// default to itself, and a specialized skill that defaults to the unspecialized skill does not register a default to a
// sibling specialization.
func TestHasDefaultToExactMatch(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, fireSpec, fxp.Four, baseDef)
	water := addRitualMagic(e, waterSpec, fxp.Four, baseDef)
	e.Recalculate()

	c.True(fire.HasDefaultTo(base), fireSpec+" defaults to the unspecialized "+ritualMagicSkill)
	c.True(water.HasDefaultTo(base), waterSpec+" defaults to the unspecialized "+ritualMagicSkill)
	c.False(water.HasDefaultTo(water), "a skill must never default to itself")
	c.False(fire.HasDefaultTo(water), "a specialized skill must not register a default to a sibling specialization")
	c.False(water.HasDefaultTo(fire), "a specialized skill must not register a default to a sibling specialization")
}

// TestDefaultSkillDoesNotFlipToHigherSpecialized verifies that the skill resolved from a recorded default (used for
// display and swapping) stays the unspecialized skill even once a sibling specialization's level climbs above it.
func TestDefaultSkillDoesNotFlipToHigherSpecialized(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	base := addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, fireSpec, fxp.One, baseDef)
	water := addRitualMagic(e, waterSpec, fxp.FromInteger(20), baseDef) // level climbs above the base
	e.Recalculate()

	c.True(water.LevelData.Level > base.LevelData.Level,
		"the test requires "+waterSpec+"'s level to exceed the base skill's")

	fireDefault := fire.DefaultSkill()
	c.NotNil(fireDefault, fireSpec+" should resolve a default skill")
	c.True(fireDefault == base, fireSpec+" must resolve to the unspecialized base skill")

	waterDefault := water.DefaultSkill()
	c.NotNil(waterDefault, waterSpec+" should resolve a default skill")
	c.True(waterDefault == base,
		waterSpec+" must resolve to the unspecialized base skill, not the higher-level "+waterSpec+" itself")
}

// TestDefaultPersistsAfterCreation verifies that a skill which defaults to any sibling picks the highest-level default
// when first resolved, then keeps that choice rather than auto-jumping to whichever sibling later becomes the highest.
func TestDefaultPersistsAfterCreation(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, fireSpec, fxp.Twelve, "") // highest at creation
	water := addRitualMagic(e, waterSpec, fxp.One, "")
	air := addRitualMagic(e, airSpec, fxp.One, anyDef)
	e.Recalculate()

	c.NotNil(air.DefaultSkill(), airSpec+" should resolve a default")
	c.Equal(fireSpec, air.DefaultSkill().SpecializationWithReplacements(),
		airSpec+" should initially default to the highest-level sibling")

	// Raise Water far above Fire; Air must keep its existing choice instead of jumping to the new highest.
	water.SetRawPoints(fxp.FromInteger(40))
	e.Recalculate()
	c.Equal(fireSpec, air.DefaultSkill().SpecializationWithReplacements(),
		airSpec+" must keep its chosen default after creation, not auto-jump to the new highest sibling")
}

// TestAlternateDefaultsAvailable verifies that only a skill with more than one resolvable default reports that
// alternates are available (which is what enables the Swap Defaults menu for non-mutual defaults).
func TestAlternateDefaultsAvailable(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	fire := addRitualMagic(e, fireSpec, fxp.Twelve, baseDef) // single resolvable default
	addRitualMagic(e, waterSpec, fxp.One, "")
	air := addRitualMagic(e, airSpec, fxp.One, anyDef) // three resolvable defaults
	e.Recalculate()

	c.True(air.AlternateDefaultsAvailable(), airSpec+" defaults to any sibling, so it has multiple choices")
	c.False(fire.AlternateDefaultsAvailable(),
		fireSpec+" has a single resolvable default, so there is nothing to swap among")
}

// TestSwapToNextDefaultCycles verifies that swapping cycles through every resolvable default and wraps back around,
// persisting each choice.
func TestSwapToNextDefaultCycles(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, fireSpec, fxp.Twelve, "")
	addRitualMagic(e, waterSpec, fxp.One, "")
	air := addRitualMagic(e, airSpec, fxp.One, anyDef)
	e.Recalculate()

	start := air.DefaultSkill().SpecializationWithReplacements()
	seen := map[string]bool{start: true}
	for range 2 {
		air.SwapToNextDefault()
		e.Recalculate()
		seen[air.DefaultSkill().SpecializationWithReplacements()] = true
	}
	c.Equal(3, len(seen), "swapping should visit all three siblings (unspecialized, "+fireSpec+", "+waterSpec+")")

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
	earth := addRitualMagic(e, earthSpec, fxp.Four, "")
	fire := addRitualMagic(e, fireSpec, fxp.Four, "")
	// Make both Earth and Fire use an optional specialization, so the base skill gains synthetic -2 defaults to them.
	earth.Specialization = ""
	earth.OptionalSpecialization = earthSpec
	fire.Specialization = ""
	fire.OptionalSpecialization = fireSpec
	e.Recalculate()

	c.NotNil(base.DefaultedFrom, "base should pick up a synthetic default to an optionally-specialized sibling")
	c.False(base.AlternateDefaultsAvailable(),
		"a skill with no declared defaults is not user-swappable, even with multiple synthetic defaults")

	// Raise Fire well above Earth; the synthetic default must follow the best sibling instead of staying on Earth.
	fire.SetRawPoints(fxp.FromInteger(40))
	e.Recalculate()
	c.Equal(fireSpec, base.DefaultedFrom.SpecializationWithReplacements(nil),
		"base's synthetic optional-spec default must track the best sibling ("+fireSpec+
			"), not stay locked on "+earthSpec+"")
}

// TestTechniqueDefaultUsesHighestLevelSkill verifies that when a skill-based default matches more than one skill (e.g.
// several core Ritual Magic skills via an empty/wildcard specialization), the highest-level match is used rather than
// whichever happens to be encountered first.
func TestTechniqueDefaultUsesHighestLevelSkill(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	low := addTestSkill(e, ritualMagicSkill, "", "", fxp.Four)             // inserted first, lower level
	high := addTestSkill(e, ritualMagicSkill, "", "", fxp.FromInteger(20)) // higher level
	e.Recalculate()
	c.True(high.LevelData.Level > low.LevelData.Level, "precondition: the second skill must be higher level")

	def := &SkillDefault{
		DefaultType: SkillID,
		Name:        criteria.Text{TextData: criteria.TextData{Compare: criteria.IsText, Qualifier: ritualMagicSkill}},
	}
	var limit fxp.Int
	result := CalculateTechniqueLevel(e, nil, "Imbue", "", nil, def, difficulty.Hard, 0, false, &limit, nil)
	c.Equal(high.LevelData.Level, result.Level, "the default must resolve to the highest-level matching skill")
}

// TestTechniqueDefaultAnySpecializationIgnoresQualifier reproduces issue #1061: a technique whose skill default was set
// to a specific specialization ("is Demolition") and then switched to "whose specialization is anything" must resolve
// against any matching specialization rather than staying locked on the leftover qualifier text.
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
	fire := addTestSkill(e, ritualMagicSkill, fireSpec, "", fxp.Twelve)
	fire.OptionalSpecialization = "Flame"
	// Same (empty) required specialization, with an optional specialization - the legitimate optional-specialty case.
	earth := addTestSkill(e, ritualMagicSkill, "", "", fxp.Four)
	earth.OptionalSpecialization = earthSpec
	e.Recalculate()

	specs := map[string]bool{}
	for _, d := range base.resolveToSpecificDefaults() {
		specs[d.Specialization.Qualifier] = true
	}
	c.True(specs[earthSpec], "base should default to the same-required-spec optional-"+earthSpec+" sibling")
	c.False(specs[fireSpec], "base must not default to a sibling with a different required specialization")
	c.False(specs["Flame"], "base must not default to a sibling with a different required specialization")
}

// TestSwapToSiblingReassignsOrphan reproduces the reported issue: with Air defaulting to Earth, cycling Earth's
// defaults eventually selects Air. Air can no longer default to Earth (that would be a cycle), so it must fall back to
// its best remaining default rather than ending up with no default at all.
func TestSwapToSiblingReassignsOrphan(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // base (best non-cyclic default)
	addRitualMagic(e, fireSpec, fxp.Four, baseDef)
	addRitualMagic(e, waterSpec, fxp.Four, baseDef)
	air := addRitualMagic(e, airSpec, fxp.Four, anyDef)
	earth := addRitualMagic(e, earthSpec, fxp.Four, anyDef)
	e.Recalculate()

	// Establish Air -> Earth.
	air.DefaultedFrom = newSkillDefaultTo(ritualMagicSkill, earthSpec, false, -fxp.Six)
	e.Recalculate()
	c.Equal(earthSpec, air.DefaultSkill().SpecializationWithReplacements(),
		"precondition: "+airSpec+" defaults to "+earthSpec)

	// Cycle Earth's defaults until it selects Air (Air is one of Earth's valid choices).
	reached := false
	for range 8 {
		earth.SwapToNextDefault()
		if ds := earth.DefaultSkill(); ds != nil && ds.SpecializationWithReplacements() == airSpec {
			reached = true
			break
		}
	}
	c.True(reached, "cycling "+earthSpec+"'s defaults should be able to select Air")

	// Air must not be orphaned: it keeps a default, resolving to the best remaining choice (the base skill).
	c.NotNil(air.DefaultedFrom, airSpec+" must keep a default when other choices remain")
	c.NotNil(air.DefaultSkill(), airSpec+" must resolve to a real skill")
	c.Equal("", air.DefaultSkill().SpecializationWithReplacements(),
		airSpec+" should fall back to the best remaining default (the base skill)")

	// The two must not default to each other.
	c.True(air.DefaultSkill() != earth, airSpec+" must not default back to "+earthSpec)
	c.Equal(airSpec, earth.DefaultSkill().SpecializationWithReplacements(),
		earthSpec+" should keep its chosen default of "+airSpec)
}

// TestSwapNeverCreatesMutualCycle verifies that no sequence of swaps (nor a hand-forced cycle) can leave two skills
// defaulting to each other, which would be invalid.
func TestSwapNeverCreatesMutualCycle(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "") // base, no default
	air := addRitualMagic(e, airSpec, fxp.Four, anyDef)
	earth := addRitualMagic(e, earthSpec, fxp.Four, anyDef)
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
	air.DefaultedFrom = newSkillDefaultTo(ritualMagicSkill, earthSpec, false, -fxp.Six)
	earth.DefaultedFrom = newSkillDefaultTo(ritualMagicSkill, airSpec, false, -fxp.Six)
	e.Recalculate()
	c.False(hasCycle(), "a forced mutual cycle must be broken on recalculation")
}

// TestMutualSwapDefaults verifies that the classic pairwise swap (two skills that default to each other) still flips
// the primary skill and stays stable, using the same routing the UI uses.
func TestMutualSwapDefaults(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	broadsword := addTestSkill(e, broadswordSkill, "", "", fxp.Four)
	broadsword.Defaults = []*SkillDefault{newSkillDefaultTo(rapierSkill, "", true, -fxp.Four)}
	rapier := addTestSkill(e, rapierSkill, "", "", fxp.Eight) // more points, so it is the primary skill
	rapier.Defaults = []*SkillDefault{newSkillDefaultTo(broadswordSkill, "", true, -fxp.Four)}
	e.Recalculate()

	c.NotNil(broadsword.DefaultedFrom, broadswordSkill+" should default to "+rapierSkill+" initially")
	c.Nil(rapier.DefaultedFrom, rapierSkill+" should stand on its own initially")

	// Drive the swap through the same decision the sheet uses.
	swapViaUI(broadsword)
	e.Recalculate()

	c.Nil(broadsword.DefaultedFrom, "after the swap, "+broadswordSkill+" should stand on its own")
	c.NotNil(rapier.DefaultedFrom, "after the swap, "+rapierSkill+" should default to "+broadswordSkill)

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
