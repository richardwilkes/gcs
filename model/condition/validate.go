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
	"slices"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Problem is one thing wrong with a tree: why Decode rejected it, or what Validate found. Path leads from the root to
// the node or operand at fault.
type Problem struct {
	Path Path
	Msg  string
}

// Error implements error.
func (p Problem) Error() string {
	if len(p.Path) == 0 {
		return p.Msg
	}
	return p.Path.String() + ": " + p.Msg
}

// Validate returns the problems found in the tree rooted at n, in pre-order. Validate never changes data. A tree with
// any problem evaluates as Failed, with the first problem as the reason.
func Validate(n Node) []Problem {
	var v validator
	v.root(n)
	return v.problems
}

// firstProblem returns the first problem Validate would report for n, or nil when there is none. It is checked before
// every evaluation, so it first looks for any problem at all, keeping no paths, which allocates nothing.
func firstProblem(n Node) *Problem {
	v := validator{quick: true}
	v.root(n)
	if !v.invalid {
		return nil
	}
	return &Validate(n)[0]
}

// validator collects problems, or in quick mode only notes that there is one.
type validator struct {
	path     Path
	problems []Problem
	depth    int
	quick    bool
	invalid  bool
}

func (v *validator) done() bool {
	return v.invalid
}

func (v *validator) report(msg string) {
	v.reportAt(nil, msg)
}

// reportAt reports a problem at rel, a path relative to the current one.
func (v *validator) reportAt(rel Path, msg string) {
	if v.quick {
		v.invalid = true
		return
	}
	v.problems = append(v.problems, Problem{Path: slices.Concat(v.path, rel), Msg: msg})
}

func (v *validator) root(n Node) {
	if n != nil {
		v.node(n)
	}
}

// node validates n one level deeper, unless that is deeper than MaxDepth.
func (v *validator) node(n Node) {
	if v.deeper() {
		n.validate(v)
	}
	v.depth--
}

// deeper steps one level down and returns true, or reports the step and returns false when it goes past MaxDepth.
// Either way, the caller steps back up.
func (v *validator) deeper() bool {
	v.depth++
	if v.depth > MaxDepth {
		v.report(reasonTooDeep())
		return false
	}
	return true
}

func (v *validator) child(s Slot, n Node) {
	if v.done() {
		return
	}
	v.enter(s)
	if n == nil {
		v.report(i18n.Text("missing child"))
	} else {
		v.node(n)
	}
	v.leave()
}

// enter steps into s, which only a full validation records.
func (v *validator) enter(s Slot) {
	if !v.quick {
		v.path = append(v.path, s)
	}
}

func (v *validator) leave() {
	if !v.quick {
		v.path = v.path[:len(v.path)-1]
	}
}

func (v *validator) ref(kind SlotKind, r Ref) {
	if !r.IsZero() {
		v.child(Slot{Kind: kind}, r.Node)
	}
}

func (v *validator) operand(s Slot, o *Operand) {
	if v.done() {
		return
	}
	v.enter(s)
	if v.deeper() {
		o.validate(v)
	}
	v.depth--
	v.leave()
}
