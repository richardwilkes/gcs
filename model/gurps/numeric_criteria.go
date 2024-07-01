// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

// NumericCriteria holds the criteria for matching a number.
type NumericCriteria struct {
	NumericCriteriaData
}

// NumericCriteriaData holds the criteria for matching a number that should be written to disk.
type NumericCriteriaData struct {
	Compare   NumericCompareType `json:"compare,omitempty"`
	Qualifier fxp.Int            `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (n NumericCriteria) ShouldOmit() bool {
	return n.Compare.EnsureValid() == AnyNumber
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *NumericCriteria) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &n.NumericCriteriaData)
	n.Compare = n.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (n NumericCriteria) Matches(value fxp.Int) bool {
	return n.Compare.Matches(n.Qualifier, value)
}

func (n NumericCriteria) String() string {
	return n.Compare.Describe(n.Qualifier)
}

// AltString returns the alternate description.
func (n NumericCriteria) AltString() string {
	return n.Compare.AltDescribe(n.Qualifier)
}

// Hash writes this object's contents into the hasher.
func (n NumericCriteria) Hash(h hash.Hash) {
	if n.ShouldOmit() {
		return
	}
	_, _ = h.Write([]byte(n.Compare))
	_ = binary.Write(h, binary.LittleEndian, n.Qualifier)
}
