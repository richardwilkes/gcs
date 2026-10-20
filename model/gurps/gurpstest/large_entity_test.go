// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurpstest_test

import (
	"os"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// LargeEntityDumpEnvVar names the environment variable that, when set to a file path, has TestWriteLargeEntity save
// the character NewLargeEntity builds for the size in the environment there, so that it can be opened in the app.
const LargeEntityDumpEnvVar = "GCS_BENCH_DUMP"

// countLeaves returns the number of non-container rows the traversal visits.
func countLeaves[T gurps.Node[T]](list []T) int {
	n := 0
	gurps.Traverse(func(T) bool { n++; return false }, false, true, list...)
	return n
}

// countConditionalModifierRows returns the number of rows, containers included, in a derived list.
func countConditionalModifierRows(list []*gurps.ConditionalModifier) int {
	n := 0
	for _, one := range list {
		if len(one.Children) == 0 {
			n++
		} else {
			n += countConditionalModifierRows(one.Children)
		}
	}
	return n
}

// Every list a sheet shows must be filled to the size asked for, with the containers open and the derived values
// resolved, or a benchmark built on the entity measures less than it claims to.
func TestNewLargeEntityFillsEveryList(t *testing.T) {
	c := check.New(t)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.1)
	e := gurpstest.NewLargeEntity(&size)

	// The traits list also holds the natural attacks and the traits with unsatisfied prerequisites.
	c.True(countLeaves(e.Traits) >= size.Traits, "traits: %d < %d", countLeaves(e.Traits), size.Traits)
	c.True(countLeaves(e.Skills) >= size.Skills, "skills: %d < %d", countLeaves(e.Skills), size.Skills)
	c.Equal(size.Spells+size.RitualMagicSpells, countLeaves(e.Spells), "spells")
	c.Equal(size.CarriedEquipment, countLeaves(e.CarriedEquipment), "carried equipment")
	c.Equal(size.OtherEquipment, countLeaves(e.OtherEquipment), "other equipment")
	c.Equal(size.Notes, countLeaves(e.Notes), "notes")
	c.Equal(size.Reactions, countConditionalModifierRows(e.Reactions()), "reactions")
	c.Equal(size.ConditionalModifiers, countConditionalModifierRows(e.ConditionalModifiers()), "conditional modifiers")

	// The weapons on unequipped and uncarried equipment stay off the page, so the page shows most, not all, of them;
	// the natural attacks add three melee weapons.
	melee := e.Weapons(true, false, true)
	ranged := e.Weapons(false, false, true)
	c.True(len(melee) >= size.MeleeWeapons*8/10 && len(melee) <= size.MeleeWeapons+3, "melee weapons: %d", len(melee))
	c.True(len(ranged) >= size.RangedWeapons*8/10 && len(ranged) <= size.RangedWeapons, "ranged weapons: %d",
		len(ranged))
	c.Equal(size.MeleeWeapons+3, len(e.Weapons(true, true, true)), "all melee weapons")
	c.Equal(size.RangedWeapons, len(e.Weapons(false, true, true)), "all ranged weapons")
	for _, w := range melee {
		c.True(w.SkillLevel(nil) > 0, "melee weapon %s has no skill level", w)
	}
	for _, w := range ranged {
		c.True(w.SkillLevel(nil) > 0, "ranged weapon %s has no skill level", w)
	}

	// The ritual magic spells resolve their levels through the college's "Ritual Magic (@Specialty@)" skill.
	rituals := 0
	gurps.Traverse(func(s *gurps.Spell) bool {
		if s.IsRitualMagic() {
			rituals++
			c.True(s.LevelData.Level > 0, "ritual magic spell %s has no level", s)
		}
		return false
	}, false, true, e.Spells...)
	c.Equal(size.RitualMagicSpells, rituals, "ritual magic spells")

	// Every container starts open, so a sheet built from the entity shows every row.
	gurps.Traverse(func(t *gurps.Trait) bool {
		if t.Container() {
			c.True(t.IsOpen(), "trait container %s is closed", t)
		}
		return false
	}, false, false, e.Traits...)
	gurps.Traverse(func(s *gurps.Spell) bool {
		if s.Container() {
			c.True(s.IsOpen(), "spell container %s is closed", s)
		}
		return false
	}, false, false, e.Spells...)
	gurps.Traverse(func(eqp *gurps.Equipment) bool {
		if eqp.Container() {
			c.True(eqp.IsOpen(), "equipment container %s is closed", eqp)
		}
		return false
	}, false, false, e.CarriedEquipment...)
}

// The same size must yield the same character, apart from the IDs, so that benchmark runs compare like with like.
func TestNewLargeEntityIsDeterministic(t *testing.T) {
	c := check.New(t)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.05)
	first := gurpstest.NewLargeEntity(&size)
	second := gurpstest.NewLargeEntity(&size)
	c.Equal(first.UnspentPoints(), second.UnspentPoints())
	c.Equal(first.TotalPoints, second.TotalPoints)
	c.Equal(len(first.Weapons(true, true, true)), len(second.Weapons(true, true, true)))
	c.Equal(first.WeightCarried(false), second.WeightCarried(false))
}

// TestWriteLargeEntity saves the character built for the size in the environment to the file named by
// LargeEntityDumpEnvVar, so that it can be opened in the app. It does nothing unless that variable is set.
func TestWriteLargeEntity(t *testing.T) {
	path := os.Getenv(LargeEntityDumpEnvVar)
	if path == "" {
		t.Skipf("%s is not set", LargeEntityDumpEnvVar)
	}
	c := check.New(t)
	size, err := gurpstest.LargeEntitySizeFromEnv()
	c.NoError(err)
	c.NoError(gurpstest.NewLargeEntity(&size).Save(path))
}
