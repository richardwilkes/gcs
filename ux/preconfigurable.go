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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/toolbox/v2/xreflect"
	"github.com/richardwilkes/unison"
)

// addPreconfigurable adds the Preconfigured check box, when the item may be marked preconfigured, anywhere but on a sheet,
// where the item's modifiers and nameables have already been settled and the mark is cleared.
func addPreconfigurable[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], parent *unison.Panel) {
	if HasOwner[*Sheet](parent) || HasOwner[*LootSheet](parent) {
		return
	}
	if p, ok := any(e.editorData).(gurps.Preconfigurable); ok && !xreflect.IsNil(p) {
		if e.target.Container() {
			if !p.CanPreconfigureContainer() {
				return
			}
		} else {
			if !p.CanPreconfigureItem() {
				return
			}
		}

		// This panel only fills the space where a label would normally be
		parent.AddChild(unison.NewPanel())
		addCheckBox(parent, i18n.Text("Preconfigured"), p.PreconfiguredRef())
	}
}

// shouldClearPreconfiguredFlag returns true if the panel belongs to a sheet, the one place the Preconfigured flag means
// nothing. Templates and libraries keep it, so that it is honored when their rows are copied onward.
func shouldClearPreconfiguredFlag(panel unison.Paneler) bool {
	return !xreflect.IsNil(panel) && transferKindOf(panel) == transferSheet
}

func clearPreconfiguredFlag[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) bool {
	if !shouldClearPreconfiguredFlag(table) {
		return false
	}
	if rows == nil {
		rows = table.SelectedRows(true)
	}
	var changes bool
	for _, row := range rows {
		node := row.Data()
		if tl, ok := any(node).(gurps.Preconfigurable); ok && !xreflect.IsNil(node) && tl.IsPreconfigured() {
			changes = true
			tl.SetPreconfigured(false)
		}
	}
	return changes
}
