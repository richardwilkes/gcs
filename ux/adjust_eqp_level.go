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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

type adjustEquipmentLevelList struct {
	Owner Rebuildable
	List  []*equipmentLevelAdjuster
}

func (a *adjustEquipmentLevelList) Apply() {
	for _, one := range a.List {
		one.Apply()
	}
	a.Finish()
}

func (a *adjustEquipmentLevelList) Finish() {
	gurps.EntityFromNode(a.List[0].Target).Recalculate()
	MarkModified(a.Owner)
}

type equipmentLevelAdjuster struct {
	Target         *gurps.Equipment
	EquipmentLevel fxp.Int
}

func newEquipmentLevelAdjuster(target *gurps.Equipment) *equipmentLevelAdjuster {
	return &equipmentLevelAdjuster{
		Target:         target,
		EquipmentLevel: target.Level,
	}
}

func (a *equipmentLevelAdjuster) Apply() {
	a.Target.Level = a.EquipmentLevel.Max(0)
}

func canAdjustEquipmentLevel(table *unison.Table[*Node[*gurps.Equipment]], amount fxp.Int) bool {
	for _, row := range table.SelectedRows(false) {
		if amount > 0 || row.Data().Level > 0 {
			return true
		}
	}
	return false
}

func adjustEquipmentLevel(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], amount fxp.Int) {
	before := &adjustEquipmentLevelList{Owner: owner}
	after := &adjustEquipmentLevelList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		equipment := row.Data()
		if amount > 0 || equipment.Level > 0 {
			before.List = append(before.List, newEquipmentLevelAdjuster(equipment))
			equipment.Level = (equipment.Level + amount).Max(0)
			after.List = append(after.List, newEquipmentLevelAdjuster(equipment))
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			var name string
			if amount < 0 {
				name = decreaseEquipmentLevelAction.Title
			} else {
				name = increaseEquipmentLevelAction.Title
			}
			mgr.Add(&unison.UndoEdit[*adjustEquipmentLevelList]{
				ID:         unison.NextUndoID(),
				EditName:   name,
				UndoFunc:   func(edit *unison.UndoEdit[*adjustEquipmentLevelList]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*adjustEquipmentLevelList]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		before.Finish()
	}
}
