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
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/calculator"
	"github.com/richardwilkes/rpgtools/dice"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/align"
)

var _ calculatorTab = &explosionCalculator{}

// The kinds of attack the calculator handles, in the order the attack type popup offers them.
const (
	explosionAttack = iota
	areaEffectAttack
	coneAttack
)

// newExplosionAttackTypes returns the kinds of attack the calculator handles, in the order the attack constants give
// them.
func newExplosionAttackTypes() []explosionAttackType {
	return []explosionAttackType{
		{name: i18n.Text("Explosion")},
		{name: i18n.Text("Area-effect attack")},
		{name: i18n.Text("Cone attack")},
	}
}

type explosionAttackType struct {
	name string
}

func (a explosionAttackType) String() string {
	return a.name
}

// weaponSource is an entry in the attack's Source popup: an explosive weapon on an open character sheet, or none for an
// attack whose dice are typed in.
type weaponSource struct {
	name   string
	sheet  *Sheet
	weapon *gurps.Weapon
}

func (s weaponSource) String() string {
	return s.name
}

// exposedChoice is an entry in the target's Exposed locations popup: every location, as a true area effect leaves them,
// or the least-protected one facing the attack, which is what Large-Area Injury averages the torso DR with (BX400).
type exposedChoice struct {
	name     string
	location *gurps.HitLocation
}

func (e exposedChoice) String() string {
	return e.name
}

// explosionCalculator works out what an explosion or an area attack does to something standing at a given distance
// (BX413-BX415). The attack is either typed in or taken from an explosive weapon on any open character sheet, and so
// is its target.
type explosionCalculator struct {
	calculatorContent
	attackSlot         *unison.Panel
	explosionRows      *unison.Panel
	coneRows           *unison.Panel
	areaRows           *unison.Panel
	results            *unison.Panel
	notes              *unison.Panel
	blastLabel         *textLabel
	blastField         *StringField
	damageTypeField    *StringField
	fragmentationField *StringField
	coneRangeField     *DecimalField
	coneWidthField     *DecimalField
	weaponPopup        *unison.PopupMenu[weaponSource]
	airburstBox        *unison.CheckBox
	hotFragmentsBox    *unison.CheckBox
	dissipatesBox      *unison.CheckBox
	htResistedBox      *unison.CheckBox
	attackTypes        []explosionAttackType
	environments       []calculator.ExplosionEnvironmentChoice
	postures           []calculator.TargetPostureChoice
	situations         []calculator.BlastSituationChoice
	target             blastTarget
	weaponSheet        *Sheet
	weapon             *gurps.Weapon
	blastDice          dice.Dice
	fragmentationDice  dice.Dice
	coneMaxRange       fxp.Int
	coneMaxWidth       fxp.Int
	blastSpec          string
	damageType         string
	fragmentationSpec  string
	attackTypeIndex    int
	environmentIndex   int
	airburst           bool
	hotFragments       bool
	dissipates         bool
	htResisted         bool
	updating           bool
	rebuilding         bool
}

// blastTarget is whatever the explosion or area attack is worked out against. Its numbers are either typed in or taken
// from an open character sheet, in which case the fields holding them are locked and refreshed whenever the sheet
// changes. Its distance, posture and situation are always typed in, since where it was standing is a matter of
// circumstance rather than something a sheet knows.
type blastTarget struct {
	sheetSourcePicker
	calc           *explosionCalculator
	panel          *unison.Panel
	situationSlot  *unison.Panel
	situationRow   *unison.Panel
	drLabel        *textLabel
	distanceLabel  *textLabel
	smField        *IntegerField
	hpField        *DecimalField
	torsoDRField   *IntegerField
	drField        *IntegerField
	distanceField  *DecimalField
	exposedPopup   *unison.PopupMenu[exposedChoice]
	hp             fxp.Int
	distance       fxp.Int
	exposed        *gurps.HitLocation
	sm             int
	torsoDR        int
	dr             int
	postureIndex   int
	situationIndex int
}

func newExplosionCalculator() *explosionCalculator {
	c := &explosionCalculator{
		attackTypes:  newExplosionAttackTypes(),
		environments: calculator.ExplosionEnvironmentChoices(),
		postures:     calculator.TargetPostureChoices(),
		situations:   calculator.BlastSituationChoices(),
		coneMaxRange: fxp.Hundred,
		coneMaxWidth: fxp.Five,
		blastSpec:    "6d",
		damageType:   "cr",
	}
	c.blastDice = gurps.Roller.Parse(c.blastSpec)
	c.target = blastTarget{calc: c, hp: fxp.Ten, distance: fxp.Five}
	c.createContent()
	return c
}

// title implements calculatorTab.
func (c *explosionCalculator) title() string {
	return i18n.Text("Explosions & Area Attacks")
}

// panel implements calculatorTab.
func (c *explosionCalculator) panel() *unison.Panel {
	return c.content
}

// preselect implements calculatorTab. The sheet becomes the target's source.
func (c *explosionCalculator) preselect(sheet *Sheet) {
	c.target.preselect(sheet)
}

// sheetChanged implements calculatorTab.
func (c *explosionCalculator) sheetChanged(sheet *Sheet) {
	if c.target.sheet == sheet || c.weaponSheet == sheet {
		c.changed()
	}
}

func (c *explosionCalculator) createContent() {
	c.initCalculatorContent()
	c.content.AddChild(c.createHeader(i18n.Text("Explosions & Area Attacks"),
		[]linkSpec{
			{pageRef: "BX413", highlight: "Area and Spreading Attacks"},
			{pageRef: "BX414", highlight: "Explosions"},
		}, 0))

	row := c.addRow(2)
	addPlainLabel(row, i18n.Text("Attack type:"))
	addIndexPopup(row, c.attackTypes, &c.attackTypeIndex, c.changed)

	c.addSubheader(i18n.Text("Attack"))
	c.createAttackRows()

	c.addSubheader(i18n.Text("Target"))
	c.target.createPanel(c.content)

	c.results, c.notes = c.addResultsSection()
}

// createAttackRows builds the rows describing the attack itself: where its numbers come from, the dice it rolls, and
// the rows that only one kind of attack has, which live in a slot the attack type swaps.
func (c *explosionCalculator) createAttackRows() {
	group := newRowGroup()
	c.content.AddChild(group)
	rows := &calculatorContent{content: group}

	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Source:"))
	c.weaponPopup = unison.NewPopupMenu[weaponSource]()
	c.weaponPopup.WillShowMenuCallback = func(_ *unison.PopupMenu[weaponSource]) { c.rebuildWeaponSources() }
	c.weaponPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[weaponSource]) {
		if c.rebuilding {
			return
		}
		if source, ok := popup.Selected(); ok {
			c.weapon = source.weapon
			c.weaponSheet = source.sheet
			c.pullFromWeapon()
		}
		c.changed()
	}
	row.AddChild(c.weaponPopup)
	c.rebuildWeaponSources()

	c.blastField = c.newSpecField(i18n.Text("Blast Damage"), &c.blastSpec, &c.blastDice)
	c.blastLabel = rows.addFieldRow(c.blastField, "")
	c.damageTypeField = NewStringField(nil, "", i18n.Text("Damage Type"),
		func() string { return c.damageType },
		func(v string) {
			c.damageType = v
			c.changed()
		})
	sizeStringField(c.damageTypeField)
	rows.addFieldRow(c.damageTypeField, i18n.Text("damage type, e.g. cr"))
	c.fragmentationField = c.newSpecField(i18n.Text("Fragmentation"), &c.fragmentationSpec, &c.fragmentationDice)
	rows.addFieldRow(c.fragmentationField, i18n.Text("fragmentation dice, e.g. 2d (blank for none)"))

	c.attackSlot = newRowGroup()
	group.AddChild(c.attackSlot)
	c.explosionRows = c.createExplosionRows()
	c.coneRows = c.createConeRows()
	c.areaRows = c.createAreaRows()
}

// newSpecField returns a field holding a dice specification, which is parsed as it is typed so that the results always
// follow the dice the spec actually names.
func (c *explosionCalculator) newSpecField(undoTitle string, spec *string, parsed *dice.Dice) *StringField {
	field := NewStringField(nil, "", undoTitle,
		func() string { return *spec },
		func(v string) {
			*spec = v
			*parsed = gurps.Roller.Parse(v)
			c.changed()
		})
	sizeStringField(field)
	return field
}

// sizeStringField sizes a calculator's text field to the same width as its numeric fields, so that the fields line up
// down the column, and keeps it from stretching across the rest of the row.
func sizeStringField(field *StringField) {
	field.SetMinimumTextWidthUsing(calculatorFieldPrototype)
	field.SetLayoutData(&unison.FlexLayoutData{HAlign: align.Start})
}

func (c *explosionCalculator) createExplosionRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Environment:"))
	addIndexPopup(row, c.environments, &c.environmentIndex, c.changed)
	c.airburstBox = rows.addCheckBox(i18n.Text("Airburst (posture does not protect against the fragments)"),
		&c.airburst, c.changed)
	c.hotFragmentsBox = rows.addCheckBox(i18n.Text("Hot fragments (white phosphorus)"), &c.hotFragments, c.changed)
	return group
}

func (c *explosionCalculator) createConeRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	c.coneRangeField = sameWidth(NewDecimalField(nil, "", i18n.Text("Maximum Range"),
		func() fxp.Int { return c.coneMaxRange },
		func(v fxp.Int) {
			c.coneMaxRange = v
			c.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(c.coneRangeField, i18n.Text("yards of maximum range"))
	c.coneWidthField = sameWidth(NewDecimalField(nil, "", i18n.Text("Maximum Width"),
		func() fxp.Int { return c.coneMaxWidth },
		func(v fxp.Int) {
			c.coneMaxWidth = v
			c.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(c.coneWidthField, i18n.Text("yards of maximum width, 0 if unspecified"))
	return group
}

func (c *explosionCalculator) createAreaRows() *unison.Panel {
	group := newRowGroup()
	rows := &calculatorContent{content: group}
	c.dissipatesBox = rows.addCheckBox(i18n.Text("The attack dissipates with distance (Dissipation)"), &c.dissipates,
		c.changed)
	c.htResistedBox = rows.addCheckBox(i18n.Text("Resisted by a HT roll (the divisor is a bonus to HT instead)"),
		&c.htResisted, c.changed)
	return group
}

// createPanel builds the target's rows into a panel of its own, added to the parent.
func (t *blastTarget) createPanel(parent *unison.Panel) {
	t.panel = newRowGroup()
	parent.AddChild(t.panel)
	rows := &calculatorContent{content: t.panel}

	t.selected = t.selectSheet
	t.refreshed = t.pullFromSheet
	t.changed = t.calc.changed
	t.addRow(rows, i18n.Text("Source:"))

	t.smField = sameWidth(NewIntegerField(nil, "", i18n.Text("Size Modifier"),
		func() int { return t.sm },
		func(v int) {
			t.sm = v
			t.calc.changed()
		},
		-100, 100, true, false))
	rows.addFieldRow(t.smField, i18n.Text("SM"))
	t.hpField = sameWidth(NewDecimalField(nil, "", i18n.Text("Hit Points"),
		func() fxp.Int { return t.hp },
		func(v fxp.Int) {
			t.hp = v
			t.calc.changed()
		},
		0, fxp.Max, false, false))
	rows.addFieldRow(t.hpField, i18n.Text("HP"))
	t.torsoDRField = sameWidth(NewIntegerField(nil, "", i18n.Text("Torso DR"),
		func() int { return t.torsoDR },
		func(v int) {
			t.torsoDR = v
			t.calc.changed()
		},
		0, 10000, false, false))
	rows.addFieldRow(t.torsoDRField, i18n.Text("torso DR"))

	row := rows.addRow(2)
	addPlainLabel(row, i18n.Text("Exposed locations:"))
	t.exposedPopup = unison.NewPopupMenu[exposedChoice]()
	t.exposedPopup.SelectionChangedCallback = func(popup *unison.PopupMenu[exposedChoice]) {
		if t.calc.rebuilding {
			return
		}
		if choice, ok := popup.Selected(); ok {
			t.exposed = choice.location
		}
		t.calc.changed()
	}
	row.AddChild(t.exposedPopup)
	t.rebuildExposedChoices(nil)

	t.drField = sameWidth(NewIntegerField(nil, "", i18n.Text("DR"),
		func() int { return t.dr },
		func(v int) {
			t.dr = v
			t.calc.changed()
		},
		0, 10000, false, false))
	t.drLabel = rows.addFieldRow(t.drField, "")

	row = rows.addRow(2)
	addPlainLabel(row, i18n.Text("Posture:"))
	addIndexPopup(row, t.calc.postures, &t.postureIndex, t.calc.changed)

	t.situationSlot = newRowGroup()
	t.panel.AddChild(t.situationSlot)
	situationRows := &calculatorContent{content: t.situationSlot}
	t.situationRow = situationRows.addRow(2)
	addPlainLabel(t.situationRow, i18n.Text("Situation:"))
	addIndexPopup(t.situationRow, t.calc.situations, &t.situationIndex, t.calc.changed)

	t.distanceField = sameWidth(NewDecimalField(nil, "", i18n.Text("Distance"),
		func() fxp.Int { return t.distance },
		func(v fxp.Int) {
			t.distance = v
			t.calc.changed()
		},
		0, fxp.Max, false, false))
	t.distanceLabel = rows.addFieldRow(t.distanceField, "")
}

// selectSheet reads the target's numbers from the sheet the picker has just made its source, doing nothing when there
// is none. The exposed location belonged to the previous sheet, so it is dropped.
func (t *blastTarget) selectSheet(sheet *Sheet) {
	if sheet == nil {
		return
	}
	t.exposed = nil
	t.pullFromSheet()
}

// pullFromSheet reads the target's numbers from its sheet. Each backing value is assigned before its field is synced,
// so that the setter the sync may run sees nothing new and does not start another round of updates.
func (t *blastTarget) pullFromSheet() {
	entity := t.sheet.Entity()
	// The locations are rebuilt first, since the DR is averaged over whichever of them is exposed.
	t.rebuildExposedChoices(entity)
	values := calculator.BlastTargetFromEntity(entity, t.calc.damageType, t.exposedFilter())
	if values.HasHP {
		t.hp = values.HP
	}
	t.sm = values.SM
	t.torsoDR = values.TorsoDR
	t.dr = values.DR
	t.smField.Sync()
	t.hpField.Sync()
	t.torsoDRField.Sync()
	t.drField.Sync()
}

// rebuildExposedChoices fills the Exposed locations popup with the entity's hit locations, keeping the current choice
// selected if it is still among them. A nil entity leaves only the choice that exposes every location, which is all a
// target whose numbers are typed in can offer.
func (t *blastTarget) rebuildExposedChoices(entity *gurps.Entity) {
	t.calc.rebuilding = true
	defer func() { t.calc.rebuilding = false }()
	t.exposedPopup.RemoveAllItems()
	t.exposedPopup.AddItem(exposedChoice{name: i18n.Text("All locations")})
	selected := 0
	if entity != nil {
		if body := gurps.SheetSettingsFor(entity).BodyType; body != nil {
			for _, loc := range body.UniqueHitLocations(entity) {
				t.exposedPopup.AddItem(exposedChoice{
					name:     i18n.Text("Least protected: %s", loc.ChoiceName),
					location: loc,
				})
				if loc == t.exposed {
					selected = t.exposedPopup.ItemCount() - 1
				}
			}
		}
	}
	if selected == 0 {
		t.exposed = nil
	}
	t.exposedPopup.SelectIndex(selected)
}

// exposedFilter returns the test LargeAreaDR uses to decide which locations face the attack: nil when every location is
// exposed (a true area effect, or no location named), otherwise one that accepts only the named location.
func (t *blastTarget) exposedFilter() func(*gurps.HitLocation) bool {
	return calculator.ExposedFilter(t.exposed, t.calc.attackTypeIndex == areaEffectAttack)
}

// update re-reads what the target's sheet supplies, dropping the sheet if it has been closed.
func (t *blastTarget) update() {
	t.refresh()
	if t.sheet == nil && t.exposedPopup.ItemCount() > 1 {
		// The locations belonged to the sheet that has just gone away.
		t.rebuildExposedChoices(nil)
	}
}

// lockSheetFields enables the fields whose values are typed in and disables those that a sheet supplies.
func (t *blastTarget) lockSheetFields() {
	manual := t.sheet == nil
	t.smField.SetEnabled(manual)
	t.hpField.SetEnabled(manual)
	t.torsoDRField.SetEnabled(manual)
	t.drField.SetEnabled(manual)
}

// posture returns the posture the target is in when the fragments arrive.
func (t *blastTarget) posture() calculator.TargetPosture {
	return t.calc.postures[t.postureIndex].Posture
}

// situation returns where the target was when the explosion went off.
func (t *blastTarget) situation() calculator.BlastSituation {
	return t.calc.situations[t.situationIndex].Situation
}

// changed implements calculatorTab.
func (c *explosionCalculator) changed() {
	if c.updating {
		return
	}
	c.updating = true
	defer func() { c.updating = false }()
	c.refreshSources()
	c.adjustControls()
	c.updateResults()
}

// refreshSources re-reads the attack from the weapon it was taken from and the target from its sheet, dropping either
// source that is no longer there. The weapon comes first, since the DR the target's sheet works out depends on the
// damage type the weapon supplies.
func (c *explosionCalculator) refreshSources() {
	c.rebuildWeaponSources()
	c.pullFromWeapon()
	c.target.update()
}

// rebuildWeaponSources fills the attack's Source popup with the explosive weapons on the sheets that are open right
// now, keeping the current weapon selected if it is still among them and otherwise dropping back to a typed-in attack,
// which keeps the values last read from the weapon.
func (c *explosionCalculator) rebuildWeaponSources() {
	c.rebuilding = true
	defer func() { c.rebuilding = false }()
	c.weaponPopup.RemoveAllItems()
	c.weaponPopup.AddItem(weaponSource{name: i18n.Text("Typed in")})
	sheets := OpenSheets(nil)
	names := sheetSourceNames(sheets)
	selected := 0
	for i, sheet := range sheets {
		entity := sheet.Entity()
		for _, w := range append(entity.Weapons(false, true, true), entity.Weapons(true, true, true)...) {
			resolved := w.Damage.ResolveDamage(nil)
			if resolved == nil || !resolved.IsExplosive() {
				continue
			}
			c.weaponPopup.AddItem(weaponSource{
				name:   i18n.Text("%s: %s, %s (%s)", names[i], w.String(), w.UsageWithReplacements(), resolved.String()),
				sheet:  sheet,
				weapon: w,
			})
			if w == c.weapon {
				selected = c.weaponPopup.ItemCount() - 1
			}
		}
	}
	if selected == 0 {
		c.weapon = nil
		c.weaponSheet = nil
	}
	c.weaponPopup.SelectIndex(selected)
}

// pullFromWeapon fills the attack's fields from the weapon chosen as its source, and does nothing at all when the
// attack is typed in. Each backing value is assigned before its field is synced, so that the setter the sync may run
// sees nothing new and does not start another round of updates.
func (c *explosionCalculator) pullFromWeapon() {
	if c.weapon == nil {
		return
	}
	resolved := c.weapon.Damage.ResolveDamage(nil)
	if resolved == nil {
		return
	}
	useExtra := gurps.SheetSettingsFor(c.weaponSheet.Entity()).UseModifyingDicePlusAdds
	c.blastDice = resolved.Dice
	c.blastSpec = gurps.FormatDice(resolved.Dice, useExtra)
	c.damageType = resolved.Type
	c.fragmentationDice = dice.Dice{}
	c.fragmentationSpec = ""
	if resolved.HasFragmentation {
		c.fragmentationDice = resolved.Fragmentation
		c.fragmentationSpec = gurps.FormatDice(resolved.Fragmentation, useExtra)
	}
	c.blastField.Sync()
	c.damageTypeField.Sync()
	c.fragmentationField.Sync()
}

// adjustControls shows the rows the attack type needs, locks the fields a source supplies and enables, disables and
// retitles the rest to match the current choices.
func (c *explosionCalculator) adjustControls() {
	explosion := c.attackTypeIndex == explosionAttack
	area := c.attackTypeIndex == areaEffectAttack
	switch c.attackTypeIndex {
	case explosionAttack:
		fillSlot(c.attackSlot, c.explosionRows)
	case coneAttack:
		fillSlot(c.attackSlot, c.coneRows, c.areaRows)
	default:
		fillSlot(c.attackSlot, c.areaRows)
	}
	if explosion {
		fillSlot(c.target.situationSlot, c.target.situationRow)
	} else {
		fillSlot(c.target.situationSlot)
	}

	// An attack taken from a weapon has its damage read from the sheet, so those fields are locked rather than typed in.
	fromWeapon := c.weapon != nil
	c.blastField.SetEnabled(!fromWeapon)
	c.damageTypeField.SetEnabled(!fromWeapon)
	c.fragmentationField.SetEnabled(!fromWeapon)
	c.blastLabel.SetTitle(i18n.Text("dice of damage (%s)", c.blastText()))

	c.target.lockSheetFields()
	if c.target.sheet != nil {
		c.target.drLabel.SetTitle(i18n.Text("effective DR (Large-Area Injury, BX400)"))
	} else {
		c.target.drLabel.SetTitle(i18n.Text("DR against the attack"))
	}
	// Every location is exposed to a true area effect (BX400), so there is nothing to choose among for one.
	if area {
		c.target.exposed = nil
		c.target.exposedPopup.SelectIndex(0)
	}
	// A control whose input would not be used is blanked as well as disabled, so that a stale value cannot be read as
	// part of the answer; one a sheet or weapon supplies keeps showing its value, since that value is in use.
	adjustPopupBlank(c.target.exposedPopup, area || c.target.sheet == nil)
	switch c.attackTypeIndex {
	case explosionAttack:
		c.target.distanceLabel.SetTitle(i18n.Text("yards from the center of the blast"))
	case coneAttack:
		c.target.distanceLabel.SetTitle(i18n.Text("yards from the apex of the cone"))
	default:
		c.target.distanceLabel.SetTitle(i18n.Text("yards from the center of the area"))
	}
	adjustFieldBlank(c.target.distanceField, explosion && c.target.situation() != calculator.CaughtInBlast)

	c.content.MarkForLayoutRecursively()
	c.content.MarkForLayoutRecursivelyUpward()
	c.content.MarkForRedraw()
}

// updateResults recomputes the attack's results and rewrites the notes that go with it.
func (c *explosionCalculator) updateResults() {
	c.results.RemoveAllChildren()
	var notes []string
	if c.attackTypeIndex == explosionAttack {
		notes = c.updateExplosionResults()
	} else {
		notes = c.updateAreaResults()
	}
	setNotes(c.notes, notes)
	c.results.MarkForLayoutRecursivelyUpward()
	c.results.MarkForRedraw()
}

func (c *explosionCalculator) addResult(label, value string) {
	addResult(c.results, label, value)
}

// useExtraDice reports whether the dice are formatted with the target's preference for modifying dice plus adds.
func (c *explosionCalculator) useExtraDice() bool {
	return gurps.SheetSettingsFor(c.target.entity()).UseModifyingDicePlusAdds
}

// blastText returns the blast's dice as the roller normalizes them, which is how a specification that does not say what
// the typist meant becomes visible.
func (c *explosionCalculator) blastText() string {
	return gurps.FormatDice(c.blastDice, c.useExtraDice())
}

// attackDamageType returns the damage type to show for the attack. An explosion's type always carries the Explosion
// modifier (B104), so it is spelled out for one that does not say so already.
func (c *explosionCalculator) attackDamageType() string {
	if c.attackTypeIndex == explosionAttack {
		return calculator.ExplosionDamageType(c.damageType)
	}
	return strings.TrimSpace(c.damageType)
}

// damageText describes the dice and type an attack inflicts, with the divisor its distance imposes.
func (c *explosionCalculator) damageText(divisor fxp.Int) string {
	text := strings.TrimSpace(c.blastText() + " " + c.attackDamageType())
	if divisor <= fxp.One {
		return text
	}
	return i18n.Text("%s ÷ %s, rounded down", text, divisor.Comma())
}

// updateExplosionResults works out what an explosion does to the target and returns the notes that go with it (BX414).
func (c *explosionCalculator) updateExplosionResults() []string {
	radius := calculator.CollateralDamageRadius(c.blastDice)
	c.addResult(i18n.Text("Collateral damage radius:"),
		i18n.Text("%d yards (%d dice)", radius, calculator.DiceOfDamage(c.blastDice)))
	notes := []string{
		i18n.Text("An armor divisor on the explosive applies to neither its collateral damage nor its fragments."),
		i18n.Text("Large-Area Injury (BX400): treat the damage as a torso hit, with no hit location wounding modifier, unless only one location is exposed. Only the locations facing the blast are exposed; those behind cover or masked by the body are not."),
		i18n.Text("Explosions are incendiary, so check for Catching Fire (BX434)."),
		i18n.Text("The only defense is Dodge and Drop (BX377): a dodge at +3 that leaves the target prone. With cover a step away, success reaches it in time; even without cover, the step puts the target a yard farther from the blast."),
	}
	divisor, divisorText, inRange := c.blastDivisor(radius)
	c.addResult(i18n.Text("Damage divisor:"), divisorText)
	if !inRange {
		// The fragments still reach further than the blast does whenever their own radius is the larger one.
		return append(notes, c.fragmentationResults()...)
	}
	c.addResult(i18n.Text("Blast damage:"), c.damageText(divisor))
	minimum, average, maximum := calculator.DividedDamage(c.blastDice, divisor)
	dr := c.target.dr
	switch c.target.situation() {
	case calculator.ThrewSelfOnExplosive:
		c.addResult(i18n.Text("Damage taken:"), i18n.Text("maximum possible damage: %d", maximum))
		c.addResult(i18n.Text("DR against the blast:"), i18n.Text("%d (DR protects normally)", dr))
		c.addResult(i18n.Text("Penetrating (maximum):"), fmt.Sprintf("%d", max(maximum-dr, 0)))
		notes = append(notes, i18n.Text("Throwing himself onto the explosive is a Sacrificial Dodge and Drop (BX377). He takes the maximum possible damage, with his DR protecting normally, and his body gives everyone else cover DR %d (his torso DR plus his HP).",
			calculator.CoverDRFromBody(c.target.torsoDR, c.target.hp)))
	case calculator.ExplosiveInsideTarget:
		c.addResult(i18n.Text("Minimum / average / maximum:"), fmt.Sprintf("%d / %d / %d", minimum, average, maximum))
		c.addResult(i18n.Text("Wounding:"), i18n.Text("×3 wounding (vitals), DR does not apply"))
		c.addResult(i18n.Text("DR against the blast:"), i18n.Text("None (internal explosion)"))
		c.addResult(i18n.Text("Penetrating (average):"), fmt.Sprintf("%d", average))
		notes = append(notes, i18n.Text("An explosion set off inside the target is worked out as an attack on the vitals, with a ×3 wounding modifier and no DR at all."))
	default:
		c.addResult(i18n.Text("Minimum / average / maximum:"), fmt.Sprintf("%d / %d / %d", minimum, average, maximum))
		c.addResult(i18n.Text("DR against the blast:"), i18n.Text("%d (Large-Area Injury)", dr))
		c.addResult(i18n.Text("Penetrating (average):"), fmt.Sprintf("%d", max(average-dr, 0)))
	}
	return append(notes, c.fragmentationResults()...)
}

// blastDivisor returns what the rolled damage is divided by for the target, the text explaining it, and whether the
// target is close enough to be hurt at all (BX414-BX415).
func (c *explosionCalculator) blastDivisor(radius int) (divisor fxp.Int, text string, inRange bool) {
	environment := c.environments[c.environmentIndex].Environment
	distance := c.target.distance
	divisor, reach := calculator.BlastDivisor(c.blastDice, c.target.situation(), environment, distance)
	switch reach {
	case calculator.BlastUndivided:
		switch c.target.situation() {
		case calculator.ThrewSelfOnExplosive:
			text = i18n.Text("None (threw himself on the explosive)")
		case calculator.ExplosiveInsideTarget:
			text = i18n.Text("None (the explosive went off inside him)")
		default:
			text = i18n.Text("None (struck directly)")
		}
	case calculator.BlastOutOfRange:
		return divisor, i18n.Text("Out of range (beyond %d yards)", radius), false
	case calculator.BlastAtCenter:
		text = i18n.Text("None (at the center of the blast)")
	default:
		switch environment {
		case calculator.ExplosionUnderwater:
			text = i18n.Text("%s (%s yards, underwater)", divisor.Comma(), distance.Comma())
		case calculator.ExplosionInVacuum:
			text = i18n.Text("%s (10 × %s yards, in vacuum)", divisor.Comma(), distance.Comma())
		default:
			text = i18n.Text("%s (3 × %s yards)", divisor.Comma(), distance.Comma())
		}
	}
	return divisor, text, true
}

// fragmentationResults adds the rows describing the fragments the explosion throws and returns the notes that go with
// them (BX414-BX415). An explosion that lists no fragmentation gets a note about what it throws anyway.
func (c *explosionCalculator) fragmentationResults() []string {
	if calculator.DiceOfDamage(c.fragmentationDice) <= 0 {
		return []string{i18n.Text("An explosive that lists no fragmentation still throws whatever it was sitting on: 1d-4 for ordinary earth, up to 1d for loose scrap.")}
	}
	f := calculator.Fragmentation(c.fragmentationDice, c.target.situation(), c.target.posture(), c.target.distance,
		c.target.sm, c.airburst)
	c.addResult(i18n.Text("Fragmentation radius:"),
		i18n.Text("%d yards (%d dice)", f.Radius, calculator.DiceOfDamage(c.fragmentationDice)))
	var notes []string
	switch {
	case f.Automatic:
		c.addResult(i18n.Text("Fragments hit:"), i18n.Text("automatically (struck directly)"))
	case f.OutOfRange:
		c.addResult(i18n.Text("Fragments hit:"), i18n.Text("Out of range"))
	default:
		text := i18n.Text("on a roll of %d or less", f.Skill)
		// The arithmetic is only worth showing when a modifier changed the base skill; with nothing to add, the total
		// says it all.
		if f.RangePenalty != 0 || f.PosturePenalty != 0 || c.target.sm != 0 {
			parts := []string{
				i18n.Text("base %d", calculator.FragmentationBaseSkill),
				i18n.Text("range %+d", f.RangePenalty),
			}
			if !c.airburst {
				parts = append(parts, i18n.Text("posture %+d", f.PosturePenalty))
			}
			parts = append(parts, i18n.Text("SM %+d", c.target.sm))
			text += " (" + strings.Join(parts, ", ") + ")"
		}
		c.addResult(i18n.Text("Fragments hit:"), text)
		notes = append(notes, i18n.Text("For every %d points by which the fragmentation roll succeeds, one more fragment hits. Roll the hit location of each fragment separately. The only defense against them is diving away (Dodge and Drop, BX377).",
			calculator.ExtraFragmentMargin))
		if f.HalfExposed {
			notes = append(notes, i18n.Text("Crawling or lying down with the blast at ground level, the target shows only half of its torso, and the fragments cannot hit its groin, legs or feet, nor its neck, eyes or face if its head is down: reroll those hit locations (BX551)."))
		}
	}
	c.addResult(i18n.Text("Fragment damage:"),
		i18n.Text("%s cut, per fragment", gurps.FormatDice(c.fragmentationDice, c.useExtraDice())))
	if c.airburst {
		notes = append(notes, i18n.Text("In an airburst the fragments come from above, so posture does not protect against them and only overhead cover does."))
	}
	if c.hotFragments {
		notes = append(notes, i18n.Text("Hot fragments (white phosphorus) each inflict 1d(0.2) burning every 10 seconds for one minute as well."))
	}
	return notes
}

// updateAreaResults works out what an area-effect or cone attack does to the target and returns the notes that go with
// it (BX413-BX414).
func (c *explosionCalculator) updateAreaResults() []string {
	cone := c.attackTypeIndex == coneAttack
	distance := c.target.distance
	spread := calculator.AreaDamageDivisor(distance)
	if cone {
		spread = calculator.ConeWidth(distance, c.coneMaxRange, c.coneMaxWidth)
		c.addResult(i18n.Text("Cone width at %s yards:", distance.Comma()), i18n.Text("%s yards", spread.Comma()))
	}
	divisor, htBonus := calculator.DissipationDivisor(spread, c.dissipates, c.htResisted)
	switch {
	case !c.dissipates:
		c.addResult(i18n.Text("Damage divisor:"), i18n.Text("None (damage does not decline with distance)"))
	case c.htResisted:
		// A dissipating attack that HT resists eases the roll instead of dividing the damage (BX414).
		c.addResult(i18n.Text("Bonus to the HT roll:"), i18n.Text("+%s", htBonus.Comma()))
	default:
		c.addResult(i18n.Text("Damage divisor:"), divisor.Comma())
	}
	c.addResult(i18n.Text("Damage:"), c.damageText(divisor))
	minimum, average, maximum := calculator.DividedDamage(c.blastDice, divisor)
	c.addResult(i18n.Text("Minimum / average / maximum:"), fmt.Sprintf("%d / %d / %d", minimum, average, maximum))
	c.addResult(i18n.Text("DR against the attack:"), i18n.Text("%d (Large-Area Injury)", c.target.dr))
	c.addResult(i18n.Text("Penetrating (average):"), fmt.Sprintf("%d", max(average-c.target.dr, 0)))
	notes := []string{
		i18n.Text("Large-Area Injury (BX400): treat the damage as a torso hit, with no hit location wounding modifier, unless only one location is exposed. Only the locations facing the attack are exposed; those behind cover or masked by the body are not."),
		i18n.Text("An attack that misses an area scatters; the Scatter calculator works out how far."),
		i18n.Text("The only defense is Dodge and Drop (BX377): a dodge at +3 that leaves the target prone. With cover a step away, success reaches it in time; even without cover, the step puts the target a yard farther from the blast."),
	}
	if cone {
		notes = append(notes, i18n.Text("A cone spreads out from its apex, so anything that would block a normal attack from the attacker screens the target from it."))
	} else {
		notes = append(notes, i18n.Text("Area-effect damage does not usually decline with distance; only an attack that dissipates does."))
	}
	return notes
}
