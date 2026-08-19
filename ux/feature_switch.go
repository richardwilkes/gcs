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

// showSwitchColumn returns true if the switch column should be present in a page list showing the given rows. The
// column is only shown on character sheets (never on loot sheets, in templates or in library lists), and only when at
// least one row, at any depth, has switchable features, so that sheets which don't make use of switchable features
// aren't cluttered by an empty column. Anywhere else, the switch of an item that has one is thrown from the item's
// editor.
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

// toggleFeatureSwitch sets the switch of the node's data to the given state, registering an undoable edit and
// recalculating the owning entity. If includeDescendants is true, everything contained within it that actually has
// something to switch is set as well. Descendants with no switchable features are deliberately left out: throwing their
// switch would have no effect the user could see, yet the new state would still be written to the file and show up as a
// change to the sheet. The node itself is always a target, since a switch cell -- the only thing that calls this -- is
// only created for an item whose HasSwitchableFeatures() is true, so the target list is never empty.
//
// The owner is rebuilt rather than merely marked as modified, since a switchable feature can be a reaction bonus, a
// conditional modifier bonus or a weapon bonus, and whether the lists showing those appear on the page at all -- along
// with which columns the weapon lists hold -- is decided only when the owner creates its lists.
//
// Returns false, without changing anything, if the node's data has no switch to throw. That can't happen for a switch
// cell, but the caller uses the answer to put the cell back the way it was rather than showing a state the model never
// took on.
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
