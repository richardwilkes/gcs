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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

const traitModifierRefKey = "trait_modifier"

var _ TableProvider[*gurps.TraitModifier] = &traitModifiersProvider{}

type traitModifiersProvider struct {
	listProvider[*gurps.TraitModifier]
	forEditor bool
}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider gurps.TraitModifierListProvider, forEditor bool) TableProvider[*gurps.TraitModifier] {
	p := &traitModifiersProvider{forEditor: forEditor}
	p.listProvider = listProvider[*gurps.TraitModifier]{
		dataOwner:  provider,
		list:       provider.TraitModifierList,
		setList:    provider.SetTraitModifierList,
		columnIDs:  p.ColumnIDs,
		headerData: gurps.TraitModifierHeaderData,
	}
	return p
}

func (p *traitModifiersProvider) RefKey() string {
	return traitModifierRefKey
}

func (p *traitModifiersProvider) DragKey() *uti.DataType {
	return traitModifierDragKey
}

func (p *traitModifiersProvider) DragSVG() *unison.SVG {
	return svg.GCSTraitModifiers
}

func (p *traitModifiersProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait Modifier"), i18n.Text("Trait Modifiers")
}

func (p *traitModifiersProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 5)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.TraitModifierEnabledColumn)
	}
	columnIDs = append(
		columnIDs,
		gurps.TraitModifierDescriptionColumn,
		gurps.TraitModifierCostColumn,
		gurps.TraitModifierTagsColumn,
		gurps.TraitModifierReferenceColumn,
	)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.TraitModifierLibSrcColumn)
	}
	return columnIDs
}

func (p *traitModifiersProvider) HierarchyColumnID() int {
	return gurps.TraitModifierDescriptionColumn
}

func (p *traitModifiersProvider) ExcessWidthColumnID() int {
	return gurps.TraitModifierDescriptionColumn
}

func (p *traitModifiersProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]]) {
	OpenEditor(table, func(item *gurps.TraitModifier) {
		EditTraitModifier(owner, item)
	})
}

func (p *traitModifiersProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.TraitModifier]], variant ItemVariant) {
	item := gurps.NewTraitModifier(p.DataOwner(), nil, variant == ContainerItemVariant)
	p.insertItems(owner, table, item)
	EditTraitModifier(owner, item)
}

func (p *traitModifiersProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		{
			Title: i18n.Text("New Trait Modifier"),
			ID:    NewTraitModifierItemID,
		},
		{
			Title: i18n.Text("New Trait Modifier Container"),
			ID:    NewTraitContainerModifierItemID,
		},
	})
}
