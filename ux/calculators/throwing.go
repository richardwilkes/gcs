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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/ux"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ calculatorTab = &throwingCalculator{}

// throwingCalculator works out how far a character can throw an object and what it does when it lands (BX355). The
// character's numbers are either typed in or taken from any open character sheet, in which case the fields holding
// them are locked and refreshed whenever the sheet changes. The object and the extra effort are always typed in.
type throwingCalculator struct {
	calculatorContent
	source             sheetSourcePicker
	stField            *ux.DecimalField
	strikingSTField    *ux.DecimalField
	throwingPopup      *unison.PopupMenu[calculator.ThrowingTier]
	throwingArtPopup   *unison.PopupMenu[calculator.ThrowingTier]
	weightField        *ux.WeightField
	distanceResult     *unison.Label
	damageResult       *unison.Label
	notes              *unison.Panel
	throwingTiers      []calculator.ThrowingTier
	throwingArtTiers   []calculator.ThrowingTier
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
		throwingTiers:    calculator.ThrowingTiers(),
		throwingArtTiers: calculator.ThrowingArtTiers(),
		st:               fxp.Ten,
		strikingST:       fxp.Ten,
		objectWeight:     fxp.Weight(fxp.One),
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
func (t *throwingCalculator) preselect(sheet *ux.Sheet) {
	t.source.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (t *throwingCalculator) sheetChanged(sheet *ux.Sheet) {
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

	t.stField = sameWidth(ux.NewDecimalField(nil, "", i18n.Text("ST"),
		func() fxp.Int { return t.st },
		func(v fxp.Int) {
			t.st = v
			t.changed()
		},
		0, fxp.Max, false, false))
	t.addFieldRow(t.stField, i18n.Text("ST the distance is worked out from"))
	t.strikingSTField = sameWidth(ux.NewDecimalField(nil, "", i18n.Text("Striking ST"),
		func() fxp.Int { return t.strikingST },
		func(v fxp.Int) {
			t.strikingST = v
			t.changed()
		},
		0, fxp.Max, false, false))
	t.addFieldRow(t.strikingSTField, i18n.Text("Striking ST the damage is worked out from"))
	row := t.addRow(2)
	addPlainLabel(row, i18n.Text("Throwing:"))
	t.throwingPopup = addIndexPopup(row, t.throwingTiers, &t.throwingIndex, t.changed)
	row = t.addRow(2)
	addPlainLabel(row, i18n.Text("Throwing Art:"))
	t.throwingArtPopup = addIndexPopup(row, t.throwingArtTiers, &t.throwingArtIndex, t.changed)

	t.addSubheader(i18n.Text("Throw"))
	t.weightField = sameWidth(newSourcedWeightField(i18n.Text("Object Weight"), t.source.entity,
		func() fxp.Weight { return t.objectWeight },
		func(v fxp.Weight) {
			t.objectWeight = v
			t.changed()
		},
		0, fxp.Weight(fxp.Max)))
	t.addFieldRow(t.weightField, i18n.Text("object"))
	t.addFieldRow(sameWidth(ux.NewIntegerField(nil, "", i18n.Text("Throwing Extra Effort Penalty"),
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
func (t *throwingCalculator) selectSheet(sheet *ux.Sheet) {
	if sheet != nil {
		t.pullFromSheet()
	}
}

// pullFromSheet reads the thrower's numbers from its sheet. Each backing value is assigned before its field is synced,
// so that the setter the sync may run sees nothing new and does not start another round of updates.
func (t *throwingCalculator) pullFromSheet() {
	stats := calculator.ThrowerStatsFromEntity(t.source.entity())
	t.st = stats.ST
	t.strikingST = stats.StrikingST
	t.throwingIndex = stats.ThrowingIndex
	t.throwingArtIndex = stats.ThrowingArtIndex
	t.stField.Sync()
	t.strikingSTField.Sync()
	t.throwingPopup.SelectIndex(t.throwingIndex)
	t.throwingArtPopup.SelectIndex(t.throwingArtIndex)
}

// thrower returns the thrower the calculator's numbers describe.
func (t *throwingCalculator) thrower() calculator.Thrower {
	return calculator.Thrower{
		Entity:      t.source.entity(),
		ST:          t.st,
		StrikingST:  t.strikingST,
		Throwing:    t.throwingTiers[t.throwingIndex],
		ThrowingArt: t.throwingArtTiers[t.throwingArtIndex],
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
	distance, damage := t.throwText(t.extraEffortPenalty)
	t.distanceResult.SetTitle(distance)
	t.damageResult.SetTitle(damage)
	var note string
	if t.extraEffortPenalty < 0 {
		distance, damage = t.throwText(0)
		note = i18n.Text("The extra effort takes a Will roll, or a Will-based Throwing roll if that is better, at %d for the +%d%% ST shown, and costs 1 FP whether it succeeds or fails. A failure leaves the throw as it would be without it: %s for %s. A critical failure costs 1 HP of injury instead and the throw fails, and on a natural 18 a HT roll is needed as well to avoid a temporary disadvantage (B357).",
			t.extraEffortPenalty, -5*t.extraEffortPenalty, distance, damage)
	} else {
		note = i18n.Text("Extra effort adds 5% to the ST the distance and damage are worked out from per -1 taken on a Will roll, or a Will-based Throwing roll if that is better, for 1 FP per attempt (B357).")
	}
	setNotes(t.notes, []string{note})
	t.distanceResult.MarkForLayoutRecursivelyUpward()
	t.damageResult.MarkForLayoutRecursivelyUpward()
}

// throwText returns the distance the object is thrown and the damage it does, as text, with the given extra effort
// penalty adding to the ST both are worked out from (BX355, BX357).
func (t *throwingCalculator) throwText(extraEffortPenalty int) (distance, damage string) {
	thrower := t.thrower()
	result := thrower.Throw(t.objectWeight, extraEffortPenalty)
	switch {
	case result.TooHeavy:
		return i18n.Text("The object is too heavy to throw"), i18n.Text("None")
	case result.Distance <= 0:
		return i18n.Text("None"), i18n.Text("None")
	default:
		entity := t.source.entity()
		return lengthToText(entity, result.Distance),
			gurps.FormatDice(result.Damage, gurps.SheetSettingsFor(entity).UseModifyingDicePlusAdds)
	}
}
