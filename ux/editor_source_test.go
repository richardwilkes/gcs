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
	"path/filepath"
	"slices"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// testSource is a complete source that no library holds.
var testSource = gurps.Source{
	Library: "Test Library",
	Path:    "Traits/Test.adq",
	TID:     "t0123456789abcdef",
}

// sourceMenuSummary returns the entries' labels, with "# " before a heading and " (disabled)" after a disabled entry.
func sourceMenuSummary(entries []menuEntry) []string {
	summary := make([]string, 0, len(entries))
	for _, one := range entries {
		switch {
		case one.Act == nil:
			summary = append(summary, "# "+one.Label)
		case one.Disabled:
			summary = append(summary, one.Label+" (disabled)")
		default:
			summary = append(summary, one.Label)
		}
	}
	return summary
}

// sourceMenuAction returns the action of the enabled entry with the label, or nil if there is no such entry.
func sourceMenuAction(entries []menuEntry, label string) func() {
	for _, one := range entries {
		if one.Label == label && !one.Disabled {
			return one.Act
		}
	}
	return nil
}

// requireSourceMenuAction returns the action of the enabled entry with the label, ending the test with why if there is
// none, since calling a nil action would bring down every test in the package.
func requireSourceMenuAction(t *testing.T, entries []menuEntry, label, why string) func() {
	t.Helper()
	action := sourceMenuAction(entries, label)
	if action == nil {
		t.Fatalf("no enabled %q in the source menu: %s", label, why)
	}
	return action
}

// newLibrarySourcedTrait saves lib into a file of the user library and returns a sheet holding a copy of it sourced
// from that file.
func newLibrarySourcedTrait(t *testing.T, c check.Checker, lib *gurps.Trait) (*Sheet, *gurps.Trait) {
	t.Helper()
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{lib}, filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	local := lib.Clone(libFile, entity, nil, gurps.Reference)
	entity.Traits = append(entity.Traits, local)
	entity.Recalculate()
	return sheet, local
}

// TestEditorsShowNoSourceFields verifies that no editor shows the ID or the source in its content, which the toolbar's
// source menu shows instead.
func TestEditorsShowNoSourceFields(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	contents := make(map[string]*unison.Panel)
	sourced := func(node interface{ SetSource(gurps.Source) }) { node.SetSource(testSource) }

	trait := gurps.NewTrait(entity, nil, false)
	sourced(trait)
	_, contents["trait"] = buildEditorContent(sheet, trait, initTraitEditor)
	traitContainer := gurps.NewTrait(entity, nil, true)
	sourced(traitContainer)
	_, contents["trait container"] = buildEditorContent(sheet, traitContainer, initTraitEditor)
	_, contents["trait choice"] = buildEditorContent(sheet, gurps.NewTraitChoiceContainer(entity, nil), initTraitEditor)
	skill := gurps.NewSkill(entity, nil, false)
	sourced(skill)
	_, contents["skill"] = buildEditorContent(sheet, skill, initSkillEditor)
	spell := gurps.NewSpell(entity, nil, false)
	sourced(spell)
	_, contents["spell"] = buildEditorContent(sheet, spell, initSpellEditor)
	note := gurps.NewNote(entity, nil, false)
	sourced(note)
	_, contents["note"] = buildEditorContent(sheet, note, initNoteEditor)
	eqp := gurps.NewEquipment(entity, nil, false)
	sourced(eqp)
	_, contents["equipment"] = buildEditorContent(sheet, eqp, initEquipmentEditor(true))
	group := gurps.NewEquipment(entity, nil, true)
	group.ContainerType = eqcontainer.Group
	sourced(group)
	_, contents["equipment group"] = buildEditorContent(sheet, group, initEquipmentEditor(true))
	_, contents["equipment choice"] = buildEditorContent(sheet, gurps.NewEquipmentChoiceContainer(entity, nil),
		initEquipmentEditor(false))
	traitMod := gurps.NewTraitModifier(entity, nil, false)
	sourced(traitMod)
	_, contents["trait modifier"] = buildEditorContent(sheet, traitMod, initTraitModifierEditor)
	traitModGroup := gurps.NewTraitModifier(entity, nil, true)
	sourced(traitModGroup)
	_, contents["trait modifier group"] = buildEditorContent(sheet, traitModGroup, initTraitModifierEditor)
	eqpMod := gurps.NewEquipmentModifier(entity, nil, false)
	sourced(eqpMod)
	_, contents["equipment modifier"] = buildEditorContent(sheet, eqpMod, initEquipmentModifierEditor)
	eqpModGroup := gurps.NewEquipmentModifier(entity, nil, true)
	sourced(eqpModGroup)
	_, contents["equipment modifier group"] = buildEditorContent(sheet, eqpModGroup, initEquipmentModifierEditor)

	for name, content := range contents {
		texts := labelTexts(content)
		c.NotEqual(0, len(texts), name+" editor must have built its content")
		for _, label := range []string{"ID", "Source ID", "Source Library", "Source Path"} {
			c.False(slices.Contains(texts, label), name+" editor must not show "+label)
		}
	}
}

// TestSourceMenuEntries verifies what the source menu offers for the states a target's source may be in.
func TestSourceMenuEntries(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	custom := gurps.NewTrait(entity, nil, false)
	e, _ := buildEditorContent(sheet, custom, initTraitEditor)
	c.Equal([]string{
		"ID: " + string(custom.ID()),
		"# " + srcstate.Custom.String(),
		"Sync with Source (disabled)",
		"Clear Source (disabled)",
	}, sourceMenuSummary(e.sourceMenuEntries()), "custom data has nothing to sync with or clear")

	partial := gurps.NewTrait(entity, nil, false)
	partial.Source = gurps.Source{Library: "Test Library"}
	e, _ = buildEditorContent(sheet, partial, initTraitEditor)
	c.Equal([]string{
		"ID: " + string(partial.ID()),
		"Source Library: Test Library",
		"# " + srcstate.Custom.String(),
		"Sync with Source (disabled)",
		"Clear Source",
	}, sourceMenuSummary(e.sourceMenuEntries()), "only the parts present are shown, and they can still be cleared")

	missing := gurps.NewTrait(entity, nil, false)
	missing.Source = testSource
	entity.Traits = append(entity.Traits, missing)
	entity.Recalculate()
	e, _ = buildEditorContent(sheet, missing, initTraitEditor)
	c.Equal([]string{
		"ID: " + string(missing.ID()),
		"Source ID: " + string(testSource.TID),
		"Source Library: " + testSource.Library,
		"Source Path: " + testSource.Path,
		"# " + srcstate.Missing.String(),
		"Sync with Source (disabled)",
		"Clear Source",
	}, sourceMenuSummary(e.sourceMenuEntries()), "a source no library holds can't be synced with")

	choice := gurps.NewTraitChoiceContainer(entity, nil)
	choice.Source = testSource
	e, _ = buildEditorContent(sheet, choice, initTraitEditor)
	c.Equal([]string{"ID: " + string(choice.ID())}, sourceMenuSummary(e.sourceMenuEntries()),
		"a template choice container shows only its ID")
}

// TestSourceMenuMatchesWhatNoSheetPrepared verifies that the source menu compares a node with its library copy even
// when nothing has prepared the library file it comes from: a modifier dropped into an open editor, which isn't on its
// sheet yet, or a node of a library list file, which has no data owner, or one without a source matcher.
func TestSourceMenuMatchesWhatNoSheetPrepared(t *testing.T) {
	c := check.New(t)
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	libMod := gurps.NewTraitModifier(nil, nil, false)
	libMod.Name = "Long"
	modFile := gurps.LibraryFile{Library: user.Key(), Path: "Modifiers/Test" + gurps.TraitModifiersExt}
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{libMod},
		filepath.Join(user.Path(), filepath.FromSlash(modFile.Path))))
	libTrait := gurps.NewTrait(nil, nil, false)
	libTrait.Name = "Claws"
	traitFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{libTrait}, filepath.Join(user.Path(), filepath.FromSlash(traitFile.Path))))
	libEqp := gurps.NewEquipment(nil, nil, false)
	libEqp.Name = "Rope"
	eqpFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{libEqp},
		filepath.Join(user.Path(), filepath.FromSlash(eqpFile.Path))))
	sheet := newTestSheetForTemplate(t)
	sheet.Entity().Recalculate()

	dropped := libMod.Clone(modFile, sheet.Entity(), nil, gurps.Reference)
	e, _ := buildEditorContent(sheet, dropped, initTraitModifierEditor)
	c.True(slices.Contains(sourceMenuSummary(e.sourceMenuEntries()), "# "+srcstate.Matched.String()),
		"a modifier from a file nothing on the sheet comes from is matched")
	e.editorData.Name = "Short"
	c.True(slices.Contains(sourceMenuSummary(e.sourceMenuEntries()), "Sync with Source"),
		"and can be synced once it differs")

	listed := libTrait.Clone(traitFile, nil, nil, gurps.Reference)
	listed.Name = "Talons"
	traitEditor, _ := buildEditorContent(sheet, listed, initTraitEditor)
	c.True(slices.Contains(sourceMenuSummary(traitEditor.sourceMenuEntries()), "# "+srcstate.Mismatched.String()),
		"a node without a data owner, as in a library list file, is matched")
	requireSourceMenuAction(t, traitEditor.sourceMenuEntries(), "Sync with Source",
		"a node without a data owner can be synced")()
	c.Equal("Claws", traitEditor.editorData.Name, "and synced")

	listedEqp := libEqp.Clone(eqpFile, &equipmentListProvider{}, nil, gurps.Reference)
	listedEqp.Name = "Cord"
	eqpEditor, _ := buildEditorContent(sheet, listedEqp, initEquipmentEditor(true))
	c.True(slices.Contains(sourceMenuSummary(eqpEditor.sourceMenuEntries()), "# "+srcstate.Mismatched.String()),
		"as is one whose data owner, an equipment list file, has no source matcher")
	requireSourceMenuAction(t, eqpEditor.sourceMenuEntries(), "Sync with Source",
		"equipment of an equipment list file can be synced")()
	c.Equal("Rope", eqpEditor.editorData.Name, "and synced")
	c.Equal("Cord", listedEqp.Name, "the equipment is left alone until the changes are applied")
	listedEqp.SyncWithSource()
	c.Equal("Rope", listedEqp.Name, "which the list's Sync with Source syncs as well")
}

// TestEditorClearSourceIsPendingAndUndoable verifies that clearing the source from the editor's menu leaves the target
// alone until the editor's changes are applied, can be undone within the editor, and is then applied, undone and redone
// along with the editor's other changes as one edit.
func TestEditorClearSourceIsPendingAndUndoable(t *testing.T) {
	c := check.New(t)
	trait := gurps.NewTrait(nil, nil, false)
	trait.Name = "Claws"
	trait.Source = testSource
	data := gurps.NewTemplate()
	data.Traits = []*gurps.Trait{trait}
	template := newTestTemplateDockable("Source", data)
	mgr := unison.UndoManagerFor(template)
	if mgr == nil {
		t.Fatal("the template must have an undo manager")
	}

	e, _ := buildEditorContent(template, trait, initTraitEditor)
	c.False(e.isModified())
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Clear Source", "a sourced trait's source can be cleared")()
	c.True(e.isModified(), "a pending clear is a change to apply")
	c.Equal(testSource, trait.Source, "the target keeps its source until the change is applied")
	c.Equal([]string{
		"ID: " + string(trait.ID()),
		"# " + srcstate.Custom.String(),
		"Sync with Source (disabled)",
		"Clear Source (disabled)",
	}, sourceMenuSummary(e.sourceMenuEntries()), "the menu describes the target as it will be")

	e.UndoManager().Undo()
	c.False(e.isModified(), "undoing the clear within the editor takes it back")
	e.UndoManager().Redo()
	c.True(e.isModified(), "and redoing it puts it back")

	e.editorData.Name = "Talons"
	e.applyEdits()
	c.Equal("Talons", trait.Name)
	c.Equal(gurps.Source{}, trait.Source, "applying clears the source")
	mgr.Undo()
	c.Equal("Claws", trait.Name)
	c.Equal(testSource, trait.Source, "undo puts the source back along with the rest")
	mgr.Redo()
	c.Equal("Talons", trait.Name)
	c.Equal(gurps.Source{}, trait.Source, "redo clears it again")
}

// TestEditorSyncWithSource verifies that syncing from the editor's menu replaces the synced fields of the editor's data
// with the library's, keeps the editor's other pending changes, rebuilds the content around the new data and leaves the
// target alone until the changes are applied.
func TestEditorSyncWithSource(t *testing.T) {
	c := check.New(t)
	lib := gurps.NewTrait(nil, nil, false)
	lib.Name = "Claws"
	lib.PageRef = "B42"
	lib.BasePoints = fxp.Five
	lib.Weapons = []*gurps.Weapon{gurps.NewWeapon(lib, true)}
	sheet, local := newLibrarySourcedTrait(t, c, lib)
	local.Name = "Claws (old)"
	local.BasePoints = fxp.Three
	local.Weapons = nil
	local.UserDesc = "Mine"
	modifier := gurps.NewTraitModifier(local.DataOwner(), nil, false)
	modifier.Name = "Long"
	local.Modifiers = []*gurps.TraitModifier{modifier}
	gurps.AttachModifiers(local, local.Modifiers)

	e, content := buildEditorContent(sheet, local, initTraitEditor)
	e.editorData.UserDesc = "Pending"
	e.editorData.PageRef = "B99"
	e.clearSource()
	e.UndoManager().Undo()
	c.True(e.UndoManager().CanRedo(), "precondition: the editor has an undo history")
	oldChildren := slices.Clone(content.Children())
	oldMelee := e.meleeWeapons
	entries := e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Mismatched.String()))
	requireSourceMenuAction(t, entries, "Sync with Source", "a trait that differs from its source can be synced")()

	c.Equal("Claws", e.editorData.Name, "synced fields come from the library")
	c.Equal("B42", e.editorData.PageRef, "including those with pending changes")
	c.Equal(fxp.Five, e.editorData.BasePoints)
	c.Equal("Pending", e.editorData.UserDesc, "pending changes to fields a sync leaves alone are kept")
	c.Equal(1, len(e.editorData.Modifiers), "the modifiers aren't synced")
	c.True(e.editorData.Modifiers[0].Target() == local, "the modifiers stay pointed at the target")
	c.Equal(1, len(e.editorData.Weapons), "the library's weapon arrives")
	c.True(e.editorData.Weapons[0].Owner == gurps.WeaponOwner(local), "pointed at the target")
	c.Equal("Claws (old)", local.Name, "the target is left alone until the changes are applied")
	c.True(e.isModified())
	c.False(e.UndoManager().CanUndo() || e.UndoManager().CanRedo(), "the editor's undo history is cleared")
	for _, child := range oldChildren {
		c.Nil(child.Parent(), "the old content is thrown away")
	}
	c.NotEqual(0, len(content.Children()), "and new content built")
	c.True(e.meleeWeapons != nil && e.meleeWeapons != oldMelee, "the weapons panel is rebuilt")
	c.Equal(1, len(e.meleeWeapons.Weapons(true, false, false)), "and shows the library's weapon")
	entries = e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Matched.String()),
		"the editor's data now matches the library")
	c.Nil(sourceMenuAction(entries, "Sync with Source"), "so there is nothing left to sync")

	e.applyEdits()
	c.Equal("Claws", local.Name)
	c.Equal("Pending", local.UserDesc)
	state, _ := gurps.MatchSource(local)
	c.Equal(srcstate.Matched, state, "once applied, the target matches its source")
}

// TestEditorSyncThatOnlyRevertsPendingEdits verifies that syncing an editor whose only difference from the library is
// a pending change leaves it with nothing to apply.
func TestEditorSyncThatOnlyRevertsPendingEdits(t *testing.T) {
	c := check.New(t)
	lib := gurps.NewTrait(nil, nil, false)
	lib.Name = "Claws"
	lib.PageRef = "B42"
	sheet, local := newLibrarySourcedTrait(t, c, lib)
	state, _ := gurps.MatchSource(local)
	c.Equal(srcstate.Matched, state, "precondition: the copy matches its source")

	e, _ := buildEditorContent(sheet, local, initTraitEditor)
	e.editorData.PageRef = "B99"
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source",
		"the pending change makes the editor's data differ from the library")()
	c.Equal("B42", e.editorData.PageRef)
	c.False(e.isModified(), "nothing is left to apply")
}

// TestEditorSyncLeavesEquipmentKindChangesToTheList verifies that the editor won't sync equipment with a source of
// another kind, whether another kind of container or a container when it isn't one, since its data can't carry the
// kind.
func TestEditorSyncLeavesEquipmentKindChangesToTheList(t *testing.T) {
	c := check.New(t)
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	lib := gurps.NewEquipment(nil, nil, true)
	lib.ContainerType = eqcontainer.Group
	lib.Name = "Pack"
	libContainer := gurps.NewEquipment(nil, nil, true)
	libContainer.Name = "Sack"
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{lib, libContainer},
		filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	local := lib.Clone(libFile, entity, nil, gurps.Reference)
	local.ContainerType = eqcontainer.Container
	local.BaseValue = "10"
	local.Modifiers = []*gurps.EquipmentModifier{gurps.NewEquipmentModifier(entity, nil, false)}
	converted := libContainer.Clone(libFile, entity, nil, gurps.Reference)
	converted.ConvertToNonContainer()
	converted.BaseValue = "5"
	entity.CarriedEquipment = append(entity.CarriedEquipment, local, converted)
	entity.Recalculate()

	e, _ := buildEditorContent(sheet, local, initEquipmentEditor(true))
	entries := e.sourceMenuEntries()
	summary := sourceMenuSummary(entries)
	c.True(slices.Contains(summary, "# "+srcstate.Mismatched.String()), "the container differs from its source")
	c.True(slices.Contains(summary, "Sync with Source (disabled)"), "but only the list can change its kind")
	e.syncWithSource()
	c.Equal("10", e.editorData.BaseValue, "nothing is synced")
	c.Equal(1, len(e.editorData.Modifiers))
	c.False(e.isModified())

	e, _ = buildEditorContent(sheet, converted, initEquipmentEditor(true))
	summary = sourceMenuSummary(e.sourceMenuEntries())
	c.True(slices.Contains(summary, "# "+srcstate.Mismatched.String()), "the converted item differs from its source")
	c.True(slices.Contains(summary, "Sync with Source (disabled)"),
		"but its source is a container, which only the list can make it again")
	e.syncWithSource()
	c.Equal("5", e.editorData.BaseValue, "nothing is synced")
	c.False(e.isModified())
}

// TestEditorSyncThatUnmakesAModifierChoice verifies that applying an editor's sync that makes a modifier choice a
// group, as its library copy is, hands its options to the choice around it, which keeps its own pick, as the list's
// sync does, and that undo puts the options back.
func TestEditorSyncThatUnmakesAModifierChoice(t *testing.T) {
	c := check.New(t)
	sheet, outer, inner := newChoiceWithinChoice(t, c, true)
	options := gurps.ModifierChoiceOptions(outer)
	c.Equal(1, len(options), "precondition: the inner choice owns its own options")

	e, _ := buildEditorContent(sheet, inner, initTraitModifierEditor)
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source",
		"the choice differs from its library copy, a group")()
	e.applyEdits()
	c.False(gurps.IsModifierChoice(inner), "the choice is a group now")
	a, b, cMod := outer.Children[0], inner.Children[0], inner.Children[1]
	c.True(a.Enabled(), "the outer choice keeps its own pick")
	c.False(b.Enabled(), "over the option the group handed it")
	c.False(cMod.Enabled())

	unison.UndoManagerFor(sheet).Undo()
	c.True(gurps.IsModifierChoice(inner))
	c.True(a.Enabled())
	c.True(b.Enabled(), "undo puts the inner choice's pick back")
	c.False(cMod.Enabled())
}

// TestEditorSyncThatMakesAModifierChoice verifies that applying an editor's sync that makes a modifier group a choice,
// as its library copy is, settles it and the choice around it as the list's sync does: the new choice keeps the pick it
// was handed, and the mandatory choice around it, which lost its pick to it, is left for the user to make rather than
// given one. Undo makes it a group again, with the options as they were.
func TestEditorSyncThatMakesAModifierChoice(t *testing.T) {
	c := check.New(t)
	sheet, outer, inner := newChoiceWithinChoice(t, c, false)
	a, b, cMod := outer.Children[0], inner.Children[0], inner.Children[1]
	c.False(a.Enabled(), "precondition: the outer choice's pick is within the group")
	c.True(b.Enabled())

	e, _ := buildEditorContent(sheet, inner, initTraitModifierEditor)
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source",
		"the group differs from its library copy, a choice")()
	e.applyEdits()
	c.True(gurps.IsModifierChoice(inner), "the group is a choice now")
	c.True(b.Enabled(), "it keeps the pick it was handed")
	c.False(cMod.Enabled())
	c.False(a.Enabled(), "and the mandatory choice around it, which lost its pick to it, isn't given one")
	c.False(gurps.ModifierChoiceIsResolved(outer), "but is left for the user to make")

	unison.UndoManagerFor(sheet).Undo()
	c.False(gurps.IsModifierChoice(inner), "undo makes it a group again")
	c.True(gurps.ModifierChoiceIsResolved(outer), "whose option is the outer choice's pick once more")
	c.False(a.Enabled())
	c.True(b.Enabled())
	c.False(cMod.Enabled())
}

// TestEditorSyncSettlesAChoiceAsItsListDoes verifies that syncing a modifier container from its editor leaves its
// options, and those of the choice around it, just as syncing it from its list does, whatever the sync makes of it: a
// choice of a group, a group of a choice, or a mandatory choice of an optional one.
func TestEditorSyncSettlesAChoiceAsItsListDoes(t *testing.T) {
	// optionalWithoutPick has the inner choice optional, with nothing picked, and its library copy a mandatory choice.
	optionalWithoutPick := func(t *testing.T, c check.Checker) (sheet *Sheet, outer, inner *gurps.TraitModifier) {
		sheet, outer, inner = newChoiceWithinChoice(t, c, true)
		inner.SetMandatoryChoice(false)
		inner.Children[0].SetEnabled(false)
		lib := gurps.NewTraitModifierChoice(nil, nil)
		lib.SetMandatoryChoice(true)
		lib.Name = inner.Name
		lib.TID = inner.Source.TID
		c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{lib},
			filepath.Join(gurps.GlobalSettings().Libraries.User().Path(), filepath.FromSlash(inner.Source.Path))))
		sheet.Entity().Recalculate()
		return sheet, outer, inner
	}
	for _, one := range []struct {
		name  string
		build func(t *testing.T, c check.Checker) (sheet *Sheet, outer, inner *gurps.TraitModifier)
	}{
		{
			name: "group to choice",
			build: func(t *testing.T, c check.Checker) (*Sheet, *gurps.TraitModifier, *gurps.TraitModifier) {
				return newChoiceWithinChoice(t, c, false)
			},
		},
		{
			name: "choice to group",
			build: func(t *testing.T, c check.Checker) (*Sheet, *gurps.TraitModifier, *gurps.TraitModifier) {
				return newChoiceWithinChoice(t, c, true)
			},
		},
		{name: "optional to mandatory", build: optionalWithoutPick},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			// outcome describes the container and the options once synced: A, of the outer choice, then B and C.
			outcome := func(outer, inner *gurps.TraitModifier) []bool {
				return []bool{
					gurps.IsModifierChoice(inner), gurps.IsMandatoryModifierChoice(inner),
					outer.Children[0].Enabled(), inner.Children[0].Enabled(), inner.Children[1].Enabled(),
				}
			}
			_, outer, inner := one.build(t, c)
			before := outcome(outer, inner)
			inner.SyncWithSource()
			fromList := outcome(outer, inner)
			c.NotEqual(before, fromList, "precondition: the sync changes the container")

			sheet, outer, inner := one.build(t, c)
			e, _ := buildEditorContent(sheet, inner, initTraitModifierEditor)
			requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source",
				"the container differs from its library copy")()
			e.applyEdits()
			c.Equal(fromList, outcome(outer, inner), "the editor's sync leaves what the list's does")
			state, _ := gurps.MatchSource(inner)
			c.Equal(srcstate.Matched, state)
		})
	}
}

// TestEditorSyncWithSourceForEquipment verifies that equipment of the same kind as its library copy can be synced from
// its editor's menu, which rebuilds the content around the library's data with the modifiers the equipment has and
// the weapons its library copy has, all pointed at the equipment.
func TestEditorSyncWithSourceForEquipment(t *testing.T) {
	c := check.New(t)
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	lib := gurps.NewEquipment(nil, nil, false)
	lib.Name = "Rope"
	lib.BaseValue = "5"
	lib.Weapons = []*gurps.Weapon{gurps.NewWeapon(lib, true)}
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{lib}, filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	local := lib.Clone(libFile, entity, nil, gurps.Reference)
	local.Name = "Cord"
	local.BaseValue = "9"
	local.Weapons = nil
	modifier := gurps.NewEquipmentModifier(entity, nil, false)
	modifier.Name = "Fine"
	local.Modifiers = []*gurps.EquipmentModifier{modifier}
	gurps.AttachModifiers(local, local.Modifiers)
	entity.CarriedEquipment = append(entity.CarriedEquipment, local)
	entity.Recalculate()

	e, content := buildEditorContent(sheet, local, initEquipmentEditor(true))
	e.editorData.Quantity = fxp.Three
	oldChildren := slices.Clone(content.Children())
	entries := e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Mismatched.String()))
	requireSourceMenuAction(t, entries, "Sync with Source",
		"equipment that differs from a source of its own kind can be synced")()

	c.Equal("Rope", e.editorData.Name, "synced fields come from the library")
	c.Equal("5", e.editorData.BaseValue)
	c.Equal(fxp.Three, e.editorData.Quantity, "pending changes to fields a sync leaves alone are kept")
	c.Equal(1, len(e.editorData.Modifiers), "the modifiers aren't synced")
	c.Equal("Fine", e.editorData.Modifiers[0].Name)
	c.True(e.editorData.Modifiers[0].Target() == local, "the modifiers stay pointed at the equipment")
	c.Equal(1, len(e.editorData.Weapons), "the library's weapon arrives")
	c.True(e.editorData.Weapons[0].Owner == gurps.WeaponOwner(local), "pointed at the equipment")
	c.Equal("Cord", local.Name, "the equipment is left alone until the changes are applied")
	for _, child := range oldChildren {
		c.Nil(child.Parent(), "the old content is thrown away")
	}
	c.True(e.meleeWeapons != nil, "the weapons panel is rebuilt")
	c.Equal(1, len(e.meleeWeapons.Weapons(true, false, false)), "and shows the library's weapon")
	entries = e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Matched.String()),
		"the editor's data now matches the library")
	c.Nil(sourceMenuAction(entries, "Sync with Source"), "so there is nothing left to sync")

	e.applyEdits()
	c.Equal("Rope", local.Name)
	c.Equal(fxp.Three, local.Quantity)
	c.Equal(1, len(local.Modifiers))
	state, _ := gurps.MatchSource(local)
	c.Equal(srcstate.Matched, state, "once applied, the equipment matches its source")
}

// TestEditorSyncSurvivesTheContentItRebuilds verifies that the content rebuilt around synced data leaves that data
// matching its source: the trait editor keeps a self-control adjustment synced onto a trait with no roll, and a
// technique matches whether its default is based on an attribute or a skill.
func TestEditorSyncSurvivesTheContentItRebuilds(t *testing.T) {
	c := check.New(t)
	lib := gurps.NewTrait(nil, nil, false)
	lib.Name = "Bad Temper"
	lib.SelfControl = selfctrl.CR12
	lib.SelfControlAdj = selfctrl.ActionPenalty
	sheet, local := newLibrarySourcedTrait(t, c, lib)
	local.SelfControl = selfctrl.None
	local.SelfControlAdj = selfctrl.NoAdjustment
	e, content := buildEditorContent(sheet, local, initTraitEditor)
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source", "the trait lacks its source's adjustment")()
	c.Equal(selfctrl.None, e.editorData.SelfControl, "the roll is the user's to choose, so a sync leaves it alone")
	c.Equal(selfctrl.ActionPenalty, e.editorData.SelfControlAdj, "the adjustment the sync brought in is kept")
	entries := e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Matched.String()),
		"the editor's data matches the library once synced")
	c.Nil(sourceMenuAction(entries, "Sync with Source"), "so there is nothing left to sync")
	e.applyEdits()
	state, _ := gurps.MatchSource(local)
	c.Equal(srcstate.Matched, state, "as does the trait once the changes are applied")

	rolls := panelsOfType[*unison.PopupMenu[selfctrl.Roll]](content)
	if len(rolls) != 1 {
		t.Fatalf("expected the trait editor to have one self-control popup, found %d", len(rolls))
	}
	rolls[0].Select(selfctrl.CR12)
	c.Equal(selfctrl.ActionPenalty, e.editorData.SelfControlAdj, "giving the trait a roll keeps the adjustment")
	rolls[0].Select(selfctrl.None)
	c.Equal(selfctrl.NoAdjustment, e.editorData.SelfControlAdj, "and taking the roll away clears it")

	for _, defaultType := range []string{gurps.DexterityID, gurps.SkillID} {
		_, user := useTestLibraries(t, c)
		libTechnique := gurps.NewTechnique(nil, nil, "Karate")
		libTechnique.Name = "Kicking"
		libTechnique.PageRef = "B230"
		libTechnique.TechniqueDefault.DefaultType = defaultType
		libFile := gurps.LibraryFile{Library: user.Key(), Path: "Skills/Test" + gurps.SkillsExt}
		c.NoError(gurps.SaveSkills([]*gurps.Skill{libTechnique},
			filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
		techniqueSheet := newTestSheetForTemplate(t)
		entity := techniqueSheet.Entity()
		technique := libTechnique.Clone(libFile, entity, nil, gurps.Reference)
		technique.PageRef = "B999"
		entity.Skills = append(entity.Skills, technique)
		entity.Recalculate()
		te, _ := buildEditorContent(techniqueSheet, technique, initSkillEditor)
		requireSourceMenuAction(t, te.sourceMenuEntries(), "Sync with Source",
			"the technique differs from its source")()
		c.Equal("B230", te.editorData.PageRef)
		c.True(slices.Contains(sourceMenuSummary(te.sourceMenuEntries()), "# "+srcstate.Matched.String()),
			"a technique defaulting to "+defaultType+" matches the library once synced")
		te.applyEdits()
		state, _ = gurps.MatchSource(technique)
		c.Equal(srcstate.Matched, state, "as it does once the changes are applied")
	}
}

// TestContentFocusIsFoundAgain verifies that the place the focus held in an editor's content is found again once the
// content has been rebuilt, even if controls have come or gone ahead of it; that the nearest control that can take the
// focus stands in for one that has gone; and that nothing is found in content with nothing to give the focus to.
func TestContentFocusIsFoundAgain(t *testing.T) {
	c := check.New(t)
	sheet := newTestSheetForTemplate(t)
	choice := gurps.NewTraitModifierChoice(sheet.Entity(), nil)
	e, content := buildEditorContent(sheet, choice, initTraitModifierEditor)
	field := func(label string) *unison.Panel {
		t.Helper()
		f := stringFieldLabeled(content, label)
		if f == nil {
			t.Fatalf("the editor has no %s field", label)
		}
		return f.AsPanel()
	}
	popups := panelsOfType[*unison.PopupMenu[string]](content)
	if len(popups) != 1 {
		t.Fatalf("expected the choice's editor to have one popup, found %d", len(popups))
	}
	onTags := newContentFocus(content, field("Tags"))
	onName := newContentFocus(content, field("Name"))
	onPopup := newContentFocus(content, popups[0].AsPanel())
	index := content.IndexOfChild(field("Tags"))
	c.Equal(index, onTags.path[0].index)
	c.Equal("Tags", onTags.path[0].label)

	e.editorData.Choice = gurps.TemplatePicker{}
	e.rebuildContent()
	c.Equal(index-2, content.IndexOfChild(field("Tags")), "precondition: a group's editor has no popup ahead of the tags")
	c.True(content.Children()[index].Self != field("Tags").Self,
		"precondition: another field is where the tags were")
	target, same := onTags.find(content)
	c.True(target == field("Tags"), "the tags field is found where it now is")
	c.True(same)
	target, same = onName.find(content)
	c.True(target == field("Name"), "a field that hasn't moved is found in its place")
	c.True(same)
	target, same = onPopup.find(content)
	c.True(target == field("Tags"), "the field now nearest its place stands in for the popup, which has gone")
	c.False(same)

	onTags = newContentFocus(content, field("Tags"))
	e.editorData.SetMandatoryChoice(true)
	e.rebuildContent()
	target, same = onTags.find(content)
	c.True(target == field("Tags"), "the tags field is found again once the popup is back ahead of it")
	c.True(same)
	field("Tags").SetEnabled(false)
	target, same = onTags.find(content)
	c.False(same, "a field that can no longer take the focus isn't where the focus goes")
	c.True(target != nil && target != field("Tags"), "the nearest control that can take it is")

	content.RemoveAllChildren()
	label := unison.NewLabel()
	label.SetTitle("Value")
	content.AddChild(label)
	var value string
	valueField := addStringField(content, "Value", "", &value)
	onValue := newContentFocus(content, valueField.AsPanel())
	valueField.RemoveFromParent()
	box := unison.NewCheckBox()
	content.AddChild(box)
	target, same = onValue.find(content)
	c.True(target == box.AsPanel(), "a control of another kind may stand in for the field whose label it now has")
	c.False(same, "but isn't taken for it")

	box.RemoveFromParent()
	target, same = onValue.find(content)
	c.Nil(target, "nothing is found in content that has nothing to take the focus")
	c.False(same)
}

// newChoiceWithinChoice returns a sheet with a trait whose modifiers are a mandatory choice holding option A and a
// container, along with the two. The container holds options B and C, and is a mandatory choice with B picked when
// innerIsChoice is true, otherwise a group with B on and A off. The container comes from a user library file holding
// it as the other kind.
func newChoiceWithinChoice(t *testing.T, c check.Checker, innerIsChoice bool) (sheet *Sheet, outer, inner *gurps.TraitModifier) {
	t.Helper()
	_, user := useTestLibraries(t, c)
	RegisterKnownFileTypes()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	outer = gurps.NewTraitModifierChoice(entity, nil)
	outer.SetMandatoryChoice(true)
	a := gurps.NewTraitModifier(entity, outer, false)
	a.Name = "A"
	a.SetEnabled(innerIsChoice)
	if innerIsChoice {
		inner = gurps.NewTraitModifierChoice(entity, outer)
		inner.SetMandatoryChoice(true)
	} else {
		inner = gurps.NewTraitModifier(entity, outer, true)
	}
	inner.Name = "Inner"
	for _, name := range []string{"B", "C"} {
		option := gurps.NewTraitModifier(entity, inner, false)
		option.Name = name
		option.SetEnabled(name == "B")
		inner.Children = append(inner.Children, option)
	}
	outer.Children = []*gurps.TraitModifier{a, inner}

	var lib *gurps.TraitModifier
	if innerIsChoice {
		lib = gurps.NewTraitModifier(nil, nil, true)
	} else {
		lib = gurps.NewTraitModifierChoice(nil, nil)
		lib.SetMandatoryChoice(true)
	}
	lib.Name = inner.Name
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Modifiers/Test" + gurps.TraitModifiersExt}
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{lib},
		filepath.Join(user.Path(), filepath.FromSlash(libFile.Path))))
	inner.Source = gurps.Source{LibraryFile: libFile, TID: lib.TID}

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	trait.Modifiers = []*gurps.TraitModifier{outer}
	gurps.AttachModifiers(trait, trait.Modifiers)
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()
	return sheet, outer, inner
}
