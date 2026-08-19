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
	"reflect"
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
// only on a character sheet (never in a template, a library list or an editor's table), and only when something in the
// list actually has a switch to throw.
func TestShowSwitchColumnOnlyForSheetsWithSwitchableRows(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	switchable := []*gurps.Trait{newSwitchableTrait(entity, "Claws")}

	c.False(showSwitchColumn(false, entity, switchable),
		"a list that isn't a sheet page must never show the switch column")

	template := gurps.NewTemplate()
	c.False(showSwitchColumn(true, template, switchable),
		"a template must not show the switch column, even with switchable rows")

	plain := gurps.NewTrait(entity, nil, false)
	plain.Name = "Plain"
	plain.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}
	c.False(showSwitchColumn(true, entity, []*gurps.Trait{plain}),
		"a sheet with nothing switchable must not show the switch column")

	c.True(showSwitchColumn(true, entity, switchable), "a sheet with a switchable row must show the switch column")
}

// switchableWeightReducer returns a container holding one child, where the container's contained weight reduction only
// applies while its switch is on. This is the case that matters on a loot sheet, since it changes what the sheet shows
// without a character being involved at all.
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

// TestLootSheetShowsSwitchColumn verifies that a loot sheet gets the switch column, too. A loot sheet never processes
// the features that need a character, but the ones that don't -- a contained weight reduction being the obvious one --
// change what it displays, so the switch that governs them has to be reachable from the list rather than only from the
// item's editor.
func TestLootSheetShowsSwitchColumn(t *testing.T) {
	c := check.New(t)
	loot := gurps.NewLoot()
	plain := gurps.NewEquipment(loot, nil, false)
	plain.Name = "Torch"
	loot.Equipment = []*gurps.Equipment{plain}
	c.False(showSwitchColumn(true, loot, loot.Equipment),
		"a loot sheet with nothing switchable must not show the switch column")

	container := switchableWeightReducer(loot)
	loot.Equipment = []*gurps.Equipment{plain, container}
	c.True(showSwitchColumn(true, loot, loot.Equipment),
		"a loot sheet with a switchable row must show the switch column")

	// The switch governs what the loot sheet displays, which is why the column has to be offered there.
	units := loot.WeightUnit()
	container.SwitchedOn = false
	off := container.ExtendedWeight(false, units)
	container.SwitchedOn = true
	c.NotEqual(off, container.ExtendedWeight(false, units),
		"the switch must change the weight a loot sheet displays")
	container.SwitchedOn = false
}

// TestLootSheetSwitchColumnInPlace verifies that the loot sheet's equipment list really does end up with the switch
// column, through the same provider the sheet itself uses.
func TestLootSheetSwitchColumnInPlace(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	saved := Workspace.DocumentDock
	t.Cleanup(func() { Workspace.DocumentDock = saved })
	Workspace.DocumentDock = NewDocumentDock()

	loot := gurps.NewLoot()
	sheet := NewLootSheet("test"+gurps.LootExt, loot)
	c.Equal(-1, switchColumnIndex(sheet.Equipment.Table.Columns, gurps.EquipmentSwitchColumn),
		"without switchable features, there is no switch column")

	loot.Equipment = []*gurps.Equipment{switchableWeightReducer(loot)}
	sheet.Rebuild(true)
	c.NotEqual(-1, switchColumnIndex(sheet.Equipment.Table.Columns, gurps.EquipmentSwitchColumn),
		"a switchable contained weight reduction must bring the switch column into view")
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

	c.True(label.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the click must be consumed")
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
// contents, bloating the undo edit and making an item that came from a library diverge from its source.
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
// other equipment list -- is still usable. It is drawn dimmed, but the cell has to stay enabled to remain clickable,
// since a disabled panel receives no mouse events at all.
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

	c.True(dimmedLabel.Enabled(), "a dimmed switch cell must stay enabled, or it would never see a click")
	c.NotEqual(reflect.ValueOf(litLabel.DrawCallback).Pointer(), reflect.ValueOf(dimmedLabel.DrawCallback).Pointer(),
		"a dimmed switch cell must not draw itself the way an undimmed one does")

	// The cell has to be part of the sheet's panel tree for the undo manager to be found, which is where the table
	// itself puts it when it lays out the row.
	table.AddChild(dimmedLabel)
	c.True(dimmedLabel.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the click must be consumed")
	c.True(dimmed.SwitchedOn, "clicking a dimmed switch cell must still throw the switch")
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
