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

package model

// Dimensions returns the paper dimensions.
func (enum PaperSize) Dimensions() (width, height PaperLength) {
	switch enum {
	case LetterPaperSize:
		return PaperLength{Length: 8.5, Units: InchPaperUnits}, PaperLength{Length: 11, Units: InchPaperUnits}
	case LegalPaperSize:
		return PaperLength{Length: 8.5, Units: InchPaperUnits}, PaperLength{Length: 14, Units: InchPaperUnits}
	case TabloidPaperSize:
		return PaperLength{Length: 11, Units: InchPaperUnits}, PaperLength{Length: 17, Units: InchPaperUnits}
	case A0PaperSize:
		return PaperLength{Length: 841, Units: MillimeterPaperUnits}, PaperLength{Length: 1189, Units: MillimeterPaperUnits}
	case A1PaperSize:
		return PaperLength{Length: 594, Units: MillimeterPaperUnits}, PaperLength{Length: 841, Units: MillimeterPaperUnits}
	case A2PaperSize:
		return PaperLength{Length: 420, Units: MillimeterPaperUnits}, PaperLength{Length: 594, Units: MillimeterPaperUnits}
	case A3PaperSize:
		return PaperLength{Length: 297, Units: MillimeterPaperUnits}, PaperLength{Length: 420, Units: MillimeterPaperUnits}
	case A4PaperSize:
		return PaperLength{Length: 210, Units: MillimeterPaperUnits}, PaperLength{Length: 297, Units: MillimeterPaperUnits}
	case A5PaperSize:
		return PaperLength{Length: 148, Units: MillimeterPaperUnits}, PaperLength{Length: 210, Units: MillimeterPaperUnits}
	case A6PaperSize:
		return PaperLength{Length: 105, Units: MillimeterPaperUnits}, PaperLength{Length: 148, Units: MillimeterPaperUnits}
	default:
		return LetterPaperSize.Dimensions()
	}
}
