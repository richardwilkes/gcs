// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package condition

// The memos and caches of an evaluator. Both passes of a call share the memos, so a script runs at most once per
// (node, self, subject) in a call and the host's replacements are fetched once per (self, subject). A Matcher's field
// cache outlives a call, since it depends only on the tree and the host.

type itemPair struct {
	self    Item
	subject Item
}

type scriptKey struct {
	script  *Script
	self    Item
	subject Item
}

type scriptOutcome struct {
	reason string
	state  State
}

type fieldKey struct {
	of  string
	key string
}

// replacementsFor returns the host's nameable replacements for the frame, memoized per (self, subject).
func (e *evaluator) replacementsFor(f frame) map[string]string {
	key := itemPair{self: e.self, subject: f.subject}
	m, ok := e.replacements[key]
	if !ok {
		m = e.host.Replacements(e.self, f.subject)
		if e.replacements == nil {
			e.replacements = make(map[itemPair]map[string]string)
		}
		e.replacements[key] = m
	}
	return m
}

// runScript runs a script through the host, memoized per (node, self, subject). A script is where an evaluation is
// most likely to wait, so the context is checked once it returns.
func (e *evaluator) runScript(s *Script, f frame) scriptOutcome {
	key := scriptKey{script: s, self: e.self, subject: f.subject}
	out, ok := e.scripts[key]
	if !ok {
		out.state, out.reason = e.host.RunScript(e.ctx, s, e.self, f.subject)
		if e.err = e.ctx.Err(); e.err != nil {
			e.exhausted = true
		}
		if e.scripts == nil {
			e.scripts = make(map[scriptKey]scriptOutcome)
		}
		e.scripts[key] = out
	}
	return out
}

// field returns a field of item. Inside a Matcher, of picks the cached FieldFunc the host's FieldResolver gave.
func (e *evaluator) field(item Item, of, key string) (Value, bool) {
	if e.fields != nil {
		if fn := e.fieldFunc(of, key); fn != nil {
			return fn(item)
		}
	}
	return e.host.Field(item, key)
}

// checkField returns the field a check reads from the subject. Inside a Matcher, the FieldFunc is also cached per
// check, which saves hashing the field key on every row. A check shared between positions with different collections
// gets the func of the first: FieldFunc must behave exactly like Field, so which one it gets makes no difference.
func (e *evaluator) checkField(c *Check, f frame) (Value, bool) {
	if e.fields != nil {
		fn, cached := e.checkFields[c]
		if !cached {
			fn = e.fieldFunc(f.collection(), c.Field)
			e.checkFields[c] = fn
		}
		if fn != nil {
			return fn(f.subject)
		}
	}
	return e.host.Field(f.subject, c.Field)
}

// fieldFunc returns the resolver's func for (of, key), asking it once; nil when it has none.
func (e *evaluator) fieldFunc(of, key string) func(Item) (Value, bool) {
	k := fieldKey{of: of, key: key}
	fn, cached := e.fields[k]
	if !cached {
		var ok bool
		if fn, ok = e.resolver.FieldFunc(of, key); !ok {
			fn = nil
		}
		e.fields[k] = fn
	}
	return fn
}
