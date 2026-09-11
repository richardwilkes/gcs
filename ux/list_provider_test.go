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

func TestEquipmentProviderUsesTheCarriedOrOtherList(t *testing.T) {
	c := check.New(t)
	lists := &listsForTest{
		carried: []*gurps.Equipment{gurps.NewEquipment(nil, nil, false), gurps.NewEquipment(nil, nil, false)},
		other:   []*gurps.Equipment{gurps.NewEquipment(nil, nil, false)},
	}
	carried := NewEquipmentProvider(lists, true, true)
	other := NewEquipmentProvider(lists, false, true)
	c.Equal(lists.carried, carried.RootData())
	c.Equal(lists.other, other.RootData())
	c.Equal(gurps.BlockEquipmentKey, carried.RefKey())
	c.Equal(gurps.BlockOtherEquipmentKey, other.RefKey())
	other.SetRootData(nil)
	c.Equal(0, len(lists.other), "the other provider must write to the other list")
	c.Equal(2, len(lists.carried), "the other provider must leave the carried list alone")
	carried.SetRootData(nil)
	c.Equal(0, len(lists.carried), "the carried provider must write to the carried list")
}

func TestWeaponsProviderOnlyEditsOffThePage(t *testing.T) {
	c := check.New(t)
	lists := &listsForTest{
		melee:  []*gurps.Weapon{gurps.NewWeapon(nil, true)},
		ranged: []*gurps.Weapon{gurps.NewWeapon(nil, false), gurps.NewWeapon(nil, false)},
	}
	c.Equal(lists.melee, NewWeaponsProvider(lists, true, false).RootData())
	c.Equal(lists.ranged, NewWeaponsProvider(lists, false, false).RootData())

	onPage := NewWeaponsProvider(lists, false, true)
	onPage.SetTable(unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.Weapon]]{}))
	_, err := onPage.Serialize()
	c.HasError(err, "the weapons on a page are not the list's own, so they must not be copied")
	c.HasError(onPage.Deserialize(nil), "the weapons on a page are not the list's own, so they must not be pasted")
	for i, header := range onPage.Headers() {
		c.False(header.SortState().Sortable, "weapon header %d must not be sortable", i)
	}

	inEditor := NewWeaponsProvider(lists, false, false)
	data, err := inEditor.Serialize()
	c.NoError(err)
	target := &listsForTest{}
	c.NoError(NewWeaponsProvider(target, false, false).Deserialize(data))
	c.Equal(2, len(target.ranged), "the ranged weapons must be pasted into the ranged list")
	c.Equal(0, len(target.melee), "the melee list must be left alone")
}

func TestCondModProviderIsReadOnly(t *testing.T) {
	c := check.New(t)
	lists := &condModListsForTest{conditional: []*gurps.ConditionalModifier{gurps.NewConditionalModifier("one", "", 0)}}
	p := NewConditionalModifiersProvider(lists)
	p.SetTable(unison.NewTable(&unison.SimpleTableModel[*Node[*gurps.ConditionalModifier]]{}))
	p.SetRootData(nil)
	p.SetRootRows(nil)
	c.Equal(1, len(lists.conditional), "the computed rows must not be replaceable")
	_, err := p.Serialize()
	c.HasError(err)
	c.HasError(p.Deserialize(nil))
	for i, header := range p.Headers() {
		c.False(header.SortState().Sortable, "conditional modifier header %d must not be sortable", i)
	}
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

// TestEquipmentProviderFollowsTLAndLCColumnSettings verifies that the equipment list's tech level and legality class
// columns follow the sheet's settings on a page and are always present off one.
func TestEquipmentProviderFollowsTLAndLCColumnSettings(t *testing.T) {
	c := check.New(t)
	entity := gurps.NewEntity()
	onPage := NewEquipmentProvider(entity, true, true)
	offPage := NewEquipmentProvider(entity, true, false)
	for _, hide := range []bool{false, true} {
		entity.SheetSettings.HideTLColumn = hide
		entity.SheetSettings.HideLCColumn = !hide
		c.Equal(!hide, slices.Contains(onPage.ColumnIDs(), gurps.EquipmentTLColumn),
			"on a page the TL column must follow the sheet's setting (hide=%v)", hide)
		c.Equal(hide, slices.Contains(onPage.ColumnIDs(), gurps.EquipmentLCColumn),
			"on a page the LC column must follow the sheet's setting (hide=%v)", !hide)
		c.True(slices.Contains(offPage.ColumnIDs(), gurps.EquipmentTLColumn), "off a page the TL column must be present")
		c.True(slices.Contains(offPage.ColumnIDs(), gurps.EquipmentLCColumn), "off a page the LC column must be present")
	}
}
