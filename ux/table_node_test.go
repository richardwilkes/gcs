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
	c.True(label.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the click must be consumed")
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
	c.True(label.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the click must be consumed")
	c.Equal(1, clicks, "the click must be reported once")
	c.True(switchCellData.Checked, "the click must flip the checked state")
	drawable, ok := seen.(*unison.DrawableSVG)
	c.True(ok, "the report must see an SVG drawable")
	c.Equal(unison.CheckmarkSVG, drawable.SVG, "the drawable must show the new state before the click is reported")
	c.True(neededLayout, "the layout must be requested before the click is reported")

	label.NeedsLayout = false
	c.True(label.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the second click must be consumed")
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
	c.True(label.MouseDownCallback(geom.Point{}, 1, 1, mod.None), "the click must be consumed")
	c.Equal(3, clicks, "the click must be reported")
	c.False(toggleCellData.Checked, "the click must flip the checked state")
	c.Nil(seen, "the drawable must be cleared before the click is reported")
	c.True(neededLayout, "the layout must be requested before the click is reported")
}
