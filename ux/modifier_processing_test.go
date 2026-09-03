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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// modifierPrompt records one invocation of a modifier enable/disable prompt.
type modifierPrompt struct {
	title     string
	modifiers []string
}

// captureModifierPrompts substitutes non-interactive modifier prompts that record what they were asked to show,
// returning the accumulator they append to. The real prompts are restored when the test finishes.
func captureModifierPrompts(t *testing.T) *[]modifierPrompt {
	t.Helper()
	origTrait := promptForTraitModifiers
	origEquipment := promptForEquipmentModifiers
	t.Cleanup(func() {
		promptForTraitModifiers = origTrait
		promptForEquipmentModifiers = origEquipment
	})
	var prompts []modifierPrompt
	promptForTraitModifiers = func(title string, modifiers []*gurps.TraitModifier) bool {
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false
	}
	promptForEquipmentModifiers = func(title string, modifiers []*gurps.EquipmentModifier) bool {
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false
	}
	return &prompts
}

// TestProcessModifiersIgnoresModifierRows documents that ProcessModifiers only has something to do for rows that can
// hold modifiers. Handing it the modifiers themselves matches nothing, which is why the alternate drop handlers must
// pass the row the modifiers were dropped onto.
func TestProcessModifiersIgnoresModifierRows(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()
	panel := unison.NewPanel()

	traitMod := gurps.NewTraitModifier(entity, nil, false)
	traitMod.Name = "Trait Modifier"
	ProcessModifiers(panel, []*gurps.TraitModifier{traitMod})
	equipmentMod := gurps.NewEquipmentModifier(entity, nil, false)
	equipmentMod.Name = "Equipment Modifier"
	ProcessModifiers(panel, []*gurps.EquipmentModifier{equipmentMod})
	c.Equal(0, len(*prompts), "modifier rows have no modifiers of their own to prompt for")

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	trait.Modifiers = []*gurps.TraitModifier{traitMod}
	ProcessModifiers(panel, []*gurps.Trait{trait})
	c.Equal([]modifierPrompt{{title: "Trait", modifiers: []string{"Trait Modifier"}}}, *prompts,
		"a trait must be prompted for with its own modifiers")
}

// TestAltDropOnTraitPromptsForTargetModifiers verifies that dropping trait modifiers onto a trait row adds them and
// then prompts for the target trait's modifiers. The prompt used to be handed the dropped modifiers, which
// ProcessModifiers matches nothing for, so it could never appear.
func TestAltDropOnTraitPromptsForTargetModifiers(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	target := gurps.NewTrait(entity, nil, false)
	target.Name = "Target Trait"
	existing := gurps.NewTraitModifier(entity, nil, false)
	existing.Name = "Existing"
	target.Modifiers = []*gurps.TraitModifier{existing}
	entity.Traits = []*gurps.Trait{target}

	provider, ok := NewTraitsProvider(entity, false).(*traitsProvider)
	c.True(ok, "the traits provider must be a *traitsProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	provider.AltDropSupport().Drop([]int{0}, &unison.TableDragData[*Node[*gurps.TraitModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.TraitModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target trait")
	c.Equal([]modifierPrompt{{title: "Target Trait", modifiers: []string{"Existing", "Dropped"}}}, *prompts,
		"the drop must prompt for the target trait's modifiers")
}

// TestAltDropOnEquipmentPromptsForTargetModifiers verifies the same for dropping equipment modifiers onto an equipment
// row.
func TestAltDropOnEquipmentPromptsForTargetModifiers(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	target := gurps.NewEquipment(entity, nil, false)
	target.Name = "Target Equipment"
	existing := gurps.NewEquipmentModifier(entity, nil, false)
	existing.Name = "Existing"
	target.Modifiers = []*gurps.EquipmentModifier{existing}
	entity.CarriedEquipment = []*gurps.Equipment{target}

	provider, ok := NewEquipmentProvider(entity, true, false).(*equipmentProvider)
	c.True(ok, "the equipment provider must be an *equipmentProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{})
	provider.AltDropSupport().Drop([]int{0}, &unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.EquipmentModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target equipment")
	c.Equal([]modifierPrompt{{title: "Target Equipment", modifiers: []string{"Existing", "Dropped"}}}, *prompts,
		"the drop must prompt for the target equipment's modifiers")
}

// TestAltDropOnSeveralTraitsGivesEachItsOwnCopy verifies that dropping a trait modifier onto several selected traits
// attaches a separate copy to each of them and prompts for each in turn. A single shared modifier would tie the traits
// together, so that enabling it on one would enable it on all and renaming it would rename it everywhere.
func TestAltDropOnSeveralTraitsGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	first := gurps.NewTrait(entity, nil, false)
	first.Name = "First Trait"
	second := gurps.NewTrait(entity, nil, false)
	second.Name = "Second Trait"
	entity.Traits = []*gurps.Trait{first, second}

	provider, ok := NewTraitsProvider(entity, false).(*traitsProvider)
	c.True(ok, "the traits provider must be a *traitsProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	provider.AltDropSupport().Drop([]int{0, 1}, &unison.TableDragData[*Node[*gurps.TraitModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.TraitModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first trait")
	c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second trait")
	c.Equal("Dropped", first.Modifiers[0].Name, "the first trait must get the dropped modifier")
	c.Equal("Dropped", second.Modifiers[0].Name, "the second trait must get the dropped modifier")
	c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each trait must get a copy of its own")
	c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
	c.Equal([]modifierPrompt{
		{title: "First Trait", modifiers: []string{"Dropped"}},
		{title: "Second Trait", modifiers: []string{"Dropped"}},
	}, *prompts, "each trait must be prompted for its own modifiers")
}

// TestAltDropOnSeveralEquipmentItemsGivesEachItsOwnCopy verifies the same for dropping equipment modifiers onto several
// selected equipment items.
func TestAltDropOnSeveralEquipmentItemsGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	first := gurps.NewEquipment(entity, nil, false)
	first.Name = "First Item"
	second := gurps.NewEquipment(entity, nil, false)
	second.Name = "Second Item"
	entity.CarriedEquipment = []*gurps.Equipment{first, second}

	provider, ok := NewEquipmentProvider(entity, true, false).(*equipmentProvider)
	c.True(ok, "the equipment provider must be an *equipmentProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{})
	provider.AltDropSupport().Drop([]int{0, 1}, &unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.EquipmentModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first item")
	c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second item")
	c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each item must get a copy of its own")
	c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
	c.Equal([]modifierPrompt{
		{title: "First Item", modifiers: []string{"Dropped"}},
		{title: "Second Item", modifiers: []string{"Dropped"}},
	}, *prompts, "each item must be prompted for its own modifiers")
}

// TestAltDropOnAContainerAndItsChildPromptsTheChildOnce verifies that a selection holding both a container and one of
// its own descendants doesn't ask about that descendant twice. ProcessModifiers walks everything below each row it is
// handed, so the child is already covered by its container being in the list and must be dropped from it, the same
// reduction the selection-driven callers get from SelectedRows(true).
func TestAltDropOnAContainerAndItsChildPromptsTheChildOnce(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	container := gurps.NewTrait(entity, nil, true)
	container.Name = "Container Trait"
	child := gurps.NewTrait(entity, container, false)
	child.Name = "Child Trait"
	container.Children = []*gurps.Trait{child}
	entity.Traits = []*gurps.Trait{container}

	provider, ok := NewTraitsProvider(entity, false).(*traitsProvider)
	c.True(ok, "the traits provider must be a *traitsProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())
	c.Equal(1, table.LastRowIndex(), "the container's child must be disclosed")

	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	provider.AltDropSupport().Drop([]int{0, 1}, &unison.TableDragData[*Node[*gurps.TraitModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.TraitModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(1, len(container.Modifiers), "the container must still receive the dropped modifier")
	c.Equal(1, len(child.Modifiers), "the child must still receive its own copy of the dropped modifier")
	c.Equal([]modifierPrompt{
		{title: "Container Trait", modifiers: []string{"Dropped"}},
		{title: "Child Trait", modifiers: []string{"Dropped"}},
	}, *prompts, "the child must be prompted for once, by way of its container")
}

// TestAltDropOnAMissingTraitRowIsANoOp verifies that a row index which resolves to nothing is quietly left out of a
// drop, and that a drop none of whose indexes resolve does nothing at all: nothing is attached and no prompt goes up.
// The indexes come from the table the drop landed in, so none should ever miss, but one that does must not take the
// handler down.
func TestAltDropOnAMissingTraitRowIsANoOp(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	entity.Traits = []*gurps.Trait{trait}
	provider, ok := NewTraitsProvider(entity, false).(*traitsProvider)
	c.True(ok, "the traits provider must be a *traitsProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	data := &unison.TableDragData[*Node[*gurps.TraitModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.TraitModifier]{NewNode(modTable, nil, dropped, false)},
	}

	provider.AltDropSupport().Drop([]int{99}, data)
	c.Equal(0, len(trait.Modifiers), "a drop with no resolvable target must attach nothing")
	c.Equal(0, len(*prompts), "a drop with no resolvable target must not prompt")

	provider.AltDropSupport().Drop([]int{99, 0}, data)
	c.Equal(1, len(trait.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
	c.Equal([]modifierPrompt{{title: "Trait", modifiers: []string{"Dropped"}}}, *prompts,
		"only the target that resolved may be prompted for")
}

// TestAltDropOnAMissingEquipmentRowIsANoOp verifies the same for the equipment drop handler.
func TestAltDropOnAMissingEquipmentRowIsANoOp(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()
	item := gurps.NewEquipment(entity, nil, false)
	item.Name = "Item"
	entity.CarriedEquipment = []*gurps.Equipment{item}
	provider, ok := NewEquipmentProvider(entity, true, false).(*equipmentProvider)
	c.True(ok, "the equipment provider must be an *equipmentProvider")
	table := unison.NewTable(provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{})
	data := &unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.EquipmentModifier]{NewNode(modTable, nil, dropped, false)},
	}

	provider.AltDropSupport().Drop([]int{99}, data)
	c.Equal(0, len(item.Modifiers), "a drop with no resolvable target must attach nothing")
	c.Equal(0, len(*prompts), "a drop with no resolvable target must not prompt")

	provider.AltDropSupport().Drop([]int{99, 0}, data)
	c.Equal(1, len(item.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
	c.Equal([]modifierPrompt{{title: "Item", modifiers: []string{"Dropped"}}}, *prompts,
		"only the target that resolved may be prompted for")
}

// TestProcessModifiersRebuildsThroughAReplacedTable verifies that answering a modifier prompt still rebuilds the sheet
// when the table ProcessModifiers was handed has since been replaced. The alternate drop path rebuilds before it
// prompts, and an earlier prompt in the same pass rebuilds too, and either rebuild can replace the table when the
// answer adds or takes away the switch column, so the lookup for the owner to rebuild has to go through the live table
// rather than upward from the orphan.
func TestProcessModifiersRebuildsThroughAReplacedTable(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	first := gurps.NewTrait(entity, nil, false)
	first.Name = "Claws"
	first.Modifiers = []*gurps.TraitModifier{newSwitchableTraitModifier("Sharp")}
	second := gurps.NewTrait(entity, nil, false)
	second.Name = "Fangs"
	plain := gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Venomous"
	plain.Disabled = true
	plain.Features = gurps.Features{gurps.NewAttributeBonus(gurps.StrengthID)} // Not switchable: applies once enabled.
	second.Modifiers = []*gurps.TraitModifier{plain}
	entity.Traits = []*gurps.Trait{first, second}
	sheet.Rebuild(true)
	stale := sheet.Traits.Table
	c.Equal(-1, switchColumnIndex(stale.Columns, gurps.TraitSwitchColumn),
		"with every modifier disabled, the traits list must start out without the switch column")
	c.Equal(fxp.Int(0), stBonusFor(entity), "nothing contributes to ST while every modifier is disabled")

	// Enabling the first trait's modifier gives it switchable features, so the rebuild the prompt asks for brings the
	// switch column in and replaces the table. Enabling the second trait's modifier changes nothing about the columns,
	// but the rebuild its answer asks for is what recalculates the entity with the modifier's bonus in play -- and it
	// is asked for through a table that the first answer orphaned.
	shown := stubTraitModifierPrompt(t, enableAllModifiers)
	entity.ModifiedOn = jio.Time{}
	ProcessModifiers(stale, []*gurps.Trait{first, second})
	c.Equal(2, *shown, "both traits must have been prompted for")
	c.NotEqual(jio.Time{}, entity.ModifiedOn,
		"an answer that changes a modifier is an edit, so it must bump the modification timestamp")
	c.NotEqual(stale, sheet.Traits.Table, "enabling the first modifier must have replaced the traits table")
	c.NotEqual(-1, switchColumnIndex(sheet.Traits.Table.Columns, gurps.TraitSwitchColumn),
		"the live traits table must have gained the switch column")
	c.False(plain.Disabled, "the second prompt's answer must have been applied to the model")
	c.Equal(fxp.One, stBonusFor(entity),
		"the rebuild after the second answer must have recalculated the entity, bringing the enabled bonus into play")
	c.True(columnsMatchProvider(sheet.Traits.Table), "the live table's columns must match its provider")
}
