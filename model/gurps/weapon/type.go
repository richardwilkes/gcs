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

package weapon

import (
	"github.com/richardwilkes/gcs/v5/res"
	"github.com/richardwilkes/unison"
)

// SVG returns the SVG that should be used for this type.
func (enum Type) SVG() *unison.SVG {
	switch enum {
	case Melee:
		return res.MeleeWeaponSVG
	case Ranged:
		return res.RangedWeaponSVG
	default:
		return nil
	}
}
