// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper

// ToPixels converts the given length in this Units to the number of 72-pixels-per-inch pixels it represents.
func (enum Unit) ToPixels(length float64) float32 {
	switch enum {
	case Inch:
		return float32(length * 72)
	case Centimeter:
		return float32((length * 72) / 2.54)
	case Millimeter:
		return float32((length * 72) / 25.4)
	default:
		return Inch.ToPixels(length)
	}
}
