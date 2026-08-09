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
	"slices"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
)

// PointsRecord holds information about when and why points were adjusted.
type PointsRecord struct {
	When   jio.Time `json:"when"`
	Points fxp.Int  `json:"points"`
	Reason string   `json:"reason,omitzero"`
}

// ClonePointsRecordList creates a clone of the provided PointsRecord list.
func ClonePointsRecordList(list []*PointsRecord) []*PointsRecord {
	clone := make([]*PointsRecord, len(list))
	for i := range list {
		record := *list[i]
		clone[i] = &record
	}
	return clone
}

// SortPointsRecordList sorts the provided PointsRecord list into the order GCS presents and stores them in, which is
// most recent first.
func SortPointsRecordList(list []*PointsRecord) {
	slices.SortFunc(list, func(a, b *PointsRecord) int { return b.When.Compare(a.When) })
}
