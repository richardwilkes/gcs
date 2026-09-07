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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/check"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

// modifierProviderCase names what a modifier table provider is expected to say about itself, so that the trait and
// equipment modifier providers, which share one implementation, can be checked by the same code.
type modifierProviderCase[T gurps.Node[T]] struct {
	name        string
	provider    func(forEditor bool) TableProvider[T]
	refKey      string
	dragKey     *uti.DataType
	dragSVG     *unison.SVG
	singular    string
	plural      string
	enabled     int
	description int
	columns     []int
	libSrc      int
	newActions  []*unison.Action
}

func checkModifierProvider[T gurps.Node[T]](t *testing.T, one modifierProviderCase[T]) {
	t.Run(one.name, func(t *testing.T) {
		c := check.New(t)
		for _, forEditor := range []bool{false, true} {
			p := one.provider(forEditor)
			c.Equal(one.refKey, p.RefKey())
			c.True(one.dragKey == p.DragKey(), "drag key must be the %s key", one.name)
			c.True(one.dragSVG == p.DragSVG(), "drag image must be the %s image", one.name)
			singular, plural := p.ItemNames()
			c.Equal(one.singular, singular)
			c.Equal(one.plural, plural)
			c.Equal(one.description, p.HierarchyColumnID())
			c.Equal(one.description, p.ExcessWidthColumnID())
			expected := one.columns
			if forEditor {
				expected = append(append([]int{one.enabled}, one.columns...), one.libSrc)
			}
			c.Equal(expected, p.ColumnIDs(), "columns for forEditor=%v", forEditor)
			items := p.ContextMenuItems()
			c.True(len(items) > len(one.newActions), "the provider's own items precede the shared ones")
			for i, action := range one.newActions {
				c.Equal(action.ID, items[i].ID, "context menu item %d", i)
			}
		}
	})
}

// TestModifierProvidersUseTheirOwnSpec verifies that the trait and equipment modifier table providers each describe
// their own kind of modifier: the keys, images and names, the description column that carries the hierarchy, the
// enabled and library source columns that only an editor's table adds around the fixed run of columns, and the
// new-item actions that lead the context menu.
func TestModifierProvidersUseTheirOwnSpec(t *testing.T) {
	checkModifierProvider(t, modifierProviderCase[*gurps.TraitModifier]{
		name: "trait modifiers",
		provider: func(forEditor bool) TableProvider[*gurps.TraitModifier] {
			return NewTraitModifiersProvider(&traitModifierListProvider{}, forEditor)
		},
		refKey:      traitModifierRefKey,
		dragKey:     traitModifierDragKey,
		dragSVG:     svg.GCSTraitModifiers,
		singular:    "Trait Modifier",
		plural:      "Trait Modifiers",
		enabled:     gurps.TraitModifierEnabledColumn,
		description: gurps.TraitModifierDescriptionColumn,
		columns: []int{
			gurps.TraitModifierDescriptionColumn,
			gurps.TraitModifierCostColumn,
			gurps.TraitModifierTagsColumn,
			gurps.TraitModifierReferenceColumn,
		},
		libSrc:     gurps.TraitModifierLibSrcColumn,
		newActions: []*unison.Action{newTraitModifierAction, newTraitContainerModifierAction},
	})
	checkModifierProvider(t, modifierProviderCase[*gurps.EquipmentModifier]{
		name: "equipment modifiers",
		provider: func(forEditor bool) TableProvider[*gurps.EquipmentModifier] {
			return NewEquipmentModifiersProvider(&equipmentModifierListProvider{}, forEditor)
		},
		refKey:      equipmentModifierRefKey,
		dragKey:     equipmentModifierDragKey,
		dragSVG:     svg.GCSEquipmentModifiers,
		singular:    "Equipment Modifier",
		plural:      "Equipment Modifiers",
		enabled:     gurps.EquipmentModifierEnabledColumn,
		description: gurps.EquipmentModifierDescriptionColumn,
		columns: []int{
			gurps.EquipmentModifierDescriptionColumn,
			gurps.EquipmentModifierTechLevelColumn,
			gurps.EquipmentModifierCostColumn,
			gurps.EquipmentModifierWeightColumn,
			gurps.EquipmentModifierTagsColumn,
			gurps.EquipmentModifierReferenceColumn,
		},
		libSrc:     gurps.EquipmentModifierLibSrcColumn,
		newActions: []*unison.Action{newEquipmentModifierAction, newEquipmentContainerModifierAction},
	})
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
