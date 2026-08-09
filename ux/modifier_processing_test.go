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
	table := unison.NewTable[*Node[*gurps.Trait]](provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewTraitModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable[*Node[*gurps.TraitModifier]](&unison.SimpleTableModel[*Node[*gurps.TraitModifier]]{})
	provider.AltDropSupport().Drop(0, &unison.TableDragData[*Node[*gurps.TraitModifier]]{
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
	table := unison.NewTable[*Node[*gurps.Equipment]](provider)
	provider.SetTable(table)
	table.SetRootRows(provider.RootRows())

	dropped := gurps.NewEquipmentModifier(entity, nil, false)
	dropped.Name = "Dropped"
	modTable := unison.NewTable[*Node[*gurps.EquipmentModifier]](
		&unison.SimpleTableModel[*Node[*gurps.EquipmentModifier]]{})
	provider.AltDropSupport().Drop(0, &unison.TableDragData[*Node[*gurps.EquipmentModifier]]{
		Table: modTable,
		Rows:  []*Node[*gurps.EquipmentModifier]{NewNode(modTable, nil, dropped, false)},
	})

	c.Equal(2, len(target.Modifiers), "the dropped modifier must be added to the target equipment")
	c.Equal("Dropped", target.Modifiers[1].Name, "the dropped modifier must be added to the target equipment")
	c.Equal([]modifierPrompt{{title: "Target Equipment", modifiers: []string{"Existing", "Dropped"}}}, *prompts,
		"the drop must prompt for the target equipment's modifiers")
}
