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
	"strconv"
	"sync"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

// matcherRows returns n rows for a check-only filter.
func matcherRows(n int) []*fakeItem {
	rows := make([]*fakeItem, n)
	for i := range rows {
		name := "Row " + strconv.Itoa(i)
		if i%3 == 0 {
			name = "Fire " + strconv.Itoa(i)
		}
		rows[i] = newItem(name, map[string]Value{
			"level": numberValue(int64(i % 5)),
			"tags":  listValue("Spell", "Tag "+strconv.Itoa(i%7)),
		})
	}
	return rows
}

// checkOnlyTree is a filter of checks only, as list filters are.
func checkOnlyTree() Node {
	return &Group{Children: Nodes{
		checkText("name", criteria.StartsWithText, "fire"),
		&Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 2)},
		&Group{Mode: GroupAny, Children: Nodes{
			checkText("tags", criteria.IsText, "Tag 3"),
			checkText("tags", criteria.ContainsText, "tag 1"),
		}},
	}}
}

// TestConcurrentEvaluation runs Evaluate and Matchers over one tree from several goroutines at once. It is listed in
// the TestRace wrapper.
func TestConcurrentEvaluation(t *testing.T) {
	c := check.New(t)
	h := &resolvingHost{fakeHost: newHost()}
	rows := matcherRows(300)
	tree := checkOnlyTree()
	big := &Group{Mode: GroupAny, Children: Nodes{tree, &Script{Script: "met"}}}
	want := make([]State, len(rows))
	for i, row := range rows {
		want[i] = eval(t, tree, h, row, yesNo).State
	}
	var wg sync.WaitGroup
	results := make([][]State, 8)
	for g := range results {
		wg.Go(func() {
			m := NewMatcher(tree, h, Options{})
			got := make([]State, len(rows))
			for i, row := range rows {
				got[i], _ = m.State(t.Context(), row) //nolint:errcheck // The context is never done.
				eval(t, big, h, row, collect)
			}
			results[g] = got
		})
	}
	wg.Wait()
	for _, got := range results {
		c.Equal(want, got)
	}
}

// BenchmarkMatcher matches 10,000 rows against check-only trees. The first tree uses comparisons whose criteria
// allocate on their own (case folding and comma splitting); the second uses only comparisons that don't.
func BenchmarkMatcher(b *testing.B) {
	h := &resolvingHost{fakeHost: newHost()}
	rows := matcherRows(10_000)
	plain := &Group{Children: Nodes{
		&Check{Field: "level", Number: numberCrit(criteria.AtLeastNumber, 2)},
		&Group{Mode: GroupAny, Children: Nodes{
			checkText("name", criteria.IsText, "fire 3"),
			&Check{Field: "level", Number: numberCrit(criteria.EqualsNumber, 4)},
		}},
	}}
	for _, tc := range []struct {
		name string
		tree Node
	}{
		{"criteria", checkOnlyTree()},
		{"plain", plain},
	} {
		b.Run(tc.name, func(b *testing.B) {
			m := NewMatcher(tc.tree, h, Options{})
			ctx := b.Context()
			matched := 0
			b.ReportAllocs()
			for b.Loop() {
				for _, row := range rows {
					if state, _ := m.State(ctx, row); state.Passes() { //nolint:errcheck // The context is never done.
						matched++
					}
				}
			}
			b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N*len(rows)), "ns/row")
			if matched == 0 {
				b.Fatal("nothing matched")
			}
		})
	}
}
