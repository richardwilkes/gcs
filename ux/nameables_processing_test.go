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
	"maps"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// slicedNameablesPrompt adapts a stand-in for the nameables prompt that takes the sections as parallel slices of their
// titles, substitution maps and visible keys. The maps are the sections' own, so whatever the stand-in fills in reaches
// the rows.
func slicedNameablesPrompt(fn func(titles []string, nameables []map[string]string, visibleKeys [][]string) bool) func(promptOperation, []nameablesSection) bool {
	return func(_ promptOperation, sections []nameablesSection) bool {
		titles := make([]string, len(sections))
		nameables := make([]map[string]string, len(sections))
		visibleKeys := make([][]string, len(sections))
		for i, section := range sections {
			titles[i] = section.Title
			nameables[i] = section.Nameables
			visibleKeys[i] = section.VisibleKeys
		}
		return fn(titles, nameables, visibleKeys)
	}
}

// stubNameablesPrompt substitutes a non-interactive nameables prompt that hands the section titles and the substitution
// maps it was asked to fill to the given responder and reports back whatever the responder returns. The count of
// prompts actually shown is returned, and the real prompt is restored when the test finishes.
func stubNameablesPrompt(t *testing.T, respond func(titles []string, nameables []map[string]string) bool) *int {
	t.Helper()
	shown := 0
	swapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, nameables []map[string]string, _ [][]string) bool {
		shown++
		return respond(titles, nameables)
	}))
	return &shown
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

// altDropNameableModifier performs an alternate drop of dropped onto the rows at rowIndexes of the provider's table,
// answering the single nameables prompt that must follow by filling in the "Material" key of each section with the
// answers in turn. The drop is checked to have shown nothing before prompting. Returns the prompt's section titles and
// the sheet's sync counter.
func altDropNameableModifier[T gurps.Node[T], M gurps.Node[M]](t *testing.T, c check.Checker, sheet *Sheet,
	provider TableProvider[T], rowIndexes []int, dropped M, answers ...string,
) (headings []string, counter *syncCounter) {
	t.Helper()
	counter = installSyncCounter(sheet)
	syncsWhenAsked := -1
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
	c.Equal(0, syncsWhenAsked, "the drop must show nothing before asking")
	return headings, counter
}

// TestAltDropOnTraitAppliesNameablesToTheLiveList verifies that dropping a trait modifier onto a trait prompts for its
// nameable keys before anything is shown, and that the rebuild which follows replaces the traits list when the
// modifier brings switchable features.
func TestAltDropOnTraitAppliesNameablesToTheLiveList(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet, targets := sheetWithNamedTraits(t, c, "Claws")
	stale := sheet.Traits.Table

	// Its switchable feature is what brings the switch column in, and its name needs a substitution.
	dropped := newSwitchableTraitModifier("@Material@ Coating")
	_, counter := altDropNameableModifier(t, c, sheet, sheet.Traits.provider, []int{0}, dropped, "Steel")

	c.Equal(1, len(targets[0].Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
		"the substitution must be applied to the dropped modifier")
	live := sheet.Traits.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the traits table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.TraitSwitchColumn),
		"the dropped modifier's switchable feature must bring the switch column into view")
	c.True(counter.count > 0,
		"the substitutions must be reflected by a rebuild of the list that replaced the one the drop was given")
}

// TestAltDropOnEquipmentAppliesNameablesToTheLiveList verifies the same for dropping equipment modifiers onto an
// equipment row.
func TestAltDropOnEquipmentAppliesNameablesToTheLiveList(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet, targets := sheetWithNamedEquipment(t, c, "Sword")
	stale := sheet.CarriedEquipment.Table

	// Its switchable feature is what brings the switch column in, and its name needs a substitution.
	dropped := newSwitchableEquipmentModifier("@Material@ Coating")
	_, counter := altDropNameableModifier(t, c, sheet, sheet.CarriedEquipment.provider, []int{0}, dropped, "Steel")

	c.Equal(1, len(targets[0].Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Steel Coating", targets[0].Modifiers[0].NameWithReplacements(),
		"the substitution must be applied to the dropped modifier")
	live := sheet.CarriedEquipment.Table
	c.NotEqual(stale, live, "gaining the switch column must replace the carried equipment table")
	c.NotEqual(-1, switchColumnIndex(live.Columns, gurps.EquipmentSwitchColumn),
		"the dropped modifier's switchable feature must bring the switch column into view")
	c.True(counter.count > 0,
		"the substitutions must be reflected by a rebuild of the list that replaced the one the drop was given")
}

// TestAltDropOnSeveralTraitsPromptsForEachCopy verifies that dropping a modifier whose name needs filling in onto
// several selected traits asks about every copy separately, in one prompt. Each target has a copy of its own, so each
// gets its own entry in the prompt and its own answer, letting the same modifier be named differently on each trait it
// was attached to. The copies are otherwise identical, so the entries have to be headed by the trait each belongs to,
// or the user has no way of telling which answer goes where.
func TestAltDropOnSeveralTraitsPromptsForEachCopy(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet, targets := sheetWithNamedTraits(t, c, "Claws", "Fangs")

	dropped := gurps.NewTraitModifier(nil, nil, false)
	dropped.Name = "@Material@ Coating"
	headings, _ := altDropNameableModifier(t, c, sheet, sheet.Traits.provider, []int{0, 1}, dropped,
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
	forbidModifierPrompts(t)
	sheet, targets := sheetWithNamedEquipment(t, c, "Sword", "Shield")

	dropped := gurps.NewEquipmentModifier(nil, nil, false)
	dropped.Name = "@Material@ Coating"
	headings, _ := altDropNameableModifier(t, c, sheet, sheet.CarriedEquipment.provider, []int{0, 1}, dropped,
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

// TestAltDropAsksAboutAKeySharedByATargetsModifiersOnce verifies that a key used by more than one of the modifiers
// dropped onto one trait is asked about once, under the first of them, and that the answer, or its clearing, is
// applied to all of them. The prompt is answered as the real dialog answers it, writing only the keys it showed.
func TestAltDropAsksAboutAKeySharedByATargetsModifiersOnce(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	sheet, targets := sheetWithNamedTraits(t, c, "@Material@ Sword")
	sword := targets[0]
	sword.Replacements = map[string]string{"Material": "Iron"}
	coating := gurps.NewTraitModifier(nil, nil, false)
	coating.Name = "@Material@ Coating"
	inlay := gurps.NewTraitModifier(nil, nil, false)
	inlay.Name = "@Material@ Inlay"
	var headings []string
	var asked []map[string]string
	var visible [][]string
	answer := "Steel"
	swapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, nameables []map[string]string, visibleKeys [][]string) bool {
		headings = titles
		asked = make([]map[string]string, 0, len(nameables))
		visible = visibleKeys
		for i, one := range nameables {
			asked = append(asked, maps.Clone(one))
			for _, k := range visibleKeys[i] {
				if answer == "" {
					delete(one, k)
				} else {
					one[k] = answer
				}
			}
		}
		return true
	}))

	altDrop(sheet.Traits.provider.AltDropSupport(), []int{0}, coating, inlay)
	c.Equal(1, len(headings), "the key must be asked about once, under the first modifier that uses it")
	c.Equal([]map[string]string{{"Material": "Iron"}}, asked, "the field must start out on the trait's value")
	c.Equal([][]string{{"Material"}}, visible)
	c.Equal(map[string]string{"Material": "Steel"}, sword.Replacements, "the answer must be applied for both copies")
	c.Equal([]string{"Steel Coating", "Steel Inlay"}, appliedModifierNames(sword.Modifiers))
	c.Equal("Steel Sword", sword.NameWithReplacements())

	// Dropping both again and clearing the field removes the key for everything on the trait that uses it.
	answer = ""
	headings = nil
	visible = nil
	altDrop(sheet.Traits.provider.AltDropSupport(), []int{0}, coating, inlay)
	c.Equal(1, len(headings), "the key must be asked about once")
	c.Equal([][]string{{"Material"}}, visible, "the second copy must have nothing of its own to ask about")
	c.Equal(0, len(sword.Replacements), "the cleared key must be removed")
	c.Equal("@Material@ Sword", sword.NameWithReplacements())
	c.Equal([]string{"@Material@ Coating", "@Material@ Inlay", "@Material@ Coating", "@Material@ Inlay"},
		appliedModifierNames(sword.Modifiers), "the clearing must be applied for both new copies")
}

// TestProcessNameableGroupsSharesOneAnswerAcrossTheEntriesThatUseIt verifies that, for a group with shared
// replacements, a key asked about under an earlier entry takes that entry's answer, or its clearing, in the later
// entries that use it before they are applied, rather than putting their starting value back.
func TestProcessNameableGroupsSharesOneAnswerAcrossTheEntriesThatUseIt(t *testing.T) {
	newSword := func() (*gurps.Trait, []NameableGroup[*gurps.TraitModifier]) {
		sword := gurps.NewTrait(nil, nil, false)
		sword.Name = "@Material@ Sword"
		sword.Replacements = map[string]string{"Material": "Iron"}
		coating := gurps.NewTraitModifier(nil, nil, false)
		coating.Name = "@Material@ Coating"
		guard := gurps.NewTraitModifier(nil, nil, false)
		guard.Name = "@Color@ @Material@ Guard"
		sword.AddModifiers(coating, guard)
		return sword, []NameableGroup[*gurps.TraitModifier]{{
			Label:              sword.String(),
			Rows:               []*gurps.TraitModifier{coating, guard},
			SharedReplacements: true,
		}}
	}
	// answerVisibleKeys substitutes a prompt that, like the real dialog, writes only the keys it was told to show,
	// clearing any with no answer.
	answerVisibleKeys := func(t *testing.T, answers map[string]string) *[][]string {
		t.Helper()
		var visible [][]string
		swapForTest(t, &promptForNameables, slicedNameablesPrompt(func(_ []string, nameables []map[string]string, visibleKeys [][]string) bool {
			visible = visibleKeys
			for i, keys := range visibleKeys {
				for _, k := range keys {
					if v, ok := answers[k]; ok {
						nameables[i][k] = v
					} else {
						delete(nameables[i], k)
					}
				}
			}
			return true
		}))
		return &visible
	}

	t.Run("answered", func(t *testing.T) {
		c := check.New(t)
		sword, groups := newSword()
		visible := answerVisibleKeys(t, map[string]string{"Material": "Steel", "Color": "Red"})
		c.True(processNameableGroups(promptOperation{}, groups))
		c.Equal([][]string{{"Material"}, {"Color"}}, *visible,
			"the shared key is shown once, under the first modifier, and the second asks only about its own")
		c.Equal(map[string]string{"Material": "Steel", "Color": "Red"}, sword.Replacements,
			"the one answer must be applied for every modifier that uses the key, over the value the trait held")
		c.Equal([]string{"Steel Coating", "Red Steel Guard"}, appliedModifierNames(sword.Modifiers))
		c.Equal("Steel Sword", sword.NameWithReplacements())
	})

	t.Run("cleared", func(t *testing.T) {
		c := check.New(t)
		sword, groups := newSword()
		answerVisibleKeys(t, map[string]string{"Color": "Red"})
		c.True(processNameableGroups(promptOperation{}, groups))
		c.Equal(map[string]string{"Color": "Red"}, sword.Replacements,
			"clearing the key under the first modifier must clear it for the second as well")
		c.Equal([]string{"@Material@ Coating", "Red @Material@ Guard"}, appliedModifierNames(sword.Modifiers))
		c.Equal("@Material@ Sword", sword.NameWithReplacements())
	})
}

// TestAltDropOfAContainerAsksAboutTheKeyItsChildSharesOnce verifies the same for a container and a modifier within it
// that use the same key: the child is asked only about the keys of its own that the container doesn't use, and is left
// out of the prompt entirely when it has none.
func TestAltDropOfAContainerAsksAboutTheKeyItsChildSharesOnce(t *testing.T) {
	c := check.New(t)
	// The container prompt is answered by leaving its contents as they are.
	containerPrompts := stubTraitModifierPrompt(t, func(_ []*gurps.TraitModifier) bool { return false })
	sheet, targets := sheetWithNamedTraits(t, c, "Sword")
	sword := targets[0]
	group := gurps.NewTraitModifier(nil, nil, true)
	group.Name = "@Material@ Fittings"
	pommel := gurps.NewTraitModifier(nil, group, false)
	pommel.Name = "@Material@ Pommel"
	guard := gurps.NewTraitModifier(nil, group, false)
	guard.Name = "@Color@ @Material@ Guard"
	group.Children = []*gurps.TraitModifier{pommel, guard}
	var visible [][]string
	shown := 0
	swapForTest(t, &promptForNameables, slicedNameablesPrompt(func(_ []string, nameables []map[string]string, visibleKeys [][]string) bool {
		shown++
		visible = visibleKeys
		// Only the keys shown are answered, as with the real dialog.
		for i, keys := range visibleKeys {
			for _, k := range keys {
				nameables[i][k] = map[string]string{"Material": "Steel", "Color": "Red"}[k]
			}
		}
		return true
	}))

	altDrop(sheet.Traits.provider.AltDropSupport(), []int{0}, group)
	c.Equal(1, *containerPrompts, "the target must have been asked about the container's contents once")
	c.Equal(1, shown, "the drop must be covered by a single prompt")
	c.Equal([][]string{{"Material"}, {"Color"}}, visible,
		"the container asks about the shared key, the guard only about its own, and the pommel has nothing to ask")
	c.Equal(map[string]string{"Material": "Steel", "Color": "Red"}, sword.Replacements)
	c.Equal(1, len(sword.Modifiers), "the container must be attached")
	c.Equal("Steel Fittings", sword.Modifiers[0].NameWithReplacements())
	c.Equal([]string{"Steel Pommel", "Red Steel Guard"}, appliedModifierNames(sword.Modifiers[0].Children))
}
