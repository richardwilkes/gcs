/*
 * Copyright ©1998-2022 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package paper

// Dimensions returns the paper dimensions.
func (enum Size) Dimensions() (width, height Length) {
	switch enum {
	case Letter:
		return Length{Length: 8.5, Units: Inch}, Length{Length: 11, Units: Inch}
	case Legal:
		return Length{Length: 8.5, Units: Inch}, Length{Length: 14, Units: Inch}
	case Tabloid:
		return Length{Length: 11, Units: Inch}, Length{Length: 17, Units: Inch}
	case A0:
		return Length{Length: 841, Units: Millimeter}, Length{Length: 1189, Units: Millimeter}
	case A1:
		return Length{Length: 594, Units: Millimeter}, Length{Length: 841, Units: Millimeter}
	case A2:
		return Length{Length: 420, Units: Millimeter}, Length{Length: 594, Units: Millimeter}
	case A3:
		return Length{Length: 297, Units: Millimeter}, Length{Length: 420, Units: Millimeter}
	case A4:
		return Length{Length: 210, Units: Millimeter}, Length{Length: 297, Units: Millimeter}
	case A5:
		return Length{Length: 148, Units: Millimeter}, Length{Length: 210, Units: Millimeter}
	case A6:
		return Length{Length: 105, Units: Millimeter}, Length{Length: 148, Units: Millimeter}
	default:
		return Letter.Dimensions()
	}
}
