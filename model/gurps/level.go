// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import "github.com/richardwilkes/gcs/v5/model/fxp"

// Level provides a level & relative level pair, plus a tooltip.
type Level struct {
	Level         fxp.Int
	RelativeLevel fxp.Int
	Tooltip       string
}

// LevelAsString returns the level as a string.
func (l Level) LevelAsString(forContainer bool) string {
	if forContainer {
		return ""
	}
	level := l.Level.Trunc()
	if level <= 0 {
		return "-"
	}
	return level.String()
}
