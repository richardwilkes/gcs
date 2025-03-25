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

type adjustTechLevelList[T gurps.NodeTypes] struct {
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
	gurps.EntityFromNode(a.List[0].Target).Recalculate()
	MarkModified(a.Owner)
}

type techLevelAdjuster[T gurps.NodeTypes] struct {
	Target    gurps.TechLevelProvider[T]
	TechLevel string
}

func newTechLevelAdjuster[T gurps.NodeTypes](target gurps.TechLevelProvider[T]) *techLevelAdjuster[T] {
	return &techLevelAdjuster[T]{
		Target:    target,
		TechLevel: target.TL(),
	}
}

func (a *techLevelAdjuster[T]) Apply() {
	a.Target.SetTL(a.TechLevel)
}

func canAdjustTechLevel[T gurps.NodeTypes](table *unison.Table[*Node[T]], amount fxp.Int) bool {
	for _, row := range table.SelectedRows(false) {
		if provider, ok := any(row.Data()).(gurps.TechLevelProvider[T]); ok {
			if provider.RequiresTL() {
				if _, changed := gurps.AdjustTechLevel(provider.TL(), amount); changed {
					return true
				}
			}
		}
	}
	return false
}

func adjustTechLevel[T gurps.NodeTypes](owner Rebuildable, table *unison.Table[*Node[T]], amount fxp.Int) {
	before := &adjustTechLevelList[T]{Owner: owner}
	after := &adjustTechLevelList[T]{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if provider, ok := any(row.Data()).(gurps.TechLevelProvider[T]); ok {
			if provider.RequiresTL() {
				if tl, changed := gurps.AdjustTechLevel(provider.TL(), amount); changed {
					before.List = append(before.List, newTechLevelAdjuster(provider))
					provider.SetTL(tl)
					after.List = append(after.List, newTechLevelAdjuster(provider))
				}
			}
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			var name string
			if amount < 0 {
				name = decreaseTechLevelAction.Title
			} else {
				name = increaseTechLevelAction.Title
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
