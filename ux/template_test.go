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

func ptr(s string) *string { return &s }

func newTestTrait(name string, levels fxp.Int, canLevel bool) *gurps.Trait {
	t := gurps.NewTrait(nil, nil, false)
	t.Name = name
	t.CanLevel = canLevel
	if canLevel {
		t.PointsPerLevel = fxp.FromInteger(5)
		t.Levels = levels
	}
	return t
}

func newTestTraitModifier(name string, disabled bool) *gurps.TraitModifier {
	m := gurps.NewTraitModifier(nil, nil, false)
	m.Name = name
	m.Disabled = disabled
	return m
}

// TestMergeSkillPoints verifies that applying a template's skills onto an existing set folds the points of identical
// skills together, correctly distinguishing skills that differ only by tech level.
func TestMergeSkillPoints(t *testing.T) {
	c := check.New(t)

	t.Run("identical skill without a TL merges points", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Brawling", fxp.FromInteger(4), nil)}
		incoming := []*gurps.Skill{newTestSkill("Brawling", fxp.FromInteger(2), nil)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
		c.Equal(true, selMap[existing[0].ID()])
	})

	t.Run("identical skill with a matching TL merges points", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), ptr("8"))}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), ptr("8"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("skill with a differing TL is not merged", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), ptr("8"))}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), ptr("9"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// This reproduces the reported bug: on the first apply the skill's empty TL is resolved to the entity's TL, so on
	// the second apply the incoming skill (still empty TL) must resolve to that same TL and merge rather than duplicate.
	t.Run("incoming empty TL resolves to the entity TL and merges", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), ptr("3"))}
		incoming := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), ptr(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(2), existing[0].Points)
		c.Equal(true, selMap[existing[0].ID()])
	})

	// A surviving (unmatched) incoming skill with an empty TL must still have that TL resolved, so a subsequent apply
	// will merge with it.
	t.Run("surviving incoming empty TL is resolved to the entity TL", func(_ *testing.T) {
		var existing []*gurps.Skill
		incoming := []*gurps.Skill{newTestSkill("Architecture", fxp.FromInteger(1), ptr(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal("3", *remaining[0].TechLevel)
	})

	// Both variants share a hash (the TL is not part of the hash). The incoming TL9 skill must merge with the existing
	// TL9 skill even when the TL8 variant happens to be considered first.
	t.Run("matches the correct TL variant when several share a hash", func(_ *testing.T) {
		// The matching TL9 variant is listed first so that a lookup keyed only by hash (retaining just the
		// last-seen TL8 variant) would fail to merge; the merge must consider every candidate for the hash.
		existingTL9 := newTestSkill("Guns", fxp.FromInteger(1), ptr("9"))
		existingTL8 := newTestSkill("Guns", fxp.FromInteger(4), ptr("8"))
		existing := []*gurps.Skill{existingTL9, existingTL8}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), ptr("9"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(4), existingTL8.Points)
		c.Equal(fxp.FromInteger(3), existingTL9.Points)
		c.Equal(true, selMap[existingTL9.ID()])
		c.Equal(false, selMap[existingTL8.ID()])
	})

	// A skill with a TL must not merge into a same-named skill that has no TL at all.
	t.Run("a skill with a TL does not merge with one lacking a TL", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), ptr("8"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// The merge match includes the nameable replacements; this is why they must be resolved before the merge runs.
	const rifle = "Rifle"
	t.Run("skills with matching replacements merge", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		existing[0].Replacements = map[string]string{"1": rifle}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), nil)}
		incoming[0].Replacements = map[string]string{"1": rifle}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("skills with differing replacements are not merged", func(_ *testing.T) {
		existing := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(4), nil)}
		existing[0].Replacements = map[string]string{"1": rifle}
		incoming := []*gurps.Skill{newTestSkill("Guns", fxp.FromInteger(2), nil)}
		incoming[0].Replacements = map[string]string{"1": "Pistol"}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(existing, incoming, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(4), existing[0].Points)
	})

	// Two identical entries within the incoming set (e.g. a template containing the same skill twice) collapse into one
	// even when there is nothing on the sheet to merge into.
	const (
		what   = "what"
		empire = "Empire"
	)
	t.Run("identical incoming skills merge with each other", func(_ *testing.T) {
		first := newTestSkill("Administration", fxp.FromInteger(1), nil)
		first.Replacements = map[string]string{what: empire}
		second := newTestSkill("Administration", fxp.FromInteger(1), nil)
		second.Replacements = map[string]string{what: empire}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(nil, []*gurps.Skill{first, second}, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(first, remaining[0])
		c.Equal(fxp.FromInteger(2), first.Points)
	})

	t.Run("incoming skills with differing replacements stay separate", func(_ *testing.T) {
		first := newTestSkill("Administration", fxp.FromInteger(1), nil)
		first.Replacements = map[string]string{what: empire}
		second := newTestSkill("Administration", fxp.FromInteger(1), nil)
		second.Replacements = map[string]string{what: "Guild"}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSkillPoints(nil, []*gurps.Skill{first, second}, "3", selMap)
		c.Equal(2, len(remaining))
	})
}

// TestMergeSpellPoints verifies the same folding behavior for spells, including the tech-level distinction.
func TestMergeSpellPoints(t *testing.T) {
	c := check.New(t)

	t.Run("identical spell with a matching TL merges points", func(_ *testing.T) {
		existing := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(4), ptr("3"))}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(2), ptr("3"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSpellPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(6), existing[0].Points)
	})

	t.Run("matches the correct TL variant when several share a hash", func(_ *testing.T) {
		// The matching TL4 variant is listed first so a lookup keyed only by hash would retain the TL3 variant and
		// fail to merge; the merge must consider every candidate for the hash.
		existingTL4 := newTestSpell("Fireball", fxp.FromInteger(1), ptr("4"))
		existingTL3 := newTestSpell("Fireball", fxp.FromInteger(4), ptr("3"))
		existing := []*gurps.Spell{existingTL4, existingTL3}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(2), ptr("4"))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSpellPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(4), existingTL3.Points)
		c.Equal(fxp.FromInteger(3), existingTL4.Points)
	})

	t.Run("incoming empty TL resolves to the entity TL and merges", func(_ *testing.T) {
		existing := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(1), ptr("3"))}
		incoming := []*gurps.Spell{newTestSpell("Fireball", fxp.FromInteger(1), ptr(""))}
		selMap := make(map[tid.TID]bool)
		remaining := mergeSpellPoints(existing, incoming, "3", selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(2), existing[0].Points)
	})

	t.Run("identical incoming spells merge with each other", func(_ *testing.T) {
		first := newTestSpell("Fireball", fxp.FromInteger(1), nil)
		second := newTestSpell("Fireball", fxp.FromInteger(1), nil)
		selMap := make(map[tid.TID]bool)
		remaining := mergeSpellPoints(nil, []*gurps.Spell{first, second}, "3", selMap)
		c.Equal(1, len(remaining))
		c.Equal(first, remaining[0])
		c.Equal(fxp.FromInteger(2), first.Points)
	})
}

func newSkillTable(skills ...*gurps.Skill) (*unison.Table[*Node[*gurps.Skill]], []*Node[*gurps.Skill]) {
	table := unison.NewTable[*Node[*gurps.Skill]](&unison.SimpleTableModel[*Node[*gurps.Skill]]{})
	nodes := make([]*Node[*gurps.Skill], len(skills))
	for i, s := range skills {
		nodes[i] = NewNode(table, nil, s, false)
	}
	table.SetRootRows(nodes)
	return table, nodes
}

// TestMergeAddedRows verifies that adding a skill to a sheet (as when dragging or copying one in, where the
// added rows are the selection) folds its points into an identical existing row and removes the redundant new row.
func TestMergeAddedRows(t *testing.T) {
	c := check.New(t)

	t.Run("adding an identical skill merges into the existing one", func(_ *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		added := newTestSkill("Brawling", fxp.FromInteger(2), nil)
		table, nodes := newSkillTable(existing, added)
		table.SetSelectionMap(map[tid.TID]bool{nodes[1].ID(): true})
		MergeAddedRows(table)
		roots := table.RootRows()
		c.Equal(1, len(roots))
		c.Equal(existing, roots[0].Data())
		c.Equal(fxp.FromInteger(6), existing.Points)
	})

	t.Run("adding a distinct skill keeps both rows", func(_ *testing.T) {
		existing := newTestSkill("Brawling", fxp.FromInteger(4), nil)
		added := newTestSkill("Climbing", fxp.FromInteger(2), nil)
		table, nodes := newSkillTable(existing, added)
		table.SetSelectionMap(map[tid.TID]bool{nodes[1].ID(): true})
		MergeAddedRows(table)
		c.Equal(2, len(table.RootRows()))
		c.Equal(fxp.FromInteger(4), existing.Points)
	})

	// Only the newly-added (selected) row is treated as incoming; two identical existing rows must not merge with each
	// other, and an added row must merge into the first matching existing candidate.
	t.Run("added row merges into an existing row that shares a hash", func(_ *testing.T) {
		existingTL8 := newTestSkill("Guns", fxp.FromInteger(4), ptr("8"))
		existingTL9 := newTestSkill("Guns", fxp.FromInteger(1), ptr("9"))
		added := newTestSkill("Guns", fxp.FromInteger(2), ptr("9"))
		table, nodes := newSkillTable(existingTL8, existingTL9, added)
		table.SetSelectionMap(map[tid.TID]bool{nodes[2].ID(): true})
		MergeAddedRows(table)
		c.Equal(2, len(table.RootRows()))
		c.Equal(fxp.FromInteger(4), existingTL8.Points)
		c.Equal(fxp.FromInteger(3), existingTL9.Points)
	})

	// Regression test for #1066: a leveled trait nested inside an added container (as when a template is added to the
	// sheet by dragging it in) merges into an identical existing trait, and the now-redundant nested row must leave the
	// table's view immediately. It used to be pruned from the data only, remaining visible until something else caused
	// the table to reload, because the merge bailed out early when the top-level row count was unchanged.
	t.Run("a trait inside an added container merges and its row leaves the view", func(_ *testing.T) {
		existing := newTestTrait("Increased Dexterity", fxp.FromInteger(1), true)
		container := gurps.NewTrait(nil, nil, true)
		container.Name = "Infernal"
		child := newTestTrait("Increased Dexterity", fxp.FromInteger(2), true)
		child.SetParent(container)
		container.Children = []*gurps.Trait{child}
		table := unison.NewTable[*Node[*gurps.Trait]](&unison.SimpleTableModel[*Node[*gurps.Trait]]{})
		existingNode := NewNode(table, nil, existing, false)
		containerNode := NewNode(table, nil, container, false)
		table.SetRootRows([]*Node[*gurps.Trait]{existingNode, containerNode})
		c.Equal(3, table.LastRowIndex()+1, "the open container's child must be showing before the merge")
		table.SetSelectionMap(map[tid.TID]bool{containerNode.ID(): true})
		MergeAddedRows(table)
		c.Equal(fxp.FromInteger(3), existing.Levels, "the existing trait must absorb the nested trait's levels")
		c.Equal(0, len(container.Children), "the merged child must be pruned from the container's data")
		roots := table.RootRows()
		c.Equal(2, len(roots), "the existing trait and the container must both remain")
		c.Equal(0, len(roots[1].Children()), "the merged child must no longer be in the view")
		c.Equal(2, table.LastRowIndex()+1, "the merged child's row must be gone from the table")
		sel := table.CopySelectionMap()
		c.Equal(true, sel[existing.ID()], "the merged-into trait must be selected")
		c.Equal(true, sel[container.ID()], "the surviving container must remain selected")
	})

	// Same as above, but with the merged trait nested two container levels deep, to ensure both the merge traversal and
	// the view refresh handle arbitrary nesting rather than just direct children of an added container.
	t.Run("a trait nested two levels deep in an added container merges and its row leaves the view", func(_ *testing.T) {
		existing := newTestTrait("Increased Dexterity", fxp.FromInteger(1), true)
		container := gurps.NewTrait(nil, nil, true)
		container.Name = "Infernal"
		subContainer := gurps.NewTrait(nil, container, true)
		subContainer.Name = "Attributes"
		container.Children = []*gurps.Trait{subContainer}
		child := newTestTrait("Increased Dexterity", fxp.FromInteger(2), true)
		child.SetParent(subContainer)
		subContainer.Children = []*gurps.Trait{child}
		table := unison.NewTable[*Node[*gurps.Trait]](&unison.SimpleTableModel[*Node[*gurps.Trait]]{})
		existingNode := NewNode(table, nil, existing, false)
		containerNode := NewNode(table, nil, container, false)
		table.SetRootRows([]*Node[*gurps.Trait]{existingNode, containerNode})
		c.Equal(4, table.LastRowIndex()+1, "all nested rows must be showing before the merge")
		table.SetSelectionMap(map[tid.TID]bool{containerNode.ID(): true})
		MergeAddedRows(table)
		c.Equal(fxp.FromInteger(3), existing.Levels, "the existing trait must absorb the deeply nested trait's levels")
		c.Equal(0, len(subContainer.Children), "the merged child must be pruned from the sub-container's data")
		c.Equal(1, len(container.Children), "the sub-container itself must remain in the added container")
		c.Equal(3, table.LastRowIndex()+1, "the merged child's row must be gone from the table")
		roots := table.RootRows()
		c.Equal(2, len(roots), "the existing trait and the container must both remain")
		containerChildren := roots[1].Children()
		c.Equal(1, len(containerChildren), "the sub-container's row must remain in the view")
		c.Equal(0, len(containerChildren[0].Children()), "the merged child must no longer be in the view")
	})
}

// TestMergeTraitLevels verifies that only leveled traits merge, that their levels are combined, and that differing
// modifiers keep them separate.
func TestMergeTraitLevels(t *testing.T) {
	c := check.New(t)

	t.Run("identical leveled traits merge their levels", func(_ *testing.T) {
		existing := []*gurps.Trait{newTestTrait("Damage Resistance", fxp.FromInteger(2), true)}
		incoming := []*gurps.Trait{newTestTrait("Damage Resistance", fxp.FromInteger(3), true)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels(existing, incoming, selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(5), existing[0].Levels)
		c.Equal(true, selMap[existing[0].ID()])
	})

	t.Run("non-leveled traits are never merged", func(_ *testing.T) {
		existing := []*gurps.Trait{newTestTrait("Combat Reflexes", fxp.FromInteger(0), false)}
		incoming := []*gurps.Trait{newTestTrait("Combat Reflexes", fxp.FromInteger(0), false)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels(existing, incoming, selMap)
		c.Equal(1, len(remaining))
	})

	t.Run("leveled traits with matching modifiers merge", func(_ *testing.T) {
		existing := newTestTrait("Damage Resistance", fxp.FromInteger(2), true)
		existing.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Hardened", false)}
		incoming := newTestTrait("Damage Resistance", fxp.FromInteger(1), true)
		incoming.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Hardened", false)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels([]*gurps.Trait{existing}, []*gurps.Trait{incoming}, selMap)
		c.Equal(0, len(remaining))
		c.Equal(fxp.FromInteger(3), existing.Levels)
	})

	t.Run("leveled traits with differing modifiers stay separate", func(_ *testing.T) {
		existing := newTestTrait("Damage Resistance", fxp.FromInteger(2), true)
		existing.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Hardened", false)}
		incoming := newTestTrait("Damage Resistance", fxp.FromInteger(1), true)
		incoming.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Tough Skin", false)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels([]*gurps.Trait{existing}, []*gurps.Trait{incoming}, selMap)
		c.Equal(1, len(remaining))
		c.Equal(fxp.FromInteger(2), existing.Levels)
	})

	// The same modifier enabled on one trait but disabled on the other means the traits are not identical.
	t.Run("a disabled modifier prevents merging", func(_ *testing.T) {
		existing := newTestTrait("Damage Resistance", fxp.FromInteger(2), true)
		existing.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Hardened", false)}
		incoming := newTestTrait("Damage Resistance", fxp.FromInteger(1), true)
		incoming.Modifiers = []*gurps.TraitModifier{newTestTraitModifier("Hardened", true)}
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels([]*gurps.Trait{existing}, []*gurps.Trait{incoming}, selMap)
		c.Equal(1, len(remaining))
	})

	t.Run("identical incoming leveled traits merge with each other", func(_ *testing.T) {
		first := newTestTrait("Damage Resistance", fxp.FromInteger(1), true)
		second := newTestTrait("Damage Resistance", fxp.FromInteger(2), true)
		selMap := make(map[tid.TID]bool)
		remaining := mergeTraitLevels(nil, []*gurps.Trait{first, second}, selMap)
		c.Equal(1, len(remaining))
		c.Equal(first, remaining[0])
		c.Equal(fxp.FromInteger(3), first.Levels)
	})
}
