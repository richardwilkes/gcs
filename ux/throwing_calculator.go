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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var (
	_ calculatorTab = &throwingCalculator{}

	// throwingTiers are the levels of the Throwing skill that matter to a throw (BX355): at DX+1 it adds 1 to the ST
	// the distance is worked out from, and at DX+2 or better it adds 2. It does nothing for the damage.
	throwingTiers = []throwingSkillTier{
		{name: i18n.Text("None, or below DX+1")},
		{name: i18n.Text("DX+1"), distance: 1},
		{name: i18n.Text("DX+2 or better"), distance: 2},
	}

	// throwingArtTiers are the levels of the Throwing Art skill that matter to a throw (BX355): at DX it adds 1 to the
	// ST the distance is worked out from and 1 per die to the damage, and at DX+1 or better it adds 2 of each.
	throwingArtTiers = []throwingSkillTier{
		{name: i18n.Text("None, or below DX")},
		{name: i18n.Text("DX"), distance: 1, damage: 1},
		{name: i18n.Text("DX+1 or better"), distance: 2, damage: 2},
	}
)

// throwingSkillTier is a level of a throwing skill, relative to DX, and what it adds to a throw.
type throwingSkillTier struct {
	name     string
	distance int
	damage   int
}

func (t throwingSkillTier) String() string {
	return t.name
}

// throwingCalculator works out how far a character can throw an object and what it does when it lands (BX355). The
// character's numbers are either typed in or taken from any open character sheet, in which case the fields holding
// them are locked and refreshed whenever the sheet changes. The object and the extra effort are always typed in.
type throwingCalculator struct {
	calculatorContent
	source             sheetSourcePicker
	stField            *DecimalField
	strikingSTField    *DecimalField
	throwingPopup      *unison.PopupMenu[throwingSkillTier]
	throwingArtPopup   *unison.PopupMenu[throwingSkillTier]
	weightField        *WeightField
	distanceResult     *unison.Label
	damageResult       *unison.Label
	notes              *unison.Panel
	st                 fxp.Int
	strikingST         fxp.Int
	objectWeight       fxp.Weight
	throwingIndex      int
	throwingArtIndex   int
	extraEffortPenalty int
	updating           bool
}

func newThrowingCalculator() *throwingCalculator {
	t := &throwingCalculator{
		st:           fxp.Ten,
		strikingST:   fxp.Ten,
		objectWeight: fxp.Weight(fxp.One),
	}
	t.createContent()
	return t
}

// title implements calculatorTab.
func (t *throwingCalculator) title() string {
	return i18n.Text("Throwing")
}

// panel implements calculatorTab.
func (t *throwingCalculator) panel() *unison.Panel {
	return t.content
}

// preselect implements calculatorTab.
func (t *throwingCalculator) preselect(sheet *Sheet) {
	t.source.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (t *throwingCalculator) sheetChanged(sheet *Sheet) {
	if t.source.sheet == sheet {
		t.changed()
	}
}

func (t *throwingCalculator) createContent() {
	t.initCalculatorContent()
	t.content.AddChild(t.createHeader(i18n.Text("Throwing"), []linkSpec{{pageRef: "BX355", highlight: "Throwing"}}, 0))

	t.addSubheader(i18n.Text("Thrower"))
	t.source.selected = t.selectSheet
	t.source.refreshed = t.pullFromSheet
	t.source.changed = t.changed
	t.source.addRow(&t.calculatorContent, i18n.Text("Source:"))

	t.stField = sameWidth(NewDecimalField(nil, "", i18n.Text("ST"),
		func() fxp.Int { return t.st },
		func(v fxp.Int) {
			t.st = v
			t.changed()
		},
		0, fxp.Max, false, false))
	t.addFieldRow(t.stField, i18n.Text("ST the distance is worked out from"))
	t.strikingSTField = sameWidth(NewDecimalField(nil, "", i18n.Text("Striking ST"),
		func() fxp.Int { return t.strikingST },
		func(v fxp.Int) {
			t.strikingST = v
			t.changed()
		},
		0, fxp.Max, false, false))
	t.addFieldRow(t.strikingSTField, i18n.Text("Striking ST the damage is worked out from"))
	row := t.addRow(2)
	addPlainLabel(row, i18n.Text("Throwing:"))
	t.throwingPopup = addIndexPopup(row, throwingTiers, &t.throwingIndex, t.changed)
	row = t.addRow(2)
	addPlainLabel(row, i18n.Text("Throwing Art:"))
	t.throwingArtPopup = addIndexPopup(row, throwingArtTiers, &t.throwingArtIndex, t.changed)

	t.addSubheader(i18n.Text("Throw"))
	t.weightField = sameWidth(newSourcedWeightField(i18n.Text("Object Weight"), t.source.entity,
		func() fxp.Weight { return t.objectWeight },
		func(v fxp.Weight) {
			t.objectWeight = v
			t.changed()
		},
		0, fxp.Weight(fxp.Max)))
	t.addFieldRow(t.weightField, i18n.Text("object"))
	t.addFieldRow(sameWidth(NewIntegerField(nil, "", i18n.Text("Throwing Extra Effort Penalty"),
		func() int { return t.extraEffortPenalty },
		func(v int) {
			t.extraEffortPenalty = v
			t.changed()
		},
		-100, 0, false, false)), i18n.Text("penalty taken for extra effort (+5% ST per -1)"))

	box := t.addResultsBox()
	row = box.addRow(2)
	addPlainLabel(row, i18n.Text("Distance:"))
	t.distanceResult = addResultLabel(row)
	addPlainLabel(row, i18n.Text("Damage:"))
	t.damageResult = addResultLabel(row)
	t.notes = box.addNotesGroup()
}

// selectSheet reads the thrower's numbers from the sheet the picker has just made its source, and does nothing at all
// when there is none.
func (t *throwingCalculator) selectSheet(sheet *Sheet) {
	if sheet != nil {
		t.pullFromSheet()
	}
}

// pullFromSheet reads the thrower's numbers from its sheet. Each backing value is assigned before its field is synced,
// so that the setter the sync may run sees nothing new and does not start another round of updates.
func (t *throwingCalculator) pullFromSheet() {
	entity := t.source.entity()
	t.st = (entity.LiftingStrength() - entity.LiftingStrengthBonus).Max(0)
	t.strikingST = entity.StrikingStrength().Max(0)
	t.throwingIndex = throwingTierIndex(entity, "Throwing", fxp.One)
	t.throwingArtIndex = throwingTierIndex(entity, "Throwing Art", 0)
	t.stField.Sync()
	t.strikingSTField.Sync()
	t.throwingPopup.SelectIndex(t.throwingIndex)
	t.throwingArtPopup.SelectIndex(t.throwingArtIndex)
}

// throwingTierIndex returns the tier the entity's level in the named skill falls in: 0 below the given relative level,
// 1 at it, and 2 a full level or more above it.
func throwingTierIndex(entity *gurps.Entity, name string, first fxp.Int) int {
	sk := entity.BestSkillNamed(name, "", false, nil)
	if sk == nil {
		return 0
	}
	switch relative := sk.CalculateLevel(nil).RelativeLevel; {
	case relative >= first+fxp.One:
		return 2
	case relative >= first:
		return 1
	default:
		return 0
	}
}

// changed implements calculatorTab.
func (t *throwingCalculator) changed() {
	if t.updating {
		return
	}
	t.updating = true
	defer func() { t.updating = false }()
	t.source.refresh()
	t.adjustControls()
	t.updateResults()
}

// adjustControls locks the fields a sheet supplies.
func (t *throwingCalculator) adjustControls() {
	manual := t.source.sheet == nil
	t.stField.SetEnabled(manual)
	t.strikingSTField.SetEnabled(manual)
	t.throwingPopup.SetEnabled(manual)
	t.throwingArtPopup.SetEnabled(manual)
	// The weight is shown in the units the source prefers, which may have just changed along with it.
	t.weightField.Sync()
	t.content.MarkForLayoutRecursively()
	t.content.MarkForLayoutRecursivelyUpward()
	t.content.MarkForRedraw()
}

// updateResults recomputes the throw and rewrites the results (BX355), along with the note explaining the extra
// effort (BX357): what it takes, what it costs, and what a failure and a critical failure leave the thrower with.
func (t *throwingCalculator) updateResults() {
	distance, damage := t.computeThrow(t.extraEffortPenalty)
	t.distanceResult.SetTitle(distance)
	t.damageResult.SetTitle(damage)
	var note string
	if t.extraEffortPenalty < 0 {
		distance, damage = t.computeThrow(0)
		note = fmt.Sprintf(i18n.Text("The extra effort takes a Will roll, or a Will-based Throwing roll if that is better, at %d for the +%d%% ST shown, and costs 1 FP whether it succeeds or fails. A failure leaves the throw as it would be without it: %s for %s. A critical failure costs 1 HP of injury instead and the throw fails, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357)."),
			t.extraEffortPenalty, -5*t.extraEffortPenalty, distance, damage)
	} else {
		note = i18n.Text("Extra effort adds 5% to the ST the distance and damage are worked out from per -1 taken on a Will roll, or a Will-based Throwing roll if that is better, for 1 FP per attempt (B357).")
	}
	setNotes(t.notes, []string{note})
	t.distanceResult.MarkForLayoutRecursivelyUpward()
	t.damageResult.MarkForLayoutRecursivelyUpward()
}

// extraEffortST returns the ST with the given extra effort penalty applied, which adds 5% per -1 (BX357).
func extraEffortST(st fxp.Int, extraEffortPenalty int) fxp.Int {
	if extraEffortPenalty >= 0 {
		return st
	}
	return st.Mul(fxp.FromInteger(-5*extraEffortPenalty).Div(fxp.Hundred) + fxp.One).Floor()
}

// computeThrow returns the distance the object is thrown and the damage it does, as text, with the given extra effort
// penalty adding to the ST both are worked out from (BX355, BX357).
func (t *throwingCalculator) computeThrow(extraEffortPenalty int) (distance, damage string) {
	if t.objectWeight <= 0 {
		return i18n.Text("None"), i18n.Text("None")
	}
	entity := t.source.entity()
	distanceBonus := max(throwingTiers[t.throwingIndex].distance, throwingArtTiers[t.throwingArtIndex].distance)
	damageBonus := throwingArtTiers[t.throwingArtIndex].damage

	// Determine distance modifier based on weight ratio
	st := extraEffortST(t.st, extraEffortPenalty) + fxp.FromInteger(distanceBonus)
	basicLift := basicLiftFor(entity, st)
	var weightRatio fxp.Int
	if basicLift > 0 {
		weightRatio = fxp.Int(t.objectWeight).Div(fxp.Int(basicLift))
	}
	var modifier fxp.Int
	switch {
	case weightRatio <= fxp.Twentieth:
		modifier = fxp.ThreeAndAHalf
	case weightRatio <= fxp.Tenth:
		modifier = fxp.TwoAndAHalf
	case weightRatio <= fxp.PointOneFive:
		modifier = fxp.Two
	case weightRatio <= fxp.Fifth:
		modifier = fxp.OneAndAHalf
	case weightRatio <= fxp.Quarter:
		modifier = fxp.OnePointTwo
	case weightRatio <= fxp.ThreeTenths:
		modifier = fxp.OnePointOne
	case weightRatio <= fxp.TwoFifths:
		modifier = fxp.One
	case weightRatio <= fxp.Half:
		modifier = fxp.FourFifths
	case weightRatio <= fxp.ThreeQuarters:
		modifier = fxp.SevenTenths
	case weightRatio <= fxp.One:
		modifier = fxp.ThreeFifths
	case weightRatio <= fxp.OneAndAHalf:
		modifier = fxp.TwoFifths
	case weightRatio <= fxp.Two:
		modifier = fxp.ThreeTenths
	case weightRatio <= fxp.TwoAndAHalf:
		modifier = fxp.Quarter
	case weightRatio <= fxp.Three:
		modifier = fxp.Fifth
	case weightRatio <= fxp.Four:
		modifier = fxp.PointOneFive
	case weightRatio <= fxp.Five:
		modifier = fxp.PointOneTwo
	case weightRatio <= fxp.Six:
		modifier = fxp.Tenth
	case weightRatio <= fxp.Seven:
		modifier = fxp.PointZeroNine
	case weightRatio <= fxp.Eight:
		modifier = fxp.PointZeroEight
	case weightRatio <= fxp.Nine:
		modifier = fxp.PointZeroSeven
	case weightRatio <= fxp.Ten:
		modifier = fxp.PointZeroSix
	case weightRatio <= fxp.Twelve:
		modifier = fxp.Twentieth
	}
	inches := st.Mul(modifier).Mul(fxp.ThirtySix).Floor()
	if inches <= fxp.One {
		return i18n.Text("The object is too heavy to throw"), i18n.Text("None")
	}

	// Determine damage based on weight ratio
	thrust := thrustFor(entity, extraEffortST(t.strikingST, extraEffortPenalty))
	thrust.Modifier += thrust.Count * damageBonus
	basicLift = basicLiftFor(entity, st-fxp.FromInteger(distanceBonus))
	if basicLift > 0 {
		weightRatio = fxp.Int(t.objectWeight).Div(fxp.Int(basicLift))
	} else {
		weightRatio = 0
	}
	switch {
	case weightRatio <= fxp.Eighth:
		thrust.Modifier -= thrust.Count * 2
	case weightRatio <= fxp.Quarter:
		thrust.Modifier -= thrust.Count
	case weightRatio <= fxp.Half:
	case weightRatio <= fxp.One:
		thrust.Modifier += thrust.Count
	case weightRatio <= fxp.Two:
	case weightRatio <= fxp.Four:
		thrust.Modifier -= thrust.Count / 2
	default:
		thrust.Modifier -= thrust.Count
	}

	return lengthToText(entity, inches), gurps.FormatDice(thrust, gurps.SheetSettingsFor(entity).UseModifyingDicePlusAdds)
}
