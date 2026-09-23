// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

func newTestSkill(name string, points fxp.Int, techLevel *string) *gurps.Skill {
	s := gurps.NewSkill(nil, nil, false)
	s.Name = name
	s.Specialization = ""
	s.Points = points
	s.TechLevel = techLevel
	return s
}

func newTestSpell(name string, points fxp.Int, techLevel *string) *gurps.Spell {
	s := gurps.NewSpell(nil, nil, false)
	s.Name = name
	s.Points = points
	s.TechLevel = techLevel
	return s
}

// TestResolveEmptyTechLevel verifies the tech level substitution shared by the drop handlers and the merge: only an
// empty (but present) tech level is replaced, and one that is absent stays absent rather than being introduced.
func TestResolveEmptyTechLevel(t *testing.T) {
	c := check.New(t)

	skill := newTestSkill("Guns", fxp.FromInteger(1), nil)
	resolveEmptyTechLevel(skill, "3")
	c.Nil(skill.TechLevel, "a skill without a TL must not gain one")

	skill = newTestSkill("Guns", fxp.FromInteger(1), new(""))
	resolveEmptyTechLevel(skill, "3")
	c.Equal("3", *skill.TechLevel, "an empty TL must be replaced with the default")

	spell := newTestSpell("Fireball", fxp.FromInteger(1), new("8"))
	resolveEmptyTechLevel(spell, "3")
	c.Equal("8", *spell.TechLevel, "a TL that is already set must be left alone")
}

func TestMergeSkillPoints(t *testing.T) {
	c := check.New(t)

	t.Run("identical skill without a TL merges points", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Brawling", fxp.FromInteger(4), nil)}
		incoming := []*gurps.Skill{newTestSkill("Brawling", fxp.FromInteger(2), nil)}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
		c.Equal(true, selMap[existing[0].ID()])
	})

	t.Run("identical skill with a matching TL merges points", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), new("8"))}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), new("8"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("skill with a differing TL is not merged", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), new("8"))}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), new("9"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// This reproduces the reported bug: on the first apply the skill's empty TL is resolved to the entity's TL, so on
	// the second apply the incoming skill (still empty TL) must resolve to that same TL and merge rather than
	// duplicate.
	t.Run("incoming empty TL resolves to the entity TL and merges", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), new("3"))}
		incoming := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), new(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(2), existing[0].Points)
		c.Equal(true, selMap[existing[0].ID()])
	})

	// A surviving (unmatched) incoming skill with an empty TL must still have that TL resolved, so a subsequent apply
	// will merge with it.
	t.Run("surviving incoming empty TL is resolved to the entity TL", func(_ *testing.T) {
		var existing []*gurps.Skill
		incoming := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), new(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal("3", *remaining[0].TechLevel)
	})

	// Both variants share a hash, since the TL is not part of it. The matching TL9 variant is listed first, so a lookup
	// keyed only by hash that retained just the last-seen (TL8) variant would fail to merge; every candidate for the
	// hash must be considered.
	t.Run("matches the correct TL variant when several share a hash", func(_ *testing.T) {
		existingTL9 := newTestSkill("Guns", fxp.FromInteger(1), new("9"))
		existingTL8 := newTestSkill("Guns", fxp.FromInteger(4), new("8"))
		existing := []*gurps.Skill{existingTL9, existingTL8}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), new("9"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(4), existingTL8.Points)
		c.Equal(fxp.FromInteger(3), existingTL9.Points)
		c.Equal(true, selMap[existingTL9.ID()])
		c.Equal(false, selMap[existingTL8.ID()])
	})

	t.Run("a skill with a TL does not merge with one lacking a TL", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), new("8"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// The merge match includes the nameable replacements; this is why they must be resolved before the merge runs.
	t.Run("skills with matching replacements merge", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		existing[0].Replacements = map[string]string{"1": "Rifle"}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), nil)}
		incoming[0].Replacements = map[string]string{"1": "Rifle"}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("skills with differing replacements are not merged", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		existing[0].Replacements = map[string]string{"1": "Rifle"}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), nil)}
		incoming[0].Replacements = map[string]string{"1": "Pistol"}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// Two identical incoming entries (a template containing the same skill twice) collapse into one even when there is
	// nothing on the sheet to merge into.
	t.Run("identical incoming skills merge with each other", func(_ *testing.T) {
		first := newTestSkill("Administration", fxp.FromInteger(1), nil)
		first.Replacements = map[string]string{"what": "Empire"}
		second := newTestSkill("Administration", fxp.FromInteger(1), nil)
		second.Replacements = map[string]string{"what": "Empire"}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(nil, []*gurps.Skill{first, second}, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(first, remaining[0])
		c.Equal(fxp.FromInteger(2), first.Points)
	})

	t.Run("incoming skills with differing replacements stay separate", func(_ *testing.T) {
		first := newTestSkill("Administration", fxp.FromInteger(1), nil)
		first.Replacements = map[string]string{"what": "Empire"}
		second := newTestSkill("Administration", fxp.FromInteger(1), nil)
		second.Replacements = map[string]string{"what": "Guild"}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(nil, []*gurps.Skill{first, second}, "3", selMap)
		c.Equal(2, len(remaining))
	})
}

func TestMergeSpellPoints(t *testing.T) {
	c := check.New(t)

	t.Run("identical spell with a matching TL merges points", func(_ *testing.T) {
		existing := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(4), new("3"))}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(2), new("3"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("matches the correct TL variant when several share a hash", func(_ *testing.T) {
		// The matching TL4 variant is listed first so a lookup keyed only by hash would retain the TL3 variant and
		// fail to merge; the merge must consider every candidate for the hash.
		existingTL4 := newTestSpell("Fireball", fxp.FromInteger(1), new("4"))
		existingTL3 := newTestSpell("Fireball", fxp.FromInteger(4), new("3"))
		existing := []*gurps.Spell{existingTL4, existingTL3}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(2), new("4"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(4), existingTL3.Points)
		c.Equal(fxp.FromInteger(3), existingTL4.Points)
	})

	t.Run("incoming empty TL resolves to the entity TL and merges", func(_ *testing.T) {
		existing := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(1), new("3"))}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(1), new(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(2), existing[0].Points)
	})

	t.Run("identical incoming spells merge with each other", func(_ *testing.T) {
		first := newTestSpell("Fireball", fxp.FromInteger(1), nil)
		second := newTestSpell("Fireball", fxp.FromInteger(1), nil)
		selMap := make(map[tid.TID]bool)
		remaining := mergePoints(nil, []*gurps.Spell{first, second}, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(first, remaining[0])
		c.Equal(fxp.FromInteger(2), first.Points)
	})
}

// placeSkills places the added skills after the existing ones already on a sheet, the way any rows arriving on a sheet
// are placed, returning the sheet's skills table afterwards.
func placeSkills(t *testing.T, existing []*gurps.Skill, added ...*gurps.Skill) (*gurps.Entity, *unison.Table[*Node[*gurps.Skill]]) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	entity.Skills = existing
	sheet.Rebuild(true)
	part := applyPart[*gurps.Skill]{table: sheet.Skills.Table, rows: added, index: -1}
	part.place(true)
	return entity, liveTable(sheet.Skills.Table)
}

// Placing rows merges any that duplicate a row already present into that row, while the rows already present never
// merge with each other.
func TestPlaceMergesIntoExistingRows(t *testing.T) {
	c := check.New(t)

	t.Run("adding an identical skill merges into the existing one", func(t *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		added := newTestSkill("Brawling", fxp.FromInteger(2), nil)
		entity, table := placeSkills(t, []*gurps.Skill{existing}, added)
		c.Equal(1, len(entity.Skills))
		c.Equal(existing, table.RootRows()[0].Data())
		c.Equal(fxp.FromInteger(6), existing.Points)
	})

	t.Run("adding a distinct skill keeps both rows", func(t *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		added := newTestSkill("Climbing", fxp.FromInteger(2), nil)
		entity, _ := placeSkills(t, []*gurps.Skill{existing}, added)
		c.Equal(2, len(entity.Skills))
		c.Equal(fxp.FromInteger(4), existing.Points)
	})

	// Two identical existing rows must not merge with each other, and an added row must merge into the matching
	// candidate among the rows sharing its hash.
	t.Run("added row merges into an existing row that shares a hash", func(t *testing.T) {
		existingTL8 := newTestSkill("Guns", fxp.FromInteger(4), new("8"))
		existingTL9 := newTestSkill("Guns", fxp.FromInteger(1), new("9"))
		added := newTestSkill("Guns", fxp.FromInteger(2), new("9"))
		entity, _ := placeSkills(t, []*gurps.Skill{existingTL8, existingTL9}, added)
		c.Equal(2, len(entity.Skills))
		c.Equal(fxp.FromInteger(4), existingTL8.Points)
		c.Equal(fxp.FromInteger(3), existingTL9.Points)
	})

	// A row dropped into a container arrives already parented to it, though not yet among its children, and the merge
	// used to take it out of the container's children -- where it wasn't -- rather than out of the rows to be placed,
	// so its points were added to the existing row and the row was placed as well.
	t.Run("a skill dropped into a container merges without also being placed", func(t *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		group := gurps.NewSkill(nil, nil, true)
		group.Name = "Combat Skills"
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		entity.Skills = []*gurps.Skill{existing, group}
		sheet.Rebuild(true)
		added := newTestSkill("Brawling", fxp.FromInteger(2), nil)
		added.SetParent(group)
		part := applyPart[*gurps.Skill]{table: sheet.Skills.Table, rows: []*gurps.Skill{added}, parent: group, index: 0}
		part.place(true)
		c.Equal(fxp.FromInteger(6), existing.Points, "the existing skill must absorb the dropped skill's points")
		c.Equal(0, len(group.Children), "the merged skill must not also be placed into the container")
		c.Equal(0, len(part.placed))
		table, _ := part.changed()
		c.NotNil(table, "a merge changes the row merged into, so the change must still be reported")
	})

	// Regression test for #1066: a row nested inside an added container (as when a template is added to the sheet by
	// dragging it in) merges into an identical existing row, and the now-redundant nested row must leave the table's
	// view immediately rather than remaining visible until something else causes the table to reload.
	t.Run("a skill inside an added container merges and its row leaves the view", func(t *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		container := gurps.NewSkill(nil, nil, true)
		container.Name = "Combat Skills"
		child := newTestSkill("Brawling", fxp.FromInteger(2), nil)
		child.SetParent(container)
		container.Children = []*gurps.Skill{child}
		_, table := placeSkills(t, []*gurps.Skill{existing}, container)
		c.Equal(fxp.FromInteger(6), existing.Points, "the existing skill must absorb the nested skill's points")
		c.Equal(0, len(container.Children), "the merged child must be pruned from the container's data")
		roots := table.RootRows()
		c.Equal(2, len(roots), "the existing skill and the container must both remain")
		c.Equal(0, len(roots[1].Children()), "the merged child must no longer be in the view")
		c.Equal(2, table.LastRowIndex()+1, "the merged child's row must be gone from the table")
		sel := table.CopySelectionMap()
		c.Equal(true, sel[existing.ID()], "the merged-into skill must be selected")
		c.Equal(true, sel[container.ID()], "the surviving container must remain selected")
	})

	// Same as above, but nested two container levels deep, so that both the merge traversal and the view refresh are
	// shown to handle arbitrary nesting rather than just direct children of an added container.
	t.Run("a skill nested two levels deep in an added container merges and its row leaves the view", func(t *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		container := gurps.NewSkill(nil, nil, true)
		container.Name = "Combat Skills"
		subContainer := gurps.NewSkill(nil, container, true)
		subContainer.Name = "Unarmed"
		container.Children = []*gurps.Skill{subContainer}
		child := newTestSkill("Brawling", fxp.FromInteger(2), nil)
		child.SetParent(subContainer)
		subContainer.Children = []*gurps.Skill{child}
		_, table := placeSkills(t, []*gurps.Skill{existing}, container)
		c.Equal(fxp.FromInteger(6), existing.Points, "the existing skill must absorb the deeply nested skill's points")
		c.Equal(0, len(subContainer.Children), "the merged child must be pruned from the sub-container's data")
		c.Equal(1, len(container.Children), "the sub-container itself must remain in the added container")
		c.Equal(3, table.LastRowIndex()+1, "the merged child's row must be gone from the table")
		roots := table.RootRows()
		c.Equal(2, len(roots), "the existing skill and the container must both remain")
		containerChildren := roots[1].Children()
		c.Equal(1, len(containerChildren), "the sub-container's row must remain in the view")
		c.Equal(0, len(containerChildren[0].Children()), "the merged child must no longer be in the view")
	})
}
