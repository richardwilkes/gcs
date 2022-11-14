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
	"github.com/richardwilkes/gcs/v5/model/gid"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

var (
	equipmentModifierColMap = map[int]int{
		0: model.EquipmentModifierDescriptionColumn,
		1: model.EquipmentModifierTechLevelColumn,
		2: model.EquipmentModifierCostColumn,
		3: model.EquipmentModifierWeightColumn,
		4: model.EquipmentModifierTagsColumn,
		5: model.EquipmentModifierReferenceColumn,
	}
	equipmentModifierInEditorColMap = map[int]int{
		0: model.EquipmentModifierEnabledColumn,
		1: model.EquipmentModifierDescriptionColumn,
		2: model.EquipmentModifierTechLevelColumn,
		3: model.EquipmentModifierCostColumn,
		4: model.EquipmentModifierWeightColumn,
		5: model.EquipmentModifierTagsColumn,
		6: model.EquipmentModifierReferenceColumn,
	}
	_ TableProvider[*model.EquipmentModifier] = &eqpModProvider{}
)

type eqpModProvider struct {
	table    *unison.Table[*Node[*model.EquipmentModifier]]
	colMap   map[int]int
	provider model.EquipmentModifierListProvider
}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider model.EquipmentModifierListProvider, forEditor bool) TableProvider[*model.EquipmentModifier] {
	p := &eqpModProvider{
		provider: provider,
	}
	if forEditor {
		p.colMap = equipmentModifierInEditorColMap
	} else {
		p.colMap = equipmentModifierColMap
	}
	return p
}

func (p *eqpModProvider) RefKey() string {
	return gid.EquipmentModifier
}

func (p *eqpModProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(modifier *model.EquipmentModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *eqpModProvider) SetTable(table *unison.Table[*Node[*model.EquipmentModifier]]) {
	p.table = table
}

func (p *eqpModProvider) RootRowCount() int {
	return len(p.provider.EquipmentModifierList())
}

func (p *eqpModProvider) RootRows() []*Node[*model.EquipmentModifier] {
	data := p.provider.EquipmentModifierList()
	rows := make([]*Node[*model.EquipmentModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.EquipmentModifier](p.table, nil, p.colMap, one, false))
	}
	return rows
}

func (p *eqpModProvider) SetRootRows(rows []*Node[*model.EquipmentModifier]) {
	p.provider.SetEquipmentModifierList(ExtractNodeDataFromList(rows))
}

func (p *eqpModProvider) RootData() []*model.EquipmentModifier {
	return p.provider.EquipmentModifierList()
}

func (p *eqpModProvider) SetRootData(data []*model.EquipmentModifier) {
	p.provider.SetEquipmentModifierList(data)
}

func (p *eqpModProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *eqpModProvider) DragKey() string {
	return gid.EquipmentModifier
}

func (p *eqpModProvider) DragSVG() *unison.SVG {
	return svg.GCSEquipmentModifiers
}

func (p *eqpModProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.EquipmentModifier]]) bool {
	return from == to
}

func (p *eqpModProvider) ProcessDropData(_, _ *unison.Table[*Node[*model.EquipmentModifier]]) {
}

func (p *eqpModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *eqpModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Equipment Modifier"), i18n.Text("Equipment Modifiers")
}

func (p *eqpModProvider) Headers() []unison.TableColumnHeader[*Node[*model.EquipmentModifier]] {
	var headers []unison.TableColumnHeader[*Node[*model.EquipmentModifier]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.EquipmentModifierEnabledColumn:
			headers = append(headers, NewEnabledHeader[*model.EquipmentModifier](false))
		case model.EquipmentModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Equipment Modifier"), "", false))
		case model.EquipmentModifierTechLevelColumn:
			headers = append(headers, NewEditorListHeader[*model.EquipmentModifier](i18n.Text("TL"), i18n.Text("Tech Level"), false))
		case model.EquipmentModifierCostColumn:
			headers = append(headers, NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Cost Adjustment"), "", false))
		case model.EquipmentModifierWeightColumn:
			headers = append(headers, NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Weight Adjustment"), "", false))
		case model.EquipmentModifierTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Tags"), "", false))
		case model.EquipmentModifierReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.EquipmentModifier](false))
		default:
			jot.Fatalf(1, "invalid equipment modifier column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *eqpModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.EquipmentModifier]]) {
}

func (p *eqpModProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.EquipmentModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *eqpModProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *eqpModProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.EquipmentModifier]]) {
	OpenEditor[*model.EquipmentModifier](table, func(item *model.EquipmentModifier) {
		EditEquipmentModifier(owner, item)
	})
}

func (p *eqpModProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.EquipmentModifier]], variant ItemVariant) {
	item := model.NewEquipmentModifier(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*model.EquipmentModifier](owner, table, p.provider.EquipmentModifierList,
		p.provider.SetEquipmentModifierList,
		func(_ *unison.Table[*Node[*model.EquipmentModifier]]) []*Node[*model.EquipmentModifier] {
			return p.RootRows()
		}, item)
	EditEquipmentModifier(owner, item)
}

func (p *eqpModProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.EquipmentModifierList())
}

func (p *eqpModProvider) Deserialize(data []byte) error {
	var rows []*model.EquipmentModifier
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.provider.SetEquipmentModifierList(rows)
	return nil
}

func (p *eqpModProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	list = append(list, EquipmentModifierExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
