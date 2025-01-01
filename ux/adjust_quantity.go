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
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type adjustQuantityListUndoEdit = *unison.UndoEdit[*adjustQuantityList]

type adjustQuantityList struct {
	Owner Rebuildable
	List  []*quantityAdjuster
}

func (a *adjustQuantityList) Apply() {
	for _, one := range a.List {
		one.Apply()
	}
	a.Finish()
}

func (a *adjustQuantityList) Finish() {
	gurps.EntityFromNode(a.List[0].Target).Recalculate()
	MarkModified(a.Owner)
}

type quantityAdjuster struct {
	Target   *gurps.Equipment
	Quantity fxp.Int
}

func newQuantityAdjuster(target *gurps.Equipment) *quantityAdjuster {
	return &quantityAdjuster{
		Target:   target,
		Quantity: target.Quantity,
	}
}

func (a *quantityAdjuster) Apply() {
	a.Target.Quantity = a.Quantity
}

func canAdjustQuantity(table *unison.Table[*Node[*gurps.Equipment]], increment bool) bool {
	for _, row := range table.SelectedRows(false) {
		if eqp := row.Data(); eqp != nil {
			if increment || eqp.Quantity > 0 {
				return true
			}
		}
	}
	return false
}

func adjustQuantity(owner Rebuildable, table *unison.Table[*Node[*gurps.Equipment]], increment bool) {
	before := &adjustQuantityList{Owner: owner}
	after := &adjustQuantityList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if eqp := row.Data(); eqp != nil {
			if increment || eqp.Quantity > 0 {
				before.List = append(before.List, newQuantityAdjuster(eqp))
				original := eqp.Quantity
				qty := original.Trunc()
				if increment {
					qty += fxp.One
				} else if original == qty {
					qty -= fxp.One
				}
				eqp.Quantity = qty.Max(0)
				after.List = append(after.List, newQuantityAdjuster(eqp))
			}
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			var name string
			if increment {
				name = i18n.Text("Increment Quantity")
			} else {
				name = i18n.Text("Decrement Quantity")
			}
			mgr.Add(&unison.UndoEdit[*adjustQuantityList]{
				ID:         unison.NextUndoID(),
				EditName:   name,
				UndoFunc:   func(edit adjustQuantityListUndoEdit) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit adjustQuantityListUndoEdit) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		before.Finish()
	}
}
