// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculators

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestCollisionCalculatorResults drives the collision calculator inside a headless workspace the way a user would: it
// opens the calculators from their menu action with a character sheet active, works the example on BX431 for that
// character's fall, switches the scenario and checks that the rows for it are swapped in, works the example on BX432
// with numbers typed in, and checks that choosing the sheet again suggests the character's Move as the velocity.
func TestCollisionCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	dockable := openCalculator(t, screen)
	calc := dockable.collision
	selectCalculatorTab(t, screen, dockable, calc)

	type state struct {
		velocityEnabled          bool
		targetShown, extrasShown bool
		sections                 []*unison.Panel
		velocity                 string
		results                  []string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.velocityEnabled = calc.mover.velocityField.Enabled()
			s.targetShown = len(calc.targetSlot.Children()) > 0
			s.extrasShown = calc.mover.extras.Parent() != nil
			s.sections = calc.sectionSlot.Children()
			s.velocity = calc.mover.velocityField.Text()
			s.results = uxtest.LabelTexts(calc.results)
		})
		return s
	}

	s := current()
	c.False(s.velocityEnabled, "the velocity of a fall is computed, not typed in")
	c.True(s.extrasShown, "the faller shows its skills and DR")
	c.False(s.targetShown, "a fall has no struck object")
	c.Equal([]*unison.Panel{calc.fallRows, calc.surfaceRows}, s.sections, "a fall shows the fall and surface rows")

	// BX431: a 10 HP character falling 17 yards onto a hard surface hits at 19 yards/second and takes 4d.
	screen.Do(func() { calc.fallDistanceField.SetText("17") })
	s = current()
	c.Equal("19", s.velocity, "the fall velocity must come from the table")
	// The first damage line is the faller's, whatever it is called.
	damage := slices.IndexFunc(s.results, func(label string) bool { return strings.HasPrefix(label, "Damage to ") })
	if damage < 0 || damage+1 >= len(s.results) {
		t.Fatalf("the results must give the damage to the faller, but are %q", s.results)
	}
	c.Equal("4d cr", s.results[damage+1], "the worked example on BX431 must come out at 4d")
	uxtest.CaptureScreen(t, c, screen, "collision_calculator_fall")

	var scenarioPopup *unison.PopupMenu[collisionScenario]
	screen.Do(func() { scenarioPopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[collisionScenario]](calc.content) })
	if scenarioPopup == nil {
		t.Fatal("the calculator must offer a scenario popup")
	}
	uxtest.ChoosePopupItem(t, screen, wnd, scenarioPopup, twoObjectScenario)
	s = current()
	c.True(s.targetShown, "a collision between two objects shows the struck object")
	c.False(s.extrasShown, "the striking object has no use for skills or DR")
	c.True(s.velocityEnabled, "the striking object's velocity is typed in")
	c.Equal([]*unison.Panel{calc.dropRows, calc.fallRows, calc.angleRows}, s.sections,
		"a collision between two objects shows the drop, fall and angle rows")

	// BX432: a 60 HP car at 25 yards/second rear-ends a 10 HP pedestrian fleeing at 5: 12d to the pedestrian, 2d back.
	var anglePopup *unison.PopupMenu[calculator.CollisionAngleChoice]
	screen.Do(func() {
		anglePopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[calculator.CollisionAngleChoice]](calc.content)
	})
	if anglePopup == nil {
		t.Fatal("the calculator must offer a collision angle popup")
	}
	uxtest.ChoosePopupItem(t, screen, wnd, calc.mover.popup, 0)
	uxtest.ChoosePopupItem(t, screen, wnd, anglePopup, slices.IndexFunc(calc.angles,
		func(a calculator.CollisionAngleChoice) bool { return a.Angle == calculator.RearEndCollision }))
	screen.Do(func() {
		calc.mover.hpField.SetText("60")
		calc.mover.stField.SetText("0") // A car has no ST score, so the overrun uses half its HP.
		calc.mover.velocityField.SetText("25")
		calc.mover.smField.SetText("2")
		calc.target.hpField.SetText("10")
		calc.target.velocityField.SetText("5")
	})
	uxtest.CaptureScreen(t, c, screen, "collision_calculator_two_objects")
	c.Equal([]string{
		"Collision velocity:", "20 yards/second (40 mph)",
		"Damage to the struck object:", "12d cr",
		"Damage to the striking object:", "2d cr",
		"Overrun damage:", "3d cr",
	}, current().results, "the worked example on BX432 must come out at 12d and 2d, with an overrun for ST 30")

	// Choosing a sheet suggests its character's Move as the velocity, in place of the car's.
	var sheetMove string
	screen.Do(func() {
		entity := sheet.Entity()
		sheetMove = strconv.Itoa(entity.Move(entity.EncumbranceLevel(false)))
	})
	uxtest.ChoosePopupItem(t, screen, wnd, calc.mover.popup, 1)
	c.Equal(sheetMove, current().velocity, "choosing a sheet must suggest its Move as the velocity")

	uxtest.CloseEditorWithoutPrompt(t, screen, dockable)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}
