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
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/stlimit"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// switchableSTBonus returns an attribute bonus of +1 ST that only applies while its owner's switch is on.
func switchableSTBonus(owner *gurps.Trait) *gurps.AttributeBonus {
	bonus := gurps.NewAttributeBonus(gurps.StrengthID)
	bonus.SetSwitchable(true)
	if owner != nil {
		bonus.SetOwner(owner)
	}
	return bonus
}

// newSwitchableTrait returns a non-container trait carrying a single switchable +1 ST bonus.
func newSwitchableTrait(entity *gurps.Entity, name string) *gurps.Trait {
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	trait.Features = gurps.Features{switchableSTBonus(trait)}
	return trait
}

// stBonusFor returns the total ST bonus the entity currently receives.
func stBonusFor(entity *gurps.Entity) fxp.Int {
	return entity.AttributeBonusFor(gurps.StrengthID, stlimit.None, nil)
}

// TestShowSwitchColumnOnlyForSheetsWithSwitchableRows verifies the conditions under which the switch column appears:
// only on a character sheet (never on a loot sheet, in a template, a library list or an editor's table), and only when
// something in the list actually has a switch to throw.
func TestShowSwitchColumnOnlyForSheetsWithSwitchableRows(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	switchable := []*gurps.Trait{newSwitchableTrait(entity, "Claws")}

	c.False(showSwitchColumn(false, entity, switchable),
		"a list that isn't a sheet page must never show the switch column")

	template := gurps.NewTemplate()
	c.False(showSwitchColumn(true, template, switchable),
		"a template must not show the switch column, even with switchable rows")

	loot := gurps.NewLoot()
	loot.Equipment = []*gurps.Equipment{switchableWeightReducer(loot)}
	c.False(showSwitchColumn(true, loot, loot.Equipment),
		"a loot sheet must not show the switch column, even with switchable rows")

	plain := gurps.NewTrait(entity, nil, false)
	plain.Name = "Plain"
	plain.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}
	c.False(showSwitchColumn(true, entity, []*gurps.Trait{plain}),
		"a sheet with nothing switchable must not show the switch column")

	c.True(showSwitchColumn(true, entity, switchable), "a sheet with a switchable row must show the switch column")
}

// switchableWeightReducer returns a container holding one child, where the container's contained weight reduction only
// applies while its switch is on. This is the strongest case a loot sheet could make for the switch column, since the
// reduction changes what the sheet shows without a character being involved at all, yet the column is still reserved
// for character sheets.
func switchableWeightReducer(owner gurps.DataOwner) *gurps.Equipment {
	container := gurps.NewEquipment(owner, nil, true)
	container.Name = "Bag of Holding"
	reduction := gurps.NewContainedWeightReduction()
	reduction.Reduction = "50%"
	reduction.SetSwitchable(true)
	container.Features = gurps.Features{reduction}
	child := gurps.NewEquipment(owner, container, false)
	child.Name = "Anvil"
	child.BaseWeight = "10 lb"
	child.SetParent(container)
	container.Children = []*gurps.Equipment{child}
	return container
}

// TestLootSheetDoesNotShowSwitchColumn verifies that a loot sheet never gets the switch column, not even when its
// equipment has switchable features that change what the sheet displays. The switch column belongs to character sheets
// alone; everywhere else the switch is thrown from the item's editor.
func TestLootSheetDoesNotShowSwitchColumn(t *testing.T) {
	c := check.New(t)
	loot := gurps.NewLoot()
	plain := gurps.NewEquipment(loot, nil, false)
	plain.Name = "Torch"
	loot.Equipment = []*gurps.Equipment{plain}
	c.False(showSwitchColumn(true, loot, loot.Equipment),
		"a loot sheet with nothing switchable must not show the switch column")

	container := switchableWeightReducer(loot)
	loot.Equipment = []*gurps.Equipment{plain, container}
	c.False(showSwitchColumn(true, loot, loot.Equipment),
		"a loot sheet must not show the switch column, even with a switchable row")

	// The switch still governs what the loot sheet displays; it is simply thrown from the item's editor there.
	units := loot.WeightUnit()
	container.SwitchedOn = false
	off := container.ExtendedWeight(false, units)
	container.SwitchedOn = true
	c.NotEqual(off, container.ExtendedWeight(false, units),
		"the switch must change the weight a loot sheet displays")
	container.SwitchedOn = false
}

// TestLootSheetHasNoSwitchColumnInPlace verifies that the loot sheet's equipment list really is built without the
// switch column, through the same provider the sheet itself uses, whether or not anything in it is switchable.
func TestLootSheetHasNoSwitchColumnInPlace(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	loot := sheet.loot
	c.Equal(-1, switchColumnIndex(sheet.Equipment.Table.Columns, gurps.EquipmentSwitchColumn),
		"without switchable features, there is no switch column")

	loot.Equipment = []*gurps.Equipment{switchableWeightReducer(loot)}
	sheet.Rebuild(true)
	c.Equal(-1, switchColumnIndex(sheet.Equipment.Table.Columns, gurps.EquipmentSwitchColumn),
		"a switchable contained weight reduction must not bring the switch column into a loot sheet")
}

// TestShowSwitchColumnFindsNestedRows verifies that the scan looks at every depth, not just the top level, so that a
// switchable item tucked inside a container still brings the column into view.
func TestShowSwitchColumnFindsNestedRows(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	parent := gurps.NewTrait(entity, nil, true)
	parent.Name = "Container"
	child := newSwitchableTrait(entity, "Claws")
	child.SetParent(parent)
	parent.Children = []*gurps.Trait{child}
	c.True(showSwitchColumn(true, entity, []*gurps.Trait{parent}),
		"a switchable child inside a container must show the switch column")

	// A container's own features are ignored, but those of its modifiers are not.
	modOwner := gurps.NewTrait(entity, nil, true)
	modOwner.Name = "Container With Modifier"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Switchable Modifier"
	modifier.Features = gurps.Features{switchableSTBonus(nil)}
	modOwner.Modifiers = []*gurps.TraitModifier{modifier}
	c.True(showSwitchColumn(true, entity, []*gurps.Trait{modOwner}),
		"a container whose modifier is switchable must show the switch column")
}

// TestSheetShowsSwitchColumnWhenNeeded verifies that each of the sheet's page lists gains its switch column -- in the
// expected position -- as soon as one of its rows has switchable features, and doesn't have it before that.
func TestSheetShowsSwitchColumnWhenNeeded(t *testing.T) {
	t.Run("traits", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		c.Equal(gurps.TraitDescriptionColumn, sheet.Traits.Table.Columns[0].ID,
			"without switchable features, the description comes first")

		entity.Traits = []*gurps.Trait{newSwitchableTrait(entity, "Claws")}
		sheet.Rebuild(true)
		c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
			"the switch column must come first once a trait has switchable features")
	})

	t.Run("skills", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		c.Equal(gurps.SkillDescriptionColumn, sheet.Skills.Table.Columns[0].ID,
			"without switchable features, the description comes first")

		skill := gurps.NewSkill(entity, nil, false)
		skill.Name = "Brawling"
		skill.Features = gurps.Features{switchableSTBonus(nil)}
		entity.Skills = []*gurps.Skill{skill}
		sheet.Rebuild(true)
		c.Equal(gurps.SkillSwitchColumn, sheet.Skills.Table.Columns[0].ID,
			"the switch column must come first once a skill has switchable features")
	})

	t.Run("spells", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		c.Equal(gurps.SpellDescriptionForPageColumn, sheet.Spells.Table.Columns[0].ID,
			"without switchable features, the description comes first")

		spell := gurps.NewSpell(entity, nil, false)
		spell.Name = "Fireball"
		spell.Features = gurps.Features{switchableSTBonus(nil)}
		entity.Spells = []*gurps.Spell{spell}
		sheet.Rebuild(true)
		c.Equal(gurps.SpellSwitchColumn, sheet.Spells.Table.Columns[0].ID,
			"the switch column must come first once a spell has switchable features")
	})

	t.Run("carried equipment", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		c.Equal(gurps.EquipmentEquippedColumn, sheet.CarriedEquipment.Table.Columns[0].ID,
			"the equipped column always comes first for carried equipment")
		c.NotEqual(gurps.EquipmentSwitchColumn, sheet.CarriedEquipment.Table.Columns[1].ID,
			"without switchable features, there is no switch column")

		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Features = gurps.Features{switchableSTBonus(nil)}
		entity.CarriedEquipment = []*gurps.Equipment{eqp}
		sheet.Rebuild(true)
		c.Equal(gurps.EquipmentEquippedColumn, sheet.CarriedEquipment.Table.Columns[0].ID,
			"the equipped column still comes first")
		c.Equal(gurps.EquipmentSwitchColumn, sheet.CarriedEquipment.Table.Columns[1].ID,
			"the switch column must follow the equipped column")
	})

	t.Run("other equipment", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		c.NotEqual(gurps.EquipmentSwitchColumn, sheet.OtherEquipment.Table.Columns[0].ID,
			"without switchable features, there is no switch column")

		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Features = gurps.Features{switchableSTBonus(nil)}
		entity.OtherEquipment = []*gurps.Equipment{eqp}
		sheet.Rebuild(true)
		c.Equal(gurps.EquipmentSwitchColumn, sheet.OtherEquipment.Table.Columns[0].ID,
			"the switch column must come first for other equipment, which has no equipped column")
	})
}

// newSheetWithSwitchableTrait returns a sheet whose entity holds exactly one trait, carrying a switchable +1 ST bonus
// that is currently switched off, along with that trait.
func newSheetWithSwitchableTrait(t *testing.T) (*Sheet, *gurps.Trait) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	return sheet, trait
}

// TestSwitchCellClickTogglesTheSwitch verifies the full path a user takes: the cell in the switch column shows a dash
// while the switch is off, clicking it turns the switch on (bringing the switchable bonus into play and changing the
// drawable to a checkmark), and undoing puts everything back.
func TestSwitchCellClickTogglesTheSwitch(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table

	c.Equal(gurps.TraitSwitchColumn, table.Columns[0].ID, "the switch column must be present")
	c.False(trait.SwitchedOn, "the switch starts out off")
	c.Equal(fxp.Int(0), stBonusFor(entity), "a switchable bonus must not apply while the switch is off")

	rows := table.RootRows()
	c.Equal(1, len(rows), "the traits table must hold the one trait")
	node := rows[0]
	cell := node.ColumnCell(0, 0, unison.Black, unison.White, false, false, false)
	label, ok := cell.(*unison.Label)
	c.True(ok, "the switch cell must be a label")
	drawable, ok := label.Drawable.(*unison.DrawableSVG)
	c.True(ok, "the switch cell must draw an SVG")
	c.Equal(unison.DashSVG, drawable.SVG, "an off switch is drawn as a dash")

	// The cell has to be part of the sheet's panel tree for the undo manager to be found, which is where the table
	// itself puts it when it lays out the row.
	table.AddChild(label)
	c.NotNil(unison.UndoManagerFor(label), "the cell must be able to find the sheet's undo manager")

	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(trait.SwitchedOn, "clicking the cell must turn the switch on")
	c.Equal(fxp.One, stBonusFor(entity), "the switchable bonus must apply once the switch is on")
	drawable, ok = label.Drawable.(*unison.DrawableSVG)
	c.True(ok, "the switch cell must still draw an SVG")
	c.Equal(unison.CheckmarkSVG, drawable.SVG, "an on switch is drawn as a checkmark")

	mgr := unison.UndoManagerFor(label)
	c.True(mgr.CanUndo(), "throwing the switch must be undoable")
	mgr.Undo()
	c.False(trait.SwitchedOn, "undo must turn the switch back off")
	c.Equal(fxp.Int(0), stBonusFor(entity), "undo must take the bonus back out of play")

	c.True(mgr.CanRedo(), "throwing the switch must be redoable")
	mgr.Redo()
	c.True(trait.SwitchedOn, "redo must turn the switch back on")
	c.Equal(fxp.One, stBonusFor(entity), "redo must put the bonus back into play")
}

// TestToggleFeatureSwitchRecalculatesAndIsUndoable verifies the action behind the cell in isolation: it changes the
// state, recalculates the entity, and registers a single undoable edit.
func TestToggleFeatureSwitchRecalculatesAndIsUndoable(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableTrait(t)
	entity := sheet.Entity()
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.False(mgr.CanUndo(), "nothing has been done yet")

	toggleFeatureSwitch(table.RootRows()[0], table, true, false)
	c.True(trait.SwitchedOn, "the switch must be on")
	c.Equal(fxp.One, stBonusFor(entity), "the entity must have been recalculated with the bonus in play")
	c.True(mgr.CanUndo(), "the toggle must be undoable")

	mgr.Undo()
	c.False(trait.SwitchedOn, "undo must turn the switch back off")
	c.Equal(fxp.Int(0), stBonusFor(entity), "undo must recalculate the entity as well")
	c.False(mgr.CanUndo(), "the toggle must have registered exactly one edit")
}

// newSheetWithSwitchableContainer returns a sheet whose entity holds a trait container with a switchable modifier and
// two switchable children, along with the container and its children.
func newSheetWithSwitchableContainer(t *testing.T) (sheet *Sheet, container *gurps.Trait, children []*gurps.Trait) {
	t.Helper()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	container = gurps.NewTrait(entity, nil, true)
	container.Name = "Cybernetics"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Powered"
	modifier.Features = gurps.Features{switchableSTBonus(nil)}
	container.Modifiers = []*gurps.TraitModifier{modifier}
	children = []*gurps.Trait{newSwitchableTrait(entity, "Claws"), newSwitchableTrait(entity, "Armor")}
	for _, child := range children {
		child.SetParent(container)
	}
	container.Children = children
	entity.Traits = []*gurps.Trait{container}
	sheet.Rebuild(true)
	return sheet, container, children
}

// TestToggleFeatureSwitchWithDescendants verifies that an option-click cascades the new state to everything inside the
// container, and that undoing restores every one of them.
func TestToggleFeatureSwitchWithDescendants(t *testing.T) {
	c := check.New(t)
	sheet, parent, children := newSheetWithSwitchableContainer(t)
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

	toggleFeatureSwitch(table.RootRows()[0], table, true, true)
	c.True(parent.SwitchedOn, "the container's switch must be on")
	for i, child := range children {
		c.True(child.SwitchedOn, "child %d's switch must be on", i)
	}
	c.Equal(fxp.FromInteger(3), stBonusFor(sheet.Entity()),
		"the container's modifier and both children must all be contributing")

	mgr.Undo()
	c.False(parent.SwitchedOn, "undo must turn the container's switch back off")
	for i, child := range children {
		c.False(child.SwitchedOn, "undo must turn child %d's switch back off", i)
	}
	c.Equal(fxp.Int(0), stBonusFor(sheet.Entity()), "undo must take all of the bonuses back out of play")
}

// TestToggleFeatureSwitchWithoutDescendants verifies that an ordinary click on a container's switch leaves the items
// inside it alone.
func TestToggleFeatureSwitchWithoutDescendants(t *testing.T) {
	c := check.New(t)
	sheet, parent, children := newSheetWithSwitchableContainer(t)
	table := sheet.Traits.Table

	toggleFeatureSwitch(table.RootRows()[0], table, true, false)
	c.True(parent.SwitchedOn, "the container's switch must be on")
	for i, child := range children {
		c.False(child.SwitchedOn, "child %d's switch must have been left alone", i)
	}
	c.Equal(fxp.One, stBonusFor(sheet.Entity()), "only the container's modifier may be contributing")
}

// TestToggleFeatureSwitchLeavesDescendantsWithNothingToSwitchAlone verifies that an option-click only reaches the
// descendants that actually have something to switch. Throwing the switch of an item with no switchable features would
// change nothing the user could see, yet the new state would still be written to the file, altering the sheet's
// contents and bloating the undo edit.
func TestToggleFeatureSwitchLeavesDescendantsWithNothingToSwitchAlone(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	container := gurps.NewTrait(entity, nil, true)
	container.Name = "Cybernetics"
	modifier := gurps.NewTraitModifier(entity, nil, false)
	modifier.Name = "Powered"
	modifier.Features = gurps.Features{switchableSTBonus(nil)}
	container.Modifiers = []*gurps.TraitModifier{modifier}
	switchable := newSwitchableTrait(entity, "Claws")
	plain := gurps.NewTrait(entity, nil, false)
	plain.Name = "Plain"
	plain.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}
	featureless := gurps.NewTrait(entity, nil, false)
	featureless.Name = "Featureless"
	nested := gurps.NewTrait(entity, nil, true)
	nested.Name = "Nested"
	deep := newSwitchableTrait(entity, "Armor")
	deep.SetParent(nested)
	nested.Children = []*gurps.Trait{deep}
	children := []*gurps.Trait{switchable, plain, featureless, nested}
	for _, child := range children {
		child.SetParent(container)
	}
	container.Children = children
	entity.Traits = []*gurps.Trait{container}
	sheet.Rebuild(true)
	c.Equal(fxp.One, stBonusFor(entity), "only the always-on bonus is in play to start with")

	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

	toggleFeatureSwitch(table.RootRows()[0], table, true, true)
	c.True(container.SwitchedOn, "the container's switch must be on")
	c.True(switchable.SwitchedOn, "a switchable child's switch must be on")
	c.True(deep.SwitchedOn, "a switchable child at any depth must be reached")
	c.False(plain.SwitchedOn, "a child whose features are all always-on must be left alone")
	c.False(featureless.SwitchedOn, "a child with no features at all must be left alone")
	c.False(nested.SwitchedOn, "a container with nothing of its own to switch must be left alone")
	c.Equal(fxp.FromInteger(4), stBonusFor(entity),
		"the container's modifier, both switchable children and the always-on bonus must all be contributing")

	mgr.Undo()
	c.False(container.SwitchedOn, "undo must turn the container's switch back off")
	c.False(switchable.SwitchedOn, "undo must turn the switchable child's switch back off")
	c.False(deep.SwitchedOn, "undo must turn the deeper switchable child's switch back off")
	c.False(plain.SwitchedOn, "undo must leave the untouched child alone as well")
	c.False(featureless.SwitchedOn, "undo must leave the untouched child alone as well")
	c.False(nested.SwitchedOn, "undo must leave the untouched container alone as well")
	c.Equal(fxp.One, stBonusFor(entity), "undo must take the switchable bonuses back out of play")
}

// TestAdjustTargetsWithUnparentedSource verifies that adjusting a value from a panel that isn't part of a window's
// panel tree -- so there is no undo manager and no owner to mark as modified -- still applies the change and
// recalculates the entity rather than panicking. Editors build such detached widgets while they are being assembled.
func TestAdjustTargetsWithUnparentedSource(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	trait := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{trait}
	entity.Recalculate()
	c.Equal(fxp.Int(0), stBonusFor(entity), "the switchable bonus starts out of play")

	source := unison.NewPanel()
	c.Nil(unison.UndoManagerFor(source), "the test requires a source with no undo manager")
	adjustTargets("Toggle Switch", nil, source, entity, []gurps.FeatureSwitcher{trait},
		func(s gurps.FeatureSwitcher) bool { return s.IsSwitchedOn() },
		func(s gurps.FeatureSwitcher, v bool) { s.SetSwitchedOn(v) },
		func(s gurps.FeatureSwitcher) { s.SetSwitchedOn(true) },
		false)

	c.True(trait.SwitchedOn, "the mutation must have been applied")
	c.Equal(fxp.One, stBonusFor(entity), "the entity must have been recalculated")
}

// switchColumnIndex returns the index of the switch column within the given columns, or -1 if it isn't present.
func switchColumnIndex(columns []unison.ColumnInfo, switchColumnID int) int {
	for i, one := range columns {
		if one.ID == switchColumnID {
			return i
		}
	}
	return -1
}

// TestDimmedSwitchCellRemainsClickable verifies that the switch of an item that isn't currently contributing its
// features -- a piece of equipment that isn't equipped, which is the very case the switch column is offered for in the
// other equipment list -- is still usable. Its cell is disabled, which is what draws it dimmed, and that costs it
// nothing: a table hands mouse events to its cells itself, without consulting their enabled state, so the switch can
// still be thrown. Dimming says the switch has no effect at the moment, not that it can't be thrown.
func TestDimmedSwitchCellRemainsClickable(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	dimmed := gurps.NewEquipment(entity, nil, false)
	dimmed.Name = "Powered Armor"
	dimmed.Equipped = false
	dimmed.Features = gurps.Features{switchableSTBonus(nil)}
	lit := gurps.NewEquipment(entity, nil, false)
	lit.Name = "Powered Boots"
	lit.Features = gurps.Features{switchableSTBonus(nil)}
	entity.CarriedEquipment = []*gurps.Equipment{dimmed, lit}
	sheet.Rebuild(true)

	table := sheet.CarriedEquipment.Table
	col := switchColumnIndex(table.Columns, gurps.EquipmentSwitchColumn)
	c.NotEqual(-1, col, "the switch column must be present")

	var cellData gurps.CellData
	dimmed.CellData(gurps.EquipmentSwitchColumn, &cellData)
	c.True(cellData.Dim, "the switch cell of equipment that isn't equipped must be dimmed")
	lit.CellData(gurps.EquipmentSwitchColumn, &cellData)
	c.False(cellData.Dim, "the switch cell of equipment that is equipped must not be dimmed")

	rows := table.RootRows()
	c.Equal(2, len(rows), "the carried equipment table must hold both items")
	dimmedLabel, ok := rows[0].ColumnCell(0, col, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the switch cell must be a label")
	litLabel, ok := rows[1].ColumnCell(1, col, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the switch cell must be a label")

	c.False(dimmedLabel.Enabled(), "a dimmed switch cell is disabled, which is what draws it with the dimming filter")
	c.True(litLabel.Enabled(), "an undimmed switch cell is drawn normally")

	// The cell has to be part of the sheet's panel tree for the undo manager to be found, which is where the table
	// itself puts it when it lays out the row.
	table.AddChild(dimmedLabel)
	mgr := unison.UndoManagerFor(dimmedLabel)
	c.NotNil(mgr, "the cell must be able to find the sheet's undo manager")
	c.True(dimmedLabel.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(dimmed.SwitchedOn, "clicking a dimmed switch cell must still throw the switch")
	c.True(mgr.CanUndo(), "throwing a dimmed switch must be undoable, just like any other")
	c.True(dimmedLabel.MouseUpCallback(geom.Point{}, unison.ButtonLeft, mod.None),
		"the release must be consumed by the dimmed cell as well")
	drawable, ok := dimmedLabel.Drawable.(*unison.DrawableSVG)
	c.True(ok, "the dimmed switch cell must draw an SVG")
	c.Equal(unison.CheckmarkSVG, drawable.SVG, "the dimmed cell must show the switch it just threw as on")
}

// clickSwitchCellThroughTable delivers a primary click to the cell at the given row and column the way a real one
// arrives: at the center of the frame the table itself computes for that cell, through the table's own mouse
// callbacks, which are what locate the cell under the pointer and hand it the event. The table has to have been sized
// beforehand, since those frames come from the column widths and row heights a layout pass produces.
func clickSwitchCellThroughTable[T gurps.Node[T]](table *unison.Table[*Node[T]], row, col int) (down, up bool) {
	where := table.CellFrame(row, col).Center()
	down = table.MouseDownCallback(where, unison.ButtonLeft, 1, mod.None)
	up = table.MouseUpCallback(where, unison.ButtonLeft, mod.None)
	return down, up
}

// TestSwitchCellIsClickableThroughTheTable verifies the property that lets a dimmed switch cell be a disabled panel,
// along the path a user actually takes rather than by calling the cell's own callback directly: the table locates the
// cell under the pointer and hands it the press itself (Table.DefaultMouseDown and friends), never consulting the
// cell's enabled state, and the window only ever sees the table, since a cell is attached to the table for the
// duration of a single event rather than being one of its children. So a dimmed switch is thrown by a click exactly as
// an undimmed one is. Both are driven identically here, so that the dimmed case is measured against a known-good
// baseline, and each ends with a press on an ordinary column that does reach the table's own row selection -- which
// shows the dispatch really is happening rather than the presses going nowhere.
func TestSwitchCellIsClickableThroughTheTable(t *testing.T) {
	for _, tc := range []struct {
		name string
		dim  bool
	}{
		{name: "dimmed", dim: true},
		{name: "undimmed", dim: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			entity := sheet.Entity()
			trait := newSwitchableTrait(entity, "Claws")
			// A trait that is turned off contributes nothing, which is what dims its switch cell; whether it has a
			// switch to throw at all doesn't depend on that.
			trait.Disabled = tc.dim
			entity.Traits = []*gurps.Trait{trait}
			sheet.Rebuild(true)
			table := sheet.Traits.Table
			col := switchColumnIndex(table.Columns, gurps.TraitSwitchColumn)
			c.NotEqual(-1, col, "the switch column must be present")
			mgr := unison.UndoManagerFor(table)
			c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
			c.False(mgr.CanUndo(), "nothing has been done yet")

			// Hit testing needs the table to know how wide its columns and how tall its rows are, which it does even
			// without a window: syncing a page list sizes its columns to their content (see sizePageTableColumns) and
			// syncing a table to its model measures every row, and the rebuild above did both. The frame is checked
			// here so that a change to any of that shows up as this rather than as presses that quietly land nowhere.
			c.False(table.CellFrame(0, col).Empty(), "the switch cell must have a frame to aim at")
			label, ok := table.RootRows()[0].ColumnCell(0, col, unison.Black, unison.White, false, false,
				false).(*unison.Label)
			c.True(ok, "the switch cell must be a label")
			if tc.dim {
				c.False(label.Enabled(), "a dimmed switch cell is disabled, which is what draws it dimmed")
			} else {
				c.True(label.Enabled(), "an undimmed switch cell is drawn normally")
			}

			down, up := clickSwitchCellThroughTable(table, 0, col)
			c.True(down, "the table must let the cell consume the press")
			c.True(up, "the table must let the cell consume the release")
			c.True(trait.SwitchedOn, "a click delivered by the table must throw the switch")
			c.False(table.HasSelection(),
				"the cell must have taken the press before the table's own row selection could run")
			c.True(mgr.CanUndo(), "the click must have registered an undo edit")
			mgr.Undo()
			c.False(trait.SwitchedOn, "undo must turn the switch back off")
			c.False(mgr.CanUndo(), "the click must have registered exactly one edit")

			descCol := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
			c.NotEqual(-1, descCol, "the description column must be present")
			down, _ = clickSwitchCellThroughTable(table, 0, descCol)
			c.True(down, "the press must be consumed")
			c.True(table.HasSelection(),
				"a press that no cell takes must reach the table's own row selection, or nothing was dispatched")
		})
	}
}

// TestOwnerRecalculatesOnlyForItsOwnSheet verifies the check that keeps a single edit from recalculating the entity
// twice: a sheet always recalculates its own entity when it is marked as modified, so the caller must not do so as
// well, while anything else leaves that to the caller.
func TestOwnerRecalculatesOnlyForItsOwnSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()

	c.True(ownerRecalculates(sheet, entity), "a sheet recalculates its own entity when it is marked as modified")
	c.True(ownerRecalculates(sheet.Traits, entity), "a panel within the sheet resolves to the sheet")
	c.False(ownerRecalculates(sheet, gurps.NewEntity()), "another entity isn't the sheet's to recalculate")
	c.False(ownerRecalculates(sheet, nil), "there is no entity to recalculate")
	c.False(ownerRecalculates(nil, entity), "there is no owner to do the recalculation")
	c.False(ownerRecalculates(unison.NewPanel(), entity),
		"a panel that isn't part of a modifiable root can't recalculate anything")
	c.False(ownerRecalculates(NewTemplate("test"+gurps.TemplatesExt, gurps.NewTemplate()), entity),
		"a template doesn't recalculate an entity when it is marked as modified")
}

// newSheetWithSwitchableReaction returns a sheet whose entity holds one trait carrying a switchable reaction bonus that
// is currently switched off, along with that trait. With the switch off, the sheet has no reactions to show, so the
// Reactions list is not on the page.
func newSheetWithSwitchableReaction(t *testing.T) (*Sheet, *gurps.Trait) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Charisma"
	bonus := gurps.NewReactionBonus()
	bonus.Situation = "from everyone"
	bonus.SetSwitchable(true)
	trait.Features = gurps.Features{bonus}
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	return sheet, trait
}

// TestToggleFeatureSwitchRebuildsConditionallyPresentLists verifies that throwing a switch rebuilds the sheet rather
// than merely marking it as modified. A switchable feature can be a reaction bonus, and the Reactions list is only
// carried on the page while there is a reaction to show -- something decided only when the sheet creates its lists --
// so a toggle that only marked the sheet as modified would leave the list missing after switching on, and leave an
// empty list behind after switching off. The rebuild is also the whole of the update: the sheet must not be synced a
// second time on top of it.
func TestToggleFeatureSwitchRebuildsConditionallyPresentLists(t *testing.T) {
	c := check.New(t)
	sheet, trait := newSheetWithSwitchableReaction(t)
	entity := sheet.Entity()
	c.False(listAttachedToSheet(sheet, sheet.Reactions), "with the switch off, the Reactions list must not be on the page")
	table := sheet.Traits.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	counter := installSyncCounter(sheet)

	c.True(toggleFeatureSwitch(table.RootRows()[0], table, true, false), "the trait must have a switch to throw")
	c.True(trait.SwitchedOn, "the switch must be on")
	c.Equal(1, len(entity.Reactions()), "the reaction must be in play once the switch is on")
	c.True(listAttachedToSheet(sheet, sheet.Reactions), "switching on must bring the Reactions list onto the page")
	c.Equal(1, sheet.Reactions.Table.RootRowCount(), "the Reactions list must show the reaction")
	c.Equal(1, counter.count, "throwing the switch must update the sheet exactly once")

	counter.count = 0
	c.True(toggleFeatureSwitch(sheet.Traits.Table.RootRows()[0], sheet.Traits.Table, false, false),
		"the trait must still have a switch to throw")
	c.False(trait.SwitchedOn, "the switch must be off again")
	c.False(listAttachedToSheet(sheet, sheet.Reactions), "switching off must take the Reactions list off the page")
	c.Equal(1, counter.count, "throwing the switch back must update the sheet exactly once")

	counter.count = 0
	c.True(mgr.CanUndo(), "throwing the switch must be undoable")
	mgr.Undo()
	c.True(trait.SwitchedOn, "undo must turn the switch back on")
	c.True(listAttachedToSheet(sheet, sheet.Reactions), "undo must bring the Reactions list back onto the page")
	c.Equal(1, counter.count, "undoing the toggle must update the sheet exactly once")

	counter.count = 0
	mgr.Undo()
	c.False(trait.SwitchedOn, "the second undo must turn the switch off again")
	c.False(listAttachedToSheet(sheet, sheet.Reactions), "the second undo must take the Reactions list off again")
	c.Equal(1, counter.count, "the second undo must update the sheet exactly once")
}

// TestToggleFeatureSwitchDeclinesRowsWithoutASwitch verifies that a row whose data has no switch is left alone and
// reported as such, so that a switch cell -- which flips its drawn state before asking -- can put itself back.
func TestToggleFeatureSwitchDeclinesRowsWithoutASwitch(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	note := gurps.NewNote(entity, nil, false)
	note.MarkDown = "A note has nothing to switch"
	entity.Notes = []*gurps.Note{note}
	sheet.Rebuild(true)
	table := sheet.Notes.Table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

	c.False(toggleFeatureSwitch(table.RootRows()[0], table, true, false), "a note has no switch to throw")
	c.False(mgr.CanUndo(), "declining must not register an undo edit")
}
