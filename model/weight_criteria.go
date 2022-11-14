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

package model

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/measure"
	"github.com/richardwilkes/json"
)

// WeightCriteria holds the criteria for matching a number.
type WeightCriteria struct {
	WeightCriteriaData
}

// WeightCriteriaData holds the criteria for matching a number that should be written to disk.
type WeightCriteriaData struct {
	Compare   NumericCompareType `json:"compare,omitempty"`
	Qualifier measure.Weight     `json:"qualifier,omitempty"`
}

// ShouldOmit implements json.Omitter.
func (w WeightCriteria) ShouldOmit() bool {
	return w.Compare.EnsureValid() == AnyNumber
}

// UnmarshalJSON implements json.Unmarshaler.
func (w *WeightCriteria) UnmarshalJSON(data []byte) error {
	err := json.Unmarshal(data, &w.WeightCriteriaData)
	w.Compare = w.Compare.EnsureValid()
	return err
}

// Matches performs a comparison and returns true if the data matches.
func (w WeightCriteria) Matches(value measure.Weight) bool {
	return w.Compare.Matches(fxp.Int(w.Qualifier), fxp.Int(value))
}

func (w WeightCriteria) String() string {
	return w.Compare.Describe(fxp.Int(w.Qualifier))
}
