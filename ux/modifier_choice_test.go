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
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	checkstate "github.com/richardwilkes/unison/enums/check"
	"github.com/richardwilkes/unison/enums/mod"
)

// newTraitModifierChoiceFor returns a trait modifier choice owned by owner, holding an option for each name, each
// enabled only if its name is among those given as enabled.
func newTraitModifierChoiceFor(owner gurps.DataOwner, mandatory bool, names []string, enabled ...string) *gurps.TraitModifier {
	choice := gurps.NewTraitModifierChoice(owner, nil)
	choice.SetMandatoryChoice(mandatory)
	for _, name := range names {
		option := gurps.NewTraitModifier(owner, choice, false)
		option.Name = name
		option.SetEnabled(slices.Contains(enabled, name))
		choice.Children = append(choice.Children, option)
	}
	return choice
}

// labelTitles returns the titles of every label found anywhere beneath the given panel.
func labelTitles(p *unison.Panel) []string {
	labels := panelsOfType[*unison.Label](p)
	titles := make([]string, 0, len(labels))
	for _, label := range labels {
		titles = append(titles, label.String())
	}
	return titles
}

// TestModifierContainerEditorsShowOnlyWhatAContainerUses verifies that a modifier group's editor leaves out every
// field that only a modifier in its own right uses, and that a choice's editor adds only what the choice asks for.
func TestModifierContainerEditorsShowOnlyWhatAContainerUses(t *testing.T) {
	c := check.New(t)
	leafOnly := []string{"Cost", "Level", "Total", "Tech Level", "Cost Modifier", "Weight Modifier", "VTT Notes"}
	checkNone := func(titles []string) {
		c.Helper()
		for _, title := range leafOnly {
			c.False(slices.Contains(titles, title), "a container's editor must not offer %q", title)
		}
	}

	_, content := buildEditorContent(nil, gurps.NewTraitModifier(nil, nil, true), initTraitModifierEditor)
	titles := labelTitles(content)
	checkNone(titles)
	c.True(slices.Contains(titles, "Notes"))
	c.False(slices.Contains(titles, "Choice"), "a group asks for no choice")
	c.Equal(0, len(checkBoxesTitled(content, "Enabled")), "a container is always enabled")
	c.Equal(0, len(checkBoxesTitled(content, "Also show notes in weapon usage")))

	choice := gurps.NewEquipmentModifierChoice(nil, nil)
	e, content := buildEditorContent(nil, choice, initEquipmentModifierEditor)
	titles = labelTitles(content)
	checkNone(titles)
	c.True(slices.Contains(titles, "Choice"), "a choice says what it asks for")
	popups := panelsOfType[*unison.PopupMenu[string]](content)
	c.Equal(1, len(popups))
	popups[0].SelectIndex(1)
	c.False(e.editorData.IsMandatoryChoice(), "the second item makes the choice optional")
	c.True(e.editorData.IsChoice(), "the choice can't be taken out of use here")
	c.True(choice.IsMandatoryChoice(), "nothing reaches the modifier until the edit is applied")
}

// TestModifierChoiceConversionInLibrary verifies that a modifier library converts a group to a choice without asking,
// since nothing is lost, converts it back after asking, since the choice is, and that both can be undone.
func TestModifierChoiceConversionInLibrary(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	group := gurps.NewTraitModifier(nil, nil, true)
	leaf := gurps.NewTraitModifier(nil, nil, false)
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{group, leaf})
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
	swapForTest(t, &askToConvertChoiceContainers, func(_, _ string) bool {
		asked++
		return true
	})
	table.PerformCmd(table, ConvertToChoiceContainerItemID)
	c.Equal(0, asked, "nothing is lost, so nothing may be asked")
	c.True(gurps.IsMandatoryModifierChoice(group), "a new choice is a mandatory one")
	toChoice, toGroup = canConvert(group)
	c.False(toChoice)
	c.True(toGroup)

	table.PerformCmd(table, ConvertToGroupContainerItemID)
	c.Equal(1, asked, "removing the choice must be confirmed first")
	c.False(gurps.IsModifierChoice(group))
	mgr.Undo()
	c.True(gurps.IsModifierChoice(group), "undo must restore the choice")
	mgr.Undo()
	c.False(gurps.IsModifierChoice(group), "undo must turn it back into a group")
}

// TestModifierChoiceConversionInEditor verifies that the conversions reach the modifier table of a trait editor, where
// only the editor's copy is converted.
func TestModifierChoiceConversionInEditor(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	e, table, _ := newTraitEditorWithModifiers(t)
	group := e.editorData.Modifiers[1]
	table.SetSelectionMap(map[tid.TID]bool{group.ID(): true})
	c.True(table.CanPerformCmd(nil, ConvertToChoiceContainerItemID))
	table.PerformCmd(nil, ConvertToChoiceContainerItemID)
	c.True(gurps.IsModifierChoice(group), "the editor's copy must have become a choice")
	c.False(gurps.IsModifierChoice(e.target.Modifiers[1]), "the trait's own modifier must wait for the edit to be applied")
	c.True(unison.UndoManagerFor(table).CanUndo())
}

// TestCreatingModifierChoices verifies that the "New ... Modifier Choice" commands are offered by a modifier library
// and by an editor's modifier table, in the menus as well, and that what they create is a mandatory choice.
func TestCreatingModifierChoices(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, nil)
	c.True(library.CanPerformCmd(nil, NewTraitModifierChoiceItemID))
	provider, ok := library.provider.(*modifiersProvider[*gurps.TraitModifier])
	c.True(ok)
	choice := provider.newChoice(nil, nil)
	c.True(gurps.IsMandatoryModifierChoice(choice))
	c.Equal("Trait Modifier Choice", choice.Name)
	items := provider.ContextMenuItems()
	titles := make([]string, 0, len(items))
	for _, item := range items {
		titles = append(titles, item.Title)
	}
	c.True(slices.Contains(titles, "New Trait Modifier Group"), "a modifier container is now called a group")
	c.True(slices.Contains(titles, "New Trait Modifier Choice"))

	e, _, _ := newEquipmentEditorWithModifiers(t)
	c.True(e.CanPerformCmd(nil, NewEquipmentModifierChoiceItemID))
	c.True(gurps.IsMandatoryModifierChoice(gurps.NewEquipmentModifierChoice(nil, nil)))
}

// TestTogglingAnOptionOnTurnsTheOthersOff verifies that turning on an option of a modifier choice turns off the one
// that was on, as a single undoable edit, and that on a sheet the pick of a mandatory choice can't be turned off,
// only replaced by another.
func TestTogglingAnOptionOnTurnsTheOthersOff(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{
		newTraitModifierChoiceFor(entity, true, []string{"A", "B", "C"}, "A"),
	}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := firstPanelOfType[*traitModifiersPanel](content)
	c.True(ok)
	table := panel.table
	options := e.editorData.Modifiers[0].Children
	enabled := func() []bool {
		return []bool{options[0].Enabled(), options[1].Enabled(), options[2].Enabled()}
	}

	table.SetSelectionMap(map[tid.TID]bool{options[1].ID(): true})
	table.PerformCmd(nil, ToggleStateItemID)
	c.Equal([]bool{false, true, false}, enabled(), "turning B on must turn A off")
	mgr := unison.UndoManagerFor(table)
	mgr.Undo()
	c.Equal([]bool{true, false, false}, enabled(), "undo must put both back")
	mgr.Redo()
	c.Equal([]bool{false, true, false}, enabled())

	table.SetSelectionMap(map[tid.TID]bool{options[0].ID(): true, options[2].ID(): true})
	table.PerformCmd(nil, ToggleStateItemID)
	c.Equal([]bool{false, false, true}, enabled(), "turning two on together leaves only the later one on")

	table.SetSelectionMap(map[tid.TID]bool{options[2].ID(): true})
	c.False(table.CanPerformCmd(nil, ToggleStateItemID), "the pick of a mandatory choice on a sheet can't be turned off")
	adjustModifierEnabled(e, table, options[2], false)
	c.Equal([]bool{false, false, true}, enabled(), "not by its checkmark either")

	adjustModifierEnabled(e, table, options[1], true)
	c.Equal([]bool{false, true, false}, enabled(), "the checkmark cell turns the others off as the command does")
}

// TestPreconfiguredAsksOnlyAboutUnresolvedChoices verifies that a preconfigured trait is asked about the mandatory
// modifier choices it has left without a pick, and about nothing else: not its other modifiers, not its optional
// choices and not the mandatory choices it has already picked for.
func TestPreconfiguredAsksOnlyAboutUnresolvedChoices(t *testing.T) {
	c := check.New(t)
	prompts := captureModifierPrompts(t)
	plain := gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Plain"
	made := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"}, "High")
	made.Name = "Made"
	open := newTraitModifierChoiceFor(nil, true, []string{"Hot", "Cold"})
	open.Name = "Open"
	optional := newTraitModifierChoiceFor(nil, false, []string{"Loud", "Quiet"})
	optional.Name = "Optional"
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Blast"
	trait.Modifiers = []*gurps.TraitModifier{plain, made, open, optional}
	trait.Preconfigured = true

	c.True(processModifiers(promptOperation{}, []*gurps.Trait{trait}, true))
	c.Equal([]modifierPrompt{{title: "Blast", modifiers: []string{"Open"}}}, *prompts,
		"only the unresolved mandatory choice is asked about")

	*prompts = nil
	open.Children[0].SetEnabled(true)
	c.True(processModifiers(promptOperation{}, []*gurps.Trait{trait}, true))
	c.Equal(0, len(*prompts), "with every mandatory choice made, a preconfigured trait isn't asked at all")

	*prompts = nil
	trait.Preconfigured = false
	c.True(processModifiers(promptOperation{}, []*gurps.Trait{trait}, true))
	c.Equal([]modifierPrompt{{title: "Blast", modifiers: []string{"Plain", "Made", "Open", "Optional"}}}, *prompts,
		"a trait that isn't preconfigured is asked about everything")
}

// TestDuplicatingAnOptionKeepsThePick verifies that an option duplicated within a choice arrives turned off, so that
// the choice keeps the pick it had, and that undo takes the duplicate away again.
func TestDuplicatingAnOptionKeepsThePick(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A")}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := firstPanelOfType[*traitModifiersPanel](content)
	c.True(ok)
	table := panel.table
	choice := e.editorData.Modifiers[0]
	table.SetSelectionMap(map[tid.TID]bool{choice.Children[0].ID(): true})
	DuplicateSelection(table)
	c.Equal(3, len(choice.Children))
	c.True(choice.Children[0].Enabled(), "the pick the choice had is kept")
	c.False(choice.Children[1].Enabled(), "the duplicate arrives turned off")
	unison.UndoManagerFor(table).Undo()
	c.Equal(2, len(e.editorData.Modifiers[0].Children), "undo takes the duplicate away")
}

// TestConvertingToAChoiceOnASheetPicksTheFirstOption verifies that a group converted to a choice on a sheet, where a
// mandatory choice must be made, gets its first option picked when it had none, and that undo puts that back.
func TestConvertingToAChoiceOnASheetPicksTheFirstOption(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	group := gurps.NewTraitModifier(entity, nil, true)
	for _, name := range []string{"A", "B"} {
		option := gurps.NewTraitModifier(entity, group, false)
		option.Name = name
		option.SetEnabled(false)
		group.Children = append(group.Children, option)
	}
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{group}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := firstPanelOfType[*traitModifiersPanel](content)
	c.True(ok)
	table := panel.table
	copyOfGroup := e.editorData.Modifiers[0]
	table.SetSelectionMap(map[tid.TID]bool{copyOfGroup.ID(): true})
	table.PerformCmd(nil, ConvertToChoiceContainerItemID)
	c.True(gurps.IsMandatoryModifierChoice(copyOfGroup))
	c.True(copyOfGroup.Children[0].Enabled(), "the first option is picked")
	c.False(copyOfGroup.Children[1].Enabled())
	mgr := unison.UndoManagerFor(table)
	mgr.Undo()
	c.False(gurps.IsModifierChoice(copyOfGroup))
	c.False(copyOfGroup.Children[0].Enabled(), "undo turns the pick back off")
	mgr.Redo()
	c.True(gurps.IsMandatoryModifierChoice(copyOfGroup))
	c.True(copyOfGroup.Children[0].Enabled(), "redo picks the first option again")
}

// TestUnresolvedChoiceLeadsTheTooltip verifies that the explanation of an unresolved modifier choice goes ahead of
// whatever else a row's tooltip says.
func TestUnresolvedChoiceLeadsTheTooltip(t *testing.T) {
	c := check.New(t)
	c.Equal("pick", labelCellTooltip(&gurps.CellData{UnresolvedChoice: "pick"}))
	c.Equal("pick\n---\nnotes", labelCellTooltip(&gurps.CellData{UnresolvedChoice: "pick", Tooltip: "notes"}))
	c.Equal("pick\n---\nunmet", labelCellTooltip(&gurps.CellData{
		UnresolvedChoice:  "pick",
		UnsatisfiedReason: "unmet",
		Tooltip:           "notes",
	}))
	c.Equal("notes", labelCellTooltip(&gurps.CellData{Tooltip: "notes"}))
}

// TestModifierSelectionTreatsChoicesByKind verifies the prompt asking which modifiers to enable: a modifier outside a
// choice gets a check box, the options of an optional choice get radio buttons along with "None", and those of a
// mandatory choice get radio buttons alone, the prompt not being complete until one is picked.
func TestModifierSelectionTreatsChoicesByKind(t *testing.T) {
	c := check.New(t)
	plain := gurps.NewTraitModifier(nil, nil, false)
	plain.Name = "Plain"
	mandatory := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"})
	optional := newTraitModifierChoiceFor(nil, false, []string{"Hot", "Cold"}, "Hot", "Cold")
	s := newModifierSelection([]*gurps.TraitModifier{plain, mandatory, optional}, true)
	c.NotNil(s)
	c.Equal(1, len(s.boxes), "only the modifier outside a choice gets a check box")
	c.Equal(2, len(s.choices))
	c.True(s.choices[0].mandatory)
	c.False(s.choices[1].mandatory)
	c.Equal(5, len(panelsOfType[*unison.RadioButton](s.list)),
		"each option of a choice gets a radio button, and only the optional choice adds one for None")
	c.False(s.complete(), "a mandatory choice with nothing picked must hold the prompt open")

	radio := func(choice *choiceRadioGroup, m *gurps.TraitModifier) *unison.RadioButton {
		for rb, gm := range choice.options {
			if gm == m {
				return rb
			}
		}
		t.Fatalf("no radio button for %s", m.Name)
		return nil
	}
	c.True(s.choices[1].group.Selected(radio(s.choices[1], optional.Children[0])),
		"the first enabled option starts out picked")
	radio(s.choices[0], mandatory.Children[1]).Click()
	c.True(s.complete(), "picking an option makes the mandatory choice")

	for cb := range s.boxes {
		cb.State = checkstate.Off
	}
	c.True(s.apply())
	c.False(plain.Enabled())
	c.False(mandatory.Children[0].Enabled())
	c.True(mandatory.Children[1].Enabled())
	c.True(optional.Children[0].Enabled())
	c.False(optional.Children[1].Enabled(), "only one option of a choice may be left on")
}

// TestTraitEditorShowsTheRangeOfAnOpenChoice verifies that a trait editor opened outside a sheet shows the range of
// point costs while a mandatory modifier choice is yet to be made, as the list does, and a single cost once the trait
// is marked preconfigured with its pick made, the editor's own pending state counting in both.
func TestTraitEditorShowsTheRangeOfAnOpenChoice(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.BasePoints = fxp.FromInteger(10)
	choice := newTraitModifierChoiceFor(nil, true, []string{"Small", "Large"})
	choice.Children[0].CostAdj = "+5"
	choice.Children[1].CostAdj = "+10"
	trait.AddModifiers(choice)
	e, content := buildEditorContent(nil, trait, initTraitEditor)
	// The Point Cost field is the first of the editor's non-editable fields.
	fields := panelsOfType[*NonEditableField](content)
	c.NotEqual(0, len(fields))
	pointCost := func() string { return fields[0].String() }
	c.Equal("15~20", pointCost())
	e.editorData.Preconfigured = true
	DeepSync(e)
	c.Equal("15~20", pointCost(), "a preconfigured trait still has a choice with no pick to make")
	e.editorData.Modifiers[0].Children[1].SetEnabled(true)
	DeepSync(e)
	c.Equal("20", pointCost(), "a preconfigured trait takes the pick already made")
}

// TestLockedPickKeepsItsCheckmark verifies that clicking the checkmark of the pick of a mandatory choice on a sheet
// leaves both the modifier and the checkmark drawn for it as they were, rather than clearing the mark over a pick that
// is still in force.
func TestLockedPickKeepsItsCheckmark(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A")}
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := firstPanelOfType[*traitModifiersPanel](content)
	c.True(ok)
	table := panel.table
	pick := table.RootRows()[0].Children()[0]
	label, ok := pick.ColumnCell(0, 0, unison.Black, unison.White, false, false, false).(*unison.Label)
	c.True(ok, "the enabled cell must be a label")
	c.NotNil(label.Drawable, "the pick starts out checked")
	table.AddChild(label)
	c.True(label.MouseDownCallback(geom.Point{}, unison.ButtonLeft, 1, mod.None), "the click must be consumed")
	label.RemoveFromParent()
	c.NotNil(label.Drawable, "the checkmark must stay, since the pick can't be turned off")
	c.True(e.editorData.Modifiers[0].Children[0].Enabled())
	c.False(unison.UndoManagerFor(table).CanUndo(), "a refused click records nothing")
}

// TestModifierEditorFollowsTheChoiceRules verifies that enabling an option in its own editor makes it the pick,
// turning the one that was picked off, and that undo puts both back. On a sheet, the pick of a mandatory choice can't
// be turned off there either.
func TestModifierEditorFollowsTheChoiceRules(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	trait := gurps.NewTrait(entity, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A")}
	traitEditor, _ := buildEditorContent(sheet, trait, initTraitEditor)
	options := traitEditor.editorData.Modifiers[0].Children

	e, _ := buildEditorContent(traitEditor, options[1], initTraitModifierEditor)
	e.editorData.Disabled = false
	e.applyEdits()
	c.False(options[0].Enabled(), "enabling B in its editor turns A off")
	c.True(options[1].Enabled())
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.True(options[0].Enabled(), "undo turns A back on")
	c.False(options[1].Enabled(), "and B back off")
	mgr.Redo()
	c.False(options[0].Enabled(), "redo turns A off again")
	c.True(options[1].Enabled(), "and B back on")
	mgr.Undo()

	e, _ = buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	e.editorData.Disabled = true
	e.applyEdits()
	c.True(options[0].Enabled(), "the pick of a mandatory choice on a sheet can't be turned off in its editor")
}

// TestModifierSelectionOutsideASheet verifies that when the rows aren't headed for a sheet, a mandatory choice is
// offered just as an optional one is, "None" included, and doesn't hold the prompt open, and that a preconfigured row
// isn't asked about at all.
func TestModifierSelectionOutsideASheet(t *testing.T) {
	c := check.New(t)
	mandatory := newTraitModifierChoiceFor(nil, true, []string{"Low", "High"})
	s := newModifierSelection([]*gurps.TraitModifier{mandatory}, false)
	c.True(s.complete(), "a template may keep a mandatory choice without its pick")
	c.Equal(3, len(panelsOfType[*unison.RadioButton](s.list)), "the options and None")
	c.Nil(s.choices[0].status)

	prompts := captureModifierPrompts(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{mandatory}
	trait.Preconfigured = true
	c.True(processModifiers(promptOperation{}, []*gurps.Trait{trait}, false))
	c.Equal(0, len(*prompts), "a preconfigured row headed for a template isn't asked about its choices")
}

// TestEmptyMandatoryChoiceDoesNotHoldThePromptOpen verifies that a mandatory choice with no options, which has
// nothing to pick from, neither holds the prompt open nor is flagged as required.
func TestEmptyMandatoryChoiceDoesNotHoldThePromptOpen(t *testing.T) {
	c := check.New(t)
	empty := gurps.NewTraitModifierChoice(nil, nil)
	s := newModifierSelection([]*gurps.TraitModifier{empty}, true)
	c.NotNil(s)
	c.True(s.complete())
	c.Nil(s.choices[0].status, "nothing is flagged")
}

// TestApplyingAChoiceToLootAsksForItsPick verifies that applying a container of modifiers to equipment on a loot sheet
// asks which of its modifiers to enable, a pick required, just as on a character sheet, since both are sheets.
func TestApplyingAChoiceToLootAsksForItsPick(t *testing.T) {
	c := check.New(t)
	sheet := newTestLootSheet(t)
	eqp := gurps.NewEquipment(sheet.loot, nil, false)
	eqp.Name = "Sword"
	sheet.loot.Equipment = []*gurps.Equipment{eqp}
	sheet.Rebuild(true)
	var asked []bool
	swapForTest(t, &promptForEquipmentModifiers,
		func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
			asked = append(asked, info.requirePicks)
			return false, false
		})
	choice := gurps.NewEquipmentModifierChoice(nil, nil)
	option := gurps.NewEquipmentModifier(nil, choice, false)
	choice.Children = []*gurps.EquipmentModifier{option}
	attachModifierClones("", []*unison.Table[*Node[*gurps.Equipment]]{sheet.Equipment.Table}, sheet.loot,
		[]*gurps.Equipment{eqp}, []*gurps.EquipmentModifier{choice}, gurps.LibraryFile{})
	c.Equal([]bool{true}, asked, "the loot sheet asks, requiring the pick")
}

// traitEditorOnSheet returns a trait editor on a sheet for a trait holding the given modifiers, along with its table of
// modifiers.
func traitEditorOnSheet(t *testing.T, sheet *Sheet, modifiers ...*gurps.TraitModifier) (*editor[*gurps.Trait, *gurps.TraitEditData],
	*unison.Table[*Node[*gurps.TraitModifier]],
) {
	t.Helper()
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Modifiers = modifiers
	e, content := buildEditorContent(sheet, trait, initTraitEditor)
	panel, ok := firstPanelOfType[*traitModifiersPanel](content)
	if !ok {
		t.Fatal("expected a trait modifiers panel in the trait editor")
	}
	return e, panel.table
}

// TestMovingAGroupIntoAChoiceKeepsThePick verifies that a group moved into a choice brings its options in turned off,
// so that the choice keeps the pick it had, even though they land ahead of it.
func TestMovingAGroupIntoAChoiceKeepsThePick(t *testing.T) {
	c := check.New(t)
	registerKeyBindingsOnce.Do(registerActions)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	group := gurps.NewTraitModifier(entity, nil, true)
	carried := gurps.NewTraitModifier(entity, group, false)
	group.Children = []*gurps.TraitModifier{carried}
	e, table := traitEditorOnSheet(t, sheet, group, newTraitModifierChoiceFor(entity, false, []string{"X", "Y"}, "X"))
	groupCopy := e.editorData.Modifiers[0]
	table.SetSelectionMap(map[tid.TID]bool{groupCopy.ID(): true})
	c.True(table.CanPerformCmd(nil, MoveIntoContainerItemID))
	table.PerformCmd(nil, MoveIntoContainerItemID)
	c.Equal(1, len(e.editorData.Modifiers), "the group went into the choice")
	options := gurps.ModifierChoiceOptions(e.editorData.Modifiers[0])
	c.Equal(3, len(options), "the option the group carried, then X and Y")
	c.False(options[0].Enabled(), "what the group carried in arrives turned off")
	c.True(options[1].Enabled(), "the choice keeps its pick, X")
	c.False(options[2].Enabled())
}

// TestDropIntoAClosedChoiceKeepsTheDroppedRowsSelected verifies that rows dropped into a closed modifier container
// are shown, the container being opened, so that they stay selected, which is what tells the choice which of its
// options just arrived.
func TestDropIntoAClosedChoiceKeepsTheDroppedRowsSelected(t *testing.T) {
	c := check.New(t)
	choice := newTraitModifierChoiceFor(nil, false, []string{"X", "Arrived"}, "X")
	choice.SetOpen(false)
	library := NewTraitModifierTableDockable("mods"+gurps.TraitModifiersExt, []*gurps.TraitModifier{choice})
	table := library.table
	table.SetSelectionMap(map[tid.TID]bool{choice.Children[1].ID(): true})
	revealModifierDropTarget(table, table.RootRows()[0])
	c.True(choice.IsOpen(), "the choice is opened")
	selected := table.SelectedRows(false)
	c.Equal(1, len(selected))
	c.Equal(choice.Children[1], selected[0].Data(), "the dropped row is still selected")
}

// TestEditorKeepsAPickMadeSinceItOpened verifies that applying an option's editor, opened while it was the pick, after
// another option has been picked in the list leaves that later pick alone, since the editor didn't change whether its
// option is enabled.
func TestEditorKeepsAPickMadeSinceItOpened(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traitEditor, table := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(entity, true, []string{"A", "B"}, "A"))
	options := traitEditor.editorData.Modifiers[0].Children
	aEditor, _ := buildEditorContent(traitEditor, options[0], initTraitModifierEditor)
	adjustModifierEnabled(traitEditor, table, options[1], true)
	c.True(options[1].Enabled())
	aEditor.editorData.Name = "Renamed"
	aEditor.applyEdits()
	c.Equal("Renamed", options[0].Name)
	c.False(options[0].Enabled(), "the editor didn't change whether A is enabled, so A stays off")
	c.True(options[1].Enabled(), "and B stays the pick")
}

// TestChoiceMadeMandatoryInItsEditorOnASheet verifies that a choice made mandatory in its editor on a sheet, with
// nothing picked, gets its first option picked, as Convert to Choice does, and that undo and redo cover it.
func TestChoiceMadeMandatoryInItsEditorOnASheet(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	traitEditor, _ := traitEditorOnSheet(t, sheet, newTraitModifierChoiceFor(entity, false, []string{"A", "B"}))
	choice := traitEditor.editorData.Modifiers[0]
	choiceEditor, _ := buildEditorContent(traitEditor, choice, initTraitModifierEditor)
	choiceEditor.editorData.SetMandatoryChoice(true)
	choiceEditor.applyEdits()
	c.True(gurps.IsMandatoryModifierChoice(choice))
	c.True(choice.Children[0].Enabled(), "the first option is picked")
	mgr := unison.UndoManagerFor(traitEditor)
	mgr.Undo()
	c.False(gurps.IsMandatoryModifierChoice(choice))
	c.False(choice.Children[0].Enabled(), "undo takes the pick back")
	mgr.Redo()
	c.True(choice.Children[0].Enabled(), "redo picks it again")
}

// TestEquipmentChoiceToggleAndLock verifies the choice rules in an equipment editor on a sheet: turning an option on
// turns the pick off, and the new pick can't be turned off.
func TestEquipmentChoiceToggleAndLock(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	choice := gurps.NewEquipmentModifierChoice(entity, nil)
	for _, name := range []string{"A", "B"} {
		option := gurps.NewEquipmentModifier(entity, choice, false)
		option.Name = name
		option.SetEnabled(name == "A")
		choice.Children = append(choice.Children, option)
	}
	equipment := gurps.NewEquipment(entity, nil, false)
	equipment.Modifiers = []*gurps.EquipmentModifier{choice}
	e, content := buildEditorContent(sheet, equipment, initEquipmentEditor(true))
	panel, ok := firstPanelOfType[*equipmentModifiersPanel](content)
	c.True(ok)
	table := panel.table
	options := e.editorData.Modifiers[0].Children
	table.SetSelectionMap(map[tid.TID]bool{options[1].ID(): true})
	table.PerformCmd(nil, ToggleStateItemID)
	c.False(options[0].Enabled(), "turning B on turns A off")
	c.True(options[1].Enabled())
	c.False(table.CanPerformCmd(nil, ToggleStateItemID), "the new pick can't be turned off on a sheet")
}

// TestPromptRequiresPicksOnlyForSheets verifies that rows headed for a template are asked about their modifiers
// without a pick being required, and rows headed for a loot sheet with one.
func TestPromptRequiresPicksOnlyForSheets(t *testing.T) {
	c := check.New(t)
	var traitAsked, equipmentAsked []bool
	swapForTest(t, &promptForTraitModifiers,
		func(info *modifierPromptInfo, _ []*gurps.TraitModifier) (changed, canceled bool) {
			traitAsked = append(traitAsked, info.requirePicks)
			return false, false
		})
	swapForTest(t, &promptForEquipmentModifiers,
		func(info *modifierPromptInfo, _ []*gurps.EquipmentModifier) (changed, canceled bool) {
			equipmentAsked = append(equipmentAsked, info.requirePicks)
			return false, false
		})
	trait := gurps.NewTrait(nil, nil, false)
	trait.Modifiers = []*gurps.TraitModifier{newTraitModifierChoiceFor(nil, true, []string{"A"})}
	template := newTestTemplateWithTraits()
	toTemplate := &applyPart[*gurps.Trait]{table: template.Traits.Table, rows: []*gurps.Trait{trait}}
	c.True(toTemplate.promptForModifiers(promptOperation{}, 0, 1))
	c.Equal([]bool{false}, traitAsked, "a template doesn't require the pick")

	loot := newTestLootSheet(t)
	equipment := gurps.NewEquipment(nil, nil, false)
	equipment.Modifiers = []*gurps.EquipmentModifier{gurps.NewEquipmentModifierChoice(nil, nil)}
	toLoot := &applyPart[*gurps.Equipment]{table: loot.Equipment.Table, rows: []*gurps.Equipment{equipment}}
	c.True(toLoot.promptForModifiers(promptOperation{}, 0, 1))
	c.Equal([]bool{true}, equipmentAsked, "a loot sheet requires it")
}
