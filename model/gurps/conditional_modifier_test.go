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
	"github.com/richardwilkes/gcs/v5/model/kinds"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
)

// addReaction adds a trait carrying one reaction bonus with the given situation, group and amount to the entity.
func addReaction(e *Entity, name, situation, group string, amt fxp.Int) *Trait {
	bonus := NewReactionBonus()
	bonus.Situation = situation
	bonus.Group = group
	bonus.Amount = amt
	return addTraitWithFeatures(e, name, bonus)
}

// addGroupedConditionalModifier adds a trait carrying one conditional modifier bonus with the given situation, group and
// amount to the entity.
func addGroupedConditionalModifier(e *Entity, name, situation, group string, amt fxp.Int) *Trait {
	bonus := NewConditionalModifierBonus()
	bonus.Situation = situation
	bonus.Group = group
	bonus.Amount = amt
	return addTraitWithFeatures(e, name, bonus)
}

// rowNames returns the From of each row, with a container's children listed in brackets after it.
func rowNames(list []*ConditionalModifier) []string {
	names := make([]string, 0, len(list))
	for _, one := range list {
		name := one.From
		if one.Container() {
			name += "["
			for i, child := range one.Children {
				if i != 0 {
					name += ","
				}
				name += child.From
			}
			name += "]"
		}
		names = append(names, name)
	}
	return names
}

// withGroupContainersOnSort sets the general "group containers when sorting" setting for the duration of the test.
func withGroupContainersOnSort(t *testing.T, on bool) {
	t.Helper()
	general := GlobalSettings().General
	was := general.GroupContainersOnSort
	general.GroupContainersOnSort = on
	t.Cleanup(func() { general.GroupContainersOnSort = was })
}

// TestConditionalModifierGroupsProduceContainerRows verifies that bonuses with a group are filed as the children of a
// container row named for the group, wired to their parent, while ungrouped bonuses stay top-level, for both the
// reactions and the conditional modifiers tables.
func TestConditionalModifierGroupsProduceContainerRows(t *testing.T) {
	withGroupContainersOnSort(t, true)
	e := NewEntity()
	addReaction(e, "Charisma", "from everyone", "", fxp.One)
	addReaction(e, "Fearsome", "from monsters", "Combat", fxp.Two)
	addReaction(e, "Sword Master", "from swordsmen", "Combat", fxp.Three)
	addGroupedConditionalModifier(e, "Gadgeteer", "to build things", "", fxp.One)
	addGroupedConditionalModifier(e, "Lucky", "to avoid traps", "Adventuring", fxp.Two)
	addGroupedConditionalModifier(e, "Wary", "to notice ambushes", "Adventuring", fxp.Three)
	e.Recalculate()

	for _, tc := range []struct {
		name  string
		rows  []*ConditionalModifier
		group string
		leaf  string
	}{
		{name: "reactions", rows: e.Reactions(), group: "Combat", leaf: "from everyone"},
		{name: "conditional modifiers", rows: e.ConditionalModifiers(), group: "Adventuring", leaf: "to build things"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			c.Equal(2, len(tc.rows), "a group container and an ungrouped entry")
			if len(tc.rows) != 2 {
				return
			}
			group, leaf := tc.rows[0], tc.rows[1]
			c.True(group.Container(), "the group is a container")
			c.True(group.HasChildren(), "the group has children")
			c.Equal(tc.group, group.From, "the container carries the group name")
			c.Equal(2, len(group.NodeChildren()), "both grouped bonuses are filed under it")
			for _, child := range group.NodeChildren() {
				c.True(child.Parent() == group, "%s: a child knows its group", child.From)
				c.False(child.Container(), "%s: a child is not a container", child.From)
				c.Equal(tc.group, child.GroupName(), "%s: a child reports its group's name", child.From)
			}
			c.False(leaf.Container(), "the ungrouped entry is not a container")
			c.Equal(tc.leaf, leaf.From, "the ungrouped entry is top-level")
			c.Nil(leaf.Parent(), "the ungrouped entry has no parent")
			c.Equal("", leaf.GroupName(), "the ungrouped entry reports no group")
			c.Equal(tc.group, group.GroupName(), "a container reports its own name as its group")
		})
	}
}

// TestConditionalModifierGroupCellDataLeavesTheValueColumnBlank verifies that a group container shows only its name:
// the modifier column is empty, with no tooltip, since the members apply in different situations and don't sum.
func TestConditionalModifierGroupCellDataLeavesTheValueColumnBlank(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewConditionalModifierGroup(e.ID, BlockReactionsKey, "Combat")
	var data CellData
	group.CellData(ConditionalModifierValueColumn, &data)
	c.Equal("", data.Primary, "the modifier column is blank for a group")
	c.Equal("", data.Tooltip, "there is no per-source tooltip for a group")
	data = CellData{}
	group.CellData(ConditionalModifierDescriptionColumn, &data)
	c.Equal("Combat", data.Primary, "the description column shows the group name")
	c.Equal(fxp.Int(0), group.Total(), "a group has no total of its own")
	c.Equal("Combat", group.String(), "a group's string form is just its name")
	c.Equal("Conditional Modifier Container", group.Kind())

	leaf := NewConditionalModifier("from trait X", "from everyone", fxp.Two)
	data = CellData{}
	leaf.CellData(ConditionalModifierValueColumn, &data)
	c.Equal("+2", data.Primary, "a modifier still shows its total")
	c.Equal("+2 from trait X", data.Tooltip, "a modifier still lists its sources")
	c.Equal("+2 from everyone", leaf.String())
	c.Equal("Conditional Modifier", leaf.Kind())
}

// TestConditionalModifierOrderingHonorsGroupContainersOnSort verifies the two orderings: with the general setting on,
// the groups come first, sorted by name, followed by the ungrouped entries; with it off, groups and ungrouped entries
// are intermixed by name. Children are always sorted by name within their group.
func TestConditionalModifierOrderingHonorsGroupContainersOnSort(t *testing.T) {
	e := NewEntity()
	addReaction(e, "A", "from apes", "", fxp.One)
	addReaction(e, "B", "from zebras", "", fxp.One)
	addReaction(e, "C", "from wolves", "Beasts", fxp.One)
	addReaction(e, "D", "from bears", "Beasts", fxp.One)
	addReaction(e, "E", "from yeti", "Monsters", fxp.One)
	e.Recalculate()

	t.Run("on", func(t *testing.T) {
		withGroupContainersOnSort(t, true)
		check.New(t).Equal([]string{"Beasts[from bears,from wolves]", "Monsters[from yeti]", "from apes", "from zebras"},
			rowNames(e.Reactions()))
	})
	t.Run("off", func(t *testing.T) {
		withGroupContainersOnSort(t, false)
		check.New(t).Equal([]string{"Beasts[from bears,from wolves]", "from apes", "from zebras", "Monsters[from yeti]"},
			rowNames(e.Reactions()))
	})
}

// TestConditionalModifierGroupBeforeSameNamedEntry verifies that when a group and an ungrouped entry carry the same
// name and are intermixed, the group is placed first, so the order doesn't depend on map iteration. That must hold
// whatever the entry's total, since a group's total is always zero: a negative entry must not slip ahead of the group.
func TestConditionalModifierGroupBeforeSameNamedEntry(t *testing.T) {
	withGroupContainersOnSort(t, false)
	for _, amt := range []fxp.Int{fxp.One, 0, -fxp.One} {
		t.Run(amt.String(), func(t *testing.T) {
			c := check.New(t)
			e := NewEntity()
			addReaction(e, "A", "Combat", "", amt)
			addReaction(e, "B", "from foes", "Combat", fxp.One)
			e.Recalculate()
			for range 20 {
				c.Equal([]string{"Combat[from foes]", "Combat"}, rowNames(e.Reactions()))
			}
		})
	}
}

// TestConditionalModifierGroupsMergeOnlyWithinTheirGroup verifies that the same situation under two groups and with no
// group yields three separate rows with distinct IDs, and that contributions merge only within a group.
func TestConditionalModifierGroupsMergeOnlyWithinTheirGroup(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	e := NewEntity()
	addReaction(e, "A", "from everyone", "", fxp.One)
	addReaction(e, "B", "from everyone", "Alpha", fxp.Two)
	addReaction(e, "C", "from everyone", "Alpha", fxp.Two)
	addReaction(e, "D", "from everyone", "Beta", fxp.Three)
	e.Recalculate()

	rows := e.Reactions()
	c.Equal([]string{"Alpha[from everyone]", "Beta[from everyone]", "from everyone"}, rowNames(rows))
	if len(rows) != 3 {
		return
	}
	alpha, beta, plain := rows[0].Children[0], rows[1].Children[0], rows[2]
	c.Equal(fxp.Four, alpha.Total(), "the two contributions under Alpha merge")
	c.Equal(fxp.Three, beta.Total())
	c.Equal(fxp.One, plain.Total())
	ids := map[tid.TID]bool{alpha.TID: true, beta.TID: true, plain.TID: true, rows[0].TID: true, rows[1].TID: true}
	c.Equal(5, len(ids), "every row, grouped or not, has its own ID")
}

// TestConditionalModifierTIDsAreStableAndPerTable verifies the ID guarantees the disclosure state and the exporters
// rely on: an ungrouped row's ID is exactly what NewConditionalModifier produces, a group's ID is the same each time the
// rows are rebuilt, and a group of the same name in the other table, or on another entity, has a different ID.
func TestConditionalModifierTIDsAreStableAndPerTable(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	e := NewEntity()
	addReaction(e, "A", "from everyone", "", fxp.One)
	addReaction(e, "B", "from foes", "Combat", fxp.One)
	addGroupedConditionalModifier(e, "C", "to hit", "Combat", fxp.One)
	e.Recalculate()

	reactions := e.Reactions()
	c.Equal([]string{"Combat[from foes]", "from everyone"}, rowNames(reactions))
	if len(reactions) != 2 {
		return
	}
	c.Equal(NewConditionalModifier("ignored", "from everyone", fxp.One).TID, reactions[1].TID,
		"an ungrouped row keeps the ID it had before groups existed")
	c.Equal(reactions[0].TID, e.Reactions()[0].TID, "a group's ID is the same across rebuilds")
	c.True(tid.IsKind(reactions[0].TID, kinds.ConditionalModifierContainer), "a group's ID carries the container kind")
	c.True(tid.IsKind(reactions[0].Children[0].TID, kinds.ConditionalModifier), "a member's ID carries the leaf kind")

	condMods := e.ConditionalModifiers()
	c.Equal([]string{"Combat[to hit]"}, rowNames(condMods))
	if len(condMods) != 1 {
		return
	}
	c.NotEqual(reactions[0].TID, condMods[0].TID, "the same group name in the other table is a different container")

	other := NewEntity()
	addReaction(other, "B", "from foes", "Combat", fxp.One)
	other.Recalculate()
	otherReactions := other.Reactions()
	c.Equal([]string{"Combat[from foes]"}, rowNames(otherReactions))
	if len(otherReactions) != 1 {
		return
	}
	c.NotEqual(reactions[0].TID, otherReactions[0].TID,
		"the same group name on another entity is a different container, so its disclosure state is per sheet")
	c.Equal(reactions[0].Children[0].TID, otherReactions[0].Children[0].TID,
		"a member's ID does not depend on the entity, as the exporters rely on")
}

// TestConditionalModifierWhitespaceGroupIsNoGroup verifies that a group of nothing but whitespace is treated as no
// group, and that surrounding whitespace doesn't split a group in two.
func TestConditionalModifierWhitespaceGroupIsNoGroup(t *testing.T) {
	withGroupContainersOnSort(t, true)
	c := check.New(t)
	e := NewEntity()
	addReaction(e, "A", "from everyone", "   ", fxp.One)
	addReaction(e, "B", "from foes", " Combat", fxp.One)
	addReaction(e, "C", "from allies", "Combat ", fxp.One)
	e.Recalculate()
	c.Equal([]string{"Combat[from allies,from foes]", "from everyone"}, rowNames(e.Reactions()))
}

// TestConditionalModifierCloneKeepsContainerKindAndChildren verifies that cloning a group produces a group, with its
// children cloned beneath it, in both clone modes.
func TestConditionalModifierCloneKeepsContainerKindAndChildren(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	group := NewConditionalModifierGroup(e.ID, BlockReactionsKey, "Combat")
	child := NewConditionalModifier("from trait X", "from foes", fxp.Two)
	child.SetParent(group)
	group.Children = []*ConditionalModifier{child}

	for _, mode := range []CloneMode{Copy, Reference} {
		clone := group.Clone(LibraryFile{}, nil, nil, mode)
		c.True(clone.Container(), "mode %d: the clone is still a container", mode)
		c.Equal("Combat", clone.From, "mode %d", mode)
		c.Equal(mode == Copy, clone.TID == group.TID, "mode %d: only a copy keeps the ID", mode)
		c.Equal(1, len(clone.Children), "mode %d: the children are cloned", mode)
		if len(clone.Children) != 1 {
			continue
		}
		c.True(clone.Children[0] != child, "mode %d: the child is a clone, not the original", mode)
		c.True(clone.Children[0].Parent() == clone, "mode %d: the cloned child belongs to the clone", mode)
		c.False(clone.Children[0].Container(), "mode %d: the cloned child is still a modifier", mode)
		c.Equal(fxp.Two, clone.Children[0].Total(), "mode %d", mode)
		c.Equal(Hash64(group), Hash64(clone), "mode %d: a clone hashes identically", mode)
	}
	c.NotEqual(Hash64(group), Hash64(NewConditionalModifierGroup(e.ID, BlockReactionsKey, "Combat")),
		"the children participate in a group's hash")
}
