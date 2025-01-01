// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package paper

// Dimensions returns the paper dimensions after orienting the paper.
func (enum Orientation) Dimensions(width, height Length) (adjustedWidth, adjustedHeight Length) {
	switch enum {
	case Portrait:
		return width, height
	case Landscape:
		return height, width
	default:
		return Portrait.Dimensions(width, height)
	}
}
