/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package criteria

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/json"
)

// Numeric holds the criteria for matching a number.
type Numeric struct {
	NumericData
}

// NumericData holds the criteria for matching a number that should be written to disk.
type NumericData struct {
	Compare   NumericCompareType `json:"compare,omitempty"`
	Qualifier fxp.Int            `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (n Numeric) ShouldOmit() bool {
	return n.Compare.EnsureValid() == AnyNumber
}

// UnmarshalJSON implements json.Unmarshaler.
func (n *Numeric) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &n.NumericData)
	n.Compare = n.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (n Numeric) Matches(value fxp.Int) bool {
	return n.Compare.Matches(n.Qualifier, value)
}

func (n Numeric) String() string {
	return n.Compare.Describe(n.Qualifier)
}

// AltString returns the alternate description.
func (n Numeric) AltString() string {
	return n.Compare.AltDescribe(n.Qualifier)
}
