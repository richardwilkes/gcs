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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/difficulty"
	"github.com/richardwilkes/toolbox/v2/check"
)

// addTestSpell creates a non-container spell owned by the entity, giving it the supplied name and points, then appends
// it to the entity's spell list.
func addTestSpell(e *Entity, name string, points fxp.Int) *Spell {
	s := NewSpell(e, nil, false)
	s.Name = name
	s.Points = points
	e.Spells = append(e.Spells, s)
	return s
}

// addTestRitualMagicSpell creates a ritual magic spell owned by the entity, giving it the supplied name and colleges,
// then appends it to the entity's spell list.
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
	t := NewTrait(e, nil, false)
	t.Name = "Magery"
	t.Features = Features{bonus}
	e.Traits = append(e.Traits, t)
	return t
}

// TestSpellAdjustedRelativeLevel verifies that AdjustedRelativeLevel returns the cached relative level for a positive
// level spell (matching the freshly-computed level), and fxp.Min for containers and unowned spells. This guards the
// switch from recomputing the whole level via CalculateLevel to reading the cached LevelData.
func TestSpellAdjustedRelativeLevel(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	s := addTestSpell(e, "Fireball", fxp.Four)
	e.Recalculate()

	// Precondition: the spell must resolve to a positive level so AdjustedRelativeLevel takes the non-Min branch.
	c.True(s.LevelData.Level > 0, "precondition: the spell must have a positive level")

	// The cached relative level must be returned, and it must match a fresh recomputation (proving the cheaper cached
	// read is equivalent to the old CalculateLevel path).
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
// have stays unresolvable (fxp.Min), even when a spell bonus applies. A phantom level of 0 would let the bonus lift the
// spell to a positive, displayable level it hasn't actually earned.
func TestRitualMagicSpellLevelWithoutRitualSkill(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	withCollege := addTestRitualMagicSpell(e, "Fireball", "Fire")
	withoutCollege := addTestRitualMagicSpell(e, "Light")
	addTestSpellBonus(e, fxp.Three)
	e.Recalculate()

	c.Equal(fxp.Min, withCollege.LevelData.Level,
		"a ritual magic spell with no matching ritual skill must remain unresolvable")
	c.Equal("-", withCollege.LevelData.LevelAsString(false), "an unresolvable level displays as \"-\"")
	c.Equal(fxp.Min, withoutCollege.LevelData.Level,
		"a college-less ritual magic spell with no matching ritual skill must remain unresolvable")
	c.Equal("-", withoutCollege.LevelData.LevelAsString(false), "an unresolvable level displays as \"-\"")
}

// TestRitualMagicSpellLevelNegativeBonusNoOverflow verifies that a net-negative spell bonus is not added to an
// unresolvable level. fxp.Min is math.MinInt64, so adding a negative amount to it wraps around to a huge positive level.
func TestRitualMagicSpellLevelNegativeBonusNoOverflow(t *testing.T) {
	c := check.New(t)

	e := NewEntity()
	withCollege := addTestRitualMagicSpell(e, "Fireball", "Fire")
	withoutCollege := addTestRitualMagicSpell(e, "Light")
	addTestSpellBonus(e, -fxp.Three)
	e.Recalculate()

	c.Equal(fxp.Min, withCollege.LevelData.Level, "a negative bonus must not wrap an unresolvable level")
	c.Equal(fxp.Min, withoutCollege.LevelData.Level, "a negative bonus must not wrap an unresolvable level")
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

// TestSpellMarshalUnsatisfiedReason verifies that the unsatisfied reason is written in both calc branches. Ritual magic
// spells that are unsatisfied typically have a level of fxp.Min, which takes the branch that used to omit the field --
// so exactly the spells that have a reason were the ones losing it.
func TestSpellMarshalUnsatisfiedReason(t *testing.T) {
	c := check.New(t)
	const reason = "Requires Magery 1"
	marshaled := func(s *Spell) string {
		data, err := json.Marshal(s)
		c.NoError(err)
		return string(data)
	}

	// A spell with no resolvable level takes the abbreviated calc branch.
	noLevel := NewSpell(nil, nil, false)
	noLevel.Name = "Fireball"
	c.True(noLevel.LevelData.Level <= 0, "precondition: the spell must have no resolvable level")
	c.False(strings.Contains(marshaled(noLevel), "unsatisfied_reason"), "no reason means no field")
	noLevel.UnsatisfiedReason = reason
	c.True(strings.Contains(marshaled(noLevel), `"unsatisfied_reason":"`+reason+`"`),
		"an unsatisfied spell with no resolvable level must still record the reason")

	// A spell with a positive level takes the full calc branch.
	e := NewEntity()
	leveled := addTestSpell(e, "Fireball", fxp.Four)
	e.Recalculate()
	c.True(leveled.LevelData.Level > 0, "precondition: the spell must have a positive level")
	leveled.UnsatisfiedReason = reason
	c.True(strings.Contains(marshaled(leveled), `"unsatisfied_reason":"`+reason+`"`),
		"an unsatisfied spell with a level must record the reason")
}
