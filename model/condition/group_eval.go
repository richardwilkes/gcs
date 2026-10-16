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
	"cmp"
	"slices"
)

// eval implements Node. The guard comes first: Unmet makes the group Skipped without evaluating its children, Failed
// makes it Failed, and Met or Skipped means the group applies.
func (g *Group) eval(e *evaluator, f frame) Result {
	if !e.collect {
		return g.test(e, f).result()
	}
	if !g.OnlyWhen.IsZero() {
		guard := e.evalYesNo(g.OnlyWhen, f)
		switch guard.State {
		case Unmet:
			r := Result{State: Skipped, Children: make([]Result, len(g.Children))}
			for i, child := range g.Children {
				r.Children[i] = Result{Node: child, State: Skipped}
			}
			return r
		case Failed:
			return failure(guard.Reason)
		default:
		}
	}
	if g.Exclusive {
		return e.exclusiveGroup(g, f)
	}
	var t groupTally
	children := make([]Result, len(g.Children))
	for i, child := range g.Children {
		children[i] = e.eval(child, f)
		t.add(children[i].State, children[i].Reason)
	}
	r := t.outcome(g).result()
	r.Children = children
	if r.State == Met && !g.Not {
		r.Specificity = g.specificity(children, t.applicable)
	}
	return r
}

// groupTally counts the states of a group's applicable children.
type groupTally struct {
	reason     string // of the first Failed child
	met        int
	unmet      int
	applicable int
	failed     int
}

func (t *groupTally) add(state State, reason string) {
	switch state {
	case Skipped:
		return
	case Met:
		t.met++
	case Unmet:
		t.unmet++
	default:
		t.failed++
		if t.reason == "" {
			t.reason = reason
		}
	}
	t.applicable++
}

// outcome returns the state the tally gives a group with this mode and Not: Skipped when no child applies, Met when
// enough are Met, Unmet when the Unmet children alone leave too few to reach the number required, so that no Failed
// child could have changed it, otherwise Failed when any applicable child is, and Unmet when none is. Not then swaps
// Met and Unmet.
func (t *groupTally) outcome(g *Group) outcome {
	switch {
	case t.applicable == 0:
		return outcome{state: Skipped}
	case t.met >= g.required(t.applicable):
		return outcome{state: metState(true, g.Not)}
	case t.failed > 0 && t.unmet <= t.applicable-g.required(t.applicable):
		return outcome{reason: t.reason, state: Failed}
	default:
		return outcome{state: metState(false, g.Not)}
	}
}

// required returns how many applicable children must be Met.
func (g *Group) required(applicable int) int {
	switch g.Mode {
	case GroupAny:
		return 1
	case GroupAtLeast:
		return min(g.N, applicable)
	default:
		return applicable
	}
}

// decided reports whether a YesNo evaluation can stop with remaining children still to go: once enough are Met
// whatever the rest are (any at one, at_least at N; all only at the end), or once the Unmet ones alone put the number
// required out of reach even if every remaining child applies and is Met, which makes the group Unmet whatever the
// rest are, Failed ones included.
func (g *Group) decided(t *groupTally, remaining int) bool {
	possible := t.applicable + remaining
	if t.unmet > possible-g.required(possible) {
		return true
	}
	switch g.Mode {
	case GroupAny:
		return t.met >= 1
	case GroupAtLeast:
		return t.met >= g.N
	default:
		return false
	}
}

// specificity returns the specificity of a Met group: the sum over applicable children for all, the maximum over Met
// children for any, and the sum of the best min(N, applicable) Met children for at_least.
func (g *Group) specificity(children []Result, applicable int) int {
	var specs []int
	for i := range children {
		if children[i].State == Met {
			specs = append(specs, children[i].Specificity)
		}
	}
	switch g.Mode {
	case GroupAny:
		return slices.Max(specs)
	case GroupAtLeast:
		slices.SortFunc(specs, func(a, b int) int { return cmp.Compare(b, a) })
		specs = specs[:min(g.N, applicable)]
	default:
	}
	total := 0
	for _, one := range specs {
		total += one
	}
	return total
}
