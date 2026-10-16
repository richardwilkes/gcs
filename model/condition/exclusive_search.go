// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// maximize fills as many values of a grouped demand as it can, each member counting toward one value only: a max flow
// when each value needs one member, otherwise a branch and bound over the values, including a value before excluding
// it, with the incumbent as the bound. Including a value opens its sink and augments from the current flow, which
// never takes flow off a sink already filled, so a value that can't be filled shows the include is infeasible. Each
// node prunes with the relaxation bound: the values still undecided can add at most min(their number, F / p), where F
// is the flow they could still take. It returns false when the budget ran out.
func (s *poolSolver) maximize(d *demand) ([]owner, bool) {
	net := s.net
	if d.p == 1 {
		target := net.open(d.node, d.sink, len(d.values))
		if !net.maxFlow(s.e, []flowTarget{target}) {
			return nil, false
		}
		return net.assigned(), true
	}
	var best []owner
	bestCount := -1
	ok := true
	reachable, _ := probes(d, 0)
	var search func(from, count int)
	search = func(from, count int) {
		if !ok || bestCount == len(reachable) {
			return
		}
		if count > bestCount {
			best, bestCount = net.assigned(), count
		}
		targets, v := probes(d, from)
		var flow int
		if flow, ok = net.probe(s.e, targets, d.p, d.p*(bestCount-count+1)); !ok {
			return
		}
		if len(targets) == 0 || count+min(len(targets), flow/d.p) <= bestCount {
			return
		}
		net.save()
		target := net.open(d.valueNodes[v], d.valueSinks[v], d.p)
		if ok = net.maxFlow(s.e, []flowTarget{target}); !ok {
			return
		}
		if net.cap[target.sinkEdge] == 0 {
			search(v+1, count+1)
		}
		net.restore()
		search(v+1, count)
	}
	search(0, 0)
	return best, ok
}

// probes returns the targets of the values of d from index from on that have at least p candidates in the whole pool,
// for measuring how much more flow they could take, and the index of the first.
func probes(d *demand, from int) (targets []flowTarget, first int) {
	first = -1
	for v := from; v < len(d.values); v++ {
		if len(d.values[v]) >= d.p {
			if first < 0 {
				first = v
			}
			targets = append(targets, flowTarget{node: d.valueNodes[v], sinkEdge: d.valueSinks[v]})
		}
	}
	return targets, first
}
