// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/json/jsontext"
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
)

// PointsRecord holds information about when and why points were adjusted.
type PointsRecord struct {
	When   jio.Time `json:"when"`
	Points fxp.Int  `json:"points"`
	Reason string   `json:"reason,omitzero"`
	// Unknown holds the members this release does not know, so that they survive a load/save cycle unchanged.
	Unknown jsontext.Value `json:",embed"`
}

// ClonePointsRecordList creates a clone of the provided PointsRecord list.
func ClonePointsRecordList(list []*PointsRecord) []*PointsRecord {
	clone := make([]*PointsRecord, len(list))
	for i, one := range list {
		if clone[i] = clonePtr(one); clone[i] != nil {
			clone[i].Unknown = slices.Clone(one.Unknown)
		}
	}
	return clone
}

// SortPointsRecordList sorts the provided PointsRecord list into the order GCS presents and stores them in, which is
// most recent first. Times are kept only to the second, so records made in the same second keep the order they had.
func SortPointsRecordList(list []*PointsRecord) {
	slices.SortStableFunc(list, func(a, b *PointsRecord) int { return b.When.Compare(a.When) })
}
