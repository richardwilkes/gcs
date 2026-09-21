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

		// This panel only fills the space where a label would normally be
		parent.AddChild(unison.NewPanel())
		addCheckBox(parent, i18n.Text("Preconfigured"), p.PreconfiguredRef())
	}
}

// allowPreconfiguredFlag reports whether rows landing on the given panel may keep their Preconfigured flag. Only the
// instance documents -- a character sheet and a loot sheet -- clear it. Everywhere else the flag is authored data: a
// template is where it is set in the first place, and a library list holds items that have been resolved once so that
// they can be dropped onto a sheet without asking again. The test used to be "anywhere but a template", which meant a
// drop into a library list quietly threw that away.
func allowPreconfiguredFlag(panel unison.Paneler) bool {
	if xreflect.IsNil(panel) {
		return true
	}
	switch unison.AncestorOrSelf[unison.Dockable](panel).(type) {
	case *Sheet, *LootSheet:
		return false
	default:
		return true
	}
}

// maybeClearPreconfiguredFlag clears the Preconfigured flag on the given rows, and everything beneath them, when the
// table they landed in is one that doesn't allow the flag. Passing nil rows uses the table's selection.
func maybeClearPreconfiguredFlag[T gurps.Node[T]](table *unison.Table[*Node[T]], rows []*Node[T]) bool {
	if allowPreconfiguredFlag(table) {
		return false
	}
	if rows == nil {
		rows = table.SelectedRows(true)
	}
	return clearPreconfiguredFlag(rows)
}

// clearPreconfiguredFlag clears the Preconfigured flag on the given rows and everything beneath them, returning
// whether anything changed.
//
// The descent matters as much as the rows themselves. A selection holds only the shallowest rows of each branch, and
// only those that are showing, so a row inside a closed container isn't in it at all -- yet the flag has to come off
// everything that arrived, not just what the user can see. A container that is itself preconfigured is cleared and
// then descended into as well: the flag says a node's own selections and substitutions are settled, and says nothing
// about its children's.
func clearPreconfiguredFlag[T gurps.Node[T]](rows []*Node[T]) bool {
	nodes := make([]T, 0, len(rows))
	for _, row := range rows {
		if node := row.Data(); !xreflect.IsNil(node) {
			nodes = append(nodes, node)
		}
	}
	var changes bool
	gurps.Traverse(func(node T) bool {
		if p, ok := any(node).(gurps.Preconfigurable); ok && !xreflect.IsNil(p) && p.IsPreconfigured() {
			changes = true
			p.SetPreconfigured(false)
		}
		return false
	}, false, false, nodes...)
	return changes
}
