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

func skillLevelExtractor[T gurps.Node[T]](increment bool) func(T) (gurps.SkillAdjustmentProvider, bool) {
	return func(data T) (gurps.SkillAdjustmentProvider, bool) {
		if provider, ok := any(data).(gurps.SkillAdjustmentProvider); ok && !provider.Container() &&
			(increment || provider.RawPoints() > 0) {
			return provider, true
		}
		return nil, false
	}
}

func canAdjustSkillLevel[T gurps.Node[T]](table *unison.Table[*Node[T]], increment bool) bool {
	return canAdjustSelection(table, skillLevelExtractor[T](increment))
}

func adjustSkillLevel[T gurps.Node[T]](owner Rebuildable, table *unison.Table[*Node[T]], increment bool) {
	title := increaseSkillLevelAction.Title
	if !increment {
		title = decreaseSkillLevelAction.Title
	}
	adjustSelection(title, owner, table, skillLevelExtractor[T](increment),
		func(p gurps.SkillAdjustmentProvider) fxp.Int { return p.RawPoints() },
		func(p gurps.SkillAdjustmentProvider, v fxp.Int) { p.SetRawPoints(v) },
		func(p gurps.SkillAdjustmentProvider) {
			if increment {
				p.IncrementSkillLevel()
			} else {
				p.DecrementSkillLevel()
			}
		},
		true, false)
}
