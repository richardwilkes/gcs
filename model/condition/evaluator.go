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
	"context"
	"sync"

	"github.com/richardwilkes/toolbox/v2/i18n"
)

// DefaultBudget is the work an evaluation pass may do when Options leaves Budget at 0.
const DefaultBudget = 1_000_000

// cancelCheckInterval is how many units of work go by between checks of the context.
const cancelCheckInterval = 256

// Options control an evaluation.
type Options struct {
	Mode Mode
	// Budget bounds the work of each pass, counting a unit for every node evaluated, every member visited and every
	// node the exclusive solver searches; 0 means DefaultBudget. A pass that runs out ends Failed, as too complex to
	// decide, wherever its state depends on what it could not do.
	Budget int
}

// evaluator holds everything one evaluation needs. Nodes are read-only during evaluation; all caches, memos and the
// budget live here. A Matcher keeps one and reuses it, so the field and self-token caches, which depend only on the
// tree and host, outlive a single evaluation.
type evaluator struct {
	ctx          context.Context
	err          error // the context's error, once it is done
	host         Host
	self         Item
	resolver     FieldResolver
	replacements map[itemPair]map[string]string
	scripts      map[scriptKey]scriptOutcome
	fields       map[fieldKey]func(Item) (Value, bool)
	checkFields  map[*Check]func(Item) (Value, bool)
	qualifiers   map[*Check]*selfQualifier
	opt          Options
	budget       int
	ticks        int
	collect      bool
	exhausted    bool
}

// frame is the context a node is evaluated in. The subject becomes each member inside a where, and of then points at
// the key of that where's collection. Self never changes during a call, so the evaluator holds it.
type frame struct {
	subject Item
	of      *string
}

// collection returns the key of the collection whose where the frame is in, or "" for the root subject.
func (f frame) collection() string {
	if f.of == nil {
		return ""
	}
	return *f.of
}

// Evaluate is EvaluateFor with self also as the root subject.
func Evaluate(ctx context.Context, n Node, h Host, self Item, opt Options) (Result, error) {
	return EvaluateFor(ctx, n, h, self, self, opt)
}

// EvaluateFor evaluates n for self with the given root subject. A nil tree is Skipped, and one that Validate finds a
// problem with is Failed, with the first problem as the reason. The error is the context's, when it was done before
// the evaluation finished, which then has no result.
func EvaluateFor(ctx context.Context, n Node, h Host, self, subject Item, opt Options) (Result, error) {
	e, ok := evaluators.Get().(*evaluator)
	if !ok {
		e = &evaluator{}
	}
	e.init(ctx, h, self, opt)
	r := e.run(n, subject)
	err := e.err
	e.release()
	evaluators.Put(e)
	if err != nil {
		return Result{}, err
	}
	return r, nil
}

// evaluators keeps one-shot evaluators, and their memo maps, for reuse. An evaluator holds nothing of a call once it
// returns.
var evaluators sync.Pool

// init readies e for a new call, keeping only its memo maps, emptied.
func (e *evaluator) init(ctx context.Context, h Host, self Item, opt Options) {
	if opt.Budget <= 0 {
		opt.Budget = DefaultBudget
	}
	clear(e.replacements)
	clear(e.scripts)
	*e = evaluator{
		ctx:          ctx,
		err:          ctx.Err(),
		host:         h,
		self:         self,
		opt:          opt,
		replacements: e.replacements,
		scripts:      e.scripts,
	}
}

// release drops what e holds of the call.
func (e *evaluator) release() {
	e.ctx = nil
	e.err = nil
	e.host = nil
	e.self = nil
	clear(e.replacements)
	clear(e.scripts)
}

// run evaluates n as the mode asks: a YesNo pass decides the state, and a Collect pass fills in the details. Collect
// runs only the Collect pass, which reaches the same state; Explain runs it only when the YesNo state doesn't pass. A
// Collect pass that runs out of budget gives way to the YesNo state, marked Truncated.
func (e *evaluator) run(n Node, subject Item) Result {
	if n == nil {
		return Result{State: Skipped}
	}
	if p := firstProblem(n); p != nil {
		r := failure(p.Error())
		r.Node = n
		return r
	}
	f := frame{subject: subject}
	if e.opt.Mode == Collect {
		if r, ok := e.pass(n, f, true); ok {
			return r
		}
	}
	r, _ := e.pass(n, f, false)
	switch e.opt.Mode {
	case Collect:
		r.Truncated = true
	case Explain:
		if r.State.Passes() {
			break
		}
		var ok bool
		var full Result
		if full, ok = e.pass(n, f, true); ok {
			return full
		}
		r.Truncated = true
	default:
	}
	return r
}

// pass evaluates n with a fresh budget, returning false when the budget ran out. Both passes of a call share the
// memos, so a script runs at most once per call.
func (e *evaluator) pass(n Node, f frame, collect bool) (Result, bool) {
	e.budget = e.opt.Budget
	e.collect = collect
	e.exhausted = e.err != nil
	r := e.eval(n, f)
	return r, !e.exhausted
}

// spend charges units of work, checking the context every so often. It returns false once the budget is spent or
// the context is done, after which everything evaluated is Failed.
func (e *evaluator) spend(units int) bool {
	if e.exhausted {
		return false
	}
	e.budget -= units
	if e.ticks += units; e.ticks >= cancelCheckInterval {
		e.ticks = 0
		if e.err = e.ctx.Err(); e.err != nil {
			e.exhausted = true
			return false
		}
	}
	if e.budget < 0 {
		e.exhausted = true
		return false
	}
	return true
}

// eval evaluates one node, which validation has made sure is not nil.
func (e *evaluator) eval(n Node, f frame) Result {
	var r Result
	if e.spend(1) {
		r = n.eval(e, f)
	} else {
		r = failure(reasonTooComplex())
	}
	r.Node = n
	return r
}

// evalRef evaluates the node of a slot; an empty one is Skipped.
func (e *evaluator) evalRef(r Ref, f frame) Result {
	if r.IsZero() {
		return Result{State: Skipped}
	}
	return e.eval(r.Node, f)
}

// evalYesNo evaluates a node without collecting details, as guards are.
func (e *evaluator) evalYesNo(r Ref, f frame) Result {
	collecting := e.collect
	e.collect = false
	result := e.evalRef(r, f)
	e.collect = collecting
	return result
}

func reasonTooComplex() string {
	return i18n.Text("too complex to decide")
}
