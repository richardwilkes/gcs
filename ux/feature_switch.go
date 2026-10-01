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
	"github.com/richardwilkes/unison"
)

// showSwitchColumn returns true if a page list showing the given rows should have the switch column: only on character
// sheets (not loot sheets, templates or library lists), and only when some row, at any depth, has switchable features,
// so sheets that don't use them aren't cluttered by an empty column. Elsewhere, switches are thrown from editors.
func showSwitchColumn[T gurps.Node[T]](forPage bool, provider gurps.DataOwnerProvider, rows []T) bool {
	if !forPage {
		return false
	}
	if _, ok := provider.(*gurps.Entity); !ok {
		return false
	}
	return anySwitchable(rows)
}

// anySwitchable returns true if any of the given rows, at any depth, has switchable features. This is deliberately not
// written in terms of gurps.Traverse, since that allocates a copy of the children of every container it descends into
// and this is called for every page list, more than once, every time a sheet is rebuilt.
func anySwitchable[T gurps.Node[T]](rows []T) bool {
	for _, row := range rows {
		if switcher, ok := any(row).(gurps.FeatureSwitcher); ok && switcher.HasSwitchableFeatures() {
			return true
		}
		if row.HasChildren() && anySwitchable(row.NodeChildren()) {
			return true
		}
	}
	return false
}

// toggleFeatureSwitch sets the switch of the node's data to the given state as an undoable edit and recalculates the
// owning entity. If includeDescendants is true, everything within it that has switchable features is set too; the rest
// are left out, since throwing their switch would change nothing visible yet still show up as a change to the sheet.
//
// The owner is rebuilt rather than merely marked as modified, since a switchable feature can be a reaction, conditional
// modifier or weapon bonus, and whether the lists showing those appear on the page at all -- along with which columns
// the weapon lists hold -- is decided only when the owner creates its lists.
//
// It returns false, changing nothing, if the node's data has no switch. That can't happen for a switch cell, but the
// caller uses the answer to put the cell back rather than show a state the model never took on.
func toggleFeatureSwitch[T gurps.Node[T]](n *Node[T], source unison.Paneler, on, includeDescendants bool) bool {
	switcher, ok := any(n.Data()).(gurps.FeatureSwitcher)
	if !ok {
		return false
	}
	targets := []gurps.FeatureSwitcher{switcher}
	if includeDescendants {
		gurps.Traverse(func(one T) bool {
			if other, ok2 := any(one).(gurps.FeatureSwitcher); ok2 && other.HasSwitchableFeatures() {
				targets = append(targets, other)
			}
			return false
		}, false, false, n.Data().NodeChildren()...)
	}
	adjustTargets(i18n.Text("Toggle Switch"), unison.AncestorOrSelf[Rebuildable](source), source,
		gurps.EntityFromNode(n.Data()), targets,
		func(s gurps.FeatureSwitcher) bool { return s.IsSwitchedOn() },
		func(s gurps.FeatureSwitcher, v bool) { s.SetSwitchedOn(v) },
		func(s gurps.FeatureSwitcher) { s.SetSwitchedOn(on) },
		true)
	return true
}
