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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

var _ calculatorTab = &collisionCalculator{}

// The scenarios the collision calculator handles, in the order the scenario popup offers them.
const (
	fallScenario = iota
	immovableScenario
	twoObjectScenario
	suddenStopScenario
)

// newCollisionScenarios returns the scenarios the calculator handles, in the order the scenario constants give them.
func newCollisionScenarios() []collisionScenario {
	return []collisionScenario{
		{name: i18n.Text("Fall onto a surface")},
		{name: i18n.Text("Collision with an immovable object")},
		{name: i18n.Text("Collision between two objects")},
		{name: i18n.Text("Sudden stop of a vehicle, elevator, etc.")},
	}
}

type collisionScenario struct {
	name string
}

func (s collisionScenario) String() string {
	return s.name
}

// collisionCalculator works out the damage from a collision or a fall (BX430-BX431). Each of the objects involved
// either has its numbers typed in or takes them from any open character sheet.
type collisionCalculator struct {
	calculatorContent
	moverHeader         *unison.Label
	sectionSlot         *unison.Panel
	targetSlot          *unison.Panel
	targetHeader        *unison.Label
	fallRows            *unison.Panel
	surfaceRows         *unison.Panel
	dropRows            *unison.Panel
	angleRows           *unison.Panel
	restraintRows       *unison.Panel
	results             *unison.Panel
	notes               *unison.Panel
	fallDistanceField   *DecimalField
	gravityField        *DecimalField
	terminalPopup       *unison.PopupMenu[calculator.TerminalVelocityChoice]
	customTerminalField *DecimalField
	pressureField       *DecimalField
	controlledFallBox   *unison.CheckBox
	elasticDRField      *IntegerField
	breakableBox        *unison.CheckBox
	obstacleHPField     *DecimalField
	obstacleDRField     *IntegerField
	cleanDiveBox        *unison.CheckBox
	scenarios           []collisionScenario
	shapes              []calculator.CollisionShape
	surfaces            []calculator.CollisionSurface
	terminals           []calculator.TerminalVelocityChoice
	angles              []calculator.CollisionAngleChoice
	restraints          []calculator.CollisionRestraint
	mover               collisionParticipant
	target              collisionParticipant
	fallDistance        fxp.Int
	gravity             fxp.Int
	pressure            fxp.Int
	customTerminal      fxp.Int
	obstacleHP          fxp.Int
	moverName           string
	scenarioIndex       int
	terminalIndex       int
	surfaceIndex        int
	angleIndex          int
	restraintIndex      int
	elasticDR           int
	obstacleDR          int
	controlledFall      bool
	cleanDive           bool
	breakable           bool
	dropped             bool
	updating            bool
}

// collisionParticipant is one of the objects in a collision: the one doing the moving, or the one it hits. Its numbers
// are either typed in or taken from an open character sheet, in which case the fields holding them are locked and
// refreshed whenever the sheet changes. The velocity is the exception: a sheet only suggests it (the character's Move),
// since how fast something was going is a matter of circumstance.
type collisionParticipant struct {
	sheetSourcePicker
	calc            *collisionCalculator
	panel           *unison.Panel
	extras          *unison.Panel
	hpField         *DecimalField
	stField         *DecimalField
	smField         *IntegerField
	velocityField   *DecimalField
	velocityLabel   *textLabel
	acrobaticsField *IntegerField
	swimmingField   *IntegerField
	armorDRField    *IntegerField
	innateDRField   *IntegerField
	hp              fxp.Int
	st              fxp.Int
	velocity        fxp.Int
	sm              int
	acrobatics      int
	swimming        int
	armorDR         int
	innateDR        int
	shapeIndex      int
}

func newCollisionCalculator() *collisionCalculator {
	c := &collisionCalculator{
		scenarios:      newCollisionScenarios(),
		shapes:         calculator.CollisionShapes(),
		surfaces:       calculator.CollisionSurfaces(),
		terminals:      calculator.TerminalVelocityChoices(),
		angles:         calculator.CollisionAngleChoices(),
		restraints:     calculator.CollisionRestraints(),
		fallDistance:   fxp.Five,
		gravity:        fxp.One,
		pressure:       fxp.One,
		customTerminal: fxp.FromInteger(200),
		elasticDR:      5,
	}
	c.mover = collisionParticipant{calc: c, hp: fxp.Ten, velocity: fxp.Five}
	c.target = collisionParticipant{calc: c, hp: fxp.Ten}
	c.createContent()
	return c
}

// title implements calculatorTab.
func (c *collisionCalculator) title() string {
	return i18n.Text("Collisions & Falls")
}

// panel implements calculatorTab.
func (c *collisionCalculator) panel() *unison.Panel {
	return c.content
}

// preselect implements calculatorTab. The sheet becomes the moving object's source.
func (c *collisionCalculator) preselect(sheet *Sheet) {
	c.mover.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (c *collisionCalculator) sheetChanged(sheet *Sheet) {
	if c.mover.sheet == sheet || c.target.sheet == sheet {
		c.changed()
	}
}

func (c *collisionCalculator) createContent() {
	c.initCalculatorContent()
	c.content.AddChild(c.createHeader(i18n.Text("Collisions & Falls"),
		[]linkSpec{{pageRef: "BX430", highlight: "Collisions and Falls"}}, 0))

	row := c.addRow(2)
	addPlainLabel(row, i18n.Text("Scenario:"))
	addIndexPopup(row, c.scenarios, &c.scenarioIndex, c.changed)

	c.moverHeader = c.addSubheader("")
	c.mover.createPanel(c.content)

	c.sectionSlot = newRowGroup()
	c.content.AddChild(c.sectionSlot)
	c.fallRows = c.createFallRows()
	c.surfaceRows = c.createSurfaceRows()
	c.dropRows = c.createDropRows()
	c.angleRows = c.createAngleRows()
	c.restraintRows = c.createRestraintRows()

	c.targetSlot = newRowGroup()
	c.content.AddChild(c.targetSlot)
	c.targetHeader = newSubheader(i18n.Text("Struck object"))
	c.target.createPanel(c.targetSlot)
	c.targetSlot.RemoveAllChildren()

	c.results, c.notes = c.addResultsSection()
}

// createPanel builds the participant's rows into a panel of its own, added to the parent.
func (p *collisionParticipant) createPanel(parent *unison.Panel) {
	p.panel = newRowGroup()
	parent.AddChild(p.panel)
	rows := &calculatorContent{content: p.panel}

	p.selected = p.selectSheet
	p.refreshed = p.pullFromSheet
	p.changed = p.calc.changed
	p.addRow(rows, i18n.Text("Source:"))

	p.hpField = sameWidth(NewDecimalField(nil, "", i18n.Text("Hit Points"),
		func() fxp.Int { return p.hp },
		func(v fxp.Int) {
			p.hp = v
			p.calc.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(p.hpField, i18n.Text("HP"))
	p.stField = sameWidth(NewDecimalField(nil, "", i18n.Text("Strength"),
		func() fxp.Int { return p.st },
		func(v fxp.Int) {
			p.st = v
			p.calc.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(p.stField, i18n.Text("ST (0 if it has no ST score)"))
	p.smField = sameWidth(NewIntegerField(nil, "", i18n.Text("Size Modifier"),
		func() int { return p.sm },
		func(v int) {
			p.sm = v
			p.calc.changed()
		},
		-100, 100, true, false))
	rows.addFieldRow(p.smField, i18n.Text("SM"))
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Shape:"))
	addIndexPopup(row, p.calc.shapes, &p.shapeIndex, p.calc.changed)
	p.velocityField = sameWidth(NewDecimalField(nil, "", i18n.Text("Velocity"),
		func() fxp.Int { return p.velocity },
		func(v fxp.Int) {
			p.velocity = v
			p.calc.changed()
		},
		0, fxp.Max, false, false))
	p.velocityLabel = rows.addFieldRow(p.velocityField, "")

	p.extras = newRowGroup()
	p.panel.AddChild(p.extras)
	rows = &calculatorContent{content: p.extras}
	p.acrobaticsField = sameWidth(NewIntegerField(nil, "", i18n.Text("Acrobatics"),
		func() int { return p.acrobatics },
		func(v int) {
			p.acrobatics = v
			p.calc.changed()
		},
		0, 100, false, false))
	rows.addFieldRow(p.acrobaticsField, i18n.Text("Acrobatics skill level"))
	p.swimmingField = sameWidth(NewIntegerField(nil, "", i18n.Text("Swimming"),
		func() int { return p.swimming },
		func(v int) {
			p.swimming = v
			p.calc.changed()
		},
		0, 100, false, false))
	rows.addFieldRow(p.swimmingField, i18n.Text("Swimming skill level"))
	p.armorDRField = sameWidth(NewIntegerField(nil, "", i18n.Text("Armor DR"),
		func() int { return p.armorDR },
		func(v int) {
			p.armorDR = v
			p.calc.changed()
		},
		0, 10000, false, false))
	rows.addFieldRow(p.armorDRField, i18n.Text("DR from armor (counts as flexible against a fall)"))
	p.innateDRField = sameWidth(NewIntegerField(nil, "", i18n.Text("Innate DR"),
		func() int { return p.innateDR },
		func(v int) {
			p.innateDR = v
			p.calc.changed()
		},
		0, 10000, false, false))
	rows.addFieldRow(p.innateDRField, i18n.Text("innate DR (does not count as flexible)"))
}

// showExtras adds or removes the rows that a two-object collision does not use.
func (p *collisionParticipant) showExtras(show bool) {
	if show == (p.extras.Parent() != nil) {
		return
	}
	if show {
		p.panel.AddChild(p.extras)
	} else {
		p.extras.RemoveFromParent()
	}
}

// selectSheet reads the participant's numbers from the sheet the picker has just made its source, and does nothing at
// all when there is none. Choosing a sheet also suggests its Move as the velocity; it is only a suggestion, so a later
// refresh leaves the velocity alone.
func (p *collisionParticipant) selectSheet(sheet *Sheet) {
	if sheet == nil {
		return
	}
	values := calculator.CollisionParticipantFromEntity(sheet.Entity())
	p.velocity = values.Velocity
	p.velocityField.Sync()
	p.apply(values)
}

// pullFromSheet reads the participant's numbers from its sheet.
func (p *collisionParticipant) pullFromSheet() {
	p.apply(calculator.CollisionParticipantFromEntity(p.sheet.Entity()))
}

// apply stores what a sheet supplies, leaving an attribute the sheet does not define as it was, and syncs the fields.
// Each backing value is assigned before its field is synced, so that the setter the sync may run sees nothing new and
// does not start another round of updates.
func (p *collisionParticipant) apply(values calculator.CollisionParticipant) {
	if values.HasHP {
		p.hp = values.HP
	}
	if values.HasST {
		p.st = values.ST
	}
	p.sm = values.SM
	p.acrobatics = values.Acrobatics
	p.swimming = values.Swimming
	p.armorDR = values.ArmorDR
	p.innateDR = values.InnateDR
	p.hpField.Sync()
	p.stField.Sync()
	p.smField.Sync()
	p.acrobaticsField.Sync()
	p.swimmingField.Sync()
	p.armorDRField.Sync()
	p.innateDRField.Sync()
}

// lockSheetFields enables the fields whose values are typed in and disables those that a sheet supplies.
func (p *collisionParticipant) lockSheetFields() {
	manual := p.sheet == nil
	p.hpField.SetEnabled(manual)
	p.stField.SetEnabled(manual)
	p.smField.SetEnabled(manual)
	p.acrobaticsField.SetEnabled(manual)
	p.swimmingField.SetEnabled(manual)
	p.armorDRField.SetEnabled(manual)
	p.innateDRField.SetEnabled(manual)
}

// name returns what to call the participant in the results: its sheet's title, or the role it plays.
func (p *collisionParticipant) name(role string) string {
	if p.sheet != nil {
		return p.sheet.String()
	}
	return role
}

func (p *collisionParticipant) shape() calculator.CollisionShape {
	return p.calc.shapes[p.shapeIndex]
}

func (c *collisionCalculator) createFallRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	c.fallDistanceField = sameWidth(NewDecimalField(nil, "", i18n.Text("Distance Fallen"),
		func() fxp.Int { return c.fallDistance },
		func(v fxp.Int) {
			c.fallDistance = v
			c.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(c.fallDistanceField, i18n.Text("yards fallen"))
	c.gravityField = sameWidth(NewDecimalField(nil, "", i18n.Text("Gravity"),
		func() fxp.Int { return c.gravity },
		func(v fxp.Int) {
			c.gravity = v
			c.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(c.gravityField, i18n.Text("gravity, in Gs"))
	row := rows.addRow(4)
	addPlainLabel(row, i18n.Text("Terminal velocity:"))
	c.terminalPopup = addIndexPopup(row, c.terminals, &c.terminalIndex, c.changed)
	c.customTerminalField = sameWidth(NewDecimalField(nil, "", i18n.Text("Custom Terminal Velocity"),
		func() fxp.Int { return c.customTerminal },
		func(v fxp.Int) {
			c.customTerminal = v
			c.changed()
		},
		0, fxp.Max, false, false))
	row.AddChild(c.customTerminalField)
	addPlainLabel(row, i18n.Text("yards/second"))
	c.pressureField = sameWidth(NewDecimalField(nil, "", i18n.Text("Atmospheric Pressure"),
		func() fxp.Int { return c.pressure },
		func(v fxp.Int) {
			c.pressure = v
			c.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(c.pressureField, i18n.Text("atmospheres of pressure (0 in a vacuum)"))
	c.controlledFallBox = rows.addCheckBox("", &c.controlledFall, c.changed)
	return group
}

func (c *collisionCalculator) createSurfaceRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Surface:"))
	addIndexPopup(row, c.surfaces, &c.surfaceIndex, c.changed)
	c.elasticDRField = sameWidth(NewIntegerField(nil, "", i18n.Text("Elastic DR"),
		func() int { return c.elasticDR },
		func(v int) {
			c.elasticDR = v
			c.changed()
		},
		0, 100, false, false))
	rows.addFieldRow(c.elasticDRField, i18n.Text("DR from the elastic surface (2 for a feather bed, 10 for a net or airbag)"))
	c.cleanDiveBox = rows.addCheckBox("", &c.cleanDive, c.changed)
	c.breakableBox = rows.addCheckBox(i18n.Text("The surface can break"), &c.breakable, c.changed)
	row = rows.addRow(4)
	c.obstacleHPField = sameWidth(NewDecimalField(nil, "", i18n.Text("Obstacle HP"),
		func() fxp.Int { return c.obstacleHP },
		func(v fxp.Int) {
			c.obstacleHP = v
			c.changed()
		},
		0, fxp.Max, false, false))
	row.AddChild(c.obstacleHPField)
	addPlainLabel(row, i18n.Text("HP"))
	c.obstacleDRField = sameWidth(NewIntegerField(nil, "", i18n.Text("Obstacle DR"),
		func() int { return c.obstacleDR },
		func(v int) {
			c.obstacleDR = v
			c.changed()
		},
		0, 10000, false, false))
	// The "HP" caption that closes the field before this one would otherwise be taken as this field's name.
	c.obstacleDRField.Accessibility.Name = i18n.Text("Obstacle DR")
	row.AddChild(c.obstacleDRField)
	addPlainLabel(row, i18n.Text("DR of the breakable surface"))
	return group
}

func (c *collisionCalculator) createDropRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	rows.addCheckBox(i18n.Text("The striking object was dropped, and its velocity comes from the fall"), &c.dropped,
		c.changed)
	return group
}

func (c *collisionCalculator) createAngleRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Collision angle:"))
	addIndexPopup(row, c.angles, &c.angleIndex, c.changed)
	return group
}

func (c *collisionCalculator) createRestraintRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Restraint:"))
	addIndexPopup(row, c.restraints, &c.restraintIndex, c.changed)
	return group
}

// changed implements calculatorTab.
func (c *collisionCalculator) changed() {
	if c.updating {
		return
	}
	c.updating = true
	defer func() { c.updating = false }()
	c.mover.refresh()
	c.target.refresh()
	c.adjustControls()
	c.updateResults()
}

// adjustControls shows the rows the scenario needs, locks the fields a sheet supplies and enables, disables and
// retitles the rest to match the current choices.
func (c *collisionCalculator) adjustControls() {
	var groups []*unison.Panel
	var moverRole string
	showTarget := false
	switch c.scenarioIndex {
	case fallScenario:
		moverRole = i18n.Text("Faller")
		c.moverName = i18n.Text("the faller")
		groups = []*unison.Panel{c.fallRows, c.surfaceRows}
	case immovableScenario:
		moverRole = i18n.Text("Moving object")
		c.moverName = i18n.Text("the moving object")
		groups = []*unison.Panel{c.surfaceRows}
	case twoObjectScenario:
		moverRole = i18n.Text("Striking object")
		c.moverName = i18n.Text("the striking object")
		groups = []*unison.Panel{c.dropRows, c.fallRows, c.angleRows}
		showTarget = true
	default:
		moverRole = i18n.Text("Occupant")
		c.moverName = i18n.Text("the occupant")
		groups = []*unison.Panel{c.restraintRows}
	}
	c.moverHeader.SetTitle(moverRole)
	fillSlot(c.sectionSlot, groups...)
	if showTarget {
		fillSlot(c.targetSlot, c.targetHeader.AsPanel(), c.target.panel)
	} else {
		fillSlot(c.targetSlot)
	}
	c.mover.showExtras(c.scenarioIndex != twoObjectScenario)
	c.target.showExtras(false)
	c.mover.lockSheetFields()
	c.target.lockSheetFields()

	fromFall := c.velocityFromFall()
	c.mover.velocityField.SetEnabled(!fromFall)
	switch {
	case fromFall:
		c.mover.velocityLabel.SetTitle(i18n.Text("yards/second, from the fall"))
	case c.scenarioIndex == suddenStopScenario:
		c.mover.velocityLabel.SetTitle(i18n.Text("yards/second lost in the stop"))
	default:
		c.mover.velocityLabel.SetTitle(i18n.Text("yards/second"))
	}
	c.target.velocityLabel.SetTitle(i18n.Text("yards/second"))
	// A control whose input would not be used is blanked as well as disabled, so that a stale value cannot be read as
	// part of the answer. The moving object's velocity is the exception: when it comes from the fall, the field shows
	// the velocity the fall reaches, which is in use.
	adjustFieldBlank(c.target.velocityField, c.angles[c.angleIndex].Angle == calculator.SideOnCollision)

	// Unison does not disable a panel's children along with it, so the fall rows are switched one by one when the
	// striking object in a two-object collision is not something that was dropped.
	adjustFieldBlank(c.fallDistanceField, !fromFall)
	adjustFieldBlank(c.gravityField, !fromFall)
	adjustPopupBlank(c.terminalPopup, !fromFall)
	terminal := c.terminals[c.terminalIndex]
	adjustFieldBlank(c.customTerminalField, !fromFall || !terminal.Custom)
	adjustFieldBlank(c.pressureField, !fromFall || (terminal.BaseVelocity <= 0 && !terminal.Custom))
	// The fall can be softened by an Acrobatics roll or by a clean dive into water, but not by both (BX431).
	c.controlledFallBox.SetTitle(i18n.Text("Made a successful Acrobatics roll for a controlled fall (-5 yards)"))
	c.controlledFallBox.SetEnabled(c.scenarioIndex == fallScenario && !c.diving())
	c.cleanDiveBox.SetTitle(i18n.Text("Made a successful Swimming roll at %d for a clean dive",
		gurps.SpeedRangePenalty(c.moverVelocity())))
	c.cleanDiveBox.SetEnabled(c.inWater() && !c.controlledFallApplies())

	surface := c.surfaces[c.surfaceIndex]
	adjustFieldBlank(c.elasticDRField, !surface.Elastic)
	adjustFieldBlank(c.obstacleHPField, !c.breakable)
	adjustFieldBlank(c.obstacleDRField, !c.breakable)

	c.content.MarkForLayoutRecursively()
	c.content.MarkForLayoutRecursivelyUpward()
	c.content.MarkForRedraw()
}

// velocityFromFall reports whether the moving object's velocity is the one it reaches in a fall rather than one that
// is typed in.
func (c *collisionCalculator) velocityFromFall() bool {
	return c.scenarioIndex == fallScenario || (c.scenarioIndex == twoObjectScenario && c.dropped)
}

// inWater reports whether the moving object lands in water, where a clean dive is possible.
func (c *collisionCalculator) inWater() bool {
	return (c.scenarioIndex == fallScenario || c.scenarioIndex == immovableScenario) &&
		c.surfaces[c.surfaceIndex].Water
}

// controlledFallApplies reports whether the Acrobatics roll shortens the fall, which only a fall allows.
func (c *collisionCalculator) controlledFallApplies() bool {
	return c.scenarioIndex == fallScenario && c.controlledFall
}

// diving reports whether a clean dive negates the damage: only in water, and not when the faller chose a controlled
// fall instead, since the rules allow one or the other (BX431).
func (c *collisionCalculator) diving() bool {
	return c.inWater() && c.cleanDive && !c.controlledFallApplies()
}

// moverVelocity returns the velocity the moving object hits at: its own, or the one it reaches in a fall.
func (c *collisionCalculator) moverVelocity() fxp.Int {
	if c.velocityFromFall() {
		return c.fall().Velocity
	}
	return c.mover.velocity
}

// fall works out the fall the moving object makes, with the terminal velocity the chosen limit allows.
func (c *collisionCalculator) fall() calculator.FallResult {
	choice := c.terminals[c.terminalIndex]
	base := choice.BaseVelocity
	if choice.Custom {
		base = c.customTerminal
	}
	return calculator.Fall(c.fallDistance, c.gravity, base, c.pressure, c.controlledFallApplies())
}

// fallNotes returns the notes explaining anything that shortened or limited the fall.
func fallNotes(fall calculator.FallResult) []string {
	var notes []string
	if fall.Controlled {
		notes = append(notes, i18n.Text("The controlled fall counts as a fall of %s yards (BX431).", fall.Distance.Comma()))
	}
	switch {
	case fall.NoGravity:
		notes = append(notes, i18n.Text("Without gravity, there is no fall."))
	case fall.Vacuum:
		notes = append(notes, i18n.Text("In a vacuum there is no terminal velocity (BX431)."))
	case fall.TerminalLimited:
		notes = append(notes, i18n.Text("The fall is limited to the terminal velocity of %s yards/second (BX431).",
			fall.Terminal.Comma()))
	}
	return notes
}

// updateResults recomputes the damage and rewrites the results and notes.
func (c *collisionCalculator) updateResults() {
	c.results.RemoveAllChildren()
	var notes []string
	if c.scenarioIndex == twoObjectScenario {
		notes = c.updateTwoObjectResults()
	} else {
		notes = c.updateSurfaceResults()
	}
	setNotes(c.notes, notes)
	c.results.MarkForLayoutRecursivelyUpward()
	c.results.MarkForRedraw()
}

func (c *collisionCalculator) addResult(label, value string) {
	addResult(c.results, label, value)
}

// velocityText describes a velocity in yards per second and miles per hour (2 mph is 1 yard/second, BX430).
func velocityText(velocity fxp.Int) string {
	return i18n.Text("%s yards/second (%s mph)", velocity.Comma(), velocity.Mul(fxp.Two).Comma())
}

// damageText formats dice of the given type, using the sheet's dice notation when the numbers came from a sheet.
func damageText(entity *gurps.Entity, count fxp.Int, damageType string) string {
	if count <= 0 {
		return i18n.Text("None")
	}
	return gurps.FormatDice(calculator.CollisionDamageDice(count), gurps.SheetSettingsFor(entity).UseModifyingDicePlusAdds) +
		" " + damageType
}

// updateSurfaceResults handles the scenarios where the moving object hits something immovable: a fall, a collision with
// an obstacle, and an occupant's sudden stop, which is a fall at the velocity lost.
func (c *collisionCalculator) updateSurfaceResults() []string {
	var notes []string
	velocity := c.mover.velocity
	if c.scenarioIndex == fallScenario {
		fall := c.fall()
		velocity, notes = fall.Velocity, fallNotes(fall)
		c.mover.velocity = velocity
		c.mover.velocityField.Sync()
	}
	surface := c.surfaces[c.surfaceIndex]
	if c.scenarioIndex == suddenStopScenario {
		surface = c.surfaces[0]
	}
	shape := c.mover.shape()
	count := calculator.ImmovableCollisionDice(calculator.CollisionObject{
		HP:         c.mover.hp,
		Velocity:   velocity,
		HalfDamage: shape.HalfDamage,
	}, surface.Hard)
	entity := c.mover.entity()
	mover := c.mover.name(c.moverName)
	c.addResult(i18n.Text("Velocity:"), velocityText(velocity))
	if c.diving() {
		c.addResult(i18n.Text("Damage to %s:", mover), i18n.Text("None"))
		notes = append(notes, i18n.Text("The clean dive negates all damage (BX431)."))
		return notes
	}
	damage := damageText(entity, count, shape.DamageType)
	c.addResult(i18n.Text("Damage to %s:", mover), damage)
	if c.scenarioIndex != suddenStopScenario {
		c.addResult(i18n.Text("Damage to the surface:"), damage)
	}
	if surface.Hard {
		notes = append(notes, i18n.Text("The surface is hard, so the damage is worked out with twice the HP (BX431)."))
	}
	if c.scenarioIndex != suddenStopScenario && c.breakable {
		notes = append(notes, i18n.Text("The surface can break, so neither side takes more than %s points, its HP + DR (BX431).",
			(c.obstacleHP+fxp.FromInteger(c.obstacleDR)).Comma()))
	}
	if surface.Elastic {
		notes = append(notes, i18n.Text("The elastic surface gives DR %d against this damage (BX431).", c.elasticDR))
	}
	if c.inWater() {
		notes = append(notes, c.swimmingNote())
	}
	if c.scenarioIndex == suddenStopScenario {
		if dr := c.restraints[c.restraintIndex].DR; dr > 0 {
			notes = append(notes, i18n.Text("The restraint gives DR %d against this damage (BX432).", dr))
		}
		notes = append(notes, i18n.Text("Anyone not strapped into an open vehicle is also thrown; work out knockback from this damage to see how far (BX432)."))
	}
	if c.scenarioIndex != immovableScenario {
		notes = append(notes, c.armorNote(count))
		if c.scenarioIndex == fallScenario {
			notes = append(notes, i18n.Text("Roll randomly for the hit location. Injury to a limb or extremity in excess of what cripples it is not ignored; if a limb is crippled, roll 1d, and on 5-6 all limbs of that type are crippled (BX431)."))
		}
	}
	return notes
}

// swimmingNote describes the Swimming roll that would have made a clean dive.
func (c *collisionCalculator) swimmingNote() string {
	penalty := gurps.SpeedRangePenalty(c.moverVelocity())
	roll := i18n.Text("Swimming roll")
	if c.scenarioIndex == immovableScenario {
		roll = i18n.Text("Swimming roll (or vehicle control roll, when ditching a vehicle)")
	}
	if c.mover.swimming > 0 {
		return i18n.Text("A successful %s at %d (effective skill %d) would be a clean dive that negates all damage (BX431).",
			roll, penalty, c.mover.swimming+penalty)
	}
	return i18n.Text("A successful %s at %d would be a clean dive that negates all damage (BX431).", roll, penalty)
}

// armorNote describes how the mover's armor fares against falling damage: all of it counts as flexible, so it lets 1 HP
// of injury through for every 5 full points it stops, even when it stops all of it (BX431). The most it can stop is its
// own DR, which bounds the blunt trauma.
func (c *collisionCalculator) armorNote(count fxp.Int) string {
	if count <= 0 {
		return ""
	}
	if c.mover.armorDR <= 0 {
		return i18n.Text("Any armor worn counts as flexible against this damage: 1 HP of injury per 5 full points it stops, even if it stops all of it (BX431).")
	}
	trauma := calculator.BluntTraumaFromFall(fxp.FromInteger(c.mover.armorDR))
	if trauma == 0 {
		return i18n.Text("Armor DR %d counts as flexible against this damage, but it cannot stop 5 full points, so no blunt trauma gets through it (BX431).",
			c.mover.armorDR)
	}
	return i18n.Text("Armor DR %d counts as flexible against this damage: 1 HP of injury per 5 full points it stops, even if it stops all of it, so up to %d HP gets through it as blunt trauma (BX431).",
		c.mover.armorDR, trauma)
}

// updateTwoObjectResults handles a collision between two objects, either of which may be moving (BX432).
func (c *collisionCalculator) updateTwoObjectResults() []string {
	var notes []string
	strikerVelocity := c.mover.velocity
	if c.dropped {
		fall := c.fall()
		strikerVelocity, notes = fall.Velocity, fallNotes(fall)
		c.mover.velocity = strikerVelocity
		c.mover.velocityField.Sync()
	}
	angle := c.angles[c.angleIndex].Angle
	struckVelocity := c.target.velocity
	if angle == calculator.SideOnCollision {
		struckVelocity = 0
	}
	result := calculator.Collision(angle,
		calculator.CollisionObject{HP: c.mover.hp, Velocity: strikerVelocity, HalfDamage: c.mover.shape().HalfDamage},
		calculator.CollisionObject{HP: c.target.hp, Velocity: struckVelocity, HalfDamage: c.target.shape().HalfDamage})
	striker := c.mover.name(i18n.Text("the striking object"))
	struck := c.target.name(i18n.Text("the struck object"))
	c.addResult(i18n.Text("Collision velocity:"), velocityText(result.Velocity))
	c.addResult(i18n.Text("Damage to %s:", struck),
		damageText(c.mover.entity(), result.StrikerDice, c.mover.shape().DamageType))
	c.addResult(i18n.Text("Damage to %s:", striker),
		damageText(c.target.entity(), result.StruckDice, c.target.shape().DamageType))
	switch {
	case result.StrikerCapped:
		notes = append(notes, i18n.Text("As the slower object, %s cannot inflict more dice than %s (BX432).", striker, struck))
	case result.StruckCapped:
		notes = append(notes, i18n.Text("As the struck object, %s cannot inflict more dice than %s (BX432).", struck, striker))
	}
	if c.dropped {
		if calculator.DroppedObjectHampers(c.mover.sm, c.target.sm) {
			notes = append(notes, i18n.Text("The falling object is at least as big as the victim, so on the victim's next turn he may move only one yard and his active defenses are at -3 (BX431)."))
		}
	} else if st, overruns := calculator.OverrunST(c.mover.sm, c.target.sm, c.mover.st, c.mover.hp); overruns {
		thrust := calculator.ThrustFor(c.mover.entity(), st)
		c.addResult(i18n.Text("Overrun damage:"),
			gurps.FormatDice(thrust, gurps.SheetSettingsFor(c.mover.entity()).UseModifyingDicePlusAdds)+" cr")
		notes = append(notes, i18n.Text("Being at least two sizes bigger, %s overruns %s and inflicts thrust damage for ST %s as well (BX432).",
			striker, struck, st.Comma()))
	}
	return notes
}
