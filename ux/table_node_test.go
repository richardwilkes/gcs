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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fonts"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/display"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
	"github.com/richardwilkes/unison/enums/role"
)

// TestCellDataForAccessibilityLeavesOutTheContainerMarker verifies that a screen reader is given a container's cell
// text without the marker the sort text keeps to group containers ahead of the other rows.
func TestCellDataForAccessibilityLeavesOutTheContainerMarker(t *testing.T) {
	c := check.New(t)
	settings := gurps.GlobalSettings()
	saved := settings.General.GroupContainersOnSort
	settings.General.GroupContainersOnSort = true
	t.Cleanup(func() { settings.General.GroupContainersOnSort = saved })

	entity := gurps.NewEntity()
	container := gurps.NewEquipment(entity, nil, true)
	container.Name = "Backpack"
	entity.CarriedEquipment = []*gurps.Equipment{container}
	provider := NewEquipmentProvider(entity, true, false)
	table := newProviderTable(provider)
	ids := provider.ColumnIDs()
	table.Columns = make([]unison.ColumnInfo, len(ids))
	for i, id := range ids {
		table.Columns[i].ID = id
	}
	rows := table.RootRows()
	c.Equal(1, len(rows), "the table must hold the single container")
	col := slices.Index(ids, gurps.EquipmentDescriptionColumn)
	c.True(col != -1, "the table must have a description column")

	c.Equal(containerMarker+"Backpack", rows[0].CellDataForSort(col),
		"the sort text carries the marker that groups containers first")
	c.Equal("Backpack", rows[0].CellDataForAccessibility(col),
		"a screen reader is given the text the cell shows, without the marker")
}

// TestLabelCellTooltipKeepsTheCellsOwnTextBehindAContradiction verifies which tooltip a label cell shows: a
// contradiction among the prerequisites is explained ahead of the cell's own tooltip rather than in place of it, since
// the row is enabled and in use, while the reason for an unsatisfied prerequisite replaces it, since the row is flagged
// as broken and the reason is what puts it right. The explanation of an unresolved modifier choice goes ahead of the
// rest, behind only where a granted item came from.
func TestLabelCellTooltipKeepsTheCellsOwnTextBehindAContradiction(t *testing.T) {
	c := check.New(t)
	c.Equal("", labelCellTooltip(&gurps.CellData{}), "a cell with nothing to say has no tooltip")
	c.Equal("own", labelCellTooltip(&gurps.CellData{Tooltip: "own"}), "a cell shows its own tooltip")
	c.Equal("why", labelCellTooltip(&gurps.CellData{Tooltip: "own", UnsatisfiedReason: "why"}),
		"the reason for an unsatisfied prerequisite replaces the cell's own tooltip")
	c.Equal("clash", labelCellTooltip(&gurps.CellData{PrereqContradiction: "clash"}),
		"a contradiction is explained on its own when the cell has no tooltip of its own")
	checkTooltipBlocks(c, labelCellTooltip(&gurps.CellData{Tooltip: "own", PrereqContradiction: "clash"}), "clash", "own",
		"a contradiction is explained ahead of the cell's own tooltip, not in place of it")
	c.Equal("granted\n\n---\nwhy", labelCellTooltip(&gurps.CellData{GrantedBy: "granted", UnsatisfiedReason: "why"}),
		"where a granted item came from leads the tooltip")
	c.Equal("pick", labelCellTooltip(&gurps.CellData{UnresolvedChoice: "pick"}),
		"an unresolved choice is explained on its own when the cell has nothing else to say")
	checkTooltipBlocks(c, labelCellTooltip(&gurps.CellData{UnresolvedChoice: "pick", Tooltip: "notes"}), "pick", "notes",
		"an unresolved choice is explained ahead of the cell's own tooltip")
	checkTooltipBlocks(c, labelCellTooltip(&gurps.CellData{
		UnresolvedChoice:  "pick",
		UnsatisfiedReason: "unmet",
		Tooltip:           "notes",
	}), "pick", "unmet", "an unresolved choice is explained ahead of the reason for an unsatisfied prerequisite")
}

// checkTooltipBlocks verifies that the tooltip renders as the two blocks, in that order, with a rule between them. It
// is rendered the way a label cell's tooltip is, so a separator that the Markdown renderer would read as the underline
// of a heading, which is what "---" is directly under a line of text, fails here.
func checkTooltipBlocks(c check.Checker, tooltip, first, second, msg string) {
	c.Helper()
	c.Equal("<p>"+first+"</p>\n<hr>\n<p>"+second+"</p>\n", renderTooltipMarkdown(c, markdownHardLineBreaks(tooltip)),
		msg)
}

// TestCheckCellDrawsTheNewStateBeforeReportingTheClick verifies the order of operations within a check cell's click
// handling. Reporting the click marks the owner as modified, which re-syncs the table and may build the cell anew,
// leaving the clicked label detached from the panel tree, where updating its drawable and marking it for layout and
// redraw would accomplish nothing at all. Both have to happen while the label is still the one on screen.
func TestCheckCellDrawsTheNewStateBeforeReportingTheClick(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	node := NewNode(table, nil, gurps.NewTrait(nil, nil, false), false)
	var seen unison.Drawable
	var neededLayout bool
	clicks := 0
	record := func(label *unison.Label, _ mod.Modifiers) bool {
		clicks++
		seen = label.Drawable
		neededLayout = label.NeedsLayout
		return true
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

// TestCheckCellLeavesNonPrimaryPressesToTheTable verifies that only the primary button works a check cell. A right
// press reaches the cell only as a right-drag unison hands back to the table, and a cell that consumed it would throw
// the switch for a drag. The drag and up callbacks have to pass the gesture through too, since the table keeps routing
// them to whichever panel saw the press go down, whether that panel consumed it or not.
// TestCheckCellRightClickOpensMenuHeadless covers the right-click itself.
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
				"a non-primary press must be left for the table, as a right-drag handed back to it is")
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
				"a non-primary press must be left for the table, as a right-drag handed back to it is")
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
// reaches it. Undoing that one edit takes what the click brought into play back out, and for the equipped cell,
// redoing it puts the item's features back.
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

		c.True(mgr.CanRedo(), "equipping the item must be redoable")
		mgr.Redo()
		c.True(eqp.Equipped, "redo must equip the item again")
		c.Equal(fxp.One, stBonusFor(entity), "redo must put the feature back into play")
	})
}

// TestCheckCellRevertsADeclinedClick verifies that a check cell whose click isn't taken -- the report answers false --
// puts its checked state and drawable back, since it flips both before reporting and would otherwise go on showing a
// state the model never took on.
func TestCheckCellRevertsADeclinedClick(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	node := NewNode(table, nil, gurps.NewTrait(nil, nil, false), false)
	data := &gurps.CellData{}
	clicks := 0
	label := node.newCheckCell(data, unison.Black,
		func(on bool) *unison.SVG {
			if on {
				return unison.CheckmarkSVG
			}
			return unison.DashSVG
		},
		func(_ *unison.Label, _ mod.Modifiers) bool {
			clicks++
			return false
		})
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None),
		"the click must still be consumed, so that the table doesn't act on it")
	c.Equal(1, clicks, "the click must have been reported")
	c.False(data.Checked, "a declined click must leave the checked state as it was")
	drawable, ok := label.Drawable.(*unison.DrawableSVG)
	c.True(ok, "the cell must still draw an SVG")
	c.Equal(unison.DashSVG, drawable.SVG, "a declined click must put the drawable back")
}

// TestEquippedCellClickInAnUnequippedContainerDimsTheCheckmark verifies that a click on the equipped cell of an item
// whose container isn't equipped leaves the cell looking as one built afresh would. The cell's data is dimmed either way
// there, so the click alone decides whether the cell is disabled, and a row kept across the sync goes on showing the
// panel the click flipped.
func TestEquippedCellClickInAnUnequippedContainerDimsTheCheckmark(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	container := gurps.NewEquipment(entity, nil, true)
	container.Name = "Backpack"
	container.Equipped = false
	container.SetOpen(true)
	item := gurps.NewEquipment(entity, container, false)
	item.Name = "Rope"
	item.Equipped = false
	container.Children = []*gurps.Equipment{item}
	entity.CarriedEquipment = []*gurps.Equipment{container}
	sheet.Rebuild(true)

	cell := func() *unison.Label {
		table := sheet.CarriedEquipment.Table
		c.Equal(item, table.RowFromIndex(1).Data(), "the item's row must follow its container's")
		label, ok := table.RowFromIndex(1).ColumnCell(1, 0, unison.Black, unison.White, false, false,
			false).(*unison.Label)
		c.True(ok, "the equipped cell must be a label")
		table.AddChild(label)
		return label
	}

	label := cell()
	c.True(label.Enabled(), "an unchecked cell must stay enabled, so that it can be pressed")
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(item.Equipped, "the click must equip the item")
	label = cell()
	c.NotNil(label.Drawable, "the cell must show the checkmark")
	c.False(label.Enabled(), "the checkmark of an item in an unequipped container must be dimmed")

	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.False(item.Equipped, "the click must unequip the item")
	label = cell()
	c.Nil(label.Drawable, "the cell must show no checkmark")
	c.True(label.Enabled(), "an unchecked cell must stay enabled, so that it can be pressed")
}

// newSheetWithReactionGrantingEquipment returns a sheet whose entity carries a single, unequipped piece of equipment
// with a reaction bonus, along with that equipment. Nothing else supplies a reaction, so the Reactions list is only on
// the page while the item is equipped.
func newSheetWithReactionGrantingEquipment(t *testing.T) (*Sheet, *gurps.Equipment) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = "Badge of Office"
	eqp.Equipped = false
	bonus := gurps.NewReactionBonus()
	bonus.Situation = "from the townsfolk"
	eqp.Features = gurps.Features{bonus}
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)
	return sheet, eqp
}

// TestEquippedCellClickRebuildsConditionallyPresentLists verifies that equipping an item through its cell or the
// Toggle Equipped command rebuilds the sheet, so that a list which only appears while the item contributes -- the
// Reactions list, for equipment granting a reaction bonus -- comes and goes with the equipped state, and that the sheet
// is updated exactly once per toggle and per undo.
func TestEquippedCellClickRebuildsConditionallyPresentLists(t *testing.T) {
	for _, one := range []struct {
		trigger string
		toggle  func(c check.Checker, sheet *Sheet)
	}{
		{"clicking the cell", func(c check.Checker, sheet *Sheet) {
			table := sheet.CarriedEquipment.Table
			label, ok := table.RootRows()[0].ColumnCell(0, 0, unison.Black, unison.White, false, false,
				false).(*unison.Label)
			c.True(ok, "the equipped cell must be a label")
			table.AddChild(label)
			c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
		}},
		{"the Toggle Equipped command", func(_ check.Checker, sheet *Sheet) {
			table := sheet.CarriedEquipment.Table
			table.SelectByIndex(0)
			toggleEquipped(sheet, table)
		}},
	} {
		t.Run(one.trigger, func(t *testing.T) {
			c := check.New(t)
			sheet, eqp := newSheetWithReactionGrantingEquipment(t)
			c.False(listAttachedToSheet(sheet, sheet.Reactions),
				"while unequipped, the Reactions list must not be on the page")
			counter := installSyncCounter(sheet)

			one.toggle(c, sheet)
			c.True(eqp.Equipped, "%s must equip the item", one.trigger)
			c.True(listAttachedToSheet(sheet, sheet.Reactions), "equipping must bring the Reactions list onto the page")
			c.Equal(1, sheet.Reactions.Table.RootRowCount(), "the Reactions list must show the item's reaction")
			c.Equal(1, counter.count, "equipping must update the sheet exactly once")

			counter.count = 0
			one.toggle(c, sheet)
			c.False(eqp.Equipped, "%s must unequip the item again", one.trigger)
			c.False(listAttachedToSheet(sheet, sheet.Reactions), "unequipping must take the Reactions list off the page")
			c.Equal(1, counter.count, "unequipping must update the sheet exactly once")

			mgr := sheet.UndoManager()
			c.NotNil(mgr, "the sheet must have an undo manager")
			counter.count = 0
			c.True(mgr.CanUndo(), "%s must be undoable", one.trigger)
			mgr.Undo()
			c.True(eqp.Equipped, "undo must equip the item again")
			c.True(listAttachedToSheet(sheet, sheet.Reactions), "undo must bring the Reactions list back onto the page")
			c.Equal(1, counter.count, "undo must update the sheet exactly once")
		})
	}
}

// TestMarkdownCellKeepsItsPanelsWhenSelectionChanges verifies that selecting and deselecting a row re-colors the
// markdown in its cells without rebuilding it. A rebuild replaces every panel within the markdown, which gives each a
// new accessibility node and tells a screen reader that the row being left has changed, pulling it back there.
func TestMarkdownCellKeepsItsPanelsWhenSelectionChanges(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Charisma"
	trait.LocalNotes = "Some *notes* with a [link](https://gurpscharactersheet.com)"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)

	table := sheet.Traits.Table
	col := -1
	for i := range table.Columns {
		if table.Columns[i].ID == gurps.TraitDescriptionColumn {
			col = i
			break
		}
	}
	c.True(col >= 0, "the traits table must have a description column")
	rows := table.RootRows()
	c.Equal(1, len(rows), "the traits table must hold the one trait")

	md := findMarkdown(rows[0].ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel())
	c.NotNil(md, "the description cell must show the notes as markdown")
	children := slices.Clone(md.Children())
	c.True(len(children) != 0, "the markdown must have built its content")
	unselectedColor := markdownTextColor(c, md)

	cell := rows[0].ColumnCell(0, col, unison.Red, unison.Blue, true, false, false).AsPanel()
	c.Equal(md, findMarkdown(cell), "selecting the row must reuse the cached markdown")
	c.Equal(children, md.Children(), "selecting the row must not rebuild the markdown")
	c.Equal(unison.Red, markdownTextColor(c, md),
		"the markdown must draw with the selection's foreground")

	rows[0].ColumnCell(0, col, unison.Black, unison.White, false, false, false)
	c.Equal(children, md.Children(), "deselecting the row must not rebuild the markdown")
	c.Equal(unselectedColor, markdownTextColor(c, md),
		"the markdown must go back to its normal color")
}

func findMarkdown(panel *unison.Panel) *unison.Markdown {
	if md, ok := panel.Self.(*unison.Markdown); ok {
		return md
	}
	for _, child := range panel.Children() {
		if md := findMarkdown(child); md != nil {
			return md
		}
	}
	return nil
}

func markdownTextColor(c check.Checker, md *unison.Markdown) unison.Color {
	cp, ok := md.OnBackgroundInk.(unison.ColorProvider)
	c.True(ok, "the markdown's text ink must be a color provider")
	if !ok {
		return 0
	}
	return cp.GetColor()
}

// TestChoiceTagFollowsTheRowColors verifies that the tag describing a template choice is drawn in the row's current
// colors when a cached cell is reused, so a choice whose cell was built while its row was selected doesn't keep the
// selected colors once the row is deselected. The prerequisite tags keep their own fixed colors.
func TestChoiceTagFollowsTheRowColors(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewTraitChoiceContainer(nil, nil)
	unsatisfied := gurps.NewTrait(nil, nil, false)
	unsatisfied.UnsatisfiedReason = "reason"
	template := newTestTemplateWithTraits(choice, unsatisfied)
	table := template.Traits.Table
	rows := table.RootRows()
	c.Equal(2, len(rows))
	findTag := func(cell unison.Paneler) *unison.Tag {
		var found *unison.Tag
		var walk func(p *unison.Panel)
		walk = func(p *unison.Panel) {
			if tag, ok := p.Self.(*unison.Tag); ok && found == nil {
				found = tag
			}
			for _, child := range p.Children() {
				walk(child)
			}
		}
		walk(cell.AsPanel())
		return found
	}

	tag := findTag(rows[0].ColumnCell(0, 0, unison.White, unison.Blue, true, false, false))
	c.NotNil(tag, "the choice's cell must hold its tag")
	c.Equal(unison.Ink(unison.White), tag.BackgroundInk, "a selected row's tag must be drawn in its colors")
	tag = findTag(rows[0].ColumnCell(0, 0, unison.Black, unison.White, false, false, false))
	c.Equal(unison.Ink(unison.Black), tag.BackgroundInk, "the tag must follow the row once it is deselected")
	c.Equal(unison.Ink(unison.White), tag.OnBackgroundInk)

	rows[1].ColumnCell(1, 0, unison.White, unison.Blue, true, false, false)
	tag = findTag(rows[1].ColumnCell(1, 0, unison.Black, unison.White, false, false, false))
	c.NotNil(tag, "the unsatisfied trait's cell must hold its tag")
	c.Equal(unison.Ink(unison.ThemeError), tag.BackgroundInk, "a prerequisite tag must keep its own colors")
}

func TestPageRefLinkTakesNoMoreRoomThanItsText(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Charisma"
	trait.PageRef = "B41"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)

	table := sheet.Traits.Table
	col := columnIndexForID(table, gurps.TraitReferenceColumn)
	c.True(col >= 0, "the traits table must have a page reference column")
	rows := table.RootRows()
	c.Equal(1, len(rows), "the traits table must hold the one trait")
	cell := rows[0].ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	links := uxtest.PanelsMatching(cell, func(p *unison.Panel) bool { return p.Accessibility.Role == role.Link })
	if len(links) != 1 {
		t.Fatalf("the page reference cell must hold one link, but holds %d", len(links))
	}
	link, ok := links[0].Self.(*unison.Label)
	if !ok {
		t.Fatalf("the link must be a label, not a %T", links[0].Self)
	}
	c.Equal("B41", link.String())
	c.False(link.Focusable(), "the focus stays with the table")
	c.Nil(link.Border(), "there is no outline to leave room for")
	plain := unison.NewLabel()
	plain.Font = link.Font
	plain.SetTitle(link.String())
	_, want, _ := plain.Sizes(geom.Size{})
	_, got, _ := link.Sizes(geom.Size{})
	c.Equal(want, got, "the link is the size of its text")
}

// TestLabelCellKeepsTheNotesButtonWithinTheCell verifies that the button which shows and hides a row's notes stays
// inside the cell whatever the cell's width, once the name beside it has to wrap. The lines a name is broken into
// keep the space they were broken at, and a label reports a width rounded up to whole pixels, so a line that fit the
// width it was wrapped to could come out wider than that and push the button past the edge of the cell, where the
// table clipped it.
func TestLabelCellKeepsTheNotesButtonWithinTheCell(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	skill := gurps.NewSkill(entity, nil, false)
	skill.Name = "Area Knowledge"
	skill.Specialization = "Kingdom of Cormyr; Lived there"
	skill.LocalNotes = "General nature of its settlements and towns"
	entity.Skills = []*gurps.Skill{skill}
	sheet.Rebuild(true)

	table := sheet.Skills.Table
	col := columnIndexForID(table, gurps.SkillDescriptionColumn)
	c.True(col >= 0, "the skills table must have a description column")
	rows := table.RootRows()
	c.Equal(1, len(rows), "the skills table must hold the one skill")
	data := rows[0].cellData(col, true)
	c.Equal("Area Knowledge (Kingdom of Cormyr; Lived there)", data.Primary)
	c.NotEqual("", data.Secondary, "the skill's notes must be shown in the cell, or there is no button")
	isButton := func(p *unison.Panel) bool {
		_, ok := p.Self.(*unison.Button)
		return ok
	}
	// Fractional widths, since a cell is usually some fraction of a pixel wide and the rounding is what pushed the
	// button out. The sweep starts where the longest word fits beside the button, since a word too long to fit can't
	// be broken and overflows whatever the width.
	longestWord := unison.NewText("Knowledge", &unison.TextDecoration{Font: fonts.PageFieldPrimary}).Width()
	for width := longestWord + 20; width < 200; width += 0.35 {
		cell := rows[0].CellFromCellData(&data, width, unison.Black, unison.White, false).AsPanel()
		_, pref, _ := cell.Sizes(geom.Size{})
		if pref.Width > width {
			t.Fatalf("at a width of %.2f the cell wants %.2f, which would push its notes button past its edge",
				width, pref.Width)
		}
		cell.SetFrameRect(geom.NewRect(0, 0, width, pref.Height))
		cell.ValidateLayout()
		buttons := uxtest.PanelsMatching(cell, isButton)
		if len(buttons) != 1 {
			t.Fatalf("the cell must hold one notes button, but holds %d", len(buttons))
		}
		right := buttons[0].FrameRect().Right()
		for p := buttons[0].Parent(); p != nil && p != cell; p = p.Parent() {
			right += p.FrameRect().X
		}
		if right > width {
			t.Fatalf("at a width of %.2f the notes button reaches %.2f, past the edge of the cell", width, right)
		}
	}
}

// TestDimmedMarkdownCellFiltersEveryInk verifies that the notes of a dimmed row draw their code, quote bars and links
// as the rest of the row is drawn, through the same color filter as the text.
func TestDimmedMarkdownCellFiltersEveryInk(t *testing.T) {
	c := check.New(t)
	_, table := NewNodeTable(NewTraitsProvider(gurps.NewEntity(), false), nil)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Dim"
	n := NewNode(table, nil, trait, false)
	md, ok := n.createMarkdownCell("`code`\n\n> quote", 200, unison.FieldFont, unison.ThemeOnSurface, unison.ThemeSurface,
		false, true).AsPanel().Self.(*unison.Markdown)
	if !ok {
		t.Fatal("the cell is not a markdown panel")
	}
	filtered := func(ink unison.Ink) bool {
		_, isFiltered := ink.(*unison.ColorFilteredInk)
		return isFiltered
	}
	c.True(filtered(md.OnBackgroundInk))
	c.True(filtered(md.LinkInk))
	c.True(filtered(md.LinkOnPressedInk))
	c.True(filtered(md.CodeBackground), "code blocks are dimmed too")
	c.True(filtered(md.OnCodeBackground))
	for _, ink := range []unison.Ink{
		md.QuoteBarColor, md.QuoteBarNoteColor, md.QuoteBarTipColor,
		md.QuoteBarImportantColor, md.QuoteBarWarningColor, md.QuoteBarCautionColor,
	} {
		c.True(filtered(ink), "as are quote bars")
	}
}

// Clicking the equipped checkmark of disabled equipment is refused, leaving the item and the checkmark as they were.
func TestEquippedCheckmarkRefusesDisabledEquipment(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Equipped = true
	eqp.SetDisabled(true)
	entity.CarriedEquipment = []*gurps.Equipment{eqp}
	_, table := NewNodeTable(NewEquipmentProvider(entity, true, false), nil)
	node := NewNode(table, nil, eqp, false)
	cellData := &gurps.CellData{Checked: true}
	label, ok := node.createToggleCell(cellData, unison.Black).(*unison.Label)
	c.True(ok, "the toggle cell must be a label")

	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	c.True(eqp.Equipped, "the click must not unequip disabled equipment")
	c.True(cellData.Checked, "the checkmark must be put back")
}

// newCellCacheTestSheet returns a sheet holding a single trait of the given name, along with the row the traits table
// shows it in and the index of the table's description column.
func newCellCacheTestSheet(t *testing.T, c check.Checker, name string) (*gurps.Trait, *Node[*gurps.Trait], int) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = name
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	c.True(col >= 0, "the traits table must have a description column")
	row := table.RowFromIndex(0)
	c.NotNil(row, "the traits table must show the trait")
	return trait, row, col
}

// TestNodeKeepsCellDataUntilAChangeIsReported verifies that a row derives its cell data once per reported change
// rather than on every call, which is what spares a draw or a sizing pass the cost of the model, and that reporting a
// change is all it takes for the next call to show it.
func TestNodeKeepsCellDataUntilAChangeIsReported(t *testing.T) {
	c := check.New(t)
	trait, row, col := newCellCacheTestSheet(t, c, "Before")
	before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(slices.Contains(uxtest.LabelTexts(before), "Before"), "the cell must show the trait's name")

	trait.Name = "After"
	again := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(again == before, "until a change is reported, the cell must come back as it was")
	c.True(slices.Contains(uxtest.LabelTexts(again), "Before"), "and must still show the name it was built with")

	invalidateDisplayData()
	after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(after != before, "once a change is reported, the cell must be built afresh")
	c.True(slices.Contains(uxtest.LabelTexts(after), "After"), "and must show the new name")
}

// TestCellKeepsPanelsForFourWidths verifies that a cell keeps the panels built for four widths, since a sync of a page
// list asks for the cells of the column taking the excess width at four: before and after the page reference columns
// take their share, both at the table's preferred width, rounded up to whole pixels, and at the width the layout then
// gives it. Keeping fewer would build those cells anew several times on every sync. The sizing passes, which ask for
// every cell at its natural size before settling on the widths, get a panel of their own that evicts none of those,
// so that measuring the table and then drawing it doesn't build every cell twice.
func TestCellKeepsPanelsForFourWidths(t *testing.T) {
	c := check.New(t)
	_, row, col := newCellCacheTestSheet(t, c, "Trait")
	table := row.table
	widths := []float32{300, 300.25, 250, 250.25}
	panels := make([]unison.Paneler, len(widths))
	var natural unison.Paneler
	for i, width := range widths {
		// The sheet is never laid out, so the column is given a width of its own.
		table.Columns[col].Current = width
		c.True(table.CellWidth(0, col) > 0, "the cell must have a width to fit")
		panels[i] = row.ColumnCell(0, col, unison.Black, unison.White, false, false, false)

		table.Columns[col].Current = 0
		c.True(table.CellWidth(0, col) <= 0, "a sizing pass asks for the cell without a width")
		sized := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false)
		c.True(sized != panels[i], "the natural size must come from a panel of its own")
		if natural == nil {
			natural = sized
		}
		c.True(sized == natural, "asking for the natural size again must reuse that panel")
	}
	for i, width := range widths {
		table.Columns[col].Current = width
		c.True(row.ColumnCell(0, col, unison.Black, unison.White, false, false, false) == panels[i],
			"the panel built for width %v must have been kept", width)
	}
}

// TestLibraryListShowsQuantityAdjustmentWithoutSync verifies that adjusting a quantity in an equipment library list
// shows the new quantity on the next draw. Marking a library list as modified only updates its title, so the table
// keeps its rows, and it is the report of the change that has them derive their cell data afresh.
func TestLibraryListShowsQuantityAdjustmentWithoutSync(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	item := gurps.NewEquipment(nil, nil, false)
	item.Name = "Rope"
	item.Quantity = fxp.One
	d := NewEquipmentTableDockable("test"+gurps.EquipmentExt, []*gurps.Equipment{item})
	col := d.table.ColumnIndexForID(gurps.EquipmentQuantityColumn)
	c.True(col >= 0, "the equipment table must have a quantity column")
	row := d.table.RowFromIndex(0)
	c.NotNil(row, "the equipment table must show the item")
	c.Equal([]string{"1"}, uxtest.LabelTexts(row.ColumnCell(0, col, unison.Black, unison.White, false, false,
		false).AsPanel()), "the quantity cell must show the starting quantity")

	d.table.SelectByIndex(0)
	adjustQuantity(d, d.table, true)
	c.Equal(fxp.Two, item.Quantity, "the quantity must have been incremented")
	c.True(d.table.RowFromIndex(0) == row, "nothing may have synced the table, or this test covers nothing")
	c.Equal([]string{"2"}, uxtest.LabelTexts(row.ColumnCell(0, col, unison.Black, unison.White, false, false,
		false).AsPanel()), "the same row must show the new quantity")
}

// TestCheckCellRightClickOpensMenuHeadless verifies that a right-click on the check box at the front of a row -- a
// trait's switch or an item's equipped box -- opens the list's menu for that row without toggling the box or recording
// anything to undo.
func TestCheckCellRightClickOpensMenuHeadless(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	screen.EnableAccessibility()
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	var trait *gurps.Trait
	var eqp *gurps.Equipment
	var mgr *unison.UndoManager
	traitRow, eqpRow := -1, -1
	var switchColumnFirst, equippedColumnFirst bool
	screen.Do(func() {
		entity := sheet.Entity()
		trait = newSwitchableTrait(entity, "Claws")
		entity.Traits = append(entity.Traits, trait)
		eqp = gurps.NewEquipment(entity, nil, false)
		eqp.Name = "Powered Armor"
		eqp.Equipped = false
		entity.CarriedEquipment = append(entity.CarriedEquipment, eqp)
		entity.Recalculate()
		sheet.Rebuild(true)
		mgr = sheet.UndoManager()
		traits := sheet.Traits.Table
		switchColumnFirst = traits.Columns[0].ID == gurps.TraitSwitchColumn
		for i := 0; i <= traits.LastRowIndex(); i++ {
			if traits.RowFromIndex(i).Data() == trait {
				traitRow = i
			}
		}
		carried := sheet.CarriedEquipment.Table
		equippedColumnFirst = carried.Columns[0].ID == gurps.EquipmentEquippedColumn
		for i := 0; i <= carried.LastRowIndex(); i++ {
			if carried.RowFromIndex(i).Data() == eqp {
				eqpRow = i
			}
		}
	})
	c.True(switchColumnFirst, "the switch column must come first in the traits list")
	c.True(equippedColumnFirst, "the equipped column must come first in the carried equipment list")
	if traitRow < 0 || eqpRow < 0 {
		t.Fatal("the lists must have rows for the trait and the equipment")
	}
	c.NotNil(mgr, "the sheet must have an undo manager")

	screen.ClickWith(cellScreenPoint(t, screen, wnd, sheet.Traits.Table, traitRow, gurps.TraitSwitchColumn),
		unison.ButtonRight, mod.None)
	titles := contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, openEditorAction.Title),
		"a right-click on the switch must open the traits list's menu; it offers %v", titles)
	var switchedOn, canUndo bool
	var selected []*gurps.Trait
	var bonus fxp.Int
	screen.Do(func() {
		switchedOn = trait.SwitchedOn
		bonus = stBonusFor(sheet.Entity())
		canUndo = mgr.CanUndo()
		for _, row := range sheet.Traits.Table.SelectedRows(false) {
			selected = append(selected, row.Data())
		}
	})
	c.False(switchedOn, "a right-click on the switch must not throw it")
	c.Equal(fxp.Int(0), bonus, "nor bring the switchable bonus into play")
	c.False(canUndo, "nor record anything to undo")
	c.Equal([]*gurps.Trait{trait}, selected, "a right-click on the switch must select its row for the menu")
	closeContextMenu(c, screen, wnd)

	screen.ClickWith(cellScreenPoint(t, screen, wnd, sheet.CarriedEquipment.Table, eqpRow,
		gurps.EquipmentEquippedColumn), unison.ButtonRight, mod.None)
	titles = contextMenuTitles(t, screen, wnd)
	c.True(slices.Contains(titles, openEditorAction.Title),
		"a right-click on the equipped box must open the equipment list's menu; it offers %v", titles)
	var equipped bool
	var selectedEqp []*gurps.Equipment
	screen.Do(func() {
		equipped = eqp.Equipped
		canUndo = mgr.CanUndo()
		for _, row := range sheet.CarriedEquipment.Table.SelectedRows(false) {
			selectedEqp = append(selectedEqp, row.Data())
		}
	})
	c.False(equipped, "a right-click on the equipped box must not equip the item")
	c.False(canUndo, "nor record anything to undo")
	c.Equal([]*gurps.Equipment{eqp}, selectedEqp, "a right-click on the equipped box must select its row for the menu")
	closeContextMenu(c, screen, wnd)
}

// TestLockedPickKeepsItsCheckmark verifies that clicking the checkmark of the pick of a mandatory choice on a sheet
// leaves both the modifier and the checkmark drawn for it as they were, rather than clearing the mark over a pick that
// is still in force.
func TestLockedPickKeepsItsCheckmark(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	e, _, table := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"}, "A"))
	pick := table.RootRows()[0].Children()[0]
	label, ok := pick.ColumnCell(0, 0, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the enabled cell must be a label")
	c.NotNil(label.Drawable, "the pick starts out checked")
	table.AddChild(label)
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	label.RemoveFromParent()
	c.NotNil(label.Drawable, "the checkmark must stay, since the pick can't be turned off")
	c.True(e.editorData.Modifiers[0].Children[0].Enabled())
	c.False(unison.UndoManagerFor(table).CanUndo(), "a refused click records nothing")
}

// insertNewTraitModifier inserts a new trait modifier into an editor's table of modifiers, as "New Trait Modifier"
// does, without opening an editor for it, and returns it.
func insertNewTraitModifier(t *testing.T, owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]]) *gurps.TraitModifier {
	t.Helper()
	provider, ok := table.Model.(*modifiersProvider[*gurps.TraitModifier])
	if !ok {
		t.Fatal("expected a trait modifiers provider")
	}
	item := gurps.NewTraitModifier(provider.DataOwner(), nil, false)
	item.Name = "New"
	provider.insertItems(owner, table, item)
	return item
}

// TestNewModifierInAChoiceArrivesOff verifies that a new modifier inserted into a choice arrives turned off: into an
// optional choice, into a mandatory choice on a sheet with no pick, which is left without one and flagged, and into a
// closed group within a choice, which keeps its pick. Undo takes it away.
func TestNewModifierInAChoiceArrivesOff(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	withGroup := newTraitModifierChoiceFor(entity, true, []string{"y"}, "y")
	group := gurps.NewTraitModifier(entity, withGroup, true)
	x := gurps.NewTraitModifier(entity, group, false)
	x.SetDisabled(true)
	group.Children = []*gurps.TraitModifier{x}
	group.SetOpen(false)
	withGroup.Children = slices.Insert(withGroup.Children, 0, group)
	e, _, table := traitEditorOnSheet(t, sheet, withGroup,
		newTraitModifierChoiceFor(entity, false, []string{"a"}),
		newTraitModifierChoiceFor(entity, true, []string{"b"}))
	optional := e.editorData.Modifiers[1]
	table.SetSelectionMap(map[tid.TID]bool{optional.ID(): true})
	item := insertNewTraitModifier(t, e, table)
	c.Equal(item, optional.Children[1])
	c.False(item.Enabled(), "the new modifier arrives turned off in an optional choice")

	mandatory := e.editorData.Modifiers[2]
	table.SetSelectionMap(map[tid.TID]bool{mandatory.Children[0].ID(): true})
	item = insertNewTraitModifier(t, e, table)
	c.Equal(item, mandatory.Children[1])
	c.False(item.Enabled(), "and in a mandatory one")
	c.Equal([]*gurps.TraitModifier{mandatory}, gurps.UnresolvedModifierChoices(e.editorData.Modifiers...),
		"which is left without its pick")

	groupCopy := e.editorData.Modifiers[0].Children[0]
	table.SetSelectionMap(map[tid.TID]bool{groupCopy.ID(): true})
	item = insertNewTraitModifier(t, e, table)
	c.Equal(item, groupCopy.Children[1], "the new modifier goes into the group")
	c.False(item.Enabled(), "and arrives turned off there too")
	c.True(e.editorData.Modifiers[0].Children[1].Enabled(), "y stays the pick")
	unison.UndoManagerFor(table).Undo()
	restored := liveTable(table).RootRows()[0].Data().Children[0]
	c.Equal(1, len(restored.Children), "undo takes the new modifier away")
}

// newReuseTestSheet returns a sheet, shown as one continuous page, whose traits list is the named traits, the one at
// boxIndex (if any) being a container holding a trait named "Inside", along with the traits at the top of the list.
func newReuseTestSheet(t *testing.T, boxIndex int, names ...string) (*Sheet, []*gurps.Trait) {
	t.Helper()
	sheet := newContinuousTestSheet(t)
	entity := sheet.Entity()
	traits := make([]*gurps.Trait, len(names))
	for i, name := range names {
		traits[i] = gurps.NewTrait(entity, nil, i == boxIndex)
		traits[i].Name = name
		if i == boxIndex {
			inside := gurps.NewTrait(entity, traits[i], false)
			inside.Name = "Inside"
			traits[i].Children = []*gurps.Trait{inside}
		}
	}
	entity.Traits = traits
	sheet.Rebuild(true)
	return sheet, traits
}

// updateSheet reports a change to the sheet and then lays its page out, as the window would before drawing it. A sync
// sizes the columns to fit the table's preferred width, which is rounded up to whole pixels, so until the page has
// been laid out again the cells are asked for at widths that differ from the settled ones by that rounding.
func updateSheet(sheet *Sheet) {
	sheet.MarkModified(sheet)
	sheet.page.ValidateLayout()
}

// traitDescriptionCells returns each row the traits table shows, in order, along with the panel of its description
// cell.
func traitDescriptionCells(c check.Checker, table *unison.Table[*Node[*gurps.Trait]]) ([]*Node[*gurps.Trait], []*unison.Panel) {
	col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
	c.True(col >= 0, "the traits table must have a description column")
	count := table.LastRowIndex() + 1
	rows := make([]*Node[*gurps.Trait], count)
	panels := make([]*unison.Panel, count)
	for i := range count {
		rows[i] = table.RowFromIndex(i)
		panels[i] = rows[i].ColumnCell(i, col, unison.Black, unison.White, false, false, false).AsPanel()
	}
	return rows, panels
}

// TestMovedRowGetsANewNodeWithTheRightParent verifies that a row whose data moves to another container is given a new
// node there rather than its old one, which still names the old container as its parent, while the rows that stayed
// put are kept.
func TestMovedRowGetsANewNodeWithTheRightParent(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	boxA := gurps.NewTrait(entity, nil, true)
	boxA.Name = "A"
	boxB := gurps.NewTrait(entity, nil, true)
	boxB.Name = "B"
	child := gurps.NewTrait(entity, boxA, false)
	child.Name = "Child"
	boxA.Children = []*gurps.Trait{child}
	entity.Traits = []*gurps.Trait{boxA, boxB}
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	c.True(boxA.IsOpen() && boxB.IsOpen(), "the containers must be open, so that their children are shown")
	rowA, oldChild, rowB := table.RowFromIndex(0), table.RowFromIndex(1), table.RowFromIndex(2)
	c.True(oldChild.Data() == child && oldChild.Parent() == rowA, "the child must start out under A")

	boxA.SetChildren(nil)
	boxB.SetChildren([]*gurps.Trait{child})
	child.SetParent(boxB)
	sheet.Rebuild(true)
	c.True(table == sheet.Traits.Table, "the rebuild must have kept the table, or this test covers nothing")
	c.True(table.RowFromIndex(0) == rowA && table.RowFromIndex(1) == rowB, "the containers must have been kept")
	newChild := table.RowFromIndex(2)
	c.True(newChild.Data() == child, "the child must now be shown under B")
	c.True(newChild.Parent() == rowB, "the child's row must have B's row as its parent")
	c.True(newChild != oldChild, "the child must have been given a new node")
	c.Equal(0, len(rowA.Children()), "A's row must no longer hold the child")
}

// TestNotesToggleRebuildsAReusedCell verifies that showing or hiding a row's notes rebuilds its description cell in a
// list that keeps the row across the sync that follows the toggle, as a sheet's list and a filtered library list do.
// The cell's data stays the same, so only the notes disclosure kept with the panel tells the panel built before the
// toggle from the one wanted after it.
func TestNotesToggleRebuildsAReusedCell(t *testing.T) {
	for _, one := range []struct {
		name  string
		table func(*testing.T, check.Checker) *unison.Table[*Node[*gurps.Trait]]
	}{
		{
			name: "sheet",
			table: func(t *testing.T, _ check.Checker) *unison.Table[*Node[*gurps.Trait]] {
				sheet := newTestSheetForTemplate(t)
				entity := sheet.Entity()
				entity.SheetSettings.NotesDisplay = display.Inline
				trait := gurps.NewTrait(entity, nil, false)
				trait.Name = "Charisma"
				trait.LocalNotes = "A note"
				entity.Traits = []*gurps.Trait{trait}
				sheet.Rebuild(true)
				return sheet.Traits.Table
			},
		},
		{
			name: "filtered library list",
			table: func(t *testing.T, c check.Checker) *unison.Table[*Node[*gurps.Trait]] {
				uxtest.SwapForTest(t, &gurps.GlobalSettings().SheetSettings().NotesDisplay, display.Inline)
				registerKeyBindingsOnce.Do(func() { registerActions() })
				trait := newTraitForTest("Charisma", nil)
				trait.LocalNotes = "A note"
				d := NewTraitTableDockable("test"+gurps.TraitsExt, []*gurps.Trait{trait})
				d.table.ApplyFilter(func(*Node[*gurps.Trait]) bool { return false })
				c.True(d.table.IsFiltered(), "the table must be filtered")
				return d.table
			},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			table := one.table(t, c)
			row := table.RowFromIndex(0)
			c.NotNil(row, "the table must show the trait")
			key := row.noteKey()
			wasClosed := gurps.IsClosed(key)
			t.Cleanup(func() { gurps.SetClosedState(key, wasClosed) })
			gurps.SetClosedState(key, false)
			col := table.ColumnIndexForID(gurps.TraitDescriptionColumn)
			c.True(col >= 0, "the traits table must have a description column")

			before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
			c.True(findMarkdown(before) != nil, "the cell must show the open notes")
			button, ok := uxtest.FirstPanelOfType[*unison.Button](before)
			c.True(ok, "the cell must hold the notes disclosure button")

			button.ClickCallback()
			c.True(gurps.IsClosed(key), "clicking the button must close the notes")
			c.True(table.RowFromIndex(0) == row, "the list must have kept its row across the sync")
			after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
			c.True(after != before, "the cell must be rebuilt once its notes are closed")
			c.True(findMarkdown(after) == nil, "and must no longer show them")
		})
	}
}

// TestNotesToggleFollowsAConvertedRow verifies that the notes disclosure button of a kept cell acts on the row's notes
// as they are keyed now. Converting an item to or from a container changes its ID, and with it the key its notes'
// disclosure is kept under, while its row and, when the disclosure is the same under both keys, its cell are kept.
func TestNotesToggleFollowsAConvertedRow(t *testing.T) {
	c := check.New(t)
	sheet := newContinuousTestSheet(t)
	entity := sheet.Entity()
	entity.SheetSettings.NotesDisplay = display.Inline
	eqp := gurps.NewEquipment(entity, nil, false)
	eqp.Name = "Backpack"
	eqp.LocalNotes = "A note"
	// Already a container, so that the list leaves room for the disclosure triangle and the conversion doesn't change
	// the width of the cell.
	bag := gurps.NewEquipment(entity, nil, true)
	bag.Name = "Bag"
	entity.CarriedEquipment = []*gurps.Equipment{eqp, bag}
	sheet.Rebuild(true)
	sheet.page.ValidateLayout()
	table := sheet.CarriedEquipment.Table
	row := table.RowFromIndex(0)
	c.NotNil(row, "the table must show the equipment")
	oldKey := row.noteKey()
	col := table.ColumnIndexForID(gurps.EquipmentDescriptionColumn)
	c.True(col >= 0, "the equipment table must have a description column")
	before := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()

	table.SetSelectionMap(map[tid.TID]bool{eqp.ID(): true})
	ConvertToContainer(sheet, table)
	sheet.page.ValidateLayout()
	c.True(eqp.Container(), "the equipment must have become a container")
	c.True(table.RowFromIndex(0) == row, "the list must have kept its row across the conversion")
	newKey := row.noteKey()
	c.NotEqual(oldKey, newKey, "the conversion must have changed the key the notes' disclosure is kept under")
	for _, key := range []string{oldKey, newKey} {
		wasClosed := gurps.IsClosed(key)
		t.Cleanup(func() { gurps.SetClosedState(key, wasClosed) })
		gurps.SetClosedState(key, false)
	}
	kept := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(kept == before, "the cell must have been kept, or this test covers nothing")
	button, ok := uxtest.FirstPanelOfType[*unison.Button](kept)
	c.True(ok, "the cell must hold the notes disclosure button")

	button.ClickCallback()
	c.True(gurps.IsClosed(newKey), "clicking the button must close the notes of the converted row")
	c.False(gurps.IsClosed(oldKey), "and must leave the key the row no longer has alone")
	after := row.ColumnCell(0, col, unison.Black, unison.White, false, false, false).AsPanel()
	c.True(findMarkdown(after) == nil, "the cell must no longer show the notes")
}

// TestFontChangeRebuildsCells verifies that a change of fonts rebuilds the panels a row keeps, although the data of
// its cells is unchanged, since their text was wrapped and measured with the old fonts.
func TestFontChangeRebuildsCells(t *testing.T) {
	c := check.New(t)
	sheet, _ := newReuseTestSheet(t, -1, "Alpha")
	rows, panels := traitDescriptionCells(c, sheet.Traits.Table)
	again, againPanels := traitDescriptionCells(c, sheet.Traits.Table)
	c.True(again[0] == rows[0] && againPanels[0] == panels[0], "until the fonts change, the cell must be kept")

	invalidateCellPanels()
	after, afterPanels := traitDescriptionCells(c, sheet.Traits.Table)
	c.True(after[0] == rows[0], "the row must be the same object")
	c.True(afterPanels[0] != panels[0], "the cell must be built anew once the fonts change")
	c.True(slices.Contains(uxtest.LabelTexts(afterPanels[0]), "Alpha"), "and must still show the trait")
}

// newEquipmentTable returns a table holding a single piece of carried equipment whose weight has more decimal places
// than the sheet is told to show, along with the node for that row and the index of the weight column. forPage says
// whether the table is one of a sheet's page lists or a library-style table.
func newEquipmentTable(t *testing.T, forPage bool) (node *Node[*gurps.Equipment], weightCol int) {
	t.Helper()
	c := check.New(t)
	entity := gurps.NewEntity()
	entity.SheetSettings.EquipmentWeightFormat = fxp.NumberFormat{Places: fxp.TwoPlaces}
	e := gurps.NewEquipment(entity, nil, false)
	e.Name = "Heavy Thing"
	e.BaseWeight = exactWeightText
	entity.CarriedEquipment = []*gurps.Equipment{e}

	provider := NewEquipmentProvider(entity, true, forPage)
	table := newProviderTable(provider)
	ids := provider.ColumnIDs()
	table.Columns = make([]unison.ColumnInfo, len(ids))
	for i, id := range ids {
		table.Columns[i].ID = id
	}
	rows := table.RootRows()
	c.Equal(1, len(rows), "the table must hold the single piece of equipment")
	weightCol = slices.Index(ids, gurps.EquipmentWeightColumn)
	c.True(weightCol != -1, "the table must have a weight column")
	return rows[0], weightCol
}

// TestEquipmentPageCellDataRoundsOnlyForDisplay verifies that a page table's weight cells show the rounded weight, with
// the exact weight available as a tooltip, while the text the rows are sorted by stays exact. Sorting on the rounded
// text would tie rows whose weights actually differ, leaving them in an arbitrary order.
func TestEquipmentPageCellDataRoundsOnlyForDisplay(t *testing.T) {
	c := check.New(t)
	node, weightCol := newEquipmentTable(t, true)

	displayed := node.cellData(weightCol, true)
	c.Equal(roundedWeightText, displayed.Primary, "a page cell shows the rounded weight")
	c.Equal(exactWeightText, displayed.Tooltip, "the exact weight is available as a tooltip")

	exact := node.cellData(weightCol, false)
	c.Equal(exactWeightText, exact.Primary, "a cell that isn't for display holds the exact weight")
	c.Equal("", exact.Tooltip, "there is nothing for a tooltip to add when the exact weight is shown")

	c.Equal(exactWeightText, node.CellDataForSort(weightCol), "sorting must order by the exact weight")
}

// TestEquipmentRowsMatchDisplayedAndExactText verifies that searching finds a page row by the rounded weight the user
// can see as well as by the exact weight behind it, while a row off the page -- where the two renderings are the same
// -- matches only the exact text. The rounded text is chosen so that it is not simply a prefix of the exact text.
func TestEquipmentRowsMatchDisplayedAndExactText(t *testing.T) {
	c := check.New(t)
	node, _ := newEquipmentTable(t, true)
	c.True(node.Match(roundedWeightText), "the sheet search must find the weight as displayed")
	c.True(node.Match("7.5127"), "the sheet search must find the exact weight")
	c.False(node.Match("7.52"), "a value that is neither rendering must not match")

	node, _ = newEquipmentTable(t, false)
	c.False(node.Match(roundedWeightText), "off the page, there is no rounded rendering to match")
	c.True(node.Match("7.5127"), "off the page, the exact weight still matches")
}

// TestInsertItemsSyncsTheSheetOnlyOnce verifies that inserting items into a sheet's list, an edit that used to report
// itself more than once, syncs the sheet a single time while still bumping its modification timestamp, which is the
// one thing a rebuild on its own leaves out.
func TestInsertItemsSyncsTheSheetOnlyOnce(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := newSwitchableTrait(entity, "Claws")
	before := len(entity.Traits)
	entity.ModifiedOn = jio.Time{}
	counter := installSyncCounter(sheet)

	InsertItems(sheet, sheet.Traits.Table, entity.TraitList, entity.SetTraitList,
		func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] {
			return sheet.Traits.provider.RootRows()
		},
		trait)

	c.Equal(before+1, len(entity.Traits), "the item must be inserted into the entity")
	c.Equal(1, counter.count, "an insertion must sync the sheet exactly once")
	c.Equal(gurps.TraitSwitchColumn, sheet.Traits.Table.Columns[0].ID,
		"the rebuild must bring the switch column in with the switchable item")
	c.NotEqual(jio.Time{}, entity.ModifiedOn, "an insertion must bump the modification timestamp")
	mgr := unison.UndoManagerFor(sheet.Traits.Table)
	c.NotNil(mgr, "the table must be able to find the sheet's undo manager")
	c.True(mgr.CanUndo(), "the insertion must be undoable")
}

func columnIndexForID[T unison.TableRowConstraint[T]](table *unison.Table[T], id int) int {
	for i, col := range table.Columns {
		if col.ID == id {
			return i
		}
	}
	return -1
}

// The table synthesizes no click for Space at cell level, so page reference and check cells must answer the press
// themselves; a cell with nothing to press leaves Space to open the row's editor.
func TestSpaceOnACellWorksWhatTheCellHolds(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	const ref = "https://example.com/ref"
	var traits *unison.Table[*Node[*gurps.Trait]]
	var equipment *unison.Table[*Node[*gurps.Equipment]]
	var eq *gurps.Equipment
	screen.Do(func() {
		entity := sheet.Entity()
		tr := gurps.NewTrait(entity, nil, false)
		tr.Name = "Alpha"
		tr.PageRef = ref
		// Training grants Night Vision, whose notes the sheet shows inline, beside the mark of a granted row.
		training := gurps.NewTrait(entity, nil, false)
		training.Name = "Training"
		g := gurps.NewGrant(feature.GrantTrait)
		nightVision := newGrantTestTrait("Night Vision", 0)
		nightVision.LocalNotes = "Seen in the dark"
		g.SetItem(nightVision, gurps.LibraryFile{})
		training.Features = gurps.Features{g}
		// Replaces the new sheet's own traits, whose book references would prompt for a PDF if opened.
		entity.Traits = []*gurps.Trait{tr, training}
		eq = gurps.NewEquipment(entity, nil, false)
		eq.Name = "Rope"
		eq.Equipped = true
		entity.CarriedEquipment = append(entity.CarriedEquipment, eq)
		entity.Recalculate()
		sheet.Rebuild(true)
		traits, _ = uxtest.FirstPanelOfType[*unison.Table[*Node[*gurps.Trait]]](sheet.AsPanel())
		for _, one := range uxtest.PanelsOfType[*unison.Table[*Node[*gurps.Equipment]]](sheet.AsPanel()) {
			if one.RootRowCount() > 0 {
				equipment = one
				break
			}
		}
	})
	if traits == nil || equipment == nil {
		t.Fatal("the character sheet must show the traits and the carried equipment")
	}
	dockables := func() (count int) {
		screen.Do(func() { count = len(AllDockables()) })
		return count
	}
	before := dockables()
	spaceOn := func(table unison.Paneler, place func() bool) {
		t.Helper()
		var placed bool
		screen.Do(func() {
			placed = place()
			table.AsPanel().RequestFocus()
		})
		if !placed {
			t.Fatal("the cell cursor must go where it is put")
		}
		screen.KeyPress(unison.KeySpace, mod.None)
	}

	refCol := columnIndexForID(traits, gurps.TraitReferenceColumn)
	if refCol < 0 {
		t.Fatal("the traits must show their page references")
	}
	c.Nil(screen.OpenedURLs(), "nothing has asked for the browser yet")
	spaceOn(traits, func() bool { return traits.SetLeadCell(0, refCol) })
	c.Equal([]string{ref}, screen.OpenedURLs(), "space on a page reference opens it")
	c.Equal(before, dockables(), "and opens no editor")

	equippedCol := columnIndexForID(equipment, gurps.EquipmentEquippedColumn)
	if equippedCol < 0 {
		t.Fatal("the carried equipment must show whether it is equipped")
	}
	spaceOn(equipment, func() bool { return equipment.SetLeadCell(0, equippedCol) })
	var equipped bool
	screen.Do(func() { equipped = eq.Equipped })
	c.False(equipped, "space on a check cell unchecks it")
	c.Equal(before, dockables(), "and opens no editor")
	screen.KeyPress(unison.KeySpace, mod.None)
	screen.Do(func() { equipped = eq.Equipped })
	c.True(equipped, "and checks it again")
	c.Nil(screen.OpenedURLs())

	nameCol := columnIndexForID(traits, gurps.TraitDescriptionColumn)
	spaceOn(traits, func() bool { return traits.SetLeadCell(0, nameCol) })
	c.Equal(before+1, dockables(), "space on a cell with nothing to press opens the editor")

	// The granted row's name cell holds the mark of a granted row ahead of its notes button.
	var granted *gurps.Trait
	screen.Do(func() {
		for _, one := range sheet.Entity().Traits {
			if gurps.IsGranted(one) {
				granted = one
			}
		}
	})
	if granted == nil {
		t.Fatal("the sheet must show the granted trait")
	}
	noteKey := "N:" + string(granted.ID())
	closed := gurps.IsClosed(noteKey)
	before = dockables()
	spaceOn(traits, func() bool { return traits.SetLeadCell(2, nameCol) })
	var selected []*gurps.Trait
	screen.Do(func() {
		for _, row := range traits.SelectedRows(false) {
			selected = append(selected, row.Data())
		}
	})
	c.Equal(!closed, gurps.IsClosed(noteKey), "space on the name cell of a granted row with notes toggles the notes")
	c.Equal([]*gurps.Trait{granted}, selected, "rather than selecting the item granting it")
	c.Equal(before, dockables(), "and opens no editor")
}
