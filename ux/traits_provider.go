/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package ux

import (
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

const traitDragKey = "trait"

var (
	traitListColMap = map[int]int{
		0: model.TraitDescriptionColumn,
		1: model.TraitPointsColumn,
		2: model.TraitTagsColumn,
		3: model.TraitReferenceColumn,
	}
	traitPageColMap = map[int]int{
		0: model.TraitDescriptionColumn,
		1: model.TraitPointsColumn,
		2: model.TraitReferenceColumn,
	}
	_ TableProvider[*model.Trait] = &traitsProvider{}
)

type traitsProvider struct {
	table    *unison.Table[*Node[*model.Trait]]
	colMap   map[int]int
	provider model.TraitListProvider
	forPage  bool
}

// NewTraitsProvider creates a new table provider for traits.
func NewTraitsProvider(provider model.TraitListProvider, forPage bool) TableProvider[*model.Trait] {
	p := &traitsProvider{
		provider: provider,
		forPage:  forPage,
	}
	if forPage {
		p.colMap = traitPageColMap
	} else {
		p.colMap = traitListColMap
	}
	return p
}

func (p *traitsProvider) RefKey() string {
	return model.BlockLayoutTraitsKey
}

func (p *traitsProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(trait *model.Trait) bool {
		for _, tag := range trait.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *traitsProvider) SetTable(table *unison.Table[*Node[*model.Trait]]) {
	p.table = table
}

func (p *traitsProvider) RootRowCount() int {
	return len(p.provider.TraitList())
}

func (p *traitsProvider) RootRows() []*Node[*model.Trait] {
	data := p.provider.TraitList()
	rows := make([]*Node[*model.Trait], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Trait](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *traitsProvider) SetRootRows(rows []*Node[*model.Trait]) {
	p.provider.SetTraitList(ExtractNodeDataFromList(rows))
}

func (p *traitsProvider) RootData() []*model.Trait {
	return p.provider.TraitList()
}

func (p *traitsProvider) SetRootData(data []*model.Trait) {
	p.provider.SetTraitList(data)
}

func (p *traitsProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *traitsProvider) DragKey() string {
	return traitDragKey
}

func (p *traitsProvider) DragSVG() *unison.SVG {
	return svg.GCSTraits
}

func (p *traitsProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Trait]]) bool {
	return from == to
}

func (p *traitsProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.Trait]]) {
}

func (p *traitsProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: traitModifierDragKey,
		Drop: func(rowIndex int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*model.TraitModifier]]); ok {
				entity := p.Entity()
				rows := make([]*model.TraitModifier, 0, len(tableDragData.Rows))
				for _, row := range tableDragData.Rows {
					rows = append(rows, row.Data().Clone(entity, nil, false))
				}
				rowData := p.table.RowFromIndex(rowIndex).Data()
				rowData.Modifiers = append(rowData.Modifiers, rows...)
				p.table.SyncToModel()
				if entity != nil {
					if rebuilder := unison.Ancestor[Rebuildable](p.table); rebuilder != nil {
						rebuilder.Rebuild(true)
					}
					ProcessModifiers(p.table, rows)
					ProcessNameables(p.table, rows)
				}
			}
		},
	}
}

func (p *traitsProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait"), i18n.Text("Traits")
}

func (p *traitsProvider) Headers() []unison.TableColumnHeader[*Node[*model.Trait]] {
	var headers []unison.TableColumnHeader[*Node[*model.Trait]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.TraitDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.Trait](i18n.Text("Trait"), "", p.forPage))
		case model.TraitPointsColumn:
			headers = append(headers, NewEditorListHeader[*model.Trait](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		case model.TraitTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.Trait](i18n.Text("Tags"), "", p.forPage))
		case model.TraitReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.Trait](p.forPage))
		default:
			jot.Fatalf(1, "invalid trait column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *traitsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.Trait]]) {
}

func (p *traitsProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.TraitDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *traitsProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *traitsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Trait]]) {
	OpenEditor[*model.Trait](table, func(item *model.Trait) { EditTrait(owner, item) })
}

func (p *traitsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Trait]], variant ItemVariant) {
	item := model.NewTrait(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*model.Trait](owner, table, p.provider.TraitList, p.provider.SetTraitList,
		func(_ *unison.Table[*Node[*model.Trait]]) []*Node[*model.Trait] { return p.RootRows() }, item)
	EditTrait(owner, item)
}

func (p *traitsProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.TraitList())
}

func (p *traitsProvider) Deserialize(data []byte) error {
	var rows []*model.Trait
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetTraitList(rows)
	return nil
}

func (p *traitsProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list, TraitExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
