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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Node = &Count{}

// Members selects the members of a collection that pass a where. It is embedded (plain Go embedding) in Count and
// Operand. The owner (self) is left out of the collection unless IncludeSelf is set.
type Members struct {
	Of          string `json:"of,omitzero"`
	IncludeSelf bool   `json:"include_self,omitzero"`
	Where       Ref    `json:"where,omitzero"` // nil matches every member
}

func (m *Members) cloneMembers() Members {
	return Members{Of: m.Of, IncludeSelf: m.IncludeSelf, Where: m.Where.clone()}
}

func (m *Members) hashMembers(h hash.Hash) {
	xhash.StringWithLen(h, m.Of)
	xhash.Bool(h, m.IncludeSelf)
	m.Where.hash(h)
}

// decodeMembers decodes the where of members at the given depth.
func (m *Members) decodeWhere(raw jsontext.Value, depth int) error {
	var err error
	if m.Where.Node, err = decodeRef(raw, depth); err != nil {
		return at(Slot{Kind: SlotWhere}, err)
	}
	return nil
}

// Count counts the members of a collection that pass a where, optionally grouped by the values of a field. An
// exclusive grouped count counts each member toward one of its values only. Wire type "count".
type Count struct {
	Members
	Quantity  *criteria.Number `json:"quantity"` // nil means at least 1; written exactly when non-nil, "any" as {}
	GroupBy   string           `json:"group_by,omitzero"`
	PerGroup  *criteria.Number `json:"per_group"` // nil means at least 1; written like Quantity
	Exclusive bool             `json:"exclusive,omitzero"`
	Not       bool             `json:"not,omitzero"`
}

// countFields is the field-by-field shadow Count is written through, with wrappers that write Quantity and PerGroup
// exactly when they are non-nil.
type countFields struct {
	Of          string      `json:"of,omitzero"`
	IncludeSelf bool        `json:"include_self,omitzero"`
	Where       Ref         `json:"where,omitzero"`
	Quantity    quantityRef `json:"quantity,omitzero"`
	GroupBy     string      `json:"group_by,omitzero"`
	PerGroup    quantityRef `json:"per_group,omitzero"`
	Exclusive   bool        `json:"exclusive,omitzero"`
	Not         bool        `json:"not,omitzero"`
}

// countWire is the raw form Count is read through.
type countWire struct {
	Type        string                  `json:"type"`
	Of          string                  `json:"of"`
	IncludeSelf bool                    `json:"include_self"`
	Where       jsontext.Value          `json:"where"`
	Quantity    *criterionWire[fxp.Int] `json:"quantity"`
	GroupBy     string                  `json:"group_by"`
	PerGroup    *criterionWire[fxp.Int] `json:"per_group"`
	Exclusive   bool                    `json:"exclusive"`
	Not         bool                    `json:"not"`
}

// Clone implements Node.
func (c *Count) Clone() Node {
	other := *c
	other.Members = c.cloneMembers()
	other.Quantity = clonePtr(c.Quantity)
	other.PerGroup = clonePtr(c.PerGroup)
	return &other
}

// Hash implements Node.
func (c *Count) Hash(h hash.Hash) {
	xhash.Num8(h, hashCount)
	c.hashMembers(h)
	hashQuantity(h, c.Quantity)
	xhash.StringWithLen(h, c.GroupBy)
	hashQuantity(h, c.PerGroup)
	xhash.Bool(h, c.Exclusive)
	xhash.Bool(h, c.Not)
}

// FillWithNameableKeys implements Node.
func (c *Count) FillWithNameableKeys(m, existing map[string]string) {
	c.Where.fillWithNameableKeys(m, existing)
}

func (c *Count) validate(v *validator) {
	if c.Of == "" {
		v.report(i18n.Text("no collection"))
	}
	if c.GroupBy == "" && c.PerGroup != nil {
		v.report(i18n.Text("a per-group quantity without a group-by field"))
	}
	if c.Exclusive {
		switch {
		case c.GroupBy == "":
			v.report(i18n.Text("exclusive without a group-by field"))
		case !atLeastWhole(c.PerGroup):
			v.report(i18n.Text(`an exclusive count's per-group quantity is not "at least" a whole number`))
		default:
		}
	}
	v.ref(SlotWhere, c.Where)
}

func (c *Count) walk(w *walker) {
	w.ref(SlotWhere, c.Where)
}

// MarshalJSONTo implements json.MarshalerTo.
func (c *Count) MarshalJSONTo(enc *jsontext.Encoder) error {
	return encodeTyped(enc, typeCount, &countFields{
		Of:          c.Of,
		IncludeSelf: c.IncludeSelf,
		Where:       c.Where,
		Quantity:    quantityRef{p: canonicalQuantity(c.Quantity)},
		GroupBy:     c.GroupBy,
		PerGroup:    quantityRef{p: canonicalQuantity(c.PerGroup)},
		Exclusive:   c.Exclusive,
		Not:         c.Not,
	})
}

func decodeCount(raw jsontext.Value, depth int) (Node, error) {
	var w countWire
	if err := unmarshalWire(raw, &w); err != nil {
		return nil, err
	}
	c := &Count{
		Of: w.Of, IncludeSelf: w.IncludeSelf,
		GroupBy:   w.GroupBy,
		Exclusive: w.Exclusive,
		Not:       w.Not,
	}
	var err error
	if c.Quantity, err = readQuantity(w.Quantity, "quantity"); err != nil {
		return nil, err
	}
	if c.PerGroup, err = readQuantity(w.PerGroup, "per_group"); err != nil {
		return nil, err
	}
	if err = c.decodeWhere(w.Where, depth+1); err != nil {
		return nil, err
	}
	return c, nil
}

func clonePtr[T any](p *T) *T {
	if p == nil {
		return nil
	}
	v := *p
	return &v
}
