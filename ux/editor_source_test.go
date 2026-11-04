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
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/eqcontainer"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/selfctrl"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/srcstate"
	"github.com/richardwilkes/gcs/v5/ux/svg"
	"github.com/richardwilkes/gcs/v5/ux/uxtest"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/geom"
	"github.com/richardwilkes/toolbox/v2/tid"
	"github.com/richardwilkes/unison"
	"github.com/richardwilkes/unison/enums/mod"
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
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{lib}, filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	local := lib.Clone(libFile, entity, nil, gurps.Reference)
	entity.Traits = append(entity.Traits, local)
	entity.Recalculate()
	return sheet, local
}

// everyEditorContent builds the content of every kind of editor, keyed by what it edits, as displayEditor would. Each
// node that can carry a source has one.
func everyEditorContent(t *testing.T) map[string]*unison.Panel {
	t.Helper()
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
	technique := gurps.NewTechnique(entity, nil, "Karate")
	sourced(technique)
	_, contents["technique"] = buildEditorContent(sheet, technique, initSkillEditor)
	skillContainer := gurps.NewSkill(entity, nil, true)
	sourced(skillContainer)
	_, contents["skill container"] = buildEditorContent(sheet, skillContainer, initSkillEditor)
	_, contents["skill choice"] = buildEditorContent(sheet, gurps.NewSkillChoiceContainer(entity, nil), initSkillEditor)
	spell := gurps.NewSpell(entity, nil, false)
	sourced(spell)
	_, contents["spell"] = buildEditorContent(sheet, spell, initSpellEditor)
	ritualMagicSpell := gurps.NewRitualMagicSpell(entity, nil, false)
	sourced(ritualMagicSpell)
	_, contents["ritual magic spell"] = buildEditorContent(sheet, ritualMagicSpell, initSpellEditor)
	spellContainer := gurps.NewSpell(entity, nil, true)
	sourced(spellContainer)
	_, contents["spell container"] = buildEditorContent(sheet, spellContainer, initSpellEditor)
	_, contents["spell choice"] = buildEditorContent(sheet, gurps.NewSpellChoiceContainer(entity, nil), initSpellEditor)
	note := gurps.NewNote(entity, nil, false)
	sourced(note)
	_, contents["note"] = buildEditorContent(sheet, note, initNoteEditor)
	noteContainer := gurps.NewNote(entity, nil, true)
	sourced(noteContainer)
	_, contents["note container"] = buildEditorContent(sheet, noteContainer, initNoteEditor)
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
	return contents
}

// TestEditorsShowNoSourceFields verifies that no editor shows the ID or the source in its content, which the toolbar's
// source menu shows instead.
func TestEditorsShowNoSourceFields(t *testing.T) {
	c := check.New(t)
	for name, content := range everyEditorContent(t) {
		texts := uxtest.LabelTexts(content)
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
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	libMod := gurps.NewTraitModifier(nil, nil, false)
	libMod.Name = "Long"
	modFile := gurps.LibraryFile{Library: user.Key(), Path: "Modifiers/Test" + gurps.TraitModifiersExt}
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{libMod},
		filepath.Join(user.Path(false), filepath.FromSlash(modFile.Path))))
	libTrait := gurps.NewTrait(nil, nil, false)
	libTrait.Name = "Claws"
	traitFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{libTrait}, filepath.Join(user.Path(false), filepath.FromSlash(traitFile.Path))))
	libEqp := gurps.NewEquipment(nil, nil, false)
	libEqp.Name = "Rope"
	eqpFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{libEqp},
		filepath.Join(user.Path(false), filepath.FromSlash(eqpFile.Path))))
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

// TestEditorSyncFollowsEquipmentKinds verifies that the editor syncs plain equipment and a physical container with each
// other in both directions, since the two have the same data to sync, but won't sync a container whose source has
// another container type, since its data can't carry the container type, nor anything whose source is incompatible.
func TestEditorSyncFollowsEquipmentKinds(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	libGroup := gurps.NewEquipmentGroup(nil, nil)
	libGroup.Name = "Pack"
	libPhysical := gurps.NewEquipment(nil, nil, true)
	libPhysical.Name = "Sack"
	libPhysical.BaseValue = "20"
	libPlain := gurps.NewEquipment(nil, nil, false)
	libPlain.Name = "Rope"
	libPlain.BaseValue = "30"
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{libGroup, libPhysical, libPlain},
		filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	physicalFromGroup := libGroup.Clone(libFile, entity, nil, gurps.Reference)
	physicalFromGroup.ConvertToPhysicalContainer()
	physicalFromGroup.BaseValue = "10"
	physicalFromGroup.Modifiers = []*gurps.EquipmentModifier{gurps.NewEquipmentModifier(entity, nil, false)}
	plainFromPhysical := libPhysical.Clone(libFile, entity, nil, gurps.Reference)
	plainFromPhysical.ConvertToNonContainer()
	plainFromPhysical.BaseValue = "5"
	physicalFromPlain := libPlain.Clone(libFile, entity, nil, gurps.Reference)
	physicalFromPlain.ConvertToContainer()
	physicalFromPlain.BaseValue = "5"
	physicalFromPlain.Children = []*gurps.Equipment{gurps.NewEquipment(entity, physicalFromPlain, false)}
	plainFromGroup := libGroup.Clone(libFile, entity, nil, gurps.Reference)
	plainFromGroup.ConvertToNonContainer()
	plainFromGroup.BaseValue = "5"
	entity.CarriedEquipment = append(entity.CarriedEquipment, physicalFromGroup, plainFromPhysical, physicalFromPlain,
		plainFromGroup)
	entity.Recalculate()
	c.True(physicalFromGroup.IsPhysicalContainer(), "precondition: the copy of the group is a physical container")
	c.True(physicalFromPlain.IsPhysicalContainer(), "precondition: the copy of the equipment is a physical container")

	e, _ := buildEditorContent(sheet, physicalFromGroup, initEquipmentEditor(true))
	summary := sourceMenuSummary(e.sourceMenuEntries())
	c.True(slices.Contains(summary, "# "+srcstate.Mismatched.String()), "the container differs from its source")
	c.True(slices.Contains(summary, "Sync with Source (disabled)"), "but only the list can change its container type")
	e.syncWithSource()
	c.Equal("10", e.editorData.BaseValue, "nothing is synced")
	c.Equal(1, len(e.editorData.Modifiers), "the modifiers are kept")
	c.False(e.isModified(), "nothing is left to apply")

	e, _ = buildEditorContent(sheet, plainFromPhysical, initEquipmentEditor(true))
	entries := e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Mismatched.String()),
		"the plain equipment differs from its source, a physical container")
	requireSourceMenuAction(t, entries, "Sync with Source", "plain equipment can be synced with a physical container")()
	c.Equal("20", e.editorData.BaseValue, "the synced fields come from the library")
	e.applyEdits()
	c.Equal("20", plainFromPhysical.BaseValue, "the synced value is applied")
	state, _ := gurps.MatchSource(plainFromPhysical)
	c.Equal(srcstate.Matched, state, "once applied, the plain equipment matches its source")

	e, _ = buildEditorContent(sheet, physicalFromPlain, initEquipmentEditor(true))
	entries = e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Mismatched.String()),
		"the container differs from its source, plain equipment")
	requireSourceMenuAction(t, entries, "Sync with Source", "a physical container can be synced with plain equipment")()
	c.Equal("30", e.editorData.BaseValue, "the synced fields come from the library")
	e.applyEdits()
	c.Equal("30", physicalFromPlain.BaseValue, "the synced value is applied")
	c.True(physicalFromPlain.IsPhysicalContainer(), "the container stays a physical container")
	c.Equal(1, len(physicalFromPlain.Children), "and keeps what it holds")
	state, _ = gurps.MatchSource(physicalFromPlain)
	c.Equal(srcstate.Matched, state, "once applied, the container matches its source")

	e, _ = buildEditorContent(sheet, plainFromGroup, initEquipmentEditor(true))
	summary = sourceMenuSummary(e.sourceMenuEntries())
	c.True(slices.Contains(summary, "# "+srcstate.Incompatible.String()), "plain equipment can't sync with a group")
	c.True(slices.Contains(summary, "Sync with Source (disabled)"), "so the sync is disabled")
	c.True(slices.Contains(summary, "Clear Source"), "but the source can still be cleared")
	e.syncWithSource()
	c.Equal("5", e.editorData.BaseValue, "nothing is synced")
	c.False(e.isModified(), "nothing is left to apply")
}

// TestEditorSyncSettlesAChoiceAsItsListDoes verifies that syncing a modifier container from its editor leaves its
// options, and those of the choice around it, just as syncing it from its list does, whatever the sync makes of it: a
// choice of a group, a group of a choice, or a mandatory choice of an optional one; and that undo puts them back as
// they were.
func TestEditorSyncSettlesAChoiceAsItsListDoes(t *testing.T) {
	// choiceOutcome describes what the container is, which of option A, of the outer choice, and options B and C, of
	// the container, are on, and whether the outer choice has its pick.
	type choiceOutcome struct {
		choice, mandatory, aOn, bOn, cOn, outerResolved bool
	}
	for _, one := range []struct {
		name   string
		build  func(t *testing.T, c check.Checker) (sheet *Sheet, outer, inner *gurps.TraitModifier)
		synced choiceOutcome
		undone choiceOutcome
	}{
		{
			name: "group to choice",
			build: func(t *testing.T, c check.Checker) (*Sheet, *gurps.TraitModifier, *gurps.TraitModifier) {
				return newChoiceWithinChoice(t, c, false)
			},
			// The new choice keeps the pick it was handed, and the mandatory choice around it, which lost its pick to
			// it, is left for the user to make rather than given one. Undo makes it a group again, whose option is the
			// outer choice's pick once more.
			synced: choiceOutcome{choice: true, mandatory: true, bOn: true},
			undone: choiceOutcome{bOn: true, outerResolved: true},
		},
		{
			name: "choice to group",
			build: func(t *testing.T, c check.Checker) (*Sheet, *gurps.TraitModifier, *gurps.TraitModifier) {
				return newChoiceWithinChoice(t, c, true)
			},
			// The group hands its options to the choice around it, which keeps its own pick over them. Undo puts the
			// inner choice's pick back.
			synced: choiceOutcome{aOn: true, outerResolved: true},
			undone: choiceOutcome{choice: true, mandatory: true, aOn: true, bOn: true, outerResolved: true},
		},
		{
			name:   "optional to mandatory",
			build:  newOptionalChoiceWithinChoice,
			synced: choiceOutcome{choice: true, mandatory: true, aOn: true, outerResolved: true},
			undone: choiceOutcome{choice: true, aOn: true, outerResolved: true},
		},
	} {
		t.Run(one.name, func(t *testing.T) {
			c := check.New(t)
			outcome := func(outer, inner *gurps.TraitModifier) choiceOutcome {
				return choiceOutcome{
					choice:        gurps.IsModifierChoice(inner),
					mandatory:     gurps.IsMandatoryModifierChoice(inner),
					aOn:           outer.Children[0].Enabled(),
					bOn:           inner.Children[0].Enabled(),
					cOn:           inner.Children[1].Enabled(),
					outerResolved: gurps.ModifierChoiceIsResolved(outer),
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
			synced := outcome(outer, inner)
			c.Equal(fromList, synced, "the editor's sync leaves what the list's does")
			c.Equal(one.synced, synced, "the editor's sync settles the container and its options")
			state, _ := gurps.MatchSource(inner)
			c.Equal(srcstate.Matched, state)

			unison.UndoManagerFor(sheet).Undo()
			c.Equal(one.undone, outcome(outer, inner), "undo puts the container and its options back")
		})
	}
}

// TestEditorSyncLeavesTheUsersOwnChoiceChangeToBeSettled verifies that a choice the user made mandatory in its editor is
// given a pick when the changes are applied, even if a sync with its library copy, which is mandatory too, came in
// between. The sync didn't make that change, so the pick isn't left for the user to make, as it is when the sync did.
func TestEditorSyncLeavesTheUsersOwnChoiceChangeToBeSettled(t *testing.T) {
	c := check.New(t)
	sheet, _, inner := newOptionalChoiceWithinChoice(t, c)
	inner.LocalNotes = "Old notes"
	c.False(gurps.IsMandatoryModifierChoice(inner), "precondition: the choice is optional")
	c.False(inner.Children[0].Enabled() || inner.Children[1].Enabled(), "precondition: with nothing picked")

	e, _ := buildEditorContent(sheet, inner, initTraitModifierEditor)
	e.editorData.SetMandatoryChoice(true)
	requireSourceMenuAction(t, e.sourceMenuEntries(), "Sync with Source",
		"the choice still differs from its library copy in its notes")()
	c.Equal("", e.editorData.LocalNotes, "the sync brings the library's notes across")
	c.Nil(e.syncedChoice, "and has no change to what the choice asks for to answer for")
	e.applyEdits()
	c.True(gurps.IsMandatoryModifierChoice(inner), "the choice is mandatory now")
	c.True(gurps.ModifierChoiceIsResolved(inner), "and, since the user made it so, is given a pick")
	c.True(inner.Children[0].Enabled() != inner.Children[1].Enabled())
	state, _ := gurps.MatchSource(inner)
	c.Equal(srcstate.Matched, state)
}

// TestEditorSyncWithSourceForEquipment verifies that equipment of the same kind as its library copy can be synced from
// its editor's menu, which rebuilds the content around the library's data with the modifiers the equipment has and
// the weapons its library copy has, all pointed at the equipment. The uses left are clamped to the library's lower
// maximum as the new content settles, which leaves nothing in the editor's undo history.
func TestEditorSyncWithSourceForEquipment(t *testing.T) {
	c := check.New(t)
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	lib := gurps.NewEquipment(nil, nil, false)
	lib.Name = "Rope"
	lib.BaseValue = "5"
	lib.MaxUses = 2
	lib.Weapons = []*gurps.Weapon{gurps.NewWeapon(lib, true)}
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Equipment/Test" + gurps.EquipmentExt}
	c.NoError(gurps.SaveEquipment([]*gurps.Equipment{lib}, filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	sheet := newTestSheetForTemplate(t)
	entity := sheet.Entity()
	local := lib.Clone(libFile, entity, nil, gurps.Reference)
	local.Name = "Cord"
	local.BaseValue = "9"
	local.MaxUses = 10
	local.Uses = 8
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
	oldMelee := e.meleeWeapons
	entries := e.sourceMenuEntries()
	c.True(slices.Contains(sourceMenuSummary(entries), "# "+srcstate.Mismatched.String()))
	requireSourceMenuAction(t, entries, "Sync with Source",
		"equipment that differs from a source of its own kind can be synced")()

	c.Equal("Rope", e.editorData.Name, "synced fields come from the library")
	c.Equal("5", e.editorData.BaseValue)
	c.Equal(fxp.Three, e.editorData.Quantity, "pending changes to fields a sync leaves alone are kept")
	c.Equal(2, e.editorData.MaxUses)
	c.Equal(2, e.editorData.Uses, "the uses left, which a sync leaves alone, are clamped to the library's maximum")
	c.False(e.UndoManager().CanUndo() || e.UndoManager().CanRedo(),
		"the clamp, made as the new content settles, isn't left in the editor's undo history")
	c.Equal(1, len(e.editorData.Modifiers), "the modifiers aren't synced")
	c.Equal("Fine", e.editorData.Modifiers[0].Name)
	c.True(e.editorData.Modifiers[0].Target() == local, "the modifiers stay pointed at the equipment")
	c.Equal(1, len(e.editorData.Weapons), "the library's weapon arrives")
	c.True(e.editorData.Weapons[0].Owner == gurps.WeaponOwner(local), "pointed at the equipment")
	c.Equal("Cord", local.Name, "the equipment is left alone until the changes are applied")
	for _, child := range oldChildren {
		c.Nil(child.Parent(), "the old content is thrown away")
	}
	c.True(e.meleeWeapons != nil && e.meleeWeapons != oldMelee, "the weapons panel is rebuilt")
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

	rolls := uxtest.PanelsOfType[*unison.PopupMenu[selfctrl.Roll]](content)
	if len(rolls) != 1 {
		t.Fatalf("expected the trait editor to have one self-control popup, found %d", len(rolls))
	}
	rolls[0].Select(selfctrl.CR12)
	c.Equal(selfctrl.ActionPenalty, e.editorData.SelfControlAdj, "giving the trait a roll keeps the adjustment")
	rolls[0].Select(selfctrl.None)
	c.Equal(selfctrl.NoAdjustment, e.editorData.SelfControlAdj, "and taking the roll away clears it")

	for _, defaultType := range []string{gurps.DexterityID, gurps.SkillID} {
		_, user := uxtest.UseTestLibraries(t, c)
		libTechnique := gurps.NewTechnique(nil, nil, "Karate")
		libTechnique.Name = "Kicking"
		libTechnique.PageRef = "B230"
		libTechnique.TechniqueDefault.DefaultType = defaultType
		libFile := gurps.LibraryFile{Library: user.Key(), Path: "Skills/Test" + gurps.SkillsExt}
		c.NoError(gurps.SaveSkills([]*gurps.Skill{libTechnique},
			filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
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
	popups := uxtest.PanelsOfType[*unison.PopupMenu[string]](content)
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

	// A panel that isn't among its parent's children, as the cell a table hands the focus to isn't, has an index of -1.
	stray := &contentFocus{path: []focusStep{{kind: reflect.TypeOf(valueField), label: "Value", index: -1}}}
	target, same = stray.find(content)
	c.True(target == box.AsPanel(), "a place that was never among its parent's children leads to the nearest control")
	c.False(same)
	stray.path[0].label = ""
	c.Nil(stray.path[0].childOf(content), "and to no child of its parent")

	box.RemoveFromParent()
	target, same = onValue.find(content)
	c.Nil(target, "nothing is found in content that has nothing to take the focus")
	c.False(same)
}

// newOptionalChoiceWithinChoice is newChoiceWithinChoice with the inner choice optional, with nothing picked, and its
// library copy a mandatory choice.
func newOptionalChoiceWithinChoice(t *testing.T, c check.Checker) (sheet *Sheet, outer, inner *gurps.TraitModifier) {
	t.Helper()
	sheet, outer, inner = newChoiceWithinChoice(t, c, true)
	inner.SetMandatoryChoice(false)
	inner.Children[0].SetDisabled(true)
	lib := gurps.NewTraitModifierChoice(nil, nil)
	lib.SetMandatoryChoice(true)
	lib.Name = inner.Name
	lib.TID = inner.Source.TID
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{lib},
		filepath.Join(gurps.GlobalSettings().Libraries.User().Path(false), filepath.FromSlash(inner.Source.Path))))
	sheet.Entity().Recalculate()
	return sheet, outer, inner
}

// newChoiceWithinChoice returns a sheet with a trait whose modifiers are a mandatory choice holding option A and a
// container, along with the two. The container holds options B and C, and is a mandatory choice with B picked when
// innerIsChoice is true, otherwise a group with B on and A off. The container comes from a user library file holding
// it as the other kind.
func newChoiceWithinChoice(t *testing.T, c check.Checker, innerIsChoice bool) (sheet *Sheet, outer, inner *gurps.TraitModifier) {
	t.Helper()
	_, user := uxtest.UseTestLibraries(t, c)
	RegisterKnownFileTypes()
	sheet = newTestSheetForTemplate(t)
	entity := sheet.Entity()
	outer = gurps.NewTraitModifierChoice(entity, nil)
	outer.SetMandatoryChoice(true)
	a := gurps.NewTraitModifier(entity, outer, false)
	a.Name = "A"
	a.SetDisabled(!innerIsChoice)
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
		option.SetDisabled(name != "B")
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
		filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	inner.Source = gurps.Source{LibraryFile: libFile, TID: lib.TID}

	trait := gurps.NewTrait(entity, nil, false)
	trait.Name = "Trait"
	trait.Modifiers = []*gurps.TraitModifier{outer}
	gurps.AttachModifiers(trait, trait.Modifiers)
	entity.Traits = append(entity.Traits, trait)
	entity.Recalculate()
	return sheet, outer, inner
}

// stringFieldLabeled returns the string field that follows the label bearing the given text among the panel's direct
// children, or nil if there is none.
func stringFieldLabeled(p *unison.Panel, labelText string) *StringField {
	children := p.Children()
	for i, child := range children {
		if label, ok := child.Self.(*unison.Label); ok && label.String() == labelText && i+1 < len(children) {
			if field, ok2 := children[i+1].Self.(*StringField); ok2 {
				return field
			}
		}
	}
	return nil
}

// openSourceMenu clicks the editor's source button and returns the panels of the items of the menu that opens.
func openSourceMenu(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, d unison.Dockable) []*unison.Panel {
	t.Helper()
	var button *unison.Button
	screen.Do(func() { button = buttonWithSVG(d.AsPanel(), svg.Database) })
	if button == nil {
		t.Fatal("the editor has no source button")
	}
	screen.Click(screen.PanelCenter(button))
	var items []*unison.Panel
	screen.Do(func() { items = slices.Clone(uxtest.MenuItemPanels(uxtest.OpenMenuPopup(wnd))) })
	if len(items) == 0 {
		t.Fatal("the source button opened no menu")
	}
	return items
}

// TestEditorSourceMenu drives the source menu of a trait editor as a user would: the button names itself, the menu
// shows the trait's ID and source and how it compares to the library, choosing the ID copies it, a command that doesn't
// apply is disabled and does nothing, and a source cleared from the menu stays put when the editor's changes are
// discarded. A weapon's editor has no such menu, since a weapon has no source.
func TestEditorSourceMenu(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	trait := gurps.NewTrait(sheet.Entity(), nil, false)
	trait.Name = "Claws"
	trait.Source = testSource
	custom := gurps.NewTrait(sheet.Entity(), nil, false)
	custom.Name = "Fangs"
	var e, customEditor *editor[*gurps.Trait, *gurps.TraitEditData]
	var tip string
	screen.Do(func() {
		sheet.Entity().Traits = append(sheet.Entity().Traits, trait, custom)
		sheet.Rebuild(true)
		customEditor = EditTrait(sheet, custom)
		if button := buttonWithSVG(customEditor.AsPanel(), svg.Database); button != nil {
			tip = tooltipText(button.Tooltip)
		}
	})
	c.Equal("ID & Library Source", tip, "the source button is named by its tooltip")

	// clearSource has no check of its own, so only the menu keeps it from a trait without a source.
	items := openSourceMenu(t, screen, wnd, customEditor)
	c.Equal([]string{
		"ID: " + string(custom.ID()),
		srcstate.Custom.String(),
		"Sync with Source",
		"Clear Source",
	}, contextMenuTitles(t, screen, wnd))
	clearItem := items[len(items)-1]
	node := screen.AccessibilityNodeFor(clearItem)
	c.True(node != nil && node.Disabled, "a trait without a source has none to clear")
	screen.Click(screen.PanelCenter(clearItem))
	var modified, menuOpen bool
	screen.Do(func() {
		modified = customEditor.isModified()
		menuOpen = uxtest.OpenMenuPopup(wnd) != nil
	})
	c.False(modified, "choosing a disabled command does nothing")
	c.True(menuOpen, "and leaves the menu open")
	closeContextMenu(c, screen, wnd)
	uxtest.CloseEditorWithoutPrompt(t, screen, customEditor)

	screen.Do(func() { e = EditTrait(sheet, trait) })
	items = openSourceMenu(t, screen, wnd, e)
	c.Equal([]string{
		"ID: " + string(trait.ID()),
		"Source ID: " + string(testSource.TID),
		"Source Library: " + testSource.Library,
		"Source Path: " + testSource.Path,
		srcstate.Missing.String(),
		"Sync with Source",
		"Clear Source",
	}, contextMenuTitles(t, screen, wnd))
	node = screen.AccessibilityNodeFor(items[len(items)-2])
	c.True(node != nil && node.Disabled, "a source no library holds can't be synced with")
	closeContextMenu(c, screen, wnd)

	items = openSourceMenu(t, screen, wnd, e)
	screen.Click(screen.PanelCenter(items[0]))
	var clip string
	screen.Do(func() { clip = unison.ClipboardGetText() })
	c.Equal(string(trait.ID()), clip, "choosing the ID copies it")

	items = openSourceMenu(t, screen, wnd, e)
	screen.Click(screen.PanelCenter(items[len(items)-1]))
	var cancelEnabled bool
	screen.Do(func() {
		modified = e.isModified()
		cancelEnabled = e.cancelButton.Enabled()
	})
	c.True(modified, "clearing the source is a change to apply")
	c.True(cancelEnabled, "or to discard")
	c.Equal(testSource, trait.Source, "the trait keeps its source until the change is applied")
	screen.Click(screen.PanelCenter(e.cancelButton))
	var open bool
	screen.Do(func() {
		open = slices.ContainsFunc(AllDockables(), func(d unison.Dockable) bool { return d.AsPanel().Self == e })
	})
	c.False(open, "Discard closes the editor")
	c.Equal(testSource, trait.Source, "and the trait keeps its source")

	var weaponEditor unison.Dockable
	var hasButton bool
	screen.Do(func() {
		before := AllDockables()
		EditWeapon(sheet, gurps.NewWeapon(trait, true))
		for _, d := range AllDockables() {
			if !slices.Contains(before, d) {
				weaponEditor = d
			}
		}
		if weaponEditor != nil {
			hasButton = buttonWithSVG(weaponEditor.AsPanel(), svg.Database) != nil
		}
	})
	if weaponEditor == nil {
		t.Fatal("the weapon editor did not open")
	}
	c.False(hasButton, "a weapon has no source, so its editor has no source menu")
	if closer, isCloser := weaponEditor.(interface {
		unison.Dockable
		unison.TabCloser
	}); isCloser {
		uxtest.CloseEditorWithoutPrompt(t, screen, closer)
	}
}

// newEditedLibraryTrait saves a trait named "Claws", with a melee weapon, further altered by prepareLib if not nil, into
// a file of the user library, adds a copy sourced from that file to the sheet as "Claws (old)", further altered by
// prepare if not nil, and opens the copy in an editor.
func newEditedLibraryTrait(t *testing.T, c check.Checker, screen *unison.HeadlessScreen, sheet *Sheet, prepareLib, prepare func(trait *gurps.Trait)) (*editor[*gurps.Trait, *gurps.TraitEditData], *gurps.Trait) {
	t.Helper()
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewTrait(nil, nil, false)
	lib.Name = "Claws"
	lib.Weapons = []*gurps.Weapon{gurps.NewWeapon(lib, true)}
	if prepareLib != nil {
		prepareLib(lib)
	}
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Traits/Test" + gurps.TraitsExt}
	c.NoError(gurps.SaveTraits([]*gurps.Trait{lib}, filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	var e *editor[*gurps.Trait, *gurps.TraitEditData]
	var local *gurps.Trait
	screen.Do(func() {
		entity := sheet.Entity()
		local = lib.Clone(libFile, entity, nil, gurps.Reference)
		local.Name = "Claws (old)"
		if prepare != nil {
			prepare(local)
		}
		entity.Traits = append(entity.Traits, local)
		sheet.Rebuild(true)
		e = EditTrait(sheet, local)
	})
	return e, local
}

// editorField returns the string field of the editor's content under the label, ending the test if there is none.
func editorField[N gurps.Node[N], D gurps.EditorData[N]](t *testing.T, screen *unison.HeadlessScreen, e *editor[N, D], label string) *StringField {
	t.Helper()
	var field *StringField
	screen.Do(func() { field = stringFieldLabeled(e.content, label) })
	if field == nil {
		t.Fatalf("the editor has no %s field", label)
	}
	return field
}

// chooseSyncWithSource chooses Sync with Source from the editor's source menu, without reading the menu's items as
// contextMenuTitles would, since that turns accessibility support on, after which clicking a button gives it the focus.
func chooseSyncWithSource(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, d unison.Dockable) {
	t.Helper()
	items := openSourceMenu(t, screen, wnd, d)
	screen.Click(screen.PanelCenter(items[len(items)-2]))
}

// TestEditorSyncAsksGrantItemEditorToDiscard verifies that syncing from the source menu asks the editor open on the
// item of one of the editor's grants whether to discard its changes, rather than whether to save them, since the sync
// replaces the features, grants and all.
func TestEditorSyncAsksGrantItemEditorToDiscard(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
		g := gurps.NewGrant(feature.GrantTrait)
		g.SetItem(newGrantTestTrait("Night Vision", 0), gurps.LibraryFile{})
		lib.Features = gurps.Features{g}
	}, nil)
	isTraitEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
		return isEditor
	}
	screen.Do(func() {
		p := uxtest.PanelsOfType[*featuresPanel](e.content)[0]
		p.editGrantedItem(grantAt(t, p, 0))
	})
	nested := uxtest.OtherEditor(t, screen, e, isTraitEditor)
	screen.Do(func() {
		nested.editorData.Name = "Dark Vision"
		nested.MarkModified(nil)
	})
	chooseSyncWithSource(t, screen, wnd, e)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var prompt []string
	var offersSave bool
	screen.Do(func() {
		prompt = uxtest.LabelTexts(dialogWnd.Content())
		offersSave = dialog.Button(unison.ModalResponseDiscard) != nil
	})
	c.True(slices.Contains(prompt, "Discard changes made to"), "the prompt asks whether to discard the item's changes: %v",
		prompt)
	c.True(slices.Contains(prompt, "Syncing with the source replaces this item."), "and says why: %v", prompt)
	c.False(offersSave, "rather than whether to save them, which the sync would undo")
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	var discardReason string
	screen.Do(func() { discardReason = nested.discardReason })
	c.Equal("", discardReason, "canceling leaves the editor asking whether to save when closed for any other reason")
}

// TestEditorDiscardAndSyncReachSubEditorsOfSubEditors verifies that Discard, and Sync with Source, ask an editor opened
// from the editor of a grant's item, on one of the item's modifiers, whether to discard its changes rather than whether
// to save them, as they ask the item's editor, since what it would save goes into the same data.
func TestEditorDiscardAndSyncReachSubEditorsOfSubEditors(t *testing.T) {
	for _, tc := range []struct {
		name string
		// act discards or syncs the outer editor.
		act func(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, e *editor[*gurps.Trait, *gurps.TraitEditData])
		// reason is how the prompt's reason starts.
		reason string
	}{
		{name: "discard", act: func(_ *testing.T, screen *unison.HeadlessScreen, _ *unison.Window, e *editor[*gurps.Trait, *gurps.TraitEditData]) {
			screen.Do(func() {
				e.editorData.Name = "Talons"
				e.MarkModified(nil)
			})
			screen.Click(screen.PanelCenter(e.cancelButton))
		}, reason: "Discarding the changes made to"},
		{name: "sync", act: func(t *testing.T, screen *unison.HeadlessScreen, wnd *unison.Window, e *editor[*gurps.Trait, *gurps.TraitEditData]) {
			chooseSyncWithSource(t, screen, wnd, e)
		}, reason: "Syncing with the source replaces this item."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := check.New(t)
			screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
			sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
			if !ok {
				t.Fatal("New Character Sheet must open a character sheet")
			}
			e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
				g := gurps.NewGrant(feature.GrantTrait)
				nightVision := newGrantTestTrait("Night Vision", 0)
				modifier := gurps.NewTraitModifier(nil, nil, false)
				modifier.Name = "Long"
				nightVision.Modifiers = []*gurps.TraitModifier{modifier}
				g.SetItem(nightVision, gurps.LibraryFile{})
				lib.Features = gurps.Features{g}
			}, nil)
			isTraitEditor := func(d unison.Dockable) bool {
				_, isEditor := d.AsPanel().Self.(*editor[*gurps.Trait, *gurps.TraitEditData])
				return isEditor
			}
			isModifierEditor := func(d unison.Dockable) bool {
				_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
				return isEditor
			}
			screen.Do(func() {
				p := uxtest.PanelsOfType[*featuresPanel](e.content)[0]
				p.editGrantedItem(grantAt(t, p, 0))
			})
			nested := uxtest.OtherEditor(t, screen, e, isTraitEditor)
			screen.Do(func() { EditTraitModifier(nested, nested.editorData.Modifiers[0]) })
			modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
				isModifierEditor)
			screen.Do(func() {
				modifierEditor.editorData.Name = "Longer"
				modifierEditor.MarkModified(nil)
			})
			tc.act(t, screen, wnd, e)
			dialogWnd, dialog := modalDialog(t, screen, wnd)
			var prompt []string
			var offersSave bool
			screen.Do(func() {
				prompt = uxtest.LabelTexts(dialogWnd.Content())
				offersSave = dialog.Button(unison.ModalResponseDiscard) != nil
			})
			c.True(slices.Contains(prompt, "Discard changes made to"),
				"the prompt asks whether to discard the modifier's changes: %v", prompt)
			c.True(slices.ContainsFunc(prompt, func(s string) bool { return strings.HasPrefix(s, tc.reason) }),
				"and says why: %v", prompt)
			c.False(offersSave, "rather than whether to save them")
			screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
			var discardReason string
			screen.Do(func() { discardReason = modifierEditor.discardReason })
			c.Equal("", discardReason, "canceling leaves the editor asking whether to save when closed for any other reason")
		})
	}
}

// TestEditorSyncKeepsTheEditInProgress verifies that syncing from the source menu while a field holds the focus keeps
// what was typed into it when the sync leaves that field alone, and leaves the focus, the caret and the editor's
// scrolling as they were, so that typing carries on there. The field gives the focus up before it is replaced, while it
// is still part of the editor.
func TestEditorSyncKeepsTheEditInProgress(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Enlarged, so that the editor's content is bigger than its view, which can then be scrolled past the field.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialEditorUIScale, 300)
	e, local := newEditedLibraryTrait(t, c, screen, sheet, nil, nil)
	vttNotes := editorField(t, screen, e, "VTT Notes")
	// Near its start, since at this size the middle of the field lies beyond the window's edge.
	screen.Click(screen.PanelPoint(vttNotes, geom.Point{X: 4, Y: 4}))
	var focused bool
	screen.Do(func() { focused = wnd.CurrentFocus() == vttNotes.AsPanel() })
	c.True(focused, "precondition: the click focuses the field")
	screen.Type("Typed")
	screen.KeyPress(unison.KeyLeft, 0)
	screen.KeyPress(unison.KeyLeft, 0)

	// Scrolled past the field, since giving the focus to a field out of view scrolls it into view, which putting the
	// focus back must not do.
	var h, v float32
	var inView bool
	var lostFocusWithinEditor []bool
	screen.Do(func() {
		e.scroll.SetPosition(0, e.content.FrameRect().Height)
		h, v = e.scroll.Position()
		inView = !visibleRect(vttNotes.AsPanel()).Empty()
		lostFocus := vttNotes.LostFocusCallback
		vttNotes.LostFocusCallback = func() {
			lostFocusWithinEditor = append(lostFocusWithinEditor, unison.AncestorIsOrSelf(vttNotes, e.content))
			lostFocus()
		}
	})
	c.False(inView, "precondition: the field is scrolled out of view")

	chooseSyncWithSource(t, screen, wnd, e)
	var name, notes string
	var modified, focusOnNotes bool
	var selStart, selEnd int
	var hAfter, vAfter float32
	screen.Do(func() {
		name = e.editorData.Name
		notes = e.editorData.VTTNotes
		modified = e.isModified()
		rebuilt := stringFieldLabeled(e.content, "VTT Notes")
		focusOnNotes = rebuilt != nil && rebuilt != vttNotes && wnd.CurrentFocus() == rebuilt.AsPanel()
		if rebuilt != nil {
			selStart, selEnd = rebuilt.Selection()
		}
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal("Claws", name, "the sync brings the library's name across")
	c.Equal("Typed", notes, "and keeps what was typed into a field it leaves alone")
	c.True(modified)
	c.True(len(lostFocusWithinEditor) != 0 && !slices.Contains(lostFocusWithinEditor, false),
		"the field gives up the focus while it is still part of the editor: %v", lostFocusWithinEditor)
	c.True(focusOnNotes, "the focus is on the rebuilt field that had it")
	c.Equal(3, selStart, "with the caret where it was")
	c.Equal(3, selEnd, "and nothing selected")
	c.Equal(h, hAfter, "the editor stays scrolled as it was")
	c.Equal(v, vAfter)
	c.Equal("Claws (old)", local.Name, "the trait is left alone until the changes are applied")

	screen.Type("X")
	screen.KeyPress(unison.KeyReturn, mod.OSMenuCommand())
	var state srcstate.Value
	screen.Do(func() { state, _ = gurps.MatchSource(local) })
	c.Equal("Claws", local.Name, "applying the changes syncs the trait")
	c.Equal("TypXed", local.VTTNotes, "typing carried on where it left off")
	c.Equal(srcstate.Matched, state)
	// Applying the changes modified the sheet, which is to be left unmodified (see uxtest.StartHeadlessWorkspace).
	screen.Do(sheet.markUnmodified)
}

// TestEditorSyncPutsTheFocusBack verifies where the focus goes after a sync that changes what held it: a field whose
// text the sync replaced gets it back with all of its text selected, and when the control that held it is gone, as the
// popup of a modifier choice is once the sync has made the choice a group, the nearest control gets it.
func TestEditorSyncPutsTheFocusBack(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, nil)
	nameField := editorField(t, screen, e, "Name")
	screen.Click(screen.PanelCenter(nameField))
	screen.KeyPress(unison.KeyEnd, 0)
	var focused bool
	var selStart, selEnd int
	screen.Do(func() {
		focused = wnd.CurrentFocus() == nameField.AsPanel()
		selStart, selEnd = nameField.Selection()
	})
	c.True(focused, "precondition: the click focuses the field")
	c.Equal(len("Claws (old)"), selStart, "precondition: the caret is at the end of the name")
	c.Equal(selStart, selEnd, "precondition: nothing is selected")

	chooseSyncWithSource(t, screen, wnd, e)
	var text string
	screen.Do(func() {
		rebuilt := stringFieldLabeled(e.content, "Name")
		focused = rebuilt != nil && rebuilt != nameField && wnd.CurrentFocus() == rebuilt.AsPanel()
		if rebuilt != nil {
			text = rebuilt.Text()
			selStart, selEnd = rebuilt.Selection()
		}
	})
	c.Equal("Claws", text, "the sync replaces the name")
	c.True(focused, "the focus is on the rebuilt field that had it")
	c.Equal(0, selStart, "with all of what the sync brought in selected")
	c.Equal(len("Claws"), selEnd)
	screen.Click(screen.PanelCenter(e.cancelButton))

	// A choice whose library copy is a group, with notes long enough to put what follows them out of view.
	user := gurps.GlobalSettings().Libraries.User()
	lib := gurps.NewTraitModifier(nil, nil, true)
	lib.Name = "Options"
	lib.LocalNotes = strings.Repeat("A line of notes.\n", 80)
	libFile := gurps.LibraryFile{Library: user.Key(), Path: "Modifiers/Test" + gurps.TraitModifiersExt}
	c.NoError(gurps.SaveTraitModifiers([]*gurps.TraitModifier{lib},
		filepath.Join(user.Path(false), filepath.FromSlash(libFile.Path))))
	isChoiceEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return isEditor
	}
	screen.Do(func() {
		entity := sheet.Entity()
		choice := gurps.NewTraitModifierChoice(entity, nil)
		choice.Name = lib.Name
		choice.LocalNotes = lib.LocalNotes
		choice.Source = gurps.Source{LibraryFile: libFile, TID: lib.TID}
		trait := gurps.NewTrait(entity, nil, false)
		trait.Modifiers = []*gurps.TraitModifier{choice}
		gurps.AttachModifiers(trait, trait.Modifiers)
		entity.Traits = append(entity.Traits, trait)
		sheet.Rebuild(true)
		EditTraitModifier(sheet, choice)
	})
	choiceEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen, isChoiceEditor)
	var popup *unison.Panel
	var popupInView bool
	screen.Do(func() {
		if popups := uxtest.PanelsOfType[*unison.PopupMenu[string]](choiceEditor.content); len(popups) == 1 {
			popup = popups[0].AsPanel()
			popup.RequestFocus()
			focused = wnd.CurrentFocus() == popup
			choiceEditor.scroll.SetPosition(0, 0)
			popupInView = !visibleRect(popup).Empty()
		}
	})
	if popup == nil {
		t.Fatal("the choice's editor must have the popup that says what the choice asks for")
	}
	c.True(focused, "precondition: the popup holds the focus")
	c.False(popupInView, "precondition: the popup is scrolled out of view")

	chooseSyncWithSource(t, screen, wnd, choiceEditor)
	var isChoice, focusOnTags, tagsInView bool
	screen.Do(func() {
		isChoice = choiceEditor.editorData.IsChoice()
		if tags := stringFieldLabeled(choiceEditor.content, "Tags"); tags != nil {
			focusOnTags = wnd.CurrentFocus() == tags.AsPanel()
			tagsInView = fullyVisible(tags.AsPanel())
		}
	})
	c.False(isChoice, "the sync makes the choice a group, whose editor has no such popup")
	c.True(focusOnTags, "the focus goes to the field that is now where the popup was")
	c.True(tagsInView, "which is scrolled into view")
	screen.Click(screen.PanelCenter(choiceEditor.cancelButton))
}

// TestEditorSyncKeepsTheFocusInPlaceAsContentGrows verifies that when a sync grows the content ahead of the field that
// holds the focus, as longer notes from the library do, the editor scrolls along with the field, which stays where it
// was within the view rather than being pushed out of it.
func TestEditorSyncKeepsTheFocusInPlaceAsContentGrows(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
		lib.LocalNotes = strings.Repeat("A line of notes.\n", 150)
	}, func(local *gurps.Trait) {
		local.LocalNotes = ""
		// Tall enough that the content doesn't fit the view even before the sync, so that it can be scrolled by all
		// of what it then grows by.
		local.UserDesc = strings.Repeat("A line of my own.\n", 20)
	})
	tags := editorField(t, screen, e, "Tags")
	screen.Do(tags.ScrollIntoView)
	screen.Click(screen.PanelCenter(tags))
	var focused bool
	var before geom.Rect
	screen.Do(func() {
		focused = wnd.CurrentFocus() == tags.AsPanel()
		before = visibleRect(tags.AsPanel())
	})
	c.True(focused, "precondition: the click focuses the field")

	chooseSyncWithSource(t, screen, wnd, e)
	var notes string
	var focusOnTags bool
	var v, viewHeight float32
	var after, whole geom.Rect
	screen.Do(func() {
		notes = e.editorData.LocalNotes
		if rebuilt := stringFieldLabeled(e.content, "Tags"); rebuilt != nil {
			focusOnTags = rebuilt != tags && wnd.CurrentFocus() == rebuilt.AsPanel()
			after = visibleRect(rebuilt.AsPanel())
			whole = rebuilt.RectToRoot(rebuilt.ContentRect(false))
		}
		_, v = e.scroll.Position()
		viewHeight = e.scroll.ContentView().ContentRect(false).Height
	})
	c.Equal(150, strings.Count(notes, "\n"), "the sync brings the library's notes across")
	c.True(focusOnTags, "the focus is on the rebuilt field that had it")
	c.True(v > viewHeight, "the notes push the field down by more than the view's height, and the editor scrolls "+
		"along with it: scrolled to %v, in a view %v high", v, viewHeight)
	c.True(nearlyWithin(whole, after), "so all of the field is in view: %v within %v", whole, after)
	c.True(nearlyWithin(before, after) && nearlyWithin(after, before),
		"right where it was: %v, was %v", after, before)
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncClosesItsSubEditors verifies that syncing from the source menu first closes the editors open on the
// modifiers and weapons within the editor's data, and syncs nothing when one of them is kept open, as canceling its
// prompt does. A weapon's editor with changes asks only whether to discard them, since the sync replaces the weapons
// and saving the changes would have them thrown away. Closing a sub-editor gives the focus to the list it was opened
// from and scrolls the editor to that list, yet the focus, its selection and the editor's scrolling all end up as they
// were when the sync was chosen.
func TestEditorSyncClosesItsSubEditors(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Enlarged, so that the lists at the end of the editor are out of view while a field near its start is in view.
	uxtest.SwapForTest(t, &gurps.GlobalSettings().General.InitialEditorUIScale, 300)
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, func(local *gurps.Trait) {
		local.Weapons = []*gurps.Weapon{gurps.NewWeapon(local, true)}
		local.Weapons[0].Usage = "Thrust"
		modifier := gurps.NewTraitModifier(local.DataOwner(), nil, false)
		modifier.Name = "Long"
		local.Modifiers = []*gurps.TraitModifier{modifier}
		gurps.AttachModifiers(local, local.Modifiers)
	})
	isWeaponEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Weapon, *gurps.Weapon])
		return isEditor
	}
	isModifierEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return isEditor
	}
	// Opened as from the list of weapons, which holds the focus as the weapon's editor opens.
	screen.Do(func() {
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
	})
	weaponEditor := uxtest.SoleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
	screen.Do(func() {
		weaponEditor.editorData.Usage = "Swung"
		weaponEditor.MarkModified(nil)
	})
	vttNotes := editorField(t, screen, e, "VTT Notes")
	// Brought into view to be clicked on, which leaves the list of weapons out of view.
	screen.Do(vttNotes.ScrollIntoView)
	screen.Click(screen.PanelPoint(vttNotes, geom.Point{X: 4, Y: 4}))
	screen.Type("Typed")
	// A selection made backwards, whose start is the end that moves as it is extended.
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	var focused, listInView bool
	var h, v float32
	screen.Do(func() {
		focused = wnd.CurrentFocus() == vttNotes.AsPanel()
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
		h, v = e.scroll.Position()
	})
	c.True(focused, "precondition: the field holds the focus")
	c.False(listInView, "precondition: the list of weapons is out of view")

	// Canceling the prompt to discard the weapon's changes keeps its editor open, so nothing is synced.
	chooseSyncWithSource(t, screen, wnd, e)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var prompt []string
	var offersSave bool
	screen.Do(func() {
		prompt = uxtest.LabelTexts(dialogWnd.Content())
		offersSave = dialog.Button(unison.ModalResponseDiscard) != nil
	})
	c.True(slices.Contains(prompt, "Discard changes made to"), "the prompt asks whether to discard the weapon's "+
		"changes: %v", prompt)
	c.True(slices.Contains(prompt, "Syncing with the source replaces this weapon."), "and says why: %v", prompt)
	c.False(offersSave, "rather than whether to save them, which the sync would undo")
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseCancel)))
	var name, discardReason string
	var weaponEditors, selStart, selEnd int
	var hAfter, vAfter float32
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		discardReason = weaponEditor.discardReason
		focused = wnd.CurrentFocus() == vttNotes.AsPanel()
		selStart, selEnd = vttNotes.Selection()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(1, weaponEditors, "canceling keeps the weapon's editor open")
	c.Equal("", discardReason, "which asks whether to save its changes again when closed for any other reason")
	c.Equal("Claws (old)", name, "so nothing is synced")
	c.True(focused, "and the focus is back on the field that had it")
	c.Equal(3, selStart, "with the selection as it was")
	c.Equal(5, selEnd)
	c.Equal(h, hAfter, "and the editor scrolled as it was")
	c.Equal(v, vAfter)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.Do(func() {
		selStart, selEnd = vttNotes.Selection()
		h, v = e.scroll.Position()
	})
	c.Equal(2, selStart, "the selection is still extended from its start")
	c.Equal(5, selEnd)

	// Discarding the weapon's changes lets its editor close, and the sync go ahead.
	chooseSyncWithSource(t, screen, wnd, e)
	_, dialog = modalDialog(t, screen, wnd)
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	rebuilt := editorField(t, screen, e, "VTT Notes")
	var notes, usage string
	screen.Do(func() {
		name = e.editorData.Name
		notes = e.editorData.VTTNotes
		if len(e.editorData.Weapons) == 1 {
			usage = e.editorData.Weapons[0].Usage
		}
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focused = rebuilt != vttNotes && wnd.CurrentFocus() == rebuilt.AsPanel()
		selStart, selEnd = rebuilt.Selection()
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.Equal("Typed", notes)
	c.Equal("", usage, "with the library's weapon in place of the one that was edited")
	c.True(focused, "the focus is on the rebuilt field that had it, not the list the weapon's editor gave it to")
	c.Equal(2, selStart, "with the selection as it was")
	c.Equal(5, selEnd)
	c.Equal(h, hAfter, "the editor is scrolled as it was, not to the list the weapon's editor was opened from")
	c.Equal(v, vAfter)
	c.False(listInView)
	screen.KeyPress(unison.KeyLeft, mod.Shift)
	screen.Do(func() { selStart, selEnd = rebuilt.Selection() })
	c.Equal(1, selStart, "the rebuilt field's selection is extended from its start as well")
	c.Equal(5, selEnd)

	// A focus that isn't within the content, as the source button's is once a screen reader has clicked it, is put back
	// too, rather than left on the list a closed sub-editor gave it to, and the editor is scrolled as it was.
	var sourceButton *unison.Button
	var modifiers *traitModifiersPanel
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
		if panels := uxtest.PanelsOfType[*traitModifiersPanel](e.content); len(panels) == 1 {
			modifiers = panels[0]
			modifiers.table.RequestFocus()
			EditTraitModifier(e, e.editorData.Modifiers[0])
		}
		if sourceButton = buttonWithSVG(e.AsPanel(), svg.Database); sourceButton != nil {
			sourceButton.RequestFocus()
			focused = wnd.CurrentFocus() == sourceButton.AsPanel()
		}
		e.scroll.SetPosition(0, 0)
		listInView = !visibleRect(e.meleeWeapons.table.AsPanel()).Empty()
	})
	if sourceButton == nil {
		t.Fatal("the editor has no source button")
	}
	if modifiers == nil {
		t.Fatal("the editor has no list of modifiers")
	}
	c.True(focused, "precondition: the source button holds the focus")
	c.False(listInView, "precondition: the list of weapons is out of view")
	uxtest.SoleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
	uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen, isModifierEditor)
	chooseSyncWithSource(t, screen, wnd, e)
	var modifierEditors int
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		modifierEditors = len(AllMatchingDockables(isModifierEditor))
		focused = wnd.CurrentFocus() == sourceButton.AsPanel()
		hAfter, vAfter = e.scroll.Position()
	})
	c.Equal(0, weaponEditors, "an unchanged weapon's editor closes without asking")
	c.Equal(0, modifierEditors, "as does an unchanged modifier's")
	c.Equal("Claws", name, "and the trait is synced")
	c.True(focused, "the focus is back on the source button")
	c.Equal(float32(0), hAfter, "and the editor scrolled as it was")
	c.Equal(float32(0), vAfter)

	// A field outside the content, as the toolbar's scale field is, gets its caret back along with the focus.
	var scaleField *PercentageField
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		e.meleeWeapons.table.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
		if fields := uxtest.PanelsOfType[*PercentageField](e.Children()[0]); len(fields) == 1 {
			scaleField = fields[0]
			scaleField.RequestFocus()
			scaleField.SetSelection(1, 1)
			focused = wnd.CurrentFocus() == scaleField.AsPanel()
		}
	})
	if scaleField == nil {
		t.Fatal("the editor's toolbar has no scale field")
	}
	c.True(focused, "precondition: the scale field holds the focus")
	chooseSyncWithSource(t, screen, wnd, e)
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focused = wnd.CurrentFocus() == scaleField.AsPanel()
		selStart, selEnd = scaleField.Selection()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focused, "the focus is back on the scale field")
	c.Equal(1, selStart, "with the caret where it was, rather than all of its text selected")
	c.Equal(1, selEnd)
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncTakesTheFocusFromAClosedSubEditor verifies that syncing while the focus is within a sub-editor, which
// the sync closes, leaves the focus within the editor, where Escape and Cmd-Return still reach it: on the list the
// sub-editor was opened from, as rebuilt, or on the editor's first control when closing the sub-editor gave the focus
// to something else.
func TestEditorSyncTakesTheFocusFromAClosedSubEditor(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, func(local *gurps.Trait) {
		local.Weapons = []*gurps.Weapon{gurps.NewWeapon(local, true)}
		local.Weapons[0].Usage = "Thrust"
	})
	isWeaponEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.Weapon, *gurps.Weapon])
		return isEditor
	}
	// typeIntoWeaponEditor types into the first field of the one weapon editor open, leaving the focus there.
	typeIntoWeaponEditor := func() {
		t.Helper()
		weaponEditor := uxtest.SoleEditor[*editor[*gurps.Weapon, *gurps.Weapon]](t, screen, isWeaponEditor)
		var field *StringField
		screen.Do(func() {
			if fields := uxtest.PanelsOfType[*StringField](weaponEditor.content); len(fields) != 0 {
				field = fields[0]
			}
		})
		if field == nil {
			t.Fatal("the weapon's editor has no field to type into")
		}
		screen.Click(screen.PanelCenter(field))
		screen.Type("Swung")
		var focused, modified bool
		screen.Do(func() {
			focused = wnd.CurrentFocus() == field.AsPanel()
			modified = weaponEditor.isModified()
		})
		c.True(focused, "precondition: the focus is within the weapon's editor")
		c.True(modified, "precondition: which has changes to save")
	}
	// syncDiscardingWeaponChanges chooses Sync with Source, then agrees to the weapon editor's prompt to discard its
	// changes.
	syncDiscardingWeaponChanges := func() {
		t.Helper()
		chooseSyncWithSource(t, screen, wnd, e)
		_, dialog := modalDialog(t, screen, wnd)
		screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	}

	// Opened as from the list of weapons, which holds the focus as the weapon's editor opens.
	var oldTable *unison.Panel
	screen.Do(func() {
		oldTable = e.meleeWeapons.table.AsPanel()
		oldTable.RequestFocus()
		EditWeapon(e, e.editorData.Weapons[0])
	})
	typeIntoWeaponEditor()
	syncDiscardingWeaponChanges()
	var name string
	var weaponEditors int
	var focusOnList, listInView bool
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		table := e.meleeWeapons.table.AsPanel()
		focusOnList = table != oldTable && wnd.CurrentFocus() == table
		listInView = !visibleRect(table).Empty()
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focusOnList, "the focus is on the list the weapon's editor was opened from, as rebuilt")
	c.True(listInView, "which is in view")

	// Opened while the sheet holds the focus, which is what closing the weapon's editor then gives it back to.
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		wnd.SetFocus(sheet)
		EditWeapon(e, e.editorData.Weapons[0])
	})
	typeIntoWeaponEditor()
	syncDiscardingWeaponChanges()
	var focusInContent bool
	screen.Do(func() {
		name = e.editorData.Name
		weaponEditors = len(AllMatchingDockables(isWeaponEditor))
		focus := wnd.CurrentFocus()
		focusInContent = focus != nil && unison.AncestorIsOrSelf(focus, e.content)
	})
	c.Equal(0, weaponEditors, "the weapon's editor is closed")
	c.Equal("Claws", name, "and the trait synced")
	c.True(focusInContent, "the focus is within the editor's content, not on the sheet")

	screen.KeyPress(unison.KeyEscape, 0)
	var open bool
	screen.Do(func() {
		open = slices.ContainsFunc(AllDockables(), func(d unison.Dockable) bool { return d.AsPanel().Self == e })
	})
	c.False(open, "Escape reaches the editor, which discards its changes and closes")
}

// TestEditorSyncKeepsSavedModifierChanges verifies that syncing while a modifier's editor has changes asks whether to
// save them, as closing it any other way does, and that the sync keeps what was saved, since it leaves the modifiers
// alone.
func TestEditorSyncKeepsSavedModifierChanges(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, nil, func(local *gurps.Trait) {
		modifier := gurps.NewTraitModifier(local.DataOwner(), nil, false)
		modifier.Name = "Long"
		local.Modifiers = []*gurps.TraitModifier{modifier}
		gurps.AttachModifiers(local, local.Modifiers)
	})
	isModifierEditor := func(d unison.Dockable) bool {
		_, isEditor := d.AsPanel().Self.(*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData])
		return isEditor
	}
	screen.Do(func() { EditTraitModifier(e, e.editorData.Modifiers[0]) })
	modifierEditor := uxtest.SoleEditor[*editor[*gurps.TraitModifier, *gurps.TraitModifierEditData]](t, screen,
		isModifierEditor)
	screen.Do(func() {
		modifierEditor.editorData.Name = "Longer"
		modifierEditor.MarkModified(nil)
	})

	chooseSyncWithSource(t, screen, wnd, e)
	dialogWnd, dialog := modalDialog(t, screen, wnd)
	var prompt []string
	screen.Do(func() { prompt = uxtest.LabelTexts(dialogWnd.Content()) })
	c.True(slices.Contains(prompt, "Save changes made to"), "the prompt asks whether to save the modifier's "+
		"changes: %v", prompt)
	screen.Click(screen.PanelCenter(dialogButton(t, screen, dialog, unison.ModalResponseOK)))
	var name, modifierName string
	var modifierEditors int
	screen.Do(func() {
		name = e.editorData.Name
		modifierEditors = len(AllMatchingDockables(isModifierEditor))
		if len(e.editorData.Modifiers) == 1 {
			modifierName = e.editorData.Modifiers[0].Name
		}
	})
	c.Equal(0, modifierEditors, "saving closes the modifier's editor")
	c.Equal("Claws", name, "and the trait is synced")
	c.Equal("Longer", modifierName, "keeping the modifier as it was saved")
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncKeepsTheSelectionOfTheFocusedList verifies that a sync leaves the rows of the list that holds the focus
// selected in the list as rebuilt: the same modifiers, which a sync leaves alone, and the weapons in the same places,
// since a sync replaces the weapons with the library's.
func TestEditorSyncKeepsTheSelectionOfTheFocusedList(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
		lib.Weapons = append(lib.Weapons, gurps.NewWeapon(lib, true))
		lib.Weapons[1].Usage = "Swung"
	}, func(local *gurps.Trait) {
		for _, name := range []string{"Long", "Sharp"} {
			modifier := gurps.NewTraitModifier(local.DataOwner(), nil, false)
			modifier.Name = name
			local.Modifiers = append(local.Modifiers, modifier)
		}
		gurps.AttachModifiers(local, local.Modifiers)
	})
	modifiersTable := func() *unison.Table[*Node[*gurps.TraitModifier]] {
		if panels := uxtest.PanelsOfType[*traitModifiersPanel](e.content); len(panels) == 1 {
			return panels[0].table
		}
		return nil
	}
	var oldModifiers *unison.Table[*Node[*gurps.TraitModifier]]
	var focused bool
	screen.Do(func() {
		if oldModifiers = modifiersTable(); oldModifiers != nil {
			oldModifiers.RequestFocus()
			oldModifiers.SelectByIndex(1)
			focused = wnd.CurrentFocus() == oldModifiers.AsPanel()
		}
	})
	if oldModifiers == nil {
		t.Fatal("the editor has no list of modifiers")
	}
	c.True(focused, "precondition: the list of modifiers holds the focus")

	chooseSyncWithSource(t, screen, wnd, e)
	var name, selected string
	var count int
	screen.Do(func() {
		name = e.editorData.Name
		table := modifiersTable()
		focused = table != nil && table != oldModifiers && wnd.CurrentFocus() == table.AsPanel()
		if table != nil {
			count = table.SelectionCount()
			if rows := table.SelectedRows(false); len(rows) == 1 {
				selected = rows[0].Data().Name
			}
		}
	})
	c.Equal("Claws", name, "the trait is synced")
	c.True(focused, "the focus is on the list of modifiers, as rebuilt")
	c.Equal(1, count, "with one modifier selected")
	c.Equal("Sharp", selected, "the one that was")

	var oldWeapons *unison.Table[*Node[*gurps.Weapon]]
	var oldID tid.TID
	screen.Do(func() {
		e.editorData.Name = "Claws (changed)"
		oldWeapons = e.meleeWeapons.table
		oldWeapons.RequestFocus()
		oldWeapons.SelectByIndex(1)
		focused = wnd.CurrentFocus() == oldWeapons.AsPanel()
		if rows := oldWeapons.SelectedRows(false); len(rows) == 1 {
			oldID = rows[0].Data().TID
		}
	})
	c.True(focused, "precondition: the list of weapons holds the focus")
	c.True(oldID != "", "precondition: with its second weapon selected")

	chooseSyncWithSource(t, screen, wnd, e)
	var usage string
	var replaced, second bool
	screen.Do(func() {
		name = e.editorData.Name
		table := e.meleeWeapons.table
		focused = table != oldWeapons && wnd.CurrentFocus() == table.AsPanel()
		count = table.SelectionCount()
		second = table.IsRowSelected(1)
		if rows := table.SelectedRows(false); len(rows) == 1 {
			usage = rows[0].Data().Usage
			replaced = rows[0].Data().TID != oldID
		}
	})
	c.Equal("Claws", name, "the trait is synced again")
	c.True(focused, "the focus is on the list of weapons, as rebuilt")
	c.Equal(1, count, "with one weapon selected")
	c.True(second, "the second")
	c.Equal("Swung", usage)
	c.True(replaced, "which is the library's weapon in place of the one that was selected")
	screen.Click(screen.PanelCenter(e.cancelButton))
}

// TestEditorSyncKeepsSectionsAsTheyWere verifies that a sync shows the sections of rows as they were before it, each
// collapsed or expanded as the user left it rather than as a new section starts out, with the same row open, so that
// the field in that row that held the focus gets it back.
func TestEditorSyncKeepsSectionsAsTheyWere(t *testing.T) {
	c := check.New(t)
	screen, wnd := uxtest.StartHeadlessWorkspace(t, c)
	sheet, ok := uxtest.OpenedByAction(t, screen, newCharacterSheetAction).(*Sheet)
	if !ok {
		t.Fatal("New Character Sheet must open a character sheet")
	}
	// Prerequisites, so that their section starts out collapsed, and no features, so that theirs starts out expanded.
	e, _ := newEditedLibraryTrait(t, c, screen, sheet, func(lib *gurps.Trait) {
		trait := gurps.NewTraitPrereq()
		trait.NameCriteria.Qualifier = "Magery"
		lib.Prereq = gurps.NewPrereqList()
		lib.Prereq.Prereqs = gurps.Prereqs{trait}
		lib.Prereq = lib.Prereq.CloneAsPrereqList(nil)
	}, nil)
	sections := func() (prereqs *prereqPanel, features *featuresPanel) {
		if panels := uxtest.PanelsOfType[*prereqPanel](e.content); len(panels) == 1 {
			prereqs = panels[0]
		}
		if panels := uxtest.PanelsOfType[*featuresPanel](e.content); len(panels) == 1 {
			features = panels[0]
		}
		return prereqs, features
	}
	var oldPrereqs *prereqPanel
	var oldFeatures *featuresPanel
	screen.Do(func() { oldPrereqs, oldFeatures = sections() })
	if oldPrereqs == nil || oldFeatures == nil {
		t.Fatal("the editor must have a prerequisites section and a features section")
	}
	screen.Do(func() {
		c.True(oldPrereqs.collapse.collapsed, "precondition: the prerequisites start out collapsed")
		c.False(oldFeatures.collapse.collapsed, "precondition: the features start out expanded")
		oldPrereqs.collapse.toggle()
		oldFeatures.collapse.toggle()
	})
	screen.Do(func() { oldPrereqs.toggle("r.0") })
	var field *unison.Panel
	screen.Do(func() {
		if field = oldPrereqs.FindRefKey("r.0:name"); field != nil {
			field.RequestFocus()
		}
	})
	if field == nil {
		t.Fatal("the open prerequisite must have a name field")
	}
	var focused bool
	screen.Do(func() { focused = wnd.CurrentFocus() == field })
	c.True(focused, "precondition: the prerequisite's name field holds the focus")

	chooseSyncWithSource(t, screen, wnd, e)
	screen.Do(func() {
		c.Equal("Claws", e.editorData.Name, "the trait is synced")
		prereqs, features := sections()
		if prereqs == nil || features == nil || prereqs == oldPrereqs || features == oldFeatures {
			t.Error("the sync must rebuild both sections")
			return
		}
		c.False(prereqs.collapse.collapsed, "the prerequisites stay expanded")
		c.Equal("r.0", prereqs.open, "with the same row open")
		rebuilt := prereqs.FindRefKey("r.0:name")
		c.True(rebuilt != nil && rebuilt != field && wnd.CurrentFocus() == rebuilt,
			"and the focus on its rebuilt name field")
		c.True(features.collapse.collapsed, "the features stay collapsed")
		c.NotNil(features.FindRefKey(sectionSummaryKey), "showing their paragraph")
	})
	screen.Click(screen.PanelCenter(e.cancelButton))
}
