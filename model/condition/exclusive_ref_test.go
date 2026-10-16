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
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

var refColleges = []string{"A", "B", "C", "D"}

// randomMembers returns up to 8 members, each with some of three tags and up to two colleges.
func randomMembers(r *rand.Rand) []*fakeItem {
	members := make([]*fakeItem, 1+r.IntN(8))
	for i := range members {
		var tags, colleges []string
		for d := range 3 {
			if r.IntN(2) == 0 {
				tags = append(tags, fmt.Sprintf("d%d", d))
			}
		}
		for _, one := range refColleges {
			if r.IntN(3) == 0 && len(colleges) < 2 {
				colleges = append(colleges, one)
			}
		}
		members[i] = newItem(fmt.Sprintf("m%d", i), map[string]Value{
			"tags": listValue(tags...), "college": listValue(colleges...),
		})
	}
	return members
}

// candidates returns the indexes of the members that have tag.
func candidates(members []*fakeItem, tag string) []int {
	var list []int
	for i, m := range members {
		if slices.Contains(m.fields["tags"].List, tag) {
			list = append(list, i)
		}
	}
	return list
}

// TestPoolAgainstBruteForce compares the pool solver with a search over every assignment of members to counts, over
// thousands of small random pools.
func TestPoolAgainstBruteForce(t *testing.T) {
	c := check.New(t)
	r := rand.New(rand.NewPCG(5, 6)) //nolint:gosec // Not for security.
	met := 0
	for i := range 4000 {
		members := randomMembers(r)
		tree := &Group{Exclusive: true}
		demands := 1 + r.IntN(4)
		cands := make([][]int, 0, demands)
		needs := make([]int, 0, demands)
		for range demands {
			tag := fmt.Sprintf("d%d", r.IntN(3))
			q := 1 + r.IntN(3)
			tree.Children = append(tree.Children, &Count{
				Of: "kids", Where: Ref{Node: checkText("tags", criteria.IsText, tag)},
				Quantity: atLeast(int64(q)),
			})
			cands = append(cands, candidates(members, tag))
			needs = append(needs, q)
		}
		got := make([]int, len(needs))
		var search func(m int) bool
		search = func(m int) bool {
			if m == len(members) {
				for d, n := range got {
					if n < needs[d] {
						return false
					}
				}
				return true
			}
			if search(m + 1) {
				return true
			}
			for d := range needs {
				if slices.Contains(cands[d], m) {
					got[d]++
					found := search(m + 1)
					got[d]--
					if found {
						return true
					}
				}
			}
			return false
		}
		want := metState(search(0), false)
		self := newItem("self", nil).with("kids", members...)
		c.Equal(want, evalAll(t, tree, newHost(), self).State, fmt.Sprintf("case %d", i))
		if want == Met {
			met++
		}
	}
	c.True(met > 600 && met < 3400, fmt.Sprint("met ", met))
}

// TestGroupedAgainstBruteForce compares the most values an exclusive grouped count fills with a search over every
// assignment of members to values, over thousands of small random sets of members.
func TestGroupedAgainstBruteForce(t *testing.T) {
	c := check.New(t)
	r := rand.New(rand.NewPCG(7, 8)) //nolint:gosec // Not for security.
	for i := range 4000 {
		members := randomMembers(r)
		p := 1 + r.IntN(3)
		count := &Count{Of: "kids", GroupBy: "college", PerGroup: atLeast(int64(p)), Exclusive: true}
		perValue := make(map[string]int)
		best := 0
		var search func(m int)
		search = func(m int) {
			if m == len(members) {
				filled := 0
				for _, n := range perValue {
					if n >= p {
						filled++
					}
				}
				best = max(best, filled)
				return
			}
			search(m + 1)
			for _, v := range members[m].fields["college"].List {
				perValue[v]++
				search(m + 1)
				perValue[v]--
			}
		}
		search(0)
		self := newItem("self", nil).with("kids", members...)
		got := evalAll(t, count, newHost(), self)
		c.Equal(int64(best), got.Have.AsInteger[int64](), fmt.Sprintf("case %d", i))
	}
}
