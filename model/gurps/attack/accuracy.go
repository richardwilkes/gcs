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

// Accuracy holds the accuracy data for a weapon.
type Accuracy struct {
	Base  fxp.Int
	Scope fxp.Int
	Jet   bool
}

// ParseAccuracy parses a string into a Accuracy.
func ParseAccuracy(s string) Accuracy {
	var wa Accuracy
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ToLower(s)
	if strings.Contains(s, "jet") {
		wa.Jet = true
	} else {
		s = strings.TrimPrefix(s, "+")
		parts := strings.Split(s, "+")
		wa.Base, _ = fxp.Extract(parts[0])
		if len(parts) > 1 {
			wa.Scope, _ = fxp.Extract(parts[1])
		}
	}
	wa.Validate()
	return wa
}

// IsZero reports whether json's omitzero option should omit this value.
func (wa Accuracy) IsZero() bool {
	return wa == Accuracy{}
}

// MarshalJSONTo implements json.MarshalerTo.
func (wa Accuracy) MarshalJSONTo(enc *jsontext.Encoder) error {
	return json.MarshalEncode(enc, wa.String())
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (wa *Accuracy) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	return jio.UnmarshalStringFromInfallible(dec, wa, ParseAccuracy)
}

// Hash writes this object's contents into the hasher.
func (wa Accuracy) Hash(h hash.Hash) {
	xhash.Num64(h, wa.Base)
	xhash.Num64(h, wa.Scope)
	xhash.Bool(h, wa.Jet)
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Use the weapon's ResolvedAccuracy method first if you want the resolved values.
func (wa Accuracy) String() string {
	if wa.Jet {
		return "Jet" // Not localized, since it is part of the data
	}
	if wa.Scope != 0 {
		return wa.Base.String() + wa.Scope.StringWithSign()
	}
	return wa.Base.String()
}

// Validate ensures that the data is valid.
func (wa *Accuracy) Validate() {
	if wa.Jet {
		wa.Base = 0
		wa.Scope = 0
		return
	}
	wa.Base = wa.Base.Max(0)
	wa.Scope = wa.Scope.Max(0)
}
