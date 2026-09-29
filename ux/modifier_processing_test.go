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
	swapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, modifiers []*gurps.TraitModifier) (changed, canceled bool) {
		title := info.name
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false, false
	})
	swapForTest(t, &promptForEquipmentModifiers, func(info *modifierPromptInfo, modifiers []*gurps.EquipmentModifier) (changed, canceled bool) {
		title := info.name
		p := modifierPrompt{title: title}
		for _, one := range modifiers {
			p.modifiers = append(p.modifiers, one.Name)
		}
		prompts = append(prompts, p)
		return false, false
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

// forbidModifierPrompts substitutes modifier and nameables prompts that fail the test if shown, so that a test which
// expects none fails rather than waits on a real dialog. A test that expects the nameables prompt substitutes its own
// responder after calling this.
func forbidModifierPrompts(t *testing.T) {
	t.Helper()
	swapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
		title := info.name
		t.Errorf("the modifier prompt must not be shown, but was shown for %q", title)
		return false, false
	})
	swapForTest(t, &promptForEquipmentModifiers, func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
		title := info.name
		t.Errorf("the modifier prompt must not be shown, but was shown for %q", title)
		return false, false
	})
	swapForTest(t, &promptForNameables, slicedNameablesPrompt(func(titles []string, _ []map[string]string, _ [][]string) bool {
		t.Errorf("the nameables prompt must not be shown, but was shown for %v", titles)
		return false
	}))
}

// processModifiers prompts for the modifiers of the rows as applyTransfer does.
func processModifiers[T gurps.Node[T]](rows []T) bool {
	targets := modifierTargets(rows)
	return promptForModifierTargets(promptOperation{}, targets, 0, len(targets))
}

// TestProcessModifiersIgnoresModifierRows documents that processModifiers only has something to do for rows that can
// hold modifiers. Handing it the modifiers themselves matches nothing, which is why its callers pass the rows that
// carry the modifiers (see applyTransfer).
func TestProcessModifiersIgnoresModifierRows(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	entity := gurps.NewEntity()

	traitMod := gurps.NewTraitModifier(entity, nil, false)
	traitMod.Name = "Trait Modifier"
	processModifiers([]*gurps.TraitModifier{traitMod})
	equipmentMod := gurps.NewEquipmentModifier(entity, nil, false)
	equipmentMod.Name = "Equipment Modifier"
	processModifiers([]*gurps.EquipmentModifier{equipmentMod})
	c.Equal(0, len(*prompts), "modifier rows have no modifiers of their own to prompt for")

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	trait.Modifiers = []*gurps.TraitModifier{traitMod}
	processModifiers([]*gurps.Trait{trait})
	c.Equal([]modifierPrompt{{title: "Trait", modifiers: []string{"Trait Modifier"}}}, *prompts,
		"a trait must be prompted for with its own modifiers")
}

// TestModifierPromptsCountOnlyRowsWithModifiers verifies that a row without modifiers is neither prompted for nor
// counted, so the count the prompts show is of the prompts the user will actually see.
func TestModifierPromptsCountOnlyRowsWithModifiers(t *testing.T) {
	c := check.New(t)
	var steps [][2]int
	swapForTest(t, &promptForTraitModifiers, func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
		steps = append(steps, [2]int{info.step, info.steps})
		return false, false
	})
	entity := gurps.NewEntity()
	plain := gurps.NewTrait(entity, nil, false)
	first := gurps.NewTrait(entity, nil, false)
	first.AddModifiers(gurps.NewTraitModifier(entity, nil, false))
	second := gurps.NewTrait(entity, nil, false)
	second.AddModifiers(gurps.NewTraitModifier(entity, nil, false))
	c.True(processModifiers([]*gurps.Trait{plain, first, plain, second}))
	c.Equal([][2]int{{1, 2}, {2, 2}}, steps)
}

// TestAltDropOnTraitSwitchesTheDroppedModifierOn verifies that dropping a trait modifier onto a trait row adds an
// enabled copy of it without prompting, leaving the trait's existing modifiers and the dragged modifier as they were.
func TestAltDropOnTraitSwitchesTheDroppedModifierOn(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity, traits := entityWithNamedTraits("Target Trait")
	target := traits[0]
	existing := gurps.NewTraitModifier(entity, nil, false)
	existing.Name = "Existing"
	existing.Disabled = true
	target.Modifiers = []*gurps.TraitModifier{existing}
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	dropped.Disabled = true

	altDrop(tabledProvider(NewTraitsProvider(entity, false)).AltDropSupport(), []int{0}, dropped)

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target trait")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target trait")
	c.False(target.Modifiers[1].Disabled, "the copy must be switched on")
	c.True(target.Modifiers[0].Disabled, "the trait's existing modifiers must be left alone")
	c.True(dropped.Disabled, "the dragged modifier itself must be left alone")
}

// TestAltDropOnEquipmentSwitchesTheDroppedModifierOn verifies the same for dropping an equipment modifier onto an
// equipment row.
func TestAltDropOnEquipmentSwitchesTheDroppedModifierOn(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity, equipment := entityWithNamedEquipment("Target Equipment")
	target := equipment[0]
	existing := gurps.NewEquipmentModifier(entity, nil, false)
	existing.Name = "Existing"
	existing.Disabled = true
	target.Modifiers = []*gurps.EquipmentModifier{existing}
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"
	dropped.Disabled = true

	altDrop(tabledProvider(NewEquipmentProvider(entity, true, false)).AltDropSupport(), []int{0}, dropped)

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target equipment")
	c.False(target.Modifiers[1].Disabled, "the copy must be switched on")
	c.True(target.Modifiers[0].Disabled, "the item's existing modifiers must be left alone")
	c.True(dropped.Disabled, "the dragged modifier itself must be left alone")
}

// TestAltDropOnSeveralTraitsGivesEachItsOwnCopy verifies that dropping a trait modifier onto several selected traits
// attaches a separate copy to each of them. A single shared modifier would tie the traits together, so that enabling
// it on one would enable it on all and renaming it would rename it everywhere.
func TestAltDropOnSeveralTraitsGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity, traits := entityWithNamedTraits("First Trait", "Second Trait")
	first, second := traits[0], traits[1]
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"

	altDrop(tabledProvider(NewTraitsProvider(entity, false)).AltDropSupport(), []int{0, 1}, dropped)

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
	forbidModifierPrompts(t)
	entity, equipment := entityWithNamedEquipment("First Item", "Second Item")
	first, second := equipment[0], equipment[1]
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"

	altDrop(tabledProvider(NewEquipmentProvider(entity, true, false)).AltDropSupport(), []int{0, 1}, dropped)

	c.Equal(1, len(first.Modifiers), "the dropped modifier must be added to the first item")
	c.Equal(1, len(second.Modifiers), "the dropped modifier must be added to the second item")
	c.Equal("Dropped", first.Modifiers[0].Name, "the first item must get the dropped modifier")
	c.Equal("Dropped", second.Modifiers[0].Name, "the second item must get the dropped modifier")
	c.NotEqual(first.Modifiers[0].ID(), second.Modifiers[0].ID(), "each item must get a copy of its own")
	c.NotEqual(dropped.ID(), first.Modifiers[0].ID(), "the dragged modifier itself must not be attached")
}

// TestAltDropOnAContainerAndItsChildGivesEachItsOwnCopy verifies that a selection holding both a container and one of
// its own descendants gives each of them a copy of the dropped modifier, since both were pointed at.
func TestAltDropOnAContainerAndItsChildGivesEachItsOwnCopy(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
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

	c.Equal(1, len(container.Modifiers), "the container must receive the dropped modifier")
	c.Equal(1, len(child.Modifiers), "the child must receive its own copy of the dropped modifier")
	c.NotEqual(container.Modifiers[0].ID(), child.Modifiers[0].ID(), "each must get a copy of its own")
}

// TestAltDropOnAMissingTraitRowIsANoOp verifies that a row index which resolves to nothing is skipped, and that a drop
// none of whose indexes resolve reports that it changed nothing, so that no edit is recorded.
func TestAltDropOnAMissingTraitRowIsANoOp(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity, traits := entityWithNamedTraits("Trait")
	trait := traits[0]
	provider := tabledProvider(NewTraitsProvider(entity, false))
	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"

	c.False(provider.AltDropSupport().Drop([]int{99}, newDragData(dropped)),
		"a drop with no resolvable target changes nothing, so it must say so")
	c.Equal(0, len(trait.Modifiers), "a drop with no resolvable target must attach nothing")

	c.True(provider.AltDropSupport().Drop([]int{99, 0}, newDragData(dropped)))
	c.Equal(1, len(trait.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
}

// TestAltDropOnAMissingEquipmentRowIsANoOp verifies the same for the equipment drop handler.
func TestAltDropOnAMissingEquipmentRowIsANoOp(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	entity, equipment := entityWithNamedEquipment("Item")
	item := equipment[0]
	provider := tabledProvider(NewEquipmentProvider(entity, true, false))
	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"

	c.False(provider.AltDropSupport().Drop([]int{99}, newDragData(dropped)),
		"a drop with no resolvable target changes nothing, so it must say so")
	c.Equal(0, len(item.Modifiers), "a drop with no resolvable target must attach nothing")

	c.True(provider.AltDropSupport().Drop([]int{99, 0}, newDragData(dropped)))
	c.Equal(1, len(item.Modifiers), "an index that resolves to nothing must be skipped rather than stop the drop")
}

// TestAltDropOfTheWrongKindOfModifierIsIgnored verifies that each provider's alternate drop handler only acts on drag
// data holding its own kind of modifier, reporting that nothing changed otherwise.
func TestAltDropOfTheWrongKindOfModifierIsIgnored(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
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

	c.False(traitsProv.AltDropSupport().Drop([]int{0}, equipmentModData),
		"a drop of the wrong kind of modifier changes nothing, so it must say so, or an empty edit would be recorded")
	c.False(equipmentProv.AltDropSupport().Drop([]int{0}, traitModData),
		"a drop of the wrong kind of modifier changes nothing, so it must say so, or an empty edit would be recorded")
	c.Equal(0, len(trait.Modifiers), "equipment modifiers must not be attached to a trait")
	c.Equal(0, len(item.Modifiers), "trait modifiers must not be attached to equipment")

	c.True(traitsProv.AltDropSupport().Drop([]int{0}, traitModData))
	c.True(equipmentProv.AltDropSupport().Drop([]int{0}, equipmentModData))
	c.Equal(1, len(trait.Modifiers), "trait modifiers must still be attached to a trait")
	c.Equal(1, len(item.Modifiers), "equipment modifiers must still be attached to equipment")
}
