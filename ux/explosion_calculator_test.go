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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestExplosionCalculatorSources drives the explosion calculator inside a headless workspace the way a user would: it
// opens the calculators from their menu action with a character sheet active, checks that the sheet is preselected as
// the target with the fields it supplies locked and filled, works the collateral damage example on BX414, switches the
// attack type to a cone and works the width example on BX413, works the scatter and demolition examples on BX414 and
// BX415 on their own tabs, and finally closes the sheet and checks that the target drops back to Manual with the fields
// unlocked.
func TestExplosionCalculatorSources(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	dockable := openCalculator(t, screen)
	calc := dockable.explosion
	selectCalculatorTab(t, screen, dockable, calc)

	type state struct {
		sheet                           *Sheet
		hp                              fxp.Int
		sourceIndex                     int
		smEnabled, hpEnabled, drEnabled bool
		exposedEnabled, situationShown  bool
		sections                        []*unison.Panel
		results                         []string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.sheet = calc.target.sheet
			s.hp = calc.target.hp
			s.sourceIndex = calc.target.popup.SelectedIndex()
			s.smEnabled = calc.target.smField.Enabled()
			s.hpEnabled = calc.target.hpField.Enabled()
			s.drEnabled = calc.target.drField.Enabled()
			s.exposedEnabled = calc.target.exposedPopup.Enabled()
			s.situationShown = len(calc.target.situationSlot.Children()) > 0
			s.sections = calc.attackSlot.Children()
			s.results = labelTexts(calc.results)
		})
		return s
	}

	var sheetHP fxp.Int
	screen.Do(func() { sheetHP = sheet.Entity().Attributes.Maximum(gurps.HitPointsID) })
	s := current()
	c.Equal(sheet, s.sheet, "the active sheet must be preselected as the target")
	c.Equal(1, s.sourceIndex, "the Source popup must show the sheet")
	c.Equal(sheetHP, s.hp, "the HP must come from the sheet")
	c.False(s.smEnabled, "a field the sheet supplies must be locked")
	c.False(s.hpEnabled, "a field the sheet supplies must be locked")
	c.False(s.drEnabled, "the DR the sheet works out per BX400 must be locked")
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
	captureScreen(t, c, screen, "explosion_calculator")

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

	attackTypePopup, found := firstPanelOfType[*unison.PopupMenu[explosionAttackType]](calc.content)
	if !found {
		t.Fatal("the calculator must offer an attack type popup")
	}
	choosePopupItem(t, screen, wnd, attackTypePopup, coneAttack)
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

	// BX414: an attack that misses by 3 at 10 yards scatters 3 yards, and one that misses by 3 with the miss squared
	// would scatter 9 but is limited to half the distance.
	scatter := dockable.scatter
	selectCalculatorTab(t, screen, dockable, scatter)
	causePopup, found := firstPanelOfType[*unison.PopupMenu[scatterCause]](scatter.content)
	if !found {
		t.Fatal("the calculator must offer a cause popup")
	}
	var scattered string
	screen.Do(func() {
		scatter.marginField.SetText("3")
		scattered = scatter.result.String()
	})
	c.Equal("3 yards", scattered, "a miss scatters by its margin")
	choosePopupItem(t, screen, wnd, causePopup, 1)
	screen.Do(func() { scattered = scatter.result.String() })
	c.Equal("5 yards (limited to half the distance)", scattered, "a squared miss is limited to half the distance")
	captureScreen(t, c, screen, "scatter_calculator")

	// BX415: a 6dx8 blast takes 16 lbs of TNT, or 20 lbs of dynamite, whose REF is 0.8.
	demolition := dockable.demolition
	selectCalculatorTab(t, screen, dockable, demolition)
	var refBlank, weightBlank, countBlank bool
	screen.Do(func() {
		refBlank = !demolition.refField.Enabled() && demolition.refField.DrawOverCallback != nil
		weightBlank = !demolition.explosiveWeightField.Enabled() && demolition.explosiveWeightField.DrawOverCallback != nil
		countBlank = !demolition.blastCountField.Enabled() || demolition.blastCountField.DrawOverCallback != nil
	})
	c.True(refBlank, "a preset explosive's REF field is disabled and blank, since what is typed in it is not used")
	c.True(weightBlank, "the weight is not used when working out the explosive a blast needs, so it is blank")
	c.False(countBlank, "the blast multiplier is in use, so it is editable and shown")
	explosivePopup, found := firstPanelOfType[*unison.PopupMenu[explosiveChoice]](demolition.content)
	if !found {
		t.Fatal("the calculator must offer an explosive popup")
	}
	choosePopupItem(t, screen, wnd, explosivePopup, slices.IndexFunc(explosiveChoices,
		func(e explosiveChoice) bool { return e.title == "Dynamite" }))
	var damage, tnt, explosiveLabel, explosiveWeight string
	screen.Do(func() {
		demolition.blastCountField.SetText("8")
		damage = demolition.damageResult.String()
		tnt = demolition.tntResult.String()
		explosiveLabel = demolition.explosiveWeightLabel.String()
		explosiveWeight = demolition.explosiveWeightResult.String()
	})
	c.Equal("6dx8 cr ex", damage, "the blast is 6d times the multiplier")
	c.Equal("16 lb", tnt, "a 6dx8 blast takes (8x8)/4 lbs of TNT")
	c.Equal("Dynamite:", explosiveLabel, "the weight is labeled with the explosive it is of")
	c.Equal("20 lb", explosiveWeight, "the worked example on BX415 must come out at 20 lbs of dynamite")
	captureScreen(t, c, screen, "demolition_calculator")

	selectCalculatorTab(t, screen, dockable, calc)
	closeEditorWithoutPrompt(t, screen, sheet)
	screen.Do(func() { calc.changed() })
	s = current()
	c.Nil(s.sheet, "closing the sheet must drop it as the source")
	c.Equal(0, s.sourceIndex, "the Source popup must show Manual once the sheet is gone")
	c.True(s.smEnabled, "the fields must be unlocked once the sheet is gone")
	c.True(s.hpEnabled, "the fields must be unlocked once the sheet is gone")
	c.True(s.drEnabled, "the DR must be typed in once the sheet is gone")
	c.False(s.exposedEnabled, "without a sheet there are no locations to choose among")

	closeEditorWithoutPrompt(t, screen, dockable)
}
