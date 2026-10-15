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
	"strconv"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// pointsRecordsAt returns a points record for each offset, in order, made offset seconds after a fixed time, holding one
// point and giving its index as its reason. The offsets are an input that an unstable sort is known to reorder: it puts
// the sixth record ahead of the fourth, though they share a time.
func pointsRecordsAt(offsets ...int) []*PointsRecord {
	base := time.Date(2026, time.October, 13, 10, 0, 0, 0, time.UTC)
	list := make([]*PointsRecord, len(offsets))
	for i, offset := range offsets {
		list[i] = &PointsRecord{
			When:   jio.Time(base.Add(time.Duration(offset) * time.Second)),
			Points: fxp.One,
			Reason: strconv.Itoa(i),
		}
	}
	return list
}

var unstableSortOffsets = []int{10, 33, 20, 35, 5, 35, 22, 7, 6, 15, 35, 17, 9, 32}

// pointsRecordReasons returns the reasons of the records, in order.
func pointsRecordReasons(list []*PointsRecord) []string {
	reasons := make([]string, len(list))
	for i, rec := range list {
		reasons[i] = rec.Reason
	}
	return reasons
}

// TestSortPointsRecordListKeepsTiedRecordsInOrder checks that records made in the same second, which is as finely as
// their times are kept, keep the order they had when sorted, as the shopping record's do, rather than swapping about.
func TestSortPointsRecordListKeepsTiedRecordsInOrder(t *testing.T) {
	c := check.New(t)
	list := pointsRecordsAt(unstableSortOffsets...)
	SortPointsRecordList(list)
	c.True(slices.IsSortedFunc(list, func(a, b *PointsRecord) int { return b.When.Compare(a.When) }))
	c.Equal([]string{"3", "5", "10"}, pointsRecordReasons(list[:3]),
		"the records sharing the latest time come first, in the order they had")
}

// TestReconciliationKeepsTiedPointsRecordsInOrder checks that the record added when the points records don't add up to
// the total, on loading a sheet, is sorted in without the records sharing a time swapping about.
func TestReconciliationKeepsTiedPointsRecordsInOrder(t *testing.T) {
	c := check.New(t)
	e := NewEntity()
	e.PointsRecord = pointsRecordsAt(unstableSortOffsets...)
	e.TotalPoints = fxp.FromInteger(len(e.PointsRecord) + 5)
	data, err := jio.Marshal(e)
	c.NoError(err)
	var loaded Entity
	c.NoError(jio.Unmarshal(data, &loaded))
	c.Equal(len(unstableSortOffsets)+1, len(loaded.PointsRecord), "a reconciliation record is added")
	c.Equal("Reconciliation", loaded.PointsRecord[0].Reason, "and comes first, being the newest")
	c.True(slices.IsSortedFunc(loaded.PointsRecord, func(a, b *PointsRecord) int { return b.When.Compare(a.When) }))
	c.Equal([]string{"3", "5", "10"}, pointsRecordReasons(loaded.PointsRecord[1:4]),
		"the records sharing the latest time keep the order they had")
}
