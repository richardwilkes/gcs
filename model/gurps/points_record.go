// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
)

// PointsRecord holds information about when and why points were adjusted.
type PointsRecord struct {
	When   jio.Time `json:"when"`
	Points fxp.Int  `json:"points"`
	Reason string   `json:"reason,omitempty"`
}

// ClonePointsRecordList creates a clone of the provided PointsRecord list.
func ClonePointsRecordList(list []*PointsRecord) []*PointsRecord {
	clone := make([]*PointsRecord, len(list))
	for i := 0; i < len(list); i++ {
		record := *list[i]
		clone[i] = &record
	}
	return clone
}
