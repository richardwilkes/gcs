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

// Recoil holds the recoil data for a weapon.
type Recoil struct {
	Shot fxp.Int
	Slug fxp.Int
}

// ParseRecoil parses a string into a Recoil.
func ParseRecoil(s string) Recoil {
	var wr Recoil
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	parts := strings.Split(s, "/")
	wr.Shot, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wr.Slug, _ = fxp.Extract(parts[1])
	}
	return wr
}

// MarshalJSONTo implements json.MarshalerTo.
func (wr Recoil) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wr.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wr *Recoil) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wr, ParseRecoil)
}

// IsZero reports whether json's omitzero option should omit this value.
func (wr Recoil) IsZero() bool {
	return wr == Recoil{}
}

// Hash writes this object's contents into the hasher.
func (wr Recoil) Hash(h hash.Hash) {
	xhash.Num64(h, wr.Shot)
	xhash.Num64(h, wr.Slug)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedRecoil method first if you want the resolved values.
func (wr Recoil) String() string {
	if wr.Shot == 0 && wr.Slug == 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wr.Shot.String())
	if wr.Slug != 0 {
		buffer.WriteByte('/')
		buffer.WriteString(wr.Slug.String())
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Use the weapon's ResolvedRecoil method first if you want the tooltip
// to be based on the resolved values.
func (wr Recoil) Tooltip() string {
	if wr.Shot != 0 && wr.Slug != 0 && wr.Shot != wr.Slug {
		return i18n.Text("First Recoil value is for shot, second is for slugs")
	}
	return ""
}

// Validate ensures that the data is valid.
func (wr *Recoil) Validate() {
	wr.Shot = wr.Shot.Max(0)
	wr.Slug = wr.Slug.Max(0)
}
