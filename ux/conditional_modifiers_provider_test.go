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
		})
	}
	// The two providers must not share row data even though they are built from the same list provider.
	check.New(t).NotEqual(NewConditionalModifiersProvider(lists).RootRowCount(), NewReactionModifiersProvider(lists).RootRowCount())
}
