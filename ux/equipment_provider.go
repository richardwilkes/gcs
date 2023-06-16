/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

const equipmentDragKey = "equipment"

var _ TableProvider[*gurps.Equipment] = &equipmentProvider{}

type equipmentProvider struct {
	table    *unison.Table[*Node[*gurps.Equipment]]
	provider gurps.EquipmentListProvider
	forPage  bool
	carried  bool
}

// NewEquipmentProvider creates a new table provider for equipment. 'carried' is only relevant if 'forPage' is true.
func NewEquipmentProvider(provider gurps.EquipmentListProvider, forPage, carried bool) TableProvider[*gurps.Equipment] {
	return &equipmentProvider{
		provider: provider,
		forPage:  forPage,
		carried:  carried,
	}
}

func (p *equipmentProvider) RefKey() string {
	if p.carried {
		return gurps.BlockLayoutEquipmentKey
	}
	return gurps.BlockLayoutOtherEquipmentKey
}

func (p *equipmentProvider) AllTags() []string {
	set := make(map[string]struct{})
	gurps.Traverse(func(modifier *gurps.Equipment) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *equipmentProvider) SetTable(table *unison.Table[*Node[*gurps.Equipment]]) {
	p.table = table
}

func (p *equipmentProvider) RootRowCount() int {
	return len(p.equipmentList())
}

func (p *equipmentProvider) RootRows() []*Node[*gurps.Equipment] {
	data := p.equipmentList()
	rows := make([]*Node[*gurps.Equipment], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*gurps.Equipment](p.table, nil, one, p.forPage))
	}
	return rows
}

func (p *equipmentProvider) SetRootRows(rows []*Node[*gurps.Equipment]) {
	p.setEquipmentList(ExtractNodeDataFromList(rows))
}

func (p *equipmentProvider) RootData() []*gurps.Equipment {
	return p.equipmentList()
}

func (p *equipmentProvider) SetRootData(data []*gurps.Equipment) {
	p.setEquipmentList(data)
}

func (p *equipmentProvider) Entity() *gurps.Entity {
	return p.provider.Entity()
}

func (p *equipmentProvider) DragKey() string {
	return equipmentDragKey
}

func (p *equipmentProvider) DragSVG() *unison.SVG {
	return svg.GCSEquipment
}

func (p *equipmentProvider) DropShouldMoveData(from, to *unison.Table[*Node[*gurps.Equipment]]) bool {
	// Within same table?
	if from == to {
		return true
	}
	// Within same dockable?
	dockable := unison.Ancestor[unison.Dockable](from)
	if dockable != nil && dockable == unison.Ancestor[unison.Dockable](to) {
		return true
	}
	return false
}

func (p *equipmentProvider) ProcessDropData(from, to *unison.Table[*Node[*gurps.Equipment]]) {
	if p.carried && from != to {
		for _, row := range to.SelectedRows(true) {
			if equipmentRow, ok := any(row).(*Node[*gurps.Equipment]); ok {
				gurps.Traverse(func(e *gurps.Equipment) bool {
					e.Equipped = true
					return false
				}, false, false, equipmentRow.Data())
			}
		}
	}
}

func (p *equipmentProvider) AltDropSupport() *AltDropSupport {
	return &AltDropSupport{
		DragKey: equipmentModifierDragKey,
		Drop: func(rowIndex int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*gurps.EquipmentModifier]]); ok {
				entity := p.Entity()
				rows := make([]*gurps.EquipmentModifier, 0, len(tableDragData.Rows))
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

func (p *equipmentProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Equipment Item"), i18n.Text("Equipment Items")
}

func (p *equipmentProvider) Headers() []unison.TableColumnHeader[*Node[*gurps.Equipment]] {
	ids := p.ColumnIDs()
	headers := make([]unison.TableColumnHeader[*Node[*gurps.Equipment]], 0, len(ids))
	for _, id := range ids {
		switch id {
		case gurps.EquipmentEquippedColumn:
			headers = append(headers, NewEditorEquippedHeader[*gurps.Equipment](p.forPage))
		case gurps.EquipmentQuantityColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](i18n.Text("#"), i18n.Text("Quantity"), p.forPage))
		case gurps.EquipmentDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](p.descriptionText(), "", p.forPage))
		case gurps.EquipmentUsesColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](i18n.Text("Uses"), i18n.Text("The number of uses remaining"), p.forPage))
		case gurps.EquipmentTLColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](i18n.Text("TL"), i18n.Text("Tech Level"), p.forPage))
		case gurps.EquipmentLCColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](i18n.Text("LC"), i18n.Text("Legality Class"), p.forPage))
		case gurps.EquipmentCostColumn:
			headers = append(headers, NewMoneyHeader[*gurps.Equipment](p.forPage))
		case gurps.EquipmentExtendedCostColumn:
			headers = append(headers, NewExtendedMoneyHeader[*gurps.Equipment](p.forPage))
		case gurps.EquipmentWeightColumn:
			headers = append(headers, NewWeightHeader[*gurps.Equipment](p.forPage))
		case gurps.EquipmentExtendedWeightColumn:
			headers = append(headers, NewEditorExtendedWeightHeader[*gurps.Equipment](p.forPage))
		case gurps.EquipmentTagsColumn:
			headers = append(headers, NewEditorListHeader[*gurps.Equipment](i18n.Text("Tags"), "", p.forPage))
		case gurps.EquipmentReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*gurps.Equipment](p.forPage))
		}
	}
	return headers
}

func (p *equipmentProvider) SyncHeader(headers []unison.TableColumnHeader[*Node[*gurps.Equipment]]) {
	if p.forPage {
		if i := p.table.ColumnIndexForID(gurps.EquipmentDescriptionColumn); i != -1 {
			if header, ok := headers[i].(*PageTableColumnHeader[*gurps.Equipment]); ok {
				header.Label.Text = p.descriptionText()
			}
		}
	}
}

func (p *equipmentProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 11)
	if p.forPage && p.carried {
		columnIDs = append(columnIDs, gurps.EquipmentEquippedColumn)
	}
	columnIDs = append(columnIDs,
		gurps.EquipmentQuantityColumn,
		gurps.EquipmentDescriptionColumn,
		gurps.EquipmentUsesColumn,
		gurps.EquipmentTLColumn,
		gurps.EquipmentLCColumn,
		gurps.EquipmentCostColumn,
		gurps.EquipmentWeightColumn,
		gurps.EquipmentExtendedCostColumn,
		gurps.EquipmentExtendedWeightColumn,
	)
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.EquipmentTagsColumn)
	}
	return append(columnIDs, gurps.EquipmentReferenceColumn)
}

func (p *equipmentProvider) HierarchyColumnID() int {
	return gurps.EquipmentDescriptionColumn
}

func (p *equipmentProvider) ExcessWidthColumnID() int {
	return gurps.EquipmentDescriptionColumn
}

func (p *equipmentProvider) descriptionText() string {
	title := i18n.Text("Equipment")
	if p.forPage {
		if entity, ok := p.provider.(*gurps.Entity); ok {
			if p.carried {
				title = fmt.Sprintf(i18n.Text("Carried Equipment (%s; $%s)"),
					entity.SheetSettings.DefaultWeightUnits.Format(entity.WeightCarried(false)),
					entity.WealthCarried().String())
			} else {
				title = fmt.Sprintf(i18n.Text("Other Equipment ($%s)"), entity.WealthNotCarried().String())
			}
		}
	}
	return title
}

func (p *equipmentProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	OpenEditor[*gurps.Equipment](table, func(item *gurps.Equipment) { EditEquipment(owner, item, p.carried) })
}

func (p *equipmentProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], variant ItemVariant) {
	topListFunc := p.provider.OtherEquipmentList
	setTopListFunc := p.provider.SetOtherEquipmentList
	if p.carried {
		topListFunc = p.provider.CarriedEquipmentList
		setTopListFunc = p.provider.SetCarriedEquipmentList
	}
	item := gurps.NewEquipment(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*gurps.Equipment](owner, table, topListFunc, setTopListFunc,
		func(_ *unison.Table[*Node[*gurps.Equipment]]) []*Node[*gurps.Equipment] {
			return p.RootRows()
		}, item)
	EditEquipment(owner, item, p.carried)
}

func (p *equipmentProvider) equipmentList() []*gurps.Equipment {
	if p.carried {
		return p.provider.CarriedEquipmentList()
	}
	return p.provider.OtherEquipmentList()
}

func (p *equipmentProvider) setEquipmentList(list []*gurps.Equipment) {
	if p.carried {
		p.provider.SetCarriedEquipmentList(list)
	} else {
		p.provider.SetOtherEquipmentList(list)
	}
}

func (p *equipmentProvider) Serialize() ([]byte, error) {
	return jio.SerializeAndCompress(p.equipmentList())
}

func (p *equipmentProvider) Deserialize(data []byte) error {
	var rows []*gurps.Equipment
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.setEquipmentList(rows)
	return nil
}

func (p *equipmentProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	if p.carried {
		list = append(list,
			ContextMenuItem{i18n.Text("New Carried Equipment"), NewCarriedEquipmentItemID},
			ContextMenuItem{i18n.Text("New Carried Equipment Container"), NewCarriedEquipmentContainerItemID},
		)
	} else {
		list = append(list,
			ContextMenuItem{i18n.Text("New Other Equipment"), NewOtherEquipmentItemID},
			ContextMenuItem{i18n.Text("New Other Equipment Container"), NewOtherEquipmentContainerItemID},
		)
	}
	return AppendDefaultContextMenuItems(list)
}
