// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package attack

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Block holds the block data for a weapon.
type Block struct {
	CanBlock bool
	Modifier fxp.Int
}

// ParseBlock parses a string into a Block.
func ParseBlock(s string) Block {
	var wb Block
	s = strings.TrimSpace(s)
	s = strings.ToLower(s)
	if s != "" && s != "-" && s != "–" && s != "—" && !strings.Contains(s, "no") {
		wb.CanBlock = true
		wb.Modifier, _ = fxp.Extract(s)
	}
	wb.Validate()
	return wb
}

// IsZero reports whether json's omitzero option should omit this value.
func (wb Block) IsZero() bool {
	return !wb.CanBlock
}

// MarshalJSONTo implements json.MarshalerTo.
func (wb Block) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wb.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wb *Block) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wb, ParseBlock)
}

// Hash writes this object's contents into the hasher.
func (wb Block) Hash(h hash.Hash) {
	xhash.Bool(h, wb.CanBlock)
	xhash.Num64(h, wb.Modifier)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedBlock method first if you want the resolved values.
func (wb Block) String() string {
	if !wb.CanBlock {
		return "No"
	}
	return wb.Modifier.String()
}

// Validate ensures that the data is valid.
func (wb *Block) Validate() {
	if !wb.CanBlock {
		wb.Modifier = 0
	}
}
