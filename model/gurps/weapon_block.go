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

// WeaponBlock holds the block data for a weapon.
type WeaponBlock struct {
	Permitted bool
	Modifier  fxp.Int
}

// ParseWeaponBlock parses a string into a WeaponBlock.
func ParseWeaponBlock(s string) WeaponBlock {
	var wb WeaponBlock
	s = strings.ToLower(s)
	wb.Permitted = !strings.Contains(s, "no")
	if wb.Permitted {
		wb.Modifier, _ = fxp.Extract(s)
	}
	wb.Validate()
	return wb
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wb WeaponBlock) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wb.Modifier)
	_ = binary.Write(h, binary.LittleEndian, wb.Permitted)
}

// Resolve any bonuses that apply.
func (wb WeaponBlock) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponBlock {
	result := wb
	result.Permitted = w.ResolveBoolFlag(CanBlockWeaponSwitchType, result.Permitted)
	if result.Permitted {
		if pc := w.PC(); pc != nil {
			var primaryTooltip *xio.ByteBuffer
			if modifiersTooltip != nil {
				primaryTooltip = &xio.ByteBuffer{}
			}
			preAdj := w.skillLevelBaseAdjustment(pc, primaryTooltip)
			postAdj := w.skillLevelPostAdjustment(pc, primaryTooltip)
			best := fxp.Min
			for _, def := range w.Defaults {
				level := def.SkillLevelFast(pc, false, nil, true)
				if level == fxp.Min {
					continue
				}
				level += preAdj
				if def.Type() != BlockID {
					level = level.Div(fxp.Two).Trunc()
				}
				level += postAdj
				if best < level {
					best = level
				}
			}
			if best != fxp.Min {
				AppendBufferOntoNewLine(modifiersTooltip, primaryTooltip)
				result.Modifier += fxp.Three + best + pc.BlockBonus
				for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponBlockBonusFeatureType) {
					result.Modifier += bonus.AdjustedAmount()
				}
				result.Modifier = result.Modifier.Max(0).Trunc()
			} else {
				result.Modifier = 0
			}
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wb WeaponBlock) String() string {
	if !wb.Permitted {
		return i18n.Text("No")
	}
	if wb.Modifier == 0 {
		return ""
	}
	return wb.Modifier.String()
}

// Validate ensures that the data is valid.
func (wb *WeaponBlock) Validate() {
	if !wb.Permitted {
		wb.Modifier = 0
	}
}
