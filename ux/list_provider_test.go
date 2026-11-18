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

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/unison"
)

// listsForTest is a bare set of the lists the providers under test are built on, so the tests can see exactly what
// each provider reads and writes.
type listsForTest struct {
	traits  []*gurps.Trait
	notes   []*gurps.Note
	carried []*gurps.Equipment
	other   []*gurps.Equipment
	melee   []*gurps.Weapon
	ranged  []*gurps.Weapon
}

func (l *listsForTest) DataOwner() gurps.DataOwner                      { return nil }
func (l *listsForTest) TraitList() []*gurps.Trait                       { return l.traits }
func (l *listsForTest) SetTraitList(list []*gurps.Trait)                { l.traits = list }
func (l *listsForTest) NoteList() []*gurps.Note                         { return l.notes }
func (l *listsForTest) SetNoteList(list []*gurps.Note)                  { l.notes = list }
func (l *listsForTest) CarriedEquipmentList() []*gurps.Equipment        { return l.carried }
func (l *listsForTest) SetCarriedEquipmentList(list []*gurps.Equipment) { l.carried = list }
func (l *listsForTest) OtherEquipmentList() []*gurps.Equipment          { return l.other }
func (l *listsForTest) SetOtherEquipmentList(list []*gurps.Equipment)   { l.other = list }
func (l *listsForTest) WeaponOwner() gurps.WeaponOwner                  { return nil }

func (l *listsForTest) Weapons(melee, _, _ bool) []*gurps.Weapon {
	if melee {
		return l.melee
	}
	return l.ranged
}

func (l *listsForTest) SetWeapons(melee bool, list []*gurps.Weapon) {
	if melee {
		l.melee = list
	} else {
		l.ranged = list
	}
}

func newTraitForTest(name string, parent *gurps.Trait, tags ...string) *gurps.Trait {
	t := gurps.NewTrait(nil, parent, false)
	t.Name = name
	t.Tags = tags
	return t
}

func TestListProviderRowsMirrorTheList(t *testing.T) {
	c := check.New(t)
	notes := []*gurps.Note{gurps.NewNote(nil, nil, false), gurps.NewNote(nil, nil, true)}
	for _, forPage := range []bool{false, true} {
		// A provider for a page asks its owner for the entity, so the page case is given a real one.
		var owner gurps.NoteListProvider = &listsForTest{notes: notes}
		if forPage {
			entity := gurps.NewEntity()
			entity.Notes = notes
			owner = entity
		}
		p := NewNotesProvider(owner, forPage)
		table := unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.Note]]{})
		p.SetTable(table)
		c.Equal(2, p.RootRowCount())
		c.Equal(notes, p.RootData())
		rows := p.RootRows()
		c.Equal(2, len(rows))
		for i, row := range rows {
			c.True(notes[i] == row.Data(), "row %d must wrap the note in the same position", i)
			c.True(table == row.table, "row %d must belong to the provider's table", i)
			c.Equal(forPage, row.forPage, "row %d must be built for the same context as the provider", i)
		}
		headers := p.Headers()
		c.Equal(len(p.ColumnIDs()), len(headers))
		for i, header := range headers {
			_, isPageHeader := header.(*PageTableColumnHeader[*gurps.Note])
			c.Equal(forPage, isPageHeader, "header %d must be built for the same context as the provider", i)
		}
	}
	lists := &listsForTest{notes: notes}
	p := NewNotesProvider(lists, false)
	p.SetTable(unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.Note]]{}))
	replacement := []*gurps.Note{gurps.NewNote(nil, nil, false)}
	p.SetRootData(replacement)
	c.Equal(replacement, lists.notes, "SetRootData must write through to the list")
	rows := p.RootRows()
	p.SetRootRows(rows[:0])
	c.Equal(0, len(lists.notes), "SetRootRows must write the rows' data through to the list")
}

func TestListProviderSerializationRoundTrips(t *testing.T) {
	c := check.New(t)
	source := &listsForTest{traits: []*gurps.Trait{newTraitForTest("one", nil, "x"), newTraitForTest("two", nil)}}
	data, err := NewTraitsProvider(source, false).Serialize()
	c.NoError(err)
	target := &listsForTest{}
	c.NoError(NewTraitsProvider(target, false).Deserialize(data))
	c.Equal(2, len(target.traits))
	for i, trait := range target.traits {
		c.Equal(source.traits[i].Name, trait.Name)
		c.Equal(source.traits[i].Tags, trait.Tags)
	}
	c.HasError(NewTraitsProvider(target, false).Deserialize([]byte("not compressed json")))
}

// referenceColumnsForTest names a provider's column list along with the reference and library source columns it may
// show, so that the same checks can run over every kind of list.
type referenceColumnsForTest struct {
	name    string
	columns func() []int
	ref     int
	src     int
}

func referenceColumnProvidersForTest(lists gurps.ListProvider, forPage bool) []referenceColumnsForTest {
	return []referenceColumnsForTest{
		{"traits", NewTraitsProvider(lists, forPage).ColumnIDs, gurps.TraitReferenceColumn, gurps.TraitLibSrcColumn},
		{"skills", NewSkillsProvider(lists, forPage).ColumnIDs, gurps.SkillReferenceColumn, gurps.SkillLibSrcColumn},
		{"spells", NewSpellsProvider(lists, forPage).ColumnIDs, gurps.SpellReferenceColumn, gurps.SpellLibSrcColumn},
		{
			"equipment", NewEquipmentProvider(lists, true, forPage).ColumnIDs,
			gurps.EquipmentReferenceColumn, gurps.EquipmentLibSrcColumn,
		},
		{"notes", NewNotesProvider(lists, forPage).ColumnIDs, gurps.NoteReferenceColumn, gurps.NoteLibSrcColumn},
	}
}

// TestProvidersFollowReferenceColumnSettings verifies that the reference and library source columns of every list on a
// page come and go with the settings that hide them -- the entity's own on a sheet, the global ones on a template --
// while off a page the reference column is always present and the source column never is.
func TestProvidersFollowReferenceColumnSettings(t *testing.T) {
	c := check.New(t)
	global := gurps.GlobalSettings().SheetSettings()
	savedRef, savedSrc := global.HidePageRefColumn, global.HideSourceMismatch
	t.Cleanup(func() { global.HidePageRefColumn, global.HideSourceMismatch = savedRef, savedSrc })
	entity := gurps.NewEntity()
	template := gurps.NewTemplate()
	for _, hideRef := range []bool{false, true} {
		for _, hideSrc := range []bool{false, true} {
			entity.SheetSettings.HidePageRefColumn = hideRef
			entity.SheetSettings.HideSourceMismatch = hideSrc
			// The global settings are set to the opposite, so a sheet that read them instead of its own would fail.
			global.HidePageRefColumn = !hideRef
			global.HideSourceMismatch = !hideSrc
			for _, p := range referenceColumnProvidersForTest(entity, true) {
				c.Equal(!hideRef, slices.Contains(p.columns(), p.ref),
					"%s on a sheet: the reference column must follow the sheet's own setting (hide=%v)", p.name, hideRef)
				c.Equal(!hideSrc, slices.Contains(p.columns(), p.src),
					"%s on a sheet: the source column must follow the sheet's own setting (hide=%v)", p.name, hideSrc)
			}
			for _, p := range referenceColumnProvidersForTest(template, true) {
				c.Equal(hideRef, slices.Contains(p.columns(), p.ref),
					"%s on a template: the reference column must follow the global setting (hide=%v)", p.name, !hideRef)
				c.Equal(hideSrc, slices.Contains(p.columns(), p.src),
					"%s on a template: the source column must follow the global setting (hide=%v)", p.name, !hideSrc)
			}
			for _, p := range referenceColumnProvidersForTest(entity, false) {
				c.True(slices.Contains(p.columns(), p.ref), "%s off a page must always show the reference column", p.name)
				c.False(slices.Contains(p.columns(), p.src), "%s off a page must never show the source column", p.name)
			}
		}
	}
}

// TestRootRowsReturnsTheTablesOwnRows verifies that a page list's provider hands out the rows the table already shows
// rather than new ones, while a provider without reuse, as off a page, builds new rows on every call.
func TestRootRowsReturnsTheTablesOwnRows(t *testing.T) {
	c := check.New(t)
	sheet, traits := newReuseTestSheet(t, -1, "Alpha", "Beta")
	table := sheet.Traits.Table
	roots := table.RootRows()
	c.Equal(2, len(roots), "the table must have both traits at the top")
	c.True(roots[0] == table.RowFromIndex(0) && roots[1] == table.RowFromIndex(1),
		"the provider must return the rows the table shows")

	lists := &listsForTest{traits: traits}
	for _, p := range []TableProvider[*gurps.Trait]{NewTraitsProvider(lists, false), NewTraitsProvider(lists, true)} {
		if tp, ok := p.(*traitsProvider); ok {
			// Reuse switched off for the page provider too, which must then behave as one off a page.
			tp.reuse = nil
		}
		p.SetTable(unison.NewTable[*Node[*gurps.Trait]](p))
		first, second := p.RootRows(), p.RootRows()
		c.Equal(2, len(first), "the provider must return a row for each trait")
		for i := range first {
			c.True(first[i].Data() == traits[i] && second[i].Data() == traits[i], "row %d must hold its trait", i)
			c.True(first[i] != second[i], "without reuse, row %d must be built anew on each call", i)
		}
	}
}

// TestFileListProvidersDetachDataOwners verifies that every list a list file's dockable is built on detaches the nodes
// it is handed from the data owner they came with, since a list file has none. The two modifier lists used to keep
// whatever owner a modifier arrived with.
func TestFileListProvidersDetachDataOwners(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	notes := &noteListProvider{}
	notes.SetNoteList([]*gurps.Note{gurps.NewNote(entity, nil, false)})
	c.Nil(notes.DataOwner())
	c.Nil(notes.NoteList()[0].DataOwner(), "a note must be detached from its entity")

	skills := &skillListProvider{}
	skills.SetSkillList([]*gurps.Skill{gurps.NewSkill(entity, nil, false)})
	c.Nil(skills.DataOwner())
	c.Nil(skills.SkillList()[0].DataOwner(), "a skill must be detached from its entity")

	spells := &spellListProvider{}
	spells.SetSpellList([]*gurps.Spell{gurps.NewSpell(entity, nil, false)})
	c.Nil(spells.DataOwner())
	c.Nil(spells.SpellList()[0].DataOwner(), "a spell must be detached from its entity")

	traits := &traitListProvider{}
	traits.SetTraitList([]*gurps.Trait{gurps.NewTrait(entity, nil, false)})
	c.Nil(traits.DataOwner())
	c.Nil(traits.TraitList()[0].DataOwner(), "a trait must be detached from its entity")

	traitMods := &traitModifierListProvider{}
	traitMods.SetTraitModifierList([]*gurps.TraitModifier{gurps.NewTraitModifier(entity, nil, false)})
	c.Nil(traitMods.DataOwner())
	c.Nil(traitMods.TraitModifierList()[0].DataOwner(), "a trait modifier must be detached from its entity")

	eqpMods := &equipmentModifierListProvider{}
	eqpMods.SetEquipmentModifierList([]*gurps.EquipmentModifier{gurps.NewEquipmentModifier(entity, nil, false)})
	c.Nil(eqpMods.DataOwner())
	c.Nil(eqpMods.EquipmentModifierList()[0].DataOwner(), "an equipment modifier must be detached from its entity")

	equipment := &equipmentListProvider{}
	equipment.SetCarriedEquipmentList([]*gurps.Equipment{gurps.NewEquipment(entity, nil, false)})
	equipment.SetOtherEquipmentList([]*gurps.Equipment{gurps.NewEquipment(entity, nil, false)})
	c.True(equipment.DataOwner() == gurps.DataOwner(equipment), "an equipment list file is its own data owner")
	c.Nil(equipment.CarriedEquipmentList()[0].DataOwner(), "carried equipment must be detached from its entity")
	c.Nil(equipment.OtherEquipmentList()[0].DataOwner(), "other equipment must be detached from its entity")
}

// createItemForTest runs the provider's CreateItem against a table of its own with the editor callback replaced by one
// that records the item it was handed, and returns that item.
func createItemForTest[T gurps.Node[T]](c check.Checker, provider TableProvider[T], p *listProvider[T], variant ItemVariant, kind string) T {
	var edited, zero T
	p.edit = func(_ Rebuildable, item T) { edited = item }
	_, table := NewNodeTable(provider, nil)
	provider.CreateItem(nil, table, variant)
	c.True(edited != zero, "%s: the new item must be handed to the editor", kind)
	return edited
}

// TestCreateItemVariants verifies the shared CreateItem: the item its variant asks for is created for the provider's
// data owner, added to the end of the list and handed to the editor, and the providers with an alternate variant
// create their own kind of item for it while going through the same tail.
func TestCreateItemVariants(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()

	notes, ok := NewNotesProvider(entity, false).(*notesProvider)
	c.True(ok, "the notes provider must be the expected type")
	for i, variant := range []ItemVariant{NoItemVariant, ContainerItemVariant} {
		item := createItemForTest(c, notes, &notes.listProvider, variant, "note")
		c.Equal(i+1, len(entity.Notes), "each note must be added to the list")
		c.True(entity.Notes[i] == item, "note %d must be the one handed to the editor", i)
		c.True(entity.DataOwner() == item.DataOwner(), "note %d must belong to the provider's owner", i)
		c.Equal(variant == ContainerItemVariant, item.Container(), "note %d", i)
	}

	skills, ok := NewSkillsProvider(entity, false).(*skillsProvider)
	c.True(ok, "the skills provider must be the expected type")
	for i, variant := range []ItemVariant{NoItemVariant, ContainerItemVariant, AlternateItemVariant} {
		item := createItemForTest(c, skills, &skills.listProvider, variant, "skill")
		c.Equal(i+1, len(entity.Skills), "each skill must be added to the list")
		c.True(entity.Skills[i] == item, "skill %d must be the one handed to the editor", i)
		c.True(entity.DataOwner() == item.DataOwner(), "skill %d must belong to the provider's owner", i)
		c.Equal(variant == ContainerItemVariant, item.Container(), "skill %d", i)
		c.Equal(variant == AlternateItemVariant, item.IsTechnique(), "skill %d", i)
	}

	spells, ok := NewSpellsProvider(entity, false).(*spellsProvider)
	c.True(ok, "the spells provider must be the expected type")
	for i, variant := range []ItemVariant{NoItemVariant, ContainerItemVariant, AlternateItemVariant} {
		item := createItemForTest(c, spells, &spells.listProvider, variant, "spell")
		c.Equal(i+1, len(entity.Spells), "each spell must be added to the list")
		c.True(entity.Spells[i] == item, "spell %d must be the one handed to the editor", i)
		c.Equal(variant == ContainerItemVariant, item.Container(), "spell %d", i)
		c.Equal(variant == AlternateItemVariant, item.IsRitualMagic(), "spell %d", i)
	}

	equipment, ok := NewEquipmentProvider(entity, true, false).(*equipmentProvider)
	c.True(ok, "the equipment provider must be the expected type")
	group := createItemForTest(c, equipment, &equipment.listProvider, GroupContainerItemVariant, "equipment")
	c.Equal([]*gurps.Equipment{group}, entity.CarriedEquipment, "the group must be added to the list")
	c.True(group != nil && group.IsGroup() && !gurps.IsTemplateChoiceContainer(group),
		"the group variant must create a group container, not a physical container or a choice")

	mods := &traitModifierListProvider{}
	traitMods, ok := NewTraitModifiersProvider(mods, false).(*modifiersProvider[*gurps.TraitModifier])
	c.True(ok, "the traitMods provider must be the expected type")
	for i, variant := range []ItemVariant{NoItemVariant, ContainerItemVariant} {
		item := createItemForTest(c, traitMods, &traitMods.listProvider, variant, "trait modifier")
		c.Equal(i+1, len(mods.TraitModifierList()), "each trait modifier must be added to the list")
		c.True(mods.TraitModifierList()[i] == item, "trait modifier %d must be the one handed to the editor", i)
		c.Equal(variant == ContainerItemVariant, item.Container(), "trait modifier %d", i)
	}
}
