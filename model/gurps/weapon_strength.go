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
	"fmt"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponStrength holds the minimum strength data for a weapon.
type WeaponStrength struct {
	Minimum          fxp.Int
	Bipod            bool
	Mounted          bool
	MusketRest       bool
	TwoHanded        bool
	TwoHandedUnready bool
}

// ParseWeaponStrength parses a string into a WeaponStrength.
func ParseWeaponStrength(s string) WeaponStrength {
	var ws WeaponStrength
	s = strings.ReplaceAll(s, " ", "")
	if s != "" {
		s = strings.ToLower(s)
		ws.Bipod = strings.Contains(s, "b")
		ws.Mounted = strings.Contains(s, "m")
		ws.MusketRest = strings.Contains(s, "r")
		ws.TwoHanded = strings.Contains(s, "†") || strings.Contains(s, "*") // bad input in some files had * instead of †
		ws.TwoHandedUnready = strings.Contains(s, "‡")
		ws.Minimum, _ = fxp.Extract(s)
		ws.Validate()
	}
	return ws
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (ws WeaponStrength) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, ws.Minimum)
	_ = binary.Write(h, binary.LittleEndian, ws.Bipod)
	_ = binary.Write(h, binary.LittleEndian, ws.Mounted)
	_ = binary.Write(h, binary.LittleEndian, ws.MusketRest)
	_ = binary.Write(h, binary.LittleEndian, ws.TwoHanded)
	_ = binary.Write(h, binary.LittleEndian, ws.TwoHandedUnready)
}

// Resolve any bonuses that apply.
func (ws WeaponStrength) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponStrength {
	result := ws
	if w.Owner != nil {
		if st := w.Owner.RatedStrength().Max(0); st != 0 {
			result.Minimum = st
		}
	}
	result.Bipod = w.ResolveBoolFlag(BipodWeaponSwitchType, result.Bipod)
	result.Mounted = w.ResolveBoolFlag(MountedWeaponSwitchType, result.Mounted)
	result.MusketRest = w.ResolveBoolFlag(MusketRestWeaponSwitchType, result.MusketRest)
	result.TwoHanded = w.ResolveBoolFlag(TwoHandedWeaponSwitchType, result.TwoHanded)
	result.TwoHandedUnready = w.ResolveBoolFlag(TwoHandedAndUnreadyAfterAttackWeaponSwitchType, result.TwoHandedUnready)
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponMinSTBonusFeatureType) {
		result.Minimum += bonus.AdjustedAmount()
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (ws WeaponStrength) String() string {
	var buffer strings.Builder
	if ws.Minimum > 0 {
		buffer.WriteString(ws.Minimum.String())
	}
	if ws.Bipod {
		buffer.WriteByte('B')
	}
	if ws.Mounted {
		buffer.WriteByte('M')
	}
	if ws.MusketRest {
		buffer.WriteByte('R')
	}
	if ws.TwoHanded || ws.TwoHandedUnready {
		if ws.TwoHandedUnready {
			buffer.WriteRune('‡')
		} else {
			buffer.WriteRune('†')
		}
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (ws WeaponStrength) Tooltip(w *Weapon) string {
	var tooltip strings.Builder
	if w.Owner != nil {
		if st := w.Owner.RatedStrength(); st > 0 {
			fmt.Fprintf(&tooltip, i18n.Text("The weapon has a rated ST of %v, which is used instead of the user's ST for calculations."), st)
		}
	}
	if ws.Minimum > 0 {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		fmt.Fprintf(&tooltip, i18n.Text("The weapon has a minimum ST of %v. If your ST is less than this, you will suffer a -1 to weapon skill per point of ST you lack and lose one extra FP at the end of any fight that lasts long enough to cost FP."), ws.Minimum)
	}
	if ws.Bipod {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("Has an attached bipod. When used from a prone position, "))
		reducedST := ws.Minimum.Mul(fxp.Two).Div(fxp.Three).Ceil()
		if reducedST > 0 && reducedST != ws.Minimum {
			fmt.Fprintf(&tooltip, i18n.Text("reduces the ST requirement to %v and"), reducedST)
		}
		tooltip.WriteString(i18n.Text("treats the attack as braced (add +1 to Accuracy)."))
	}
	if ws.Mounted {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("Mounted. Ignore listed ST and Bulk when firing from its mount. Takes at least 3 Ready maneuvers to unmount or remount the weapon."))
	}
	if ws.MusketRest {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("Uses a Musket Rest. Any aimed shot fired while stationary and standing up is automatically braced (add +1 to Accuracy)."))
	}
	if ws.TwoHanded || ws.TwoHandedUnready {
		if tooltip.Len() != 0 {
			tooltip.WriteString("\n\n")
		}
		tooltip.WriteString(i18n.Text("Requires two hands"))
		if ws.TwoHandedUnready {
			fmt.Fprintf(&tooltip, i18n.Text(" and becomes unready after you attack with it. If you have at least ST %v, you can used it two-handed without it becoming unready. If you have at least ST %v, you can use it one-handed with no readiness penalty."), ws.Minimum.Mul(fxp.OneAndAHalf).Ceil(), ws.Minimum.Mul(fxp.Three).Ceil())
		} else {
			fmt.Fprintf(&tooltip, i18n.Text(". If you have at least ST %v, you can use it one-handed, but it becomes unready after you attack with it. If you have at least ST %v, you can use it one-handed with no readiness penalty."), ws.Minimum.Mul(fxp.OneAndAHalf).Ceil(), ws.Minimum.Mul(fxp.Two).Ceil())
		}
	}
	return tooltip.String()
}

// Validate ensures that the data is valid.
func (ws *WeaponStrength) Validate() {
	ws.Minimum = ws.Minimum.Max(0)
	if ws.TwoHanded && ws.TwoHandedUnready {
		ws.TwoHanded = false
	}
}
