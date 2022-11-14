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
	"fmt"

	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/jio"
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/log/jot"
	"github.com/richardwilkes/toolbox/txt"
	"github.com/richardwilkes/unison"
	"golang.org/x/exp/maps"
)

const equipmentDragKey = "equipment"

var (
	equipmentListColMap = map[int]int{
		0: model.EquipmentDescriptionColumn,
		1: model.EquipmentMaxUsesColumn,
		2: model.EquipmentTLColumn,
		3: model.EquipmentLCColumn,
		4: model.EquipmentCostColumn,
		5: model.EquipmentWeightColumn,
		6: model.EquipmentTagsColumn,
		7: model.EquipmentReferenceColumn,
	}
	carriedEquipmentPageColMap = map[int]int{
		0:  model.EquipmentEquippedColumn,
		1:  model.EquipmentQuantityColumn,
		2:  model.EquipmentDescriptionColumn,
		3:  model.EquipmentUsesColumn,
		4:  model.EquipmentTLColumn,
		5:  model.EquipmentLCColumn,
		6:  model.EquipmentCostColumn,
		7:  model.EquipmentWeightColumn,
		8:  model.EquipmentExtendedCostColumn,
		9:  model.EquipmentExtendedWeightColumn,
		10: model.EquipmentReferenceColumn,
	}
	otherEquipmentPageColMap = map[int]int{
		0: model.EquipmentQuantityColumn,
		1: model.EquipmentDescriptionColumn,
		2: model.EquipmentUsesColumn,
		3: model.EquipmentTLColumn,
		4: model.EquipmentLCColumn,
		5: model.EquipmentCostColumn,
		6: model.EquipmentWeightColumn,
		7: model.EquipmentExtendedCostColumn,
		8: model.EquipmentExtendedWeightColumn,
		9: model.EquipmentReferenceColumn,
	}
	_ TableProvider[*model.Equipment] = &equipmentProvider{}
)

type equipmentProvider struct {
	table    *unison.Table[*Node[*model.Equipment]]
	colMap   map[int]int
	provider model.EquipmentListProvider
	forPage  bool
	carried  bool
}

// NewEquipmentProvider creates a new table provider for equipment. 'carried' is only relevant if 'forPage' is true.
func NewEquipmentProvider(provider model.EquipmentListProvider, forPage, carried bool) TableProvider[*model.Equipment] {
	p := &equipmentProvider{
		provider: provider,
		forPage:  forPage,
		carried:  carried,
	}
	if forPage {
		if carried {
			p.colMap = carriedEquipmentPageColMap
		} else {
			p.colMap = otherEquipmentPageColMap
		}
	} else {
		p.colMap = equipmentListColMap
	}
	return p
}

func (p *equipmentProvider) RefKey() string {
	if p.carried {
		return model.BlockLayoutEquipmentKey
	}
	return model.BlockLayoutOtherEquipmentKey
}

func (p *equipmentProvider) AllTags() []string {
	set := make(map[string]struct{})
	model.Traverse(func(modifier *model.Equipment) bool {
		for _, tag := range modifier.Tags {
			set[tag] = struct{}{}
		}
		return false
	}, false, false, p.RootData()...)
	tags := maps.Keys(set)
	txt.SortStringsNaturalAscending(tags)
	return tags
}

func (p *equipmentProvider) SetTable(table *unison.Table[*Node[*model.Equipment]]) {
	p.table = table
}

func (p *equipmentProvider) RootRowCount() int {
	return len(p.equipmentList())
}

func (p *equipmentProvider) RootRows() []*Node[*model.Equipment] {
	data := p.equipmentList()
	rows := make([]*Node[*model.Equipment], 0, len(data))
	for _, one := range data {
		rows = append(rows, NewNode[*model.Equipment](p.table, nil, p.colMap, one, p.forPage))
	}
	return rows
}

func (p *equipmentProvider) SetRootRows(rows []*Node[*model.Equipment]) {
	p.setEquipmentList(ExtractNodeDataFromList(rows))
}

func (p *equipmentProvider) RootData() []*model.Equipment {
	return p.equipmentList()
}

func (p *equipmentProvider) SetRootData(data []*model.Equipment) {
	p.setEquipmentList(data)
}

func (p *equipmentProvider) Entity() *model.Entity {
	return p.provider.Entity()
}

func (p *equipmentProvider) DragKey() string {
	return equipmentDragKey
}

func (p *equipmentProvider) DragSVG() *unison.SVG {
	return svg.GCSEquipment
}

func (p *equipmentProvider) DropShouldMoveData(from, to *unison.Table[*Node[*model.Equipment]]) bool {
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

func (p *equipmentProvider) ProcessDropData(from, to *unison.Table[*Node[*model.Equipment]]) {
	if p.carried && from != to {
		for _, row := range to.SelectedRows(true) {
			if equipmentRow, ok := any(row).(*Node[*model.Equipment]); ok {
				model.Traverse(func(e *model.Equipment) bool {
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
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*model.EquipmentModifier]]); ok {
				entity := p.Entity()
				rows := make([]*model.EquipmentModifier, 0, len(tableDragData.Rows))
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

func (p *equipmentProvider) Headers() []unison.TableColumnHeader[*Node[*model.Equipment]] {
	var headers []unison.TableColumnHeader[*Node[*model.Equipment]]
	for i := 0; i < len(p.colMap); i++ {
		switch p.colMap[i] {
		case model.EquipmentEquippedColumn:
			headers = append(headers, NewEditorEquippedHeader[*model.Equipment](p.forPage))
		case model.EquipmentQuantityColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("#"), i18n.Text("Quantity"), p.forPage))
		case model.EquipmentDescriptionColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](p.descriptionText(), "", p.forPage))
		case model.EquipmentUsesColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("Uses"), i18n.Text("The number of uses remaining"), p.forPage))
		case model.EquipmentMaxUsesColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("Uses"), i18n.Text("The maximum number of uses"), p.forPage))
		case model.EquipmentTLColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("TL"), i18n.Text("Tech Level"), p.forPage))
		case model.EquipmentLCColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("LC"), i18n.Text("Legality Class"), p.forPage))
		case model.EquipmentCostColumn:
			headers = append(headers, NewMoneyHeader[*model.Equipment](p.forPage))
		case model.EquipmentExtendedCostColumn:
			headers = append(headers, NewExtendedMoneyHeader[*model.Equipment](p.forPage))
		case model.EquipmentWeightColumn:
			headers = append(headers, NewWeightHeader[*model.Equipment](p.forPage))
		case model.EquipmentExtendedWeightColumn:
			headers = append(headers, NewEditorExtendedWeightHeader[*model.Equipment](p.forPage))
		case model.EquipmentTagsColumn:
			headers = append(headers, NewEditorListHeader[*model.Equipment](i18n.Text("Tags"), "", p.forPage))
		case model.EquipmentReferenceColumn:
			headers = append(headers, NewEditorPageRefHeader[*model.Equipment](p.forPage))
		default:
			jot.Fatalf(1, "invalid equipment column: %d", p.colMap[i])
		}
	}
	return headers
}

func (p *equipmentProvider) SyncHeader(headers []unison.TableColumnHeader[*Node[*model.Equipment]]) {
	if p.forPage {
		for i := 0; i < len(p.colMap); i++ {
			if p.colMap[i] == model.EquipmentDescriptionColumn {
				if header, ok2 := headers[i].(*PageTableColumnHeader[*model.Equipment]); ok2 {
					header.Label.Text = p.descriptionText()
				}
				break
			}
		}
	}
}

func (p *equipmentProvider) HierarchyColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.EquipmentDescriptionColumn {
			return k
		}
	}
	return -1
}

func (p *equipmentProvider) ExcessWidthColumnIndex() int {
	for k, v := range p.colMap {
		if v == model.EquipmentDescriptionColumn {
			return k
		}
	}
	return 0
}

func (p *equipmentProvider) descriptionText() string {
	title := i18n.Text("Equipment")
	if p.forPage {
		if entity, ok := p.provider.(*model.Entity); ok {
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

func (p *equipmentProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*model.Equipment]]) {
	OpenEditor[*model.Equipment](table, func(item *model.Equipment) { EditEquipment(owner, item, p.carried) })
}

func (p *equipmentProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*model.Equipment]], variant ItemVariant) {
	topListFunc := p.provider.OtherEquipmentList
	setTopListFunc := p.provider.SetOtherEquipmentList
	if p.carried {
		topListFunc = p.provider.CarriedEquipmentList
		setTopListFunc = p.provider.SetCarriedEquipmentList
	}
	item := model.NewEquipment(p.Entity(), nil, variant == ContainerItemVariant)
	InsertItems[*model.Equipment](owner, table, topListFunc, setTopListFunc,
		func(_ *unison.Table[*Node[*model.Equipment]]) []*Node[*model.Equipment] {
			return p.RootRows()
		}, item)
	EditEquipment(owner, item, p.carried)
}

func (p *equipmentProvider) equipmentList() []*model.Equipment {
	if p.carried {
		return p.provider.CarriedEquipmentList()
	}
	return p.provider.OtherEquipmentList()
}

func (p *equipmentProvider) setEquipmentList(list []*model.Equipment) {
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
	var rows []*model.Equipment
	if err := jio.DecompressAndDeserialize(data, &rows); err != nil {
		return err
	}
	p.setEquipmentList(rows)
	return nil
}

func (p *equipmentProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	if p.carried {
		list = append(list, CarriedEquipmentExtraContextMenuItems...)
	} else {
		list = append(list, OtherEquipmentExtraContextMenuItems...)
	}
	return append(list, DefaultContextMenuItems...)
}
