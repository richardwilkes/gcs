// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package ux

import (
	"slices"
	"testing"
	"time"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
)

// newEntityWithAscendingPointsRecord returns an entity whose points record list is stored oldest first, which is the
// opposite of the order the editor uses. Files not written by GCS may hold the list in any order, since Entity's
// unmarshaling only sorts it when the point-total reconciliation branch runs.
func newEntityWithAscendingPointsRecord() *gurps.Entity {
	base := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.Local)
	entity := gurps.NewEntity()
	entity.PointsRecord = []*gurps.PointsRecord{
		{When: jio.Time(base), Points: fxp.FromInteger(100), Reason: "Starting points"},
		{When: jio.Time(base.AddDate(0, 1, 0)), Points: fxp.FromInteger(5), Reason: "First session"},
		{When: jio.Time(base.AddDate(0, 2, 0)), Points: fxp.FromInteger(3), Reason: "Second session"},
	}
	return entity
}

// TestPointsEditorOpensUnmodifiedForAnUnsortedRecordList verifies that an editor opened on a points record list that
// isn't already in the editor's display order reports no changes. The editor sorts the copy it edits, so if the copy it
// compares against isn't sorted the same way, an untouched editor enables Apply and Cancel, shows the modified marker
// on its tab, and prompts to save on the way out.
func TestPointsEditorOpensUnmodifiedForAnUnsortedRecordList(t *testing.T) {
	c := check.New(t)
	entity := newEntityWithAscendingPointsRecord()
	e := newPointsEditor(nil, entity)
	c.False(e.isModified(), "an editor that has not been touched must report no changes")

	for i, rec := range e.current {
		if i != 0 {
			c.True(e.current[i-1].When.Compare(rec.When) >= 0, "the editor must show the records most recent first")
		}
	}

	e.current[0].Points += fxp.One
	c.True(e.isModified(), "editing a record must be reported as a change")

	e.current[0].Points -= fxp.One
	c.False(e.isModified(), "putting the original value back must clear the change")
}

// TestPointsEditorSortsBothCopiesTheSameWay verifies that the two copies the editor holds are independent of each
// other, but hold equal values in the same order, no matter what order the entity stored them in.
func TestPointsEditorSortsBothCopiesTheSameWay(t *testing.T) {
	c := check.New(t)
	entity := newEntityWithAscendingPointsRecord()
	e := newPointsEditor(nil, entity)
	c.Equal(len(entity.PointsRecord), len(e.before), "the editor must hold every record")
	c.Equal(len(e.before), len(e.current), "both copies must hold the same number of records")
	for i := range e.before {
		c.True(e.before[i] != e.current[i], "the copies must not share records")
		c.Equal(*e.before[i], *e.current[i], "the copies must hold equal records in the same order")
		c.True(!slices.Contains(entity.PointsRecord, e.current[i]),
			"the editor must not share records with the entity")
	}
	c.Equal("Second session", e.current[0].Reason, "the most recent record must come first")
}
