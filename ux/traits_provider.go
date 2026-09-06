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

var _ TableProvider[*gurps.Trait] = &traitsProvider{}

type traitsProvider struct {
	listProvider[*gurps.Trait]
	provider gurps.TraitListProvider
}

// NewTraitsProvider creates a new table provider for traits.
func NewTraitsProvider(provider gurps.TraitListProvider, forPage bool) TableProvider[*gurps.Trait] {
	p := &traitsProvider{provider: provider}
	p.listProvider = listProvider[*gurps.Trait]{
		dataOwner:  provider,
		list:       provider.TraitList,
		setList:    provider.SetTraitList,
		columnIDs:  p.ColumnIDs,
		headerData: gurps.TraitsHeaderData,
		forPage:    forPage,
	}
	return p
}

func (p *traitsProvider) RefKey() string {
	return gurps.BlockTraitsKey
}

func (p *traitsProvider) DragKey() *uti.DataType {
	return traitDragKey
}

func (p *traitsProvider) DragSVG() *unison.SVG {
	return svg.GCSTraits
}

func (p *traitsProvider) AltDropSupport() *AltDropSupport {
	return modifierAltDropSupport(&p.listProvider, traitModifierDragKey,
		func(target *gurps.Trait, clones []*gurps.TraitModifier) {
			target.Modifiers = append(target.Modifiers, clones...)
		})
}

func (p *traitsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait"), i18n.Text("Traits")
}

func (p *traitsProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 5)
	if showSwitchColumn(p.forPage, p.provider, p.RootData()) {
		columnIDs = append(columnIDs, gurps.TraitSwitchColumn)
	}
	columnIDs = append(
		columnIDs,
		gurps.TraitDescriptionColumn,
		gurps.TraitPointsColumn,
	)
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.TraitTagsColumn)
	}
	return p.appendReferenceColumns(columnIDs, gurps.TraitReferenceColumn, gurps.TraitLibSrcColumn)
}

func (p *traitsProvider) HierarchyColumnID() int {
	return gurps.TraitDescriptionColumn
}

func (p *traitsProvider) ExcessWidthColumnID() int {
	return gurps.TraitDescriptionColumn
}

func (p *traitsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	OpenEditor(table, func(item *gurps.Trait) { EditTrait(owner, item) })
}

func (p *traitsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]], variant ItemVariant) {
	item := gurps.NewTrait(p.DataOwner(), nil, variant == ContainerItemVariant)
	p.insertItems(owner, table, item)
	EditTrait(owner, item)
}

func (p *traitsProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		{
			Title: i18n.Text("New Trait"),
			ID:    NewTraitItemID,
		},
		{
			Title: i18n.Text("New Trait Container"),
			ID:    NewTraitContainerItemID,
		},
		{
			Title: i18n.Text("Add Natural Attacks"),
			ID:    AddNaturalAttacksItemID,
		},
		{
			Title: organizeTraitsAction.Title,
			ID:    OrganizeTraitsItemID,
		},
	})
}
