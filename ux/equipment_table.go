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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

// equipmentListProvider holds the two lists of an equipment list file. Unlike the other list files, it is its own data
// owner, since the equipment needs a weight unit to report its weights in.
type equipmentListProvider struct {
	carried fileListProvider[*gurps.Equipment]
	other   fileListProvider[*gurps.Equipment]
}

func (p *equipmentListProvider) DataOwner() gurps.DataOwner {
	return p
}

func (p *equipmentListProvider) OwningEntity() *gurps.Entity {
	return nil
}

func (p *equipmentListProvider) SourceMatcher() *gurps.SrcMatcher {
	return nil
}

func (p *equipmentListProvider) WeightUnit() fxp.WeightUnit {
	return gurps.GlobalSettings().SheetSettings().DefaultWeightUnits
}

func (p *equipmentListProvider) CarriedEquipmentList() []*gurps.Equipment {
	return p.carried.rows()
}

func (p *equipmentListProvider) SetCarriedEquipmentList(list []*gurps.Equipment) {
	p.carried.setRows(list)
}

func (p *equipmentListProvider) OtherEquipmentList() []*gurps.Equipment {
	return p.other.rows()
}

func (p *equipmentListProvider) SetOtherEquipmentList(list []*gurps.Equipment) {
	p.other.setRows(list)
}

// NewEquipmentTableDockableFromFile loads a list of equipment from a file and creates a new unison.Dockable for them.
func NewEquipmentTableDockableFromFile(filePath string) (unison.Dockable, error) {
	return openDockableFromFile(filePath, gurps.NewEquipmentFromFile, NewEquipmentTableDockable)
}

// NewEquipmentTableDockable creates a new unison.Dockable for equipment list files.
func NewEquipmentTableDockable(filePath string, equipment []*gurps.Equipment) *TableDockable[*gurps.Equipment] {
	provider := &equipmentListProvider{other: fileListProvider[*gurps.Equipment]{list: equipment}}
	d := NewTableDockable(filePath, gurps.EquipmentExt, NewEquipmentProvider(provider, false, false),
		func(path string) error { return gurps.SaveEquipment(provider.OtherEquipmentList(), path) },
		NewOtherEquipmentItemID, NewOtherEquipmentContainerItemID)
	InstallContainerConversionHandlers(d, d, d.table)
	d.InstallCmdHandlers(IncrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(d.table, fxp.One) },
		func(_ any) { adjustTechLevel(d, d.table, fxp.One) })
	d.InstallCmdHandlers(DecrementTechLevelItemID,
		func(_ any) bool { return canAdjustTechLevel(d.table, -fxp.One) },
		func(_ any) { adjustTechLevel(d, d.table, -fxp.One) })
	d.InstallCmdHandlers(IncrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(d.table, fxp.One) },
		func(_ any) { adjustEquipmentLevel(d, d.table, fxp.One) })
	d.InstallCmdHandlers(DecrementEquipmentLevelItemID,
		func(_ any) bool { return canAdjustEquipmentLevel(d.table, -fxp.One) },
		func(_ any) { adjustEquipmentLevel(d, d.table, -fxp.One) })
	return d
}
