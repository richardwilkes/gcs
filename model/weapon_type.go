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

import (
	"github.com/richardwilkes/gcs/v5/svg"
	"github.com/richardwilkes/unison"
)

// SVG returns the SVG that should be used for this type.
func (enum WeaponType) SVG() *unison.SVG {
	switch enum {
	case MeleeWeaponType:
		return svg.MeleeWeapon
	case RangedWeaponType:
		return svg.RangedWeapon
	default:
		return nil
	}
}
