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
	"github.com/richardwilkes/toolbox/v2/i18n"
	"github.com/richardwilkes/unison"
)

func rawPointsExtractor[T gurps.NodeTypes](increment bool) func(T) (gurps.RawPointsAdjuster[T], bool) {
	return func(data T) (gurps.RawPointsAdjuster[T], bool) {
		if provider, ok := any(data).(gurps.RawPointsAdjuster[T]); ok && !provider.Container() &&
			(increment || provider.RawPoints() > 0) {
			return provider, true
		}
		return nil, false
	}
}

func canAdjustRawPoints[T gurps.NodeTypes](table *unison.Table[*Node[T]], increment bool) bool {
	return canAdjustSelection(table, rawPointsExtractor[T](increment))
}

func adjustRawPoints[T gurps.NodeTypes](owner Rebuildable, table *unison.Table[*Node[T]], increment bool) {
	title := i18n.Text("Decrement Points")
	if increment {
		title = i18n.Text("Increment Points")
	}
	adjustSelection(title, owner, table, rawPointsExtractor[T](increment),
		func(p gurps.RawPointsAdjuster[T]) fxp.Int { return p.RawPoints() },
		func(p gurps.RawPointsAdjuster[T], v fxp.Int) { p.SetRawPoints(v) },
		func(p gurps.RawPointsAdjuster[T]) {
			rawPts := p.RawPoints()
			pts := rawPts.Floor()
			if increment {
				pts += fxp.One
			} else if rawPts == pts {
				pts -= fxp.One
			}
			p.SetRawPoints(pts.Max(0))
		},
		true, false)
}
