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
	"github.com/richardwilkes/toolbox/collection/dict"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
)

const equipmentModifierDragKey = "equipment_modifier"

var _ TableProvider[*gurps.EquipmentModifier] = &eqpModProvider{}

type eqpModProvider struct {
	table     *unison.Table[*Node[*gurps.EquipmentModifier]]
	provider  gurps.EquipmentModifierListProvider
	forEditor bool
}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider gurps.EquipmentModifierListProvider, forEditor bool) TableProvider[*gurps.EquipmentModifier] {
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
	gurps.Traverse(func(modifier *gurps.EquipmentModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := dict.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *eqpModProvider) SetTable(table *unison.Table[*Node[*gurps.EquipmentModifier]]) {
	p.table = table
}

func (p *eqpModProvider) RootRowCount() int {
	return len(p.provider.EquipmentModifierList())
}

func (p *eqpModProvider) RootRows() []*Node[*gurps.EquipmentModifier] {
	data := p.provider.EquipmentModifierList()
	rows := make([]*Node[*gurps.EquipmentModifier], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode(p.table, nil, one, false))
	}
	return rows
}

func (p *eqpModProvider) SetRootRows(rows []*Node[*gurps.EquipmentModifier]) {
	p.provider.SetEquipmentModifierList(ExtractNodeDataFromList(rows))
}

func (p *eqpModProvider) RootData() []*gurps.EquipmentModifier {
	return p.provider.EquipmentModifierList()
}

func (p *eqpModProvider) SetRootData(data []*gurps.EquipmentModifier) {
	p.provider.SetEquipmentModifierList(data)
}

func (p *eqpModProvider) DataOwner() gurps.DataOwner {
	return p.provider.DataOwner()
}

func (p *eqpModProvider) DragKey() string {
	return equipmentModifierDragKey
}

func (p *eqpModProvider) DragSVG() *unison.SVG {
	return svg.GCSEquipmentModifiers
}

func (p *eqpModProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.EquipmentModifier]]) bool {
	return from == to
}

func (p *eqpModProvider) ProcessDropData(_, _ *unison.Table[*Node[*gurps.EquipmentModifier]]) {
}

func (p *eqpModProvider) AltDropSupport() *AltDropSupport {
	return nil
}

func (p *eqpModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Equipment Modifier"), i18n.Text("Equipment Modifiers")
}

func (p *eqpModProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.EquipmentModifier]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.EquipmentModifier]], 0, len(ids))
	for _, id := range ids {
		headers = append(headers, headerFromData[*gurps.EquipmentModifier](gurps.EquipmentModifierHeaderData(id), false))
	}
	return headers
}

func (p *eqpModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.EquipmentModifier]]) {
}

func (p *eqpModProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 7)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.EquipmentModifierEnabledColumn)
	}
	columnIDs = append(columnIDs,
		gurps.EquipmentModifierDescriptionColumn,
		gurps.EquipmentModifierTechLevelColumn,
		gurps.EquipmentModifierCostColumn,
		gurps.EquipmentModifierWeightColumn,
		gurps.EquipmentModifierTagsColumn,
		gurps.EquipmentModifierReferenceColumn,
	)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.EquipmentModifierLibSrcColumn)
	}
	return columnIDs
}

func (p *eqpModProvider) HierarchyColumnID() int {
	return gurps.EquipmentModifierDescriptionColumn
}

func (p *eqpModProvider) ExcessWidthColumnID() int {
	return gurps.EquipmentModifierDescriptionColumn
}

func (p *eqpModProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.EquipmentModifier]]) {
	OpenEditor(table, func(item *gurps.EquipmentModifier) {
		EditEquipmentModifier(owner, item)
	})
}

func (p *eqpModProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.EquipmentModifier]], variant ItemVariant) {
	item := gurps.NewEquipmentModifier(p.DataOwner(), nil, variant == ContainerItemVariant)
	InsertItems(owner, table, p.provider.EquipmentModifierList,
		p.provider.SetEquipmentModifierList,
		func(_ *unison.Table[*Node[*gurps.EquipmentModifier]]) []*Node[*gurps.EquipmentModifier] {
			return p.RootRows()
		}, item)
	EditEquipmentModifier(owner, item)
}

func (p *eqpModProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.provider.EquipmentModifierList())
}

func (p *eqpModProvider) Deserialize(data []byte) error {
	var rows []*gurps.EquipmentModifier
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
