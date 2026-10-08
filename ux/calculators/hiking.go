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
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var _ calculatorTab = &hikingCalculator{}

// hikingCalculator works out how far a character covers in a day of travel over the given ground (BX351, HT55), and
// what the day costs in FP (BX426). The character's numbers are either typed in or taken from any open character
// sheet, in which case the fields holding them are locked and refreshed whenever the sheet changes. The journey itself
// is always typed in.
type hikingCalculator struct {
	calculatorContent
	source                       sheetSourcePicker
	moveField                    *ux.IntegerField
	enhancedMoveField            *ux.DecimalField
	encumbrancePopup             *unison.PopupMenu[encumbrance.Level]
	fpField                      *ux.DecimalField
	fitnessPopup                 *unison.PopupMenu[calculator.HikingFitness]
	recoverEnergyPopup           *unison.PopupMenu[calculator.HikingRecoverEnergy]
	restField                    *ux.IntegerField
	restMealCheckBox             *unison.CheckBox
	restExtraField               *ux.IntegerField
	breakdown                    *unison.Panel
	notes                        *unison.Panel
	hikingResult                 *unison.Label
	hikingDistanceLabel          *ux.TextLabel
	hikingTimeLabel              *unison.Label
	fpResult                     *unison.Label
	fpLabel                      *ux.TextLabel
	hikingHoursField             *ux.DecimalField
	hikingExtraEffortField       *ux.IntegerField
	roadsAreClearedCheckBox      *unison.CheckBox
	usingSkisCheckBox            *unison.CheckBox
	usingSkatesCheckBox          *unison.CheckBox
	successfulHikingRollCheckBox *unison.CheckBox
	hikingRollPageLabel          *ux.TextLabel
	terrain                      []calculator.TerrainModifier
	weather                      []calculator.TerrainModifier
	heat                         []calculator.HikingHeat
	intensity                    []calculator.HikingIntensity
	fitness                      []calculator.HikingFitness
	recoverEnergy                []calculator.HikingRecoverEnergy
	enhancedMove                 fxp.Int
	fp                           fxp.Int
	hikingHours                  fxp.Int
	hikingDistance               fxp.Int
	move                         int
	encumbranceIndex             int
	fitnessIndex                 int
	recoverEnergyIndex           int
	restMinutes                  int
	restExtraFP                  int
	hikingExtraEffortPenalty     int
	terrainIndex                 int
	weatherIndex                 int
	heatIndex                    int
	hikingIntensityIndex         int
	usingSkis                    bool
	usingSkates                  bool
	roadsAreCleared              bool
	successfulHikingRoll         bool
	restMeal                     bool
	updating                     bool
}

func newHikingCalculator() *hikingCalculator {
	h := &hikingCalculator{
		terrain:       calculator.TerrainChoices(),
		weather:       calculator.WeatherChoices(),
		heat:          calculator.HikingHeatChoices(),
		intensity:     calculator.HikingIntensityChoices(),
		fitness:       calculator.HikingFitnessChoices(),
		recoverEnergy: calculator.HikingRecoverEnergyChoices(),
		move:          5,
		fp:            fxp.Ten,
		hikingHours:   fxp.Eight,
	}
	h.terrainIndex = slices.IndexFunc(h.terrain, func(t calculator.TerrainModifier) bool { return t.Default })
	h.weatherIndex = slices.IndexFunc(h.weather, func(t calculator.TerrainModifier) bool { return t.Default })
	h.hikingIntensityIndex = slices.IndexFunc(h.intensity, func(t calculator.HikingIntensity) bool { return t.Default })
	h.createContent()
	return h
}

// title implements calculatorTab.
func (h *hikingCalculator) title() string {
	return i18n.Text("Hiking")
}

// panel implements calculatorTab.
func (h *hikingCalculator) panel() *unison.Panel {
	return h.content
}

// preselect implements calculatorTab.
func (h *hikingCalculator) preselect(sheet *ux.Sheet) {
	h.source.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (h *hikingCalculator) sheetChanged(sheet *ux.Sheet) {
	if h.source.sheet == sheet {
		h.changed()
	}
}

func (h *hikingCalculator) createContent() {
	h.initCalculatorContent()
	h.content.AddChild(h.createHeader(i18n.Text("Hiking"),
		[]linkSpec{
			{pageRef: "BX351", highlight: "Hiking"},
			{pageRef: "HT55", highlight: "Hiking"},
		}, 0))

	h.addSubheader(i18n.Text("Hiker"))
	h.source.selected = h.selectSheet
	h.source.refreshed = h.pullFromSheet
	h.source.changed = h.changed
	h.source.addRow(&h.calculatorContent, i18n.Text("Source:"))
	h.moveField = sameWidth(ux.NewIntegerField(nil, "", i18n.Text("Move"),
		func() int { return h.move },
		func(v int) {
			h.move = v
			h.changed()
		},
		0, 10000, false, false))
	h.addFieldRow(h.moveField, i18n.Text("Move, with encumbrance"))
	h.enhancedMoveField = sameWidth(ux.NewDecimalField(nil, "", i18n.Text("Enhanced Move"),
		func() fxp.Int { return h.enhancedMove },
		func(v fxp.Int) {
			h.enhancedMove = v
			h.changed()
		},
		0, fxp.Max, false, false))
	h.addFieldRow(h.enhancedMoveField, i18n.Text("levels of Enhanced Move (Ground)"))
	row := h.addRow(2)
	addPlainLabel(row, i18n.Text("Encumbrance:"))
	h.encumbrancePopup = addIndexPopup(row, encumbrance.Levels, &h.encumbranceIndex, h.changed)
	h.fpField = sameWidth(ux.NewDecimalField(nil, "", i18n.Text("Fatigue Points"),
		func() fxp.Int { return h.fp },
		func(v fxp.Int) {
			h.fp = v
			h.changed()
		},
		0, fxp.Max, false, false))
	h.addFieldRow(h.fpField, i18n.Text("FP"))
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Fitness:"))
	h.fitnessPopup = addIndexPopup(row, h.fitness, &h.fitnessIndex, h.changed)
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Recover Energy:"))
	h.recoverEnergyPopup = addIndexPopup(row, h.recoverEnergy, &h.recoverEnergyIndex, h.changed)

	h.addSubheader(i18n.Text("Journey"))
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Terrain:"))
	addIndexPopup(row, h.terrain, &h.terrainIndex, h.changed)
	addPlainLabel(row, i18n.Text("Weather:"))
	addIndexPopup(row, h.weather, &h.weatherIndex, h.changed)
	addPlainLabel(row, i18n.Text("Heat:"))
	addIndexPopup(row, h.heat, &h.heatIndex, h.changed)
	addPlainLabel(row, i18n.Text("Intensity:"))
	addIndexPopup(row, h.intensity, &h.hikingIntensityIndex, h.changed)

	h.roadsAreClearedCheckBox = h.addCheckBox(i18n.Text("Roads are cleared"), &h.roadsAreCleared, h.changed)
	h.usingSkisCheckBox = h.addCheckBox(i18n.Text("Using skis"), &h.usingSkis, h.changed)
	h.usingSkatesCheckBox = h.addCheckBox(i18n.Text("Using skates"), &h.usingSkates, h.changed)
	// The title names the skill the roll is against, which depends on the mode of travel, so adjustControls sets it,
	// along with the page the skill is on, which sits beside the checkbox as a link.
	row = h.addRow(2)
	h.successfulHikingRollCheckBox = newCheckBox("", &h.successfulHikingRoll, h.changed)
	row.AddChild(h.successfulHikingRollCheckBox)
	h.hikingRollPageLabel = addPlainLabel(row, "")

	h.hikingHoursField = sameWidth(ux.NewDecimalField(nil, "", i18n.Text("Traveling Hours per Day"),
		func() fxp.Int { return h.hikingHours },
		func(v fxp.Int) {
			h.hikingHours = v
			h.changed()
		},
		0, fxp.TwentyFour, false, false))
	h.addFieldRow(h.hikingHoursField, i18n.Text("hours of hiking per day"))
	h.restField = sameWidth(ux.NewIntegerField(nil, "", i18n.Text("Rest"),
		func() int { return h.restMinutes },
		func(v int) {
			h.restMinutes = v
			h.changed()
		},
		0, 24*60, false, false))
	h.addFieldRow(h.restField, i18n.Text("minutes of rest halfway through the day (0 for none)"))
	h.restMealCheckBox = h.addCheckBox(i18n.Text("Eats a decent meal while resting (+1 FP)"), &h.restMeal, h.changed)
	h.restExtraField = sameWidth(ux.NewIntegerField(nil, "", i18n.Text("FP Restored While Resting"),
		func() int { return h.restExtraFP },
		func(v int) {
			h.restExtraFP = v
			h.changed()
		},
		0, 1000, false, false))
	h.addFieldRow(h.restExtraField, i18n.Text("FP restored by Lend Energy, potions, etc. while resting"))
	h.hikingExtraEffortField = sameWidth(ux.NewIntegerField(nil, "", i18n.Text("Hiking Extra Effort Penalty"),
		func() int { return h.hikingExtraEffortPenalty },
		func(v int) {
			h.hikingExtraEffortPenalty = v
			h.changed()
		},
		-100, 0, false, false))
	h.addFieldRow(h.hikingExtraEffortField, i18n.Text("penalty taken for extra effort (+5% distance per -1)"))
	h.hikingDistanceLabel = h.addFieldRow(sameWidth(ux.NewDecimalField(nil, "", i18n.Text("Distance to Cover"),
		func() fxp.Int { return h.hikingDistance },
		func(v fxp.Int) {
			h.hikingDistance = v
			h.changed()
		},
		0, fxp.Max, false, false)), "")

	box := h.addResultsBox()
	row = box.addRow(2)
	h.hikingResult = addResultLabel(row)
	addPlainLabel(row, i18n.Text("per day"))
	h.hikingTimeLabel = addResultLabel(row)
	addPlainLabel(row, i18n.Text("to hike"))
	h.fpResult = addResultLabel(row)
	h.fpLabel = addPlainLabel(row, "")
	box.addSubheader(i18n.Text("Hour by hour"))
	h.breakdown = box.addRow(6)
	h.breakdown.SetLayout(&unison.FlexLayout{Columns: 6, HSpacing: unison.StdHSpacing * 2, VSpacing: unison.StdVSpacing})
	h.notes = box.addNotesGroup()
}

// selectSheet reads the hiker's numbers from the sheet the picker has just made its source, and does nothing at all
// when there is none.
func (h *hikingCalculator) selectSheet(sheet *ux.Sheet) {
	if sheet != nil {
		h.pullFromSheet()
	}
}

// pullFromSheet reads the hiker's numbers from its sheet. Each backing value is assigned before its field is synced, so
// that the setter the sync may run sees nothing new and does not start another round of updates.
func (h *hikingCalculator) pullFromSheet() {
	stats := calculator.HikerStatsFromEntity(h.source.entity())
	h.encumbranceIndex = int(stats.Encumbrance)
	h.move = stats.Move
	h.enhancedMove = stats.EnhancedMove
	if stats.HasFP {
		h.fp = stats.FP
	}
	h.fitnessIndex = stats.FitnessIndex
	h.recoverEnergyIndex = stats.RecoverEnergyIndex
	h.moveField.Sync()
	h.enhancedMoveField.Sync()
	h.encumbrancePopup.SelectIndex(h.encumbranceIndex)
	h.fpField.Sync()
	h.fitnessPopup.SelectIndex(h.fitnessIndex)
	h.recoverEnergyPopup.SelectIndex(h.recoverEnergyIndex)
}

// changed implements calculatorTab.
func (h *hikingCalculator) changed() {
	if h.updating {
		return
	}
	h.updating = true
	defer func() { h.updating = false }()
	h.source.refresh()
	h.adjustControls()
	h.updateResults()
}

// adjustControls locks the fields a sheet supplies, then enables, disables and retitles the rest to match the current
// selections: skis and skates exclude one another and decide which skill the roll is against, the hours are only
// editable for a custom intensity, roads can only be cleared of snow or ice, and extra effort needs a successful roll.
func (h *hikingCalculator) adjustControls() {
	manual := h.source.sheet == nil
	h.moveField.SetEnabled(manual)
	h.enhancedMoveField.SetEnabled(manual)
	h.encumbrancePopup.SetEnabled(manual)
	h.fpField.SetEnabled(manual)
	h.fitnessPopup.SetEnabled(manual)
	h.recoverEnergyPopup.SetEnabled(manual)

	switch {
	case h.usingSkis:
		h.successfulHikingRollCheckBox.SetTitle(i18n.Text("Made a successful Skiing roll"))
		h.hikingRollPageLabel.SetTitle("(B221)")
		h.usingSkatesCheckBox.SetEnabled(false)
	case h.usingSkates:
		h.successfulHikingRollCheckBox.SetTitle(i18n.Text("Made a successful Skating roll"))
		h.hikingRollPageLabel.SetTitle("(B220)")
		h.usingSkisCheckBox.SetEnabled(false)
	default:
		h.successfulHikingRollCheckBox.SetTitle(i18n.Text("Made a successful Hiking roll"))
		h.hikingRollPageLabel.SetTitle("(B200)")
		h.usingSkatesCheckBox.SetEnabled(true)
		h.usingSkisCheckBox.SetEnabled(true)
	}

	i := h.intensity[h.hikingIntensityIndex]
	h.hikingHoursField.SetEnabled(true)
	if !i.IsCustom {
		h.hikingHours = i.HoursHiking
		h.hikingHoursField.Sync()
		h.hikingHoursField.SetEnabled(false)
	}

	hike := h.hike()
	h.roadsAreClearedCheckBox.SetEnabled(hike.RoadsCanBeCleared())
	// The penalty is not used without a successful roll, so the field is blanked as well as disabled, and the same
	// goes for what the rest brings when there is no rest.
	ux.AdjustFieldBlank(h.hikingExtraEffortField, !h.successfulHikingRoll)
	h.restMealCheckBox.SetEnabled(h.restMinutes > 0)
	ux.AdjustFieldBlank(h.restExtraField, h.restMinutes <= 0)
	h.content.MarkForLayoutRecursively()
	h.content.MarkForLayoutRecursivelyUpward()
	h.content.MarkForRedraw()
}

// hike returns the day of travel the calculator's numbers describe.
func (h *hikingCalculator) hike() calculator.Hike {
	return calculator.Hike{
		Hiker: calculator.Hiker{
			Move:          h.move,
			EnhancedMove:  h.enhancedMove,
			Encumbrance:   encumbrance.Level(h.encumbranceIndex),
			FP:            h.fp,
			Fitness:       h.fitness[h.fitnessIndex],
			RecoverEnergy: h.recoverEnergy[h.recoverEnergyIndex],
		},
		Terrain:            h.terrain[h.terrainIndex],
		Weather:            h.weather[h.weatherIndex],
		Heat:               h.heat[h.heatIndex],
		Hours:              h.hikingHours,
		RestMinutes:        h.restMinutes,
		RestExtraFP:        h.restExtraFP,
		ExtraEffortPenalty: h.hikingExtraEffortPenalty,
		RestMeal:           h.restMeal,
		UsingSkis:          h.usingSkis,
		UsingSkates:        h.usingSkates,
		RoadsAreCleared:    h.roadsAreCleared,
		SuccessfulRoll:     h.successfulHikingRoll,
		Metric:             useMetersFor(h.source.entity()),
	}
}

// unitsFor returns the name of the length units the source prefers, for the given distance in them.
func (h *hikingCalculator) unitsFor(distance fxp.Int) string {
	if useMetersFor(h.source.entity()) {
		if distance == fxp.One {
			return i18n.Text("kilometer")
		}
		return i18n.Text("kilometers")
	}
	if distance == fxp.One {
		return i18n.Text("mile")
	}
	return i18n.Text("miles")
}

// distanceText returns the day's distance as it would be without the extra effort, as it is shown.
func (h *hikingCalculator) distanceText(hike *calculator.Hike) string {
	base := *hike
	base.ExtraEffortPenalty = 0
	distance := base.Day().Distance
	return fmt.Sprintf("%s %s", distance.Round().Comma(), h.unitsFor(distance))
}

// updateResults recomputes the day's travel, the time the journey takes and the FP the day costs, and rewrites the
// results.
func (h *hikingCalculator) updateResults() {
	hike := h.hike()
	day := hike.Day()
	units := h.unitsFor(day.Distance)
	h.hikingResult.SetTitle(fmt.Sprintf("%s %s", day.Distance.Round().Comma(), units))
	h.hikingDistanceLabel.SetTitle(i18n.Text("%s to travel", units))

	if timeInDays, ok := calculator.HikingTimeInDays(h.hikingDistance, day.Distance); !ok {
		// No ground is covered at 0 Move or with no hours of travel, so the travel time is undefined.
		h.hikingTimeLabel.SetTitle("—")
	} else if timeInDays == fxp.One {
		h.hikingTimeLabel.SetTitle(i18n.Text("1 day"))
	} else {
		h.hikingTimeLabel.SetTitle(i18n.Text("%s days", timeInDays))
	}

	h.updateFatigue(&hike, day)
	h.updateBreakdown(day)
	h.hikingTimeLabel.MarkForLayoutRecursivelyUpward()
}

// updateFatigue rewrites the FP result and the notes that go with the day. Each hour costs 1 FP, plus 1 per level of
// encumbrance and whatever the heat adds, and extra effort adds a flat amount when the hiker stops (BX357). The notes
// say when the hiker tires, runs out of FP and falls unconscious (BX426), along with the other things the intensity
// brings with it.
func (h *hikingCalculator) updateFatigue(hike *calculator.Hike, day calculator.HikingDay) {
	perHour := hike.FPPerHour().Comma()
	label := i18n.Text("lost by the end of the day (%s per hour)", perHour)
	if hike.ExtraEffort() {
		label = i18n.Text("lost by the end of the day (%s per hour, plus %d for extra effort)", perHour,
			calculator.HikingExtraEffortFP)
	}
	h.fpResult.SetTitle(i18n.Text("%s FP", day.FPCost.Comma()))
	h.fpLabel.SetTitle(label)

	notes := []string{
		i18n.Text("Each hour of hiking costs 1 FP, plus 1 per level of encumbrance, plus 1 on a hot day or 2 in plate armor, an overcoat, etc. (B426)."),
	}
	if hike.Hiker.Fitness.VeryFit {
		notes = append(notes, i18n.Text("Being Very Fit, the hiker loses FP at half that rate (B55)."))
	}
	if day.TiredAfter > 0 {
		notes = append(notes, i18n.Text("With %s FP, the hiker has fewer than a third left after %s hours; Move, Dodge and ST are halved from then on, and the hours below allow for it (B426).",
			h.fp.Comma(), day.TiredAfter.Comma()))
	}
	if day.OutAfter > 0 {
		notes = append(notes, i18n.Text("The hiker is out of FP after %s hours; going on takes a Will roll, which the hours below assume is made, and each further FP lost also costs 1 HP (B426).",
			day.OutAfter.Comma()))
	}
	if day.UnconsciousAfter > 0 {
		notes = append(notes, i18n.Text("At -%s FP the hiker falls unconscious, after %s hours, and the day ends there; any further FP cost comes off HP instead (B426).",
			h.fp.Comma(), day.UnconsciousAfter.Comma()))
	}
	if hike.RestAfter() > 0 {
		notes = append(notes, h.restNote(hike, day))
	}
	switch {
	case hike.ExtraEffort():
		notes = append(notes, i18n.Text("The extra effort makes the Hiking roll a single Will-based Hiking roll at %d for the +%d%% beyond the +20%% a successful roll gives, and adds %d FP to the loss when the hiker stops. A failure leaves the day at the +20%% alone: %s. A critical failure turns the whole loss, %s FP, into HP of injury at the end of the day, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357).",
			h.hikingExtraEffortPenalty, -5*h.hikingExtraEffortPenalty, calculator.HikingExtraEffortFP, h.distanceText(hike),
			day.FPCost.Comma()))
	case h.successfulHikingRoll:
		notes = append(notes, i18n.Text("Extra effort adds 5%% to the distance per -1 taken on the Hiking roll, made as a single Will-based Hiking roll, and %d FP to the loss when the hiker stops (B357).",
			calculator.HikingExtraEffortFP))
	default:
		notes = append(notes, i18n.Text("Extra effort needs the Hiking roll, which it makes a single Will-based Hiking roll at -1 per 5% of distance beyond the +20% a success gives (B357)."))
	}
	if h.intensity[h.hikingIntensityIndex].IsForaging {
		notes = append(notes, i18n.Text("The rest of the day goes to foraging; each attempt takes an hour, during which no progress is made (B427)."))
	}
	if h.hikingHours > fxp.Sixteen {
		notes = append(notes, i18n.Text("More than 16 hours on the march cuts into the 8 hours of sleep the average human needs (B426)."))
	}
	setNotes(h.notes, notes)
	h.fpResult.MarkForLayoutRecursivelyUpward()
}

// restNote describes what the rest halfway through the day gives back and why (BX427, BX55, BX248).
func (h *hikingCalculator) restNote(hike *calculator.Hike, day calculator.HikingDay) string {
	var sources []string
	minutes := hike.Hiker.RestMinutesPerFP()
	spell := hike.Hiker.RecoverEnergy.MinutesPerFP
	switch {
	case spell > 0 && spell <= minutes:
		sources = append(sources, i18n.Text("1 FP per %d minutes with Recover Energy (B248)", minutes))
	case hike.Hiker.Fitness.Fit:
		sources = append(sources, i18n.Text("1 FP per %d minutes, twice the usual rate, for being fit (B55)", minutes))
	default:
		sources = append(sources, i18n.Text("1 FP per %d minutes of quiet rest", minutes))
	}
	if h.restMeal {
		sources = append(sources, i18n.Text("1 for a decent meal"))
	}
	if h.restExtraFP > 0 {
		sources = append(sources, i18n.Text("%d from Lend Energy, potions, etc.", h.restExtraFP))
	}
	return i18n.Text("The rest of %d minutes after %s hours recovers %s FP: %s, and never more than was lost (B427).",
		h.restMinutes, hike.RestAfter().Comma(), day.RestRecovered.Comma(), strings.Join(sources, ", "))
}

// updateBreakdown rewrites the hour-by-hour table: the hours elapsed, the distance covered in the hour and so far, the
// change to the FP in the hour and what is left after it, and the state that leaves the hiker in. The FP left and the
// state are only known when the hiker's FP are.
func (h *hikingCalculator) updateBreakdown(day calculator.HikingDay) {
	h.breakdown.RemoveAllChildren()
	var units string
	if useMetersFor(h.source.entity()) {
		units = i18n.Text("Kilometers")
	} else {
		units = i18n.Text("Miles")
	}
	for i, title := range []string{
		i18n.Text("Hours"), units, i18n.Text("Total"), i18n.Text("FP"),
		i18n.Text("FP left"), i18n.Text("Condition"),
	} {
		label := addResultLabel(h.breakdown)
		label.SetTitle(title)
		if i < 5 {
			label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
		}
	}
	known := h.fp > 0
	for _, hour := range day.Hours {
		addBreakdownCell(h.breakdown, hikingHourLabel(hour), true)
		if hour.Kind == calculator.MarchHour {
			addBreakdownCell(h.breakdown, tenths(hour.Distance), true)
			addBreakdownCell(h.breakdown, tenths(hour.Total), true)
		} else {
			addBreakdownCell(h.breakdown, "", true)
			addBreakdownCell(h.breakdown, "", true)
		}
		addBreakdownCell(h.breakdown, signedTenths(hour.FPChange), true)
		if known {
			addBreakdownCell(h.breakdown, tenths(hour.FPLeft), true)
			addBreakdownCell(h.breakdown, hikingConditionText(hour.Condition), false)
		} else {
			addBreakdownCell(h.breakdown, "—", true)
			addBreakdownCell(h.breakdown, "", false)
		}
	}
	h.breakdown.MarkForLayoutRecursivelyUpward()
}

// hikingHourLabel returns what the first column of the breakdown shows for the row: the hours elapsed, or what the
// row stands for when it covers no distance.
func hikingHourLabel(hour calculator.HikingHour) string {
	switch hour.Kind {
	case calculator.RestHour:
		return i18n.Text("Rest")
	case calculator.StopHour:
		return i18n.Text("Stop")
	default:
		return hour.Elapsed.Comma()
	}
}

// hikingConditionText describes the state the FP lost so far leaves the hiker in (BX426), or is empty when it leaves
// him unaffected.
func hikingConditionText(condition calculator.HikerCondition) string {
	switch condition {
	case calculator.VeryTiredHiker:
		return i18n.Text("Very tired: Move halved")
	case calculator.ExhaustedHiker:
		return i18n.Text("Exhausted: Will roll to go on, 1 HP per FP lost")
	case calculator.UnconsciousHiker:
		return i18n.Text("Unconscious")
	default:
		return ""
	}
}

// addBreakdownCell adds a cell to the hour-by-hour table, aligned to the right when it holds a number.
func addBreakdownCell(table *unison.Panel, text string, number bool) {
	label := addPlainLabel(table, text)
	if number {
		label.SetLayoutData(&unison.FlexLayoutData{HAlign: align.End})
	}
}

// tenths formats a value to the nearest tenth.
func tenths(value fxp.Int) string {
	return value.Mul(fxp.Ten).Round().Div(fxp.Ten).Comma()
}

// signedTenths formats a change to the nearest tenth, with its sign.
func signedTenths(value fxp.Int) string {
	if value > 0 {
		return "+" + tenths(value)
	}
	return tenths(value)
}
