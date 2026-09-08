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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/dgroup"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/errs"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
	"github.com/richardwilkes/unison/enums/behavior"
)

var (
	_ unison.Dockable            = &CollisionCalculator{}
	_ unison.TabCloser           = &CollisionCalculator{}
	_ unison.UndoManagerProvider = &CollisionCalculator{}
)

// The scenarios the collision calculator handles, in the order the scenario popup offers them.
const (
	fallScenario = iota
	immovableScenario
	twoObjectScenario
	suddenStopScenario
)

var (
	collisionScenarios = []collisionScenario{
		{name: i18n.Text("Fall onto a surface")},
		{name: i18n.Text("Collision with an immovable object")},
		{name: i18n.Text("Collision between two objects")},
		{name: i18n.Text("Sudden stop of a vehicle, elevator, etc.")},
	}

	collisionShapes = []collisionShape{
		{name: i18n.Text("Blunt"), damageType: "cr"},
		{name: i18n.Text("Bullet-shaped"), damageType: "pi", halves: true},
		{name: i18n.Text("Sharp"), damageType: "cut", halves: true},
		{name: i18n.Text("Spiked"), damageType: "imp", halves: true},
	}

	collisionSurfaces = []collisionSurface{
		{name: i18n.Text("Hard (ground, concrete, a wall)"), hard: true},
		{name: i18n.Text("Soft (forest litter, hay, swamp)")},
		{name: i18n.Text("Soft and elastic (mattress, net, airbag)"), elastic: true},
		{name: i18n.Text("Water or another fluid"), water: true},
	}

	terminalVelocities = []terminalVelocityChoice{
		{name: i18n.Text("Human, spread-eagled (60)"), base: fxp.Sixty},
		{name: i18n.Text("Human, swan dive (100)"), base: fxp.Hundred},
		{name: i18n.Text("Dense or streamlined object (200)"), base: fxp.FromInteger(200)},
		{name: i18n.Text("None")},
		{name: i18n.Text("Custom"), custom: true},
	}

	collisionAngles = []collisionAngleChoice{
		{name: i18n.Text("Head-on"), angle: gurps.HeadOnCollision},
		{name: i18n.Text("Rear-end"), angle: gurps.RearEndCollision},
		{name: i18n.Text("Side-on, or the struck object is stationary"), angle: gurps.SideOnCollision},
	}

	collisionRestraints = []collisionRestraint{
		{name: i18n.Text("None")},
		{name: i18n.Text("Seatbelt or straps (DR 5)"), dr: 5},
		{name: i18n.Text("Airbag (DR 10)"), dr: 10},
	}
)

type collisionScenario struct {
	name string
}

func (s collisionScenario) String() string {
	return s.name
}

// collisionShape says what an object's shape does to the damage it inflicts: a bullet-shaped, sharp or spiked object
// does half damage, but of its own type rather than crushing (BX430).
type collisionShape struct {
	name       string
	damageType string
	halves     bool
}

func (s collisionShape) String() string {
	return s.name
}

// collisionSurface is a kind of immovable object (BX431): a hard one is hit as if the mover had twice its HP, an elastic
// one gives extra DR, and water can be dived into cleanly.
type collisionSurface struct {
	name    string
	hard    bool
	elastic bool
	water   bool
}

func (s collisionSurface) String() string {
	return s.name
}

// terminalVelocityChoice is a terminal velocity at one G in one atmosphere (BX431); base is zero for the choice that
// applies no limit, and custom marks the one whose value is typed in.
type terminalVelocityChoice struct {
	name   string
	base   fxp.Int
	custom bool
}

func (t terminalVelocityChoice) String() string {
	return t.name
}

type collisionAngleChoice struct {
	name  string
	angle gurps.CollisionAngle
}

func (a collisionAngleChoice) String() string {
	return a.name
}

// collisionRestraint is what holds an occupant in place during a sudden stop, and the DR it gives against the damage
// (BX432).
type collisionRestraint struct {
	name string
	dr   int
}

func (r collisionRestraint) String() string {
	return r.name
}

// CollisionCalculator works out the damage from a collision or a fall (BX430-BX431). Unlike the per-sheet Calculator it
// belongs to no document: each of the objects involved either has its numbers typed in or takes them from any open
// character sheet.
type CollisionCalculator struct {
	unison.Panel
	calculatorContent
	undoMgr             *unison.UndoManager
	scroll              *unison.ScrollPanel
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
	terminalPopup       *unison.PopupMenu[terminalVelocityChoice]
	customTerminalField *DecimalField
	pressureField       *DecimalField
	controlledFallBox   *unison.CheckBox
	elasticDRField      *IntegerField
	breakableBox        *unison.CheckBox
	obstacleHPField     *DecimalField
	obstacleDRField     *IntegerField
	cleanDiveBox        *unison.CheckBox
	mover               collisionParticipant
	target              collisionParticipant
	fallDistance        fxp.Int
	gravity             fxp.Int
	pressure            fxp.Int
	customTerminal      fxp.Int
	obstacleHP          fxp.Int
	moverName           string
	scale               int
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
	calc            *CollisionCalculator
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

// DisplayCollisionCalculator brings the collision calculator forward, opening it if it is not already open. preselect,
// when not nil, is the sheet the moving object starts out taking its numbers from; it is ignored when the calculator
// is already open, so that re-choosing the menu item never disturbs what has been entered.
func DisplayCollisionCalculator(preselect *Sheet) {
	if activateDockable[*CollisionCalculator](nil) {
		return
	}
	c := &CollisionCalculator{
		scale:          gurps.GlobalSettings().General.InitialEditorUIScale,
		fallDistance:   fxp.Five,
		gravity:        fxp.One,
		pressure:       fxp.One,
		customTerminal: fxp.FromInteger(200),
		elasticDR:      5,
	}
	c.Self = c
	c.mover = collisionParticipant{calc: c, hp: fxp.Ten, velocity: fxp.Five}
	c.target = collisionParticipant{calc: c, hp: fxp.Ten}

	c.undoMgr = unison.NewUndoManager(100, func(err error) { errs.Log(err) })
	c.SetLayout(&unison.FlexLayout{Columns: 1})

	c.createContent()

	c.scroll = unison.NewScrollPanel()
	c.scroll.SetContent(c.content, behavior.HintedFill, behavior.Fill)
	c.scroll.SetLayoutData(&unison.FlexLayoutData{
		HAlign: align.Fill,
		VAlign: align.Fill,
		HGrab:  true,
		VGrab:  true,
	})

	c.AddChild(c.createToolbar())
	c.AddChild(c.scroll)
	if preselect != nil {
		c.mover.sheet = preselect
		c.mover.selectSheet(preselect)
		c.mover.rebuild()
	}
	c.changed()
	c.content.ValidateScrollRoot()
	PlaceInDock(c, dgroup.Editors, false)
	c.content.RequestFocus()
}

// sheetChanged implements sheetSourceUser.
func (c *CollisionCalculator) sheetChanged(sheet *Sheet) {
	if c.mover.sheet == sheet || c.target.sheet == sheet {
		c.changed()
	}
}

func (c *CollisionCalculator) createToolbar() *unison.Panel {
	toolbar := newToolbar()
	toolbar.AddChild(NewDefaultInfoPop())
	addUIScaleField(toolbar, func() int { return gurps.GlobalSettings().General.InitialEditorUIScale },
		func() int { return c.scale }, func(scale int) { c.scale = scale }, false, c.scroll)
	finishToolbarLayout(toolbar)
	return toolbar
}

func (c *CollisionCalculator) createContent() {
	c.initCalculatorContent()
	c.content.AddChild(c.createHeader(i18n.Text("Collisions and Falls"),
		[]linkSpec{{pageRef: "BX430", highlight: "Collisions and Falls"}}, 0))

	row := c.addRow(2)
	addPlainLabel(row, i18n.Text("Scenario:"))
	addIndexPopup(row, collisionScenarios, &c.scenarioIndex, c.changed)

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

	divider := unison.NewSeparator()
	divider.SetBorder(unison.NewEmptyBorder(geom.NewVerticalInsets(unison.StdVSpacing * 2)))
	divider.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Fill, HGrab: true})
	c.content.AddChild(divider)
	c.addSubheader(i18n.Text("Results")).SetBorder(nil)
	c.results = c.addRow(2)
	c.results.SetLayout(&unison.FlexLayout{Columns: 2, HSpacing: unison.StdHSpacing * 2, VSpacing: unison.StdVSpacing})
	c.notes = newRowGroup()
	c.notes.SetBorder(unison.NewEmptyBorder(geom.Insets{Top: unison.StdVSpacing * 2, Left: unison.StdHSpacing * 2}))
	c.content.AddChild(c.notes)
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
	addIndexPopup(row, collisionShapes, &p.shapeIndex, p.calc.changed)
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

// showExtras adds or removes the rows that only matter for a fall.
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
	entity := sheet.Entity()
	p.velocity = fxp.FromInteger(entity.Move(entity.EncumbranceLevel(false)))
	p.velocityField.Sync()
	p.pullFromSheet()
}

// pullFromSheet reads the participant's numbers from its sheet. Each backing value is assigned before its field is
// synced, so that the setter the sync may run sees nothing new and does not start another round of updates.
func (p *collisionParticipant) pullFromSheet() {
	entity := p.sheet.Entity()
	if entity.ResolveAttribute(gurps.HitPointsID) != nil {
		p.hp = entity.Attributes.Maximum(gurps.HitPointsID).Max(0)
	}
	if entity.ResolveAttribute(gurps.StrengthID) != nil {
		p.st = entity.Attributes.Maximum(gurps.StrengthID).Max(0)
	}
	p.sm = entity.Profile.AdjustedSizeModifier()
	p.acrobatics = skillLevelOrDefault(entity, "Acrobatics", gurps.DexterityID, -6)
	p.swimming = skillLevelOrDefault(entity, "Swimming", gurps.HealthID, -4)
	total, armor := torsoDR(entity)
	p.armorDR = armor
	p.innateDR = total - armor
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

func (p *collisionParticipant) shape() collisionShape {
	return collisionShapes[p.shapeIndex]
}

func (c *CollisionCalculator) createFallRows() *unison.Panel {
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
	c.terminalPopup = addIndexPopup(row, terminalVelocities, &c.terminalIndex, c.changed)
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

func (c *CollisionCalculator) createSurfaceRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Surface:"))
	addIndexPopup(row, collisionSurfaces, &c.surfaceIndex, c.changed)
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
	row.AddChild(c.obstacleDRField)
	addPlainLabel(row, i18n.Text("DR of the breakable surface"))
	return group
}

func (c *CollisionCalculator) createDropRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	rows.addCheckBox(i18n.Text("The striking object was dropped, and its velocity comes from the fall"), &c.dropped,
		c.changed)
	return group
}

func (c *CollisionCalculator) createAngleRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Collision angle:"))
	addIndexPopup(row, collisionAngles, &c.angleIndex, c.changed)
	return group
}

func (c *CollisionCalculator) createRestraintRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Restraint:"))
	addIndexPopup(row, collisionRestraints, &c.restraintIndex, c.changed)
	return group
}

// changed is what every control runs once it has stored its value: the sources are re-read, the dependent controls are
// brought into line, and the results are recomputed.
func (c *CollisionCalculator) changed() {
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
func (c *CollisionCalculator) adjustControls() {
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
	adjustFieldBlank(c.target.velocityField, collisionAngles[c.angleIndex].angle == gurps.SideOnCollision)

	// Unison does not disable a panel's children along with it, so the fall rows are switched one by one when the
	// striking object in a two-object collision is not something that was dropped.
	adjustFieldBlank(c.fallDistanceField, !fromFall)
	adjustFieldBlank(c.gravityField, !fromFall)
	adjustPopupBlank(c.terminalPopup, !fromFall)
	terminal := terminalVelocities[c.terminalIndex]
	adjustFieldBlank(c.customTerminalField, !fromFall || !terminal.custom)
	adjustFieldBlank(c.pressureField, !fromFall || (terminal.base <= 0 && !terminal.custom))
	// The fall can be softened by an Acrobatics roll or by a clean dive into water, but not by both (BX431).
	c.controlledFallBox.SetTitle(i18n.Text("Made a successful Acrobatics roll for a controlled fall (-5 yards)"))
	c.controlledFallBox.SetEnabled(c.scenarioIndex == fallScenario && !c.diving())
	c.cleanDiveBox.SetTitle(fmt.Sprintf(i18n.Text("Made a successful Swimming roll at %d for a clean dive"),
		gurps.SpeedRangePenalty(c.moverVelocity())))
	c.cleanDiveBox.SetEnabled(c.inWater() && !c.controlledFallApplies())

	surface := collisionSurfaces[c.surfaceIndex]
	adjustFieldBlank(c.elasticDRField, !surface.elastic)
	adjustFieldBlank(c.obstacleHPField, !c.breakable)
	adjustFieldBlank(c.obstacleDRField, !c.breakable)

	c.content.MarkForLayoutRecursively()
	c.content.MarkForLayoutRecursivelyUpward()
	c.content.MarkForRedraw()
}

// velocityFromFall reports whether the moving object's velocity is the one it reaches in a fall rather than one that
// is typed in.
func (c *CollisionCalculator) velocityFromFall() bool {
	return c.scenarioIndex == fallScenario || (c.scenarioIndex == twoObjectScenario && c.dropped)
}

// inWater reports whether the moving object lands in water, where a clean dive is possible.
func (c *CollisionCalculator) inWater() bool {
	return (c.scenarioIndex == fallScenario || c.scenarioIndex == immovableScenario) &&
		collisionSurfaces[c.surfaceIndex].water
}

// controlledFallApplies reports whether the Acrobatics roll shortens the fall, which only a fall allows.
func (c *CollisionCalculator) controlledFallApplies() bool {
	return c.scenarioIndex == fallScenario && c.controlledFall
}

// diving reports whether a clean dive negates the damage: only in water, and not when the faller chose a controlled
// fall instead, since the rules allow one or the other (BX431).
func (c *CollisionCalculator) diving() bool {
	return c.inWater() && c.cleanDive && !c.controlledFallApplies()
}

// moverVelocity returns the velocity the moving object hits at: its own, or the one it reaches in a fall.
func (c *CollisionCalculator) moverVelocity() fxp.Int {
	if c.scenarioIndex == fallScenario || (c.scenarioIndex == twoObjectScenario && c.dropped) {
		v, _ := c.fallVelocity()
		return v
	}
	return c.mover.velocity
}

// fallVelocity returns the velocity reached in the fall and the notes explaining anything that limited it.
func (c *CollisionCalculator) fallVelocity() (velocity fxp.Int, notes []string) {
	distance := c.fallDistance
	if c.controlledFallApplies() {
		distance = (distance - fxp.Five).Max(0)
		notes = append(notes, fmt.Sprintf(i18n.Text("The controlled fall counts as a fall of %s yards (BX431)."), distance.Comma()))
	}
	velocity = gurps.FallingVelocity(distance, c.gravity)
	if c.gravity <= 0 {
		notes = append(notes, i18n.Text("Without gravity, there is no fall."))
		return velocity, notes
	}
	choice := terminalVelocities[c.terminalIndex]
	base := choice.base
	if choice.custom {
		base = c.customTerminal
	}
	if base <= 0 {
		return velocity, notes
	}
	limit, unlimited := gurps.TerminalVelocity(base, c.gravity, c.pressure)
	switch {
	case unlimited:
		notes = append(notes, i18n.Text("In a vacuum there is no terminal velocity (BX431)."))
	case velocity > limit:
		velocity = limit
		notes = append(notes, fmt.Sprintf(i18n.Text("The fall is limited to the terminal velocity of %s yards/second (BX431)."),
			limit.Comma()))
	}
	return velocity, notes
}

// updateResults recomputes the damage and rewrites the results and notes.
func (c *CollisionCalculator) updateResults() {
	c.results.RemoveAllChildren()
	var notes []string
	if c.scenarioIndex == twoObjectScenario {
		notes = c.updateTwoObjectResults()
	} else {
		notes = c.updateSurfaceResults()
	}
	c.notes.RemoveAllChildren()
	for _, note := range notes {
		if note != "" {
			c.notes.AddChild(newNoteRow(note))
		}
	}
	c.results.MarkForLayoutRecursivelyUpward()
	c.results.MarkForRedraw()
}

// addResult adds a labeled result to the results panel.
func (c *CollisionCalculator) addResult(label, value string) {
	addResult(c.results, label, value)
}

// velocityText describes a velocity in yards per second and miles per hour (2 mph is 1 yard/second, BX430).
func velocityText(velocity fxp.Int) string {
	return fmt.Sprintf(i18n.Text("%s yards/second (%s mph)"), velocity.Comma(), velocity.Mul(fxp.Two).Comma())
}

// damageText formats dice of the given type, using the sheet's dice notation when the numbers came from a sheet.
func damageText(entity *gurps.Entity, count fxp.Int, damageType string) string {
	if count <= 0 {
		return i18n.Text("None")
	}
	return gurps.FormatDice(gurps.CollisionDamageDice(count), gurps.SheetSettingsFor(entity).UseModifyingDicePlusAdds) +
		" " + damageType
}

// updateSurfaceResults handles the scenarios where the moving object hits something immovable: a fall, a collision with
// an obstacle, and an occupant's sudden stop, which is a fall at the velocity lost.
func (c *CollisionCalculator) updateSurfaceResults() []string {
	var notes []string
	velocity := c.mover.velocity
	if c.scenarioIndex == fallScenario {
		velocity, notes = c.fallVelocity()
		c.mover.velocity = velocity
		c.mover.velocityField.Sync()
	}
	surface := collisionSurfaces[c.surfaceIndex]
	if c.scenarioIndex == suddenStopScenario {
		surface = collisionSurfaces[0]
	}
	shape := c.mover.shape()
	count := gurps.CollisionDiceCount(gurps.ImmovableCollisionHP(c.mover.hp, surface.hard), velocity)
	if shape.halves {
		count = count.Div(fxp.Two)
	}
	entity := c.mover.entity()
	mover := c.mover.name(c.moverName)
	c.addResult(i18n.Text("Velocity:"), velocityText(velocity))
	if c.diving() {
		c.addResult(fmt.Sprintf(i18n.Text("Damage to %s:"), mover), i18n.Text("None"))
		notes = append(notes, i18n.Text("The clean dive negates all damage (BX431)."))
		return notes
	}
	damage := damageText(entity, count, shape.damageType)
	c.addResult(fmt.Sprintf(i18n.Text("Damage to %s:"), mover), damage)
	if c.scenarioIndex != suddenStopScenario {
		c.addResult(i18n.Text("Damage to the surface:"), damage)
	}
	if surface.hard {
		notes = append(notes, i18n.Text("The surface is hard, so the damage is worked out with twice the HP (BX431)."))
	}
	if c.scenarioIndex != suddenStopScenario && c.breakable {
		notes = append(notes, fmt.Sprintf(i18n.Text("The surface can break, so neither side takes more than %s points, its HP + DR (BX431)."),
			(c.obstacleHP+fxp.FromInteger(c.obstacleDR)).Comma()))
	}
	if surface.elastic {
		notes = append(notes, fmt.Sprintf(i18n.Text("The elastic surface gives DR %d against this damage (BX431)."), c.elasticDR))
	}
	if c.inWater() {
		notes = append(notes, c.swimmingNote())
	}
	if c.scenarioIndex == suddenStopScenario {
		if dr := collisionRestraints[c.restraintIndex].dr; dr > 0 {
			notes = append(notes, fmt.Sprintf(i18n.Text("The restraint gives DR %d against this damage (BX432)."), dr))
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
func (c *CollisionCalculator) swimmingNote() string {
	penalty := gurps.SpeedRangePenalty(c.moverVelocity())
	roll := i18n.Text("Swimming roll")
	if c.scenarioIndex == immovableScenario {
		roll = i18n.Text("Swimming roll (or vehicle control roll, when ditching a vehicle)")
	}
	if c.mover.swimming > 0 {
		return fmt.Sprintf(i18n.Text("A successful %s at %d (effective skill %d) would be a clean dive that negates all damage (BX431)."),
			roll, penalty, c.mover.swimming+penalty)
	}
	return fmt.Sprintf(i18n.Text("A successful %s at %d would be a clean dive that negates all damage (BX431)."), roll, penalty)
}

// armorNote describes how the mover's armor fares against falling damage: all of it counts as flexible, so it lets 1 HP
// of injury through for every 5 full points it stops, even when it stops all of it (BX431). The most it can stop is its
// own DR, which bounds the blunt trauma.
func (c *CollisionCalculator) armorNote(count fxp.Int) string {
	if count <= 0 {
		return ""
	}
	if c.mover.armorDR <= 0 {
		return i18n.Text("Any armor worn counts as flexible against this damage: 1 HP of injury per 5 full points it stops, even if it stops all of it (BX431).")
	}
	trauma := gurps.BluntTraumaFromFall(fxp.FromInteger(c.mover.armorDR))
	if trauma == 0 {
		return fmt.Sprintf(i18n.Text("Armor DR %d counts as flexible against this damage, but it cannot stop 5 full points, so no blunt trauma gets through it (BX431)."),
			c.mover.armorDR)
	}
	return fmt.Sprintf(i18n.Text("Armor DR %d counts as flexible against this damage: 1 HP of injury per 5 full points it stops, even if it stops all of it, so up to %d HP gets through it as blunt trauma (BX431)."),
		c.mover.armorDR, trauma)
}

// updateTwoObjectResults handles a collision between two objects, either of which may be moving (BX432).
func (c *CollisionCalculator) updateTwoObjectResults() []string {
	var notes []string
	strikerVelocity := c.mover.velocity
	if c.dropped {
		strikerVelocity, notes = c.fallVelocity()
		c.mover.velocity = strikerVelocity
		c.mover.velocityField.Sync()
	}
	angle := collisionAngles[c.angleIndex].angle
	struckVelocity := c.target.velocity
	if angle == gurps.SideOnCollision {
		struckVelocity = 0
	}
	result := gurps.Collision(angle,
		gurps.CollisionObject{HP: c.mover.hp, Velocity: strikerVelocity, HalfDamage: c.mover.shape().halves},
		gurps.CollisionObject{HP: c.target.hp, Velocity: struckVelocity, HalfDamage: c.target.shape().halves})
	striker := c.mover.name(i18n.Text("the striking object"))
	struck := c.target.name(i18n.Text("the struck object"))
	c.addResult(i18n.Text("Collision velocity:"), velocityText(result.Velocity))
	c.addResult(fmt.Sprintf(i18n.Text("Damage to %s:"), struck),
		damageText(c.mover.entity(), result.StrikerDice, c.mover.shape().damageType))
	c.addResult(fmt.Sprintf(i18n.Text("Damage to %s:"), striker),
		damageText(c.target.entity(), result.StruckDice, c.target.shape().damageType))
	switch {
	case result.StrikerCapped:
		notes = append(notes, fmt.Sprintf(i18n.Text("As the slower object, %s cannot inflict more dice than %s (BX432)."), striker, struck))
	case result.StruckCapped:
		notes = append(notes, fmt.Sprintf(i18n.Text("As the struck object, %s cannot inflict more dice than %s (BX432)."), struck, striker))
	}
	if c.dropped {
		if c.mover.sm >= c.target.sm {
			notes = append(notes, i18n.Text("The falling object is at least as big as the victim, so on the victim's next turn he may move only one yard and his active defenses are at -3 (BX431)."))
		}
	} else if c.mover.sm >= c.target.sm+2 {
		st := c.mover.st
		if st <= 0 {
			st = c.mover.hp
		}
		st = st.Div(fxp.Two).Floor()
		thrust := gurps.SheetSettingsFor(c.mover.entity()).DamageProgression.Thrust(st.AsInteger[int]())
		c.addResult(i18n.Text("Overrun damage:"),
			gurps.FormatDice(thrust, gurps.SheetSettingsFor(c.mover.entity()).UseModifyingDicePlusAdds)+" cr")
		notes = append(notes, fmt.Sprintf(i18n.Text("Being at least two sizes bigger, %s overruns %s and inflicts thrust damage for ST %s as well (BX432)."),
			striker, struck, st.Comma()))
	}
	return notes
}

// TitleIcon implements unison.Dockable
func (c *CollisionCalculator) TitleIcon(suggestedSize geom.Size) unison.Drawable {
	return &unison.DrawableSVG{
		SVG:  svg.Calculator,
		Size: suggestedSize,
	}
}

// Title implements unison.Dockable
func (c *CollisionCalculator) Title() string {
	return i18n.Text("Collision & Falling Damage Calculator")
}

func (c *CollisionCalculator) String() string {
	return c.Title()
}

// Tooltip implements unison.Dockable
func (c *CollisionCalculator) Tooltip() string {
	return ""
}

// Modified implements unison.Dockable
func (c *CollisionCalculator) Modified() bool {
	return false
}

// MayAttemptClose implements unison.TabCloser
func (c *CollisionCalculator) MayAttemptClose() bool {
	return true
}

// AttemptClose implements unison.TabCloser
func (c *CollisionCalculator) AttemptClose() bool {
	return AttemptCloseForDockable(c)
}

// UndoManager implements unison.UndoManagerProvider
func (c *CollisionCalculator) UndoManager() *unison.UndoManager {
	return c.undoMgr
}
