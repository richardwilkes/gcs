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
	"errors"
	"hash"

	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Node = &Compare{}

// Compare compares two operands. Wire type "compare".
type Compare struct {
	Left  Operand   `json:"left"`
	Op    CompareOp `json:"op"`
	Right Operand   `json:"right"`
}

// compareFields is the method-less shadow Compare is written through.
type compareFields Compare

// compareWire is the raw form Compare is read through.
type compareWire struct {
	Type  string         `json:"type"`
	Left  jsontext.Value `json:"left"`
	Op    *string        `json:"op"`
	Right jsontext.Value `json:"right"`
}

// Clone implements Node.
func (c *Compare) Clone() Node {
	return &Compare{Left: c.Left.clone(), Op: c.Op, Right: c.Right.clone()}
}

// Hash implements Node.
func (c *Compare) Hash(h hash.Hash) {
	xhash.Num8(h, hashCompare)
	c.Left.hash(h)
	xhash.StringWithLen(h, c.Op.Key())
	c.Right.hash(h)
}

// FillWithNameableKeys implements Node.
func (c *Compare) FillWithNameableKeys(m, existing map[string]string) {
	c.Left.fillWithNameableKeys(m, existing)
	c.Right.fillWithNameableKeys(m, existing)
}

func (c *Compare) validate(v *validator) {
	if c.Op != c.Op.EnsureValid() {
		v.report(i18n.Text("unknown operator"))
	}
	leftNumber, leftWeight := c.Left.constants()
	rightNumber, rightWeight := c.Right.constants()
	if (leftNumber || rightNumber) && (leftWeight || rightWeight) {
		v.report(i18n.Text("number and weight constants are mixed"))
	}
	v.operand(Slot{Kind: SlotLeft}, &c.Left)
	v.operand(Slot{Kind: SlotRight}, &c.Right)
}

func (c *Compare) walk(w *walker) {
	w.operand(Slot{Kind: SlotLeft}, &c.Left)
	w.operand(Slot{Kind: SlotRight}, &c.Right)
}

// MarshalJSONTo implements json.MarshalerTo.
func (c *Compare) MarshalJSONTo(enc *jsontext.Encoder) error {
	if c.Op != c.Op.EnsureValid() {
		return errors.New(i18n.Text("unknown operator"))
	}
	return encodeTyped(enc, typeCompare, (*compareFields)(c))
}

func decodeCompare(raw jsontext.Value, depth int) (Node, error) {
	var w compareWire
	if err := unmarshalWire(raw, &w); err != nil {
		return nil, err
	}
	op, err := requiredKey(w.Op, "op", ExtractKnownCompareOp)
	if err != nil {
		return nil, err
	}
	c := &Compare{Op: op}
	if err = c.Left.decodeSide(w.Left, depth+1); err != nil {
		return nil, at(Slot{Kind: SlotLeft}, err)
	}
	if err = c.Right.decodeSide(w.Right, depth+1); err != nil {
		return nil, at(Slot{Kind: SlotRight}, err)
	}
	return c, nil
}

// eval implements Node. Both operands must resolve, and a Number may not be compared with a Weight.
func (c *Compare) eval(e *evaluator, f frame) Result {
	oe := operandEval{e: e, collect: e.collect}
	left, fail := oe.resolve(&c.Left, f)
	if fail != nil {
		return *fail
	}
	right, fail := oe.resolve(&c.Right, f)
	if fail != nil {
		return *fail
	}
	kind, fail := combineKinds(left.kind, right.kind)
	if fail != nil {
		return *fail
	}
	if kind == ValueAbsent {
		kind = ValueNumber
	}
	r := Result{
		State:    metState(c.Op.Compare(left.value, right.value), false),
		Have:     left.value,
		HaveKind: kind,
		Need:     right.value,
	}
	if oe.collect {
		r.hits = append(make([]Hit, 0, len(oe.contrib)+1), Hit{Node: c})
		for i := range oe.contrib {
			one := &oe.contrib[i]
			r.hits = append(r.hits, Hit{Node: c, Item: one.item})
			if one.where.Node != nil {
				r.Children = append(r.Children, one.where)
			}
		}
		if r.State == Met {
			r.Specificity = 1
		}
	}
	return r
}
