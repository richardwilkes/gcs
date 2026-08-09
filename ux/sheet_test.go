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

// TestMarkModifiedRecalculates verifies that telling the sheet something changed brings the entity's derived state up
// to date. Everything MarkModified goes on to do -- the panels, the tables and the calculator -- displays that state.
// It used to be refreshed only as a side effect of the tab asking whether the sheet had unsaved changes, which
// recalculated the entity on its way to hashing it; hashing no longer does that, since recalculating rewrites part of
// what gets saved.
func TestMarkModifiedRecalculates(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	entity.Recalculate()
	before := skill.LevelData.Level

	entity.Attributes.Set["dx"].Adjustment = fxp.Four
	sheet.MarkModified(nil)

	c.Equal(before+fxp.Four, skill.LevelData.Level, "the skill level must reflect the raised attribute")
}
