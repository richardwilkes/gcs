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
	"github.com/richardwilkes/unison"
)

func stBonusFor(entity *gurps.Entity) fxp.Int {
	return entity.AttributeBonusFor(gurps.StrengthID, stlimit.None, nil)
}

// The switch column appears only on a character sheet -- never on a loot sheet, in a template, a library list or an
// editor's table -- and only when something in the list, at any depth, actually has a switch to throw. A loot sheet's
// switches still govern what it displays; they are simply thrown from the item's editor there.
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

// switchableWeightReducer returns a container holding one child, where the container's contained weight reduction only
// applies while its switch is on. This is the strongest case a loot sheet could make for the switch column, since the
// reduction changes what the sheet shows without a character being involved at all.
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

// The loot sheet's equipment list really is built without the switch column, through the same provider the sheet itself
// uses, whether or not anything in it is switchable.
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

// Each of the sheet's page lists gains its switch column, in the expected position, as soon as one of its rows has
// switchable features, and does without it before that.
func TestSheetShowsSwitchColumnWhenNeeded(t *testing.T) {
	for _, tc := range []struct {
		name         string
		add          func(entity *gurps.Entity)
		columns      func(sheet *Sheet) []unison.ColumnInfo
		switchColumn int
		before       int // The column that holds the switch column's slot until the switch column is needed.
		switchAt     int // Where the switch column lands; 1 for carried equipment, whose equipped column stays first.
	}{
		{
			name:         "traits",
			add:          func(entity *gurps.Entity) { entity.Traits = []*gurps.Trait{newSwitchableTrait(entity, "Claws")} },
			columns:      func(sheet *Sheet) []unison.ColumnInfo { return sheet.Traits.Table.Columns },
			switchColumn: gurps.TraitSwitchColumn,
			before:       gurps.TraitDescriptionColumn,
		},
		{
			name:         "skills",
			add:          func(entity *gurps.Entity) { entity.Skills = []*gurps.Skill{newSwitchableSkill(entity, "Brawling")} },
			columns:      func(sheet *Sheet) []unison.ColumnInfo { return sheet.Skills.Table.Columns },
			switchColumn: gurps.SkillSwitchColumn,
			before:       gurps.SkillDescriptionColumn,
		},
		{
			name:         "spells",
			add:          func(entity *gurps.Entity) { entity.Spells = []*gurps.Spell{newSwitchableSpell(entity, "Fireball")} },
			columns:      func(sheet *Sheet) []unison.ColumnInfo { return sheet.Spells.Table.Columns },
			switchColumn: gurps.SpellSwitchColumn,
			before:       gurps.SpellDescriptionForPageColumn,
		},
		{
			name: "carried equipment",
			add: func(entity *gurps.Entity) {
				entity.CarriedEquipment = []*gurps.Equipment{newSwitchableEquipment(entity, "Powered Armor")}
			},
			columns:      func(sheet *Sheet) []unison.ColumnInfo { return sheet.CarriedEquipment.Table.Columns },
			switchColumn: gurps.EquipmentSwitchColumn,
			before:       gurps.EquipmentQuantityColumn,
			switchAt:     1,
		},
		{
			name: "other equipment",
			add: func(entity *gurps.Entity) {
				entity.OtherEquipment = []*gurps.Equipment{newSwitchableEquipment(entity, "Powered Armor")}
			},
			columns:      func(sheet *Sheet) []unison.ColumnInfo { return sheet.OtherEquipment.Table.Columns },
			switchColumn: gurps.EquipmentSwitchColumn,
			before:       gurps.EquipmentQuantityColumn,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			sheet := newTestSheetForTemplate(t)
			columns := tc.columns(sheet)
			c.Equal(-1, switchColumnIndex(columns, tc.switchColumn),
				"without switchable features, there is no switch column")
			c.Equal(tc.before, columns[tc.switchAt].ID,
				"without switchable features, another column holds the switch column's slot")
			first := columns[0].ID

			tc.add(sheet.Entity())
			sheet.Rebuild(true)
			columns = tc.columns(sheet)
			c.Equal(tc.switchAt, switchColumnIndex(columns, tc.switchColumn),
				"the switch column must take its slot once a row has switchable features")
			c.Equal(tc.before, columns[tc.switchAt+1].ID,
				"the switch column must push the column that held its slot along by one")
			if tc.switchAt > 0 {
				c.Equal(first, columns[0].ID, "the column ahead of the switch column must stay first")
			}
		})
	}
}

// newSheetWithSwitchableTrait returns a sheet whose entity holds exactly one trait, carrying a switchable +1 ST bonus
// that starts switched off, along with that trait.
func newSheetWithSwitchableTrait(t *testing.T) (*Sheet, *gurps.Trait) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := newSwitchableTrait(entity, "Claws")
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	return sheet, trait
}

// An option-click cascades a container's new state to the descendants that actually have something to switch, and an
// ordinary click to none of them. Throwing the switch of an item with no switchable features would change nothing the
// user could see, yet the new state would still be written to the file, altering the sheet's contents and bloating the
// undo edit.
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

	toggleFeatureSwitch(sheet.Traits.Table.RootRows()[0], sheet.Traits.Table, true, false)
	c.True(container.SwitchedOn, "an ordinary click must turn the container's switch on")
	c.False(switchable.SwitchedOn, "an ordinary click must leave the switchable child alone")
	c.False(deep.SwitchedOn, "an ordinary click must leave the deeper switchable child alone")
	c.Equal(fxp.FromInteger(2), stBonusFor(entity),
		"only the container's modifier and the always-on bonus may be contributing")
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

// newSheetWithSwitchableReaction returns a sheet whose entity holds one trait carrying a switchable reaction bonus that
// starts switched off, along with that trait. With the switch off there are no reactions to show, so the Reactions list
// is not on the page.
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

// Throwing a switch must rebuild the sheet, not merely mark it as modified: a switchable feature can be a reaction
// bonus, and the Reactions list is on the page only while there is a reaction to show, which is decided only when the
// sheet creates its lists. The rebuild must be the whole update, with no second sync on top of it.
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

// A row whose data has no switch is left alone and reported as such, so a switch cell -- which flips its drawn state
// before asking -- can put itself back.
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
