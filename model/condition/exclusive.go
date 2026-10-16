// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import "github.com/richardwilkes/gcs/v5/model/fxp"

// poolSolver solves one pool: the counts of an exclusive group, or the values of an exclusive grouped count, which
// share the members so that each counts toward one of them at most. Members are keyed by identity.
type poolSolver struct {
	e     *evaluator
	index map[Item]int
	items []Item
	net   *poolNet
}

func newPoolSolver(e *evaluator) *poolSolver {
	return &poolSolver{e: e, index: make(map[Item]int)}
}

// member returns the pool index of item, adding it if new.
func (s *poolSolver) member(item Item) int {
	m, ok := s.index[item]
	if !ok {
		m = len(s.items)
		s.index[item] = m
		s.items = append(s.items, item)
	}
	return m
}

// exclusiveGroup evaluates an exclusive group, once its guard has let it apply. Its counts form a pool, and the group
// is Met when one assignment of the members meets every count, Unmet when the counts that could be evaluated can't all
// be met, which no Failed count could change, and otherwise Failed when any count is. Not swaps Met and Unmet. In
// Collect mode, each count reports the members the assignment gave it, along with any it could have taken that no
// other count holds.
func (e *evaluator) exclusiveGroup(g *Group, f frame) Result {
	if len(g.Children) == 0 {
		return Result{State: Skipped}
	}
	s := newPoolSolver(e)
	children := make([]Result, len(g.Children))
	demands := make([]*demand, 0, len(g.Children))
	var failed *Result
	for i, child := range g.Children {
		c, ok := child.(*Count)
		if !ok { // Validate allows nothing else
			continue
		}
		d, fail := s.demand(c, f)
		if fail == nil {
			demands = append(demands, d)
			continue
		}
		children[i] = *fail
		children[i].Node = c
		if failed == nil {
			failed = fail
		}
	}
	s.net = newPoolNet(len(s.items), demands)
	owners, deficit, ok := s.solve(demands)
	var r Result
	switch {
	case !ok:
		return failure(reasonTooComplex())
	case deficit > 0:
		r.State = metState(false, g.Not)
	case failed != nil:
		r.State, r.Reason = Failed, failed.Reason
	default:
		r.State = metState(true, g.Not)
	}
	if e.collect {
		met := make([]bool, len(demands))
		for i, d := range demands {
			members, _, _ := d.tally(owners)
			met[i] = members >= d.q
		}
		surplus(demands, owners)
		j := 0
		for i := range children {
			if children[i].Node == nil {
				children[i] = s.demandResult(demands[j], owners, met[j])
				j++
			}
		}
		r.Children = children
		if r.State == Met && !g.Not {
			r.Specificity = len(children)
		}
	}
	return r
}

// demand finds the candidates of a count of the pool. A count whose collection the host lacks, or whose where ends
// Failed, returns the Failed result instead.
func (s *poolSolver) demand(c *Count, f frame) (*demand, *Result) {
	var cands []int
	var wheres []Result
	if fail := s.e.eachMember(&c.Members, f, func(item Item, where *Result) bool {
		cands = append(cands, s.member(item))
		if where != nil {
			wheres = append(wheres, *where)
		}
		return true
	}); fail != nil {
		return nil, fail
	}
	return newDemand(c, cands, wheres), nil
}

// exclusiveGroupedCount groups the matched members of an exclusive grouped count, each member counting toward one
// value only, choosing the assignment that fills the most values. It returns the number of values filled and which
// members they hold, which in Collect mode includes those that could join a filled value that no other value holds.
func (e *evaluator) exclusiveGroupedCount(c *Count, members []Item) (have int, used []bool, fail *Result) {
	s := newPoolSolver(e)
	cands := make([]int, len(members))
	for i, item := range members {
		cands[i] = s.member(item)
	}
	d := newDemand(c, cands, nil)
	if fail = d.group(e, s.items); fail != nil {
		return 0, nil, fail
	}
	s.net = newPoolNet(len(s.items), []*demand{d})
	owners, ok := s.maximize(d)
	if !ok {
		r := failure(reasonTooComplex())
		return 0, nil, &r
	}
	_, _, have = d.tally(owners)
	surplus([]*demand{d}, owners)
	used = make([]bool, len(members))
	for i, m := range cands {
		used[i] = owners[m].d == d
	}
	return have, used, nil
}

// solve assigns pool members to the demands, returning the assignment and how many members the demands are still
// short of, or false when the budget ran out.
func (s *poolSolver) solve(demands []*demand) (owners []owner, deficit int, ok bool) {
	targets := make([]flowTarget, len(demands))
	for i, d := range demands {
		targets[i] = s.net.open(d.node, d.sink, d.q)
	}
	if !s.net.maxFlow(s.e, targets) {
		return nil, 0, false
	}
	return s.net.assigned(), s.net.deficit(targets), true
}

// demandResult reports a count of a pool: the members assigned to it as Have, each a hit, with the result of its
// where as a child.
func (s *poolSolver) demandResult(d *demand, owners []owner, met bool) Result {
	members, _, _ := d.tally(owners)
	r := Result{
		Node: d.count, State: metState(met, false), Have: fxp.FromInteger(members), HaveKind: ValueNumber,
		Need: fxp.One,
	}
	if d.count.Quantity != nil {
		r.Need = d.count.Quantity.Qualifier
	}
	for i, m := range d.cands {
		if owners[m].d != d {
			continue
		}
		r.hits = append(r.hits, Hit{Node: d.count, Item: s.items[m]})
		if !d.count.Where.IsZero() {
			r.Children = append(r.Children, d.wheres[i])
		}
	}
	if met {
		r.Specificity = 1
	}
	return r
}
