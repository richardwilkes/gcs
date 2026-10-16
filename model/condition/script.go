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

	"github.com/richardwilkes/gcs/v5/model/nameable"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Node = &Script{}

// Script is evaluated by the host. It has no Not. Wire type "script".
type Script struct {
	Name   string `json:"name,omitzero"`
	Script string `json:"script"`
	Lang   string `json:"lang,omitzero"` // "" is js, the only language there is
}

// scriptFields is the method-less shadow Script is written through.
type scriptFields Script

// scriptWire is the raw form Script is read through.
type scriptWire struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Script string `json:"script"`
	Lang   string `json:"lang"`
}

// Clone implements Node.
func (s *Script) Clone() Node {
	other := *s
	return &other
}

// Hash implements Node.
func (s *Script) Hash(h hash.Hash) {
	xhash.Num8(h, hashScript)
	xhash.StringWithLen(h, s.Name)
	xhash.StringWithLen(h, s.Script)
	xhash.StringWithLen(h, s.Lang)
}

// FillWithNameableKeys implements Node.
func (s *Script) FillWithNameableKeys(m, existing map[string]string) {
	nameable.Extract(m, existing, s.Name, s.Script)
}

func (s *Script) validate(v *validator) {
	if s.Lang != "" {
		v.report(i18n.Text("unknown script language"))
	}
}

func (s *Script) walk(_ *walker) {}

// MarshalJSONTo implements json.MarshalerTo.
func (s *Script) MarshalJSONTo(enc *jsontext.Encoder) error {
	return encodeTyped(enc, typeScript, (*scriptFields)(s))
}

func decodeScript(raw jsontext.Value) (Node, error) {
	var w scriptWire
	if err := unmarshalWire(raw, &w); err != nil {
		return nil, err
	}
	return &Script{Name: w.Name, Script: w.Script, Lang: w.Lang}, nil
}

// eval implements Node. The host runs the script, at most once per (node, self, subject) in one evaluation. A script
// that fails is Failed.
func (s *Script) eval(e *evaluator, f frame) Result {
	out := e.runScript(s, f)
	return Result{State: out.state, Reason: out.reason}
}
