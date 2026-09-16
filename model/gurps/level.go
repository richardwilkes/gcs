// Copyright (c) 1998-2026 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"hash"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/v2/xhash"
)

// Level provides a level & relative level pair, plus a tooltip.
type Level struct {
	Level         fxp.Int
	RelativeLevel fxp.Int
	Tooltip       string
}

// Hash writes this object's contents into the hasher.
func (l Level) Hash(h hash.Hash) {
	xhash.Num64(h, l.Level)
	xhash.Num64(h, l.RelativeLevel)
	xhash.StringWithLen(h, l.Tooltip)
}

// LevelAsString returns the floored level as a string, or an empty string for a container and "-" when the level isn't
// positive.
func (l Level) LevelAsString(forContainer bool) string {
	if forContainer {
		return ""
	}
	level := l.Level.Floor()
	if level <= 0 {
		return "-"
	}
	return level.String()
}
