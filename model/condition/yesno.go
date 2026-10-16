// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// outcome is what a YesNo evaluation needs of a node below the root: its state and the reason. It is much smaller than
// a Result, which below the root nobody sees in YesNo mode.
type outcome struct {
	reason string
	state  State
}

func (o outcome) result() Result {
	return Result{State: o.state, Reason: o.reason}
}

// test evaluates a node, which validation has made sure is not nil, in YesNo mode for a parent that needs only its
// outcome. It must only be called while not collecting. Groups, checks and consts have dedicated paths; other nodes
// are evaluated in full.
func (e *evaluator) test(n Node, f frame) outcome {
	if !e.spend(1) {
		return outcome{reason: reasonTooComplex(), state: Failed}
	}
	switch t := n.(type) {
	case *Group:
		return t.test(e, f)
	case *Check:
		return t.test(e, f)
	case *Const:
		return outcome{state: metState(t.Value, false)}
	default:
		r := n.eval(e, f)
		return outcome{reason: r.Reason, state: r.State}
	}
}

// testRef tests the node of a slot; an empty one is Skipped.
func (e *evaluator) testRef(r Ref, f frame) outcome {
	if r.IsZero() {
		return outcome{state: Skipped}
	}
	return e.test(r.Node, f)
}

// test is the YesNo evaluation of a group. The guard comes first, then the children, stopping once nothing further
// can change the state (see decided): all at the first Unmet, any at the first Met, at_least once N are Met or the
// Unmet children alone put N out of reach.
func (g *Group) test(e *evaluator, f frame) outcome {
	if !g.OnlyWhen.IsZero() {
		guard := e.testRef(g.OnlyWhen, f)
		switch guard.state {
		case Unmet:
			return outcome{state: Skipped}
		case Failed:
			return guard
		default:
		}
	}
	if g.Exclusive {
		r := e.exclusiveGroup(g, f)
		return outcome{reason: r.Reason, state: r.State}
	}
	var t groupTally
	for i, child := range g.Children {
		o := e.test(child, f)
		t.add(o.state, o.reason)
		if g.decided(&t, len(g.Children)-i-1) {
			break
		}
	}
	return t.outcome(g)
}
