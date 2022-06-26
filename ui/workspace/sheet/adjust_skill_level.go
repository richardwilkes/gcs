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

package sheet

import (
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/gcs/v5/ui/widget"
	"github.com/richardwilkes/gcs/v5/ui/widget/ntable"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

func canAdjustSkillLevel[T gurps.NodeConstraint[T]](table *unison.Table[*ntable.Node[T]], increment bool) bool {
	for _, row := range table.SelectedRows(false) {
		if provider, ok := interface{}(row.Data()).(gurps.SkillAdjustmentProvider[T]); ok && !provider.Container() {
			if increment || provider.RawPoints() > 0 {
				return true
			}
		}
	}
	return false
}

func adjustSkillLevel[T gurps.NodeConstraint[T]](owner widget.Rebuildable, table *unison.Table[*ntable.Node[T]], increment bool) {
	before := &adjustRawPointsList[T]{Owner: owner}
	after := &adjustRawPointsList[T]{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if provider, ok := interface{}(row.Data()).(gurps.SkillAdjustmentProvider[T]); ok {
			if increment || provider.RawPoints() > 0 {
				before.List = append(before.List, newRawPointsAdjuster[T](provider))
				if increment {
					provider.IncrementSkillLevel()
				} else {
					provider.DecrementSkillLevel()
				}
				after.List = append(after.List, newRawPointsAdjuster[T](provider))
			}
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			var name string
			if increment {
				name = i18n.Text("Increase Skill Level")
			} else {
				name = i18n.Text("Decrease Skill Level")
			}
			mgr.Add(&unison.UndoEdit[*adjustRawPointsList[T]]{
				ID:         unison.NextUndoID(),
				EditName:   name,
				UndoFunc:   func(edit *unison.UndoEdit[*adjustRawPointsList[T]]) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit *unison.UndoEdit[*adjustRawPointsList[T]]) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		before.Finish()
	}
}
