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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/gcs/v5/model/jio"
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

// countTitled returns the number of rows in the list with title notes.
func countTitled[T interface {
	gurps.Node[T]
	TitleNotes() []string
}](list []T) int {
	n := 0
	gurps.Traverse(func(one T) bool {
		if len(one.TitleNotes()) != 0 {
			n++
		}
		return false
	}, false, true, list...)
	return n
}

// countConditionalModifierRows returns the number of rows in a derived list, leaving out the groups they are filed
// under.
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

// checkOpen fails for every container in the list that is closed.
func checkOpen[T gurps.Node[T]](c check.Checker, name string, list []T) {
	gurps.Traverse(func(one T) bool {
		if one.Container() {
			c.True(one.IsOpen(), "%s container %s is closed", name, one)
		}
		return false
	}, false, false, list...)
}

// checkParents fails for every row whose parent is not the container it sits in.
func checkParents[T gurps.Node[T]](c check.Checker, name string, list []T, parent T) {
	for _, one := range list {
		c.True(one.Parent() == parent, "%s %s has parent %v, not %v", name, one, one.Parent(), parent)
		checkParents(c, name, one.NodeChildren(), one)
	}
}

// volatileFields matches the parts of a character's JSON that differ between two builds of the same character: its
// IDs and its timestamps.
var volatileFields = regexp.MustCompile(`"(id|created_date|modified_date|when)":\s*"[^"]*"`)

// linkKeys matches the keys of the nameable links, which are the IDs of the modifiers they are for.
var linkKeys = regexp.MustCompile(`"[a-zA-Z][\w-]{16}":\s*\{`)

// describe returns the lines of the character's JSON, in the form Entity.Save writes, with the IDs, including those
// that key the nameable links, and timestamps blanked. The deterministic option keeps the keys of the replacement maps in a fixed order.
func describe(c check.Checker, e *gurps.Entity) []string {
	data, err := jio.Marshal(e, jsontext.WithIndent("\t"), json.Deterministic(true))
	c.NoError(err)
	text := volatileFields.ReplaceAllString(string(data), `"${1}":""`)
	return strings.Split(linkKeys.ReplaceAllString(text, `"": {`), "\n")
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

	// Every list's names have title notes to assemble, some of them from modifiers shown in the title.
	c.True(countTitled(e.Traits) > 0, "traits with title notes")
	c.True(countTitled(e.Skills) > 0, "skills with title notes")
	c.True(countTitled(e.Spells) > 0, "spells with title notes")
	c.True(countTitled(e.CarriedEquipment) > 0, "carried equipment with title notes")
	c.True(countTitled(e.OtherEquipment) > 0, "other equipment with title notes")
	var first *gurps.Trait
	gurps.Traverse(func(t *gurps.Trait) bool {
		if t.Name == "Trait 1" || strings.HasPrefix(t.Name, "Trait 1 ") {
			first = t
		}
		return first != nil
	}, false, true, e.Traits...)
	c.True(first != nil && strings.Contains(first.String(), "Modifier 1"), "a trait modifier shown in the title: %v",
		first)

	// Every container starts open, so a sheet built from the entity shows every row.
	checkOpen(c, "trait", e.Traits)
	checkOpen(c, "skill", e.Skills)
	checkOpen(c, "spell", e.Spells)
	checkOpen(c, "carried equipment", e.CarriedEquipment)
	checkOpen(c, "other equipment", e.OtherEquipment)
	checkOpen(c, "note", e.Notes)

	// Every fourth trait's linked modifier code takes the trait's answer, where it has one.
	var linked []*gurps.Trait
	gurps.Traverse(func(t *gurps.Trait) bool {
		if len(t.NameableLinks) != 0 {
			linked = append(linked, t)
		}
		return false
	}, false, true, e.Traits...)
	c.Equal((size.Traits+3)/4, len(linked), "traits with linked modifier codes")
	if len(linked) != 0 {
		c.Equal("Focus on Judo", linked[0].Modifiers[len(linked[0].Modifiers)-1].NameWithReplacements())
	}
}

// Every row's parent must be the container it sits in, or anything that climbs the tree from a generated row, such as
// a delete or a move, acts on the wrong container.
func TestNewLargeEntityParentsMatchContainers(t *testing.T) {
	c := check.New(t)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.1)
	e := gurpstest.NewLargeEntity(&size)
	checkParents(c, "trait", e.Traits, nil)
	checkParents(c, "skill", e.Skills, nil)
	checkParents(c, "spell", e.Spells, nil)
	checkParents(c, "carried equipment", e.CarriedEquipment, nil)
	checkParents(c, "other equipment", e.OtherEquipment, nil)
	checkParents(c, "note", e.Notes, nil)
}

// A size with fewer traits than reactions and conditional modifiers must still yield the number of each asked for, so
// that a small benchmark still has both derived lists on its page.
func TestNewLargeEntityDerivedListsOutnumberTraits(t *testing.T) {
	c := check.New(t)
	size := gurpstest.LargeEntitySize{Traits: 1, Reactions: 3, ConditionalModifiers: 2}
	e := gurpstest.NewLargeEntity(&size)
	c.Equal(size.Reactions, countConditionalModifierRows(e.Reactions()), "reactions")
	c.Equal(size.ConditionalModifiers, countConditionalModifierRows(e.ConditionalModifiers()), "conditional modifiers")
}

// The same size must yield the same character, apart from the IDs and timestamps, so that benchmark runs compare like
// with like. Comparing the saved form covers every name, value, weapon and nesting at once.
func TestNewLargeEntityIsDeterministic(t *testing.T) {
	c := check.New(t)
	size := gurpstest.DefaultLargeEntitySize.Scaled(0.05)
	first := describe(c, gurpstest.NewLargeEntity(&size))
	second := describe(c, gurpstest.NewLargeEntity(&size))
	for i := range min(len(first), len(second)) {
		if first[i] != second[i] {
			c.Equal(first[i], second[i], "line %d", i+1)
			return
		}
	}
	c.Equal(len(first), len(second), "lines")
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
