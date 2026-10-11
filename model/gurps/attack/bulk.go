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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Bulk holds the bulk data for a weapon.
type Bulk struct {
	Normal          fxp.Int
	Giant           fxp.Int
	RetractingStock bool
}

// ParseBulk parses a string into a Bulk.
func ParseBulk(s string) Bulk {
	var wb Bulk
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	wb.RetractingStock = strings.Contains(s, "*")
	parts := strings.Split(s, "/")
	wb.Normal, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wb.Giant, _ = fxp.Extract(parts[1])
	}
	wb.Validate()
	return wb
}

// IsZero reports whether json's omitzero option should omit this value.
func (wb Bulk) IsZero() bool {
	return wb == Bulk{}
}

// MarshalJSONTo implements json.MarshalerTo.
func (wb Bulk) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wb.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wb *Bulk) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wb, ParseBulk)
}

// Hash writes this object's contents into the hasher.
func (wb Bulk) Hash(h hash.Hash) {
	xhash.Num64(h, wb.Normal)
	xhash.Num64(h, wb.Giant)
	xhash.Bool(h, wb.RetractingStock)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedBulk method first if you want the resolved values.
func (wb Bulk) String() string {
	if wb.Normal >= 0 && wb.Giant >= 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wb.Normal.String())
	if wb.Giant != 0 && wb.Giant != wb.Normal {
		buffer.WriteByte('/')
		buffer.WriteString(wb.Giant.String())
	}
	if wb.RetractingStock {
		buffer.WriteByte('*')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedBulk method first if you want the tooltip to
// be based on the resolved values. accuracy, recoil and strength are the weapon's resolved values, which the
// folded-stock figures are derived from.
func (wb Bulk) Tooltip(accuracy Accuracy, recoil Recoil, strength Strength) string {
	if !wb.RetractingStock {
		return ""
	}
	if wb.Normal < 0 {
		wb.Normal += fxp.One
	}
	if wb.Giant < 0 {
		wb.Giant += fxp.One
	}
	wb.Validate()
	accuracy.Base -= fxp.One
	accuracy.Validate()
	if recoil.Shot > fxp.One {
		recoil.Shot += fxp.One
	}
	if recoil.Slug > fxp.One {
		recoil.Slug += fxp.One
	}
	recoil.Validate()
	minST := strength
	minST.Min = minST.Min.Mul(fxp.OnePointTwo).Ceil()
	minST.Validate()
	return i18n.Text("*: Has a retracting stock. With the stock folded, the weapon's stats change to Bulk %s, Accuracy %s, Recoil %s, and minimum ST %s. Folding or unfolding the stock takes one Ready maneuver.",
		wb.String(), accuracy.String(), recoil.String(), minST.String())
}

// Validate ensures that the data is valid.
func (wb *Bulk) Validate() {
	wb.Normal = wb.Normal.Min(0)
	wb.Giant = wb.Giant.Min(0)
}
