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
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

const equipmentModifierDragKey = "equipment_modifier"

var _ TableProvider[*model.EquipmentModifier] = &eqpModProvider{}

type eqpModProvider struct {
	table     *unison.Table[*Node[*model.EquipmentModifier]]
	provider  model.EquipmentModifierListProvider
	forEditor bool
}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider model.EquipmentModifierListProvider, forEditor bool) TableProvider[*model.EquipmentModifier] {
	return &eqpModProvider{
		provider:  provider,
		forEditor: forEditor,
	}
}

func (p *eqpModProvider) RefKey() string {
	return equipmentModifierDragKey
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
		rows = append(rows, NewNode[*model.EquipmentModifier](p.table, nil, one, false))
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
	return equipmentModifierDragKey
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
	headers := make([]unison.TableColumnHeader[*Node[*model.EquipmentModifier]], 0, 7)
	if p.forEditor {
		headers = append(headers, NewEnabledHeader[*model.EquipmentModifier](false))
	}
	return append(headers,
		NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Equipment Modifier"), "", false),
		NewEditorListHeader[*model.EquipmentModifier](i18n.Text("TL"), i18n.Text("Tech Level"), false),
		NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Cost Adjustment"), "", false),
		NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Weight Adjustment"), "", false),
		NewEditorListHeader[*model.EquipmentModifier](i18n.Text("Tags"), "", false),
		NewEditorPageRefHeader[*model.EquipmentModifier](false),
	)
}

func (p *eqpModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*model.EquipmentModifier]]) {
}

func (p *eqpModProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 7)
	if p.forEditor {
		columnIDs = append(columnIDs, model.EquipmentModifierEnabledColumn)
	}
	return append(columnIDs,
		model.EquipmentModifierDescriptionColumn,
		model.EquipmentModifierTechLevelColumn,
		model.EquipmentModifierCostColumn,
		model.EquipmentModifierWeightColumn,
		model.EquipmentModifierTagsColumn,
		model.EquipmentModifierReferenceColumn,
	)
}

func (p *eqpModProvider) HierarchyColumnID() int {
	return model.EquipmentModifierDescriptionColumn
}

func (p *eqpModProvider) ExcessWidthColumnID() int {
	return model.EquipmentModifierDescriptionColumn
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
	list = append(list,
		ContextMenuItem{i18n.Text("New Equipment Modifier"), NewEquipmentModifierItemID},
		ContextMenuItem{i18n.Text("New Equipment Modifier Container"), NewEquipmentContainerModifierItemID},
	)
	return AppendDefaultContextMenuItems(list)
}
