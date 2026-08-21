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

func addPreconfigurable[N gurps.Node[N], D gurps.EditorData[N]](e *editor[N, D], parent *unison.Panel) {
	if !HasOwner[*Template](parent) {
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

		// This panel serves only to fill the space where a label would normally be located
		parent.AddChild(unison.NewPanel())
		addCheckBox(parent, i18n.Text("Preconfigured"), p.PreconfiguredRef())
	}
}

func shouldClearPreconfiguredFlag(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return false
	}
	if _, ok := unison.AncestorOrSelf[unison.Dockable](panel).(*Template); ok {
		return false
	}
	return true
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
