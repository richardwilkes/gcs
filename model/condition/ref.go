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
	"hash"
	"slices"

	"github.com/richardwilkes/toolbox/v2/xhash"
)

// MaxDepth is the deepest a node may sit, counting the root as 1 and every step through children, where, only_when,
// operands and sums. Decode rejects a tree nested deeper; Validate reports one built in code.
const MaxDepth = 32

// MarshalJSONTo implements json.MarshalerTo. A zero Ref is written as null, and a broken one as the bytes it was read
// from.
func (r Ref) MarshalJSONTo(enc *jsontext.Encoder) error {
	if r.IsZero() {
		return enc.WriteToken(jsontext.Null)
	}
	return json.MarshalEncode(enc, r.Node)
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom. A tree that can't be decoded doesn't fail the file holding it:
// the Ref is broken instead, keeping the bytes as read and the error, which Err returns.
func (r *Ref) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	raw, err := dec.ReadValue()
	if err != nil {
		return err
	}
	var n Node
	if n, err = decodeRef(raw, 1); err != nil {
		n = &broken{data: slices.Clone(raw), err: err}
	}
	r.Node = n
	return nil
}

// Err returns the error that kept the Ref's tree from being decoded, or nil when it is not broken. A broken Ref
// evaluates as Failed with the error as the reason, and is written back unchanged.
func (r Ref) Err() error {
	if b, ok := r.Node.(*broken); ok {
		return b.err
	}
	return nil
}

// broken stands in for the root of a tree that a Ref could not decode. It is never part of a tree.
type broken struct {
	err  error
	data jsontext.Value
}

func (b *broken) Clone() Node {
	return &broken{data: slices.Clone(b.data), err: b.err}
}

func (b *broken) Hash(h hash.Hash) {
	xhash.Num8(h, hashBroken)
	xhash.BytesWithLen(h, b.data)
}

func (b *broken) FillWithNameableKeys(_, _ map[string]string) {}

func (b *broken) validate(v *validator) {
	var p Problem
	errors.As(b.err, &p)
	v.reportAt(p.Path, p.Msg)
}

func (b *broken) walk(_ *walker) {}

func (b *broken) eval(_ *evaluator, _ frame) Result {
	return failure(b.err.Error())
}

// MarshalJSONTo implements json.MarshalerTo, writing the bytes as read.
func (b *broken) MarshalJSONTo(enc *jsontext.Encoder) error {
	return enc.WriteValue(b.data)
}
