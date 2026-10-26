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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/grantmerge"
	"github.com/richardwilkes/toolbox/v2/check"
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
		def.Specialization = criteria.Text{Compare: criteria.IsText, Qualifier: specialization}
	}
	return def
}

func textCriteria(compare criteria.StringComparison, qualifier string) criteria.Text {
	return criteria.Text{Compare: compare, Qualifier: qualifier}
}

// addRitualMagic creates a Ritual Magic skill with the given specialization and points. When defaultMode is non-empty
// it also installs a default: "base" defaults to the unspecialized Ritual Magic (an exact, empty-specialization match);
// "any" defaults to any Ritual Magic skill.
func addRitualMagic(e *Entity, specialization string, points fxp.Int, defaultMode string) *Skill {
	sk := addTestSkill(e, "Ritual Magic", specialization, "", points)
	switch defaultMode {
	case "base":
		sk.Defaults = []*SkillDefault{newSkillDefaultTo("Ritual Magic", "", false, -fxp.Six)}
	case "any":
		sk.Defaults = []*SkillDefault{newSkillDefaultTo("Ritual Magic", "", true, -fxp.Six)}
	}
	return sk
}

// TestSkillMatchingEmptySpecialization verifies that an exact ("is") match against an empty specialization matches only
// the truly unspecialized skill, not one that has only a required or only an optional specialization, or both, and
// that an exact match against a required or an optional specialization matches the skill carrying it. The first is
// the bug that caused a specialized Ritual Magic to default to a sibling specialization instead of the unspecialized
// one.
func TestSkillMatchingEmptySpecialization(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	addRitualMagic(e, "", fxp.Eight, "")
	addRitualMagic(e, "Fire", fxp.Four, "")
	addRitualMagic(e, "Water", fxp.Four, "")
	earth := addRitualMagic(e, "Earth", fxp.Four, "")
	earth.Specialization = ""
	earth.OptionalSpecialization = "Earth"
	light := addTestSkill(e, "Ritual Magic", "Light", "", fxp.Four)
	light.OptionalSpecialization = "Dark"
	e.Recalculate()

	nameIs := textCriteria(criteria.IsText, "Ritual Magic")
	for _, one := range []struct {
		specialization string
		wantRequired   string
		wantOptional   string
		msg            string
	}{
		{"", "", "", "an empty 'is' specialization must match only the truly unspecialized skill"},
		{"Fire", "Fire", "", "an 'is Fire' specialization must match only the Fire skill"},
		{"Earth", "", "Earth", "an 'is Earth' specialization must match the optionally-specialized skill"},
		{"Light", "Light", "Dark", "the required specialization should match"},
		{"Dark", "Light", "Dark", "the optional specialization should match"},
	} {
		matches := e.SkillMatching(nameIs, textCriteria(criteria.IsText, one.specialization), nil, false, nil)
		c.Equal(1, len(matches), one.msg)
		if len(matches) == 1 {
			c.Equal(one.wantRequired, matches[0].SpecializationWithReplacements(), one.msg)
			c.Equal(one.wantOptional, matches[0].OptionalSpecializationWithReplacements(), one.msg)
		}
	}

	c.Equal(5, len(e.SkillMatching(nameIs, criteria.Text{}, nil, false, nil)),
		"an empty (any) specialization criteria must match every skill")
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
