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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/picker"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// newTestTemplateWithEquipment returns a template dockable holding the equipment.
func newTestTemplateWithEquipment(equipment ...*gurps.Equipment) *Template {
	data := gurps.NewTemplate()
	data.Equipment = equipment
	return newTestTemplateDockable("Source", data)
}

// newTestSheetWithEquipment returns a sheet dockable carrying the equipment.
func newTestSheetWithEquipment(t *testing.T, equipment ...*gurps.Equipment) *Sheet {
	t.Helper()
	sheet := newTestSheetForTemplate(t)
	for _, one := range equipment {
		one.SetDataOwner(sheet.Entity())
	}
	sheet.Entity().CarriedEquipment = equipment
	sheet.Rebuild(true)
	return sheet
}

// equipmentConversions returns whether each of the conversion commands is available for the selected equipment.
func equipmentConversions(table *unison.Table[*Node[*gurps.Equipment]], selected ...*gurps.Equipment) (toChoice, toGroup, toContainer, toNonContainer bool) {
	selection := make(map[tid.TID]bool)
	for _, one := range selected {
		selection[one.ID()] = true
	}
	table.SetSelectionMap(selection)
	return table.CanPerformCmd(table, ConvertToChoiceContainerItemID),
		table.CanPerformCmd(table, ConvertToGroupContainerItemID),
		table.CanPerformCmd(table, ConvertToContainerItemID),
		table.CanPerformCmd(table, ConvertToNonContainerItemID)
}

// TestEquipmentConversionCommandsFollowTheSelection verifies that each kind of equipment offers exactly the
// conversions it can make, and that only a template offers to make a choice.
func TestEquipmentConversionCommandsFollowTheSelection(t *testing.T) {
	c := check.New(t)
	item := gurps.NewEquipment(nil, nil, false)
	backpack := gurps.NewEquipment(nil, nil, true)
	group := gurps.NewEquipmentGroup(nil, nil)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	template := newTestTemplateWithEquipment(item, backpack, group, choice)
	table := template.Equipment.Table

	toChoice, toGroup, toContainer, toNonContainer := equipmentConversions(table, item)
	c.False(toChoice || toGroup, "an item can only become a container")
	c.True(toContainer)
	c.False(toNonContainer)

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, backpack)
	c.False(toChoice, "a physical container must become a group before it can become a choice")
	c.True(toGroup)
	c.False(toContainer, "a physical container already is one")
	c.True(toNonContainer, "an empty physical container can become an item")

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, group)
	c.True(toChoice)
	c.False(toGroup, "a group already is one")
	c.True(toContainer, "a group can become a physical container")
	c.True(toNonContainer, "an empty group can become an item")

	toChoice, toGroup, toContainer, toNonContainer = equipmentConversions(table, choice)
	c.False(toChoice, "a choice already is one")
	c.True(toGroup)
	c.False(toContainer, "a choice must become a group before it can become a physical container")
	c.False(toNonContainer, "a choice must become a group before it can become an item")

	sheetBackpack := gurps.NewEquipment(nil, nil, true)
	sheetGroup := gurps.NewEquipmentGroup(nil, nil)
	sheet := newTestSheetWithEquipment(t, sheetBackpack, sheetGroup)
	sheetTable := sheet.CarriedEquipment.Table
	_, toGroup, _, _ = equipmentConversions(sheetTable, sheetBackpack)
	c.True(toGroup, "a sheet may hold groups")
	toChoice, _, toContainer, _ = equipmentConversions(sheetTable, sheetGroup)
	c.False(toChoice, "only a template may hold choices")
	c.True(toContainer, "a sheet may hold physical containers")
}

// TestEquipmentGroupConversionWarnsAndIsUndoable verifies that converting a physical container to a group asks before
// removing what a group can't hold, that converting back asks nothing, and that undo restores each step.
func TestEquipmentGroupConversionWarnsAndIsUndoable(t *testing.T) {
	c := check.New(t)
	backpack := gurps.NewEquipment(nil, nil, true)
	backpack.BaseValue = "60"
	backpack.BaseWeight = "3 lb"
	sheet := newTestSheetWithEquipment(t, backpack)
	table := sheet.CarriedEquipment.Table
	table.SetSelectionMap(map[tid.TID]bool{backpack.ID(): true})
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)

	var asked int
	answer := false
	swapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return answer
	})

	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing data must be confirmed first")
	c.True(backpack.IsPhysicalContainer(), "declining must leave the container alone")
	c.False(mgr.CanUndo(), "declining must not record an edit")

	answer = true
	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.True(backpack.IsGroup(), "the container must have become a group")
	c.Equal("", backpack.BaseValue, "the value must be removed")
	c.Equal("", backpack.BaseWeight, "the weight must be removed")

	mgr.Undo()
	c.True(backpack.IsPhysicalContainer(), "undo must turn it back into a physical container")
	c.Equal("60", backpack.BaseValue, "undo must restore the value")
	c.Equal("3 lb", backpack.BaseWeight, "undo must restore the weight")

	mgr.Redo()
	c.True(backpack.IsGroup(), "redo must convert it again")

	asked = 0
	table.SetSelectionMap(map[tid.TID]bool{backpack.ID(): true})
	table.PerformCmd(table, ConvertToContainerItemID)
	c.Equal(0, asked, "nothing is lost, so nothing may be asked")
	c.True(backpack.IsPhysicalContainer(), "the group must have become a physical container")
	mgr.Undo()
	c.True(backpack.IsGroup(), "undo must turn it back into a group")
}

// TestConvertToContainerTakesItemsAndGroupsTogether verifies that "Convert to Container" turns the selected items into
// containers and the selected groups into physical containers at once, as a single undoable edit.
func TestConvertToContainerTakesItemsAndGroupsTogether(t *testing.T) {
	c := check.New(t)
	item := gurps.NewEquipment(nil, nil, false)
	group := gurps.NewEquipmentGroup(nil, nil)
	sheet := newTestSheetWithEquipment(t, item, group)
	table := sheet.CarriedEquipment.Table
	table.SetSelectionMap(map[tid.TID]bool{item.ID(): true, group.ID(): true})
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)

	table.PerformCmd(table, ConvertToContainerItemID)
	c.True(item.IsPhysicalContainer(), "the item must have become a physical container")
	c.True(group.IsPhysicalContainer(), "the group must have become a physical container")

	mgr.Undo()
	c.False(item.Container(), "undo must turn the item back")
	c.True(group.IsGroup(), "the same undo must turn the group back")
	c.False(mgr.CanUndo(), "the conversion must have been a single edit")
}

// TestEquipmentChoiceConversionWarnsOnlyWhenUnequipped verifies that a group becomes a choice without a question
// unless it is unequipped, which a choice can't be.
func TestEquipmentChoiceConversionWarnsOnlyWhenUnequipped(t *testing.T) {
	c := check.New(t)
	group := gurps.NewEquipmentGroup(nil, nil)
	template := newTestTemplateWithEquipment(group)
	table := template.Equipment.Table
	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	var asked int
	swapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return true
	})

	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(0, asked, "an equipped group loses nothing")
	c.True(gurps.IsTemplateChoiceContainer(group))

	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing the choices must be confirmed first")
	c.False(gurps.IsTemplateChoiceContainer(group))

	group.Equipped = false
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(2, asked, "an unequipped group must ask first")
	c.True(group.Equipped, "the choice must be left equipped")
}

// TestAltDropSkipsEquipmentGroups verifies that dropping modifiers never targets an equipment group or choice, which
// can't hold modifiers (see gurps.CanTakeModifiers).
func TestAltDropSkipsEquipmentGroups(t *testing.T) {
	c := check.New(t)
	first := gurps.NewEquipment(nil, nil, false)
	group := gurps.NewEquipmentGroup(nil, nil)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	backpack := gurps.NewEquipment(nil, nil, true)
	template := newTestTemplateWithEquipment(first, group, choice, backpack)
	table := template.Equipment.Table
	c.Equal(0, len(altDropTargets(table, 1)), "a group under the pointer must not be targeted")
	c.Equal(0, len(altDropTargets(table, 2)), "a choice under the pointer must not be targeted")
	c.Equal([]int{3}, altDropTargets(table, 3), "a physical container may be targeted")

	table.SelectByIndex(0, 1, 2, 3)
	c.Equal([]int{0, 3}, altDropTargets(table, 0), "selected groups must be left out of the batch")
}

// TestToggleEquippedSkipsChoiceContainers verifies that a choice can't be unequipped from the list, since its editor
// offers no way to equip it again.
func TestToggleEquippedSkipsChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	item := gurps.NewEquipment(nil, nil, false)
	template := newTestTemplateWithEquipment(choice, item)
	table := template.Equipment.Table

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.False(canToggleEquipped(table), "a choice alone must not offer to be unequipped")

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true, item.ID(): true})
	toggleEquipped(template, table)
	c.True(choice.Equipped, "the choice must be left equipped")
	c.False(item.Equipped, "the item must be unequipped")
}

// TestPickerRowDetailsForEquipment verifies that the template picker dialog shows an equipment option's quantity along
// with the value and weight of all of it, and shows no such details for other kinds of options.
func TestPickerRowDetailsForEquipment(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewEquipmentChoiceContainer(nil, nil)
	c.Equal([]string{"Qty", "Value", "Weight"}, pickerRowDetailHeaders(choice))
	c.Equal(0, len(pickerRowDetailHeaders(gurps.NewTraitChoiceContainer(nil, nil))), "traits have no details")

	rope := gurps.NewEquipment(nil, choice, false)
	rope.Quantity = fxp.FromInteger(3)
	rope.BaseValue = "10"
	rope.BaseWeight = "2 lb"
	c.Equal([]string{"3", "$30", "6 lb"}, pickerRowDetails(rope))

	group := gurps.NewEquipmentGroup(nil, choice)
	rope.SetParent(group)
	group.Children = []*gurps.Equipment{rope}
	c.Equal([]string{"", "$30", "6 lb"}, pickerRowDetails(group), "a group has no quantity of its own")
}

// TestEquipmentContextMenuLeavesOutTheListName verifies that an equipment list's context menu offers to add to that list
// without naming it, naming what it holds only in its first item, while still invoking the commands for that list, and
// that only the carried list offers a choice.
func TestEquipmentContextMenuLeavesOutTheListName(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(func() { registerActions() })
	titles := func(carried bool) map[string]int {
		provider, ok := NewEquipmentProvider(gurps.NewTemplate(), carried, true).(*equipmentProvider)
		c.True(ok)
		m := make(map[string]int)
		for _, one := range provider.ContextMenuItems() {
			m[one.Title] = one.ID
		}
		return m
	}
	carried := titles(true)
	c.Equal(NewCarriedEquipmentItemID, carried["New Equipment"])
	c.Equal(NewCarriedEquipmentContainerItemID, carried["New Container"])
	c.Equal(NewCarriedEquipmentGroupItemID, carried["New Group"])
	c.Equal(NewEquipmentChoiceContainerItemID, carried["New Choice"])

	other := titles(false)
	c.Equal(NewOtherEquipmentItemID, other["New Equipment"])
	c.Equal(NewOtherEquipmentContainerItemID, other["New Container"])
	c.Equal(NewOtherEquipmentGroupItemID, other["New Group"])
	_, hasChoice := other["New Choice"]
	c.False(hasChoice, "the other equipment list never holds choices")
}

// TestPickerQuantityUpdatesDetailsAndTotal verifies that setting the quantity of an option of a choice made by value or
// weight updates what is shown for it and what it counts toward the choice.
func TestPickerQuantityUpdatesDetailsAndTotal(t *testing.T) {
	c := check.New(t)
	torch := gurps.NewEquipment(nil, nil, false)
	torch.BaseValue = "3"
	torch.BaseWeight = "1 lb"
	details := []*unison.Label{unison.NewLabel(), unison.NewLabel(), unison.NewLabel()}
	setPickerRowQuantity(torch, fxp.FromInteger(4), details)
	c.Equal(fxp.FromInteger(4), torch.Quantity)
	c.Equal("4", details[0].String())
	c.Equal("$12", details[1].String())
	c.Equal("$12", formatPickerTotal(torch, picker.Value, pickerMeasureRange(torch, picker.Value)))
	units := gurps.SheetSettingsFor(nil).DefaultWeightUnits
	c.Equal(units.Format(torch.ExtendedWeight(false, units)), details[2].String())
	c.Equal(details[2].String(), formatPickerTotal(torch, picker.Weight, pickerMeasureRange(torch, picker.Weight)))
}

// TestQuantityCommandsSkipGroups verifies that Increment and Decrement, which adjust equipment's quantity, leave a
// group alone, since its quantity is always one.
func TestQuantityCommandsSkipGroups(t *testing.T) {
	c := check.New(t)
	item := gurps.NewEquipment(nil, nil, false)
	group := gurps.NewEquipmentGroup(nil, nil)
	sheet := newTestSheetWithEquipment(t, item, group)
	table := sheet.CarriedEquipment.Table

	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	c.False(canAdjustQuantity(table, true), "a group alone must not offer to be incremented")
	c.False(canAdjustQuantity(table, false), "a group alone must not offer to be decremented")

	table.SetSelectionMap(map[tid.TID]bool{item.ID(): true, group.ID(): true})
	adjustQuantity(sheet, table, true)
	c.Equal(fxp.FromInteger(2), item.Quantity, "the item must be incremented")
	c.Equal(fxp.One, group.Quantity, "the group must be left at one")
}

// TestPickerWeightUnitsFollowTheSheet verifies that the picker dialog shows weights in the units of the sheet the rows
// are headed for, and in the default units for rows with no sheet.
func TestPickerWeightUnitsFollowTheSheet(t *testing.T) {
	c := check.New(t)
	c.Equal(gurps.SheetSettingsFor(nil).DefaultWeightUnits, pickerWeightUnits(gurps.NewEquipment(nil, nil, false)))
	entity := gurps.NewEntity()
	entity.SheetSettings.DefaultWeightUnits = fxp.Kilogram
	c.Equal(fxp.Kilogram, pickerWeightUnits(gurps.NewEquipment(entity, nil, false)))
}

// TestApplyModifierSkipsEquipmentGroups verifies that the Apply Modifier command never gives an equipment group or
// choice modifiers: its target prompt leaves them out while still offering what they hold, and the shared attach step
// skips one handed to it anyway.
func TestApplyModifierSkipsEquipmentGroups(t *testing.T) {
	c := check.New(t)
	forbidModifierPrompts(t)
	group := gurps.NewEquipmentGroup(nil, nil)
	group.Name = "Kit"
	rope := gurps.NewEquipment(nil, group, false)
	rope.Name = "Rope"
	group.Children = []*gurps.Equipment{rope}
	backpack := gurps.NewEquipment(nil, nil, true)
	backpack.Name = "Backpack"
	template := newTestTemplateWithEquipment(group, backpack)

	lists := equipmentModifierTargetKind().lists(template)
	c.Equal(1, len(lists))
	labels := make([]string, 0, 2)
	for _, one := range modifierTargetChoices(lists) {
		labels = append(labels, one.String())
	}
	c.Equal([]string{"Rope (in Kit)", "Backpack"}, labels,
		"the group must be left out of the prompt, but what it holds must still be offered")

	sturdy := gurps.NewEquipmentModifier(nil, nil, false)
	sturdy.Name = "Sturdy"
	tables := []*unison.Table[*Node[*gurps.Equipment]]{template.Equipment.Table}
	c.True(attachModifierClones("", tables, template.template, []*gurps.Equipment{group, backpack},
		[]*gurps.EquipmentModifier{sturdy}, gurps.LibraryFile{}))
	c.Equal(0, len(group.Modifiers), "the group must not be given the modifier")
	c.Equal(1, len(backpack.Modifiers), "the physical container must still get it")
}
