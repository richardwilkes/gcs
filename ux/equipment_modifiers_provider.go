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

const equipmentModifierRefKey = "equipment_modifier"

var _ TableProvider[*gurps.EquipmentModifier] = &eqpModProvider{}

type eqpModProvider struct {
	listProvider[*gurps.EquipmentModifier]
	forEditor bool
}

// NewEquipmentModifiersProvider creates a new table provider for equipment modifiers.
func NewEquipmentModifiersProvider(provider gurps.EquipmentModifierListProvider, forEditor bool) TableProvider[*gurps.EquipmentModifier] {
	p := &eqpModProvider{forEditor: forEditor}
	p.listProvider = listProvider[*gurps.EquipmentModifier]{
		dataOwner:  provider,
		list:       provider.EquipmentModifierList,
		setList:    provider.SetEquipmentModifierList,
		columnIDs:  p.ColumnIDs,
		headerData: gurps.EquipmentModifierHeaderData,
	}
	return p
}

func (p *eqpModProvider) RefKey() string {
	return equipmentModifierRefKey
}

func (p *eqpModProvider) DragKey() *uti.DataType {
	return equipmentModifierDragKey
}

func (p *eqpModProvider) DragSVG() *unison.SVG {
	return svg.GCSEquipmentModifiers
}

func (p *eqpModProvider) ItemNames() (singular, plural string) {
	return i18n.Text("Equipment Modifier"), i18n.Text("Equipment Modifiers")
}

func (p *eqpModProvider) ColumnIDs() []int {
	columnIDs := make([]int, 0, 7)
	if p.forEditor {
		columnIDs = append(columnIDs, gurps.EquipmentModifierEnabledColumn)
	}
	columnIDs = append(
		columnIDs,
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
	p.insertItems(owner, table, item)
	EditEquipmentModifier(owner, item)
}

func (p *eqpModProvider) ContextMenuItems() []ContextMenuItem {
	return AppendDefaultContextMenuItems([]ContextMenuItem{
		contextMenuItemFor(newEquipmentModifierAction),
		contextMenuItemFor(newEquipmentContainerModifierAction),
	})
}
