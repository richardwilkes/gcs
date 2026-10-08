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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/encumbrance"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ calculatorTab = &jumpingCalculator{}

// jumpingCalculator works out how high and how far a character can jump (BX352). The character's numbers are either
// typed in or taken from any open character sheet, in which case the fields holding them are locked and refreshed
// whenever the sheet changes. The running start and the extra effort are always typed in, since they are a matter of
// circumstance rather than something a sheet knows.
type jumpingCalculator struct {
	calculatorContent
	source             sheetSourcePicker
	basicMoveField     *DecimalField
	liftingSTField     *DecimalField
	weightField        *WeightField
	encumbrancePopup   *unison.PopupMenu[encumbrance.Level]
	jumpingSkillField  *IntegerField
	enhancedMoveField  *DecimalField
	superJumpField     *DecimalField
	runningStartLabel  *textLabel
	highJumpResult     *unison.Label
	broadJumpResult    *unison.Label
	notes              *unison.Panel
	basicMove          fxp.Int
	liftingST          fxp.Int
	weight             fxp.Weight
	enhancedMove       fxp.Int
	superJump          fxp.Int
	runningStart       fxp.Int
	encumbranceIndex   int
	jumpingSkill       int
	extraEffortPenalty int
	updating           bool
}

func newJumpingCalculator() *jumpingCalculator {
	j := &jumpingCalculator{
		basicMove: fxp.Five,
		liftingST: fxp.Ten,
		weight:    fxp.Weight(fxp.FromInteger(150)),
	}
	j.createContent()
	return j
}

// title implements calculatorTab.
func (j *jumpingCalculator) title() string {
	return i18n.Text("Jumping")
}

// panel implements calculatorTab.
func (j *jumpingCalculator) panel() *unison.Panel {
	return j.content
}

// preselect implements calculatorTab.
func (j *jumpingCalculator) preselect(sheet *Sheet) {
	j.source.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (j *jumpingCalculator) sheetChanged(sheet *Sheet) {
	if j.source.sheet == sheet {
		j.changed()
	}
}

func (j *jumpingCalculator) createContent() {
	j.initCalculatorContent()
	j.content.AddChild(j.createHeader(i18n.Text("Jumping"), []linkSpec{{pageRef: "BX352", highlight: "Jumping"}}, 0))

	j.addSubheader(i18n.Text("Jumper"))
	j.source.selected = j.selectSheet
	j.source.refreshed = j.pullFromSheet
	j.source.changed = j.changed
	j.source.addRow(&j.calculatorContent, i18n.Text("Source:"))

	j.basicMoveField = sameWidth(NewDecimalField(nil, "", i18n.Text("Basic Move"),
		func() fxp.Int { return j.basicMove },
		func(v fxp.Int) {
			j.basicMove = v
			j.changed()
		},
		0, fxp.Max, false, false))
	j.addFieldRow(j.basicMoveField, i18n.Text("Basic Move"))
	j.liftingSTField = sameWidth(NewDecimalField(nil, "", i18n.Text("Lifting ST"),
		func() fxp.Int { return j.liftingST },
		func(v fxp.Int) {
			j.liftingST = v
			j.changed()
		},
		0, fxp.Max, false, false))
	j.addFieldRow(j.liftingSTField, i18n.Text("Lifting ST"))
	j.weightField = sameWidth(newSourcedWeightField(i18n.Text("Body Weight"), j.source.entity,
		func() fxp.Weight { return j.weight },
		func(v fxp.Weight) {
			j.weight = v
			j.changed()
		},
		0, fxp.Weight(fxp.Max)))
	j.addFieldRow(j.weightField, i18n.Text("body weight (a Basic Lift above it allows ST/4 as the Basic Move)"))
	row := j.addRow(2)
	addPlainLabel(row, i18n.Text("Encumbrance:"))
	j.encumbrancePopup = addIndexPopup(row, encumbrance.Levels, &j.encumbranceIndex, j.changed)
	j.jumpingSkillField = sameWidth(NewIntegerField(nil, "", i18n.Text("Jumping Skill"),
		func() int { return j.jumpingSkill },
		func(v int) {
			j.jumpingSkill = v
			j.changed()
		},
		0, 100, false, false))
	j.addFieldRow(j.jumpingSkillField, i18n.Text("Jumping skill level (0 for none)"))
	j.enhancedMoveField = sameWidth(NewDecimalField(nil, "", i18n.Text("Enhanced Move"),
		func() fxp.Int { return j.enhancedMove },
		func(v fxp.Int) {
			j.enhancedMove = v
			j.changed()
		},
		0, fxp.Max, false, false))
	j.addFieldRow(j.enhancedMoveField, i18n.Text("levels of Enhanced Move (Ground)"))
	j.superJumpField = sameWidth(NewDecimalField(nil, "", i18n.Text("Super Jump"),
		func() fxp.Int { return j.superJump },
		func(v fxp.Int) {
			j.superJump = v
			j.changed()
		},
		0, fxp.Max, false, false))
	j.addFieldRow(j.superJumpField, i18n.Text("levels of Super Jump"))

	j.addSubheader(i18n.Text("Jump"))
	j.runningStartLabel = j.addFieldRow(sameWidth(NewDecimalField(nil, "", i18n.Text("Jump Running Start"),
		func() fxp.Int { return j.runningStart },
		func(v fxp.Int) {
			j.runningStart = v
			j.changed()
		},
		0, fxp.Max, false, false)), "")
	j.addFieldRow(sameWidth(NewIntegerField(nil, "", i18n.Text("Jumping Extra Effort Penalty"),
		func() int { return j.extraEffortPenalty },
		func(v int) {
			j.extraEffortPenalty = v
			j.changed()
		},
		-100, 0, false, false)), i18n.Text("penalty taken for extra effort (+5% distance per -1)"))

	box := j.addResultsBox()
	row = box.addRow(2)
	addPlainLabel(row, i18n.Text("High Jump:"))
	j.highJumpResult = addResultLabel(row)
	addPlainLabel(row, i18n.Text("Broad Jump:"))
	j.broadJumpResult = addResultLabel(row)
	j.notes = box.addNotesGroup()
}

// selectSheet reads the jumper's numbers from the sheet the picker has just made its source, and does nothing at all
// when there is none.
func (j *jumpingCalculator) selectSheet(sheet *Sheet) {
	if sheet != nil {
		j.pullFromSheet()
	}
}

// pullFromSheet reads the jumper's numbers from its sheet. Each backing value is assigned before its field is synced,
// so that the setter the sync may run sees nothing new and does not start another round of updates.
func (j *jumpingCalculator) pullFromSheet() {
	jumper := calculator.JumperFromEntity(j.source.entity())
	j.basicMove = jumper.BasicMove
	j.liftingST = jumper.LiftingST
	j.weight = jumper.Weight
	j.encumbranceIndex = int(jumper.Encumbrance)
	j.jumpingSkill = jumper.JumpingSkill
	j.enhancedMove = jumper.EnhancedMove
	j.superJump = jumper.SuperJump
	j.basicMoveField.Sync()
	j.liftingSTField.Sync()
	j.weightField.Sync()
	j.encumbrancePopup.SelectIndex(j.encumbranceIndex)
	j.jumpingSkillField.Sync()
	j.enhancedMoveField.Sync()
	j.superJumpField.Sync()
}

// jumper returns the jumper the calculator's numbers describe.
func (j *jumpingCalculator) jumper() calculator.Jumper {
	return calculator.Jumper{
		Entity:       j.source.entity(),
		BasicMove:    j.basicMove,
		LiftingST:    j.liftingST,
		Weight:       j.weight,
		Encumbrance:  encumbrance.Level(j.encumbranceIndex),
		JumpingSkill: j.jumpingSkill,
		EnhancedMove: j.enhancedMove,
		SuperJump:    j.superJump,
	}
}

// changed implements calculatorTab.
func (j *jumpingCalculator) changed() {
	if j.updating {
		return
	}
	j.updating = true
	defer func() { j.updating = false }()
	j.source.refresh()
	j.adjustControls()
	j.updateResults()
}

// adjustControls locks the fields a sheet supplies and names the units the running start is measured in.
func (j *jumpingCalculator) adjustControls() {
	manual := j.source.sheet == nil
	j.basicMoveField.SetEnabled(manual)
	j.liftingSTField.SetEnabled(manual)
	j.weightField.SetEnabled(manual)
	j.encumbrancePopup.SetEnabled(manual)
	j.jumpingSkillField.SetEnabled(manual)
	j.enhancedMoveField.SetEnabled(manual)
	j.superJumpField.SetEnabled(manual)
	// The weight is shown in the units the source prefers, which may have just changed along with it.
	j.weightField.Sync()
	var units string
	if useMetersFor(j.source.entity()) {
		units = i18n.Text("meter")
	} else {
		units = i18n.Text("yard")
	}
	j.runningStartLabel.SetTitle(i18n.Text("%s running start", units))
	j.content.MarkForLayoutRecursively()
	j.content.MarkForLayoutRecursivelyUpward()
	j.content.MarkForRedraw()
}

// updateResults recomputes the jumps and rewrites the results, along with the note explaining the extra effort
// (BX357): what it takes, what it costs, and what a failure and a critical failure leave the jumper with.
func (j *jumpingCalculator) updateResults() {
	entity := j.source.entity()
	jumper := j.jumper()
	j.highJumpResult.SetTitle(lengthToText(entity, jumper.HighJump(j.runningStart, j.extraEffortPenalty)))
	j.broadJumpResult.SetTitle(lengthToText(entity, jumper.BroadJump(j.runningStart, j.extraEffortPenalty)))
	var note string
	if j.extraEffortPenalty < 0 {
		note = i18n.Text("The extra effort takes a Will roll, or a Will-based Jumping roll if that is better, at %d for the +%d%% shown, and costs 1 FP whether it succeeds or fails. A failure leaves the jump as it would be without it: %s high and %s broad. A critical failure costs 1 HP of injury to a foot or leg instead and the jump fails, and on a natural 18 a HT roll is needed as well to avoid a temporary Crippled Leg (B357).",
			j.extraEffortPenalty, -5*j.extraEffortPenalty, lengthToText(entity, jumper.HighJump(j.runningStart, 0)),
			lengthToText(entity, jumper.BroadJump(j.runningStart, 0)))
	} else {
		note = i18n.Text("Extra effort adds 5% to the distance per -1 taken on a Will roll, or a Will-based Jumping roll if that is better, for 1 FP per attempt (B357).")
	}
	setNotes(j.notes, []string{note})
	j.highJumpResult.MarkForLayoutRecursivelyUpward()
	j.broadJumpResult.MarkForLayoutRecursivelyUpward()
}
