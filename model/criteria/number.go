// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package criteria

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Number holds the criteria for matching a number.
type Number struct {
	NumberData
}

// NumberData holds the criteria for matching a number that should be written to disk.
type NumberData struct {
	Compare   NumericComparison `json:"compare,omitempty"`
	Qualifier fxp.Int           `json:"qualifier,omitzero"`
}

// IsZero implements json.isZero.
func (n Number) IsZero() bool {
	return n.Compare.EnsureValid() == AnyNumber
}

// UnmarshalJSONFrom implements json.UnmarshalerFrom.
func (n *Number) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	err := json.UnmarshalDecode(dec, &n.NumberData)
	n.Compare = n.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (n Number) Matches(value fxp.Int) bool {
	return n.Compare.Matches(n.Qualifier, value)
}

func (n Number) String() string {
	return n.Compare.Describe(n.Qualifier)
}

// AltString returns the alternate description.
func (n Number) AltString() string {
	return n.Compare.AltDescribe(n.Qualifier)
}

// Hash writes this object's contents into the hasher.
func (n Number) Hash(h hash.Hash) {
	xhash.StringWithLen(h, n.Compare)
	xhash.Num64(h, n.Qualifier)
}
