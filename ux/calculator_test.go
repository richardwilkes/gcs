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
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
)

// openCalculator opens the calculators from their menu action, which preselects the active sheet, and returns the
// dockable holding them.
func openCalculator(t *testing.T, screen *unison.HeadlessScreen) *Calculator {
	t.Helper()
	calc, ok := openedByAction(t, screen, calculatorAction).(*Calculator)
	if !ok {
		t.Fatal("the action must open the calculators")
	}
	return calc
}

// selectCalculatorTab clicks the tab for the given calculator the way a user would and checks that its content is
// what the dockable then shows.
func selectCalculatorTab(t *testing.T, screen *unison.HeadlessScreen, calc *Calculator, tab calculatorTab) {
	t.Helper()
	index := slices.Index(calc.tabs, tab)
	if index < 0 {
		t.Fatalf("%s is not one of the calculators", tab.title())
	}
	screen.Click(screen.PanelCenter(calc.tabBar.buttons[index]))
	var shown []*unison.Panel
	var selected int
	screen.Do(func() {
		shown = calc.slot.Children()
		selected = calc.tabBar.selectedIndex()
	})
	if selected != index || len(shown) != 1 || shown[0] != tab.panel() {
		t.Fatalf("clicking the %s tab must show that calculator alone", tab.title())
	}
}

// TestCalculatorTabs verifies the calculators open as one dockable with a tab per calculator, in the documented order,
// that clicking a tab swaps that calculator's content in beneath the tab bar, that the sheet the calculators were
// opened from is preselected as the source on every tab that takes a character's numbers, and that a second request
// brings the open dockable forward rather than opening another.
func TestCalculatorTabs(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	calc := openCalculator(t, screen)

	var titles []string
	var shown []*unison.Panel
	var sources []*Sheet
	screen.Do(func() {
		for _, button := range calc.tabBar.buttons {
			titles = append(titles, button.Text.String())
		}
		shown = calc.slot.Children()
		sources = []*Sheet{
			calc.collision.mover.sheet, calc.jumping.source.sheet, calc.throwing.source.sheet,
			calc.hiking.source.sheet, calc.explosion.target.sheet,
		}
	})
	c.Equal([]string{
		"Explosions & Area Attacks", "Scatter", "Demolition", "Collisions & Falls", "Jumping", "Throwing", "Hiking",
	}, titles, "the tabs must be labeled with the calculators, in order")
	c.Equal([]*unison.Panel{calc.explosion.content}, shown, "the first tab must be showing when the calculators open")
	c.Equal([]*Sheet{sheet, sheet, sheet, sheet, sheet}, sources,
		"the active sheet must be preselected on every calculator that takes a character's numbers")

	for _, tab := range calc.tabs {
		selectCalculatorTab(t, screen, calc, tab)
	}
	captureScreen(t, c, screen, "calculator_hiking_tab")
	selectCalculatorTab(t, screen, calc, calc.jumping)

	// Choosing the menu item again brings the open calculators forward, leaving the chosen tab alone.
	var opened int
	screen.Do(func() {
		before := len(AllDockables())
		calculatorAction.Execute(nil)
		opened = len(AllDockables()) - before
	})
	c.Equal(0, opened, "a second request must not open a second set of calculators")
	var selected int
	screen.Do(func() { selected = calc.tabBar.selectedIndex() })
	c.Equal(slices.Index(calc.tabs, calculatorTab(calc.jumping)), selected,
		"bringing the calculators forward must not change the tab")

	// Squeezing the workspace narrower than the rows must leave the calculators scrolling sideways rather than
	// clipping: the scroll panel's content has to stay as wide as the calculator needs, which includes the rows
	// nested in the groups that come and go, such as the target's posture popup on the explosions tab.
	selectCalculatorTab(t, screen, calc, calc.collision)
	var viewWidth, contentWidth, neededWidth float32
	screen.Do(func() {
		wnd.SetFrameRect(geom.NewRect(0, 0, 900, 900))
		wnd.ValidateLayout()
		viewWidth = calc.scroll.ContentView().FrameRect().Width
		contentWidth = calc.scroll.Content().AsPanel().FrameRect().Width
		_, needed, _ := calc.collision.content.Sizes(geom.Size{Width: viewWidth})
		neededWidth = needed.Width
	})
	c.True(viewWidth < neededWidth, "the narrowed view must be too narrow for the collision rows")
	c.Equal(neededWidth, contentWidth, "the scroll content must stay as wide as the calculator needs")
	selectCalculatorTab(t, screen, calc, calc.explosion)
	var popupRight, viewRight, contentRight float32
	screen.Do(func() {
		wnd.ValidateLayout()
		view := calc.scroll.ContentView()
		viewRight = view.PointToRoot(geom.NewPoint(view.FrameRect().Width, 0)).X
		content := calc.scroll.Content().AsPanel()
		contentRight = content.PointToRoot(geom.NewPoint(content.FrameRect().Width, 0)).X
		if popup, found := firstPanelOfType[*unison.PopupMenu[explosionPosture]](calc.explosion.content); found {
			popupRight = popup.PointToRoot(geom.NewPoint(popup.FrameRect().Width, 0)).X
		}
	})
	captureScreen(t, c, screen, "calculator_narrow_explosion")
	c.True(popupRight > 0, "the explosions tab must offer a posture popup")
	c.True(popupRight > viewRight, "the narrowed view must be too narrow for the posture popup")
	c.True(popupRight <= contentRight, "the posture popup must lie within the scroll content, however narrow the view")

	closeEditorWithoutPrompt(t, screen, calc)
	closeEditorWithoutPrompt(t, screen, sheet)
}

// TestJumpingCalculatorSources drives the jumping calculator inside a headless workspace the way a user would: it
// checks that the preselected sheet fills and locks the jumper's fields, works the basic example on BX352 with numbers
// typed in after switching the source to Manual, and checks that closing the sheet drops it as the source.
func TestJumpingCalculatorSources(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	calc := openCalculator(t, screen)
	jumping := calc.jumping
	selectCalculatorTab(t, screen, calc, jumping)

	type state struct {
		sheet                     *Sheet
		basicMove, liftingST      fxp.Int
		weight                    fxp.Weight
		sourceIndex               int
		moveEnabled, encEnabled   bool
		high, broad, runningStart string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.sheet = jumping.source.sheet
			s.basicMove = jumping.basicMove
			s.liftingST = jumping.liftingST
			s.weight = jumping.weight
			s.sourceIndex = jumping.source.popup.SelectedIndex()
			s.moveEnabled = jumping.basicMoveField.Enabled()
			s.encEnabled = jumping.encumbrancePopup.Enabled()
			s.high = jumping.highJumpResult.String()
			s.broad = jumping.broadJumpResult.String()
			s.runningStart = jumping.runningStartLabel.String()
		})
		return s
	}

	var sheetMove, sheetST fxp.Int
	var sheetWeight fxp.Weight
	screen.Do(func() {
		entity := sheet.Entity()
		sheetMove = entity.ResolveAttributeCurrent(gurps.BasicMoveID)
		sheetST = entity.LiftingStrength()
		sheetWeight = entity.Profile.Weight
	})
	s := current()
	c.Equal(sheet, s.sheet, "the active sheet must be preselected as the jumper")
	c.Equal(1, s.sourceIndex, "the Source popup must show the sheet")
	c.Equal(sheetMove, s.basicMove, "the Basic Move must come from the sheet")
	c.Equal(sheetST, s.liftingST, "the Lifting ST must come from the sheet")
	c.Equal(sheetWeight, s.weight, "the body weight must come from the sheet")
	c.False(s.moveEnabled, "a field the sheet supplies must be locked")
	c.False(s.encEnabled, "the encumbrance the sheet supplies must be locked")
	c.Equal("yard running start", s.runningStart, "a sheet in yards measures the running start in yards")
	captureScreen(t, c, screen, "jumping_calculator")

	// BX352: with Basic Move 5 and ST 10, a standing high jump is 6×5−10 = 20 inches and a broad jump 2×5−3 = 7 feet.
	choosePopupItem(t, screen, wnd, jumping.source.popup, 0)
	s = current()
	c.Nil(s.sheet, "choosing Manual must drop the sheet")
	c.True(s.moveEnabled, "a typed-in field must be unlocked")
	c.True(s.encEnabled, "a typed-in encumbrance must be unlocked")
	screen.Do(func() {
		jumping.basicMoveField.SetText("5")
		jumping.liftingSTField.SetText("10")
		jumping.weightField.SetText("150")
	})
	s = current()
	c.Equal("1 foot, 8 inches", s.high, "the worked example on BX352 must come out at 20 inches")
	c.Equal("2 yards, 1 foot", s.broad, "the worked example on BX352 must come out at 7 feet")

	// BX357: extra effort at -2 adds 10% to the distances, and the note says what a failure leaves.
	var notes []string
	screen.Do(func() {
		jumping.extraEffortPenalty = -2
		jumping.changed()
		notes = slices.DeleteFunc(labelTexts(jumping.notes), func(text string) bool { return text == "•" })
	})
	s = current()
	c.Equal("1 foot, 10 inches", s.high, "extra effort at -2 must add 10% to the high jump")
	c.Equal("2 yards, 1 foot, 8 inches", s.broad, "extra effort at -2 must add 10% to the broad jump")
	c.Equal([]string{
		"The extra effort takes a Will roll, or a Will-based Jumping roll if that is better, at -2 for the +10% shown, and costs 1 FP whether it succeeds or fails. A failure leaves the jump as it would be without it: 1 foot, 8 inches high and 2 yards, 1 foot broad. A critical failure costs 1 HP of injury to a foot or leg instead and the jump fails, and on a natural 18 a HT roll is needed as well to avoid a temporary Crippled Leg (B357).",
	}, notes, "extra effort must be explained, with what a failure leaves")
	screen.Do(func() {
		jumping.extraEffortPenalty = 0
		jumping.changed()
	})

	choosePopupItem(t, screen, wnd, jumping.source.popup, 1)
	c.Equal(sheet, current().sheet, "choosing the sheet must make it the source again")
	closeEditorWithoutPrompt(t, screen, sheet)
	screen.Do(func() { jumping.changed() })
	s = current()
	c.Nil(s.sheet, "closing the sheet must drop it as the source")
	c.Equal(0, s.sourceIndex, "the Source popup must show Manual once the sheet is gone")
	c.True(s.moveEnabled, "the fields must be unlocked once the sheet is gone")

	closeEditorWithoutPrompt(t, screen, calc)
}

// TestThrowingCalculatorSources drives the throwing calculator inside a headless workspace the way a user would: it
// checks that the preselected sheet fills and locks the thrower's fields, works the example on BX355 with numbers
// typed in after switching the source to Manual, and checks that the throwing skills add to the distance and damage.
func TestThrowingCalculatorSources(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	calc := openCalculator(t, screen)
	throwing := calc.throwing
	selectCalculatorTab(t, screen, calc, throwing)

	type state struct {
		sheet                      *Sheet
		st, strikingST             fxp.Int
		sourceIndex                int
		stEnabled, throwingEnabled bool
		distance, damage           string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.sheet = throwing.source.sheet
			s.st = throwing.st
			s.strikingST = throwing.strikingST
			s.sourceIndex = throwing.source.popup.SelectedIndex()
			s.stEnabled = throwing.stField.Enabled()
			s.throwingEnabled = throwing.throwingPopup.Enabled()
			s.distance = throwing.distanceResult.String()
			s.damage = throwing.damageResult.String()
		})
		return s
	}

	var sheetST, sheetStrikingST fxp.Int
	screen.Do(func() {
		entity := sheet.Entity()
		sheetST = entity.LiftingStrength() - entity.LiftingStrengthBonus
		sheetStrikingST = entity.StrikingStrength()
	})
	s := current()
	c.Equal(sheet, s.sheet, "the active sheet must be preselected as the thrower")
	c.Equal(1, s.sourceIndex, "the Source popup must show the sheet")
	c.Equal(sheetST, s.st, "the ST must come from the sheet")
	c.Equal(sheetStrikingST, s.strikingST, "the Striking ST must come from the sheet")
	c.False(s.stEnabled, "a field the sheet supplies must be locked")
	c.False(s.throwingEnabled, "the skill the sheet supplies must be locked")

	// BX355: ST 10 throws a 1 lb object, a twentieth of its Basic Lift of 20, 3.5×ST = 35 yards, for thrust−2 per die:
	// 1d−2 becomes 1d−4.
	choosePopupItem(t, screen, wnd, throwing.source.popup, 0)
	s = current()
	c.Nil(s.sheet, "choosing Manual must drop the sheet")
	c.True(s.stEnabled, "a typed-in field must be unlocked")
	c.True(s.throwingEnabled, "a typed-in skill must be unlocked")
	screen.Do(func() {
		throwing.stField.SetText("10")
		throwing.strikingSTField.SetText("10")
		throwing.weightField.SetText("1")
	})
	s = current()
	c.Equal("35 yards", s.distance, "the worked example on BX355 must come out at 35 yards")
	c.Equal("1d-4", s.damage, "a light object does thrust−2 per die")
	captureScreen(t, c, screen, "throwing_calculator")

	// BX357: extra effort at -2 raises the ST for both distance and damage by 10%, to 11: 11 × 3.5 = 38.5 yards, and
	// thrust for ST 11 is 1d-1, less 2 for the light object.
	var notes []string
	screen.Do(func() {
		throwing.extraEffortPenalty = -2
		throwing.changed()
		notes = slices.DeleteFunc(labelTexts(throwing.notes), func(text string) bool { return text == "•" })
	})
	s = current()
	c.Equal("38 yards, 1 foot, 6 inches", s.distance, "extra effort at -2 must add 10% to the ST for distance")
	c.Equal("1d-3", s.damage, "extra effort at -2 must add 10% to the ST for damage")
	c.Equal([]string{
		"The extra effort takes a Will roll, or a Will-based Throwing roll if that is better, at -2 for the +10% ST shown, and costs 1 FP whether it succeeds or fails. A failure leaves the throw as it would be without it: 35 yards for 1d-4. A critical failure costs 1 HP of injury instead and the throw fails, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357).",
	}, notes, "extra effort must be explained, with what a failure leaves")
	screen.Do(func() {
		throwing.extraEffortPenalty = 0
		throwing.changed()
	})

	// Throwing Art at DX+1 adds 2 to the ST for distance and 2 per die to the damage: 12×3.5 = 42 yards and 1d−2.
	throwingArtPopup, found := firstPanelOfType[*unison.PopupMenu[throwingSkillTier]](throwing.content)
	if !found {
		t.Fatal("the calculator must offer a Throwing popup")
	}
	c.Equal(throwing.throwingPopup, throwingArtPopup, "the Throwing popup comes first")
	choosePopupItem(t, screen, wnd, throwing.throwingArtPopup, 2)
	s = current()
	c.Equal("42 yards", s.distance, "Throwing Art at DX+1 must add 2 to the ST the distance is worked out from")
	c.Equal("1d-2", s.damage, "Throwing Art at DX+1 must add 2 per die to the damage")

	closeEditorWithoutPrompt(t, screen, calc)
	closeEditorWithoutPrompt(t, screen, sheet)
}

// TestHikingTimeInDays verifies the hiking travel-time calculation, including the 0 Move case that previously divided by
// zero and crashed the Calculator the moment a non-zero "Distance to Cover" was entered.
func TestHikingTimeInDays(t *testing.T) {
	c := check.New(t)
	for _, one := range []struct {
		name            string
		distanceToCover fxp.Int
		distancePerDay  fxp.Int
		wantDays        fxp.Int
		wantOK          bool
	}{
		// Regression: distancePerDay of 0 (Move resolved to 0) must not divide by zero, even with distance to cover.
		{name: "0 move with distance to cover", distanceToCover: fxp.FromInteger(100), distancePerDay: 0, wantDays: 0, wantOK: false},
		{name: "0 move with no distance to cover", distanceToCover: 0, distancePerDay: 0, wantDays: 0, wantOK: false},
		// Nothing to cover is already "there": 0 days, and no division hazard.
		{name: "no distance to cover", distanceToCover: 0, distancePerDay: fxp.FromInteger(20), wantDays: 0, wantOK: true},
		// 100 miles to cover at 20 miles/day -> 5 days exactly.
		{name: "even multiple", distanceToCover: fxp.FromInteger(100), distancePerDay: fxp.FromInteger(20), wantDays: fxp.FromInteger(5), wantOK: true},
		// Rounds to a tenth of a day: 10 / 3 = 3.333... -> 3.3 days.
		{name: "rounds to tenths", distanceToCover: fxp.FromInteger(10), distancePerDay: fxp.FromInteger(3), wantDays: fxp.FromStringForced("3.3"), wantOK: true},
	} {
		days, ok := hikingTimeInDays(one.distanceToCover, one.distancePerDay)
		c.Equal(one.wantOK, ok, one.name)
		c.Equal(one.wantDays, days, one.name)
	}
}

// TestCalculatorHikingControls drives the hiking calculator inside a headless workspace the way a user would: it checks
// that the preselected sheet fills and locks the hiker's Move, picks terrain, weather and intensity from the popups and
// clicks the checkboxes, checking that each choice is stored and that the controls depending on it are enabled,
// disabled or retitled to match, and works a day's travel from numbers typed in.
func TestCalculatorHikingControls(t *testing.T) {
	c := check.New(t)
	screen, wnd := startHeadlessWorkspace(t, c)
	sheet, ok := openedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	screen.Do(func() { DisplayCalculator(sheet) })
	calc := soleEditor[*Calculator](t, screen, func(d unison.Dockable) bool {
		_, isCalculator := d.AsPanel().Self.(*Calculator)
		return isCalculator
	})
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)

	var terrainPopup, weatherPopup *unison.PopupMenu[terrainModifier]
	var intensityPopup *unison.PopupMenu[hikingIntensityHours]
	screen.Do(func() {
		if popups := panelsOfType[*unison.PopupMenu[terrainModifier]](hiking.content); len(popups) == 2 {
			terrainPopup, weatherPopup = popups[0], popups[1]
		}
		intensityPopup, _ = firstPanelOfType[*unison.PopupMenu[hikingIntensityHours]](hiking.content)
	})
	if terrainPopup == nil || weatherPopup == nil || intensityPopup == nil {
		t.Fatal("the calculator must offer terrain, weather and intensity popups")
	}
	// Every row and checkbox is a direct child of the content, indented beneath a header: the hiker's source, two
	// fields, encumbrance, FP, fitness and Recover Energy, then the journey's popups, five checkboxes and five fields.
	// The header, the two subheaders and the results box are the rest, and the box holds its own subheaders and rows.
	var children, indented int
	var headers, boxHeaders []string
	screen.Do(func() {
		children = len(hiking.content.Children())
		for _, child := range hiking.content.Children() {
			// The header, the subheaders and the box have either no border or one that sets a different left inset.
			if border := child.Border(); border != nil && border.Insets().Left == unison.StdHSpacing*2 {
				indented++
			}
			if label, isLabel := child.Self.(*unison.Label); isLabel {
				headers = append(headers, label.String())
			}
		}
		for _, child := range hiking.resultsBox.Children() {
			if label, isLabel := child.Self.(*unison.Label); isLabel {
				boxHeaders = append(boxHeaders, label.String())
			}
		}
	})
	c.Equal(22, children, "the content must hold the header, the subheaders, their rows and the results box")
	c.Equal(18, indented, "every row and checkbox must be indented beneath its header")
	c.Equal([]string{"Hiker", "Journey"}, headers, "the parts of the calculator must be labeled")
	c.Equal([]string{"Results"}, boxHeaders, "the results box must be labeled")

	type state struct {
		sheet                                                                  *Sheet
		terrain, weather, intensity, move, encumbrance                         int
		hours, fp                                                              fxp.Int
		moveEnabled                                                            bool
		hoursEnabled, roadsEnabled, skisEnabled, skatesEnabled, penaltyEnabled bool
		roadsCleared, skis, skates, roll                                       bool
		rollTitle, perDay, fpCost                                              string
		notes                                                                  []string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.sheet = hiking.source.sheet
			s.terrain = hiking.terrainIndex
			s.weather = hiking.weatherIndex
			s.intensity = hiking.hikingIntensityIndex
			s.move = hiking.move
			s.encumbrance = hiking.encumbranceIndex
			s.hours = hiking.hikingHours
			s.fp = hiking.fp
			s.moveEnabled = hiking.moveField.Enabled()
			s.hoursEnabled = hiking.hikingHoursField.Enabled()
			s.roadsEnabled = hiking.roadsAreClearedCheckBox.Enabled()
			s.skisEnabled = hiking.usingSkisCheckBox.Enabled()
			s.skatesEnabled = hiking.usingSkatesCheckBox.Enabled()
			s.penaltyEnabled = hiking.hikingExtraEffortField.Enabled()
			s.roadsCleared = hiking.roadsAreCleared
			s.skis = hiking.usingSkis
			s.skates = hiking.usingSkates
			s.roll = hiking.successfulHikingRoll
			// The page the skill is on sits beside the checkbox as a link; reading it with the title keeps one field.
			s.rollTitle = hiking.successfulHikingRollCheckBox.Text.String() + " " + hiking.hikingRollPageLabel.Title()
			s.perDay = hiking.hikingResult.String()
			s.fpCost = hiking.fpResult.String() + " " + hiking.fpLabel.String()
			s.notes = slices.DeleteFunc(labelTexts(hiking.notes), func(text string) bool { return text == "•" })
		})
		return s
	}
	indexOf := func(names []string, name string) int {
		i := slices.Index(names, name)
		if i < 0 {
			t.Fatalf("no %q option", name)
		}
		return i
	}
	terrainNames := make([]string, len(terrain))
	for i, one := range terrain {
		terrainNames[i] = one.Name
	}
	weatherNames := make([]string, len(weather))
	for i, one := range weather {
		weatherNames[i] = one.Name
	}
	intensityNames := make([]string, len(hikingIntensity))
	for i, one := range hikingIntensity {
		intensityNames[i] = one.Name
	}
	dirtRoad := indexOf(terrainNames, "Road, Dirt")
	swamp := indexOf(terrainNames, "Swamp")
	normalWeather := indexOf(weatherNames, "Normal")
	snow := indexOf(weatherNames, "Snow")
	normalIntensity := indexOf(intensityNames, "Normal")
	custom := indexOf(intensityNames, "Custom")
	forcedMarch := indexOf(intensityNames, "Forced March")

	// A fresh calculator takes the hiker from the sheet it was opened for and offers the defaults for the journey, with
	// the hours fixed by the intensity, no road to clear of snow, and no extra effort without a successful roll. An
	// unencumbered normal day costs 8 FP (BX426), which leaves a hiker with 10 FP fewer than a third of them after the
	// 7th hour, so the 8th is walked at half Move: 7 × 3.125 + 1.5625 = 23.4 miles rather than 25.
	var sheetMove int
	var sheetFP fxp.Int
	screen.Do(func() {
		entity := sheet.Entity()
		sheetMove = entity.Move(entity.EncumbranceLevel(false))
		sheetFP = entity.Attributes.Maximum(gurps.FatiguePointsID)
	})
	c.Equal(state{
		sheet:         sheet,
		terrain:       dirtRoad,
		weather:       normalWeather,
		intensity:     normalIntensity,
		move:          sheetMove,
		hours:         fxp.Eight,
		fp:            sheetFP,
		skisEnabled:   true,
		skatesEnabled: true,
		rollTitle:     "Made a successful Hiking roll (B200)",
		perDay:        "23 miles",
		fpCost:        "8 FP lost by the end of the day (1 per hour)",
		notes: []string{
			"Each hour of hiking costs 1 FP, plus 1 per level of encumbrance, plus 1 on a hot day or 2 in plate armor, an overcoat, etc. (B426).",
			"With 10 FP, the hiker has fewer than a third left after 7 hours; Move, Dodge and ST are halved from then on, and the hours below allow for it (B426).",
			"Extra effort needs the Hiking roll, which it makes a single Will-based Hiking roll at -1 per 5% of distance beyond the +20% a success gives (B357).",
		},
	}, current(), "the calculator must start from the sheet's Move and FP and the defaults")
	breakdownRow := func(index int) []string {
		var row []string
		screen.Do(func() {
			cells := labelTexts(hiking.breakdown)
			if 6*index+6 <= len(cells) {
				row = cells[6*index : 6*index+6]
			}
		})
		return row
	}
	c.Equal([]string{"Hours", "Miles", "Total", "FP", "FP left", "Condition"}, breakdownRow(0),
		"the breakdown must be headed")
	c.Equal([]string{"1", "3.1", "3.1", "-1", "9", ""}, breakdownRow(1), "the first hour is walked at full Move")
	c.Equal([]string{"7", "3.1", "21.9", "-1", "3", "Very tired: Move halved"}, breakdownRow(7),
		"the 7th hour leaves fewer than a third of the FP")
	c.Equal([]string{"8", "1.6", "23.4", "-1", "2", "Very tired: Move halved"}, breakdownRow(8),
		"the 8th hour is walked at half Move")

	// Snow on a road lets it be cleared; a swamp cannot be.
	choosePopupItem(t, screen, wnd, weatherPopup, snow)
	s := current()
	c.Equal(snow, s.weather, "choosing weather must store its index")
	c.True(s.roadsEnabled, "snow on a road must allow the road to be cleared")
	choosePopupItem(t, screen, wnd, terrainPopup, swamp)
	s = current()
	c.Equal(swamp, s.terrain, "choosing terrain must store its index")
	c.False(s.roadsEnabled, "a swamp cannot be cleared of snow")
	choosePopupItem(t, screen, wnd, terrainPopup, dirtRoad)
	c.True(current().roadsEnabled, "back on the road, it can be cleared again")

	// The hours are only editable for a custom intensity; any other sets them.
	choosePopupItem(t, screen, wnd, intensityPopup, custom)
	s = current()
	c.Equal(custom, s.intensity, "choosing an intensity must store its index")
	c.True(s.hoursEnabled, "a custom intensity must free the hours field")
	c.Equal(fxp.Eight, s.hours, "switching to custom must keep the hours")
	choosePopupItem(t, screen, wnd, intensityPopup, forcedMarch)
	s = current()
	c.Equal(fxp.Sixteen, s.hours, "a forced march must set the hours")
	c.False(s.hoursEnabled, "a fixed intensity must lock the hours field")
	// BX426: 16 hours at 1 FP per hour is 16 FP; a hiker with 10 FP has fewer than a third left after the 7th hour and
	// none after the 10th, and walks the last 9 hours at half Move. Snow halves the pace on the uncleared road to
	// 1.5625 miles an hour, so the march comes to 7 × 1.5625 + 9 × 0.78125 = 18 miles rather than 25.
	c.Equal("16 FP lost by the end of the day (1 per hour)", s.fpCost, "a forced march costs 16 FP unencumbered")
	c.Equal("18 miles", s.perDay, "a forced march must allow for the hours walked at half Move")
	c.Equal([]string{
		"Each hour of hiking costs 1 FP, plus 1 per level of encumbrance, plus 1 on a hot day or 2 in plate armor, an overcoat, etc. (B426).",
		"With 10 FP, the hiker has fewer than a third left after 7 hours; Move, Dodge and ST are halved from then on, and the hours below allow for it (B426).",
		"The hiker is out of FP after 10 hours; going on takes a Will roll, which the hours below assume is made, and each further FP lost also costs 1 HP (B426).",
		"Extra effort needs the Hiking roll, which it makes a single Will-based Hiking roll at -1 per 5% of distance beyond the +20% a success gives (B357).",
	}, s.notes, "a forced march must warn when the hiker tires and runs out of FP")
	c.Equal([]string{"10", "0.8", "13.3", "-1", "0", "Exhausted: Will roll to go on, 1 HP per FP lost"}, breakdownRow(10),
		"the 10th hour uses up the FP")
	c.Equal([]string{"16", "0.8", "18", "-1", "-6", "Exhausted: Will roll to go on, 1 HP per FP lost"}, breakdownRow(16),
		"the march goes on into negative FP")

	// Skis and skates exclude one another and change which skill the roll is against; a successful roll unlocks the
	// extra effort penalty.
	clickCheckBox := func(cb *unison.CheckBox) {
		screen.Do(func() { cb.ScrollIntoView() })
		screen.Click(screen.PanelCenter(cb))
	}
	clickCheckBox(hiking.usingSkisCheckBox)
	s = current()
	c.True(s.skis, "clicking the skis checkbox must store the choice")
	c.False(s.skatesEnabled, "skis must rule out skates")
	c.Equal("Made a successful Skiing roll (B221)", s.rollTitle, "on skis, the roll is against Skiing")
	clickCheckBox(hiking.successfulHikingRollCheckBox)
	s = current()
	c.True(s.roll, "clicking the roll checkbox must store the choice")
	c.True(s.penaltyEnabled, "a successful roll must unlock the extra effort penalty")
	clickCheckBox(hiking.usingSkisCheckBox)
	s = current()
	c.False(s.skis, "clicking the skis checkbox again must clear the choice")
	c.True(s.skatesEnabled, "without skis, skates are allowed again")
	c.Equal("Made a successful Hiking roll (B200)", s.rollTitle, "on foot, the roll is against Hiking")
	clickCheckBox(hiking.usingSkatesCheckBox)
	s = current()
	c.True(s.skates, "clicking the skates checkbox must store the choice")
	c.False(s.skisEnabled, "skates must rule out skis")
	c.Equal("Made a successful Skating roll (B220)", s.rollTitle, "on skates, the roll is against Skating")
	clickCheckBox(hiking.roadsAreClearedCheckBox)
	c.True(current().roadsCleared, "clicking the roads checkbox must store the choice")
	captureScreen(t, c, screen, "hiking_calculator")

	// The Move is locked while the sheet supplies it and typed in once the source is Manual: Move 6 over a normal
	// eight-hour day on a dirt road in snow that has been cleared, on skates, after a successful Skating roll, is 4.5
	// miles an hour, with the 8th hour at half Move once the 10 FP have run low: 33.75 miles. With light encumbrance on
	// a hot day it costs 3 FP per hour, and extra effort at -2 adds 10% and 2 FP when the hiker stops (BX357): three
	// hours at 4.875 miles, then half Move from 1 FP left, and unconsciousness at -10 FP after the 7th hour, for 24.4
	// miles and 21 + 2 = 23 FP.
	c.False(current().moveEnabled, "the Move the sheet supplies must be locked")
	choosePopupItem(t, screen, wnd, hiking.source.popup, 0)
	s = current()
	c.Nil(s.sheet, "choosing Manual must drop the sheet")
	c.True(s.moveEnabled, "a typed-in Move must be unlocked")
	choosePopupItem(t, screen, wnd, intensityPopup, normalIntensity)
	screen.Do(func() { hiking.moveField.SetText("6") })
	c.Equal("34 miles", current().perDay, "the day's travel must follow the typed-in Move")
	choosePopupItem(t, screen, wnd, hiking.encumbrancePopup, 1)
	heatPopup, found := firstPanelOfType[*unison.PopupMenu[hikingHeatChoice]](hiking.content)
	if !found {
		t.Fatal("the calculator must offer a heat popup")
	}
	choosePopupItem(t, screen, wnd, heatPopup, 1)
	screen.Do(func() { hiking.hikingExtraEffortField.SetText("-2") })
	s = current()
	c.Equal(1, s.encumbrance, "choosing an encumbrance must store its index")
	c.Equal("23 FP lost by the end of the day (3 per hour, plus 2 for extra effort)", s.fpCost,
		"encumbrance, heat and extra effort must all add to the FP cost, up to the collapse")
	c.Equal("24 miles", s.perDay, "the day must end when the hiker falls unconscious")
	c.Equal([]string{"3", "4.9", "14.6", "-3", "1", "Very tired: Move halved"}, breakdownRow(3),
		"the 3rd hour leaves fewer than a third of the FP")
	c.Equal([]string{"4", "2.4", "17.1", "-3", "-2", "Exhausted: Will roll to go on, 1 HP per FP lost"}, breakdownRow(4),
		"the 4th hour is walked at half Move and uses up the FP")
	c.Equal([]string{"7", "2.4", "24.4", "-3", "-10", "Unconscious"}, breakdownRow(7),
		"the 7th hour reaches -FP")
	c.Equal([]string{"Stop", "", "", "-2", "-10", "Unconscious"}, breakdownRow(8),
		"the extra effort takes its toll at the stop, off HP once the FP are at their floor")
	c.Nil(breakdownRow(9), "the day must end at the collapse")
	c.True(slices.Contains(s.notes, "At -10 FP the hiker falls unconscious, after 7 hours, and the day ends there; any further FP cost comes off HP instead (B426)."),
		"the collapse must be explained")
	c.True(slices.Contains(s.notes, "The extra effort makes the Hiking roll a single Will-based Hiking roll at -2 for the +10% beyond the +20% a successful roll gives, and adds 2 FP to the loss when the hiker stops. A failure leaves the day at the +20% alone: 23 miles. A critical failure turns the whole loss, 23 FP, into HP of injury at the end of the day, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357)."),
		"extra effort must be explained, with what a failure leaves")
	captureScreen(t, c, screen, "hiking_calculator_fatigue")

	// BX427, BX55: an hour's rest after the 4th hour, with a decent meal, gives a Fit hiker 12 + 1 FP, capped at the 12
	// lost, so the second half of the day goes like the first: 34.1 miles instead of a collapse. The meal and the FP
	// from spells only count once there is a rest to take them in.
	var mealEnabled, extraBlank bool
	screen.Do(func() {
		mealEnabled = hiking.restMealCheckBox.Enabled()
		extraBlank = !hiking.restExtraField.Enabled() && hiking.restExtraField.DrawOverCallback != nil
	})
	c.False(mealEnabled, "without a rest there is no meal to take")
	c.True(extraBlank, "without a rest there is nothing to restore FP during")
	screen.Do(func() { hiking.restField.SetText("60") })
	clickCheckBox(hiking.restMealCheckBox)
	choosePopupItem(t, screen, wnd, hiking.fitnessPopup, 1)
	s = current()
	c.Equal("34 miles", s.perDay, "the rest must let the hiker finish the day")
	c.Equal("26 FP lost by the end of the day (3 per hour, plus 2 for extra effort)", s.fpCost,
		"the whole day is walked once the rest allows it")
	c.Equal([]string{"Rest", "", "", "+12", "10", ""}, breakdownRow(5), "the rest recovers what was lost, and no more")
	c.Equal([]string{"5", "4.9", "21.9", "-3", "7", ""}, breakdownRow(6), "the 5th hour is walked at full Move again")
	c.Equal([]string{"Stop", "", "", "-2", "-4", "Exhausted: Will roll to go on, 1 HP per FP lost"}, breakdownRow(10),
		"the day ends at the stop")
	c.True(slices.Contains(s.notes, "The rest of 60 minutes after 4 hours recovers 12 FP: 1 FP per 5 minutes, twice the usual rate, for being fit (B55), 1 for a decent meal, and never more than was lost (B427)."),
		"the rest must be explained")

	// BX55, BX248: a Very Fit hiker loses 1.5 FP an hour instead, and 3 FP from Lend Energy add to the rest, though the
	// rest still cannot exceed the 6 FP lost by then: 39 miles for 14 FP, and only the stop leaves the hiker very tired.
	choosePopupItem(t, screen, wnd, hiking.fitnessPopup, 2)
	screen.Do(func() { hiking.restExtraField.SetText("3") })
	s = current()
	c.Equal("39 miles", s.perDay, "a Very Fit hiker keeps full Move all day")
	c.Equal("14 FP lost by the end of the day (1.5 per hour, plus 2 for extra effort)", s.fpCost,
		"a Very Fit hiker loses FP at half the rate")
	c.Equal([]string{"Rest", "", "", "+6", "10", ""}, breakdownRow(5), "the rest recovers what was lost, and no more")
	c.Equal([]string{"Stop", "", "", "-2", "2", "Very tired: Move halved"}, breakdownRow(10),
		"the stop's extra effort leaves fewer than a third of the FP")
	c.True(slices.Contains(s.notes, "Being Very Fit, the hiker loses FP at half that rate (B55)."),
		"Very Fit must be explained")
	captureScreen(t, c, screen, "hiking_calculator_rest")

	closeEditorWithoutPrompt(t, screen, calc)
	closeEditorWithoutPrompt(t, screen, sheet)
}
