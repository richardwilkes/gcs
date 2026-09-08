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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

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

// TestCalculatorHikingControls drives the hiking section of the calculator inside a headless workspace the way a user
// would: it picks terrain, weather and intensity from the popups and clicks the checkboxes, checking that each choice
// is stored and that the controls depending on it are enabled, disabled or retitled to match.
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

	var terrainPopup, weatherPopup *unison.PopupMenu[terrainModifier]
	var intensityPopup *unison.PopupMenu[hikingIntensityHours]
	screen.Do(func() {
		if popups := panelsOfType[*unison.PopupMenu[terrainModifier]](calc.content); len(popups) == 2 {
			terrainPopup, weatherPopup = popups[0], popups[1]
		}
		intensityPopup, _ = firstPanelOfType[*unison.PopupMenu[hikingIntensityHours]](calc.content)
	})
	if terrainPopup == nil || weatherPopup == nil || intensityPopup == nil {
		t.Fatal("the calculator must offer terrain, weather and intensity popups")
	}
	// Every row and checkbox is a direct child of the content, indented beneath its section's header: three rows for
	// jumping and three for throwing, then the popups, four checkboxes, three fields and the results for hiking.
	var children, indented int
	screen.Do(func() {
		children = len(calc.content.Children())
		for _, child := range calc.content.Children() {
			// The section headers have either no border or one that only sets a top margin.
			if border := child.Border(); border != nil && border.Insets().Left == unison.StdHSpacing*2 {
				indented++
			}
		}
	})
	c.Equal(18, children, "the content must hold the three headers and their rows")
	c.Equal(15, indented, "every row and checkbox must be indented beneath its section header")

	type state struct {
		terrain, weather, intensity                                            int
		hours                                                                  fxp.Int
		hoursEnabled, roadsEnabled, skisEnabled, skatesEnabled, penaltyEnabled bool
		roadsCleared, skis, skates, roll                                       bool
		rollTitle                                                              string
	}
	current := func() state {
		var s state
		screen.Do(func() {
			s.terrain = calc.terrainIndex
			s.weather = calc.weatherIndex
			s.intensity = calc.hikingIntensityIndex
			s.hours = calc.hikingHours
			s.hoursEnabled = calc.hikingHoursField.Enabled()
			s.roadsEnabled = calc.roadsAreClearedCheckBox.Enabled()
			s.skisEnabled = calc.usingSkisCheckBox.Enabled()
			s.skatesEnabled = calc.usingSkatesCheckBox.Enabled()
			s.penaltyEnabled = calc.hikingExtraEffortField.Enabled()
			s.roadsCleared = calc.roadsAreCleared
			s.skis = calc.usingSkis
			s.skates = calc.usingSkates
			s.roll = calc.successfulHikingRoll
			// The page the skill is on sits beside the checkbox as a link; reading it with the title keeps one field.
			s.rollTitle = calc.successfulHikingRollCheckBox.Text.String() + " " + calc.hikingRollPageLabel.Title()
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

	// A fresh calculator offers the defaults, with the hours fixed by the intensity, no road to clear of snow, and no
	// extra effort without a successful roll.
	c.Equal(state{
		terrain:       dirtRoad,
		weather:       normalWeather,
		intensity:     normalIntensity,
		hours:         fxp.Eight,
		skisEnabled:   true,
		skatesEnabled: true,
		rollTitle:     "Made a successful Hiking roll (B200)",
	}, current(), "the calculator must start from the defaults")

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

	// Skis and skates exclude one another and change which skill the roll is against; a successful roll unlocks the
	// extra effort penalty.
	clickCheckBox := func(cb *unison.CheckBox) {
		screen.Do(func() { cb.ScrollIntoView() })
		screen.Click(screen.PanelCenter(cb))
	}
	clickCheckBox(calc.usingSkisCheckBox)
	s = current()
	c.True(s.skis, "clicking the skis checkbox must store the choice")
	c.False(s.skatesEnabled, "skis must rule out skates")
	c.Equal("Made a successful Skiing roll (B221)", s.rollTitle, "on skis, the roll is against Skiing")
	clickCheckBox(calc.successfulHikingRollCheckBox)
	s = current()
	c.True(s.roll, "clicking the roll checkbox must store the choice")
	c.True(s.penaltyEnabled, "a successful roll must unlock the extra effort penalty")
	clickCheckBox(calc.usingSkisCheckBox)
	s = current()
	c.False(s.skis, "clicking the skis checkbox again must clear the choice")
	c.True(s.skatesEnabled, "without skis, skates are allowed again")
	c.Equal("Made a successful Hiking roll (B200)", s.rollTitle, "on foot, the roll is against Hiking")
	clickCheckBox(calc.usingSkatesCheckBox)
	s = current()
	c.True(s.skates, "clicking the skates checkbox must store the choice")
	c.False(s.skisEnabled, "skates must rule out skis")
	c.Equal("Made a successful Skating roll (B220)", s.rollTitle, "on skates, the roll is against Skating")
	clickCheckBox(calc.roadsAreClearedCheckBox)
	c.True(current().roadsCleared, "clicking the roads checkbox must store the choice")

	closeEditorWithoutPrompt(t, screen, calc)
	closeEditorWithoutPrompt(t, screen, sheet)
}
