// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
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
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/uti"
	"github.com/richardwilkes/unison"
)

var _ TableProvider[*gurps.Equipment] = &equipmentProvider{}

type equipmentProvider struct {
	listProvider[*gurps.Equipment]
	provider gurps.EquipmentListProvider
	carried  bool
}

// NewEquipmentProvider creates a new table provider for equipment. 'carried' is only relevant if 'forPage' is true.
func NewEquipmentProvider(provider gurps.EquipmentListProvider, carried, forPage bool) TableProvider[*gurps.Equipment] {
	p := &equipmentProvider{provider: provider, carried: carried}
	list, setList := provider.OtherEquipmentList, provider.SetOtherEquipmentList
	if carried {
		list, setList = provider.CarriedEquipmentList, provider.SetCarriedEquipmentList
	}
	p.listProvider = listProvider[*gurps.Equipment]{
		dataOwner: provider,
		list:      list,
		setList:   setList,
		columnIDs: p.ColumnIDs,
		headerData: func(columnID int) gurps.HeaderData {
			return gurps.EquipmentHeaderData(columnID, provider, carried, forPage)
		},
		forPage: forPage,
	}
	return p
}

func (p *equipmentProvider) RefKey() string {
	if p.carried {
		return gurps.BlockEquipmentKey
	}
	return gurps.BlockOtherEquipmentKey
}

func (p *equipmentProvider) DragKey() *uti.DataType {
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
	dockable := from.Ancestor[unison.Dockable]()
	if dockable != nil && dockable == to.Ancestor[unison.Dockable]() {
		return true
	}
	return false
}

func (p *equipmentProvider) ProcessDropData(from, to *unison.Table[*Node[*gurps.Equipment]]) {
	if p.carried && from != to {
		for _, row := range to.SelectedRows(true) {
			gurps.Traverse(func(e *gurps.Equipment) bool {
				e.Equipped = true
				return false
			}, false, false, row.Data())
		}
	}
}

func (p *equipmentProvider) AltDropSupport() *AltDropSupport {
	return modifierAltDropSupport(&p.listProvider, equipmentModifierDragKey,
		func(target *gurps.Equipment, clones []*gurps.EquipmentModifier) {
			target.Modifiers = append(target.Modifiers, clones...)
		})
}

func (p *equipmentProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Equipment Item"), i18n.Text("Equipment Items")
}

func (p *equipmentProvider) SyncHeader(headers []unison.TableColumnHeader[*Node[*gurps.Equipment]]) {
	if p.forPage {
		if i := p.table.ColumnIndexForID(gurps.EquipmentDescriptionColumn); i != -1 {
			if header, ok := headers[i].(*PageTableColumnHeader[*gurps.Equipment]); ok {
				// The totals in the title change as the equipment and the display formats do, and with them whether
				// there are exact totals to offer as the tooltip, so both are refreshed here.
				data := p.headerData(gurps.EquipmentDescriptionColumn)
				header.Text = unison.NewSmallCapsText(data.Title, &header.TextDecoration)
				header.SetTooltipText(data.Detail)
			}
		}
	}
}

func (p *equipmentProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 12)
	if p.forPage && p.carried {
		columnIDs = append(columnIDs, gurps.EquipmentEquippedColumn)
	}
	// Unlike the equipped column, the switch column applies to other equipment, too, since some features (e.g.
	// contained weight reductions) take effect regardless of whether the equipment is carried.
	if showSwitchColumn(p.forPage, p.provider, p.RootData()) {
		columnIDs = append(columnIDs, gurps.EquipmentSwitchColumn)
	}
	columnIDs = append(
		columnIDs,
		gurps.EquipmentQuantityColumn,
		gurps.EquipmentDescriptionColumn,
	)
	if settings := p.pageSheetSettings(); settings != nil {
		if !settings.HideTLColumn {
			columnIDs = append(columnIDs, gurps.EquipmentTLColumn)
		}
		if !settings.HideLCColumn {
			columnIDs = append(columnIDs, gurps.EquipmentLCColumn)
		}
	} else {
		columnIDs = append(
			columnIDs,
			gurps.EquipmentTLColumn,
			gurps.EquipmentLCColumn,
		)
	}
	columnIDs = append(
		columnIDs,
		gurps.EquipmentCostColumn,
		gurps.EquipmentWeightColumn,
		gurps.EquipmentExtendedCostColumn,
		gurps.EquipmentExtendedWeightColumn,
	)
	if !p.forPage {
		columnIDs = append(columnIDs, gurps.EquipmentTagsColumn)
	}
	return p.appendReferenceColumns(columnIDs, gurps.EquipmentReferenceColumn, gurps.EquipmentLibSrcColumn)
}

func (p *equipmentProvider) HierarchyColumnID() int {
	return gurps.EquipmentDescriptionColumn
}

func (p *equipmentProvider) ExcessWidthColumnID() int {
	return gurps.EquipmentDescriptionColumn
}

func (p *equipmentProvider) OpenEditor(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]]) {
	OpenEditor(table, func(item *gurps.Equipment) { EditEquipment(owner, item, p.carried) })
}

func (p *equipmentProvider) CreateItem(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], variant ItemVariant) {
	item := gurps.NewEquipment(p.DataOwner(), nil, variant == ContainerItemVariant)
	p.insertItems(owner, table, item)
	EditEquipment(owner, item, p.carried)
}

func (p *equipmentProvider) ContextMenuItems() []ContextMenuItem {
	var list []ContextMenuItem
	if p.carried {
		list = append(
			list,
			ContextMenuItem{i18n.Text("New Carried Equipment"), NewCarriedEquipmentItemID},
			ContextMenuItem{i18n.Text("New Carried Equipment Container"), NewCarriedEquipmentContainerItemID},
		)
	} else {
		list = append(
			list,
			ContextMenuItem{i18n.Text("New Other Equipment"), NewOtherEquipmentItemID},
			ContextMenuItem{i18n.Text("New Other Equipment Container"), NewOtherEquipmentContainerItemID},
		)
	}
	return AppendDefaultContextMenuItems(list)
}
