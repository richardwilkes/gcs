// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import (
	"context"
	"errors"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestMatcher checks that a Matcher gives the state Evaluate gives, asks a FieldResolver once per field and falls back
// to Field for a field the resolver can't give, and works with self tokens and hosts that resolve nothing.
func TestMatcher(t *testing.T) {
	c := check.New(t)
	base := newHost()
	h := &resolvingHost{fakeHost: base}
	tree := &Group{Children: Nodes{
		checkText("name", criteria.StartsWithText, "F"),
		&Group{Mode: GroupAny, Children: Nodes{
			checkText("unresolved", criteria.IsText, "{self.name}"),
			&Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 2)},
			&Script{Script: "failed"},
		}},
	}}
	rows := []*fakeItem{
		newItem("Fireball", map[string]Value{"level": numberValue(1), "unresolved": textValue("Fireball")}),
		newItem("Flame Jet", map[string]Value{"level": numberValue(3), "unresolved": textValue("x")}),
		newItem("Water", map[string]Value{"level": numberValue(3), "unresolved": textValue("x")}),
		newItem("Fog", map[string]Value{"unresolved": textValue("x"), "level": numberValue(0)}),
	}
	want := []State{Met, Met, Unmet, Failed}
	for _, host := range []Host{h, base} {
		m := NewMatcher(tree, host, Options{})
		for range 3 {
			for i, row := range rows {
				state, err := m.State(t.Context(), row)
				c.NoError(err)
				c.Equal(want[i], state, row.name)
				c.Equal(eval(t, tree, host, row, yesNo).State, state, row.name)
			}
		}
	}
	c.Equal(int64(3), h.asked.Load())
	state, err := NewMatcher(nil, h, Options{}).State(t.Context(), rows[0])
	c.NoError(err)
	c.Equal(Skipped, state, "an empty tree passes everything")

	// A check inside a where asks the resolver for its collection's field.
	kids := &Count{Of: "kids", Where: Ref{Node: &Check{Field: "level"}}, Quantity: atLeast(2)}
	m := NewMatcher(kids, h, Options{})
	parent := newItem("parent", nil).with("kids", rows...)
	state, err = m.State(t.Context(), parent)
	c.NoError(err)
	c.Equal(Met, state)
}

// TestMatcherCancel checks that a Matcher reports a done context instead of a state, and works again afterward.
func TestMatcherCancel(t *testing.T) {
	c := check.New(t)
	h := newHost()
	m := NewMatcher(&Script{Script: "block"}, h, Options{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	state, err := m.State(ctx, newItem("row", nil))
	c.True(errors.Is(err, context.Canceled))
	c.Equal(Failed, state)
	m = NewMatcher(&Const{Value: true}, h, Options{})
	state, err = m.State(t.Context(), newItem("row", nil))
	c.NoError(err)
	c.Equal(Met, state)
}

// TestMatcherAllocations checks that matching a row allocates nothing, given criteria that don't allocate on their own.
func TestMatcherAllocations(t *testing.T) {
	h := &resolvingHost{fakeHost: newHost()}
	m := NewMatcher(&Group{Children: Nodes{
		&Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 2)},
		&Group{Mode: GroupAny, Children: Nodes{
			checkText("name", criteria.IsText, "fire 3"),
			&Check{Field: "level", Number: numberCrit(criteria.EqualsNumber, 4)},
		}},
	}}, h, Options{})
	rows := matcherRows(100)
	ctx := t.Context()
	allocs := testing.AllocsPerRun(10, func() {
		for _, row := range rows {
			if _, err := m.State(ctx, row); err != nil {
				t.Fatal(err)
			}
		}
	})
	check.New(t).Equal(0.0, allocs)
}
