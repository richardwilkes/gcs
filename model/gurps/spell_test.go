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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

// addTestSpell adds a spell with the given name and points to the entity.
func addTestSpell(e *Entity, name string, points fxp.Int) *Spell {
	s := NewSpell(e, nil, false)
	s.Name = name
	s.Points = points
	e.Spells = append(e.Spells, s)
	return s
}

// addTestRitualMagicSpell adds a ritual magic spell with the given name and colleges to the entity.
func addTestRitualMagicSpell(e *Entity, name string, colleges ...string) *Spell {
	s := NewRitualMagicSpell(e, nil, false)
	s.Name = name
	s.College = CollegeList(colleges)
	e.Spells = append(e.Spells, s)
	return s
}

// addTestSpellBonus attaches a trait to the entity that grants a flat bonus to all of the entity's spells.
func addTestSpellBonus(e *Entity, amount fxp.Int) *Trait {
	bonus := NewSpellBonus()
	bonus.Amount = amount
	return addTraitWithFeatures(e, "Magery", bonus)
}

// TestSpellAdjustedRelativeLevel verifies that AdjustedRelativeLevel returns the cached relative level for a positive
// level spell, matching a fresh CalculateLevel, and fxp.Min for containers and unowned spells.
func TestSpellAdjustedRelativeLevel(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	s := addTestSpell(e, "Fireball", fxp.Four)
	e.Recalculate()

	// Precondition: the spell must resolve to a positive level so AdjustedRelativeLevel takes the non-Min branch.
	c.True(s.LevelData.Level > 0, "precondition: the spell must have a positive level")

	// The cached relative level must be returned, and it must match a fresh recomputation.
	c.Equal(s.LevelData.RelativeLevel, s.AdjustedRelativeLevel())
	c.Equal(s.CalculateLevel().RelativeLevel, s.AdjustedRelativeLevel())

	// A container spell has no meaningful relative level.
	container := NewSpell(e, nil, true)
	c.Equal(fxp.Min, container.AdjustedRelativeLevel())

	// A spell with no owning entity has no relative level either.
	orphan := NewSpell(nil, nil, false)
	orphan.Name = "Fireball"
	orphan.Points = fxp.Four
	c.Equal(fxp.Min, orphan.AdjustedRelativeLevel())
}

// TestNewRitualMagicSpellPoints verifies that a new ritual magic spell starts with no points (it can still be used at
// its default level from the ritual magic skill), unlike a regular spell, which starts with one point.
func TestNewRitualMagicSpellPoints(t *testing.T) {
	c := check.New(t)
	c.Equal(fxp.Int(0), NewRitualMagicSpell(nil, nil, false).RawPoints(),
		"a new ritual magic spell starts with no points")
	c.Equal(fxp.One, NewSpell(nil, nil, false).RawPoints(), "a new regular spell starts with one point")
}

// TestRitualMagicSpellLevelWithoutRitualSkill verifies that a ritual magic spell whose ritual skill the entity does not
// have stays unresolvable (fxp.Min), even when a spell bonus applies. A phantom level of 0 would let a positive bonus
// lift the spell to a positive, displayable level it hasn't actually earned, and since fxp.Min is math.MinInt64, adding
// a negative bonus to it would wrap it around to a huge positive level.
func TestRitualMagicSpellLevelWithoutRitualSkill(t *testing.T) {
	c := check.New(t)
	for _, bonus := range []fxp.Int{fxp.Three, -fxp.Three} {
		e := NewEntity()
		withCollege := addTestRitualMagicSpell(e, "Fireball", "Fire")
		withoutCollege := addTestRitualMagicSpell(e, "Light")
		addTestSpellBonus(e, bonus)
		e.Recalculate()

		c.Equal(fxp.Min, withCollege.LevelData.Level,
			"bonus %s: a ritual magic spell with no matching ritual skill must remain unresolvable", bonus)
		c.Equal("-", withCollege.LevelData.LevelAsString(false),
			"bonus %s: an unresolvable level displays as \"-\"", bonus)
		c.Equal(fxp.Min, withoutCollege.LevelData.Level,
			"bonus %s: a college-less ritual magic spell with no matching ritual skill must remain unresolvable", bonus)
		c.Equal("-", withoutCollege.LevelData.LevelAsString(false),
			"bonus %s: an unresolvable level displays as \"-\"", bonus)
	}
}

// TestRitualMagicSpellLevelWithRitualSkill verifies that a resolvable ritual magic spell still picks the best college
// and still receives spell bonuses -- i.e. that guarding against the unresolvable case didn't suppress the normal path.
func TestRitualMagicSpellLevelWithRitualSkill(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	ritual := NewSkill(e, nil, false)
	ritual.Name = "Ritual Magic"
	ritual.Specialization = "Fire"
	ritual.Difficulty.Attribute = IntelligenceID
	ritual.Difficulty.Difficulty = difficulty.Hard
	ritual.Points = fxp.Four
	e.Skills = append(e.Skills, ritual)

	// "Water" has no ritual skill, so the level must come from the "Fire" college rather than a phantom 0 for "Water".
	spell := addTestRitualMagicSpell(e, "Fireball", "Water", "Fire")
	e.Recalculate()

	base := spell.LevelData.Level
	c.True(base > 0, "a ritual magic spell backed by a ritual skill must resolve to a positive level")

	addTestSpellBonus(e, fxp.Three)
	e.Recalculate()
	c.Equal(base+fxp.Three, spell.LevelData.Level, "a spell bonus still applies to a resolvable level")
}

// TestSpellRelativeLevelColumn verifies that the RSL column agrees with the exported "rsl" value produced by
// RelativeLevel(): a ritual magic spell's relative level is measured against its ritual magic skill, so it must be
// displayed as a bare signed number rather than being prefixed with the difficulty's attribute name, while a regular
// spell keeps the attribute prefix.
func TestSpellRelativeLevelColumn(t *testing.T) {
	c := check.New(t)

	primary := func(s *Spell) string {
		var data CellData
		s.CellData(SpellRelativeLevelColumn, &data)
		return data.Primary
	}

	e := NewEntity()
	ritualSkill := NewSkill(e, nil, false)
	ritualSkill.Name = "Ritual Magic"
	ritualSkill.Specialization = "Fire"
	ritualSkill.Difficulty.Attribute = IntelligenceID
	ritualSkill.Difficulty.Difficulty = difficulty.Hard
	ritualSkill.Points = fxp.Four
	e.Skills = append(e.Skills, ritualSkill)

	// The prereq count and the point total are chosen so that neither spell's relative level lands on zero, where the
	// attribute prefix would stand alone and the difference would be invisible.
	ritual := addTestRitualMagicSpell(e, "Fireball", "Fire")
	ritual.PrereqCount = 2
	regular := addTestSpell(e, "Ice Dagger", fxp.Eight)
	e.Recalculate()

	// Preconditions: both spells must resolve, and the ritual magic spell must have a non-zero relative level, so the
	// attribute prefix would be visible if it were being emitted.
	c.True(ritual.LevelData.Level > 0, "precondition: the ritual magic spell must have a positive level")
	c.True(regular.LevelData.Level > 0, "precondition: the regular spell must have a positive level")
	c.NotEqual(fxp.Int(0), ritual.AdjustedRelativeLevel(),
		"precondition: the ritual magic spell must have a non-zero relative level")
	c.NotEqual(fxp.Int(0), regular.AdjustedRelativeLevel(),
		"precondition: the regular spell must have a non-zero relative level")

	c.Equal(ritual.AdjustedRelativeLevel().StringWithSign(), primary(ritual),
		"a ritual magic spell's RSL is relative to its ritual magic skill, so no attribute prefix")
	c.Equal(ritual.RelativeLevel(), primary(ritual), "the RSL column must agree with the exported \"rsl\" value")

	iq := ResolveAttributeName(e, regular.Difficulty.Attribute)
	c.Equal(iq+regular.AdjustedRelativeLevel().StringWithSign(), primary(regular),
		"a regular spell's RSL is relative to its attribute, so it keeps the prefix")
	c.Equal(regular.RelativeLevel(), primary(regular), "the RSL column must agree with the exported \"rsl\" value")

	// A spell that cannot be resolved displays as "-" regardless of its kind.
	unresolvable := addTestRitualMagicSpell(e, "Light", "Water")
	e.Recalculate()
	c.Equal(fxp.Min, unresolvable.AdjustedRelativeLevel(),
		"precondition: the spell must have no resolvable relative level")
	c.Equal("-", primary(unresolvable), "an unresolvable relative level displays as \"-\"")
}

// TestSpellEquipmentPrereqPenaltyIsScopedToItsSpell verifies that the penalty generated when a spell's equipment
// prerequisite is unmet applies only to that spell, and is -10 rather than -5 when the spell has a tech level, matching
// the rule applied to skills. The penalty is built from NewSpellBonus(), whose default match type is
// spellmatch.AllColleges (which matches every spell), so it must explicitly be narrowed to spellmatch.Name.
func TestSpellEquipmentPrereqPenaltyIsScopedToItsSpell(t *testing.T) {
	c := check.New(t)

	// A control entity with a single spell and no prerequisites, used to establish the unpenalized level.
	control := NewEntity()
	baseline := addTestSpell(control, "Ice Dagger", fxp.Four)
	control.Recalculate()

	for _, tc := range []struct {
		name      string
		techLevel *string
		penalty   fxp.Int
	}{
		{name: "no tech level", penalty: -fxp.Five},
		{name: "tech level 8", techLevel: new("8"), penalty: -fxp.Ten},
	} {
		e := NewEntity()
		needsFocus := addTestSpell(e, "Fireball", fxp.Four)
		needsFocus.TechLevel = tc.techLevel
		eqpPrereq := NewEquippedEquipmentPrereq()
		eqpPrereq.NameCriteria.Qualifier = "Wizard's Focus"
		needsFocus.Prereq = NewPrereqList()
		needsFocus.Prereq.Prereqs = append(needsFocus.Prereq.Prereqs, eqpPrereq)
		eqpPrereq.Parent = needsFocus.Prereq
		other := addTestSpell(e, "Ice Dagger", fxp.Four)
		e.Recalculate()

		// Precondition: the entity has no equipment, so the prerequisite is unmet and the penalty is generated.
		c.NotEqual("", needsFocus.UnsatisfiedReason,
			"%s: precondition: the equipment prerequisite must be unsatisfied", tc.name)

		// The penalty applies to the spell that owns the unmet prerequisite...
		c.Equal(tc.penalty, e.SpellBonusFor(needsFocus.NameWithReplacements(), "", nil, nil, nil),
			"%s: the spell with the unmet equipment prerequisite takes the %s penalty", tc.name, tc.penalty)

		// ...and to nothing else, no matter how it is queried.
		c.Equal(fxp.Int(0), e.SpellBonusFor(other.NameWithReplacements(), "", nil, nil, nil),
			"%s: an unrelated spell must not take the penalty", tc.name)
		c.Equal(fxp.Int(0), e.SpellBonusFor("Ice Dagger", "Arcane", []string{"Water"}, []string{"Attack"}, nil),
			"%s: the penalty must not match via power source, college, or tags either", tc.name)

		// The unrelated spell's level must be unchanged from the control, while the owning spell's is lowered.
		c.Equal(baseline.LevelData.Level, other.LevelData.Level,
			"%s: an unrelated spell's level must not be reduced", tc.name)
		c.Equal(baseline.LevelData.Level+tc.penalty, needsFocus.LevelData.Level,
			"%s: the spell with the unmet equipment prerequisite must be lowered by the penalty", tc.name)
	}
}

// TestSpellEquipmentPrereqPenaltyHonorsAnyOfLists verifies that the equipment penalty follows the prerequisite list
// as a whole: a spell whose unmet equipment prerequisite sits in an "any of" list alongside a met skill prerequisite
// has satisfied prerequisites and so takes no penalty, and it takes the penalty once that skill prerequisite is no
// longer met. Whether the alternative is met depends on a skill level, so this also checks that the penalty settles
// correctly even though it is generated before the levels are updated.
func TestSpellEquipmentPrereqPenaltyHonorsAnyOfLists(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	control := addTestSpell(e, "Ice Dagger", fxp.Four)
	needsFocus := addTestSpell(e, "Fireball", fxp.Four)
	needsFocus.Prereq = NewPrereqList()
	needsFocus.Prereq.All = false
	eqpPrereq := NewEquippedEquipmentPrereq()
	eqpPrereq.Parent = needsFocus.Prereq
	eqpPrereq.NameCriteria.Qualifier = "Wizard's Focus"
	needsFocus.Prereq.Prereqs = append(needsFocus.Prereq.Prereqs, eqpPrereq)
	addSkillPrereq(needsFocus.Prereq, "Thaumatology", fxp.Ten)
	thaumatology := addTestSkill(e, "Thaumatology", "", "", fxp.Four) // IQ+1, so 11

	e.Recalculate()
	c.Equal("", needsFocus.UnsatisfiedReason, "the met skill prerequisite satisfies the list")
	c.Equal(fxp.Int(0), e.SpellBonusFor(needsFocus.NameWithReplacements(), "", nil, nil, nil),
		"a satisfied list carries no penalty")
	c.Equal(control.LevelData.Level, needsFocus.LevelData.Level, "the level is not penalized")

	thaumatology.Points = fxp.One // IQ-1, so 9, which no longer meets the alternative
	e.Recalculate()
	c.NotEqual("", needsFocus.UnsatisfiedReason, "with neither alternative met, the list is unsatisfied")
	c.Equal(-fxp.Five, e.SpellBonusFor(needsFocus.NameWithReplacements(), "", nil, nil, nil),
		"an unsatisfied list carries the penalty")
	c.Equal(control.LevelData.Level-fxp.Five, needsFocus.LevelData.Level, "the level is penalized")
	c.Equal(fxp.Int(0), e.SpellBonusFor(control.NameWithReplacements(), "", nil, nil, nil),
		"an unrelated spell must not take the penalty")

	thaumatology.Points = fxp.Four
	e.Recalculate()
	c.Equal("", needsFocus.UnsatisfiedReason, "meeting the alternative again satisfies the list")
	c.Equal(control.LevelData.Level, needsFocus.LevelData.Level, "and lifts the penalty")
}

// TestRitualMagicSpellWithBlankRitualSkillName verifies that a ritual magic spell whose ritual skill name is blank
// draws no level from a skill of another name, just as RitualMagicSatisfied finds no such skill to satisfy it.
func TestRitualMagicSpellWithBlankRitualSkillName(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	thaumatology := NewSkill(e, nil, false)
	thaumatology.Name = "Thaumatology"
	thaumatology.Specialization = "Fire"
	thaumatology.Difficulty.Attribute = IntelligenceID
	thaumatology.Difficulty.Difficulty = difficulty.VeryHard
	thaumatology.Points = fxp.Four
	e.Skills = append(e.Skills, thaumatology)
	generic := NewSkill(e, nil, false)
	generic.Name = "Broadsword"
	generic.Difficulty.Attribute = DexterityID
	generic.Difficulty.Difficulty = difficulty.Average
	generic.Points = fxp.Four
	e.Skills = append(e.Skills, generic)

	spell := addTestRitualMagicSpell(e, "Fireball", "Fire")
	spell.RitualSkillName = ""
	e.Recalculate()
	c.False(spell.RitualMagicSatisfied(nil, ""), "no skill has a blank name")
	c.Equal(fxp.Min, spell.LevelData.Level, "with no ritual skill, the level must remain unresolvable")

	spell.RitualSkillName = "Thaumatology"
	e.Recalculate()
	c.True(spell.RitualMagicSatisfied(nil, ""), "Thaumatology (Fire) satisfies the spell")
	c.True(spell.LevelData.Level > 0, "Thaumatology (Fire) gives the spell a level")
}

// TestRitualMagicSpellWithOptionallySpecializedRitualSkill verifies that a ritual skill with only an optional
// specialization serves the college it names but not, as an unspecialized skill would, every other college, so that
// RitualMagicSatisfied agrees with the level calculation.
func TestRitualMagicSpellWithOptionallySpecializedRitualSkill(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	ritual := NewSkill(e, nil, false)
	ritual.Name = "Ritual Magic"
	ritual.OptionalSpecialization = "Fire"
	ritual.Difficulty.Attribute = IntelligenceID
	ritual.Difficulty.Difficulty = difficulty.VeryHard
	ritual.Points = fxp.Four
	e.Skills = append(e.Skills, ritual)

	fire := addTestRitualMagicSpell(e, "Fireball", "Fire")
	water := addTestRitualMagicSpell(e, "Ice Dagger", "Water")
	e.Recalculate()
	c.True(fire.RitualMagicSatisfied(nil, ""), "Ritual Magic (Fire) satisfies a Fire spell")
	c.True(fire.LevelData.Level > 0, "Ritual Magic (Fire) gives a Fire spell a level")
	c.False(water.RitualMagicSatisfied(nil, ""), "Ritual Magic (Fire) is not an unspecialized skill")
	c.Equal(fxp.Min, water.LevelData.Level, "Ritual Magic (Fire) gives a Water spell no level")

	ritual.OptionalSpecialization = ""
	e.Recalculate()
	c.True(water.RitualMagicSatisfied(nil, ""), "an unspecialized Ritual Magic satisfies a Water spell")
	c.True(water.LevelData.Level > 0, "an unspecialized Ritual Magic gives a Water spell a level")
}
