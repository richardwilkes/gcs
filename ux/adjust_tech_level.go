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
	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps"
	"github.com/richardwilkes/unison"
)

func techLevelExtractor[T gurps.NodeTypes](amount fxp.Int) func(T) (gurps.TechLevelProvider[T], bool) {
	return func(data T) (gurps.TechLevelProvider[T], bool) {
		if provider, ok := any(data).(gurps.TechLevelProvider[T]); ok && provider.RequiresTL() {
			if _, changed := gurps.AdjustTechLevel(provider.TL(), amount); changed {
				return provider, true
			}
		}
		return nil, false
	}
}

func canAdjustTechLevel[T gurps.NodeTypes](table *unison.Table[*Node[T]], amount fxp.Int) bool {
	return canAdjustSelection(table, techLevelExtractor[T](amount))
}

func adjustTechLevel[T gurps.NodeTypes](owner Rebuildable, table *unison.Table[*Node[T]], amount fxp.Int) {
	title := increaseTechLevelAction.Title
	if amount < 0 {
		title = decreaseTechLevelAction.Title
	}
	adjustSelection(title, owner, table, techLevelExtractor[T](amount),
		func(p gurps.TechLevelProvider[T]) string { return p.TL() },
		func(p gurps.TechLevelProvider[T], v string) { p.SetTL(v) },
		func(p gurps.TechLevelProvider[T]) {
			tl, _ := gurps.AdjustTechLevel(p.TL(), amount)
			p.SetTL(tl)
		},
		true, false)
}
