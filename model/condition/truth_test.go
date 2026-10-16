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
	"github.com/richardwilkes/toolbox/v2/check"
)

var allStates = []State{Met, Unmet, Skipped, Failed}

// stateSequences returns every sequence of states of length 0 to max.
func stateSequences(maxLen int) [][]State {
	result := make([][]State, 1, 128)
	frontier := [][]State{nil}
	for range maxLen {
		next := make([][]State, 0, len(allStates)*len(frontier))
		for _, seq := range frontier {
			for _, s := range allStates {
				next = append(next, append(append([]State(nil), seq...), s))
			}
		}
		result = append(result, next...)
		frontier = next
	}
	return result
}

// expectedGroupState is the reference model, written independently of the engine: the legacy rule of PR #1182, where
// a child that decides the group does so even when others failed.
func expectedGroupState(mode GroupMode, n int, not bool, guard *State, states []State) State {
	if guard != nil {
		switch *guard {
		case Unmet:
			return Skipped
		case Failed:
			return Failed
		default:
		}
	}
	applicable, met, unmet, failed := 0, 0, 0, 0
	for _, s := range states {
		switch s {
		case Skipped:
			continue
		case Met:
			met++
		case Unmet:
			unmet++
		default:
			failed++
		}
		applicable++
	}
	if applicable == 0 {
		return Skipped
	}
	need := applicable
	switch mode {
	case GroupAny:
		need = 1
	case GroupAtLeast:
		need = min(n, applicable)
	default:
	}
	met = min(met, need)
	switch {
	case met == need:
		if not {
			return Unmet
		}
		return Met
	case applicable-unmet >= need && failed > 0: // the threshold is still reachable, so the Failed ones decide
		return Failed
	case not:
		return Met
	default:
		return Unmet
	}
}

// TestGroupTruthTables checks all, any and at_least(N) over every sequence of up to three child states, with and
// without Not and each kind of guard, in every mode.
func TestGroupTruthTables(t *testing.T) {
	c := check.New(t)
	h := newHost()
	guards := []*State{nil, new(Met), new(Unmet), new(Skipped), new(Failed)}
	for _, states := range stateSequences(3) {
		children := make(Nodes, len(states))
		for i, s := range states {
			children[i] = leafFor(s)
		}
		for _, mode := range GroupModes {
			ns := []int{0}
			if mode == GroupAtLeast {
				ns = nil
				for n := 1; n <= len(states); n++ {
					ns = append(ns, n)
				}
			}
			for _, n := range ns {
				for _, not := range []bool{false, true} {
					for _, guard := range guards {
						g := &Group{Mode: mode, N: n, Not: not, Children: children}
						if guard != nil {
							g.OnlyWhen = Ref{Node: leafFor(*guard)}
						}
						want := expectedGroupState(mode, n, not, guard, states)
						desc := fmt.Sprintf("%v n=%d not=%v guard=%v states=%v", mode, n, not, guardName(guard), states)
						full := evalAll(t, g, h, nil)
						c.Equal(want, full.State, desc)
						wantChildren := len(states)
						if guard != nil && *guard == Failed {
							wantChildren = 0
						}
						c.Equal(wantChildren, len(full.Children), desc+" (children)")
					}
				}
			}
		}
	}
}

// TestFailuresReachCounts checks that a where that ends Failed for any member fails its count whatever the
// quantity or Not, since the count can't be known, and that a where with no member to fail on doesn't.
func TestFailuresReachCounts(t *testing.T) {
	c := check.New(t)
	h := newHost()
	self := newItem("self", nil).with("kids", newItem("kid", nil)).with("none")
	failing := Ref{Node: &Script{Script: "failed"}}
	for _, tc := range []struct {
		node  Node
		state State
	}{
		{&Count{Of: "kids", Where: failing}, Failed},
		{&Count{Of: "kids", Where: failing, Not: true}, Failed},
		{&Count{Of: "kids", Where: failing, Quantity: &criteria.Number{}}, Failed},
		{&Count{Of: "kids", Where: Ref{Node: &Check{Field: "level"}}}, Failed},
		{&Count{Of: "none", Where: failing, Not: true}, Met},
		{&Group{Mode: GroupAny, Children: Nodes{&Count{Of: "kids", Where: failing}, &Const{Value: true}}}, Met},
	} {
		r := evalAll(t, tc.node, h, self)
		c.Equal(tc.state, r.State, encode(t, tc.node))
	}
}

func guardName(guard *State) string {
	if guard == nil {
		return "none"
	}
	return guard.String()
}
