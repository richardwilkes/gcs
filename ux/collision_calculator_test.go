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
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestCollisionCalculatorSources drives the collision calculator inside a headless workspace the way a user would: it
// opens it from its menu action with a character sheet active, checks that the sheet is preselected as the faller with
// the fields it supplies locked and filled, switches the source to Manual and back through the popup, switches the
// scenario and checks that the rows for it are swapped in, reads the damage from the worked example on B431, and
// finally closes the sheet and checks that the source drops back to Manual with the fields unlocked.
func TestCollisionCalculatorSources(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	calc, ok := openedByAction(t, screen, collisionCalculatorAction).(*CollisionCalculator)
	if !ok {
		t.Fatal("the action must open the collision calculator")
	}

	type state struct {
		sheet                      *Sheet
		hp                         fxp.Int
		sourceIndex                int
		hpEnabled, velocityEnabled bool
		targetShown, extrasShown   bool
		sections                   []*unison.Panel
		damage                     string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.sheet = calc.mover.sheet
			s.hp = calc.mover.hp
			s.sourceIndex = calc.mover.sourcePopup.SelectedIndex()
			s.hpEnabled = calc.mover.hpField.Enabled()
			s.velocityEnabled = calc.mover.velocityField.Enabled()
			s.targetShown = len(calc.targetSlot.Children()) > 0
			s.extrasShown = calc.mover.extras.Parent() != nil
			s.sections = calc.sectionSlot.Children()
			// The first damage line is the faller's, whatever it is called.
			labels := panelsOfType[*unison.Label](calc.results)
			for i, label := range labels {
				if strings.HasPrefix(label.String(), "Damage to ") && i+1 < len(labels) {
					s.damage = labels[i+1].String()
					break
				}
			}
		})
		return s
	}

	var sheetHP fxp.Int
	screen.Do(func() { sheetHP = sheet.Entity().Attributes.Maximum(gurps.HitPointsID) })
	s := current()
	c.Equal(sheet, s.sheet, "the active sheet must be preselected as the faller")
	c.Equal(1, s.sourceIndex, "the Source popup must show the sheet")
	c.Equal(sheetHP, s.hp, "the HP must come from the sheet")
	c.False(s.hpEnabled, "a field the sheet supplies must be locked")
	c.False(s.velocityEnabled, "the velocity of a fall is computed, not typed in")
	c.True(s.extrasShown, "the faller shows its skills and DR")
	c.False(s.targetShown, "a fall has no struck object")
	c.Equal([]*unison.Panel{calc.fallRows, calc.surfaceRows}, s.sections, "a fall shows the fall and surface rows")

	// B431: a 10 HP character falling 17 yards onto a hard surface hits at 19 yards/second and takes 4d.
	screen.Do(func() {
		calc.fallDistance = fxp.FromInteger(17)
		calc.changed()
	})
	s = current()
	c.Equal(fxp.FromInteger(19), calc.mover.velocity, "the fall velocity must come from the table")
	c.Equal("4d cr", s.damage, "the worked example on B431 must come out at 4d")
	captureScreen(t, c, screen, "collision_calculator_fall")

	choosePopupItem(t, screen, wnd, calc.mover.sourcePopup, 0)
	s = current()
	c.Nil(s.sheet, "choosing Manual must drop the sheet")
	c.Equal(sheetHP, s.hp, "the numbers last read from the sheet must be kept")
	c.True(s.hpEnabled, "a typed-in field must be unlocked")

	choosePopupItem(t, screen, wnd, calc.mover.sourcePopup, 1)
	s = current()
	c.Equal(sheet, s.sheet, "choosing the sheet must make it the source again")
	c.False(s.hpEnabled, "a field the sheet supplies must be locked again")

	scenarioPopup, found := firstPanelOfType[*unison.PopupMenu[collisionScenario]](calc.content)
	if !found {
		t.Fatal("the calculator must offer a scenario popup")
	}
	choosePopupItem(t, screen, wnd, scenarioPopup, twoObjectScenario)
	s = current()
	c.True(s.targetShown, "a collision between two objects shows the struck object")
	c.False(s.extrasShown, "the striking object has no use for skills or DR")
	c.True(s.velocityEnabled, "the striking object's velocity is typed in")
	c.Equal([]*unison.Panel{calc.dropRows, calc.fallRows, calc.angleRows}, s.sections,
		"a collision between two objects shows the drop, fall and angle rows")

	// B430: a 60 HP car at 25 yards/second rear-ends a 10 HP pedestrian fleeing at 5: 12d to the pedestrian, 2d back.
	choosePopupItem(t, screen, wnd, calc.mover.sourcePopup, 0)
	screen.Do(func() {
		calc.mover.hp = fxp.FromInteger(60)
		calc.mover.st = 0 // A car has no ST score, so the overrun uses half its HP.
		calc.mover.velocity = fxp.FromInteger(25)
		calc.mover.sm = 2
		calc.target.hp = fxp.FromInteger(10)
		calc.target.velocity = fxp.Five
		calc.angleIndex = slices.IndexFunc(collisionAngles, func(a collisionAngleChoice) bool {
			return a.angle == gurps.RearEndCollision
		})
		calc.changed()
	})
	captureScreen(t, c, screen, "collision_calculator_two_objects")
	var results []string
	screen.Do(func() {
		for _, label := range panelsOfType[*unison.Label](calc.results) {
			results = append(results, label.String())
		}
	})
	c.Equal([]string{
		"Collision velocity:", "20 yards/second (40 mph)",
		"Damage to the struck object:", "12d cr",
		"Damage to the striking object:", "2d cr",
		"Overrun damage:", "3d cr",
	}, results, "the worked example on B430 must come out at 12d and 2d, with an overrun for ST 30")

	choosePopupItem(t, screen, wnd, scenarioPopup, fallScenario)
	choosePopupItem(t, screen, wnd, calc.mover.sourcePopup, 1)
	closeEditorWithoutPrompt(t, screen, sheet)
	screen.Do(func() { calc.changed() })
	s = current()
	c.Nil(s.sheet, "closing the sheet must drop it as the source")
	c.Equal(0, s.sourceIndex, "the Source popup must show Manual once the sheet is gone")
	c.True(s.hpEnabled, "the fields must be unlocked once the sheet is gone")

	closeEditorWithoutPrompt(t, screen, calc)
}
