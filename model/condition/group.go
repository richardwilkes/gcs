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

	"github.com/richardwilkes/gcs/v5/model/criteria"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Node = &Group{}

// Group combines its children: all of them, any of them, or at least N of them must be met. An exclusive group is an
// all group of counts over one collection that share its members, each member counting toward one of them at most.
// Wire type "group".
type Group struct {
	Mode      GroupMode `json:"mode"`
	N         int       `json:"n,omitzero"` // at_least only
	Not       bool      `json:"not,omitzero"`
	Exclusive bool      `json:"exclusive,omitzero"`
	OnlyWhen  Ref       `json:"only_when,omitzero"`
	Children  Nodes     `json:"children,omitempty"`
}

// groupFields is the method-less shadow Group is written through.
type groupFields Group

// groupWire is the raw form Group is read through.
type groupWire struct {
	Type      string           `json:"type"`
	Mode      *string          `json:"mode"`
	N         int              `json:"n"`
	Not       bool             `json:"not"`
	Exclusive bool             `json:"exclusive"`
	OnlyWhen  jsontext.Value   `json:"only_when"`
	Children  []jsontext.Value `json:"children"`
}

// Clone implements Node.
func (g *Group) Clone() Node {
	c := *g
	c.OnlyWhen = g.OnlyWhen.clone()
	c.Children = g.Children.clone()
	return &c
}

// Hash implements Node.
func (g *Group) Hash(h hash.Hash) {
	xhash.Num8(h, hashGroup)
	xhash.StringWithLen(h, g.Mode.Key())
	xhash.Num64(h, g.N)
	xhash.Bool(h, g.Not)
	xhash.Bool(h, g.Exclusive)
	g.OnlyWhen.hash(h)
	g.Children.hash(h)
}

// FillWithNameableKeys implements Node.
func (g *Group) FillWithNameableKeys(m, existing map[string]string) {
	g.OnlyWhen.fillWithNameableKeys(m, existing)
	g.Children.fillWithNameableKeys(m, existing)
}

func (g *Group) validate(v *validator) {
	switch {
	case g.Mode != g.Mode.EnsureValid():
		v.report(i18n.Text("unknown group mode"))
	case g.Mode == GroupAtLeast:
		if g.N < 1 || g.N > len(g.Children) {
			v.report(i18n.Text("the number of children needed is outside 1 to the number of children"))
		}
	case g.N != 0:
		v.report(i18n.Text(`a number of children needed is set, but the mode is not "at least"`))
	}
	if g.Exclusive {
		g.validateExclusive(v)
	}
	v.ref(SlotOnlyWhen, g.OnlyWhen)
	for i, child := range g.Children {
		v.child(Slot{Kind: SlotChild, Index: i}, child)
	}
}

// validateExclusive reports what keeps an exclusive group from being a pool: it must be an all group whose children
// are all counts over the same collection, none negated, grouped or exclusive itself, each needing at least a whole
// number of members.
func (g *Group) validateExclusive(v *validator) {
	if g.Mode != GroupAll {
		v.report(i18n.Text(`an exclusive group's mode is not "all"`))
	}
	var of *string
	for _, child := range g.Children {
		c, ok := child.(*Count)
		switch {
		case !ok:
			if child != nil {
				v.report(i18n.Text("an exclusive group holds something other than a count"))
				return
			}
			continue
		case c.Not || c.GroupBy != "" || c.Exclusive:
			v.report(i18n.Text("an exclusive group holds a count that is negated or grouped"))
			return
		case !atLeastWhole(c.Quantity):
			v.report(i18n.Text(`an exclusive group holds a count whose quantity is not "at least" a whole number`))
			return
		case of != nil && *of != c.Of:
			v.report(i18n.Text("an exclusive group holds counts over different collections"))
			return
		}
		of = &c.Of
	}
}

// atLeastWhole returns true for a quantity that needs some whole number of members: nil, which means at least 1, or
// at_least a whole number of at least 1.
func atLeastWhole(q *criteria.Number) bool {
	return q == nil || (q.Compare == criteria.AtLeastNumber && q.Qualifier >= fxp.One && q.Qualifier == q.Qualifier.Trunc())
}

func (g *Group) walk(w *walker) {
	w.ref(SlotOnlyWhen, g.OnlyWhen)
	for i, child := range g.Children {
		w.child(Slot{Kind: SlotChild, Index: i}, child)
	}
}

// MarshalJSONTo implements json.MarshalerTo.
func (g *Group) MarshalJSONTo(enc *jsontext.Encoder) error {
	if g.Mode != g.Mode.EnsureValid() {
		return errors.New(i18n.Text("unknown group mode"))
	}
	return encodeTyped(enc, typeGroup, (*groupFields)(g))
}

func decodeGroup(raw jsontext.Value, depth int) (Node, error) {
	var w groupWire
	if err := unmarshalWire(raw, &w); err != nil {
		return nil, err
	}
	mode, err := requiredKey(w.Mode, "mode", ExtractKnownGroupMode)
	if err != nil {
		return nil, err
	}
	g := &Group{Mode: mode, N: w.N, Not: w.Not, Exclusive: w.Exclusive}
	if g.OnlyWhen.Node, err = decodeRef(w.OnlyWhen, depth+1); err != nil {
		return nil, at(Slot{Kind: SlotOnlyWhen}, err)
	}
	if g.Children, err = decodeNodes(w.Children, depth+1); err != nil {
		return nil, err
	}
	return g, nil
}
