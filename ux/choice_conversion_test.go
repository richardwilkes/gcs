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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/container"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
)

// TestChoiceConversionCommandsFollowTheSelection verifies that the conversion commands are only available for the
// containers they can convert, and only in a template.
func TestChoiceConversionCommandsFollowTheSelection(t *testing.T) {
	c := check.New(t)
	group := gurps.NewTrait(nil, nil, true)
	meta := gurps.NewTrait(nil, nil, true)
	meta.ContainerType = container.MetaTrait
	choice := gurps.NewTraitChoiceContainer(nil, nil)
	plain := gurps.NewTrait(nil, nil, false)
	template := newTestTemplateWithTraits(group, meta, choice, plain)
	table := template.Traits.Table
	canConvert := func(selected *gurps.Trait) (toChoice, toGroup bool) {
		table.SetSelectionMap(map[tid.TID]bool{selected.ID(): true})
		return table.CanPerformCmd(table, ConvertToChoiceContainerItemID),
			table.CanPerformCmd(table, ConvertToGroupContainerItemID)
	}

	toChoice, toGroup := canConvert(group)
	c.True(toChoice, "a group can become a choice container")
	c.False(toGroup, "a group is already a group")

	toChoice, toGroup = canConvert(choice)
	c.False(toChoice, "a choice container is already one")
	c.True(toGroup, "a choice container can become a group")

	toChoice, toGroup = canConvert(meta)
	c.False(toChoice, "only a group can become a choice container")
	c.False(toGroup, "a meta-trait holds no choices to remove")

	toChoice, toGroup = canConvert(plain)
	c.False(toChoice || toGroup, "a trait that isn't a container can't be converted either way")

	sheet := newTestSheetForTemplate(t)
	sheetGroup := gurps.NewTrait(sheet.Entity(), nil, true)
	sheet.Entity().Traits = []*gurps.Trait{sheetGroup}
	sheet.Rebuild(true)
	sheetTable := sheet.Traits.Table
	sheetTable.SetSelectionMap(map[tid.TID]bool{sheetGroup.ID(): true})
	c.False(sheetTable.CanPerformCmd(sheetTable, ConvertToChoiceContainerItemID),
		"only a template may hold choices")
}

// TestChoiceConversionWarnsAndIsUndoable verifies that converting a group holding data a choice container can't keep
// asks first, removes that data, and that undo puts it all back, while a group holding nothing it can't keep is
// converted without a question.
func TestChoiceConversionWarnsAndIsUndoable(t *testing.T) {
	c := check.New(t)
	group := gurps.NewTrait(nil, nil, true)
	group.SelfControl = selfctrl.CR12
	group.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
	group.Source = gurps.Source{Library: "lib", Path: "group.adq", TID: group.ID()}
	clean := gurps.NewTrait(nil, nil, true)
	template := newTestTemplateWithTraits(group, clean)
	table := template.Traits.Table
	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)

	var asked int
	answer := false
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return answer
	})

	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(1, asked, "removing data must be confirmed first")
	c.False(gurps.IsTemplateChoiceContainer(group), "declining must leave the group alone")
	c.Equal(1, len(group.Modifiers))
	c.False(mgr.CanUndo(), "declining must not record an edit")

	answer = true
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.True(gurps.IsTemplateChoiceContainer(group), "the group must have become a choice container")
	c.Equal(0, len(group.Modifiers), "the modifiers must be removed")
	c.Equal(selfctrl.None, group.SelfControl, "the self-control roll must be removed")
	c.Equal(gurps.Source{}, group.Source, "the source must be removed")

	mgr.Undo()
	c.False(gurps.IsTemplateChoiceContainer(group), "undo must turn it back into a group")
	c.Equal(1, len(group.Modifiers), "undo must restore the modifiers")
	c.Equal(selfctrl.CR12, group.SelfControl, "undo must restore the self-control roll")
	c.Equal("group.adq", group.Source.Path, "undo must restore the source")

	mgr.Redo()
	c.True(gurps.IsTemplateChoiceContainer(group), "redo must convert it again")

	asked = 0
	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing the choices must be confirmed first")
	c.False(gurps.IsTemplateChoiceContainer(group), "the choice container must have become a group")
	mgr.Undo()
	c.True(gurps.IsTemplateChoiceContainer(group), "undo must restore the choices")

	asked = 0
	answer = false
	table.SetSelectionMap(map[tid.TID]bool{clean.ID(): true})
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(0, asked, "nothing is lost, so nothing may be asked")
	c.True(gurps.IsTemplateChoiceContainer(clean), "the clean group must have become a choice container")
}

// TestToggleDisabledSkipsChoiceContainers verifies that a choice container can't be disabled from the list, since its
// editor offers no way to enable it again, while one that is already disabled can still be enabled.
func TestToggleDisabledSkipsChoiceContainers(t *testing.T) {
	c := check.New(t)
	choice := gurps.NewTraitChoiceContainer(nil, nil)
	plain := gurps.NewTrait(nil, nil, false)
	template := newTestTemplateWithTraits(choice, plain)
	table := template.Traits.Table

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.False(canToggleDisabled(table), "a choice alone must not offer to be disabled")

	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true, plain.ID(): true})
	c.True(canToggleDisabled(table), "the rest of the selection may still be toggled")
	toggleDisabled(template, table)
	c.False(choice.IsDisabled(), "the choice must be left enabled")
	c.True(plain.IsDisabled(), "the plain trait must be disabled")

	choice.SetDisabled(true)
	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.True(canToggleDisabled(table), "a disabled choice must offer to be enabled")
	toggleDisabled(template, table)
	c.False(choice.IsDisabled(), "the disabled choice must be enabled")
}

// TestChoiceConversionClosesEditors verifies that no editor is left open on a container whose kind has changed under
// it. Converting closes the editor first, and undo and redo discard an editor opened on the other kind of container,
// since applying it would silently convert the container back and throw away what the undo or redo restored. A modified
// editor opened from that editor, on one of the group's modifiers, is discarded with it without being asked, since what
// it would save goes into the data being dropped.
func TestChoiceConversionClosesEditors(t *testing.T) {
	c := check.New(t)
	screen, _ := uxtest.StartHeadlessWorkspace(t, c)
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool { return true })
	group := gurps.NewTrait(nil, nil, true)
	group.Name = "Advantages"
	group.Tags = []string{"Advantage"}
	modifier := gurps.NewTraitModifier(nil, nil, false)
	modifier.Name = "Long"
	group.Modifiers = []*gurps.TraitModifier{modifier}
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{group}
	var template *Template
	screen.Do(func() {
		template = NewTemplate("test"+gurps.TemplatesExt, data)
		DisplayNewDockable(template)
	})
	isGroupEditor := func(d unison.Dockable) bool {
		e, ok := d.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return ok && e.target == group
	}
	isModifierEditor := func(d unison.Dockable) bool {
		_, ok := d.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return ok
	}
	open := func(match func(unison.Dockable) bool) int {
		var count int
		screen.Do(func() { count = len(AllMatchingDockables(match)) })
		return count
	}
	editorsOpen := func() int { return open(isGroupEditor) }
	var mgr *unison.UndoManager
	screen.Do(func() {
		mgr = unison.UndoManagerFor(template.Traits.Table)
		EditTrait(template, group)
	})
	c.NotNil(mgr, "the template must have an undo manager")
	c.Equal(1, editorsOpen(), "the group's editor must be open")

	screen.Do(func() {
		table := template.Traits.Table
		table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
		convertSelectedContainers[*gurps.Trait, *gurps.TraitEditData](template, table, choiceContainerKind)
	})
	c.True(gurps.IsTemplateChoiceContainer(group), "the group must have become a choice")
	c.Equal(0, editorsOpen(), "converting must close the group's editor")

	screen.Do(func() { EditTrait(template, group) })
	c.Equal(1, editorsOpen(), "the choice's editor must be open")
	screen.Do(mgr.Undo)
	c.False(gurps.IsTemplateChoiceContainer(group), "undo must turn the choice back into a group")
	c.Equal([]string{"Advantage"}, group.Tags, "undo must restore what the conversion removed")
	c.Equal(0, editorsOpen(), "undo must discard the editor opened on the choice")

	screen.Do(func() { EditTrait(template, group) })
	c.Equal(1, editorsOpen(), "the group's editor must be open")
	groupEditor := uxtest.SoleEditor[*editor[*gurps.Trait, *gurps.TraitEditData]](t, screen, isGroupEditor)
	screen.Do(func() { EditTraitModifier(groupEditor, groupEditor.editorData.Modifiers[0]) })
	modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		isModifierEditor)
	screen.Do(func() {
		modifierEditor.editorData.Name = "Longer"
		modifierEditor.MarkModified(nil)
	})
	// The redo is posted rather than run through Do, since a prompt about the modifier editor's changes, which there
	// must not be, would run a modal loop that Do would wait on.
	screen.Post(mgr.Redo)
	screen.Sync()
	var windows int
	screen.Do(func() { windows = len(unison.Windows()) })
	c.Equal(1, windows, "redo must not ask about the changes made in the editor opened from the group's editor")
	c.True(gurps.IsTemplateChoiceContainer(group), "redo must convert the group again")
	c.Equal(0, editorsOpen(), "redo must discard the editor opened on the group")
	c.Equal(0, open(isModifierEditor), "and the editor opened from it")
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
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
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

// TestModifierChoiceConversionInLibrary verifies that a modifier library converts a group to a choice without asking,
// since nothing is lost, converts it back after asking, since the choice is, and that both can be undone.
func TestModifierChoiceConversionInLibrary(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	group := gurps.NewTraitModifier(nil, nil, true)
	for range 2 {
		group.Children = append(group.Children, gurps.NewTraitModifier(nil, group, false))
	}
	leaf := gurps.NewTraitModifier(nil, nil, false)
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{group, leaf})
	group = library.table.RootRows()[0].Data()
	enabled := func() []bool {
		return []bool{group.Children[0].Enabled(), group.Children[1].Enabled()}
	}
	table := library.table
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)
	canConvert := func(selected *gurps.TraitModifier) (toChoice, toGroup bool) {
		table.SetSelectionMap(map[tid.TID]bool{selected.ID(): true})
		return table.CanPerformCmd(table, ConvertToChoiceContainerItemID),
			table.CanPerformCmd(table, ConvertToGroupContainerItemID)
	}

	toChoice, toGroup := canConvert(leaf)
	c.False(toChoice || toGroup, "a modifier that isn't a container can't be converted either way")
	toChoice, toGroup = canConvert(group)
	c.True(toChoice, "a group can become a choice, outside a template too")
	c.False(toGroup, "a group is already a group")

	var asked int
	uxtest.SwapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return true
	})
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(0, asked, "nothing is lost, so nothing may be asked")
	c.True(gurps.IsMandatoryModifierChoice(group), "a new choice is a mandatory one")
	c.Equal([]bool{true, false}, enabled(), "a choice keeps no more than one option on")
	toChoice, toGroup = canConvert(group)
	c.False(toChoice)
	c.True(toGroup)

	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing the choice must be confirmed first")
	c.False(gurps.IsModifierChoice(group))
	c.Equal([]bool{true, false}, enabled())
	mgr.Undo()
	c.True(gurps.IsModifierChoice(group), "undo must restore the choice")
	c.Equal([]bool{true, false}, enabled())
	mgr.Undo()
	c.False(gurps.IsModifierChoice(group), "undo must turn it back into a group")
	c.Equal([]bool{true, true}, enabled(), "undo must turn the second option back on")
}

// TestConvertingNestedGroupsTogetherUndoesCleanly verifies that converting a group and a group within it to choices in
// one go, which changes the enabled states of the tree they share, puts every one of them back on undo, and takes them
// to where the conversion left them on redo.
func TestConvertingNestedGroupsTogetherUndoesCleanly(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	newLeaf := func(parent *gurps.TraitModifier, name string) *gurps.TraitModifier {
		leaf := gurps.NewTraitModifier(nil, parent, false)
		leaf.Name = name
		parent.Children = append(parent.Children, leaf)
		return leaf
	}
	g1 := gurps.NewTraitModifier(nil, nil, true)
	a := newLeaf(g1, "A")
	b := newLeaf(g1, "B")
	g2 := gurps.NewTraitModifier(nil, g1, true)
	g1.Children = append(g1.Children, g2)
	d1 := newLeaf(g2, "C")
	d2 := newLeaf(g2, "D")
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{g1})
	table := library.table
	enabled := func() []bool { return []bool{a.Enabled(), b.Enabled(), d1.Enabled(), d2.Enabled()} }
	c.Equal([]bool{true, true, true, true}, enabled())

	table.SetSelectionMap(map[tid.TID]bool{g1.ID(): true, g2.ID(): true})
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.True(gurps.IsModifierChoice(g1))
	c.True(gurps.IsModifierChoice(g2))
	converted := enabled()
	c.Equal([]bool{true, false, false, false}, converted, "only A is left on")

	mgr := unison.UndoManagerFor(table)
	mgr.Undo()
	c.False(gurps.IsModifierChoice(g1))
	c.False(gurps.IsModifierChoice(g2))
	c.Equal([]bool{true, true, true, true}, enabled(), "undo must turn every option back on")
	mgr.Redo()
	c.Equal(converted, enabled(), "redo must leave them as the conversion did")
}

// TestConvertingToAChoiceOnASheetPicksTheFirstOption verifies that a group converted to a choice in the editor of a
// trait on a sheet, where a mandatory choice must be made, gets its first option picked when it had none, that the
// trait's own modifier waits for the edit to be applied, and that undo puts the pick back.
func TestConvertingToAChoiceOnASheetPicksTheFirstOption(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	sheet := newTestSheetForTemplate(t)
	group := newTraitModifierChoiceFor(sheet.Entity(), true, []string{"A", "B"})
	gurps.ConvertFromModifierChoice(group)
	e, _, table := traitEditorOnSheet(t, sheet, group)
	copyOfGroup := e.editorData.Modifiers[0]
	table.SetSelectionMap(map[tid.TID]bool{copyOfGroup.ID(): true})
	table.PerformCmd(nil, ConvertToChoiceContainerItemID)
	c.True(gurps.IsMandatoryModifierChoice(copyOfGroup))
	c.True(copyOfGroup.Children[0].Enabled(), "the first option is picked")
	c.False(copyOfGroup.Children[1].Enabled())
	c.False(gurps.IsModifierChoice(e.target.Modifiers[0]), "the trait's own modifier waits for the edit to be applied")
	mgr := unison.UndoManagerFor(table)
	mgr.Undo()
	c.False(gurps.IsModifierChoice(copyOfGroup))
	c.False(copyOfGroup.Children[0].Enabled(), "undo turns the pick back off")
	mgr.Redo()
	c.True(gurps.IsMandatoryModifierChoice(copyOfGroup))
	c.True(copyOfGroup.Children[0].Enabled(), "redo picks the first option again")
}
