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

// The points block shows nine parts of the point breakdown, which walks everything on the sheet, so the breakdown is
// derived once per update rather than once per field, and every field must still follow a change.
func TestPointsPanelFollowsChanges(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	panel, ok := sheet.blocks[gurps.BlockPointsKey].(*PointsPanel)
	c.True(ok, "the sheet must have a points block")
	first := panel.pointsBreakdown()
	c.True(first == panel.pointsBreakdown(), "the breakdown is derived once per update")
	c.Equal(fxp.Int(0), first.Skills)

	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	sheet.MarkModified(sheet)
	c.NotEqual(first, panel.pointsBreakdown(), "an update derives the breakdown afresh")
	c.Equal(fxp.Four, panel.pointsBreakdown().Skills)
	c.Equal(entity.UnspentPoints(), panel.unspentPoints())
	found := false
	for _, child := range panel.ptsList.Children() {
		if field, isField := child.Self.(*NonEditablePageField); isField && field.Text.String() == "4" {
			found = true
		}
	}
	c.True(found, "the skills field must show the points spent on the skill")
}
