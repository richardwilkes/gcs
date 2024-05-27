// Copyright (c) 1998-2024 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"encoding/binary"
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ json.Omitter     = WeaponReach{}
	_ json.Marshaler   = WeaponReach{}
	_ json.Unmarshaler = &(WeaponReach{})
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

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wr WeaponReach) ShouldOmit() bool {
	return wr == WeaponReach{}
}

// MarshalJSON marshals the data to JSON.
func (wr WeaponReach) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (wr *WeaponReach) UnmarshalJSON(data []byte) error {
	var s string
	err := json.Unmarshal(data, &s)
	if err != nil {
		return err
	}
	*wr = ParseWeaponReach(s)
	return nil
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
	result.CloseCombat = w.ResolveBoolFlag(wswitch.CloseCombat, result.CloseCombat)
	result.ChangeRequiresReady = w.ResolveBoolFlag(wswitch.ReachChangeRequiresReady, result.ChangeRequiresReady)
	var percentMin, percentMax fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponMinReachBonus, feature.WeaponMaxReachBonus) {
		amt := bonus.AdjustedAmountForWeapon(w)
		if bonus.Type == feature.WeaponMinReachBonus {
			if bonus.Percent {
				percentMin += amt
			} else {
				result.Min += amt
			}
		} else if bonus.Type == feature.WeaponMaxReachBonus {
			if bonus.Percent {
				percentMax += amt
			} else {
				result.Max += amt
			}
		}
	}
	if percentMin != 0 {
		result.Min += result.Min.Mul(percentMin).Div(fxp.Hundred).Trunc()
	}
	if percentMax != 0 {
		result.Max += result.Max.Mul(percentMax).Div(fxp.Hundred).Trunc()
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
		return i18n.Text("*: Changing reach requires a Ready maneuver.")
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
