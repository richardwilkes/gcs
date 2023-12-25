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

// WeaponReach holds the reach data for a weapon.
type WeaponReach struct {
	Min                 fxp.Int
	Max                 fxp.Int
	CloseCombat         bool
	ChangeRequiresReady bool
}

// ParseWeaponReach parses a string into a WeaponReach.
func ParseWeaponReach(s string) WeaponReach {
	var wr WeaponReach
	s = strings.ReplaceAll(s, " ", "")
	if s != "" {
		s = strings.ToLower(s)
		if !strings.Contains(s, "spec") {
			s = strings.ReplaceAll(s, "-", ",")
			wr.CloseCombat = strings.Contains(s, "c")
			wr.ChangeRequiresReady = strings.Contains(s, "*")
			s = strings.ReplaceAll(s, "*", "")
			parts := strings.Split(s, ",")
			wr.Min, _ = fxp.Extract(parts[0])
			if len(parts) > 1 {
				for _, one := range parts[1:] {
					reach, _ := fxp.Extract(one)
					if reach > wr.Max {
						wr.Max = reach
					}
				}
			}
			wr.Validate()
		}
	}
	return wr
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wr WeaponReach) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wr.Min)
	_ = binary.Write(h, binary.LittleEndian, wr.Max)
	_ = binary.Write(h, binary.LittleEndian, wr.CloseCombat)
	_ = binary.Write(h, binary.LittleEndian, wr.ChangeRequiresReady)
}

// Resolve any bonuses that apply.
func (wr WeaponReach) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponReach {
	result := wr
	result.CloseCombat = w.ResolveBoolFlag(CloseCombatWeaponSwitchType, result.CloseCombat)
	result.ChangeRequiresReady = w.ResolveBoolFlag(ReachChangeRequiresReadyWeaponSwitchType, result.ChangeRequiresReady)
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponMinReachBonusFeatureType, WeaponMaxReachBonusFeatureType) {
		if bonus.Type == WeaponMinReachBonusFeatureType {
			result.Min += bonus.AdjustedAmount()
		} else if bonus.Type == WeaponMaxReachBonusFeatureType {
			result.Max += bonus.AdjustedAmount()
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponReach) String() string {
	var buffer strings.Builder
	if wr.CloseCombat {
		buffer.WriteByte('C')
	}
	if wr.Min != 0 || wr.Max != 0 {
		if buffer.Len() != 0 {
			buffer.WriteByte(',')
		}
		buffer.WriteString(wr.Min.String())
		if wr.Min != wr.Max {
			buffer.WriteByte('-')
			buffer.WriteString(wr.Max.String())
		}
	}
	if wr.ChangeRequiresReady {
		buffer.WriteByte('*')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wr WeaponReach) Tooltip() string {
	if wr.ChangeRequiresReady {
		return i18n.Text("Changing reach requires a Ready maneuver.")
	}
	return ""
}

// Validate ensures that the data is valid.
func (wr *WeaponReach) Validate() {
	wr.Min = wr.Min.Max(0)
	wr.Max = wr.Max.Max(0)
	if wr.Min == 0 && wr.Max != 0 {
		wr.Min = fxp.One
	} else if wr.Min != 0 && wr.Max == 0 {
		wr.Max = wr.Min
	}
	wr.Max = max(wr.Max, wr.Min)
}
