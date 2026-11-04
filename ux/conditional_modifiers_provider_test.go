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
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

// condModListsForTest serves distinct conditional and reaction modifier lists so the tests can tell which accessor a
// provider is wired to.
type condModListsForTest struct {
	conditional []*gurps.ConditionalModifier
	reactions   []*gurps.ConditionalModifier
}

func (l *condModListsForTest) DataOwner() gurps.DataOwner { return nil }
func (l *condModListsForTest) ConditionalModifiers() []*gurps.ConditionalModifier {
	return l.conditional
}

func (l *condModListsForTest) Reactions() []*gurps.ConditionalModifier { return l.reactions }

func TestConditionalAndReactionProvidersUseTheirOwnSpec(t *testing.T) {
	lists := &condModListsForTest{
		conditional: []*gurps.ConditionalModifier{
			gurps.NewConditionalModifier("cond one", "", 0),
			gurps.NewConditionalModifier("cond two", "", 0),
		},
		reactions: []*gurps.ConditionalModifier{gurps.NewReaction("reaction", "", 0)},
	}
	for _, one := range []struct {
		name        string
		provider    TableProvider[*gurps.ConditionalModifier]
		refKey      string
		dragKey     *uti.DataType
		singular    string
		plural      string
		rows        []*gurps.ConditionalModifier
		descTooltip string
	}{
		{
			name:     "conditional modifiers",
			provider: NewConditionalModifiersProvider(lists),
			refKey:   gurps.BlockConditionalModifiersKey,
			dragKey:  conditionalModifierDragKey,
			singular: "Conditional Modifier",
			plural:   "Conditional Modifiers",
			rows:     lists.conditional,
		},
		{
			name:     "reaction modifiers",
			provider: NewReactionModifiersProvider(lists),
			refKey:   gurps.BlockReactionsKey,
			dragKey:  reactionModifierDragKey,
			singular: "Reaction Modifier",
			plural:   "Reaction Modifiers",
			rows:     lists.reactions,
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			p := one.provider
			p.SetTable(unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.ConditionalModifier]]{}))
			c.Equal(one.refKey, p.RefKey())
			c.True(one.dragKey == p.DragKey(), "drag key must be the %s key", one.name)
			singular, plural := p.ItemNames()
			c.Equal(one.singular, singular)
			c.Equal(one.plural, plural)
			c.Equal(len(one.rows), p.RootRowCount())
			c.Equal(one.rows, p.RootData())
			rows := p.RootRows()
			c.Equal(len(one.rows), len(rows))
			for i, row := range rows {
				c.True(one.rows[i] == row.Data(), "row %d must wrap the %s data", i, one.name)
			}
			headers := p.Headers()
			c.Equal(len(p.ColumnIDs()), len(headers))
			c.Equal(gurps.ConditionalModifierDescriptionColumn, p.HierarchyColumnID(),
				"the description column carries the disclosure triangles for the group containers")
		})
	}
	// The two providers must not share row data even though they are built from the same list provider.
	check.New(t).NotEqual(NewConditionalModifiersProvider(lists).RootRowCount(), NewReactionModifiersProvider(lists).RootRowCount())
}

// TestCondModProviderWrapsGroupChildren verifies that a group's members are wrapped beneath its node in order, while an
// ungrouped row is a leaf.
func TestCondModProviderWrapsGroupChildren(t *testing.T) {
	c := check.New(t)
	group := gurps.NewConditionalModifierGroup(gurps.NewEntity().ID, gurps.BlockReactionsKey, "Combat")
	first := gurps.NewReaction("from trait A", "from allies", fxp.One)
	second := gurps.NewReaction("from trait B", "from foes", fxp.Two)
	for _, child := range []*gurps.ConditionalModifier{first, second} {
		child.SetParent(group)
		group.Children = append(group.Children, child)
	}
	leaf := gurps.NewReaction("from trait C", "from everyone", fxp.Three)
	lists := &condModListsForTest{reactions: []*gurps.ConditionalModifier{group, leaf}}
	p := NewReactionModifiersProvider(lists)
	p.SetTable(unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.ConditionalModifier]]{}))
	rows := p.RootRows()
	c.Equal(2, len(rows))
	if len(rows) != 2 {
		return
	}
	c.True(rows[0].CanHaveChildren(), "the group row can have children")
	children := rows[0].Children()
	c.Equal(2, len(children), "the group's members are wrapped beneath it")
	if len(children) == 2 {
		c.True(first == children[0].Data(), "the members keep their order")
		c.True(second == children[1].Data(), "the members keep their order")
		c.True(rows[0] == children[0].Parent(), "a member's node knows its parent node")
	}
	c.False(rows[1].CanHaveChildren(), "an ungrouped row is a leaf")
	c.Equal(0, len(rows[1].Children()))
}

// TestSheetGroupedReactionsDiscloseAndPersist verifies that grouped reactions show as a container row whose closed
// state is recorded in the global settings and survives a rebuild, which regenerates the rows from scratch.
func TestSheetGroupedReactionsDiscloseAndPersist(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Warrior"
	for _, one := range []struct {
		situation string
		group     string
	}{
		{situation: "from foes", group: "Combat"},
		{situation: "from allies", group: "Combat"},
		{situation: "from everyone", group: ""},
	} {
		bonus := gurps.NewReactionBonus()
		bonus.Situation = one.situation
		bonus.Group = one.group
		bonus.Amount = fxp.One
		trait.Features = append(trait.Features, bonus)
	}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)

	c.True(listAttachedToSheet(sheet, sheet.Reactions), "the Reactions list must be on the page")
	table := sheet.Reactions.Table
	c.Equal(2, table.RootRowCount(), "a group container and an ungrouped entry")
	c.Equal(4, sheet.Reactions.RowCount(), "with the group open, its members are shown beneath it")
	var group *Node[*gurps.ConditionalModifier]
	for _, row := range table.RootRows() {
		if row.CanHaveChildren() {
			group = row
		}
	}
	c.NotNil(group, "the group row can have children")
	if group == nil {
		return
	}
	closedKey := "n:" + string(group.Data().ID())
	t.Cleanup(func() { gurps.SetClosedState(closedKey, false) })
	c.True(group.IsOpen(), "the group starts out open")
	c.False(gurps.IsClosed(closedKey))

	open, exists := sheet.Reactions.FirstDisclosureState()
	c.True(exists, "the list reports having something to disclose")
	c.True(open)
	sheet.Reactions.SetDisclosureState(false)
	c.True(gurps.IsClosed(closedKey), "closing the group records it in the global settings by ID")
	c.Equal(2, sheet.Reactions.RowCount(), "with the group closed, its members are hidden")

	sheet.Rebuild(true)
	c.True(listAttachedToSheet(sheet, sheet.Reactions), "the Reactions list is still on the page")
	c.Equal(2, sheet.Reactions.Table.RootRowCount(), "the rows are rebuilt")
	c.Equal(2, sheet.Reactions.RowCount(), "and the group stays closed, since its ID is derived from its name")
	open, exists = sheet.Reactions.FirstDisclosureState()
	c.True(exists)
	c.False(open)

	sheet.Reactions.SetDisclosureState(true)
	c.False(gurps.IsClosed(closedKey), "opening the group clears the recorded state")
	c.Equal(4, sheet.Reactions.RowCount())
}
