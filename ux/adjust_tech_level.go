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
	"github.com/richardwilkes/gcs/v5/model"
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type adjustTechLevelList[T model.NodeTypes] struct {
	Owner Rebuildable
	List  []*techLevelAdjuster[T]
}

func (a *adjustTechLevelList[T]) Apply() {
	for _, one := range a.List {
		one.Apply()
	}
	a.Finish()
}

func (a *adjustTechLevelList[T]) Finish() {
	entity := a.List[0].Target.OwningEntity()
	if entity != nil {
		entity.Recalculate()
	}
	MarkModified(a.Owner)
}

type techLevelAdjuster[T model.NodeTypes] struct {
	Target    model.TechLevelProvider[T]
	TechLevel string
}

func newTechLevelAdjuster[T model.NodeTypes](target model.TechLevelProvider[T]) *techLevelAdjuster[T] {
	return &techLevelAdjuster[T]{
		Target:    target,
		TechLevel: target.TL(),
	}
}

func (a *techLevelAdjuster[T]) Apply() {
	a.Target.SetTL(a.TechLevel)
}

func canAdjustTechLevel[T model.NodeTypes](table *unison.Table[*Node[T]], amount fxp.Int) bool {
	for _, row := range table.SelectedRows(false) {
		if provider, ok := any(row.Data()).(model.TechLevelProvider[T]); ok {
			if provider.RequiresTL() {
				if _, changed := model.AdjustTechLevel(provider.TL(), amount); changed {
					return true
				}
			}
		}
	}
	return false
}

func adjustTechLevel[T model.NodeTypes](owner Rebuildable, table *unison.Table[*Node[T]], amount fxp.Int) {
	before := &adjustTechLevelList[T]{Owner: owner}
	after := &adjustTechLevelList[T]{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if provider, ok := any(row.Data()).(model.TechLevelProvider[T]); ok {
			if provider.RequiresTL() {
				if tl, changed := model.AdjustTechLevel(provider.TL(), amount); changed {
					before.List = append(before.List, newTechLevelAdjuster[T](provider))
					provider.SetTL(tl)
					after.List = append(after.List, newTechLevelAdjuster[T](provider))
				}
			}
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			var name string
			if amount < 0 {
				name = i18n.Text("Decrease Tech Level")
			} else {
				name = i18n.Text("Increase Tech Level")
			}
			mgr.Add(&unison.UndoEdit[*adjustTechLevelList[T]]{
				ID:         unison.NextUndoID(),
				EditName:   name,
				UndoFunc:   func(edit *unison.UndoEdit[*adjustTechLevelList[T]]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*adjustTechLevelList[T]]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		before.Finish()
	}
}
