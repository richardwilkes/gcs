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
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
)

// TestEquippedCellClickRecalculatesTheSheet verifies that clicking the equipped checkbox brings the item's features
// into play and that undoing takes them back out. The recalculation is left to the sheet, which performs it when it is
// marked as modified, so this covers the path where the cell itself no longer asks for one.
func TestEquippedCellClickRecalculatesTheSheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = "Powered Armor"
	eqp.Equipped = false
	eqp.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)

	table := sheet.CarriedEquipment.Table
	c.Equal(gurps.EquipmentEquippedColumn, table.Columns[0].ID, "the equipped column must come first")
	c.Equal(fxp.Int(0), stBonusFor(entity), "equipment that isn't equipped must not contribute its features")

	rows := table.RootRows()
	c.Equal(1, len(rows), "the carried equipment table must hold the one item")
	label, ok := rows[0].ColumnCell(0, 0, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the equipped cell must be a label")
	c.True(label.Enabled(), "an unchecked equipped cell must be enabled")

	// The cell has to be part of the sheet's panel tree for the undo manager and the sheet itself to be found, which
	// is where the table puts it when it lays out the row.
	table.AddChild(label)
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(eqp.Equipped, "clicking the cell must equip the item")
	c.Equal(fxp.One, stBonusFor(entity), "the entity must have been recalculated with the feature in play")

	mgr := unison.UndoManagerFor(label)
	c.NotNil(mgr, "the cell must be able to find the sheet's undo manager")
	c.True(mgr.CanUndo(), "equipping the item must be undoable")
	mgr.Undo()
	c.False(eqp.Equipped, "undo must unequip the item")
	c.Equal(fxp.Int(0), stBonusFor(entity), "undo must recalculate the entity as well")

	c.True(mgr.CanRedo(), "equipping the item must be redoable")
	mgr.Redo()
	c.True(eqp.Equipped, "redo must equip the item again")
	c.Equal(fxp.One, stBonusFor(entity), "redo must put the feature back into play")
}

// TestCheckCellDrawsTheNewStateBeforeReportingTheClick verifies the order of operations within a check cell's click
// handling. Reporting the click marks the owner as modified, which re-syncs the table and recreates every cell,
// leaving the clicked label detached from the panel tree, where updating its drawable and marking it for layout and
// redraw would accomplish nothing at all. Both have to happen while the label is still the one on screen.
func TestCheckCellDrawsTheNewStateBeforeReportingTheClick(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	node := NewNode(table, nil, gurps.NewTrait(nil, nil, false), false)
	var seen unison.Drawable
	var neededLayout bool
	clicks := 0
	record := func(label *unison.Label, _ mod.Modifiers) {
		clicks++
		seen = label.Drawable
		neededLayout = label.NeedsLayout
	}

	// A switch cell draws both states, so the report must see the new drawable in either direction.
	switchCellData := &gurps.CellData{}
	label := node.newCheckCell(switchCellData, unison.Black,
		func(on bool) *unison.SVG {
			if on {
				return unison.CheckmarkSVG
			}
			return unison.DashSVG
		}, record)

	label.NeedsLayout = false
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.Equal(1, clicks, "the click must be reported once")
	c.True(switchCellData.Checked, "the click must flip the checked state")
	drawable, ok := seen.(*unison.DrawableSVG)
	c.True(ok, "the report must see an SVG drawable")
	c.Equal(unison.CheckmarkSVG, drawable.SVG, "the drawable must show the new state before the click is reported")
	c.True(neededLayout, "the layout must be requested before the click is reported")

	label.NeedsLayout = false
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the second click must be consumed")
	c.Equal(2, clicks, "the second click must be reported")
	c.False(switchCellData.Checked, "the second click must flip the checked state back")
	drawable, ok = seen.(*unison.DrawableSVG)
	c.True(ok, "the report must see an SVG drawable")
	c.Equal(unison.DashSVG, drawable.SVG, "the drawable must show the new state before the click is reported")
	c.True(neededLayout, "the layout must be requested before the click is reported")

	// A toggle cell draws nothing at all when it is off, so the report must see the drawable already cleared.
	toggleCellData := &gurps.CellData{Checked: true}
	label = node.newCheckCell(toggleCellData, unison.Black,
		func(on bool) *unison.SVG {
			if on {
				return unison.CheckmarkSVG
			}
			return nil
		}, record)

	label.NeedsLayout = false
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.Equal(3, clicks, "the click must be reported")
	c.False(toggleCellData.Checked, "the click must flip the checked state")
	c.Nil(seen, "the drawable must be cleared before the click is reported")
	c.True(neededLayout, "the layout must be requested before the click is reported")
}

// TestCheckCellLeavesNonPrimaryPressesToTheTable verifies that only the primary button works a check cell. The switch
// column sits at the very front of the traits, skills and spells page lists, which makes it a natural right-click
// target, and the table runs the cell first and then pops up its context menu: a cell that consumed a right-press
// would throw the switch and hand the user the context menu anyway. The drag and up callbacks have to pass the gesture
// through too, since the table keeps routing them to whichever panel saw the press go down, whether that panel
// consumed the press or not.
func TestCheckCellLeavesNonPrimaryPressesToTheTable(t *testing.T) {
	t.Run("switch cell", func(t *testing.T) {
		c := check.New(t)
		sheet, trait := newSheetWithSwitchableTrait(t)
		table := sheet.Traits.Table
		c.Equal(gurps.TraitSwitchColumn, table.Columns[0].ID, "the switch column must come first")
		mgr := unison.UndoManagerFor(table)
		c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

		// The cells have to be part of the sheet's panel tree for the undo manager to be found, which is where the
		// table itself puts them when it lays out the row.
		for _, button := range []int{unison.ButtonRight, unison.ButtonMiddle} {
			label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
				false).(*unison.Label)
			c.True(ok, "the switch cell must be a label")
			table.AddChild(label)
			c.False(label.MouseDownCallback(geom.Point{}, button, 1, mod.None),
				"a non-primary press must be left for the table, which selects the row and pops up its context menu")
			c.False(trait.SwitchedOn, "a non-primary press must not throw the switch")
			c.Equal(fxp.Int(0), stBonusFor(sheet.Entity()), "a non-primary press must not bring the bonus into play")
			c.False(mgr.CanUndo(), "a non-primary press must not register an undo edit")
			c.False(label.MouseDragCallback(geom.Point{}, button, mod.None),
				"a drag from a press the cell didn't take must be left for the table as well")
			c.False(label.MouseUpCallback(geom.Point{}, button, mod.None),
				"the release of a press the cell didn't take must be left for the table as well")
		}

		label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
			false).(*unison.Label)
		c.True(ok, "the switch cell must be a label")
		table.AddChild(label)
		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None),
			"a primary single click must be consumed")
		c.True(trait.SwitchedOn, "a primary single click must throw the switch")
		c.Equal(fxp.One, stBonusFor(sheet.Entity()), "a primary single click must bring the bonus into play")
		c.True(mgr.CanUndo(), "a primary single click must register an undo edit")
		c.True(label.MouseDragCallback(geom.Point{}, unison.ButtonLeft, mod.None),
			"a drag from a press the cell took must be consumed as well")
		c.True(label.MouseUpCallback(geom.Point{}, unison.ButtonLeft, mod.None),
			"the release of a press the cell took must be consumed as well")
	})

	t.Run("equipped cell", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Equipped = false
		entity.CarriedEquipment = []*gurps.Equipment{eqp}
		sheet.Rebuild(true)

		table := sheet.CarriedEquipment.Table
		c.Equal(gurps.EquipmentEquippedColumn, table.Columns[0].ID, "the equipped column must come first")
		mgr := unison.UndoManagerFor(table)
		c.NotNil(mgr, "the table must be able to find the sheet's undo manager")

		for _, button := range []int{unison.ButtonRight, unison.ButtonMiddle} {
			label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
				false).(*unison.Label)
			c.True(ok, "the equipped cell must be a label")
			table.AddChild(label)
			c.False(label.MouseDownCallback(geom.Point{}, button, 1, mod.None),
				"a non-primary press must be left for the table, which selects the row and pops up its context menu")
			c.False(eqp.Equipped, "a non-primary press must not equip the item")
			c.False(mgr.CanUndo(), "a non-primary press must not register an undo edit")
			c.False(label.MouseDragCallback(geom.Point{}, button, mod.None),
				"a drag from a press the cell didn't take must be left for the table as well")
			c.False(label.MouseUpCallback(geom.Point{}, button, mod.None),
				"the release of a press the cell didn't take must be left for the table as well")
		}
	})
}

// TestCheckCellDoubleClickTogglesOnlyOnce verifies that a fast double-click on a check cell changes the item exactly
// once. The window reports the two presses with click counts of 1 and then 2, so toggling on both would leave the item
// right back where it started along with two undo edits, costing the user two undos to return to a state they never
// left. The second press still has to be consumed, since the table opens the row's editor for a double-click that
// reaches it.
func TestCheckCellDoubleClickTogglesOnlyOnce(t *testing.T) {
	t.Run("switch cell", func(t *testing.T) {
		c := check.New(t)
		sheet, trait := newSheetWithSwitchableTrait(t)
		entity := sheet.Entity()
		table := sheet.Traits.Table
		mgr := unison.UndoManagerFor(table)
		c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
		c.False(mgr.CanUndo(), "nothing has been done yet")

		// The cell has to be part of the sheet's panel tree for the undo manager to be found, which is where the table
		// itself puts it when it lays out the row.
		label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
			false).(*unison.Label)
		c.True(ok, "the switch cell must be a label")
		table.AddChild(label)

		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None),
			"the first click must be consumed")
		c.True(trait.SwitchedOn, "the first click must throw the switch")
		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 2, mod.None),
			"the second click must be consumed, or the table would open the row's editor")
		c.True(trait.SwitchedOn, "the second click of a double-click must leave the switch alone")
		c.Equal(fxp.One, stBonusFor(entity), "the bonus must be in play exactly once")
		drawable, ok := label.Drawable.(*unison.DrawableSVG)
		c.True(ok, "the switch cell must draw an SVG")
		c.Equal(unison.CheckmarkSVG, drawable.SVG, "the cell must still show the switch as on")

		c.True(mgr.CanUndo(), "the double-click must be undoable")
		mgr.Undo()
		c.False(trait.SwitchedOn, "undo must turn the switch back off")
		c.Equal(fxp.Int(0), stBonusFor(entity), "undo must take the bonus back out of play")
		c.False(mgr.CanUndo(), "a double-click must register exactly one edit")
	})

	t.Run("equipped cell", func(t *testing.T) {
		c := check.New(t)
		sheet := newTestSheetForTemplate(t)
		entity := sheet.Entity()
		eqp := gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Equipped = false
		eqp.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)}
		entity.CarriedEquipment = []*gurps.Equipment{eqp}
		sheet.Rebuild(true)

		table := sheet.CarriedEquipment.Table
		mgr := unison.UndoManagerFor(table)
		c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
		c.False(mgr.CanUndo(), "nothing has been done yet")

		label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
			false).(*unison.Label)
		c.True(ok, "the equipped cell must be a label")
		table.AddChild(label)

		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None),
			"the first click must be consumed")
		c.True(eqp.Equipped, "the first click must equip the item")
		c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 2, mod.None),
			"the second click must be consumed, or the table would open the row's editor")
		c.True(eqp.Equipped, "the second click of a double-click must leave the item equipped")
		c.Equal(fxp.One, stBonusFor(entity), "the feature must be in play exactly once")

		c.True(mgr.CanUndo(), "the double-click must be undoable")
		mgr.Undo()
		c.False(eqp.Equipped, "undo must unequip the item")
		c.Equal(fxp.Int(0), stBonusFor(entity), "undo must take the feature back out of play")
		c.False(mgr.CanUndo(), "a double-click must register exactly one edit")
	})
}
