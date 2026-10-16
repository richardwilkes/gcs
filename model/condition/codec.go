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
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
)

// Decode decodes a tree. Decoding is strict: an unknown node type, an unknown member, an unknown key for an enum or a
// comparison, a missing node and a tree nested deeper than MaxDepth are all errors, reported as a Problem whose path
// leads to the node at fault. null and the empty group decode as nil.
func Decode(data []byte) (Node, error) {
	raw := jsontext.Value(data)
	if !raw.IsValid() {
		return nil, Problem{Msg: i18n.Text("not valid JSON")}
	}
	return decodeRef(raw, 1)
}

// decodeRef decodes the node of a Ref slot at the given depth. An absent or null slot and the empty group both become
// nil.
func decodeRef(raw jsontext.Value, depth int) (Node, error) {
	if len(raw) == 0 || raw.Kind() == 'n' {
		return nil, nil
	}
	n, err := decodeNode(raw, depth)
	if err != nil || (Ref{Node: n}).IsZero() {
		return nil, err
	}
	return n, nil
}

// decodeNodes decodes a list of nodes, each at the given depth.
func decodeNodes(list []jsontext.Value, depth int) (Nodes, error) {
	if list == nil {
		return nil, nil
	}
	nodes := make(Nodes, len(list))
	for i, raw := range list {
		n, err := decodeNode(raw, depth)
		if err != nil {
			return nil, at(Slot{Kind: SlotChild, Index: i}, err)
		}
		nodes[i] = n
	}
	return nodes, nil
}

// decodeNode decodes one node at the given depth.
func decodeNode(raw jsontext.Value, depth int) (Node, error) {
	switch {
	case raw.Kind() == 'n':
		return nil, problem(i18n.Text("missing node"))
	case raw.Kind() != '{':
		return nil, problem(i18n.Text("a node is not an object"))
	case depth > MaxDepth:
		return nil, problem(reasonTooDeep())
	}
	var head struct {
		Type *string `json:"type"`
	}
	if err := json.Unmarshal(raw, &head); err != nil {
		return nil, wireError(err)
	}
	if head.Type == nil {
		return nil, problem(i18n.Text(`missing "type"`))
	}
	switch *head.Type {
	case typeGroup:
		return decodeGroup(raw, depth)
	case typeCheck:
		return decodeCheck(raw)
	case typeConst:
		return decodeConst(raw)
	case typeCount:
		return decodeCount(raw, depth)
	case typeCompare:
		return decodeCompare(raw, depth)
	case typeScript:
		return decodeScript(raw)
	default:
		return nil, problem(fmt.Sprintf(i18n.Text("unknown node type %q"), *head.Type))
	}
}

// unmarshalWire reads raw into the wire form w, rejecting unknown members.
func unmarshalWire(raw jsontext.Value, w any) error {
	if err := json.Unmarshal(raw, w, json.RejectUnknownMembers(true)); err != nil {
		return wireError(err)
	}
	return nil
}

// wireError returns the Problem a failed unmarshal of a wire form reports. It names the member at fault, but never
// repeats the underlying error, which may carry a stack trace.
func wireError(err error) error {
	var member string
	if se := (*json.SemanticError)(nil); errors.As(err, &se) {
		member = strings.Join(slices.Collect(se.JSONPointer.Tokens()), ".")
	}
	if errors.Is(err, json.ErrUnknownName) {
		return problem(fmt.Sprintf(i18n.Text("unknown member %q"), member))
	}
	return problem(fmt.Sprintf(i18n.Text("invalid value for %q"), member))
}

func problem(msg string) error {
	return Problem{Msg: msg}
}

// at returns err, which decoding always makes a Problem, with s put in front of its path.
func at(s Slot, err error) error {
	var p Problem
	errors.As(err, &p)
	p.Path = append(Path{s}, p.Path...)
	return p
}

func reasonTooDeep() string {
	return fmt.Sprintf(i18n.Text("nested more than %d deep"), MaxDepth)
}

// typedShadow writes a node's wire type ahead of the fields of its method-less shadow type.
type typedShadow[T any] struct {
	Type   string `json:"type"`
	Fields *T     `json:",embed"`
}

// encodeTyped writes a node through its shadow type, with the wire type first.
func encodeTyped[T any](enc *jsontext.Encoder, typ string, fields *T) error {
	return json.MarshalEncode(enc, &typedShadow[T]{Type: typ, Fields: fields})
}

// requiredKey resolves a persisted enum key that must be present and known.
func requiredKey[E any](key *string, member string, extract func(string) (E, bool)) (E, error) {
	if key == nil {
		var zero E
		return zero, problem(fmt.Sprintf(i18n.Text("missing %q"), member))
	}
	return knownKey(*key, member, extract)
}

// knownKey resolves a persisted enum key that must be known.
func knownKey[E any](key, member string, extract func(string) (E, bool)) (E, error) {
	v, ok := extract(key)
	if !ok {
		return v, problem(fmt.Sprintf(i18n.Text("unknown %q value %q"), member, key))
	}
	return v, nil
}

// criterionWire is the raw form of a criterion: a comparison, which must be known, and a qualifier.
type criterionWire[Q any] struct {
	Compare   *string `json:"compare"`
	Qualifier *Q      `json:"qualifier"`
}

// read resolves the comparison and qualifier of a criterion of a check, which must compare something: one that is
// present needs a comparison other than "any".
func read[C comparable, Q any](w *criterionWire[Q], member string, extract func(string) (C, bool)) (cmp C, q Q, err error) {
	if w == nil {
		return cmp, q, nil
	}
	var none C
	if cmp, err = requiredKey(w.Compare, member+".compare", extract); err == nil && cmp == none {
		err = problem(fmt.Sprintf(i18n.Text("%q compares nothing"), member))
	}
	if w.Qualifier != nil {
		q = *w.Qualifier
	}
	return cmp, q, err
}

func readText(w *criterionWire[string], member string) (criteria.Text, error) {
	cmp, q, err := read(w, member, criteria.ExtractKnownStringComparison)
	return criteria.Text{Compare: cmp, Qualifier: q}, err
}

func readNumber(w *criterionWire[fxp.Int], member string) (criteria.Number, error) {
	cmp, q, err := read(w, member, criteria.ExtractKnownNumericComparison)
	return criteria.Number{Compare: cmp, Qualifier: q}, err
}

func readWeight(w *criterionWire[fxp.Weight], member string) (criteria.Weight, error) {
	cmp, q, err := read(w, member, criteria.ExtractKnownNumericComparison)
	return criteria.Weight{Compare: cmp, Qualifier: q}, err
}

// readQuantity resolves a count quantity or per-group: nil when absent, and "any" allowed, written as {}, but never
// with a qualifier.
func readQuantity(w *criterionWire[fxp.Int], member string) (*criteria.Number, error) {
	if w == nil {
		return nil, nil
	}
	cmp := criteria.AnyNumber
	if w.Compare != nil {
		var err error
		if cmp, err = knownKey(*w.Compare, member+".compare", criteria.ExtractKnownNumericComparison); err != nil {
			return nil, err
		}
	}
	n := &criteria.Number{Compare: cmp}
	if w.Qualifier != nil {
		if cmp == criteria.AnyNumber {
			return nil, problem(fmt.Sprintf(i18n.Text("%q has a qualifier but compares nothing"), member))
		}
		n.Qualifier = *w.Qualifier
	}
	return canonicalQuantity(n), nil
}

// quantityRef writes a Count quantity exactly when it is non-nil, "any" as {}. json/v2's omitzero on the pointer
// itself would drop a non-nil "any", since criteria.Number reports itself as zero then.
type quantityRef struct {
	p *criteria.Number
}

// IsZero implements json's omitzero check.
func (q quantityRef) IsZero() bool {
	return q.p == nil
}

// MarshalJSONTo implements json.MarshalerTo.
func (q quantityRef) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, &q.p.NumberData)
}
