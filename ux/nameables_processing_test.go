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

// stubNameablesPrompt substitutes a non-interactive nameables prompt that hands the section titles and the substitution
// maps it was asked to fill to the given responder and reports back whatever the responder returns, letting a test
// drive the rebuild that answering the prompt triggers. The count of prompts actually shown is returned, and the real
// prompt is restored when the test finishes.
func stubNameablesPrompt(t *testing.T, respond func(titles []string, nameables []map[string]string) bool) *int {
	t.Helper()
	shown := 0
	swapForTest(t, &promptForNameables, func(titles []string, nameables []map[string]string, _ [][]string) bool {
		shown++
		return respond(titles, nameables)
	})
	return &shown
}

// fillNameables returns a responder for stubNameablesPrompt that answers every section it is shown by filling in the
// given key with the given value.
func fillNameables(key, value string) func(titles []string, nameables []map[string]string) bool {
	return func(_ []string, nameables []map[string]string) bool {
		for _, one := range nameables {
			one[key] = value
		}
		return true
	}
}

// sheetWithNamedTraits returns a sheet whose traits list holds one non-container trait per name, along with those
// traits in the order a drop will visit them, which is the order they appear in the list rather than the order they
// were handed over in. The list is verified to be showing every one of them and to be without the switch column, so
// that a drop which brings that column in can be seen to have replaced the table.
func sheetWithNamedTraits(t *testing.T, c check.Checker, names ...string) (*Sheet, []*gurps.Trait) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	namedTraits(sheet.Entity(), names...)
	sheet.Rebuild(true)
	table := sheet.Traits.Table
	c.Equal(len(names)-1, table.LastRowIndex(), "the sheet must show every trait")
	c.Equal(-1, switchColumnIndex(table.Columns, gurps.TraitSwitchColumn),
		"the traits list must start out without the switch column")
	return sheet, rowData(table)
}

// sheetWithNamedEquipment returns a sheet whose carried equipment list holds one non-container item per name, along
// with those items in the order a drop will visit them, which is the order they appear in the list rather than the
// order they were handed over in. The list is verified to be showing every one of them and to be without the switch
// column, so that a drop which brings that column in can be seen to have replaced the table.
func sheetWithNamedEquipment(t *testing.T, c check.Checker, names ...string) (*Sheet, []*gurps.Equipment) {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	namedEquipment(sheet.Entity(), names...)
	sheet.Rebuild(true)
	table := sheet.CarriedEquipment.Table
	c.Equal(len(names)-1, table.LastRowIndex(), "the sheet must show every item")
	c.Equal(-1, switchColumnIndex(table.Columns, gurps.EquipmentSwitchColumn),
		"the carried equipment list must start out without the switch column")
	return sheet, rowData(table)
}

// rowData returns the data of each of the table's rows, in the order the rows appear in it.
func rowData[T gurps.Node[T]](table *unison.Table[*Node[T]]) []T {
	data := make([]T, table.LastRowIndex()+1)
	for i := range data {
		data[i] = table.RowFromIndex(i).Data()
	}
	return data
}

// sheetWithStaleTraitsTable returns a sheet holding a single trait named traitName whose traits table has since been
// replaced, along with that trait, the enabled switchable modifier named modifierName that was attached to it and the
// table the replacement left orphaned. Attaching that modifier stands in for the rebuild the alternate drop path
// performs before it prompts: the modifier carries a switchable feature, so the trait now has switchable features,
// which brings the switch column into view, and a list can only gain a column by being built anew.
func sheetWithStaleTraitsTable(t *testing.T, c check.Checker, traitName, modifierName string) (sheet *Sheet,
	trait *gurps.Trait, modifier *gurps.TraitModifier, stale *unison.Table[*Node[*gurps.Trait]],
) {
	t.Helper()
	sheet, traits := sheetWithNamedTraits(t, c, traitName)
	trait = traits[0]
	stale = sheet.Traits.Table
	modifier = newSwitchableTraitModifier(modifierName)
	modifier.Disabled = false
	trait.Modifiers = []*gurps.TraitModifier{modifier}
	sheet.Rebuild(true)
	c.NotEqual(stale, sheet.Traits.Table, "gaining the switch column must replace the traits table")
	c.Nil(stale.Ancestor[Rebuildable](), "an orphaned table must have no rebuildable above it")
	return sheet, trait, modifier, stale
}

// altDropNameableModifier performs an alternate drop of dropped onto the rows at rowIndexes of the provider's table,
// answering the nameables prompt that follows by filling in the "Material" key of each section with the answers in
// turn, so that the copy of the dropped modifier each target received can be told from the others by the name it ends
// up with. The sheet is given a sync counter first and the counter's tally at the moment the prompt went up is
// reported along with it, so that the rebuild the answers ask for can be told from the ones the drop itself performed.
// The whole drop must be covered by a single prompt; that prompt's section titles are returned.
func altDropNameableModifier[T gurps.Node[T], M gurps.Node[M]](t *testing.T, c check.Checker, sheet *Sheet,
	provider TableProvider[T], rowIndexes []int, dropped M, answers ...string,
) (headings []string, syncsWhenAsked int, counter *syncCounter) {
	t.Helper()
	counter = installSyncCounter(sheet)
	syncsWhenAsked = -1
	shown := stubNameablesPrompt(t, func(titles []string, nameables []map[string]string) bool {
		headings = titles
		syncsWhenAsked = counter.count
		for i, one := range nameables {
			one["Material"] = answers[i%len(answers)]
		}
		return true
	})
	altDrop(provider.AltDropSupport(), rowIndexes, dropped)
	c.Equal(1, *shown, "the whole drop must be covered by a single prompt")
	return headings, syncsWhenAsked, counter
}

// TestProcessNameablesRebuildsThroughAReplacedTable verifies that answering the nameables prompt still rebuilds the
// sheet when the table ProcessNameables was handed has since been replaced. The substitutions are applied to the model
// no matter what, so a rebuild that never happens leaves the list the user is looking at showing the raw keys and the
// values derived from them until some unrelated edit comes along.
func TestProcessNameablesRebuildsThroughAReplacedTable(t *testing.T) {
	c := check.New(t)
	sheet, trait, _, stale := sheetWithStaleTraitsTable(t, c, "@Adjective@ Claws", "Retractable")

	counter := installSyncCounter(sheet)
	shown := stubNameablesPrompt(t, fillNameables("Adjective", "Sharp"))
	ProcessNameables(stale, []*gurps.Trait{trait})

	c.Equal(1, *shown, "the trait's nameable key must be prompted for")
	c.Equal("Sharp Claws", trait.NameWithReplacements(), "the substitution must be applied to the trait")
	c.Equal(1, counter.count, "the substitutions must be reflected by a rebuild of the live list's owner")
}

// TestProcessNameablesRebuildsThroughAReplacedTableForModifierRows verifies the same for the alternate drop path's
// call, where the rows handed over are the dropped modifiers rather than the rows of the table that came with them.
// The lookup for the live table therefore can't be driven by the row type, since it doesn't match the table's. The
// modifier's own name is what needs a substitution here, so it is the modifier, not the trait, that the prompt is
// asked for.
func TestProcessNameablesRebuildsThroughAReplacedTableForModifierRows(t *testing.T) {
	c := check.New(t)
	sheet, _, modifier, stale := sheetWithStaleTraitsTable(t, c, "Claws", "@Material@ Coating")

	counter := installSyncCounter(sheet)
	shown := stubNameablesPrompt(t, fillNameables("Material", "Steel"))
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
	sheet, targets := sheetWithNamedTraits(t, c, "Claws")
	stale := sheet.Traits.Table

	// The dropped modifier is enabled and carries a switchable feature, so adding it is what makes the switch column
	// necessary, and its name needs a substitution.
	dropped := newSwitchableTraitModifier("@Material@ Coating")
	dropped.Disabled = false
	_, syncsWhenAsked, counter := altDropNameableModifier(t, c, sheet, sheet.Traits.provider, []int{0}, dropped,
		"Steel")

	c.Equal(1, len(targets[0].Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
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
	sheet, targets := sheetWithNamedEquipment(t, c, "Sword")
	stale := sheet.CarriedEquipment.Table

	// The dropped modifier is enabled and carries a switchable feature, so adding it is what makes the switch column
	// necessary, and its name needs a substitution.
	dropped := newSwitchableEquipmentModifier("@Material@ Coating")
	dropped.Disabled = false
	_, syncsWhenAsked, counter := altDropNameableModifier(t, c, sheet, sheet.CarriedEquipment.provider, []int{0},
		dropped, "Steel")

	c.Equal(1, len(targets[0].Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
		"the substitution must be applied to the dropped modifier")
	live := sheet.CarriedEquipment.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the carried equipment table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.EquipmentSwitchColumn),
		"the dropped modifier's switchable feature must bring the switch column into view")
	c.True(counter.count > syncsWhenAsked,
		"the substitutions must be reflected by a rebuild of the list that replaced the one the drop was given")
}

// TestAltDropOnSeveralTraitsPromptsForEachCopy verifies that dropping a modifier whose name needs filling in onto
// several selected traits asks about every copy separately, in one prompt. Each target has a copy of its own, so each
// gets its own entry in the prompt and its own answer, letting the same modifier be named differently on each trait it
// was attached to. The copies are otherwise identical, so the entries have to be headed by the trait each belongs to,
// or the user has no way of telling which answer goes where.
func TestAltDropOnSeveralTraitsPromptsForEachCopy(t *testing.T) {
	c := check.New(t)
	captureModifierPrompts(t) // The modifier prompt comes first and must not try to put up a real dialog.
	sheet, targets := sheetWithNamedTraits(t, c, "Claws", "Fangs")

	dropped := gurps.NewTraitModifier(nil, nil, false)
	dropped.Name = "@Material@ Coating"
	headings, _, _ := altDropNameableModifier(t, c, sheet, sheet.Traits.provider, []int{0, 1}, dropped,
		"Steel", "Silver")

	c.Equal([]string{"Claws: @Material@ Coating", "Fangs: @Material@ Coating"}, headings,
		"the prompt must hold an entry for each copy of the dropped modifier, headed by the trait it belongs to")
	c.Equal(1, len(targets[0].Modifiers), "the first trait must receive a copy of the dropped modifier")
	c.Equal(1, len(targets[1].Modifiers), "the second trait must receive a copy of the dropped modifier")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
		"the first copy must get the first answer")
	c.Equal("Silver Coating", targets[1].Modifiers[0].NameWithReplacements(),
		"the second copy must get its own answer rather than the first one's")
}

// TestAltDropOnSeveralEquipmentItemsPromptsForEachCopy verifies the same for dropping an equipment modifier onto
// several selected equipment items.
func TestAltDropOnSeveralEquipmentItemsPromptsForEachCopy(t *testing.T) {
	c := check.New(t)
	captureModifierPrompts(t) // The modifier prompt comes first and must not try to put up a real dialog.
	sheet, targets := sheetWithNamedEquipment(t, c, "Sword", "Shield")

	dropped := gurps.NewEquipmentModifier(nil, nil, false)
	dropped.Name = "@Material@ Coating"
	headings, _, _ := altDropNameableModifier(t, c, sheet, sheet.CarriedEquipment.provider, []int{0, 1}, dropped,
		"Steel", "Silver")

	c.Equal([]string{"Sword: @Material@ Coating", "Shield: @Material@ Coating"}, headings,
		"the prompt must hold an entry for each copy of the dropped modifier, headed by the item it belongs to")
	c.Equal(1, len(targets[0].Modifiers), "the first item must receive a copy of the dropped modifier")
	c.Equal(1, len(targets[1].Modifiers), "the second item must receive a copy of the dropped modifier")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
		"the first copy must get the first answer")
	c.Equal("Silver Coating", targets[1].Modifiers[0].NameWithReplacements(),
		"the second copy must get its own answer rather than the first one's")
}
