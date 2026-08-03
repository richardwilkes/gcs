// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"fmt"
	"hash"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/gurps/enums/prereq"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xbytes"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

var _ Prereq = &UnknownPrereq{}

// UnknownPrereq holds a prerequisite whose type this version of GCS doesn't understand, most likely because it was
// written by a newer version. The original JSON is retained verbatim and written back out unchanged when the file is
// saved, so that merely loading and saving a file cannot destroy data. Unlike an unknown feature, which can safely be
// ignored, an unknown prerequisite is reported as unsatisfied: treating a constraint that can't be evaluated as met
// would silently permit something the data says isn't permitted.
type UnknownPrereq struct {
	// Parent is the owning list.
	Parent *PrereqList
	// Kind is the unrecognized value that was found in the "type" field.
	Kind string
	// Data is the original JSON for the prerequisite, exactly as it was read.
	Data jsontext.Value
}

// NewUnknownPrereq creates a new UnknownPrereq holding a copy of the passed-in data.
func NewUnknownPrereq(kind string, data jsontext.Value) *UnknownPrereq {
	return &UnknownPrereq{
		Kind: kind,
		Data: slices.Clone(data),
	}
}

// PrereqType implements Prereq.
func (p *UnknownPrereq) PrereqType() prereq.Type {
	return prereq.Unknown
}

// ParentList implements Prereq.
func (p *UnknownPrereq) ParentList() *PrereqList {
	return p.Parent
}

// Clone implements Prereq.
func (p *UnknownPrereq) Clone(parent *PrereqList) Prereq {
	clone := NewUnknownPrereq(p.Kind, p.Data)
	clone.Parent = parent
	return clone
}

// FillWithNameableKeys implements Prereq.
func (p *UnknownPrereq) FillWithNameableKeys(_, _ map[string]string) {
}

// Satisfied implements Prereq. An unknown prerequisite is never satisfied.
func (p *UnknownPrereq) Satisfied(_ *Entity, _ any, tooltip *xbytes.InsertBuffer, prefix string, _ *bool) bool {
	if tooltip != nil {
		tooltip.WriteString(prefix)
		fmt.Fprintf(tooltip, i18n.Text("an unknown prerequisite type (%q) that requires a newer version of GCS to evaluate"),
			p.Kind)
	}
	return false
}

// MarshalJSONTo implements json.MarshalerTo. The original data is written back out as-is.
func (p *UnknownPrereq) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(p.Data)
}

// Hash implements Prereq.
func (p *UnknownPrereq) Hash(h hash.Hash) {
	if p == nil {
		xhash.Num8(h, uint8(255))
		return
	}
	xhash.Num8(h, prereq.Unknown)
	xhash.StringWithLen(h, p.Kind)
	xhash.BytesWithLen(h, p.Data)
}
