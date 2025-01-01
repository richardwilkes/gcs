// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package difficulty

import (
	"github.com/richardwilkes/gcs/v5/model/fxp"
)

// TechniqueLevels holds all possible values when used with Techniques.
var TechniqueLevels = []Level{
	Average,
	Hard,
}

// BaseRelativeLevel returns the base relative skill level at 0 points.
func (enum Level) BaseRelativeLevel() fxp.Int {
	switch enum {
	case Easy:
		return 0
	case Average:
		return -fxp.One
	case Hard:
		return -fxp.Two
	case VeryHard, Wildcard:
		return -fxp.Three
	default:
		return Easy.BaseRelativeLevel()
	}
}
