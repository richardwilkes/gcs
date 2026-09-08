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
	"fmt"
	"slices"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var (
	_ calculatorTab = &hikingCalculator{}

	terrain = []terrainModifier{
		{Name: i18n.Text("Broken Ground"), Modifier: fxp.Half},
		{Name: i18n.Text("Deep Snow"), Modifier: fxp.Fifth, IsSnow: true},
		{Name: i18n.Text("Desert"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Desert, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Forest"), Modifier: fxp.Half},
		{Name: i18n.Text("Forest, Dense"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Forest, Light"), Modifier: fxp.One},
		{Name: i18n.Text("Frozen Lake"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Frozen River"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Hills, Rolling"), Modifier: fxp.One},
		{Name: i18n.Text("Hills, Steep"), Modifier: fxp.Half},
		{Name: i18n.Text("Jungle"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mountains"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Mud"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Plains, Level"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Road, Cobblestone"), Modifier: fxp.One, IsRoad: true},
		{Name: i18n.Text("Road, Dirt"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true, Default: true},
		{Name: i18n.Text("Road, Gravel"), Modifier: fxp.One, ModifierInRain: fxp.Fifth, IsRoad: true},
		{Name: i18n.Text("Road, Paved"), Modifier: fxp.OneAndAQuarter, ModifierInRain: fxp.One, IsRoad: true},
		{Name: i18n.Text("Sand"), Modifier: fxp.Fifth},
		{Name: i18n.Text("Sand, Hard-packed"), Modifier: fxp.OneAndAQuarter},
		{Name: i18n.Text("Swamp"), Modifier: fxp.Fifth},
	}

	weather = []terrainModifier{
		{Name: i18n.Text("Normal"), Modifier: fxp.One, Default: true},
		{Name: i18n.Text("Rain"), Modifier: fxp.Half, IsRain: true},
		{Name: i18n.Text("Sleet"), Modifier: fxp.Half, IsIce: true},
		{Name: i18n.Text("Snow"), Modifier: fxp.Half, IsSnow: true},
		{Name: i18n.Text("Snow, Heavy"), Modifier: fxp.Quarter, IsSnow: true},
	}

	hikingIntensity = []hikingIntensityHours{
		{Name: i18n.Text("Forced March"), HoursHiking: fxp.Sixteen},
		{Name: i18n.Text("Long March"), HoursHiking: fxp.Twelve},
		{Name: i18n.Text("Normal"), HoursHiking: fxp.Eight, Default: true},
		{Name: i18n.Text("Foraging"), HoursHiking: fxp.Four, IsForaging: true},
		{Name: i18n.Text("Custom"), IsCustom: true},
	}

	// hikingHeat is what the day's heat adds to each hour's FP cost (BX426).
	hikingHeat = []hikingHeatChoice{
		{name: i18n.Text("Temperate")},
		{name: i18n.Text("Hot (+1 FP per hour)"), fp: 1},
		{name: i18n.Text("Hot, in plate armor, an overcoat, etc. (+2 FP per hour)"), fp: 2},
	}

	// hikingFitness is how fit the hiker is (BX55): Fit and Very Fit recover FP at twice the usual rate, and Very Fit
	// loses FP to exertion at half the usual rate.
	hikingFitness = []hikingFitnessChoice{
		{name: i18n.Text("Average")},
		{name: i18n.Text("Fit (recovers FP twice as fast)"), fit: true},
		{name: i18n.Text("Very Fit (recovers FP twice as fast, loses FP half as fast)"), fit: true, veryFit: true},
	}

	// hikingRecoverEnergy is what the Recover Energy spell (BX248) does for the hiker's rest, by the skill it is known at.
	hikingRecoverEnergy = []hikingRecoverEnergyChoice{
		{name: i18n.Text("None")},
		{name: i18n.Text("Skill 15+ (1 FP per 5 minutes of rest)"), minutesPerFP: 5},
		{name: i18n.Text("Skill 20+ (1 FP per 2 minutes of rest)"), minutesPerFP: 2},
	}
)

// hikingExtraEffortFP is what extra effort on the march adds to the FP lost when the hiker stops (BX357).
const hikingExtraEffortFP = 2

// hikingRestMinutesPerFP is how long a rest takes to recover 1 FP for someone with no help (BX427).
const hikingRestMinutesPerFP = 10

type terrainModifier struct {
	Name           string
	Modifier       fxp.Int
	ModifierInRain fxp.Int
	IsRoad         bool
	IsRain         bool
	IsSnow         bool
	IsIce          bool
	Default        bool
}

func (t terrainModifier) String() string {
	return t.Name
}

type hikingIntensityHours struct {
	Name        string
	HoursHiking fxp.Int
	IsCustom    bool
	IsForaging  bool
	Default     bool
}

func (t hikingIntensityHours) String() string {
	return t.Name
}

type hikingHeatChoice struct {
	name string
	fp   int
}

func (h hikingHeatChoice) String() string {
	return h.name
}

type hikingFitnessChoice struct {
	name    string
	fit     bool
	veryFit bool
}

func (h hikingFitnessChoice) String() string {
	return h.name
}

type hikingRecoverEnergyChoice struct {
	name         string
	minutesPerFP int // 0 when the spell is no help
}

func (h hikingRecoverEnergyChoice) String() string {
	return h.name
}

// hikingCalculator works out how far a character covers in a day of travel over the given ground (BX351, HT55), and
// what the day costs in FP (BX426). The character's numbers are either typed in or taken from any open character
// sheet, in which case the fields holding them are locked and refreshed whenever the sheet changes. The journey itself
// is always typed in.
type hikingCalculator struct {
	calculatorContent
	source                       sheetSourcePicker
	moveField                    *IntegerField
	enhancedMoveField            *DecimalField
	encumbrancePopup             *unison.PopupMenu[encumbrance.Level]
	fpField                      *DecimalField
	fitnessPopup                 *unison.PopupMenu[hikingFitnessChoice]
	recoverEnergyPopup           *unison.PopupMenu[hikingRecoverEnergyChoice]
	restField                    *IntegerField
	restMealCheckBox             *unison.CheckBox
	restExtraField               *IntegerField
	breakdown                    *unison.Panel
	notes                        *unison.Panel
	hikingResult                 *unison.Label
	hikingDistanceLabel          *textLabel
	hikingTimeLabel              *unison.Label
	fpResult                     *unison.Label
	fpLabel                      *textLabel
	hikingHoursField             *DecimalField
	hikingExtraEffortField       *IntegerField
	roadsAreClearedCheckBox      *unison.CheckBox
	usingSkisCheckBox            *unison.CheckBox
	usingSkatesCheckBox          *unison.CheckBox
	successfulHikingRollCheckBox *unison.CheckBox
	hikingRollPageLabel          *textLabel
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
		move:                 5,
		fp:                   fxp.Ten,
		terrainIndex:         slices.IndexFunc(terrain, func(t terrainModifier) bool { return t.Default }),
		weatherIndex:         slices.IndexFunc(weather, func(t terrainModifier) bool { return t.Default }),
		hikingIntensityIndex: slices.IndexFunc(hikingIntensity, func(t hikingIntensityHours) bool { return t.Default }),
		hikingHours:          fxp.Eight,
	}
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
func (h *hikingCalculator) preselect(sheet *Sheet) {
	h.source.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (h *hikingCalculator) sheetChanged(sheet *Sheet) {
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
	h.moveField = sameWidth(NewIntegerField(nil, "", i18n.Text("Move"),
		func() int { return h.move },
		func(v int) {
			h.move = v
			h.changed()
		},
		0, 10000, false, false))
	h.addFieldRow(h.moveField, i18n.Text("Move, with encumbrance"))
	h.enhancedMoveField = sameWidth(NewDecimalField(nil, "", i18n.Text("Enhanced Move"),
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
	h.fpField = sameWidth(NewDecimalField(nil, "", i18n.Text("Fatigue Points"),
		func() fxp.Int { return h.fp },
		func(v fxp.Int) {
			h.fp = v
			h.changed()
		},
		0, fxp.Max, false, false))
	h.addFieldRow(h.fpField, i18n.Text("FP"))
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Fitness:"))
	h.fitnessPopup = addIndexPopup(row, hikingFitness, &h.fitnessIndex, h.changed)
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Recover Energy:"))
	h.recoverEnergyPopup = addIndexPopup(row, hikingRecoverEnergy, &h.recoverEnergyIndex, h.changed)

	h.addSubheader(i18n.Text("Journey"))
	row = h.addRow(2)
	addPlainLabel(row, i18n.Text("Terrain:"))
	addIndexPopup(row, terrain, &h.terrainIndex, h.changed)
	addPlainLabel(row, i18n.Text("Weather:"))
	addIndexPopup(row, weather, &h.weatherIndex, h.changed)
	addPlainLabel(row, i18n.Text("Heat:"))
	addIndexPopup(row, hikingHeat, &h.heatIndex, h.changed)
	addPlainLabel(row, i18n.Text("Intensity:"))
	addIndexPopup(row, hikingIntensity, &h.hikingIntensityIndex, h.changed)

	h.roadsAreClearedCheckBox = h.addCheckBox(i18n.Text("Roads are cleared"), &h.roadsAreCleared, h.changed)
	h.usingSkisCheckBox = h.addCheckBox(i18n.Text("Using skis"), &h.usingSkis, h.changed)
	h.usingSkatesCheckBox = h.addCheckBox(i18n.Text("Using skates"), &h.usingSkates, h.changed)
	// The title names the skill the roll is against, which depends on the mode of travel, so adjustControls sets it,
	// along with the page the skill is on, which sits beside the checkbox as a link.
	row = h.addRow(2)
	h.successfulHikingRollCheckBox = newCheckBox("", &h.successfulHikingRoll, h.changed)
	row.AddChild(h.successfulHikingRollCheckBox)
	h.hikingRollPageLabel = addPlainLabel(row, "")

	h.hikingHoursField = sameWidth(NewDecimalField(nil, "", i18n.Text("Traveling Hours per Day"),
		func() fxp.Int { return h.hikingHours },
		func(v fxp.Int) {
			h.hikingHours = v
			h.changed()
		},
		0, fxp.TwentyFour, false, false))
	h.addFieldRow(h.hikingHoursField, i18n.Text("hours of hiking per day"))
	h.restField = sameWidth(NewIntegerField(nil, "", i18n.Text("Rest"),
		func() int { return h.restMinutes },
		func(v int) {
			h.restMinutes = v
			h.changed()
		},
		0, 24*60, false, false))
	h.addFieldRow(h.restField, i18n.Text("minutes of rest halfway through the day (0 for none)"))
	h.restMealCheckBox = h.addCheckBox(i18n.Text("Eats a decent meal while resting (+1 FP)"), &h.restMeal, h.changed)
	h.restExtraField = sameWidth(NewIntegerField(nil, "", i18n.Text("FP Restored While Resting"),
		func() int { return h.restExtraFP },
		func(v int) {
			h.restExtraFP = v
			h.changed()
		},
		0, 1000, false, false))
	h.addFieldRow(h.restExtraField, i18n.Text("FP restored by Lend Energy, potions, etc. while resting"))
	h.hikingExtraEffortField = sameWidth(NewIntegerField(nil, "", i18n.Text("Hiking Extra Effort Penalty"),
		func() int { return h.hikingExtraEffortPenalty },
		func(v int) {
			h.hikingExtraEffortPenalty = v
			h.changed()
		},
		-100, 0, false, false))
	h.addFieldRow(h.hikingExtraEffortField, i18n.Text("penalty taken for extra effort (+5% distance per -1)"))
	h.hikingDistanceLabel = h.addFieldRow(sameWidth(NewDecimalField(nil, "", i18n.Text("Distance to Cover"),
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
func (h *hikingCalculator) selectSheet(sheet *Sheet) {
	if sheet != nil {
		h.pullFromSheet()
	}
}

// pullFromSheet reads the hiker's numbers from its sheet. Each backing value is assigned before its field is synced, so
// that the setter the sync may run sees nothing new and does not start another round of updates.
func (h *hikingCalculator) pullFromSheet() {
	entity := h.source.entity()
	h.encumbranceIndex = int(entity.EncumbranceLevel(false))
	h.move = entity.Move(encumbrance.Level(h.encumbranceIndex))
	h.enhancedMove, _ = entity.TraitLevels("enhanced move (ground)")
	if entity.ResolveAttribute(gurps.FatiguePointsID) != nil {
		h.fp = entity.Attributes.Maximum(gurps.FatiguePointsID).Max(0)
	}
	switch {
	case entity.HasTraitNamed("Very Fit"):
		h.fitnessIndex = 2
	case entity.HasTraitNamed("Fit"):
		h.fitnessIndex = 1
	default:
		h.fitnessIndex = 0
	}
	h.recoverEnergyIndex = 0
	if level := spellLevel(entity, "Recover Energy"); level >= 20 {
		h.recoverEnergyIndex = 2
	} else if level >= 15 {
		h.recoverEnergyIndex = 1
	}
	h.moveField.Sync()
	h.enhancedMoveField.Sync()
	h.encumbrancePopup.SelectIndex(h.encumbranceIndex)
	h.fpField.Sync()
	h.fitnessPopup.SelectIndex(h.fitnessIndex)
	h.recoverEnergyPopup.SelectIndex(h.recoverEnergyIndex)
}

// spellLevel returns the highest level the entity knows the named spell at, or 0 when it does not know it.
func spellLevel(entity *gurps.Entity, name string) int {
	var level fxp.Int
	gurps.Traverse(func(sp *gurps.Spell) bool {
		if strings.EqualFold(sp.NameWithReplacements(), name) {
			level = level.Max(sp.CalculateLevel().Level)
		}
		return false
	}, true, true, entity.Spells...)
	return level.AsInteger[int]()
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

	i := hikingIntensity[h.hikingIntensityIndex]
	h.hikingHoursField.SetEnabled(true)
	if !i.IsCustom {
		h.hikingHours = i.HoursHiking
		h.hikingHoursField.Sync()
		h.hikingHoursField.SetEnabled(false)
	}

	w := weather[h.weatherIndex]
	h.roadsAreClearedCheckBox.SetEnabled(terrain[h.terrainIndex].IsRoad && (w.IsIce || w.IsSnow))
	// The penalty is not used without a successful roll, so the field is blanked as well as disabled, and the same
	// goes for what the rest brings when there is no rest.
	adjustFieldBlank(h.hikingExtraEffortField, !h.successfulHikingRoll)
	h.restMealCheckBox.SetEnabled(h.restMinutes > 0)
	adjustFieldBlank(h.restExtraField, h.restMinutes <= 0)
	h.content.MarkForLayoutRecursively()
	h.content.MarkForLayoutRecursivelyUpward()
	h.content.MarkForRedraw()
}

// extraEffort reports whether the hiker is putting in extra effort, which takes a successful roll and a penalty on it.
func (h *hikingCalculator) extraEffort() bool {
	return h.successfulHikingRoll && h.hikingExtraEffortPenalty < 0
}

// distanceForHours returns the distance covered in the given hours of travel at the hiker's full pace, with the given
// extra effort penalty, in the length units the source prefers (BX351).
func (h *hikingCalculator) distanceForHours(hours fxp.Int, extraEffortPenalty int) fxp.Int {
	distance := fxp.FromInteger(h.move * 10)

	// Adjust for hours hiking
	distance = distance.Mul(hours).Div(fxp.Sixteen)

	// Adjust for enhanced move (ground), if any
	if h.enhancedMove > 0 {
		distance = distance.Mul(fxp.One + h.enhancedMove)
	}

	// Adjust for terrain
	t := terrain[h.terrainIndex]
	mod := t.Modifier
	if t.IsIce && h.usingSkates {
		mod = fxp.OneAndAQuarter
	}
	if t.IsSnow && h.usingSkis {
		mod = fxp.One
	}

	// Adjust for weather
	w := weather[h.weatherIndex]
	switch {
	case w.IsRain:
		if t.IsRoad {
			if t.ModifierInRain != 0 {
				mod = t.ModifierInRain
			}
		} else {
			mod = mod.Mul(w.Modifier)
		}
	case w.IsSnow:
		if t.IsRoad {
			mod = fxp.One
		}
		if (!t.IsRoad || !h.roadsAreCleared) && !h.usingSkis {
			mod = mod.Mul(w.Modifier)
		}
	case w.IsIce:
		if t.IsRoad {
			mod = fxp.One
		}
		if (!t.IsRoad || !h.roadsAreCleared) && !h.usingSkates {
			mod = mod.Mul(w.Modifier)
		}
	}
	distance = distance.Mul(mod)

	// Adjust for making the hiking/skiing/skating check
	mod = fxp.One
	if h.successfulHikingRoll {
		mod = fxp.OnePointTwo
		if extraEffortPenalty < 0 {
			mod += fxp.FromInteger(-5 * extraEffortPenalty).Div(fxp.Hundred)
		}
	}
	distance = distance.Mul(mod)

	if useMetersFor(h.source.entity()) {
		// miles -> inches -> GURPS kilometers
		distance = fxp.Kilometer.FromInches(fxp.Mile.ToInches(distance))
	}
	return distance
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

// hikingHour is one row of a day's hour-by-hour breakdown: what the hour covered, what it did to the hiker's FP, and
// the state it left him in. The rows for the rest halfway through and for the stop at the end of the day, which is
// where extra effort takes its toll, cover nothing.
type hikingHour struct {
	label       string
	condition   string
	distance    fxp.Int
	total       fxp.Int
	fpChange    fxp.Int // Negative for a loss.
	fpLeft      fxp.Int
	hasDistance bool
}

// hikingDay is what a day of travel comes to, worked out hour by hour with the effects of the FP lost applied as they
// arrive (BX426): Move is halved once fewer than a third of the hiker's FP are left, going on at 0 FP or less takes a
// Will roll, which the day assumes is made, and at -FP the hiker falls unconscious and the day ends. The hours are
// what the hiker can be expected to manage, and the cost is what they and the stop at the end come to.
type hikingDay struct {
	hours            []hikingHour
	distance         fxp.Int
	fpCost           fxp.Int
	restRecovered    fxp.Int // What the rest halfway through gave back; 0 when there was no rest.
	tiredAfter       fxp.Int
	outAfter         fxp.Int
	unconsciousAfter fxp.Int
}

// fpPerHour returns what each hour of the march costs in FP (BX426): 1, plus 1 per level of encumbrance, plus what
// the heat adds, halved for a Very Fit hiker (BX55).
func (h *hikingCalculator) fpPerHour() fxp.Int {
	perHour := fxp.FromInteger(1 + h.encumbranceIndex + hikingHeat[h.heatIndex].fp)
	if hikingFitness[h.fitnessIndex].veryFit {
		perHour = perHour.Div(fxp.Two)
	}
	return perHour
}

// restMinutesPerFP returns how long the hiker's rest takes to recover 1 FP: 10 minutes (BX427), or 5 for a Fit or
// Very Fit hiker (BX55), or what Recover Energy gives if that is faster (BX248).
func (h *hikingCalculator) restMinutesPerFP() int {
	minutes := hikingRestMinutesPerFP
	if hikingFitness[h.fitnessIndex].fit {
		minutes /= 2
	}
	if spell := hikingRecoverEnergy[h.recoverEnergyIndex].minutesPerFP; spell > 0 && spell < minutes {
		minutes = spell
	}
	return minutes
}

// restAfter returns how many hours into the march the hiker rests, or 0 when there is no rest: halfway through, at the
// end of a whole hour, and only when some of the march is still to come.
func (h *hikingCalculator) restAfter() fxp.Int {
	if h.restMinutes <= 0 {
		return 0
	}
	after := h.hikingHours.Div(fxp.Two).Floor().Max(fxp.One)
	if after >= h.hikingHours {
		return 0
	}
	return after
}

// restRecovery returns what the rest gives back to a hiker with the given FP left: 1 FP per restMinutesPerFP of it,
// plus 1 for a decent meal and whatever spells and the like restore, but never more than was lost (BX427).
func (h *hikingCalculator) restRecovery(fpLeft fxp.Int) fxp.Int {
	recovered := h.restMinutes / h.restMinutesPerFP()
	if h.restMeal {
		recovered++
	}
	recovered += h.restExtraFP
	amount := fxp.FromInteger(recovered)
	if h.fp > 0 {
		amount = amount.Min(h.fp - fpLeft).Max(0)
	}
	return amount
}

// computeDay works out the day of travel with the given extra effort penalty. Without FP to go on, the hiker is
// assumed to keep his full pace all day.
func (h *hikingCalculator) computeDay(extraEffortPenalty int) hikingDay {
	var day hikingDay
	hourly := h.distanceForHours(fxp.One, extraEffortPenalty)
	perHour := h.fpPerHour()
	known := h.fp > 0
	fpLeft := h.fp
	pace := fxp.One
	restAfter := h.restAfter()
	var elapsed fxp.Int
	for remaining := h.hikingHours; remaining > 0; {
		step := remaining.Min(fxp.One)
		remaining -= step
		elapsed += step
		hour := hikingHour{
			label:       elapsed.Comma(),
			distance:    hourly.Mul(step).Mul(pace),
			fpChange:    -perHour.Mul(step),
			hasDistance: true,
		}
		day.distance += hour.distance
		hour.total = day.distance
		day.fpCost -= hour.fpChange
		if known {
			// FP never fall below -FP; whatever would take them further comes off HP instead.
			fpLeft = (fpLeft + hour.fpChange).Max(-h.fp)
			hour.fpLeft = fpLeft
			hour.condition, pace = h.condition(fpLeft)
			if day.tiredAfter == 0 && pace < fxp.One {
				day.tiredAfter = elapsed
			}
			if day.outAfter == 0 && fpLeft <= 0 {
				day.outAfter = elapsed
			}
			if fpLeft <= -h.fp {
				day.unconsciousAfter = elapsed
				remaining = 0
			}
		}
		day.hours = append(day.hours, hour)
		if remaining > 0 && elapsed == restAfter {
			rest := hikingHour{label: i18n.Text("Rest"), fpChange: h.restRecovery(fpLeft)}
			day.restRecovered = rest.fpChange
			if known {
				fpLeft += rest.fpChange
				rest.fpLeft = fpLeft
				rest.condition, pace = h.condition(fpLeft)
			}
			day.hours = append(day.hours, rest)
		}
	}
	if h.successfulHikingRoll && extraEffortPenalty < 0 {
		hour := hikingHour{label: i18n.Text("Stop"), fpChange: -fxp.FromInteger(hikingExtraEffortFP)}
		day.fpCost -= hour.fpChange
		if known {
			fpLeft = (fpLeft + hour.fpChange).Max(-h.fp)
			hour.fpLeft = fpLeft
			hour.condition, _ = h.condition(fpLeft)
			if day.unconsciousAfter == 0 && fpLeft <= -h.fp {
				day.unconsciousAfter = elapsed
			}
		}
		day.hours = append(day.hours, hour)
	}
	return day
}

// condition returns the state the hiker is in with the given FP left, and the pace he can keep up as a fraction of
// his usual one: half once fewer than a third of his FP are left, and none once he is unconscious (BX426).
func (h *hikingCalculator) condition(fpLeft fxp.Int) (text string, pace fxp.Int) {
	switch {
	case fpLeft <= -h.fp:
		return i18n.Text("Unconscious"), 0
	case fpLeft <= 0:
		return i18n.Text("Exhausted: Will roll to go on, 1 HP per FP lost"), fxp.Half
	case fpLeft.Mul(fxp.Three) < h.fp:
		return i18n.Text("Very tired: Move halved"), fxp.Half
	default:
		return "", fxp.One
	}
}

// distanceText returns the day's distance with the given extra effort penalty, as it is shown.
func (h *hikingCalculator) distanceText(extraEffortPenalty int) string {
	distance := h.computeDay(extraEffortPenalty).distance
	return fmt.Sprintf("%s %s", distance.Round().Comma(), h.unitsFor(distance))
}

// updateResults recomputes the day's travel, the time the journey takes and the FP the day costs, and rewrites the
// results.
func (h *hikingCalculator) updateResults() {
	day := h.computeDay(h.hikingExtraEffortPenalty)
	units := h.unitsFor(day.distance)
	h.hikingResult.SetTitle(fmt.Sprintf("%s %s", day.distance.Round().Comma(), units))
	h.hikingDistanceLabel.SetTitle(fmt.Sprintf(i18n.Text("%s to travel"), units))

	if timeInDays, ok := hikingTimeInDays(h.hikingDistance, day.distance); !ok {
		// Ground can't be covered at 0 Move (very low DX/HT, heavy encumbrance, or a Move-reducing effect), so the
		// travel time is undefined.
		h.hikingTimeLabel.SetTitle("—")
	} else if timeInDays == fxp.One {
		h.hikingTimeLabel.SetTitle(i18n.Text("1 day"))
	} else {
		h.hikingTimeLabel.SetTitle(fmt.Sprintf(i18n.Text("%s days"), timeInDays))
	}

	h.updateFatigue(day)
	h.updateBreakdown(day)
	h.hikingTimeLabel.MarkForLayoutRecursivelyUpward()
}

// updateFatigue rewrites the FP result and the notes that go with the day. Each hour costs 1 FP, plus 1 per level of
// encumbrance and whatever the heat adds, and extra effort adds a flat amount when the hiker stops (BX357). The notes
// say when the hiker tires, runs out of FP and falls unconscious (BX426), along with the other things the intensity
// brings with it.
func (h *hikingCalculator) updateFatigue(day hikingDay) {
	perHour := h.fpPerHour().Comma()
	label := fmt.Sprintf(i18n.Text("lost by the end of the day (%s per hour)"), perHour)
	if h.extraEffort() {
		label = fmt.Sprintf(i18n.Text("lost by the end of the day (%s per hour, plus %d for extra effort)"), perHour,
			hikingExtraEffortFP)
	}
	h.fpResult.SetTitle(fmt.Sprintf(i18n.Text("%s FP"), day.fpCost.Comma()))
	h.fpLabel.SetTitle(label)

	notes := []string{
		i18n.Text("Each hour of hiking costs 1 FP, plus 1 per level of encumbrance, plus 1 on a hot day or 2 in plate armor, an overcoat, etc. (B426)."),
	}
	if hikingFitness[h.fitnessIndex].veryFit {
		notes = append(notes, i18n.Text("Being Very Fit, the hiker loses FP at half that rate (B55)."))
	}
	if day.tiredAfter > 0 {
		notes = append(notes, fmt.Sprintf(i18n.Text("With %s FP, the hiker has fewer than a third left after %s hours; Move, Dodge and ST are halved from then on, and the hours below allow for it (B426)."),
			h.fp.Comma(), day.tiredAfter.Comma()))
	}
	if day.outAfter > 0 {
		notes = append(notes, fmt.Sprintf(i18n.Text("The hiker is out of FP after %s hours; going on takes a Will roll, which the hours below assume is made, and each further FP lost also costs 1 HP (B426)."),
			day.outAfter.Comma()))
	}
	if day.unconsciousAfter > 0 {
		notes = append(notes, fmt.Sprintf(i18n.Text("At -%s FP the hiker falls unconscious, after %s hours, and the day ends there; any further FP cost comes off HP instead (B426)."),
			h.fp.Comma(), day.unconsciousAfter.Comma()))
	}
	if h.restAfter() > 0 {
		notes = append(notes, h.restNote(day))
	}
	switch {
	case h.extraEffort():
		notes = append(notes, fmt.Sprintf(i18n.Text("The extra effort makes the Hiking roll a single Will-based Hiking roll at %d for the +%d%% beyond the +20%% a successful roll gives, and adds %d FP to the loss when the hiker stops. A failure leaves the day at the +20%% alone: %s. A critical failure turns the whole loss, %s FP, into HP of injury at the end of the day, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357)."),
			h.hikingExtraEffortPenalty, -5*h.hikingExtraEffortPenalty, hikingExtraEffortFP, h.distanceText(0),
			day.fpCost.Comma()))
	case h.successfulHikingRoll:
		notes = append(notes, fmt.Sprintf(i18n.Text("Extra effort adds 5%% to the distance per -1 taken on the Hiking roll, made as a single Will-based Hiking roll, and %d FP to the loss when the hiker stops (B357)."),
			hikingExtraEffortFP))
	default:
		notes = append(notes, i18n.Text("Extra effort needs the Hiking roll, which it makes a single Will-based Hiking roll at -1 per 5% of distance beyond the +20% a success gives (B357)."))
	}
	if hikingIntensity[h.hikingIntensityIndex].IsForaging {
		notes = append(notes, i18n.Text("The rest of the day goes to foraging; each attempt takes an hour, during which no progress is made (B427)."))
	}
	if h.hikingHours > fxp.Sixteen {
		notes = append(notes, i18n.Text("More than 16 hours on the march cuts into the 8 hours of sleep the average human needs (B426)."))
	}
	setNotes(h.notes, notes)
	h.fpResult.MarkForLayoutRecursivelyUpward()
}

// restNote describes what the rest halfway through the day gives back and why (BX427, BX55, BX248).
func (h *hikingCalculator) restNote(day hikingDay) string {
	var sources []string
	spell := hikingRecoverEnergy[h.recoverEnergyIndex].minutesPerFP
	switch {
	case spell > 0 && spell <= h.restMinutesPerFP():
		sources = append(sources, fmt.Sprintf(i18n.Text("1 FP per %d minutes with Recover Energy (B248)"),
			h.restMinutesPerFP()))
	case hikingFitness[h.fitnessIndex].fit:
		sources = append(sources, fmt.Sprintf(i18n.Text("1 FP per %d minutes, twice the usual rate, for being fit (B55)"),
			h.restMinutesPerFP()))
	default:
		sources = append(sources, fmt.Sprintf(i18n.Text("1 FP per %d minutes of quiet rest"), h.restMinutesPerFP()))
	}
	if h.restMeal {
		sources = append(sources, i18n.Text("1 for a decent meal"))
	}
	if h.restExtraFP > 0 {
		sources = append(sources, fmt.Sprintf(i18n.Text("%d from Lend Energy, potions, etc."), h.restExtraFP))
	}
	return fmt.Sprintf(i18n.Text("The rest of %d minutes after %s hours recovers %s FP: %s, and never more than was lost (B427)."),
		h.restMinutes, h.restAfter().Comma(), day.restRecovered.Comma(), strings.Join(sources, ", "))
}

// updateBreakdown rewrites the hour-by-hour table: the hours elapsed, the distance covered in the hour and so far, the
// change to the FP in the hour and what is left after it, and the state that leaves the hiker in. The FP left and the
// state are only known when the hiker's FP are.
func (h *hikingCalculator) updateBreakdown(day hikingDay) {
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
	for _, hour := range day.hours {
		addBreakdownCell(h.breakdown, hour.label, true)
		if hour.hasDistance {
			addBreakdownCell(h.breakdown, tenths(hour.distance), true)
			addBreakdownCell(h.breakdown, tenths(hour.total), true)
		} else {
			addBreakdownCell(h.breakdown, "", true)
			addBreakdownCell(h.breakdown, "", true)
		}
		addBreakdownCell(h.breakdown, signedTenths(hour.fpChange), true)
		if known {
			addBreakdownCell(h.breakdown, tenths(hour.fpLeft), true)
			addBreakdownCell(h.breakdown, hour.condition, false)
		} else {
			addBreakdownCell(h.breakdown, "—", true)
			addBreakdownCell(h.breakdown, "", false)
		}
	}
	h.breakdown.MarkForLayoutRecursivelyUpward()
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

// hikingTimeInDays returns the number of days needed to cover distanceToCover while traveling distancePerDay each day,
// rounded to a tenth of a day. ok is false when distancePerDay is 0, since no ground can be covered at 0 Move and the
// travel time is therefore undefined; the guard also avoids a division by zero (fxp.Int.Div panics on a zero divisor
// with a non-zero numerator).
func hikingTimeInDays(distanceToCover, distancePerDay fxp.Int) (days fxp.Int, ok bool) {
	if distancePerDay == 0 {
		return 0, false
	}
	return distanceToCover.Mul(fxp.Ten).Div(distancePerDay).Round().Div(fxp.Ten), true
}
