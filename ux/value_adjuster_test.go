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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
)

// TestOwnerRecalculatesSkipsSheetAlreadyUpdating verifies that an edit only leaves the recalculation to the owner when
// the owner is actually going to perform one. A sheet that is already inside an update pass is not, since
// Sheet.MarkModified does nothing at all while awaitingUpdate is set.
func TestOwnerRecalculatesSkipsSheetAlreadyUpdating(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	c.True(ownerRecalculates(sheet, entity), "a sheet recalculates its own entity when marked as modified")
	sheet.awaitingUpdate = true
	c.False(ownerRecalculates(sheet, entity), "a sheet already inside an update pass won't recalculate again")
	c.False(ownerRecalculates(sheet.Traits, entity), "a panel within that sheet resolves to the same answer")
	sheet.awaitingUpdate = false
	c.True(ownerRecalculates(sheet, entity), "the sheet recalculates again once the update pass is done")
}

// TestValueAdjustmentRecalculatesDuringUpdatePass verifies that an adjustment made while the sheet is already inside an
// update pass still brings the derived state up to date. Sheet.MarkModified is a no-op in that case, so an adjustment
// that left the recalculation to it would silently drop it, leaving the sheet showing stale levels and points until
// some unrelated later edit happened to trigger a pass of its own.
func TestValueAdjustmentRecalculatesDuringUpdatePass(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	entity.Recalculate()
	before := skill.LevelData.Level

	sheet.awaitingUpdate = true
	t.Cleanup(func() { sheet.awaitingUpdate = false })
	adjustment := &snapshotList[*gurps.Attribute, fxp.Int]{
		owner:  sheet,
		entity: entity,
		set:    func(attr *gurps.Attribute, value fxp.Int) { attr.Adjustment = value },
		list: []valueSnapshot[*gurps.Attribute, fxp.Int]{
			{target: entity.Attributes.Set["dx"], value: fxp.Four},
		},
	}
	adjustment.apply()

	c.Equal(before+fxp.Four, skill.LevelData.Level, "the skill level must reflect the raised attribute")
}
