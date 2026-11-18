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
	"github.com/richardwilkes/unison"
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

// pointsPanelWithSkill returns the points block of a sheet whose character has spent 4 points on a skill.
func pointsPanelWithSkill(t *testing.T) (*Sheet, *PointsPanel) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	panel, ok := sheet.blocks[gurps.BlockPointsKey].(*PointsPanel)
	if !ok {
		t.Fatal("the sheet must have a points block")
	}
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Brawling"
	skill.Points = fxp.Four
	entity.Skills = append(entity.Skills, skill)
	sheet.MarkModified(sheet)
	return sheet, panel
}

// The unspent points are called overspent, in the error colors, while more points are spent than the character has,
// and unspent again once they are not, none left over included.
func TestPointsPanelSaysWhenOverspent(t *testing.T) {
	c := check.New(t)
	sheet, panel := pointsPanelWithSkill(t)
	entity := sheet.Entity()
	spent := entity.PointsBreakdown().Total()
	for _, one := range []struct {
		left fxp.Int
		want string
	}{
		{left: fxp.One, want: "Unspent"},
		{left: 0, want: "Unspent"},
		{left: -fxp.One, want: "Overspent"},
		{left: 0, want: "Unspent"},
	} {
		entity.TotalPoints = spent + one.left
		sheet.MarkModified(sheet)
		c.Equal(one.left.String(), panel.unspentField.Text.String(), "%v left: the points left", one.left)
		c.Equal(one.want, panel.unspentLabel.Text.String(), "%v left: the label", one.left)
		c.Equal(one.want == "Overspent", panel.unspentField.OnBackgroundInk == unison.ThemeOnError,
			"%v left: the error color", one.left)
	}
}

// The header gives the character's total points or, once the sheet is set to leave the unspent points out of it, the
// points spent, whether the setting is turned on while the block is shown or was on when it was built.
func TestPointsPanelTotalCanLeaveOutTheUnspentPoints(t *testing.T) {
	c := check.New(t)
	sheet, panel := pointsPanelWithSkill(t)
	entity := sheet.Entity()
	spent := entity.PointsBreakdown().Total()
	c.NotEqual(entity.TotalPoints, spent, "precondition: some points are left unspent")
	c.Equal(entity.TotalPoints.String()+" Points", panel.total.Text.String())
	entity.SheetSettings.ExcludeUnspentPointsFromTotal = true
	sheet.MarkModified(sheet)
	c.Equal(spent.String()+" Points", panel.total.Text.String(), "turning the setting on leaves the unspent points out")
	built := NewPointsPanel(entity, NewTargetMgr(unison.NewPanel()))
	c.Equal(spent.String()+" Points", built.total.Text.String(), "as does a block built with the setting on")
}
