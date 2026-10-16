// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

import "context"

// Matcher matches items against one tree, as filters do, reusing its scratch space and caches between items. It is
// not safe for concurrent use; give each goroutine its own.
type Matcher struct {
	root    Node
	e       *evaluator
	problem string
}

// NewMatcher returns a Matcher for n. The tree is validated once, here. When h implements FieldResolver, each field a
// check reads is looked up once with FieldFunc and the func cached.
func NewMatcher(n Node, h Host, opt Options) *Matcher {
	if opt.Budget <= 0 {
		opt.Budget = DefaultBudget
	}
	m := &Matcher{root: n, e: &evaluator{host: h, opt: opt, qualifiers: make(map[*Check]*selfQualifier)}}
	if p := firstProblem(n); p != nil {
		m.problem = p.Error()
	}
	if resolver, ok := h.(FieldResolver); ok {
		m.e.resolver = resolver
		m.e.fields = make(map[fieldKey]func(Item) (Value, bool))
		m.e.checkFields = make(map[*Check]func(Item) (Value, bool))
	}
	return m
}

// State returns the state of the tree for item, as Evaluate with YesNo would. A nil tree is Skipped, and one with a
// problem Failed. The error is the context's, when it was done before the state was decided.
func (m *Matcher) State(ctx context.Context, item Item) (State, error) {
	switch {
	case m.root == nil:
		return Skipped, nil
	case m.problem != "":
		return Failed, nil
	}
	e := m.e
	clear(e.replacements)
	clear(e.scripts)
	e.ctx = ctx
	e.self = item
	e.budget = e.opt.Budget
	e.exhausted = false
	state := e.test(m.root, frame{subject: item}).state
	err := e.err
	e.ctx, e.err, e.self = nil, nil, nil
	if err != nil {
		return Failed, err
	}
	return state, nil
}
