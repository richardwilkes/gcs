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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// stubNameablesPrompt substitutes a non-interactive nameables prompt that hands the substitution maps it was asked to
// fill to the given responder and reports back whatever the responder returns, letting a test drive the rebuild that
// answering the prompt triggers. The count of prompts actually shown is returned, and the real prompt is restored when
// the test finishes.
func stubNameablesPrompt(t *testing.T, respond func(nameables []map[string]string) bool) *int {
	t.Helper()
	original := promptForNameables
	t.Cleanup(func() { promptForNameables = original })
	shown := 0
	promptForNameables = func(_ []string, nameables []map[string]string, visibleKeys [][]string) bool {
		shown++
		return respond(nameables)
	}
	return &shown
}

// TestProcessNameablesRebuildsThroughAReplacedTable verifies that answering the nameables prompt still rebuilds the
// sheet when the table ProcessNameables was handed has since been replaced. The substitutions are applied to the model
// no matter what, so a rebuild that never happens leaves the list the user is looking at showing the raw keys and the
// values derived from them until some unrelated edit comes along.
func TestProcessNameablesRebuildsThroughAReplacedTable(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "@Adjective@ Claws"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	stale := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")

	// Stand in for the rebuild the alternate drop path performs before it prompts: the modifier that was dropped onto
	// the trait carries a switchable feature, so the trait now has switchable features, which brings the switch column
	// into view, and a list can only gain a column by being built anew -- leaving the table captured above orphaned.
	modifier := newSwitchableTraitModifier("Retractable")
	modifier.Disabled = false
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits.Table, "gaining the switch column must replace the traits table")
	c.Nil(unison.Ancestor[Rebuildable](stale), "an orphaned table must have no rebuildable above it")

	counter := installSyncCounter(sheet)
	shown := stubNameablesPrompt(t, func(nameables []map[string]string) bool {
		for _, one := range nameables {
			one["Adjective"] = "Sharp"
		}
		return true
	})
	ProcessNameables(stale, []*gurps.Trait{trait})

	c.Equal(1, *shown, "the trait's nameable key must be prompted for")
	c.Equal("Sharp Claws", trait.NameWithReplacements(), "the substitution must be applied to the trait")
	c.Equal(1, counter.count, "the substitutions must be reflected by a rebuild of the live list's owner")
}

// TestProcessNameablesRebuildsThroughAReplacedTableForModifierRows verifies the same for the alternate drop path's
// call, where the rows handed over are the dropped modifiers rather than the rows of the table that came with them.
// The lookup for the live table therefore can't be driven by the row type, since it doesn't match the table's.
func TestProcessNameablesRebuildsThroughAReplacedTableForModifierRows(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Claws"
	entity.Traits = []*gurps.Trait{trait}
	sheet.Rebuild(true)
	stale := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")

	// Stand in for the rebuild the alternate drop path performs before it prompts: the modifier that was dropped onto
	// the trait is enabled and carries a switchable feature, so the trait now has switchable features, which brings the
	// switch column into view, and a list can only gain a column by being built anew -- leaving the table captured
	// above orphaned. The modifier's own name is what needs a substitution, so it is the modifier, not the trait, that
	// the prompt is asked for.
	modifier := newSwitchableTraitModifier("@Material@ Coating")
	modifier.Disabled = false
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits.Table, "gaining the switch column must replace the traits table")
	c.Nil(unison.Ancestor[Rebuildable](stale), "an orphaned table must have no rebuildable above it")

	counter := installSyncCounter(sheet)
	shown := stubNameablesPrompt(t, func(nameables []map[string]string) bool {
		for _, one := range nameables {
			one["Material"] = "Steel"
		}
		return true
	})
	ProcessNameables(stale, []*gurps.TraitModifier{modifier})

	c.Equal(1, *shown, "the modifier's nameable key must be prompted for")
	c.Equal("Steel Coating", modifier.NameWithReplacements(), "the substitution must be applied to the modifier")
	c.Equal(1, counter.count, "the substitutions must be reflected by a rebuild of the live list's owner")
}

// TestAltDropOnTraitAppliesNameablesToTheLiveList verifies the path the fix above was made for: dropping a trait
// modifier from a library onto a trait rebuilds the sheet before it prompts, and that rebuild replaces the traits list
// when the dropped modifier gives the trait switchable features. The nameables prompt that follows is handed the
// provider's own table, which is the orphan the rebuild left behind.
func TestAltDropOnTraitAppliesNameablesToTheLiveList(t *testing.T) {
	c := check.New(t)
	captureModifierPrompts(t) // The modifier prompt comes first and must not try to put up a real dialog.
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	target := gurps.NewTrait(entity, nil, false)
	target.Name = "Claws"
	entity.Traits = []*gurps.Trait{target}
	sheet.Rebuild(true)
	provider := sheet.Traits.provider
	stale := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")

	// The dropped modifier is enabled and carries a switchable feature, so adding it is what makes the switch column
	// necessary, and its name needs a substitution.
	dropped := newSwitchableTraitModifier("@Material@ Coating")
	dropped.Disabled = false
	modTable := unison.NewTable[*Node[*gurps.TraitModifier]](&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	counter := installSyncCounter(sheet)
	syncsWhenAsked := -1
	shown := stubNameablesPrompt(t, func(nameables []map[string]string) bool {
		syncsWhenAsked = counter.count
		for _, one := range nameables {
			one["Material"] = "Steel"
		}
		return true
	})

	provider.AltDropSupport().Drop(0, &unison.TableDragData[*Node[*gurps.TraitModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.TraitModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(1, *shown, "the drop must prompt for the dropped modifier's nameable keys")
	c.Equal(1, len(target.Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Steel Coating", target.Modifiers[0].NameWithReplacements(),
		"the substitution must be applied to the dropped modifier")
	live := sheet.Traits.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the traits table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"the dropped modifier's switchable feature must bring the switch column into view")
	c.True(counter.count > syncsWhenAsked,
		"the substitutions must be reflected by a rebuild of the list that replaced the one the drop was given")
}

// TestAltDropOnEquipmentAppliesNameablesToTheLiveList verifies the same for dropping equipment modifiers onto an
// equipment row.
func TestAltDropOnEquipmentAppliesNameablesToTheLiveList(t *testing.T) {
	c := check.New(t)
	captureModifierPrompts(t) // The modifier prompt comes first and must not try to put up a real dialog.
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	target := gurps.NewEquipment(entity, nil, false)
	target.Name = "Sword"
	entity.CarriedEquipment = []*gurps.Equipment{target}
	sheet.Rebuild(true)
	provider := sheet.CarriedEquipment.provider
	stale := sheet.CarriedEquipment.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.EquipmentSwitchColumn),
		"the carried equipment list must start out without the switch column")

	dropped := gurps.NewEquipmentModifier(nil, nil, false)
	dropped.Name = "@Material@ Coating"
	dropped.Features = gurps.Features{switchableSTBonus(nil)}
	modTable := unison.NewTable[*Node[*gurps.EquipmentModifier]](
		&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{},
	)
	counter := installSyncCounter(sheet)
	syncsWhenAsked := -1
	shown := stubNameablesPrompt(t, func(nameables []map[string]string) bool {
		syncsWhenAsked = counter.count
		for _, one := range nameables {
			one["Material"] = "Steel"
		}
		return true
	})

	provider.AltDropSupport().Drop(0, &unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.EquipmentModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(1, *shown, "the drop must prompt for the dropped modifier's nameable keys")
	c.Equal(1, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Steel Coating", target.Modifiers[0].NameWithReplacements(),
		"the substitution must be applied to the dropped modifier")
	live := sheet.CarriedEquipment.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the carried equipment table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.EquipmentSwitchColumn),
		"the dropped modifier's switchable feature must bring the switch column into view")
	c.True(counter.count > syncsWhenAsked,
		"the substitutions must be reflected by a rebuild of the list that replaced the one the drop was given")
}
