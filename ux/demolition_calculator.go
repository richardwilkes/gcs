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

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

// The directions the demolition calculator works in, in the order its mode popup offers them.
const (
	explosiveForBlastMode = iota
	blastFromExplosiveMode
)

// demolitionDamageType is what an explosive charge inflicts: crushing damage with the Explosion modifier (BX415).
const demolitionDamageType = "cr ex"

var (
	_ calculatorTab = &demolitionCalculator{}

	demolitionModes = []demolitionMode{
		{name: i18n.Text("Explosive needed for a blast")},
		{name: i18n.Text("Blast from a quantity of explosive")},
	}

	explosiveChoices = newExplosiveChoices()

	// defaultExplosiveIndex is TNT's place among the choices. The Relative Explosive Force Table measures every other
	// explosive against TNT, so it is the one the calculator starts on.
	defaultExplosiveIndex = slices.IndexFunc(explosiveChoices, func(e explosiveChoice) bool { return e.title == "TNT" })
)

type demolitionMode struct {
	name string
}

func (m demolitionMode) String() string {
	return m.name
}

// explosiveChoice is a row of the Relative Explosive Force Table (BX415), or the entry whose REF is typed in.
type explosiveChoice struct {
	name   string  // What the popup shows: the tech level, the name and the REF.
	title  string  // The explosive's own name, which labels the weight it takes.
	ref    fxp.Int // Its relative explosive force, with TNT as 1.
	custom bool    // Whether the REF comes from the REF field rather than the table.
}

func (e explosiveChoice) String() string {
	return e.name
}

// newExplosiveChoices returns the Relative Explosive Force Table (BX415) as popup entries, with a custom one at the end
// for an explosive the table does not list.
func newExplosiveChoices() []explosiveChoice {
	choices := make([]explosiveChoice, 0, len(gurps.ExplosiveTypes)+1)
	for _, one := range gurps.ExplosiveTypes {
		choices = append(choices, explosiveChoice{
			name:  fmt.Sprintf(i18n.Text("TL%d %s (REF %s)"), one.TL, one.Name, one.REF.Comma()),
			title: one.Name,
			ref:   one.REF,
		})
	}
	return append(choices, explosiveChoice{name: i18n.Text("Custom"), title: i18n.Text("Explosive"), custom: true})
}

// demolitionCalculator converts between the size of a blast and the weight of explosive it takes (BX415). Everything it
// needs is typed in.
type demolitionCalculator struct {
	calculatorContent
	weightSlot            *unison.Panel
	weightRow             *unison.Panel
	damageResult          *unison.Label
	tntResult             *unison.Label
	explosiveWeightLabel  *textLabel
	explosiveWeightResult *unison.Label
	blastCountField       *IntegerField
	refField              *DecimalField
	explosiveWeightField  *WeightField
	customREF             fxp.Int
	explosiveWeight       fxp.Weight
	modeIndex             int
	explosiveIndex        int
	blastCount            int
}

func newDemolitionCalculator() *demolitionCalculator {
	d := &demolitionCalculator{
		customREF:       fxp.One,
		explosiveWeight: fxp.Weight(fxp.One),
		explosiveIndex:  defaultExplosiveIndex,
		blastCount:      1,
	}
	d.createContent()
	return d
}

// title implements calculatorTab.
func (d *demolitionCalculator) title() string {
	return i18n.Text("Demolition")
}

// panel implements calculatorTab.
func (d *demolitionCalculator) panel() *unison.Panel {
	return d.content
}

// preselect implements calculatorTab. Nothing here comes from a sheet.
func (d *demolitionCalculator) preselect(_ *Sheet) {
}

// sheetChanged implements calculatorTab. Nothing here comes from a sheet.
func (d *demolitionCalculator) sheetChanged(_ *Sheet) {
}

func (d *demolitionCalculator) createContent() {
	d.initCalculatorContent()
	d.content.AddChild(d.createHeader(i18n.Text("Demolition"), []linkSpec{{pageRef: "BX415", highlight: "Demolition"}}, 0))
	row := d.addRow(2)
	addPlainLabel(row, i18n.Text("Mode:"))
	addIndexPopup(row, demolitionModes, &d.modeIndex, d.changed)
	row = d.addRow(2)
	addPlainLabel(row, i18n.Text("Explosive:"))
	addIndexPopup(row, explosiveChoices, &d.explosiveIndex, d.changed)
	d.refField = sameWidth(NewDecimalField(nil, "", i18n.Text("Relative Explosive Force"),
		func() fxp.Int { return d.customREF },
		func(v fxp.Int) {
			d.customREF = v
			d.changed()
		},
		0, fxp.Max, false, false))
	d.addFieldRow(d.refField, i18n.Text("relative explosive force (REF), with TNT as 1"))
	d.blastCountField = sameWidth(NewIntegerField(nil, "", i18n.Text("Blast Multiplier"),
		func() int { return d.blastCount },
		func(v int) {
			d.blastCount = v
			d.changed()
		},
		0, 9999, false, false))
	d.addFieldRow(d.blastCountField, i18n.Text("n, where the blast is 6dxn"))
	d.explosiveWeightField = NewWeightField(nil, "", i18n.Text("Explosive Weight"), nil,
		func() fxp.Weight { return d.explosiveWeight },
		func(v fxp.Weight) {
			d.explosiveWeight = v
			d.changed()
		},
		0, fxp.Weight(fxp.Max), false)
	d.addFieldRow(d.explosiveWeightField, i18n.Text("of the explosive"))

	box := d.addResultsBox()
	row = box.addRow(2)
	addPlainLabel(row, i18n.Text("Damage:"))
	d.damageResult = addResultLabel(row)
	d.weightSlot = newRowGroup()
	box.content.AddChild(d.weightSlot)
	weightRows := &calculatorContent{content: d.weightSlot, flush: true}
	d.weightRow = weightRows.addRow(4)
	addPlainLabel(d.weightRow, i18n.Text("TNT:"))
	d.tntResult = addResultLabel(d.weightRow)
	d.explosiveWeightLabel = addPlainLabel(d.weightRow, "")
	d.explosiveWeightResult = addResultLabel(d.weightRow)

	d.addNotes(i18n.Text("Explosives normally do crushing damage with the Explosion modifier (B104), often with Fragmentation (B104)."))
}

// changed implements calculatorTab.
func (d *demolitionCalculator) changed() {
	d.adjustControls()
	d.updateResults()
}

// adjustControls enables, disables and blanks the fields to match the mode and the explosive: a control whose input
// would not be used is blanked as well as disabled, so that a stale value cannot be read as part of the answer.
func (d *demolitionCalculator) adjustControls() {
	adjustFieldBlank(d.refField, !explosiveChoices[d.explosiveIndex].custom)
	adjustFieldBlank(d.blastCountField, d.modeIndex != explosiveForBlastMode)
	adjustFieldBlank(d.explosiveWeightField, d.modeIndex != blastFromExplosiveMode)
	if d.modeIndex == explosiveForBlastMode {
		fillSlot(d.weightSlot, d.weightRow)
	} else {
		fillSlot(d.weightSlot)
	}
	d.content.MarkForLayoutRecursively()
	d.content.MarkForLayoutRecursivelyUpward()
	d.content.MarkForRedraw()
}

// updateResults recomputes the weight of explosive a blast takes, or the blast a weight of it makes (BX415).
func (d *demolitionCalculator) updateResults() {
	explosive := explosiveChoices[d.explosiveIndex]
	ref := explosive.ref
	if explosive.custom {
		ref = d.customREF
	}
	d.explosiveWeightLabel.SetTitle(fmt.Sprintf(i18n.Text("%s:"), explosive.title))
	if d.modeIndex == explosiveForBlastMode {
		n := fxp.FromInteger(d.blastCount)
		d.damageResult.SetTitle(blastDamageText(n))
		d.tntResult.SetTitle(explosiveWeightText(gurps.TNTForBlast(n)))
		d.explosiveWeightResult.SetTitle(explosiveWeightText(gurps.ExplosiveForBlast(n, ref)))
		return
	}
	d.damageResult.SetTitle(blastDamageText(gurps.BlastForExplosive(fxp.Int(d.explosiveWeight), ref)))
}

// blastDamageText describes the 6dxn blast the demolition rules measure an explosive by. A weight of explosive rarely
// works out to a whole number of multiples, so a fractional one is shown as it is, along with the dice it comes to.
func blastDamageText(n fxp.Int) string {
	if n <= 0 {
		return i18n.Text("None")
	}
	if n == n.Floor() {
		return gurps.FormatDice(dice.Dice{Count: 6, Sides: 6, Multiplier: n.AsInteger[int]()},
			gurps.SheetSettingsFor(nil).UseModifyingDicePlusAdds) + " " + demolitionDamageType
	}
	return fmt.Sprintf(i18n.Text("6dx%s (about %dd) %s"), n.Mul(fxp.Hundred).Round().Div(fxp.Hundred).Comma(),
		n.Mul(fxp.Six).Round().AsInteger[int](), demolitionDamageType)
}

// explosiveWeightText formats a weight of explosive, which the demolition rules always express in pounds, using the
// weight units the global sheet settings prefer.
func explosiveWeightText(pounds fxp.Int) string {
	return gurps.SheetSettingsFor(nil).DefaultWeightUnits.Format(fxp.Weight(pounds))
}
