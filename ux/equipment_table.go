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
	"os"
	"path/filepath"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type equipmentListProvider struct {
	carried []*gurps.Equipment
	other   []*gurps.Equipment
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
	return p.carried
}

func (p *equipmentListProvider) SetCarriedEquipmentList(list []*gurps.Equipment) {
	p.carried = list
}

func (p *equipmentListProvider) OtherEquipmentList() []*gurps.Equipment {
	return p.other
}

func (p *equipmentListProvider) SetOtherEquipmentList(list []*gurps.Equipment) {
	p.other = list
}

// NewEquipmentTableDockableFromFile loads a list of equipment from a file and creates a new unison.Dockable for them.
func NewEquipmentTableDockableFromFile(filePath string) (unison.Dockable, error) {
	equipment, err := gurps.NewEquipmentFromFile(os.DirFS(filepath.Dir(filePath)), filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	d := NewEquipmentTableDockable(filePath, equipment)
	d.needsSaveAsPrompt = false
	return d, nil
}

// NewEquipmentTableDockable creates a new unison.Dockable for equipment list files.
func NewEquipmentTableDockable(filePath string, equipment []*gurps.Equipment) *TableDockable[*gurps.Equipment] {
	provider := &equipmentListProvider{other: equipment}
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
