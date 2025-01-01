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
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

// Weight holds the criteria for matching a weight.
type Weight struct {
	WeightData
}

// WeightData holds the criteria for matching a weight that should be written to disk.
type WeightData struct {
	Compare   NumericComparison `json:"compare,omitempty"`
	Qualifier fxp.Weight        `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (w Weight) ShouldOmit() bool {
	return w.Compare.EnsureValid() == AnyNumber
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *Weight) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &w.WeightData)
	w.Compare = w.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (w Weight) Matches(value fxp.Weight) bool {
	return w.Compare.Matches(fxp.Int(w.Qualifier), fxp.Int(value))
}

func (w Weight) String() string {
	return w.Compare.Describe(fxp.Int(w.Qualifier))
}

// Hash writes this object's contents into the hasher.
func (w Weight) Hash(h hash.Hash) {
	if w.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.String(h, w.Compare)
	hashhelper.Num64(h, w.Qualifier)
}
