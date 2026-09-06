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
	"github.com/richardwilkes/toolbox/v2/xreflect"
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
	return &AltDropSupport{
		DragKey: equipmentModifierDragKey,
		Drop: func(rowIndexes []int, data any) {
			if tableDragData, ok := data.(*unison.TableDragData[*Node[*gurps.EquipmentModifier]]); ok {
				// Every target is resolved up front, since the rebuild below replaces this table with a new one --
				// leaving this very table an orphan whose rows are no longer the ones on screen -- so the row indexes
				// only mean something before it runs. The sync in between is harmless: attaching modifiers adds and
				// removes no rows and changes no disclosure, so it rebuilds the row cache with the same rows in the
				// same order.
				targets := make([]*gurps.Equipment, 0, len(rowIndexes))
				for _, rowIndex := range rowIndexes {
					if row := p.table.RowFromIndex(rowIndex); row != nil {
						targets = append(targets, row.Data())
					}
				}
				if len(targets) == 0 {
					return
				}
				dataOwner := p.DataOwner()
				libraryFile := libraryFileFromTable(tableDragData.Table)
				// Each target has to be given its own clones. They are separate modifiers from here on -- enabled,
				// renamed and edited independently -- so sharing one set among the targets would tie them together.
				// The clones are kept grouped by target for the nameables prompt below, which would otherwise show the
				// copies of one modifier as a run of identically titled sections with nothing to say which item each
				// belongs to.
				groups := make([]NameableGroup[*gurps.EquipmentModifier], 0, len(targets))
				for _, target := range targets {
					clones := make([]*gurps.EquipmentModifier, 0, len(tableDragData.Rows))
					for _, row := range tableDragData.Rows {
						clones = append(clones, row.Data().Clone(libraryFile, dataOwner, nil, gurps.Reference))
					}
					target.Modifiers = append(target.Modifiers, clones...)
					groups = append(groups, NameableGroup[*gurps.EquipmentModifier]{Label: target.String(), Rows: clones})
				}
				p.table.SyncToModel()
				if !xreflect.IsNil(dataOwner) {
					if entity := dataOwner.OwningEntity(); entity != nil {
						// Rebuilding is also what reports the drop when the rows belong to an entity (see
						// dropRebuilder), so the owner is rebuilt as modified rather than just rebuilt.
						rebuildAsModified(dropRebuilder(p.table), true)
						// That rebuild can have replaced this very list: an enabled modifier carrying a
						// switchable feature gives the rows it was dropped onto switchable features, which brings
						// the switch column into view, and a list can only change its columns by building a new
						// table. p belongs to the list that was replaced and its table field is never updated, so
						// each prompt below has to be aimed at the table that took its place -- an orphan has no
						// Rebuildable above it, so the rebuild its answer asks for would silently be skipped. The
						// lookup is made twice because answering the modifier prompt rebuilds as well, which can
						// replace the list a second time.
						//
						// The modifier prompt has to be given the rows the modifiers were dropped onto, since
						// modifiers themselves aren't something ProcessModifiers can process, and only the
						// topmost of them, since it walks each row's descendants as well.
						ProcessModifiers(liveTable(p.table), minimalNodes(targets))
						ProcessNameableGroups(liveTable(p.table), groups)
					}
				}
			}
		},
	}
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
	var sheetSettings *gurps.SheetSettings
	if p.forPage {
		if entity := p.DataOwner().OwningEntity(); entity != nil {
			sheetSettings = entity.SheetSettings
		} else {
			sheetSettings = gurps.GlobalSettings().SheetSettings()
		}
	}
	if p.forPage && sheetSettings != nil {
		if !sheetSettings.HideTLColumn {
			columnIDs = append(columnIDs, gurps.EquipmentTLColumn)
		}
		if !sheetSettings.HideLCColumn {
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
	if p.forPage {
		if sheetSettings == nil || !sheetSettings.HidePageRefColumn {
			columnIDs = append(columnIDs, gurps.EquipmentReferenceColumn)
		}
		if sheetSettings == nil || !sheetSettings.HideSourceMismatch {
			columnIDs = append(columnIDs, gurps.EquipmentLibSrcColumn)
		}
	} else {
		columnIDs = append(columnIDs, gurps.EquipmentReferenceColumn)
	}
	return columnIDs
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
