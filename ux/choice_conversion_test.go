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
// asks first, removes that data, and that undo puts it all back.
func TestChoiceConversionWarnsAndIsUndoable(t *testing.T) {
	c := check.New(t)
	group := gurps.NewTrait(nil, nil, true)
	group.SelfControl = selfctrl.CR12
	group.Modifiers = []*gurps.TraitModifier{gurps.NewTraitModifier(nil, nil, false)}
	group.Source = gurps.Source{Library: "lib", Path: "group.adq", TID: group.ID()}
	template := newTestTemplateWithTraits(group)
	table := template.Traits.Table
	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	mgr := unison.UndoManagerFor(table)
	c.NotNil(mgr)

	var asked int
	answer := false
	swapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
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
}

// TestChoiceConversionOfCleanGroupDoesNotAsk verifies that a group holding nothing a choice container can't keep is
// converted without a question.
func TestChoiceConversionOfCleanGroupDoesNotAsk(t *testing.T) {
	c := check.New(t)
	group := gurps.NewTrait(nil, nil, true)
	template := newTestTemplateWithTraits(group)
	table := template.Traits.Table
	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	swapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		t.Error("nothing is lost, so nothing may be asked")
		return false
	})
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.True(gurps.IsTemplateChoiceContainer(group))
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
	c.False(choice.Disabled, "the choice must be left enabled")
	c.True(plain.Disabled, "the plain trait must be disabled")

	choice.Disabled = true
	table.SetSelectionMap(map[tid.TID]bool{choice.ID(): true})
	c.True(canToggleDisabled(table), "a disabled choice must offer to be enabled")
	toggleDisabled(template, table)
	c.False(choice.Disabled, "the disabled choice must be enabled")
}
