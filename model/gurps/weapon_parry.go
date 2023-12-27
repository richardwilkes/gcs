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
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ json.Omitter     = WeaponParry{}
	_ json.Marshaler   = WeaponParry{}
	_ json.Unmarshaler = &(WeaponParry{})
)

// WeaponParry holds the parry data for a weapon.
type WeaponParry struct {
	No         bool
	Fencing    bool
	Unbalanced bool
	Modifier   fxp.Int
}

// ParseWeaponParry parses a string into a WeaponParry.
func ParseWeaponParry(s string) WeaponParry {
	var wp WeaponParry
	s = strings.ToLower(s)
	wp.No = strings.Contains(s, "no")
	if !wp.No {
		wp.Fencing = strings.Contains(s, "f")
		wp.Unbalanced = strings.Contains(s, "u")
		wp.Modifier, _ = fxp.Extract(s)
	}
	wp.Validate()
	return wp
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wp WeaponParry) ShouldOmit() bool {
	return wp == WeaponParry{}
}

// MarshalJSON marshals the data to JSON.
func (wp WeaponParry) MarshalJSON() ([]byte, error) {
	return json.Marshal(wp.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wp *WeaponParry) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*wp = ParseWeaponParry(s)
	return nil
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wp WeaponParry) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wp.Modifier)
	_ = binary.Write(h, binary.LittleEndian, wp.No)
	_ = binary.Write(h, binary.LittleEndian, wp.Fencing)
	_ = binary.Write(h, binary.LittleEndian, wp.Unbalanced)
}

// Resolve any bonuses that apply.
func (wp WeaponParry) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponParry {
	result := wp
	result.No = !w.ResolveBoolFlag(CanParryWeaponSwitchType, !result.No)
	result.Fencing = w.ResolveBoolFlag(FencingWeaponSwitchType, result.Fencing)
	result.Unbalanced = w.ResolveBoolFlag(UnbalancedWeaponSwitchType, result.Unbalanced)
	if !result.No {
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
				if def.Type() != ParryID {
					level = level.Div(fxp.Two).Trunc()
				}
				level += postAdj
				if best < level {
					best = level
				}
			}
			if best != fxp.Min {
				AppendBufferOntoNewLine(modifiersTooltip, primaryTooltip)
				result.Modifier += fxp.Three + best + pc.ParryBonus
				for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponParryBonusFeatureType) {
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
func (wp WeaponParry) String() string {
	if wp.No {
		return "No" // Not localized, since it is part of the data
	}
	if wp.Modifier == 0 && !wp.Fencing && !wp.Unbalanced {
		return ""
	}
	var buffer strings.Builder
	buffer.WriteString(wp.Modifier.String())
	if wp.Fencing {
		buffer.WriteString("F")
	}
	if wp.Unbalanced {
		buffer.WriteString("U")
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (wp WeaponParry) Tooltip(w *Weapon) string {
	if wp.No || (!wp.Fencing && !wp.Unbalanced) {
		return ""
	}
	var buffer strings.Builder
	if wp.Fencing {
		fmt.Fprintf(&buffer, i18n.Text("Fencing weapon. When retreating, your Parry is %v (instead of %v). You suffer a -2 cumulative penalty for multiple parries after the first on the same turn instead of the usual -4. Flails cannot be parried by this weapon."), wp.Modifier+fxp.Three, wp.Modifier+fxp.One)
	}
	if wp.Unbalanced {
		if buffer.Len() != 0 {
			buffer.WriteString("\n\n")
		}
		fmt.Fprintf(&buffer, i18n.Text("Unbalanced weapon. You cannot use it to parry if you have already used it to attack this turn (or vice-versa) unless your current ST is %v or greater."), w.Strength.Resolve(w, nil).Minimum.Mul(fxp.OneAndAHalf).Ceil())
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wp *WeaponParry) Validate() {
	if wp.No {
		wp.Modifier = 0
		wp.Fencing = false
		wp.Unbalanced = false
	}
}
