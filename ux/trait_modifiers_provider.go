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

const traitModifierDragKey = "trait_modifier"

var (
	traitModifierColMap = map[int]int{
		0: model.TraitModifierDescriptionColumn,
		1: model.TraitModifierCostColumn,
		2: model.TraitModifierTagsColumn,
		3: model.TraitModifierReferenceColumn,
	}
	traitModifierInEditorColMap = map[int]int{
		0: model.TraitModifierEnabledColumn,
		1: model.TraitModifierDescriptionColumn,
		2: model.TraitModifierCostColumn,
		3: model.TraitModifierTagsColumn,
		4: model.TraitModifierReferenceColumn,
	}
	_ TableProvider[*model.TraitModifier] = &traitModifiersProvider{}
)

type traitModifiersProvider struct {
	table    *unison.Table[*Node[*model.TraitModifier]]
	colMap   map[int]int
	provider model.TraitModifierListProvider
}

// NewTraitModifiersProvider creates a new table provider for trait modifiers.
func NewTraitModifiersProvider(provider model.TraitModifierListProvider, forEditor bool) TableProvider[*model.TraitModifier] {
	p := &traitModifiersProvider{
		provider: provider,
	}
	if forEditor {
		p.colMap = traitModifierInEditorColMap
	} else {
		p.colMap = traitModifierColMap
	}
	return p
}

func (p *traitModifiersProvider) RefKey() string {
	return traitModifierDragKey
}

func (p *traitModifiersProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(modifier *model.TraitModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *traitModifiersProvider) SetTable(table *unison.Table[*Node[*model.TraitModifier]]) {
	p.table = table
}

func (p *traitModifiersProvider) RootRowCount() int {
	return len(p.provider.TraitModifierList())
}

func (p *traitModifiersProvider) RootRows() []*Node[*model.TraitModifier] {
	data := p.provider.TraitModifierList()
	rows := make([]*Node[*model.TraitModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.TraitModifier](p.table, nil, p.colMap, one, false))
	}
	return rows
}

func (p *traitModifiersProvider) SetRootRows(rows []*Node[*model.TraitModifier]) {
	p.provider.SetTraitModifierList(ExtractNodeDataFromList(rows))
}

func (p *traitModifiersProvider) RootData() []*model.TraitModifier {
	return p.provider.TraitModifierList()
}

func (p *traitModifiersProvider) SetRootData(data []*model.TraitModifier) {
	p.provider.SetTraitModifierList(data)
}

func (p *traitModifiersProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *traitModifiersProvider) DragKey() string {
	return traitModifierDragKey
}

func (p *traitModifiersProvider) DragSVG() *unison.SVG {
	return svg.GCSTraitModifiers
}

func (p *traitModifiersProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.TraitModifier]]) bool {
	return from == to
}

func (p *traitModifiersProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.TraitModifier]]) {
}

func (p *traitModifiersProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *traitModifiersProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Trait Modifier"), i18n.Text("Trait Modifiers")
}

func (p *traitModifiersProvider) Headers() []unison.TableColumnHeader[*Node[*model.TraitModifier]] {
	var headers []unison.TableColumnHeader[*Node[*model.TraitModifier]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.TraitModifierEnabledColumn:
			headers = append(headers, NewEnabledHeader[*model.TraitModifier](false))
		case model.TraitModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.TraitModifier](i18n.Text("Trait Modifier"), "", false))
		case model.TraitModifierCostColumn:
			headers = append(headers, NewEditorListHeader[*model.TraitModifier](i18n.Text("Cost Modifier"), "", false))
		case model.TraitModifierTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.TraitModifier](i18n.Text("Tags"), "", false))
		case model.TraitModifierReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.TraitModifier](false))
		default:
			jot.Fatalf(1, "invalid trait modifier column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *traitModifiersProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.TraitModifier]]) {
}

func (p *traitModifiersProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.TraitModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *traitModifiersProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *traitModifiersProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.TraitModifier]]) {
	OpenEditor[*model.TraitModifier](table, func(item *model.TraitModifier) {
		EditTraitModifier(owner, item)
	})
}

func (p *traitModifiersProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.TraitModifier]], variant ItemVariant) {
	item := model.NewTraitModifier(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*model.TraitModifier](owner, table, p.provider.TraitModifierList, p.provider.SetTraitModifierList,
		func(_ *unison.Table[*Node[*model.TraitModifier]]) []*Node[*model.TraitModifier] {
			return p.RootRows()
		}, item)
	EditTraitModifier(owner, item)
}

func (p *traitModifiersProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.TraitModifierList())
}

func (p *traitModifiersProvider) Deserialize(data []byte) error {
	var rows []*model.TraitModifier
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetTraitModifierList(rows)
	return nil
}

func (p *traitModifiersProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list, TraitModifierExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
