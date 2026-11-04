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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/spellmatch"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/model/nameable"
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

// TestSpellRitualsFollowLevelClassAndOverrides pins the rituals text each level band gives, the parts the missile,
// blocking and enchantment classes leave out, and the ritual overrides, which replace a part at every level, drop it
// when "-", and take replacements.
func TestSpellRitualsFollowLevelClassAndOverrides(t *testing.T) {
	const (
		fullRitual  = "**Ritual:** need both hands and feet free and must speak"
		quietRitual = "**Ritual:** speak quietly and make a gesture"
		smallRitual = "**Ritual:** speak a word or two OR make a small gesture"
		noRitual    = "**Ritual:** none"
		halfTime    = "; **Time:** x1/2, rounded up (min 1 second)"
	)
	for _, tc := range []struct {
		name      string
		iq        fxp.Int
		class     string
		overrides RitualOverrides
		answers   map[string]string
		want      string
	}{
		{name: "below 10", iq: fxp.Nine, want: fullRitual + "; **Time:** Doubled"},
		{name: "10 to 14", iq: fxp.Ten, want: quietRitual},
		{name: "15 to 19", iq: fxp.Fifteen, want: smallRitual + "; **Cost:** -1"},
		{name: "15 to 19 blocking", iq: fxp.Fifteen, class: "Blocking", want: smallRitual},
		{name: "20 to 24", iq: fxp.Twenty, want: noRitual + halfTime + "; **Cost:** -2"},
		{
			name: "25 to 29",
			iq:   fxp.TwentyFive,
			want: noRitual + "; **Time:** x1/4, rounded up (min 1 second); **Cost:** -3",
		},
		{name: "20 to 24 missile", iq: fxp.Twenty, class: "Missile", want: noRitual + "; **Cost:** -2"},
		{name: "20 to 24 blocking", iq: fxp.Twenty, class: "Blocking", want: noRitual + halfTime},
		{name: "15 to 19 enchantment", iq: fxp.Fifteen, class: "Enchantment", want: smallRitual},
		{name: "20 to 24 enchantment", iq: fxp.Twenty, class: "Enchantment", want: noRitual},
		{
			name:      "override beats the class",
			iq:        fxp.Twenty,
			class:     "Enchantment",
			overrides: RitualOverrides{Cost: "full"},
			want:      noRitual + "; **Cost:** full",
		},
		{
			name:      "dash omits the part",
			iq:        fxp.Twenty,
			overrides: RitualOverrides{Time: "-", Cost: "-"},
			want:      noRitual,
		},
		{
			name:      "override applies where the level gives no part",
			iq:        fxp.Ten,
			overrides: RitualOverrides{Time: "not reduced", Cost: "not reduced"},
			want:      quietRitual + "; **Time:** not reduced; **Cost:** not reduced",
		},
		{
			name:      "override beats blocking",
			iq:        fxp.Fifteen,
			class:     "Blocking",
			overrides: RitualOverrides{Cost: "full"},
			want:      smallRitual + "; **Cost:** full",
		},
		{
			name:      "ritual override leaves the other parts alone",
			iq:        fxp.Nine,
			overrides: RitualOverrides{Ritual: "must speak"},
			want:      "**Ritual:** must speak; **Time:** Doubled",
		},
		{
			name:      "overrides take replacements",
			iq:        fxp.Twenty,
			overrides: RitualOverrides{Ritual: "@Ritual@", Time: "@Time@", Cost: "@Cost@"},
			answers:   map[string]string{"Ritual": "must speak", "Time": "as listed", "Cost": "full"},
			want:      "**Ritual:** must speak; **Time:** as listed; **Cost:** full",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			e.SheetSettings.ShowSpellAdj = true
			e.Attributes.Set[IntelligenceID].Adjustment = tc.iq - fxp.Ten
			sp := addTestSpell(e, "Spell", fxp.Four)
			if tc.class != "" {
				sp.Class = tc.class
			}
			sp.RitualOverrides = tc.overrides
			if tc.answers != nil {
				sp.ApplyNameableKeys(tc.answers)
			}
			e.Recalculate()
			c.Equal(tc.iq, sp.LevelData.Level, "a Hard spell at 4 points is IQ+0")
			c.Equal(tc.want, sp.Rituals())
		})
	}
}

// TestSpellRitualOverridesAreSourceDataSavedOnlyWhenSet checks that the ritual overrides are saved only when one is
// set, so that existing files re-save unchanged, load back, and are hashed, so that a library spell differing only
// in one is reported as mismatched.
func TestSpellRitualOverridesAreSourceDataSavedOnlyWhenSet(t *testing.T) {
	c := check.New(t)
	spell := NewSpell(NewEntity(), nil, false)
	spell.Name = "Enchant"
	data, err := jio.Marshal(spell)
	c.NoError(err)
	c.NotContains(string(data), `"rituals"`, "unset overrides aren't saved")
	for _, one := range []struct {
		part string
		set  func()
	}{
		{"ritual", func() { spell.RitualOverrides.Ritual = "must speak" }},
		{"time", func() { spell.RitualOverrides.Time = "not reduced" }},
		{"cost", func() { spell.RitualOverrides.Cost = "-" }},
	} {
		hash := Hash64(spell)
		one.set()
		c.NotEqual(hash, Hash64(spell), "the %s override is source data and is hashed", one.part)
	}
	data, err = jio.Marshal(spell)
	c.NoError(err)
	c.Contains(string(data), `"rituals":{"ritual":"must speak","time":"not reduced","cost":"-"}`)
	var restored Spell
	c.NoError(jio.Unmarshal(data, &restored))
	c.Equal(spell.RitualOverrides, restored.RitualOverrides, "the overrides load back")
}

func TestSpellLevelStepping(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	sp := NewSpell(e, nil, false)
	sp.Name = "Stepping"
	e.Spells = append(e.Spells, sp)
	c.Equal(difficulty.Hard, sp.Difficulty.Difficulty)
	sp.SetRawPoints(fxp.One)
	cached := func() Level {
		c.Equal(sp.CalculateLevel(), sp.LevelData, "cached level must track the points")
		return sp.LevelData
	}
	stepLevelUpAndDown(t, sp, cached,
		[]fxp.Int{fxp.Two, fxp.Four, fxp.Eight, fxp.Twelve},
		[]fxp.Int{fxp.Eight, fxp.Four, fxp.Two, fxp.One})

	sp.DecrementSkillLevel()
	c.Equal(fxp.Int(0), sp.Points)

	container := NewSpell(e, nil, true)
	container.IncrementSkillLevel()
	c.Equal(fxp.Int(0), container.Points)
}

// TestSpellFeaturesRoundTrip verifies that a spell's features are written and read back, including the switchable flag,
// and that a spell with none writes no "features" key at all.
func TestSpellFeaturesRoundTrip(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	data, err := jio.Marshal(spell)
	c.NoError(err, "a spell with no features should marshal")
	c.NotContains(string(data), `"features"`, "a spell with no features writes no features key")

	bonus := newSkillBonusTo("Alchemy", fxp.Two)
	bonus.Switchable = true
	spell.Features = Features{bonus, NewAttributeBonus(StrengthID)}
	data, err = jio.Marshal(spell)
	c.NoError(err, "a spell with features should marshal")
	c.Contains(string(data), `"features"`, "a spell's features are written")
	c.Contains(string(data), `"switchable":true`, "a spell's switchable feature keeps its flag")

	var restored Spell
	c.NoError(jio.Unmarshal(data, &restored), "a spell with features should load")
	c.Equal(2, len(restored.Features), "both features come back")
	if len(restored.Features) == 2 {
		reloaded, ok := restored.Features[0].(*SkillBonus)
		c.True(ok, "the first feature is still a skill bonus")
		if ok {
			c.Equal("Alchemy", reloaded.NameCriteria.Qualifier, "the skill bonus keeps its criteria")
			c.Equal(fxp.Two, reloaded.Amount, "the skill bonus keeps its amount")
			c.True(reloaded.IsSwitchable(), "the skill bonus is still switchable")
		}
		c.False(restored.Features[1].IsSwitchable(), "the second feature is still always-on")
	}

	// The stored form is stable across a round-trip. The comparison is made over the form without the derived "calc"
	// values, since the reloaded spell has no owning entity to derive them from.
	stored, err := MarshalWithoutCalc(spell)
	c.NoError(err, "the spell should marshal without its derived values")
	again, err := MarshalWithoutCalc(&restored)
	c.NoError(err, "the reloaded spell should marshal without its derived values")
	c.Equal(string(stored), string(again), "the stored form is stable across a round-trip")
}

// TestSpellEditDataCopiesFeatures verifies that a spell editor works on its own copy of the features, in both
// directions, so that editing them in an editor doesn't reach back into the spell it was opened on.
func TestSpellEditDataCopiesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Features = Features{newSkillBonusTo("Alchemy", fxp.Two)}

	// What the editor is handed when it opens.
	var edit SpellEditData
	edit.CopyFrom(spell)
	c.Equal(1, len(edit.Features), "the copy holds the spell's features")
	c.False(edit.Features[0] == spell.Features[0], "the copy holds its own features, not the spell's")

	// Editing the copy leaves the original alone.
	editBonus := skillBonusAt(c, edit.Features, 0)
	spellBonus := skillBonusAt(c, spell.Features, 0)
	editBonus.NameCriteria.Qualifier = "Astronomy"
	editBonus.SetSwitchable(true)
	c.Equal("Alchemy", spellBonus.NameCriteria.Qualifier, "the original criteria is untouched")
	c.False(spellBonus.IsSwitchable(), "the original switchable flag is untouched")

	// Applying the editor's data hands the target its own copy in turn.
	target := NewSpell(e, nil, false)
	edit.ApplyTo(target)
	c.Equal(1, len(target.Features), "the applied data holds the edited features")
	targetBonus := skillBonusAt(c, target.Features, 0)
	c.Equal("Astronomy", targetBonus.NameCriteria.Qualifier, "the edit was applied")
	c.True(targetBonus.IsSwitchable(), "the switchable flag was applied")
	c.False(target.Features[0] == edit.Features[0], "the target holds its own features, not the editor's")
	targetBonus.Amount = fxp.Ten
	c.Equal(fxp.Two, editBonus.Amount, "editing the target doesn't reach back into the editor")
}

// skillBonusAt returns the feature at the given index as a *SkillBonus, failing the test if it isn't one.
func skillBonusAt(c check.Checker, features Features, index int) *SkillBonus {
	c.True(index < len(features), "the feature index must be in range")
	bonus, ok := features[index].(*SkillBonus)
	c.True(ok, "the feature must be a skill bonus")
	if !ok {
		return NewSkillBonus()
	}
	return bonus
}

// TestSpellHashIncludesFeatures verifies that a spell's features participate in the source-data hash, so that a spell
// whose features differ from its library source is reported as out of sync.
func TestSpellHashIncludesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	before := Hash64(spell)

	spell.Features = Features{newSkillBonusTo("Alchemy", fxp.Two)}
	withFeature := Hash64(spell)
	c.NotEqual(before, withFeature, "adding a feature changes the spell's hash")

	spell.Features[0].SetSwitchable(true)
	c.NotEqual(withFeature, Hash64(spell), "making the feature switchable changes the spell's hash")

	spell.Features = nil
	c.Equal(before, Hash64(spell), "removing the feature restores the original hash")
}

// TestSpellSyncWithSourceClonesFeatures verifies that syncing a spell with its library source brings the source's
// features across as clones, so that later edits to one don't reach into the other. The library data is injected
// directly into the source matcher, which is what a real sync gets from the library file it loaded.
func TestSpellSyncWithSourceClonesFeatures(t *testing.T) {
	c := check.New(t)
	e := NewEntity()

	// Stand in for the spell as it appears in a library file: same identity, but carrying a feature.
	source := NewSpell(nil, nil, false)
	source.Name = "Fireball"
	bonus := newSkillBonusTo("Alchemy", fxp.Two)
	bonus.Switchable = true
	source.Features = Features{bonus}

	local := NewSpell(e, nil, false)
	local.Name = "Fireball"
	e.Spells = append(e.Spells, local)
	libFile := LibraryFile{Library: "Test Library", Path: "Test.spl"}
	local.Source = Source{LibraryFile: libFile, TID: source.TID}
	stubLibrarySources(t, e.SourceMatcher(), libFile, source)

	c.Equal(0, len(local.Features), "precondition: the local copy starts with no features")
	local.SyncWithSource()
	c.Equal(1, len(local.Features), "the sync brings the source's features across")
	if len(local.Features) == 1 {
		c.True(local.Features[0].IsSwitchable(), "the synced feature keeps its switchable flag")
		c.False(local.Features[0] == source.Features[0], "the synced feature is a clone, not the source's own")
		skillBonusAt(c, local.Features, 0).Amount = fxp.Ten
		c.Equal(fxp.Two, skillBonusAt(c, source.Features, 0).Amount, "editing the synced copy leaves the source alone")
	}
}

// TestSpellBonusFromItsOwnSpell verifies that a spell carrying a bonus that names itself resolves to a sane level
// rather than recursing forever. The spell's level is what the bonus is collected during, so a bonus pointing back at
// its own spell is the shortest cycle the machinery has to survive.
func TestSpellBonusFromItsOwnSpell(t *testing.T) {
	c := check.New(t)

	// A control entity, to establish the level the spell would have without any bonus.
	control := NewEntity()
	baseline := NewSpell(control, nil, false)
	baseline.Name = "Fireball"
	baseline.Points = fxp.Four
	control.Spells = append(control.Spells, baseline)
	control.Recalculate()
	c.True(baseline.LevelData.Level > 0, "precondition: the spell resolves to a positive level")

	e := NewEntity()
	spell := NewSpell(e, nil, false)
	spell.Name = "Fireball"
	spell.Points = fxp.Four
	bonus := NewSpellBonus()
	bonus.SpellMatchType = spellmatch.Name
	bonus.NameCriteria.Compare = criteria.IsText
	bonus.NameCriteria.Qualifier = "Fireball"
	bonus.Amount = fxp.Two
	spell.Features = Features{bonus}
	e.Spells = append(e.Spells, spell)

	c.NotPanics(func() { e.Recalculate() }, "a self-referencing spell bonus must not blow the stack")
	c.Equal(baseline.LevelData.Level+fxp.Two, spell.LevelData.Level, "the spell's own bonus raises its own level")

	// The per-level form reads the spell's own cached level rather than recomputing it, so it also terminates.
	bonus.PerLevel = true
	c.NotPanics(func() { e.Recalculate() }, "a per-level self-referencing spell bonus must not blow the stack")
	c.True(spell.LevelData.Level > 0, "the spell still resolves to a positive level")
	c.NotEqual(fxp.Min, spell.LevelData.Level, "the spell's level stays resolvable")
}

// TestSpellFillWithNameableKeysIncludesFeatures verifies that the substitution keys used inside a spell's features are
// offered to the user along with the ones in the spell's own text, so a spell can carry a feature whose target is
// filled in when the spell is added to a sheet.
func TestSpellFillWithNameableKeysIncludesFeatures(t *testing.T) {
	c := check.New(t)

	spell := NewSpell(nil, nil, false)
	spell.Name = "Enhance @what@"
	bonus := newSkillBonusTo("@skill@", fxp.Two)
	bonus.SpecializationCriteria.Compare = criteria.IsText
	bonus.SpecializationCriteria.Qualifier = "@spec@"
	spell.Features = Features{bonus}

	m := make(map[string]string)
	spell.FillWithNameableKeys(m, nil)
	c.Equal(map[string]string{
		"what":  nameable.Unset,
		"skill": nameable.Unset,
		"spec":  nameable.Unset,
	}, m, "the keys from the spell's features are offered along with its own")
}
