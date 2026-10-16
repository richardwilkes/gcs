// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// poolNet is the flow network of a pool, built once for all of its demands. Every member has an edge from the source
// and every demand's sink edge starts closed (capacity 0); a solve opens the sinks it needs. A closed demand is a dead
// end, so it takes no flow.
type poolNet struct {
	flowNet
	owners  []owner // what each node assigns a member to
	members int
}

// newPoolNet builds the network for members pool members and the demands.
func newPoolNet(members int, demands []*demand) *poolNet {
	n := &poolNet{flowNet: *newFlowNet(), members: members}
	n.owners = make([]owner, 2, 2+members+4*len(demands))
	node := func(o owner) int {
		n.owners = append(n.owners, o)
		return n.addNode()
	}
	for range members {
		n.addEdge(flowSource, node(owner{}), 1)
	}
	link := func(cands []int, to int) {
		for _, m := range cands {
			n.addEdge(2+m, to, 1)
		}
	}
	for _, d := range demands {
		switch {
		case !d.grouped:
			d.node = node(owner{d: d, value: -1})
			link(d.cands, d.node)
			d.sink = n.addEdge(d.node, flowSink, 0)
		case d.p == 1:
			d.node = node(owner{})
			for v, cands := range d.values {
				vn := node(owner{d: d, value: v})
				link(cands, vn)
				n.addEdge(vn, d.node, 1)
			}
			d.sink = n.addEdge(d.node, flowSink, 0)
		default:
			d.valueNodes = make([]int, len(d.values))
			d.valueSinks = make([]int, len(d.values))
			for v, cands := range d.values {
				d.valueNodes[v] = node(owner{d: d, value: v})
				link(cands, d.valueNodes[v])
				d.valueSinks[v] = n.addEdge(d.valueNodes[v], flowSink, 0)
			}
		}
	}
	return n
}

// open sets the capacity of a sink edge and returns it as a target.
func (n *poolNet) open(node, sink, capacity int) flowTarget {
	n.setCap(sink, capacity)
	return flowTarget{node: node, sinkEdge: sink}
}

// deficit returns the capacity still open on the targets' sink edges.
func (n *poolNet) deficit(targets []flowTarget) int {
	total := 0
	for _, t := range targets {
		total += n.cap[t.sinkEdge]
	}
	return total
}

// assigned returns where the current flow sends each member.
func (n *poolNet) assigned() []owner {
	owners := make([]owner, n.members)
	for m := range owners {
		for _, id := range n.adj[2+m] {
			if id&1 == 0 && n.flow(id) > 0 {
				owners[m] = n.owners[n.to[id]]
			}
		}
	}
	return owners
}
