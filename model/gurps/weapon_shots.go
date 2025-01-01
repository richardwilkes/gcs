// Copyright (c) 1998-2025 by Richard A. Wilkes. All rights reserved.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, version 2.0. If a copy of the MPL was not distributed with
// this file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// This Source Code Form is "Incompatible With Secondary Licenses", as
// defined by the Mozilla Public License, version 2.0.

package gurps

import (
	"hash"
	"strings"

	"github.com/richardwilkes/gcs/v5/model/fxp"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/feature"
	"github.com/richardwilkes/gcs/v5/model/gurps/enums/wswitch"
	"github.com/richardwilkes/json"
	"github.com/richardwilkes/toolbox/i18n"
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var (
	_ json.Omitter     = WeaponShots{}
	_ json.Marshaler   = WeaponShots{}
	_ json.Unmarshaler = &(WeaponShots{})
)

// WeaponShots holds the shots data for a weapon.
type WeaponShots struct {
	Count               fxp.Int
	InChamber           fxp.Int
	Duration            fxp.Int
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
			ws.Count, s = fxp.Extract(s)
			if strings.HasPrefix(s, "+") {
				ws.InChamber, s = fxp.Extract(s)
			}
			if strings.HasPrefix(s, "x") {
				ws.Duration, s = fxp.Extract(s[1:])
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

// Hash writes this object's contents into the hasher.
func (ws WeaponShots) Hash(h hash.Hash) {
	if ws.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, ws.Count)
	hashhelper.Num64(h, ws.InChamber)
	hashhelper.Num64(h, ws.Duration)
	hashhelper.Num64(h, ws.ReloadTime)
	hashhelper.Bool(h, ws.ReloadTimeIsPerShot)
	hashhelper.Bool(h, ws.Thrown)
}

// Resolve any bonuses that apply.
func (ws WeaponShots) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponShots {
	result := ws
	result.ReloadTimeIsPerShot = w.ResolveBoolFlag(wswitch.ReloadTimeIsPerShot, result.ReloadTimeIsPerShot)
	result.Thrown = w.ResolveBoolFlag(wswitch.Thrown, result.Thrown)
	var percentCount, percentInChamber, percentDuration, percentReloadTime fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponNonChamberShotsBonus, feature.WeaponChamberShotsBonus, feature.WeaponShotDurationBonus, feature.WeaponReloadTimeBonus) {
		amt := bonus.AdjustedAmountForWeapon(w)
		switch bonus.Type {
		case feature.WeaponNonChamberShotsBonus:
			if bonus.Percent {
				percentCount += amt
			} else {
				result.Count += amt
			}
		case feature.WeaponChamberShotsBonus:
			if bonus.Percent {
				percentInChamber += amt
			} else {
				result.InChamber += amt
			}
		case feature.WeaponShotDurationBonus:
			if bonus.Percent {
				percentDuration += amt
			} else {
				result.Duration += amt
			}
		case feature.WeaponReloadTimeBonus:
			if bonus.Percent {
				percentReloadTime += amt
			} else {
				result.ReloadTime += amt
			}
		default:
		}
	}
	if percentCount != 0 {
		result.Count += result.Count.Mul(percentCount).Div(fxp.Hundred).Trunc()
	}
	if percentInChamber != 0 {
		result.InChamber += result.InChamber.Mul(percentInChamber).Div(fxp.Hundred).Trunc()
	}
	if percentDuration != 0 {
		result.Duration += result.Duration.Mul(percentDuration).Div(fxp.Hundred).Trunc()
	}
	if percentReloadTime != 0 {
		result.ReloadTime += result.ReloadTime.Mul(percentReloadTime).Div(fxp.Hundred).Trunc()
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
		if ws.Count <= 0 && ws.InChamber <= 0 {
			return ""
		}
		buffer.WriteString(ws.Count.Max(0).String())
		if ws.InChamber > 0 {
			buffer.WriteByte('+')
			buffer.WriteString(ws.InChamber.String())
		}
		if ws.Duration > 0 {
			buffer.WriteByte('x')
			buffer.WriteString(ws.Duration.String())
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
		return i18n.Text("i: Reload time is per shot")
	}
	return ""
}

// Validate ensures that the data is valid.
func (ws *WeaponShots) Validate() {
	ws.ReloadTime = ws.ReloadTime.Max(0)
	if ws.Thrown {
		ws.Count = 0
		ws.InChamber = 0
		ws.Duration = 0
		return
	}
	ws.Count = ws.Count.Max(0)
	ws.InChamber = ws.InChamber.Max(0)
	if ws.Count == 0 && ws.InChamber == 0 {
		ws.Duration = 0
		ws.ReloadTime = 0
		return
	}
	ws.Duration = ws.Duration.Max(0)
}
