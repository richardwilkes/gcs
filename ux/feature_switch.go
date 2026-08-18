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
// column is only shown on character sheets (never in templates, loot sheets or library lists), and only when at least
// one row, at any depth, has switchable features, so that sheets which don't make use of switchable features aren't
// cluttered by an empty column. Scanning every row is cheap relative to what building the headers already does (e.g.
// computing the extended weight and value of every piece of equipment).
func showSwitchColumn[T gurps.Node[T]](forPage bool, provider gurps.DataOwnerProvider, rows []T) bool {
	if !forPage {
		return false
	}
	if _, ok := provider.(*gurps.Entity); !ok {
		return false
	}
	found := false
	gurps.Traverse(func(one T) bool {
		if switcher, ok := any(one).(gurps.FeatureSwitcher); ok && switcher.HasSwitchableFeatures() {
			found = true
		}
		return found
	}, false, false, rows...)
	return found
}

// toggleFeatureSwitch sets the switch of the node's data (and, if includeDescendants is true, of everything contained
// within it) to the given state, registering an undoable edit and recalculating the owning entity.
func toggleFeatureSwitch[T gurps.Node[T]](n *Node[T], source unison.Paneler, on, includeDescendants bool) {
	switcher, ok := any(n.Data()).(gurps.FeatureSwitcher)
	if !ok {
		return
	}
	targets := []gurps.FeatureSwitcher{switcher}
	if includeDescendants {
		gurps.Traverse(func(one T) bool {
			if switcher, ok = any(one).(gurps.FeatureSwitcher); ok {
				targets = append(targets, switcher)
			}
			return false
		}, false, false, n.Data().NodeChildren()...)
	}
	adjustTargets(i18n.Text("Toggle Switch"), unison.AncestorOrSelf[Rebuildable](source), source,
		gurps.EntityFromNode(n.Data()), targets,
		func(s gurps.FeatureSwitcher) bool { return s.IsSwitchedOn() },
		func(s gurps.FeatureSwitcher, v bool) { s.SetSwitchedOn(v) },
		func(s gurps.FeatureSwitcher) { s.SetSwitchedOn(on) },
		false)
}
