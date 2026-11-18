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
	"github.com/richardwilkes/gcs/v5/ux/svg"
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
	// The shared context menu items name actions that registerActions creates.
	registerKeyBindingsOnce.Do(func() { registerActions() })
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
