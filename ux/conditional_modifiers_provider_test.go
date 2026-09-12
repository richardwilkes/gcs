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
		reactions: []*gurps.ConditionalModifier{gurps.NewConditionalModifier("reaction", "", 0)},
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

// TestCondModProviderWrapsGroupChildren verifies that a group container row is wrapped as a node that can have
// children, with its members wrapped beneath it in order, while an ungrouped row is a leaf.
func TestCondModProviderWrapsGroupChildren(t *testing.T) {
	c := check.New(t)
	group := gurps.NewConditionalModifierGroup(gurps.NewEntity().ID, gurps.BlockReactionsKey, "Combat")
	first := gurps.NewConditionalModifier("from trait A", "from allies", fxp.One)
	second := gurps.NewConditionalModifier("from trait B", "from foes", fxp.Two)
	for _, child := range []*gurps.ConditionalModifier{first, second} {
		child.SetParent(group)
		group.Children = append(group.Children, child)
	}
	leaf := gurps.NewConditionalModifier("from trait C", "from everyone", fxp.Three)
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
