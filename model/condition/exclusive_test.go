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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/check"
)

func spell(name string, colleges ...string) *fakeItem {
	return newItem(name, map[string]Value{"college": listValue(colleges...), "tags": listValue()})
}

func spellBook(spells ...*fakeItem) (h *fakeHost, self *fakeItem) {
	h = newHost()
	h.entity["spells"] = spells
	return h, newItem("Owner", nil)
}

func spellsWhere(where Node, q *criteria.Number) *Count {
	return &Count{Of: "spells", Where: Ref{Node: where}, Quantity: q}
}

func inCollege(c string) Node {
	return checkText("college", criteria.IsText, c)
}

func exclusive(children ...Node) *Group {
	return &Group{Exclusive: true, Children: children}
}

func numbered(prefix string, n int, colleges ...string) []*fakeItem {
	list := make([]*fakeItem, n)
	for i := range list {
		list[i] = spell(fmt.Sprintf("%s %d", prefix, i), colleges...)
	}
	return list
}

func TestDispelMagic(t *testing.T) {
	c := check.New(t)
	tree := exclusive(spellsWhere(checkText("name", criteria.IsText, "Counterspell"), nil), spellsWhere(nil, atLeast(12)))
	h, self := spellBook(append([]*fakeItem{spell("Counterspell")}, numbered("Other", 12)...)...)
	r := evalAll(t, tree, h, self)
	c.Equal(Met, r.State)
	c.Equal(2, r.Specificity)
	c.Equal(fxp.FromInteger(12), r.Children[1].Have)
	h, self = spellBook(append([]*fakeItem{spell("Counterspell")}, numbered("Other", 11)...)...)
	r = evalAll(t, tree, h, self)
	c.Equal(Unmet, r.State)
	c.Equal(Met, r.Children[0].State)
	c.Equal(Unmet, r.Children[1].State)
	c.Equal(fxp.FromInteger(11), r.Children[1].Have)
	c.Equal(fxp.FromInteger(12), r.Children[1].Need)
	c.Equal(Met, eval(t, &Group{Children: tree.Children}, h, self, yesNo).State)
}

// TestWeatherDome needs a spell in two colleges to fill only one of them.
func TestWeatherDome(t *testing.T) {
	c := check.New(t)
	tree := exclusive(
		spellsWhere(inCollege("Air"), atLeast(2)),
		spellsWhere(inCollege("Earth"), atLeast(2)),
		spellsWhere(inCollege("Fire"), atLeast(2)),
		spellsWhere(inCollege("Water"), atLeast(2)),
	)
	spells := []*fakeItem{
		spell("Air 1", "Air"), spell("Air 2", "Air"), spell("Earth 1", "Earth"), spell("Mud Jet", "Earth", "Water"),
		spell("Water 1", "Water"), spell("Fire 1", "Fire"), spell("Fire 2", "Fire"),
	}
	h, self := spellBook(spells...)
	c.Equal(Unmet, evalAll(t, tree, h, self).State)
	c.Equal(Met, eval(t, &Group{Children: tree.Children}, h, self, yesNo).State)
	h, self = spellBook(append(spells, spell("Water 2", "Water"))...)
	c.Equal(Met, evalAll(t, tree, h, self).State)
}

func TestGreatShapeshift(t *testing.T) {
	c := check.New(t)
	shapeshifting := checkText("tags", criteria.IsText, "Shapeshifting")
	tree := exclusive(
		spellsWhere(checkText("name", criteria.IsText, "Alter Body"), nil),
		spellsWhere(shapeshifting, atLeast(4)),
		spellsWhere(nil, atLeast(10)),
	)
	shifts := numbered("Shift", 4)
	for _, one := range shifts {
		one.fields["tags"] = listValue("Shapeshifting")
	}
	base := append([]*fakeItem{spell("Alter Body")}, shifts...)
	h, self := spellBook(append(base, numbered("Other", 10)...)...)
	c.Equal(Met, evalAll(t, tree, h, self).State)
	h, self = spellBook(append(base, numbered("Other", 9)...)...)
	r := evalAll(t, tree, h, self)
	c.Equal(Unmet, r.State)
	c.Equal(fxp.FromInteger(9), r.Children[2].Have)
}

// TestGreedyCounterexample needs the flow solver: a greedy pass gives Mud Jet to Earth and leaves Water short.
func TestGreedyCounterexample(t *testing.T) {
	c := check.New(t)
	mud, earth := spell("Mud Jet", "Earth", "Water"), spell("Shape Earth", "Earth")
	h, self := spellBook(mud, earth)
	tree := exclusive(spellsWhere(inCollege("Earth"), nil), spellsWhere(inCollege("Water"), nil))
	r := evalAll(t, tree, h, self)
	c.Equal(Met, r.State)
	c.Equal([]Item{earth}, hits(&r, tree.Children[0]))
	c.Equal([]Item{mud}, hits(&r, tree.Children[1]))
}

// TestPoolSurplus checks that a count of a pool reports the members it could have taken that no other count holds,
// and that each member is counted once.
func TestPoolSurplus(t *testing.T) {
	c := check.New(t)
	fire := numbered("Fire", 3, "Fire")
	h, self := spellBook(fire...)
	tree := exclusive(spellsWhere(inCollege("Fire"), nil), spellsWhere(nil, nil))
	r := evalAll(t, tree, h, self)
	c.Equal(Met, r.State)
	c.Equal(fxp.FromInteger(2), r.Children[0].Have)
	c.Equal(fxp.FromInteger(1), r.Children[1].Have)
	c.Equal(2, len(r.Children[0].Children))
	c.Equal(0, len(r.Children[1].Children))
}

// TestPoolFailures checks the states of a pool with a count that can't be evaluated: Unmet when the others alone can't
// all be met, which the Failed one couldn't change, and otherwise Failed. Not swaps Met and Unmet only.
func TestPoolFailures(t *testing.T) {
	c := check.New(t)
	h, self := spellBook(spell("Fire 1", "Fire"))
	broken := &Count{Of: "spells", Where: Ref{Node: &Check{Field: "nope"}}}
	for _, tc := range []struct {
		name  string
		tree  *Group
		state State
	}{
		{"failed and met", exclusive(spellsWhere(inCollege("Fire"), nil), broken), Failed},
		{"failed and unmet", exclusive(broken, spellsWhere(inCollege("Fire"), atLeast(2))), Unmet},
		{"failed where", exclusive(spellsWhere(&Script{Script: "failed"}, nil), spellsWhere(inCollege("Fire"), nil)), Failed},
		{"not met", &Group{Exclusive: true, Not: true, Children: Nodes{spellsWhere(nil, nil)}}, Unmet},
		{"not unmet", &Group{Exclusive: true, Not: true, Children: Nodes{spellsWhere(nil, atLeast(2))}}, Met},
		{"not failed", &Group{Exclusive: true, Not: true, Children: Nodes{broken}}, Failed},
		{"guarded", &Group{Exclusive: true, OnlyWhen: Ref{Node: &Const{}}, Children: Nodes{broken}}, Skipped},
		{"empty", &Group{Mode: GroupAny, Children: Nodes{&Group{Exclusive: true}}}, Skipped},
	} {
		r := evalAll(t, tc.tree, h, self)
		c.Equal(tc.state, r.State, tc.name)
		if tc.state != Skipped && len(tc.tree.Children) == 2 {
			c.Equal(2, len(r.Children), tc.name)
		}
	}
	r := evalAll(t, exclusive(spellsWhere(inCollege("Fire"), nil), broken), h, self)
	c.Equal(`no field "nope"`, r.Reason)
	c.Equal(Met, r.Children[0].State)
	c.Equal(Failed, r.Children[1].State)
	c.Equal(Node(broken), r.Children[1].Node)
	c.Equal(0, r.Specificity)
}

// TestAccelerateTime checks an exclusive grouped count against the plain one: a spell in two colleges fills only one
// of them.
func TestAccelerateTime(t *testing.T) {
	c := check.New(t)
	count := &Count{
		Of:        "spells",
		GroupBy:   "college",
		PerGroup:  atLeast(2),
		Quantity:  atLeast(3),
		Exclusive: true,
	}
	spells := []*fakeItem{
		spell("A1", "A"), spell("AB", "A", "B"), spell("B1", "B"), spell("C1", "C"), spell("C2", "C"),
	}
	h, self := spellBook(spells...)
	r := evalAll(t, count, h, self)
	c.Equal(Unmet, r.State)
	c.Equal(fxp.FromInteger(2), r.Have)
	plain := *count
	plain.Exclusive = false
	r = evalAll(t, &plain, h, self)
	c.Equal(Met, r.State)
	c.Equal(fxp.FromInteger(3), r.Have)
	h, self = spellBook(append(spells, spell("B2", "B"), spell("C3", "C"))...)
	r = evalAll(t, count, h, self)
	c.Equal(Met, r.State)
	c.Equal(fxp.FromInteger(3), r.Have)
	c.Equal(7, len(hits(&r, count)), "a surplus member joins a filled value")
}

// TestPerGroupRerouting needs the branch and bound to reroute: with P = 2, value A {m1, m2, m3} and value B {m2, x} are
// both filled only when m2 goes to B.
func TestPerGroupRerouting(t *testing.T) {
	c := check.New(t)
	h, self := spellBook(spell("m1", "A"), spell("m2", "A", "B"), spell("m3", "A"), spell("x", "B"))
	grouped := &Count{
		Of: "spells", GroupBy: "college", PerGroup: atLeast(2), Quantity: atLeast(2),
		Exclusive: true,
	}
	r := evalAll(t, grouped, h, self)
	c.Equal(Met, r.State)
	c.Equal(fxp.FromInteger(2), r.Have)
	grouped.Quantity = atLeast(3)
	c.Equal(Unmet, evalAll(t, grouped, h, self).State)
	one := &Count{Of: "spells", GroupBy: "college", Quantity: atLeast(2), Exclusive: true}
	r = evalAll(t, one, h, self)
	c.Equal(Met, r.State)
	c.Equal(4, len(hits(&r, one)))
}

// TestPoolBudget checks that a pool or exclusive count that runs out of budget is Failed, as too complex, wherever it
// runs out, and is otherwise decided.
func TestPoolBudget(t *testing.T) {
	c := check.New(t)
	h, self := spellBook(numbered("Fire", 40, "Fire", "Air")...)
	for _, tree := range []Node{
		exclusive(spellsWhere(inCollege("Fire"), atLeast(20)), spellsWhere(inCollege("Air"), atLeast(20))),
		&Count{Of: "spells", GroupBy: "college", PerGroup: atLeast(20), Exclusive: true},
		&Count{Of: "spells", GroupBy: "college", Exclusive: true},
	} {
		r := eval(t, tree, h, self, Options{Budget: 100})
		c.Equal(Failed, r.State)
		c.Equal("too complex to decide", r.Reason)
		c.Equal(Met, evalAll(t, tree, h, self).State)
		for budget := 100; budget < 3000; budget += 7 {
			r = eval(t, tree, h, self, Options{Budget: budget})
			c.True(r.State == Met || r.Reason == "too complex to decide", budget)
		}
	}
}
