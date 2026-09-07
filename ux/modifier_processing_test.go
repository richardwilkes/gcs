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
	var prompts []modifierPrompt
	swapForTest(t, &promptForTraitModifiers, func(title string, modifiers []*gurps.TraitModifier) bool {
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false
	})
	swapForTest(t, &promptForEquipmentModifiers, func(title string, modifiers []*gurps.EquipmentModifier) bool {
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false
	})
	return &prompts
}

// namedTraits fills the entity's traits list with one non-container trait per name and returns them.
func namedTraits(entity *gurps.Entity, names ...string) []*gurps.Trait {
	traits := make([]*gurps.Trait, len(names))
	for i, name := range names {
		traits[i] = gurps.NewTrait(entity, nil, false)
		traits[i].Name = name
	}
	entity.Traits = traits
	return traits
}

// namedEquipment fills the entity's carried equipment list with one non-container item per name and returns them.
func namedEquipment(entity *gurps.Entity, names ...string) []*gurps.Equipment {
	equipment := make([]*gurps.Equipment, len(names))
	for i, name := range names {
		equipment[i] = gurps.NewEquipment(entity, nil, false)
		equipment[i].Name = name
	}
	entity.CarriedEquipment = equipment
	return equipment
}

// entityWithNamedTraits returns a new entity whose traits list holds one non-container trait per name, along with those
// traits.
func entityWithNamedTraits(names ...string) (*gurps.Entity, []*gurps.Trait) {
	entity := gurps.NewEntity()
	return entity, namedTraits(entity, names...)
}

// entityWithNamedEquipment returns a new entity whose carried equipment list holds one non-container item per name,
// along with those items.
func entityWithNamedEquipment(names ...string) (*gurps.Entity, []*gurps.Equipment) {
	entity := gurps.NewEntity()
	return entity, namedEquipment(entity, names...)
}

// tabledProvider gives the provider the table and root rows that NewNodeTable would, then hands the provider back, so
// that a provider a test only needs in order to drive a drop can be had in a single expression.
func tabledProvider[T gurps.Node[T]](provider TableProvider[T]) TableProvider[T] {
	newProviderTable(provider)
	return provider
}

// checkAltDropPrompts performs an alternate drop of dropped onto the rows at rowIndexes of the provider's table and
// verifies that the modifier prompts that go up are exactly the wanted ones, in order. The prompts are captured for the
// remainder of the test, so nothing tries to put up a real dialog.
func checkAltDropPrompts[T gurps.Node[T], M gurps.Node[M]](t *testing.T, c check.Checker, provider TableProvider[T],
	rowIndexes []int, dropped M, wantPrompts []modifierPrompt,
) {
	t.Helper()
	prompts := captureModifierPrompts(t)
	altDrop(provider.AltDropSupport(), rowIndexes, dropped)
	c.Equal(wantPrompts, *prompts, "the drop must prompt for the modifiers of each target it resolved, and no others")
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
	entity, traits := entityWithNamedTraits("Target Trait")
	target := traits[0]
	existing := gurps.NewTraitModifier(entity, nil, false)
	existing.Name = "Existing"
	target.Modifiers = []*gurps.TraitModifier{existing}
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, tabledProvider(NewTraitsProvider(entity, false)), []int{0}, dropped,
		[]modifierPrompt{{title: "Target Trait", modifiers: []string{"Existing", "Dropped"}}})

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target trait")
}

// TestAltDropOnEquipmentPromptsForTargetModifiers verifies the same for dropping equipment modifiers onto an equipment
// row.
func TestAltDropOnEquipmentPromptsForTargetModifiers(t *testing.T) {
	c := check.New(t)
	entity, equipment := entityWithNamedEquipment("Target Equipment")
	target := equipment[0]
	existing := gurps.NewEquipmentModifier(entity, nil, false)
	existing.Name = "Existing"
	target.Modifiers = []*gurps.EquipmentModifier{existing}
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, tabledProvider(NewEquipmentProvider(entity, true, false)), []int{0}, dropped,
		[]modifierPrompt{{title: "Target Equipment", modifiers: []string{"Existing", "Dropped"}}})

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target equipment")
}

// TestAltDropOnSeveralTraitsGivesEachItsOwnCopy verifies that dropping a trait modifier onto several selected traits
// attaches a separate copy to each of them and prompts for each in turn. A single shared modifier would tie the traits
// together, so that enabling it on one would enable it on all and renaming it would rename it everywhere.
func TestAltDropOnSeveralTraitsGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	entity, traits := entityWithNamedTraits("First Trait", "Second Trait")
	first, second := traits[0], traits[1]
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, tabledProvider(NewTraitsProvider(entity, false)), []int{0, 1}, dropped,
		[]modifierPrompt{
			{title: "First Trait", modifiers: []string{"Dropped"}},
			{title: "Second Trait", modifiers: []string{"Dropped"}},
		})

	c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first trait")
	c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second trait")
	c.Equal("Dropped", first.Modifiers[0].Name, "the first trait must get the dropped modifier")
	c.Equal("Dropped", second.Modifiers[0].Name, "the second trait must get the dropped modifier")
	c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each trait must get a copy of its own")
	c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
}

// TestAltDropOnSeveralEquipmentItemsGivesEachItsOwnCopy verifies the same for dropping equipment modifiers onto several
// selected equipment items.
func TestAltDropOnSeveralEquipmentItemsGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	entity, equipment := entityWithNamedEquipment("First Item", "Second Item")
	first, second := equipment[0], equipment[1]
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, tabledProvider(NewEquipmentProvider(entity, true, false)), []int{0, 1}, dropped,
		[]modifierPrompt{
			{title: "First Item", modifiers: []string{"Dropped"}},
			{title: "Second Item", modifiers: []string{"Dropped"}},
		})

	c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first item")
	c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second item")
	c.Equal("Dropped", first.Modifiers[0].Name, "the first item must get the dropped modifier")
	c.Equal("Dropped", second.Modifiers[0].Name, "the second item must get the dropped modifier")
	c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each item must get a copy of its own")
	c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
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

	provider := NewTraitsProvider(entity, false)
	table := newProviderTable(provider)
	c.Equal(1, table.LastRowIndex(), "the container's child must be disclosed")

	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	altDrop(provider.AltDropSupport(), []int{0, 1}, dropped)

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
	entity, traits := entityWithNamedTraits("Trait")
	trait := traits[0]
	provider := tabledProvider(NewTraitsProvider(entity, false))
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, provider, []int{99}, dropped, nil)
	c.Equal(0, len(trait.Modifiers), "a drop with no resolvable target must attach nothing")

	checkAltDropPrompts(t, c, provider, []int{99, 0}, dropped,
		[]modifierPrompt{{title: "Trait", modifiers: []string{"Dropped"}}})
	c.Equal(1, len(trait.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
}

// TestAltDropOnAMissingEquipmentRowIsANoOp verifies the same for the equipment drop handler.
func TestAltDropOnAMissingEquipmentRowIsANoOp(t *testing.T) {
	c := check.New(t)
	entity, equipment := entityWithNamedEquipment("Item")
	item := equipment[0]
	provider := tabledProvider(NewEquipmentProvider(entity, true, false))
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"

	checkAltDropPrompts(t, c, provider, []int{99}, dropped, nil)
	c.Equal(0, len(item.Modifiers), "a drop with no resolvable target must attach nothing")

	checkAltDropPrompts(t, c, provider, []int{99, 0}, dropped,
		[]modifierPrompt{{title: "Item", modifiers: []string{"Dropped"}}})
	c.Equal(1, len(item.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
}

// TestAltDropOfTheWrongKindOfModifierIsIgnored verifies that each provider's alternate drop handler only acts on drag
// data holding its own kind of modifier: equipment modifiers handed to the traits handler, or trait modifiers handed
// to the equipment handler, attach nothing and put up no prompt. The handlers share one implementation, so this is
// what pins each provider to the modifier type its rows actually carry.
func TestAltDropOfTheWrongKindOfModifierIsIgnored(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()
	trait := namedTraits(entity, "Trait")[0]
	item := namedEquipment(entity, "Item")[0]

	traitsProv := tabledProvider(NewTraitsProvider(entity, false))
	equipmentProv := tabledProvider(NewEquipmentProvider(entity, true, false))

	traitMod := gurps.NewTraitModifier(entity, nil, false)
	traitMod.Name = "Trait Modifier"
	traitModData := newDragData(traitMod)
	equipmentMod := gurps.NewEquipmentModifier(entity, nil, false)
	equipmentMod.Name = "Equipment Modifier"
	equipmentModData := newDragData(equipmentMod)

	traitsProv.AltDropSupport().Drop([]int{0}, equipmentModData)
	equipmentProv.AltDropSupport().Drop([]int{0}, traitModData)
	c.Equal(0, len(trait.Modifiers), "equipment modifiers must not be attached to a trait")
	c.Equal(0, len(item.Modifiers), "trait modifiers must not be attached to equipment")
	c.Equal(0, len(*prompts), "a drop of the wrong kind of modifier must not prompt")

	traitsProv.AltDropSupport().Drop([]int{0}, traitModData)
	equipmentProv.AltDropSupport().Drop([]int{0}, equipmentModData)
	c.Equal(1, len(trait.Modifiers), "trait modifiers must still be attached to a trait")
	c.Equal(1, len(item.Modifiers), "equipment modifiers must still be attached to equipment")
	c.Equal([]modifierPrompt{
		{title: "Trait", modifiers: []string{"Trait Modifier"}},
		{title: "Item", modifiers: []string{"Equipment Modifier"}},
	}, *prompts, "each provider must prompt for the modifiers of its own kind")
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
