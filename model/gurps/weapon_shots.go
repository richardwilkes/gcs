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
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
)

var (
	_ json.Omitter     = WeaponShots{}
	_ json.Marshaler   = WeaponShots{}
	_ json.Unmarshaler = &(WeaponShots{})
)

// WeaponShots holds the shots data for a weapon.
type WeaponShots struct {
	NonChamberShots     fxp.Int
	ChamberShots        fxp.Int
	ShotDuration        fxp.Int
	ReloadTime          fxp.Int
	ReloadTimeIsPerShot bool
	Thrown              bool
}

// ParseWeaponShots parses a string into a WeaponShots.
func ParseWeaponShots(s string) WeaponShots {
	var ws WeaponShots
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, ",", "")
	if !strings.Contains(s, "fp") &&
		!strings.Contains(s, "hrs") &&
		!strings.Contains(s, "day") {
		ws.Thrown = strings.Contains(s, "t")
		if !strings.Contains(s, "spec") {
			ws.NonChamberShots, s = fxp.Extract(s)
			if strings.HasPrefix(s, "+") {
				ws.ChamberShots, s = fxp.Extract(s)
			}
			if strings.HasPrefix(s, "x") {
				ws.ShotDuration, s = fxp.Extract(s[1:])
			}
			if strings.HasPrefix(s, "(") {
				ws.ReloadTime, _ = fxp.Extract(s[1:])
				ws.ReloadTimeIsPerShot = strings.Contains(s, "i")
			}
		}
	}
	ws.Validate()
	return ws
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (ws WeaponShots) ShouldOmit() bool {
	return ws == WeaponShots{}
}

// MarshalJSON marshals the data to JSON.
func (ws WeaponShots) MarshalJSON() ([]byte, error) {
	return json.Marshal(ws.String())
}

// UnmarshalJSON unmarshals the data from JSON.
func (ws *WeaponShots) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*ws = ParseWeaponShots(s)
	return nil
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (ws WeaponShots) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, ws.NonChamberShots)
	_ = binary.Write(h, binary.LittleEndian, ws.ChamberShots)
	_ = binary.Write(h, binary.LittleEndian, ws.ShotDuration)
	_ = binary.Write(h, binary.LittleEndian, ws.ReloadTime)
	_ = binary.Write(h, binary.LittleEndian, ws.ReloadTimeIsPerShot)
	_ = binary.Write(h, binary.LittleEndian, ws.Thrown)
}

// Resolve any bonuses that apply.
func (ws WeaponShots) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponShots {
	result := ws
	result.ReloadTimeIsPerShot = w.ResolveBoolFlag(wswitch.ReloadTimeIsPerShot, result.ReloadTimeIsPerShot)
	result.Thrown = w.ResolveBoolFlag(wswitch.Thrown, result.Thrown)
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponNonChamberShotsBonus, feature.WeaponChamberShotsBonus, feature.WeaponShotDurationBonus, feature.WeaponReloadTimeBonus) {
		switch bonus.Type {
		case feature.WeaponNonChamberShotsBonus:
			result.NonChamberShots += bonus.AdjustedAmount()
		case feature.WeaponChamberShotsBonus:
			result.ChamberShots += bonus.AdjustedAmount()
		case feature.WeaponShotDurationBonus:
			result.ShotDuration += bonus.AdjustedAmount()
		case feature.WeaponReloadTimeBonus:
			result.ReloadTime += bonus.AdjustedAmount()
		default:
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (ws WeaponShots) String() string {
	var buffer strings.Builder
	if ws.Thrown {
		buffer.WriteByte('T')
	} else {
		if ws.NonChamberShots <= 0 {
			return ""
		}
		buffer.WriteString(ws.NonChamberShots.String())
		if ws.ChamberShots > 0 {
			buffer.WriteByte('+')
			buffer.WriteString(ws.ChamberShots.String())
		}
		if ws.ShotDuration > 0 {
			buffer.WriteByte('x')
			buffer.WriteString(ws.ShotDuration.String())
			buffer.WriteByte('s')
		}
	}
	if ws.ReloadTime > 0 {
		buffer.WriteByte('(')
		buffer.WriteString(ws.ReloadTime.String())
		if ws.ReloadTimeIsPerShot {
			buffer.WriteByte('i')
		}
		buffer.WriteByte(')')
	}
	return buffer.String()
}

// Tooltip returns a tooltip for the data, if any. Call .Resolve() prior to calling this method if you want the tooltip
// to be based on the resolved values.
func (ws WeaponShots) Tooltip() string {
	if ws.ReloadTimeIsPerShot {
		return i18n.Text("Reload time is per shot")
	}
	return ""
}

// Validate ensures that the data is valid.
func (ws *WeaponShots) Validate() {
	ws.ReloadTime = ws.ReloadTime.Max(0)
	if ws.Thrown {
		ws.NonChamberShots = 0
		ws.ChamberShots = 0
		ws.ShotDuration = 0
		return
	}
	ws.NonChamberShots = ws.NonChamberShots.Max(0)
	if ws.NonChamberShots == 0 {
		ws.ChamberShots = 0
		ws.ShotDuration = 0
		ws.ReloadTime = 0
		return
	}
	ws.ChamberShots = ws.ChamberShots.Max(0)
	ws.ShotDuration = ws.ShotDuration.Max(0)
}
