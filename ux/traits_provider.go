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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/gurps/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var (
	traitListColMap = map[int]int{
		0: gurps.TraitDescriptionColumn,
		1: gurps.TraitPointsColumn,
		2: gurps.TraitTagsColumn,
		3: gurps.TraitReferenceColumn,
	}
	traitPageColMap = map[int]int{
		0: gurps.TraitDescriptionColumn,
		1: gurps.TraitPointsColumn,
		2: gurps.TraitReferenceColumn,
	}
	_ TableProvider[*gurps.Trait] = &traitsProvider{}
)

type traitsProvider struct {
	table    *unison.Table[*Node[*gurps.Trait]]
	colMap   map[int]int
	provider gurps.TraitListProvider
	forPage  bool
}

// NewTraitsProvider creates a new table provider for traits.
func NewTraitsProvider(provider gurps.TraitListProvider, forPage bool) TableProvider[*gurps.Trait] {
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
	tags := maps.Keys(set)
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
		rows = append(rows, NewNode[*gurps.Trait](p.table, nil, p.colMap, one, p.forPage))
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

func (p *traitsProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *traitsProvider) DragKey() string {
	return gid.Trait
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
		DragKey: gid.TraitModifier,
		Drop: func(rowIndex int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*gurps.TraitModifier]]); ok {
				entity := p.Entity()
				rows := make([]*gurps.TraitModifier, 0, len(tableDragData.Rows))
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

func (p *traitsProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Trait]] {
	var headers []unison.TableColumnHeader[*Node[*gurps.Trait]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.TraitDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Trait](i18n.Text("Trait"), "", p.forPage))
		case gurps.TraitPointsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Trait](i18n.Text("Pts"), i18n.Text("Points"), p.forPage))
		case gurps.TraitTagsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Trait](i18n.Text("Tags"), "", p.forPage))
		case gurps.TraitReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*gurps.Trait](p.forPage))
		default:
			jot.Fatalf(1, "invalid trait column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *traitsProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.Trait]]) {
}

func (p *traitsProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.TraitDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *traitsProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *traitsProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	OpenEditor[*gurps.Trait](table, func(item *gurps.Trait) { EditTrait(owner, item) })
}

func (p *traitsProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]], variant ItemVariant) {
	item := gurps.NewTrait(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*gurps.Trait](owner, table, p.provider.TraitList, p.provider.SetTraitList,
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
	list = append(list, TraitExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
