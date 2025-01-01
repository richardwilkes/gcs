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
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/unison"
)

type toggleDisabledUndoEdit = *unison.UndoEdit[*toggleDisabledList]

type toggleDisabledList struct {
	Owner Rebuildable
	List  []*disabledAdjuster
}

func (a *toggleDisabledList) Apply() {
	for _, one := range a.List {
		one.Apply()
	}
	a.Finish()
}

func (a *toggleDisabledList) Finish() {
	gurps.EntityFromNode(a.List[0].Target).Recalculate()
	MarkModified(a.Owner)
	a.Owner.Rebuild(true)
}

type disabledAdjuster struct {
	Target   *gurps.Trait
	Disabled bool
}

func newDisabledAdjuster(target *gurps.Trait) *disabledAdjuster {
	return &disabledAdjuster{
		Target:   target,
		Disabled: target.Disabled,
	}
}

func (a *disabledAdjuster) Apply() {
	a.Target.Disabled = a.Disabled
}

func canToggleDisabled(table *unison.Table[*Node[*gurps.Trait]]) bool {
	for _, row := range table.SelectedRows(false) {
		if t := row.Data(); t != nil {
			return true
		}
	}
	return false
}

func toggleDisabled(owner Rebuildable, table *unison.Table[*Node[*gurps.Trait]]) {
	before := &toggleDisabledList{Owner: owner}
	after := &toggleDisabledList{Owner: owner}
	for _, row := range table.SelectedRows(false) {
		if t := row.Data(); t != nil {
			before.List = append(before.List, newDisabledAdjuster(t))
			t.Disabled = !t.Disabled
			after.List = append(after.List, newDisabledAdjuster(t))
		}
	}
	if len(before.List) > 0 {
		if mgr := unison.UndoManagerFor(table); mgr != nil {
			mgr.Add(&unison.UndoEdit[*toggleDisabledList]{
				ID:         unison.NextUndoID(),
				EditName:   i18n.Text("Toggle Enablement"),
				UndoFunc:   func(edit toggleDisabledUndoEdit) { edit.BeforeData.Apply() },
				RedoFunc:   func(edit toggleDisabledUndoEdit) { edit.AfterData.Apply() },
				BeforeData: before,
				AfterData:  after,
			})
		}
		before.Finish()
	}
}
