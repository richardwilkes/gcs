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
	"encoding/json/v2"
	"hash"
	"sync/atomic"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/zeebo/xxh3"
)

// fakeItem is an in-memory host object: its fields, and the relative collections it owns.
type fakeItem struct {
	fields   map[string]Value
	relative map[string][]*fakeItem
	name     string
}

func newItem(name string, fields map[string]Value) *fakeItem {
	if fields == nil {
		fields = make(map[string]Value)
	}
	return &fakeItem{name: name, fields: fields}
}

// with adds a relative collection to the item and returns it.
func (it *fakeItem) with(key string, members ...*fakeItem) *fakeItem {
	if it.relative == nil {
		it.relative = make(map[string][]*fakeItem)
	}
	it.relative[key] = members
	return it
}

// fakeScript is the outcome of a fake script, which the host looks up by the script's text.
type fakeScript struct {
	reason string
	state  State
}

// fakeHost is an in-memory Host. Entity collections are listed in full, self included; relative collections belong
// to the subject. Every item has a "name" field, and the fields it lists. Scripts are looked up by their text and
// counted; the script "block" waits until the context is done.
type fakeHost struct {
	entity       map[string][]*fakeItem
	attributes   map[string]fxp.Int
	entities     map[string]fxp.Int
	replacements map[string]string
	scripts      map[string]fakeScript
	runs         atomic.Int64
	replaced     atomic.Int64
}

func newHost() *fakeHost {
	return &fakeHost{
		entity:     make(map[string][]*fakeItem),
		attributes: make(map[string]fxp.Int),
		entities:   make(map[string]fxp.Int),
		scripts: map[string]fakeScript{
			"met":     {state: Met},
			"unmet":   {state: Unmet},
			"skipped": {state: Skipped},
			"failed":  {state: Failed, reason: "script failed"},
		},
	}
}

func (h *fakeHost) Collection(key string, _, subject Item) ([]Item, bool) {
	list, ok := h.entity[key]
	if !ok {
		var it *fakeItem
		if it, ok = subject.(*fakeItem); ok {
			list, ok = it.relative[key]
		}
	}
	if !ok {
		return nil, false
	}
	items := make([]Item, len(list))
	for i, one := range list {
		items[i] = one
	}
	return items, true
}

func (h *fakeHost) Field(item Item, key string) (Value, bool) {
	it, ok := item.(*fakeItem)
	if !ok {
		return Value{}, false
	}
	if v, has := it.fields[key]; has {
		return v, true
	}
	if key == "name" {
		return Value{Kind: ValueText, Text: it.name}, true
	}
	return Value{}, false
}

func (h *fakeHost) Attribute(key string) (fxp.Int, bool) {
	v, ok := h.attributes[key]
	return v, ok
}

func (h *fakeHost) Entity(key string) (fxp.Int, bool) {
	v, ok := h.entities[key]
	return v, ok
}

func (h *fakeHost) Replacements(_, _ Item) map[string]string {
	h.replaced.Add(1)
	return h.replacements
}

func (h *fakeHost) RunScript(ctx context.Context, s *Script, _, _ Item) (state State, reason string) {
	h.runs.Add(1)
	if s.Script == "block" {
		<-ctx.Done()
		return Failed, ctx.Err().Error()
	}
	out, ok := h.scripts[s.Script]
	if !ok {
		return Failed, "no such script"
	}
	return out.state, out.reason
}

// resolvingHost adds a FieldResolver to fakeHost. It resolves every field except "unresolved", and counts how often
// it is asked.
type resolvingHost struct {
	*fakeHost
	asked atomic.Int64
}

func (h *resolvingHost) FieldFunc(_, key string) (func(Item) (Value, bool), bool) {
	h.asked.Add(1)
	if key == "unresolved" {
		return nil, false
	}
	return func(item Item) (Value, bool) { return h.Field(item, key) }, true
}

func textValue(s string) Value {
	return Value{Kind: ValueText, Text: s}
}

func listValue(list ...string) Value {
	return Value{Kind: ValueList, List: list}
}

func numberValue(n int64) Value {
	return Value{Kind: ValueNumber, Number: fxp.FromInteger(n)}
}

func weightValue(n int64) Value {
	return Value{Kind: ValueWeight, Number: fxp.FromInteger(n)}
}

func boolValue(b bool) Value {
	return Value{Kind: ValueBool, Bool: b}
}

func absentValue() Value {
	return Value{Kind: ValueAbsent}
}

func intPtr(v int64) *fxp.Int {
	x := fxp.FromInteger(v)
	return &x
}

func weightPtr(v int64) *fxp.Weight {
	x := fxp.Weight(fxp.FromInteger(v))
	return &x
}

func textCrit(cmp criteria.StringComparison, qualifier string) criteria.Text {
	return criteria.Text{Compare: cmp, Qualifier: qualifier}
}

func numberCrit(cmp criteria.NumericComparison, qualifier int64) criteria.Number {
	return criteria.Number{Compare: cmp, Qualifier: fxp.FromInteger(qualifier)}
}

func numberCritPtr(cmp criteria.NumericComparison, qualifier int64) *criteria.Number {
	n := numberCrit(cmp, qualifier)
	return &n
}

func weightCrit(cmp criteria.NumericComparison, qualifier int64) criteria.Weight {
	return criteria.Weight{Compare: cmp, Qualifier: fxp.Weight(fxp.FromInteger(qualifier))}
}

func atLeast(n int64) *criteria.Number {
	return numberCritPtr(criteria.AtLeastNumber, n)
}

func checkText(field string, cmp criteria.StringComparison, qualifier string) *Check {
	return &Check{Field: field, Text: textCrit(cmp, qualifier)}
}

func value(n int64) Operand {
	return Operand{Value: intPtr(n)}
}

// leafFor returns a leaf that evaluates to the given state: a const for Met and Unmet, and a script otherwise.
func leafFor(s State) Node {
	switch s {
	case Met:
		return &Const{Value: true}
	case Unmet:
		return &Const{}
	case Skipped:
		return &Script{Script: "skipped"}
	default:
		return &Script{Script: "failed"}
	}
}

var (
	yesNo   = Options{}
	collect = Options{Mode: Collect}
	explain = Options{Mode: Explain}
)

// eval evaluates n for self with a background context, failing the test on an error.
func eval(t testing.TB, n Node, h Host, self Item, opt Options) Result {
	t.Helper()
	r, err := Evaluate(context.Background(), n, h, self, opt)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	return r
}

// evalFor evaluates n for self with the given subject and a background context, failing the test on an error.
func evalFor(t testing.TB, n Node, h Host, self, subject Item, opt Options) Result {
	t.Helper()
	r, err := EvaluateFor(context.Background(), n, h, self, subject, opt)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	return r
}

// evalAll evaluates n in every mode, fails the test unless they agree on the state, and returns the Collect result.
func evalAll(t testing.TB, n Node, h Host, self Item) Result {
	t.Helper()
	yes := eval(t, n, h, self, yesNo)
	full := eval(t, n, h, self, collect)
	why := eval(t, n, h, self, explain)
	if yes.State != full.State || yes.State != why.State {
		t.Fatalf("modes disagree: yes/no %v, collect %v, explain %v", yes.State, full.State, why.State)
	}
	return full
}

// hits returns the items of r's hits for node n, in order.
func hits(r *Result, n Node) []Item {
	var list []Item
	for hit := range r.Hits() {
		if hit.Node == n {
			list = append(list, hit.Item)
		}
	}
	return list
}

func nodeHash(n Node) uint64 {
	return hashWith(func(h hash.Hash) { HashNode(h, n) })
}

func hashWith(fn func(h hash.Hash)) uint64 {
	h := xxh3.New()
	fn(h)
	return h.Sum64()
}

func encode(t testing.TB, v any) string {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return string(data)
}

func decode(t testing.TB, data string) Node {
	t.Helper()
	n, err := Decode([]byte(data))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	return n
}
