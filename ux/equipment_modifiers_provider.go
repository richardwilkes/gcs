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
	equipmentModifierColMap = map[int]int{
		0: gurps.EquipmentModifierDescriptionColumn,
		1: gurps.EquipmentModifierTechLevelColumn,
		2: gurps.EquipmentModifierCostColumn,
		3: gurps.EquipmentModifierWeightColumn,
		4: gurps.EquipmentModifierTagsColumn,
		5: gurps.EquipmentModifierReferenceColumn,
	}
	equipmentModifierInEditorColMap = map[int]int{
		0: gurps.EquipmentModifierEnabledColumn,
		1: gurps.EquipmentModifierDescriptionColumn,
		2: gurps.EquipmentModifierTechLevelColumn,
		3: gurps.EquipmentModifierCostColumn,
		4: gurps.EquipmentModifierWeightColumn,
		5: gurps.EquipmentModifierTagsColumn,
		6: gurps.EquipmentModifierReferenceColumn,
	}
	_ TableProvider[*gurps.EquipmentModifier] = &eqpModProvider{}
)

type eqpModProvider struct {
	table    *unison.Table[*Node[*gurps.EquipmentModifier]]
	colMap   map[int]int
	provider gurps.EquipmentModifierListProvider
}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider gurps.EquipmentModifierListProvider, forEditor bool) TableProvider[*gurps.EquipmentModifier] {
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
	gurps.Traverse(func(modifier *gurps.EquipmentModifier) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
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
		rows = append(rows, NewNode[*gurps.EquipmentModifier](p.table, nil, p.colMap, one, false))
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

func (p *eqpModProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *eqpModProvider) DragKey() string {
	return gid.EquipmentModifier
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
	var headers []unison.TableColumnHeader[*Node[*gurps.EquipmentModifier]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case gurps.EquipmentModifierEnabledColumn:
			headers = append(headers, NewEnabledHeader[*gurps.EquipmentModifier](false))
		case gurps.EquipmentModifierDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.EquipmentModifier](i18n.Text("Equipment Modifier"), "", false))
		case gurps.EquipmentModifierTechLevelColumn:
			headers = append(headers, NewEditorListHeader[*gurps.EquipmentModifier](i18n.Text("TL"), i18n.Text("Tech Level"), false))
		case gurps.EquipmentModifierCostColumn:
			headers = append(headers, NewEditorListHeader[*gurps.EquipmentModifier](i18n.Text("Cost Adjustment"), "", false))
		case gurps.EquipmentModifierWeightColumn:
			headers = append(headers, NewEditorListHeader[*gurps.EquipmentModifier](i18n.Text("Weight Adjustment"), "", false))
		case gurps.EquipmentModifierTagsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.EquipmentModifier](i18n.Text("Tags"), "", false))
		case gurps.EquipmentModifierReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*gurps.EquipmentModifier](false))
		default:
			jot.Fatalf(1, "invalid equipment modifier column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *eqpModProvider) SyncHeader(_ []unison.TableColumnHeader[*Node[*gurps.EquipmentModifier]]) {
}

func (p *eqpModProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == gurps.EquipmentModifierDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *eqpModProvider) ExcessWidthColumnIndex() int {
	return p.HierarchyColumnIndex()
}

func (p *eqpModProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.EquipmentModifier]]) {
	OpenEditor[*gurps.EquipmentModifier](table, func(item *gurps.EquipmentModifier) {
		EditEquipmentModifier(owner, item)
	})
}

func (p *eqpModProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.EquipmentModifier]], variant ItemVariant) {
	item := gurps.NewEquipmentModifier(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*gurps.EquipmentModifier](owner, table, p.provider.EquipmentModifierList,
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
	list = append(list, EquipmentModifierExtraContextMenuItems...)
	return append(list, DefaultContextMenuItems...)
}
