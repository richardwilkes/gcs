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
	"github.com/richardwilkes/toolbox/xio"
	"github.com/richardwilkes/toolbox/xmath/hashhelper"
)

var (
	_ json.Omitter     = WeaponRange{}
	_ json.Marshaler   = WeaponRange{}
	_ json.Unmarshaler = &(WeaponRange{})
)

// WeaponRange holds the range data for a weapon.
type WeaponRange struct {
	HalfDamage    fxp.Int
	Min           fxp.Int
	Max           fxp.Int
	MusclePowered bool
	InMiles       bool
}

// ParseWeaponRange parses a string into a WeaponRange.
func ParseWeaponRange(s string) WeaponRange {
	var wr WeaponRange
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "×", "x")
	if !strings.Contains(s, "sight") &&
		!strings.Contains(s, "spec") &&
		!strings.Contains(s, "skill") &&
		!strings.Contains(s, "point") &&
		!strings.Contains(s, "pbaoe") &&
		!strings.HasPrefix(s, "b") {
		s = strings.ReplaceAll(s, ",max", "/")
		s = strings.ReplaceAll(s, "max", "")
		s = strings.ReplaceAll(s, "1/2d", "")
		wr.MusclePowered = strings.Contains(s, "x")
		s = strings.ReplaceAll(s, "x", "")
		s = strings.ReplaceAll(s, "st", "")
		s = strings.ReplaceAll(s, "c/", "")
		wr.InMiles = strings.Contains(s, "mi")
		s = strings.ReplaceAll(s, "mi.", "")
		s = strings.ReplaceAll(s, "mi", "")
		s = strings.ReplaceAll(s, ",", "")
		parts := strings.Split(s, "/")
		if len(parts) > 1 {
			wr.HalfDamage, _ = fxp.Extract(parts[0])
			parts[0] = parts[1]
		}
		parts = strings.Split(parts[0], "-")
		if len(parts) > 1 {
			wr.Min, _ = fxp.Extract(parts[0])
			wr.Max, _ = fxp.Extract(parts[1])
		} else {
			wr.Max, _ = fxp.Extract(parts[0])
		}
		wr.Validate()
	}
	return wr
}

// ShouldOmit returns true if the data should be omitted from JSON output.
func (wr WeaponRange) ShouldOmit() bool {
	return wr == WeaponRange{}
}

// MarshalJSON marshals the data to JSON.
func (wr WeaponRange) MarshalJSON() ([]byte, error) {
	return json.Marshal(wr.String(false))
}

// UnmarshalJSON unmarshals the data from JSON.
func (wr *WeaponRange) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*wr = ParseWeaponRange(s)
	return nil
}

// Hash writes this object's contents into the hasher.
func (wr WeaponRange) Hash(h hash.Hash) {
	if wr.ShouldOmit() {
		hashhelper.Num8(h, uint8(255))
		return
	}
	hashhelper.Num64(h, wr.HalfDamage)
	hashhelper.Num64(h, wr.Min)
	hashhelper.Num64(h, wr.Max)
	hashhelper.Bool(h, wr.MusclePowered)
	hashhelper.Bool(h, wr.InMiles)
}

// Resolve any bonuses that apply.
func (wr WeaponRange) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRange {
	result := wr
	result.MusclePowered = w.ResolveBoolFlag(wswitch.MusclePowered, result.MusclePowered)
	result.InMiles = w.ResolveBoolFlag(wswitch.RangeInMiles, result.InMiles)
	if result.MusclePowered {
		var st fxp.Int
		maxST := w.Strength.Resolve(w, nil).Min.Mul(fxp.Three)
		if w.Owner != nil {
			st = w.Owner.RatedStrength()
		}
		if st == 0 {
			if entity := w.Entity(); entity != nil {
				st = entity.ThrowingStrength()
			}
		}
		var percentMin fxp.Int
		for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponEffectiveSTBonus) {
			amt := bonus.AdjustedAmountForWeapon(w)
			if bonus.Percent {
				percentMin += amt
			} else {
				st += amt
			}
		}
		if percentMin != 0 {
			st += st.Mul(percentMin).Div(fxp.Hundred).Trunc()
		}
		if st < 0 {
			st = 0
		}
		if maxST > 0 && maxST < st {
			st = maxST
		}
		if st > 0 {
			result.HalfDamage = result.HalfDamage.Mul(st).Trunc().Max(0)
			result.Min = result.Min.Mul(st).Trunc().Max(0)
			result.Max = result.Max.Mul(st).Trunc().Max(0)
		}
	}
	var percentHalfDamage, percentMin, percentMax fxp.Int
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, feature.WeaponHalfDamageRangeBonus, feature.WeaponMinRangeBonus, feature.WeaponMaxRangeBonus) {
		amt := bonus.AdjustedAmountForWeapon(w)
		switch bonus.Type {
		case feature.WeaponHalfDamageRangeBonus:
			if bonus.Percent {
				percentHalfDamage += amt
			} else {
				result.HalfDamage += amt
			}
		case feature.WeaponMinRangeBonus:
			if bonus.Percent {
				percentMin += amt
			} else {
				result.Min += amt
			}
		case feature.WeaponMaxRangeBonus:
			if bonus.Percent {
				percentMax += amt
			} else {
				result.Max += amt
			}
		default:
		}
	}
	if percentHalfDamage != 0 {
		result.HalfDamage += result.HalfDamage.Mul(percentHalfDamage).Div(fxp.Hundred).Trunc()
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
func (wr WeaponRange) String(musclePowerIsResolved bool) string {
	var buffer strings.Builder
	if wr.HalfDamage != 0 {
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.HalfDamage.Comma())
		buffer.WriteByte('/')
	}
	if wr.Min != 0 || wr.Max != 0 {
		if wr.Min != 0 && wr.Min != wr.Max {
			if wr.MusclePowered && !musclePowerIsResolved {
				buffer.WriteByte('x')
			}
			buffer.WriteString(wr.Min.Comma())
			buffer.WriteByte('-')
		}
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.Max.Comma())
	}
	if wr.InMiles && buffer.Len() != 0 {
		buffer.WriteByte(' ')
		buffer.WriteString(fxp.Mile.String())
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRange) Validate() {
	wr.HalfDamage = wr.HalfDamage.Max(0)
	wr.Min = wr.Min.Max(0)
	wr.Max = wr.Max.Max(0)
	if wr.Min > wr.Max {
		wr.Min, wr.Max = wr.Max, wr.Min
	}
	if wr.HalfDamage < wr.Min || wr.HalfDamage >= wr.Max {
		wr.HalfDamage = 0
	}
}
