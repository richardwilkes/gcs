// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// flowNet is a flow network. Edges are stored in pairs, so the reverse of edge i is i^1, and an edge's flow is the
// residual capacity of its reverse. Adjacency lists keep insertion order, which makes every search deterministic.
// While a snapshot is held, every capacity change is logged so that restoring undoes it.
type flowNet struct {
	adj    [][]int
	to     []int
	cap    []int
	parent []int
	seen   []uint32 // the search that last reached each node
	goal   []int    // during a probe, each target node's sink edge plus 1
	queue  []int
	log    []flowChange
	marks  []int
	search uint32
}

// flowChange is a logged change: delta added to the capacity of an edge, and when paired taken from its reverse.
type flowChange struct {
	edge   int
	delta  int
	paired bool
}

const (
	flowSource = 0
	flowSink   = 1
)

// flowTarget is a node whose edge to the sink is a demand to saturate.
type flowTarget struct {
	node     int
	sinkEdge int
}

func newFlowNet() *flowNet {
	n := &flowNet{}
	n.addNode()
	n.addNode()
	return n
}

func (n *flowNet) addNode() int {
	n.adj = append(n.adj, nil)
	return len(n.adj) - 1
}

// addEdge adds an edge from u to v and returns its id.
func (n *flowNet) addEdge(u, v, capacity int) int {
	id := len(n.to)
	n.to = append(n.to, v, u)
	n.cap = append(n.cap, capacity, 0)
	n.adj[u] = append(n.adj[u], id)
	n.adj[v] = append(n.adj[v], id+1)
	return id
}

func (n *flowNet) flow(edge int) int {
	return n.cap[edge^1]
}

// setCap sets an edge's capacity, logging the change while a snapshot is held.
func (n *flowNet) setCap(edge, capacity int) {
	if len(n.marks) != 0 {
		n.log = append(n.log, flowChange{edge: edge, delta: capacity - n.cap[edge]})
	}
	n.cap[edge] = capacity
}

// push moves one unit of flow along an edge, logging it while a snapshot is held.
func (n *flowNet) push(edge int) {
	n.cap[edge]--
	n.cap[edge^1]++
	if len(n.marks) != 0 {
		n.log = append(n.log, flowChange{edge: edge, delta: -1, paired: true})
	}
}

// save takes a snapshot; restore undoes every change since the matching save.
func (n *flowNet) save() {
	n.marks = append(n.marks, len(n.log))
}

func (n *flowNet) restore() {
	k := len(n.marks) - 1
	mark := n.marks[k]
	for i := len(n.log) - 1; i >= mark; i-- {
		c := n.log[i]
		n.cap[c.edge] -= c.delta
		if c.paired {
			n.cap[c.edge^1] += c.delta
		}
	}
	n.log = n.log[:mark]
	n.marks = n.marks[:k]
}

// maxFlow saturates the targets' sink edges as far as possible: one target at a time in order, then again in rounds
// until a round changes nothing. Augmenting paths never pass through the sink, so flow already on a sink edge is never
// reduced. Each search costs the nodes it reached. It returns false when the budget ran out.
func (n *flowNet) maxFlow(e *evaluator, targets []flowTarget) bool {
	for {
		progress := false
		for _, t := range targets {
			for n.cap[t.sinkEdge] > 0 {
				found := n.findPath(t.node, false)
				if !e.spend(len(n.queue)) {
					return false
				}
				if found < 0 {
					break
				}
				n.pushPath(t.node, t.sinkEdge)
				progress = true
			}
		}
		if !progress {
			return true
		}
	}
}

// probe opens the targets' sink edges with the given capacity and pushes flow into any of them until limit units have
// gone in or no path remains, then undoes everything. It returns the flow that went in, which is the most the targets
// can take whenever it is below limit. The order of the targets doesn't matter, so each search ends at the first one
// it reaches.
func (n *flowNet) probe(e *evaluator, targets []flowTarget, capacity, limit int) (flow int, ok bool) {
	if limit <= 0 || len(targets) == 0 {
		return 0, true
	}
	if len(n.goal) < len(n.adj) {
		n.goal = make([]int, len(n.adj))
	}
	n.save()
	for _, t := range targets {
		n.setCap(t.sinkEdge, capacity)
		n.goal[t.node] = t.sinkEdge + 1
	}
	ok = true
	for flow < limit {
		node := n.findPath(-1, true)
		if !e.spend(len(n.queue)) {
			ok = false
			break
		}
		if node < 0 {
			break
		}
		n.pushPath(node, n.goal[node]-1)
		flow++
	}
	for _, t := range targets {
		n.goal[t.node] = 0
	}
	n.restore()
	return flow, ok
}

// pushPath moves one unit along the path findPath left in parent, from the source to node, then out its sink edge.
func (n *flowNet) pushPath(node, sinkEdge int) {
	for v := node; v != flowSource; {
		id := n.parent[v]
		n.push(id)
		v = n.to[id^1]
	}
	n.push(sinkEdge)
}

// findPath searches breadth first for a residual path from the source that avoids the sink, to target, or when
// anyGoal is set to the first probe target whose sink edge is still open. It leaves the edge used to reach each node
// in parent and returns the node reached, or -1.
func (n *flowNet) findPath(target int, anyGoal bool) int {
	if len(n.seen) < len(n.adj) {
		n.seen = make([]uint32, len(n.adj))
		n.parent = make([]int, len(n.adj))
	}
	n.search++
	n.seen[flowSource] = n.search
	n.queue = append(n.queue[:0], flowSource)
	for i := 0; i < len(n.queue); i++ {
		for _, id := range n.adj[n.queue[i]] {
			v := n.to[id]
			if n.cap[id] <= 0 || v == flowSink || n.seen[v] == n.search {
				continue
			}
			n.seen[v] = n.search
			n.parent[v] = id
			if v == target || (anyGoal && n.goal[v] != 0 && n.cap[n.goal[v]-1] > 0) {
				return v
			}
			n.queue = append(n.queue, v)
		}
	}
	return -1
}
