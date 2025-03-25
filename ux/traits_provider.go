// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox"
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

const traitDragKey = "trait"

var _ TableProvider[*gurps.Trait] = &traitsProvider{}

type traitsProvider struct {
	table    *unison.Table[*Node[*gurps.Trait]]
	provider gurps.TraitListProvider
	forPage  bool
}

// NewTraitsProvider creates a new table provider for traits.
func NewTraitsProvider(provider gurps.TraitListProvider, forPage bool) TableProvider[*gurps.Trait] {
	return &traitsProvider{
		provider: provider,
		forPage:  forPage,
	}
}

func (p *traitsProvider) RefKey() string {
	return gurps.BlockLayoutTraitsKey
}

func (p *traitsProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(trait *gurps.Trait) bool {
		for _, tag := range trait.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := dict.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *traitsProvider) SetTable(table *unison.Table[*Node[*gurps.Trait]]) {
	p.table = table
}

func (p *traitsProvider) RootRowCount() int {
	return len(p.provider.TraitList())
}

func (p *traitsProvider) RootRows() []*Node[*gurps.Trait] {
	data := p.provider.TraitList()
	rows := make([]*Node[*gurps.Trait], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *traitsProvider) SetRootRows(rows []*Node[*gurps.Trait]) {
	p.provider.SetTraitList(ExtractNodeDataFromList(rows))
}

func (p *traitsProvider) RootData() []*gurps.Trait {
	return p.provider.TraitList()
}

func (p *traitsProvider) SetRootData(data []*gurps.Trait) {
	p.provider.SetTraitList(data)
}

func (p *traitsProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *traitsProvider) DragKey() string {
	return traitDragKey
}

func (p *traitsProvider) DragSVG() *unison.SVG {
	return svg.GCSTraits
}

func (p *traitsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Trait]]) bool {
	return from == to
}

func (p *traitsProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.Trait]]) {
}

func (p *traitsProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: traitModifierDragKey,
		Drop: func(rowIndex int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*gurps.TraitModifier]]); ok {
				dataOwner := p.DataOwner()
				rows := make([]*gurps.TraitModifier, 0, len(tableDragData.Rows))
				libraryFile := libraryFileFromTable(tableDragData.Table)
				for _, row := range tableDragData.Rows {
					rows = append(rows, row.Data().Clone(libraryFile, dataOwner, nil, false))
				}
				rowData := p.table.RowFromIndex(rowIndex).Data()
				rowData.Modifiers = append(rowData.Modifiers, rows...)
				p.table.SyncToModel()
				if !toolbox.IsNil(dataOwner) {
					if entity := dataOwner.OwningEntity(); entity != nil {
						if rebuilder := unison.Ancestor[Rebuildable](p.table); rebuilder != nil {
							rebuilder.Rebuild(true)
						}
						ProcessModifiers(p.table, rows)
						ProcessNameables(p.table, rows)
					}
				}
			}
		},
	}
}

func (p *traitsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait"), i18n.Text("Traits")
}

func (p *traitsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Trait]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Trait]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.Trait](gurps.TraitsHeaderData(id), p.forPage))
	}
	return headers
}

func (p *traitsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Trait]]) {
}

func (p *traitsProvider) ColumnIDs() []int {
	columnIDs := append(make([]int, 0, 4),
		gurps.TraitDescriptionColumn,
		gurps.TraitPointsColumn,
	)
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.TraitTagsColumn)
	}
	columnIDs = append(columnIDs, gurps.TraitReferenceColumn)
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity == nil || !entity.SheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.TraitLibSrcColumn)
		}
	}
	return columnIDs
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
	InsertItems(owner, table, p.provider.TraitList, p.provider.SetTraitList,
		func(_ *unison.Table[*Node[*gurps.Trait]]) []*Node[*gurps.Trait] { return p.RootRows() }, item)
	EditTrait(owner, item)
}

func (p *traitsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.TraitList())
}

func (p *traitsProvider) Deserialize(data []byte) error {
	var rows []*gurps.Trait
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetTraitList(rows)
	return nil
}

func (p *traitsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list,
		ContextMenuItem{i18n.Text("New Trait"), NewTraitItemID},
		ContextMenuItem{i18n.Text("New Trait Container"), NewTraitContainerItemID},
		ContextMenuItem{i18n.Text("Add Natural Attacks"), AddNaturalAttacksItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
