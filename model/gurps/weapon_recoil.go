/*
 * Copyright ©1998-2023 by Richard A. Wilkes. All rights reserved.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, version 2.0. If a copy of the MPL was not distributed with
 * this file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * This Source Code Form is "Incompatible With Secondary Licenses", as
 * defined by the Mozilla Public License, version 2.0.
 */

package gurps

import (
	"encoding/binary"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponRecoil holds the recoil data for a weapon.
type WeaponRecoil struct {
	ShotRecoil fxp.Int
	SlugRecoil fxp.Int
}

// ParseWeaponRecoil parses a string into a WeaponRecoil.
func ParseWeaponRecoil(s string) WeaponRecoil {
	var wr WeaponRecoil
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	parts := strings.Split(s, "/")
	wr.ShotRecoil, _ = fxp.Extract(parts[0])
	if len(parts) > 1 {
		wr.SlugRecoil, _ = fxp.Extract(parts[1])
	}
	return wr
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wr WeaponRecoil) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wr.ShotRecoil)
	_ = binary.Write(h, binary.LittleEndian, wr.SlugRecoil)
}

// Resolve any bonuses that apply.
func (wr WeaponRecoil) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRecoil {
	result := wr
	// 0 means recoil isn't used; 1+ means it is.
	if wr.ShotRecoil > 0 || wr.SlugRecoil > 0 {
		for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponRecoilBonusFeatureType) {
			result.ShotRecoil += bonus.AdjustedAmount()
			result.SlugRecoil += bonus.AdjustedAmount()
		}
		if wr.ShotRecoil > 0 {
			result.ShotRecoil = result.ShotRecoil.Max(fxp.One)
		} else {
			result.ShotRecoil = 0
		}
		if wr.SlugRecoil > 0 {
			result.SlugRecoil = result.SlugRecoil.Max(fxp.One)
		} else {
			result.SlugRecoil = 0
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRecoil) String() string {
	if wr.ShotRecoil == 0 && wr.SlugRecoil == 0 {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wr.ShotRecoil.String())
	if wr.SlugRecoil != 0 {
		buffer.WriteByte('/')
		buffer.WriteString(wr.SlugRecoil.String())
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wr WeaponRecoil) Tooltip() string {
	if wr.ShotRecoil == wr.SlugRecoil || (wr.ShotRecoil == 0 && wr.SlugRecoil == 0) {
		return ""
	}
	return i18n.Text("First Recoil value is for shot, second is for slugs")
}

// Validate ensures that the data is valid.
func (wr *WeaponRecoil) Validate() {
	wr.ShotRecoil = wr.ShotRecoil.Max(0)
	wr.SlugRecoil = wr.SlugRecoil.Max(0)
}
