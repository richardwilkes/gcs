// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package calculator

import (
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/model/gurps/gurpstest"
	"github.com/richardwilkes/toolbox/v2/check"
)

// terrainNamed returns the entry of the given table with the given name.
func terrainNamed(c check.Checker, table []TerrainModifier, name string) TerrainModifier {
	c.Helper()
	i := slices.IndexFunc(table, func(t TerrainModifier) bool { return t.Name == name })
	if i < 0 {
		c.Fatalf("no entry named %q", name)
	}
	return table[i]
}

// testHike returns the hike the ux tests start from: a hiker with Move 5 and 10 FP on a dirt road in normal weather
// for a normal eight-hour day.
func testHike(c check.Checker) Hike {
	c.Helper()
	return Hike{
		Hiker: Hiker{
			Move:          5,
			FP:            fxp.Ten,
			Fitness:       HikingFitnessChoices()[0],
			RecoverEnergy: HikingRecoverEnergyChoices()[0],
		},
		Terrain: terrainNamed(c, TerrainChoices(), "Road, Dirt"),
		Weather: terrainNamed(c, WeatherChoices(), "Normal"),
		Heat:    HikingHeatChoices()[0],
		Hours:   fxp.Eight,
	}
}

// hourFacts is the part of a HikingHour a test pins.
type hourFacts struct {
	kind      HikingHourKind
	condition HikerCondition
	distance  string
	total     string
	fpChange  string
	fpLeft    string
}

// facts reduces a row to what the tests pin, with the distances and FP as decimal strings.
func facts(hour HikingHour) hourFacts {
	return hourFacts{
		kind:      hour.Kind,
		condition: hour.Condition,
		distance:  hour.Distance.String(),
		total:     hour.Total.String(),
		fpChange:  hour.FPChange.String(),
		fpLeft:    hour.FPLeft.String(),
	}
}

// TestHikingDay verifies a day's travel worked out hour by hour (BX351, BX426), through the scenarios the ux tests
// drive: a normal day, a forced march in snow, a day on skates after a successful roll, one that ends in collapse,
// and the rest and fitness that stave the collapse off (BX427, BX55).
func TestHikingDay(t *testing.T) {
	c := check.New(t)

	// A normal day: 3.125 miles an hour for 7 hours, then half Move once fewer than a third of the FP are left.
	hike := testHike(c)
	day := hike.Day()
	c.Equal(fxp.FromStringForced("23.4375"), day.Distance, "the default day covers 23.4375 miles")
	c.Equal(fxp.Eight, day.FPCost, "and costs 8 FP")
	c.Equal(fxp.Seven, day.TiredAfter, "the hiker tires after the 7th hour")
	c.Equal(fxp.Int(0), day.OutAfter, "but does not run out")
	c.Equal(fxp.Int(0), day.UnconsciousAfter, "nor collapse")
	c.Equal(8, len(day.Hours), "one row per hour")
	c.Equal(hourFacts{kind: MarchHour, distance: "3.125", total: "3.125", fpChange: "-1", fpLeft: "9"}, facts(day.Hours[0]),
		"the first hour is walked at full Move")
	c.Equal(fxp.One, day.Hours[0].Elapsed, "an hour in")
	c.Equal(hourFacts{kind: MarchHour, condition: VeryTiredHiker, distance: "3.125", total: "21.875", fpChange: "-1", fpLeft: "3"},
		facts(day.Hours[6]), "the 7th hour leaves fewer than a third of the FP")
	c.Equal(hourFacts{kind: MarchHour, condition: VeryTiredHiker, distance: "1.5625", total: "23.4375", fpChange: "-1", fpLeft: "2"},
		facts(day.Hours[7]), "so the 8th is walked at half Move")

	metric := hike
	metric.Metric = true
	c.Equal(fxp.FromStringForced("37.125"), metric.Day().Distance, "a metric day is reported in GURPS kilometers")

	// A forced march on an uncleared road in snow: 16 hours at 1.5625 miles, the last 9 at half Move.
	forced := hike
	forced.Hours = fxp.Sixteen
	forced.Weather = terrainNamed(c, WeatherChoices(), "Snow")
	day = forced.Day()
	c.Equal(fxp.FromInteger(18), day.Distance.Round(), "7 x 1.5625 + 9 x 0.78125 is 18 miles, give or take the fixed-point truncation")
	c.Equal(fxp.Sixteen, day.FPCost, "16 hours cost 16 FP")
	c.Equal(fxp.Ten, day.OutAfter, "the FP run out after the 10th hour")
	c.Equal(fxp.Int(0), day.UnconsciousAfter, "but the day ends at -6 FP, before the collapse")
	c.Equal(16, len(day.Hours), "every hour is walked")
	c.Equal(ExhaustedHiker, day.Hours[9].Condition, "the 10th hour leaves the hiker exhausted")
	c.Equal(fxp.FromStringForced("13.2811"), day.Hours[9].Total, "and 13.3 miles along")

	// Skates on a road cleared of snow, after a successful Skating roll, with Move 6: 4.5 miles an hour.
	skating := hike
	skating.Hiker.Move = 6
	skating.Weather = forced.Weather
	skating.UsingSkates = true
	skating.RoadsAreCleared = true
	skating.SuccessfulRoll = true
	day = skating.Day()
	c.Equal(fxp.FromStringForced("33.75"), day.Distance, "7 x 4.5 + 2.25")
	c.False(skating.ExtraEffort(), "a successful roll alone is not extra effort")

	// Light encumbrance on a hot day with extra effort at -2: 3 FP an hour and 4.875 miles, until the collapse.
	collapse := skating
	collapse.Hiker.Encumbrance = encumbrance.Light
	collapse.Heat = HikingHeatChoices()[1]
	collapse.ExtraEffortPenalty = -2
	c.True(collapse.ExtraEffort(), "a penalty on a successful roll is extra effort")
	c.Equal(fxp.Three, collapse.FPPerHour(), "1, plus 1 for the encumbrance, plus 1 for the heat")
	day = collapse.Day()
	c.Equal(fxp.FromStringForced("24.375"), day.Distance, "3 x 4.875 + 4 x 2.4375 before the collapse")
	c.Equal(fxp.FromInteger(23), day.FPCost, "7 x 3, plus 2 when the hiker stops")
	c.Equal(fxp.Three, day.TiredAfter, "the 3rd hour leaves fewer than a third of the FP")
	c.Equal(fxp.Four, day.OutAfter, "the 4th uses them up")
	c.Equal(fxp.Seven, day.UnconsciousAfter, "the 7th reaches -FP")
	c.Equal(8, len(day.Hours), "7 hours and the stop")
	c.Equal(hourFacts{kind: MarchHour, condition: VeryTiredHiker, distance: "4.875", total: "14.625", fpChange: "-3", fpLeft: "1"},
		facts(day.Hours[2]), "the 3rd hour")
	c.Equal(hourFacts{kind: MarchHour, condition: ExhaustedHiker, distance: "2.4375", total: "17.0625", fpChange: "-3", fpLeft: "-2"},
		facts(day.Hours[3]), "the 4th hour, at half Move")
	c.Equal(hourFacts{kind: MarchHour, condition: UnconsciousHiker, distance: "2.4375", total: "24.375", fpChange: "-3", fpLeft: "-10"},
		facts(day.Hours[6]), "the 7th hour, which ends the day")
	c.Equal(hourFacts{kind: StopHour, condition: UnconsciousHiker, distance: "0", total: "0", fpChange: "-2", fpLeft: "-10"},
		facts(day.Hours[7]), "the stop, whose extra effort can take no more")
	c.Equal(fxp.Seven, day.Hours[7].Elapsed, "the stop comes after the 7th hour")

	// An hour's rest after the 4th hour with a decent meal gives a Fit hiker 12 + 1 FP, capped at the 12 lost.
	rested := collapse
	rested.RestMinutes = 60
	rested.RestMeal = true
	rested.Hiker.Fitness = HikingFitnessChoices()[1]
	c.Equal(fxp.Four, rested.RestAfter(), "the rest comes halfway through the 8 hours")
	day = rested.Day()
	c.Equal(fxp.FromStringForced("34.125"), day.Distance, "the second half of the day goes like the first")
	c.Equal(fxp.FromInteger(26), day.FPCost, "8 x 3, plus 2 when the hiker stops")
	c.Equal(fxp.Twelve, day.RestRecovered, "the rest recovers what was lost, and no more")
	c.Equal(fxp.Int(0), day.UnconsciousAfter, "there is no collapse")
	c.Equal(10, len(day.Hours), "8 hours, the rest and the stop")
	c.Equal(hourFacts{kind: RestHour, distance: "0", total: "0", fpChange: "12", fpLeft: "10"}, facts(day.Hours[4]), "the rest")
	c.Equal(hourFacts{kind: MarchHour, distance: "4.875", total: "21.9375", fpChange: "-3", fpLeft: "7"}, facts(day.Hours[5]),
		"the 5th hour is walked at full Move again")
	c.Equal(hourFacts{kind: StopHour, condition: ExhaustedHiker, distance: "0", total: "0", fpChange: "-2", fpLeft: "-4"},
		facts(day.Hours[9]), "the day ends at the stop")

	// A Very Fit hiker loses 1.5 FP an hour, and 3 FP from Lend Energy add to the rest, which still cannot exceed the 6
	// FP lost by then.
	veryFit := rested
	veryFit.Hiker.Fitness = HikingFitnessChoices()[2]
	veryFit.RestExtraFP = 3
	c.Equal(fxp.OneAndAHalf, veryFit.FPPerHour(), "Very Fit halves the cost")
	day = veryFit.Day()
	c.Equal(fxp.FromInteger(39), day.Distance, "full Move all day: 8 x 4.875")
	c.Equal(fxp.FromInteger(14), day.FPCost, "8 x 1.5, plus 2 when the hiker stops")
	c.Equal(fxp.Six, day.RestRecovered, "the rest recovers the 6 FP lost, and no more")
	c.Equal(hourFacts{kind: StopHour, condition: VeryTiredHiker, distance: "0", total: "0", fpChange: "-2", fpLeft: "2"},
		facts(day.Hours[9]), "only the stop leaves the hiker very tired")

	// Without FP to go on, the hiker keeps full pace all day.
	unknown := hike
	unknown.Hiker.FP = 0
	day = unknown.Day()
	c.Equal(fxp.FromInteger(25), day.Distance, "8 x 3.125")
	c.Equal(fxp.Int(0), day.TiredAfter, "nothing is known about tiring")
	for i, hour := range day.Hours {
		c.Equal(FreshHiker, hour.Condition, "hour %d has no condition", i)
		c.Equal(fxp.Int(0), hour.FPLeft, "hour %d has no FP left to report", i)
	}
}

// TestHikingDistanceModifiers verifies what the ground, the weather and the gear do to an hour's travel (BX351),
// starting from the 3.125 miles an hour of Move 5.
func TestHikingDistanceModifiers(t *testing.T) {
	c := check.New(t)
	hike := testHike(c)
	terrains := TerrainChoices()
	weathers := WeatherChoices()
	hourly := func(terrain, weather string, adjust func(h *Hike)) string {
		h := hike
		h.Terrain = terrainNamed(c, terrains, terrain)
		h.Weather = terrainNamed(c, weathers, weather)
		if adjust != nil {
			adjust(&h)
		}
		return h.distanceForHours(fxp.One).String()
	}
	none := func(_ *Hike) {}
	c.Equal("3.125", hourly("Road, Dirt", "Normal", none), "a dirt road in fair weather is the full pace")
	c.Equal("0.625", hourly("Road, Dirt", "Rain", none), "rain turns a dirt road to mud")
	c.Equal("3.125", hourly("Road, Paved", "Rain", none), "and takes a paved road down from its 1.25 to a plain one")
	c.Equal("3.125", hourly("Road, Cobblestone", "Rain", none), "while leaving a cobblestone road alone")
	c.Equal("0.7812", hourly("Broken Ground", "Rain", none), "off the road, rain halves the ground's own modifier")
	c.Equal("1.5625", hourly("Road, Dirt", "Snow", none), "snow halves an uncleared road")
	c.Equal("3.125", hourly("Road, Dirt", "Snow", func(h *Hike) { h.RoadsAreCleared = true }), "unless it has been cleared")
	c.Equal("3.125", hourly("Road, Dirt", "Snow", func(h *Hike) { h.UsingSkis = true }), "or the hiker is on skis")
	c.Equal("1.5625", hourly("Road, Paved", "Snow", none), "snow also takes a paved road down to a plain one")
	c.Equal("0.1562", hourly("Jungle", "Snow, Heavy", none), "heavy snow quarters the ground's own modifier")
	c.Equal("1.5625", hourly("Road, Dirt", "Sleet", none), "sleet halves an uncleared road")
	c.Equal("3.125", hourly("Road, Dirt", "Sleet", func(h *Hike) { h.UsingSkates = true }), "unless the hiker is on skates")
	c.Equal("1.5625", hourly("Frozen Lake", "Normal", none), "a frozen lake is half pace on foot")
	c.Equal("3.9062", hourly("Frozen Lake", "Normal", func(h *Hike) { h.UsingSkates = true }), "and 1.25 on skates")
	c.Equal("3.125", hourly("Deep Snow", "Normal", func(h *Hike) { h.UsingSkis = true }), "deep snow is full pace on skis")
	c.Equal("0.625", hourly("Deep Snow", "Normal", none), "and a fifth of it on foot")
	c.Equal("3.75", hourly("Road, Dirt", "Normal", func(h *Hike) { h.SuccessfulRoll = true }), "a successful roll adds 20%")
	c.Equal("4.0625", hourly("Road, Dirt", "Normal", func(h *Hike) { h.SuccessfulRoll = true; h.ExtraEffortPenalty = -2 }),
		"and extra effort at -2 adds 10% more")
	c.Equal("3.125", hourly("Road, Dirt", "Normal", func(h *Hike) { h.ExtraEffortPenalty = -2 }),
		"extra effort does nothing without the roll")
	c.Equal("6.25", hourly("Road, Dirt", "Normal", func(h *Hike) { h.Hiker.EnhancedMove = fxp.One }),
		"a level of Enhanced Move doubles the pace")
}

// TestHikingRest verifies when the rest comes and how fast it recovers FP (BX427, BX55, BX248).
func TestHikingRest(t *testing.T) {
	c := check.New(t)
	hike := testHike(c)
	c.Equal(fxp.Int(0), hike.RestAfter(), "without minutes of rest there is no rest")
	hike.RestMinutes = 30
	c.Equal(fxp.Four, hike.RestAfter(), "the rest comes halfway through 8 hours")
	hike.Hours = fxp.Seven
	c.Equal(fxp.Three, hike.RestAfter(), "at the end of a whole hour")
	hike.Hours = fxp.Two
	c.Equal(fxp.One, hike.RestAfter(), "and after the first hour of two")
	hike.Hours = fxp.One
	c.Equal(fxp.Int(0), hike.RestAfter(), "but not when nothing would follow it")

	hiker := hike.Hiker
	c.Equal(10, hiker.RestMinutesPerFP(), "quiet rest recovers 1 FP per 10 minutes")
	hiker.RecoverEnergy = HikingRecoverEnergyChoices()[1]
	c.Equal(5, hiker.RestMinutesPerFP(), "Recover Energy at 15+ makes it 5")
	hiker.Fitness = HikingFitnessChoices()[1]
	c.Equal(5, hiker.RestMinutesPerFP(), "being Fit makes it 5 as well, and the two do not stack")
	hiker.RecoverEnergy = HikingRecoverEnergyChoices()[2]
	c.Equal(2, hiker.RestMinutesPerFP(), "Recover Energy at 20+ makes it 2")
	c.True(hike.RoadsCanBeCleared() == false, "a road in fair weather has nothing to clear")
	hike.Weather = terrainNamed(c, WeatherChoices(), "Sleet")
	c.True(hike.RoadsCanBeCleared(), "a road under ice can be cleared")
	hike.Terrain = terrainNamed(c, TerrainChoices(), "Swamp")
	c.False(hike.RoadsCanBeCleared(), "a swamp cannot")
}

// TestHikingChoices verifies the hiking tables: that each has exactly one default where the calculator needs one, and
// the facts the entries carry.
func TestHikingChoices(t *testing.T) {
	c := check.New(t)
	terrains := TerrainChoices()
	checkNamed(c, "terrain", terrains, func(t TerrainModifier) string { return t.Name })
	weathers := WeatherChoices()
	checkNamed(c, "weather", weathers, func(t TerrainModifier) string { return t.Name })
	intensities := HikingIntensityChoices()
	checkNamed(c, "intensity", intensities, func(h HikingIntensity) string { return h.Name })
	defaults := func(terrain []TerrainModifier) int {
		n := 0
		for _, t := range terrain {
			if t.Default {
				n++
			}
		}
		return n
	}
	c.Equal(1, defaults(terrains), "exactly one terrain is the default")
	c.Equal(1, defaults(weathers), "exactly one weather is the default")
	c.Equal(1, len(slices.DeleteFunc(slices.Clone(intensities), func(h HikingIntensity) bool { return !h.Default })),
		"exactly one intensity is the default")
	c.Equal(1, len(slices.DeleteFunc(slices.Clone(intensities), func(h HikingIntensity) bool { return !h.IsCustom })),
		"exactly one intensity is custom")
	for i, t := range terrains {
		c.False(t.IsRain, "terrain %d is not weather", i)
		c.True(t.Modifier > 0, "terrain %d slows or speeds the hike", i)
	}
	for i, w := range weathers {
		c.False(w.IsRoad, "weather %d is not ground", i)
		c.True(w.Modifier > 0, "weather %d slows or speeds the hike", i)
	}

	heat := HikingHeatChoices()
	checkNamed(c, "heat", heat, func(h HikingHeat) string { return h.Name })
	c.Equal([]int{0, 1, 2}, []int{heat[0].FP, heat[1].FP, heat[2].FP}, "heat adds 0, 1 or 2 FP an hour")

	fitness := HikingFitnessChoices()
	checkNamed(c, "fitness", fitness, func(h HikingFitness) string { return h.Name })
	c.Equal([]HikingFitness{{Name: fitness[0].Name}, {Name: fitness[1].Name, Fit: true}, {Name: fitness[2].Name, Fit: true, VeryFit: true}},
		fitness, "Very Fit is also Fit")

	recoverEnergy := HikingRecoverEnergyChoices()
	checkNamed(c, "Recover Energy", recoverEnergy, func(h HikingRecoverEnergy) string { return h.Name })
	c.Equal([]int{0, 5, 2}, []int{recoverEnergy[0].MinutesPerFP, recoverEnergy[1].MinutesPerFP, recoverEnergy[2].MinutesPerFP},
		"Recover Energy gives 1 FP per 5 minutes at 15+ and per 2 at 20+")
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
		days, ok := HikingTimeInDays(one.distanceToCover, one.distancePerDay)
		c.Equal(one.wantOK, ok, one.name)
		c.Equal(one.wantDays, days, one.name)
	}
}

// TestHikerStatsFromEntity verifies what a sheet supplies about a hiker: Move and encumbrance, Enhanced Move, FP, the
// fitness traits and the Recover Energy spell.
func TestHikerStatsFromEntity(t *testing.T) {
	c := check.New(t)
	e := gurps.NewEntity()
	s := HikerStatsFromEntity(e)
	c.Equal(encumbrance.No, s.Encumbrance, "an empty sheet carries nothing")
	c.Equal(e.Move(encumbrance.No), s.Move, "the Move is the sheet's at that encumbrance")
	c.Equal(fxp.Int(0), s.EnhancedMove, "no Enhanced Move is 0")
	c.True(s.HasFP, "the default sheet defines FP")
	c.Equal(e.Attributes.Maximum(gurps.FatiguePointsID), s.FP, "the FP is the maximum")
	c.Equal(0, s.FitnessIndex, "an average hiker")
	c.Equal(0, s.RecoverEnergyIndex, "with no Recover Energy")

	addLeveledTrait(e, "Enhanced Move (Ground)", fxp.One)
	gurpstest.AddTraitWithFeatures(e, "Fit")
	spell := addSpell(e, "Recover Energy", fxp.FromInteger(28))
	e.Recalculate()
	s = HikerStatsFromEntity(e)
	c.Equal(fxp.One, s.EnhancedMove, "the trait's levels are read")
	c.Equal(1, s.FitnessIndex, "Fit is the second choice")
	level := spell.CalculateLevel().Level.AsInteger[int]()
	c.True(level >= 15 && level < 20, "28 points must put the spell at 15 or better but short of 20, not %d", level)
	c.Equal(1, s.RecoverEnergyIndex, "Recover Energy at 15+ is the second choice")

	gurpstest.AddTraitWithFeatures(e, "Very Fit")
	spell.Points = fxp.FromInteger(44)
	e.Recalculate()
	s = HikerStatsFromEntity(e)
	c.Equal(2, s.FitnessIndex, "Very Fit is the third choice, and wins over Fit")
	level = spell.CalculateLevel().Level.AsInteger[int]()
	c.True(level >= 20, "44 points must put the spell at 20 or better, not %d", level)
	c.Equal(2, s.RecoverEnergyIndex, "Recover Energy at 20+ is the third choice")
}
