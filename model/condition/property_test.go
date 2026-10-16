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
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/toolbox/v2/check"
)

var (
	randomFields  = []string{"name", "level", "tags", "flag", "tl", "missing"}
	randomTexts   = []string{"a", "b", "", "A, b"}
	randomScripts = []string{"met", "unmet", "skipped", "failed"}
)

// randomHost returns a host whose self has a "kids" collection of random items, each with its own random kids, and
// whose entity has "spells", "skills" and "traits" collections of random items with colleges.
func randomHost(r *rand.Rand) (h *fakeHost, self *fakeItem) {
	var build func(level int) *fakeItem
	build = func(level int) *fakeItem {
		fields := map[string]Value{
			"level": numberValue(int64(r.IntN(4))),
			"tags":  listValue(randomTexts[:r.IntN(len(randomTexts))]...),
			"flag":  boolValue(r.IntN(2) == 0),
		}
		if r.IntN(2) == 0 {
			fields["tl"] = absentValue()
		}
		if r.IntN(4) == 0 {
			fields["level"] = textValue("odd")
		}
		it := newItem(randomTexts[r.IntN(len(randomTexts))], fields)
		if level < 2 {
			kids := make([]*fakeItem, r.IntN(4))
			for i := range kids {
				kids[i] = build(level + 1)
			}
			it.with("kids", kids...)
		}
		return it
	}
	h = newHost()
	h.attributes["st"] = numberValue(int64(r.IntN(3))).Number
	for _, key := range []string{"spells", "skills", "traits"} {
		list := make([]*fakeItem, r.IntN(8))
		for i := range list {
			list[i] = build(2)
			var colleges []string
			for _, one := range []string{"Fire", "Air", "Water"} {
				if r.IntN(3) == 0 {
					colleges = append(colleges, one)
				}
			}
			list[i].fields["college"] = listValue(colleges...)
		}
		h.entity[key] = list
	}
	return h, build(0)
}

// randomNode returns a random valid tree, exclusive groups and exclusive grouped counts included.
func randomNode(r *rand.Rand, depth int) Node {
	kind := r.IntN(8)
	if depth > 3 {
		kind = 2 + r.IntN(3)
	}
	switch kind {
	case 0:
		g := &Group{Mode: GroupModes[r.IntN(len(GroupModes))], Not: r.IntN(3) == 0}
		for range r.IntN(4) {
			g.Children = append(g.Children, randomNode(r, depth+1))
		}
		if g.Mode == GroupAtLeast {
			if len(g.Children) == 0 {
				g.Mode = GroupAll
			} else {
				g.N = 1 + r.IntN(len(g.Children))
			}
		}
		if r.IntN(4) == 0 {
			g.OnlyWhen = Ref{Node: randomNode(r, depth+1)}
		}
		return g
	case 1:
		g := &Group{Exclusive: true, Not: r.IntN(3) == 0}
		for range 1 + r.IntN(3) {
			count := &Count{Of: "kids", Quantity: atLeast(int64(1 + r.IntN(3)))}
			if r.IntN(3) != 0 {
				count.Where = Ref{Node: randomNode(r, depth+2)}
			}
			g.Children = append(g.Children, count)
		}
		return g
	case 2:
		field := randomFields[r.IntN(len(randomFields))]
		c := &Check{Field: field, Not: r.IntN(3) == 0}
		switch r.IntN(4) {
		case 0:
			c.Text = textCrit(criteria.StringComparisons[1+r.IntN(len(criteria.StringComparisons)-1)],
				randomTexts[r.IntN(len(randomTexts))])
		case 1:
			c.Number = numberCrit(criteria.NumericComparisons[1+r.IntN(len(criteria.NumericComparisons)-1)],
				int64(r.IntN(4)))
		case 2:
			c.Text = textCrit(criteria.IsText, "{self.name}")
		default:
		}
		return c
	case 3:
		return &Const{Value: r.IntN(2) == 0}
	case 4:
		return &Script{Script: randomScripts[r.IntN(len(randomScripts))]}
	case 5:
		count := &Count{Of: "kids", IncludeSelf: r.IntN(4) == 0, Not: r.IntN(3) == 0}
		if r.IntN(3) != 0 {
			count.Where = Ref{Node: randomNode(r, depth+1)}
		}
		if r.IntN(2) == 0 {
			count.Quantity = numberCritPtr(criteria.NumericComparisons[r.IntN(len(criteria.NumericComparisons))],
				int64(r.IntN(3)))
		}
		if r.IntN(3) == 0 {
			count.GroupBy = "tags"
			count.Exclusive = r.IntN(2) == 0
			if r.IntN(2) == 0 {
				count.PerGroup = atLeast(int64(1 + r.IntN(2)))
			}
		}
		return count
	default:
		left := Operand{Aggregate: Aggregates[1+r.IntN(len(Aggregates)-1)], Of: "kids", Field: "level"}
		if left.Aggregate == AggregateCount {
			left.Field = ""
		}
		if r.IntN(2) == 0 {
			left.Where = Ref{Node: randomNode(r, depth+2)}
		}
		return &Compare{Left: left, Op: CompareOps[r.IntN(len(CompareOps))], Right: Operand{Sum: []Operand{
			{Attribute: "st"}, value(int64(r.IntN(3))),
		}}}
	}
}

// TestModesAgree checks, over random valid trees and hosts, that every mode reaches the same state, and that Collect
// does so without running out of budget.
func TestModesAgree(t *testing.T) {
	c := check.New(t)
	r := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // Not for security.
	for i := range 3000 {
		h, self := randomHost(r)
		tree := randomNode(r, 0)
		c.Equal(0, len(Validate(tree)), i)
		full := evalAll(t, tree, h, self)
		c.False(full.Truncated, i)
	}
}

// TestAnySpecificityIgnoresOrder checks that permuting the children of an any group changes neither its state nor its
// specificity.
func TestAnySpecificityIgnoresOrder(t *testing.T) {
	c := check.New(t)
	r := rand.New(rand.NewPCG(3, 4)) //nolint:gosec // Not for security.
	for i := range 500 {
		h, self := randomHost(r)
		g := &Group{Mode: GroupAny}
		for range 1 + r.IntN(4) {
			g.Children = append(g.Children, randomNode(r, 1))
		}
		want := eval(t, g, h, self, collect)
		for range 4 {
			shuffled := &Group{Mode: GroupAny, Children: slices.Clone(g.Children)}
			r.Shuffle(len(shuffled.Children), func(a, b int) {
				shuffled.Children[a], shuffled.Children[b] = shuffled.Children[b], shuffled.Children[a]
			})
			got := eval(t, shuffled, h, self, collect)
			c.Equal(want.State, got.State, i)
			c.Equal(want.Specificity, got.Specificity, i)
		}
	}
}
