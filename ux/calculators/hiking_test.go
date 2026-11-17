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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// TestCalculatorHikingControls drives the hiking calculator inside a headless workspace the way a user would: it checks
// that the calculator starts from the preselected sheet's Move and FP, picks terrain, weather and intensity from the
// popups and clicks the checkboxes, checking that each choice is stored and reaches the day's travel and that the
// controls depending on it are enabled, disabled or retitled to match, and works a day's travel from numbers typed in.
// The model's TestHikingDay pins the hours of each of these days; the rows read here are the ones that tell the ways a
// row is shown apart.
func TestCalculatorHikingControls(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet := openNewCharacterSheet(t, screen)
	calc := openCalculator(t, screen)
	hiking := calc.hiking
	selectCalculatorTab(t, screen, calc, hiking)

	var terrainPopup, weatherPopup *unison.PopupMenu[calculator.TerrainModifier]
	var intensityPopup *unison.PopupMenu[calculator.HikingIntensity]
	screen.Do(func() {
		if popups := uxtest.PanelsOfType[*unison.PopupMenu[calculator.TerrainModifier]](hiking.content); len(popups) == 2 {
			terrainPopup, weatherPopup = popups[0], popups[1]
		}
		intensityPopup, _ = uxtest.FirstPanelOfType[*unison.PopupMenu[calculator.HikingIntensity]](hiking.content)
	})
	if terrainPopup == nil || weatherPopup == nil || intensityPopup == nil {
		t.Fatal("the calculator must offer terrain, weather and intensity popups")
	}
	// Every row and checkbox is a direct child of the content, indented beneath a header. The header, which comes first,
	// the subheaders and the results box are the rest, and are not indented.
	var rows int
	var misplaced []int
	screen.Do(func() {
		for i, child := range hiking.content.Children() {
			_, isSubheader := child.Self.(*unison.Label)
			isRow := i != 0 && !isSubheader && child != hiking.resultsBox
			if isRow {
				rows++
			}
			border := child.Border()
			if isRow != (border != nil && border.Insets().Left == unison.StdHSpacing*2) {
				misplaced = append(misplaced, i)
			}
		}
	})
	c.True(rows > 0, "the content must hold the calculator's rows")
	c.Equal(0, len(misplaced),
		"the rows and checkboxes, and nothing else, must be indented beneath their headers, but children %v are not",
		misplaced)

	type state struct {
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
			s.notes = slices.DeleteFunc(uxtest.LabelTexts(hiking.notes), func(text string) bool { return text == "•" })
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
	terrainNames := make([]string, len(hiking.terrain))
	for i, one := range hiking.terrain {
		terrainNames[i] = one.Name
	}
	weatherNames := make([]string, len(hiking.weather))
	for i, one := range hiking.weather {
		weatherNames[i] = one.Name
	}
	intensityNames := make([]string, len(hiking.intensity))
	for i, one := range hiking.intensity {
		intensityNames[i] = one.Name
	}
	dirtRoad := indexOf(terrainNames, "Road, Dirt")
	swamp := indexOf(terrainNames, "Swamp")
	normalWeather := indexOf(weatherNames, "Normal")
	sleet := indexOf(weatherNames, "Sleet")
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
			cells := uxtest.LabelTexts(hiking.breakdown)
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

	// Snow on a road lets it be cleared; a swamp cannot be.
	uxtest.ChoosePopupItem(t, screen, wnd, weatherPopup, snow)
	s := current()
	c.Equal(snow, s.weather, "choosing weather must store its index")
	c.True(s.roadsEnabled, "snow on a road must allow the road to be cleared")
	uxtest.ChoosePopupItem(t, screen, wnd, terrainPopup, swamp)
	s = current()
	c.Equal(swamp, s.terrain, "choosing terrain must store its index")
	c.False(s.roadsEnabled, "a swamp cannot be cleared of snow")
	uxtest.ChoosePopupItem(t, screen, wnd, terrainPopup, dirtRoad)
	c.True(current().roadsEnabled, "back on the road, it can be cleared again")

	// The hours are only editable for a custom intensity; any other sets them.
	uxtest.ChoosePopupItem(t, screen, wnd, intensityPopup, custom)
	s = current()
	c.Equal(custom, s.intensity, "choosing an intensity must store its index")
	c.True(s.hoursEnabled, "a custom intensity must free the hours field")
	c.Equal(fxp.Eight, s.hours, "switching to custom must keep the hours")
	uxtest.ChoosePopupItem(t, screen, wnd, intensityPopup, forcedMarch)
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

	// Skis and skates exclude one another and change which skill the roll is against; a successful roll unlocks the
	// extra effort penalty. Skis take the snow at the road's own pace, 3.125 miles an hour, so the march comes to
	// 7 × 3.125 + 9 × 1.5625 = 35.9 miles, and skates do the same for ice.
	clickCheckBox := func(cb *unison.CheckBox) {
		screen.Do(func() { cb.ScrollIntoView() })
		screen.Click(screen.PanelCenter(cb))
	}
	clickCheckBox(hiking.usingSkisCheckBox)
	s = current()
	c.True(s.skis, "clicking the skis checkbox must store the choice")
	c.False(s.skatesEnabled, "skis must rule out skates")
	c.Equal("Made a successful Skiing roll (B221)", s.rollTitle, "on skis, the roll is against Skiing")
	c.Equal("36 miles", s.perDay, "skis must spare the hiker what the snow costs")
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
	// With the roll's +20%, the road in sleet is skated at 3.75 miles an hour: 7 × 3.75 + 9 × 1.875 = 43.1 miles.
	uxtest.ChoosePopupItem(t, screen, wnd, weatherPopup, sleet)
	c.Equal("43 miles", current().perDay, "skates must spare the hiker what the ice costs")
	uxtest.ChoosePopupItem(t, screen, wnd, weatherPopup, snow)
	clickCheckBox(hiking.roadsAreClearedCheckBox)
	c.True(current().roadsCleared, "clicking the roads checkbox must store the choice")
	uxtest.CaptureScreen(t, c, screen, "hiking_calculator")

	// The Move is locked while the sheet supplies it, whatever the journey, and typed in once the source is Manual: Move
	// 6 over a normal eight-hour day on a dirt road in snow that has been cleared, on skates, after a successful Skating
	// roll, is 4.5 miles an hour, with the 8th hour at half Move once the 10 FP have run low: 33.75 miles. Two levels of
	// Enhanced Move triple that. With light encumbrance on a hot day it costs 3 FP per hour, and extra effort at -2 adds
	// 10% and 2 FP when the hiker stops (BX357): three hours at 4.875 miles, then half Move from 1 FP left, and
	// unconsciousness at -10 FP after the 7th hour, for 24.4 miles and 21 + 2 = 23 FP.
	c.False(current().moveEnabled, "the Move the sheet supplies must be locked")
	uxtest.ChoosePopupItem(t, screen, wnd, hiking.source.popup, 0)
	c.True(current().moveEnabled, "a typed-in Move must be unlocked")
	uxtest.ChoosePopupItem(t, screen, wnd, intensityPopup, normalIntensity)
	screen.Do(func() { hiking.moveField.SetText("6") })
	c.Equal("34 miles", current().perDay, "the day's travel must follow the typed-in Move")
	screen.Do(func() { hiking.enhancedMoveField.SetText("2") })
	c.Equal("101 miles", current().perDay, "each level of Enhanced Move must add the Move again")
	screen.Do(func() { hiking.enhancedMoveField.SetText("0") })
	uxtest.ChoosePopupItem(t, screen, wnd, hiking.encumbrancePopup, 1)
	heatPopup, found := uxtest.FirstPanelOfType[*unison.PopupMenu[calculator.HikingHeat]](hiking.content)
	if !found {
		t.Fatal("the calculator must offer a heat popup")
	}
	uxtest.ChoosePopupItem(t, screen, wnd, heatPopup, 1)
	screen.Do(func() { hiking.hikingExtraEffortField.SetText("-2") })
	s = current()
	c.Equal(1, s.encumbrance, "choosing an encumbrance must store its index")
	c.Equal("23 FP lost by the end of the day (3 per hour, plus 2 for extra effort)", s.fpCost,
		"encumbrance, heat and extra effort must all add to the FP cost, up to the collapse")
	c.Equal("24 miles", s.perDay, "the day must end when the hiker falls unconscious")
	c.Equal([]string{"7", "2.4", "24.4", "-3", "-10", "Unconscious"}, breakdownRow(7),
		"the 7th hour reaches -FP")
	c.Equal([]string{"Stop", "", "", "-2", "-10", "Unconscious"}, breakdownRow(8),
		"the extra effort takes its toll at the stop, off HP once the FP are at their floor")
	c.Nil(breakdownRow(9), "the day must end at the collapse")
	c.True(slices.Contains(s.notes, "At -10 FP the hiker falls unconscious, after 7 hours, and the day ends there; any further FP cost comes off HP instead (B426)."),
		"the collapse must be explained")
	c.True(slices.Contains(s.notes, "The extra effort makes the Hiking roll a single Will-based Hiking roll at -2 for the +10% beyond the +20% a successful roll gives, and adds 2 FP to the loss when the hiker stops. A failure leaves the day at the +20% alone: 23 miles. A critical failure turns the whole loss, 23 FP, into HP of injury at the end of the day, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357)."),
		"extra effort must be explained, with what a failure leaves")
	uxtest.CaptureScreen(t, c, screen, "hiking_calculator_fatigue")

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
	uxtest.ChoosePopupItem(t, screen, wnd, hiking.fitnessPopup, 1)
	s = current()
	c.Equal("34 miles", s.perDay, "the rest must let the hiker finish the day")
	c.Equal("26 FP lost by the end of the day (3 per hour, plus 2 for extra effort)", s.fpCost,
		"the whole day is walked once the rest allows it")
	c.Equal([]string{"Rest", "", "", "+12", "10", ""}, breakdownRow(5), "the rest recovers what was lost, and no more")
	c.Equal([]string{"5", "4.9", "21.9", "-3", "7", ""}, breakdownRow(6), "the 5th hour is walked at full Move again")
	c.True(slices.Contains(s.notes, "The rest of 60 minutes after 4 hours recovers 12 FP: 1 FP per 5 minutes, twice the usual rate, for being fit (B55), 1 for a decent meal, and never more than was lost (B427)."),
		"the rest must be explained")

	// BX55, BX248: a Very Fit hiker loses 1.5 FP an hour instead, and 3 FP from Lend Energy add to the rest, though the
	// rest still cannot exceed the 6 FP lost by then: 39 miles for 14 FP.
	uxtest.ChoosePopupItem(t, screen, wnd, hiking.fitnessPopup, 2)
	screen.Do(func() { hiking.restExtraField.SetText("3") })
	s = current()
	c.Equal("39 miles", s.perDay, "a Very Fit hiker keeps full Move all day")
	c.Equal("14 FP lost by the end of the day (1.5 per hour, plus 2 for extra effort)", s.fpCost,
		"a Very Fit hiker loses FP at half the rate")
	c.True(slices.Contains(s.notes, "Being Very Fit, the hiker loses FP at half that rate (B55)."),
		"Very Fit must be explained")
	uxtest.CaptureScreen(t, c, screen, "hiking_calculator_rest")

	// A rest of 5 minutes is too short to recover all 6 FP, so each of its parts shows: 1 FP for the 5 minutes, 1 for
	// the meal and 3 from Lend Energy. Recover Energy at skill 20+ turns the 5 minutes into 2 FP (BX248).
	screen.Do(func() { hiking.restField.SetText("5") })
	c.Equal([]string{"Rest", "", "", "+5", "9", ""}, breakdownRow(5), "the rest, the meal and the FP restored must add up")
	c.True(slices.Contains(current().notes, "The rest of 5 minutes after 4 hours recovers 5 FP: 1 FP per 5 minutes, twice the usual rate, for being fit (B55), 1 for a decent meal, 3 from Lend Energy, potions, etc., and never more than was lost (B427)."),
		"each part of what the rest recovers must be explained")
	uxtest.ChoosePopupItem(t, screen, wnd, hiking.recoverEnergyPopup, 2)
	c.Equal([]string{"Rest", "", "", "+6", "10", ""}, breakdownRow(5), "Recover Energy must speed the rest up")
	c.True(slices.Contains(current().notes, "The rest of 5 minutes after 4 hours recovers 6 FP: 1 FP per 2 minutes with Recover Energy (B248), 1 for a decent meal, 3 from Lend Energy, potions, etc., and never more than was lost (B427)."),
		"Recover Energy must be explained")

	uxtest.CloseEditorWithoutPrompt(t, screen, calc)
	uxtest.CloseEditorWithoutPrompt(t, screen, sheet)
}
