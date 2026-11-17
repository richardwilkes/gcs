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
	"testing"

	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestExplosionCalculatorResults drives the explosion calculator in a headless workspace with a character sheet active
// as the target: the worked examples on BX413-BX415 must come out for an explosion, its fragments and a cone, each of
// the two must show its own rows, and closing the sheet while the attack is a cone must drop the target back to Manual
// with the fields unlocked and no exposed locations to choose among.
func TestExplosionCalculatorResults(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	dockable := openCalculator(t, screen)
	calc := dockable.explosion
	selectCalculatorTab(t, screen, dockable, calc)

	type state struct {
		exposedEnabled, situationShown bool
		sections                       []*unison.Panel
		results                        []string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.exposedEnabled = calc.target.exposedPopup.Enabled()
			s.situationShown = len(calc.target.situationSlot.Children()) > 0
			s.sections = calc.attackSlot.Children()
			s.results = uxtest.LabelTexts(calc.results)
		})
		return s
	}

	s := current()
	c.True(s.exposedEnabled, "the exposed locations of a sheet may be chosen among for an explosion")
	c.True(s.situationShown, "an explosion asks where the target was")
	c.Equal([]*unison.Panel{calc.explosionRows}, s.sections, "an explosion shows the environment and its checkboxes")
	// BX414: a 6dx2 blast reaches 24 yards, and 3 yards out the damage is divided by 9, which is 1/4/8.
	screen.Do(func() {
		calc.blastField.SetText("6dx2")
		calc.target.distanceField.SetText("3")
	})
	s = current()
	c.Equal([]string{
		"Collateral damage radius:", "24 yards (12 dice)",
		"Damage divisor:", "9 (3 × 3 yards)",
		"Blast damage:", "6dx2 cr ex ÷ 9, rounded down",
		"Minimum / average / maximum:", "1 / 4 / 8",
		"DR against the blast:", "0 (Large-Area Injury)",
		"Penetrating (average):", "4",
	}, s.results, "the worked example on BX414 must come out at 1/4/8")
	uxtest.CaptureScreen(t, c, screen, "explosion_calculator")

	// BX414-BX415: [2d] fragmentation reaches 10 yards, and the fragments roll against 15 less the range penalty, which
	// is -1 at 3 yards; the arithmetic is shown only while a modifier applies.
	screen.Do(func() { calc.fragmentationField.SetText("2d") })
	s = current()
	c.Equal([]string{
		"Fragmentation radius:", "10 yards (2 dice)",
		"Fragments hit:", "on a roll of 14 or less (base 15, range -1, posture +0, SM +0)",
		"Fragment damage:", "2d cut, per fragment",
	}, s.results[12:], "the fragments must roll against 15 less the range penalty")
	screen.Do(func() { calc.target.distanceField.SetText("2") })
	s = current()
	c.Equal([]string{"Fragments hit:", "on a roll of 15 or less"}, s.results[14:16],
		"with no modifier the total needs no arithmetic")
	screen.Do(func() {
		calc.fragmentationField.SetText("")
		calc.target.distanceField.SetText("3")
	})

	var attackTypePopup *unison.PopupMenu[explosionAttackType]
	screen.Do(func() {
		attackTypePopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[explosionAttackType]](calc.content)
	})
	if attackTypePopup == nil {
		t.Fatal("the calculator must offer an attack type popup")
	}
	uxtest.ChoosePopupItem(t, screen, wnd, attackTypePopup, coneAttack)
	// BX413: a cone with a maximum range of 100 yards and a maximum width of 5 is 3 yards wide at 60 yards.
	screen.Do(func() {
		calc.coneRangeField.SetText("100")
		calc.coneWidthField.SetText("5")
		calc.target.distanceField.SetText("60")
	})
	s = current()
	c.Equal([]*unison.Panel{calc.coneRows, calc.areaRows}, s.sections, "a cone shows the cone and dissipation rows")
	c.False(s.situationShown, "only an explosion asks where the target was")
	c.Equal([]string{
		"Cone width at 60 yards:", "3 yards",
		"Damage divisor:", "None (damage does not decline with distance)",
		"Damage:", "6dx2 cr",
		"Minimum / average / maximum:", "12 / 42 / 72",
		"DR against the attack:", "0 (Large-Area Injury)",
		"Penetrating (average):", "42",
	}, s.results, "the worked example on BX413 must come out 3 yards wide")

	c.True(s.exposedEnabled, "the exposed locations of a sheet may be chosen among for a cone too")

	// TestCalculatorSheetSources closes the sheet on an explosion; a cone must drop it just the same.
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
	var source *ux.Sheet
	var sourceIndex int
	var unlocked []bool
	screen.Do(func() {
		calc.changed()
		source = calc.target.sheet
		sourceIndex = calc.target.popup.SelectedIndex()
		for _, field := range []unison.Paneler{
			calc.target.smField, calc.target.hpField, calc.target.torsoDRField, calc.target.drField,
		} {
			unlocked = append(unlocked, field.AsPanel().Enabled())
		}
	})
	c.Nil(source, "closing the sheet must drop it as the source")
	c.Equal(0, sourceIndex, "the Source popup must show Manual once the sheet is gone")
	c.Equal([]bool{true, true, true, true}, unlocked, "the SM, HP and DR must be typed in once the sheet is gone")
	c.False(current().exposedEnabled, "without a sheet there are no locations to choose among")

	uxtest.CloseEditorWithoutPrompt(t, screen, dockable)
}
