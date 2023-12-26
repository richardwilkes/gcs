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
	"github.com/richardwilkes/toolbox/xio"
)

// WeaponRange holds the range data for a weapon.
type WeaponRange struct {
	HalfDamageRange fxp.Int
	MinRange        fxp.Int
	MaxRange        fxp.Int
	MusclePowered   bool
	RangeInMiles    bool
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
		wr.RangeInMiles = strings.Contains(s, "mi")
		s = strings.ReplaceAll(s, "mi.", "")
		s = strings.ReplaceAll(s, "mi", "")
		s = strings.ReplaceAll(s, ",", "")
		parts := strings.Split(s, "/")
		if len(parts) > 1 {
			wr.HalfDamageRange, _ = fxp.Extract(parts[0])
			parts[0] = parts[1]
		}
		parts = strings.Split(parts[0], "-")
		if len(parts) > 1 {
			wr.MinRange, _ = fxp.Extract(parts[0])
			wr.MaxRange, _ = fxp.Extract(parts[1])
		} else {
			wr.MaxRange, _ = fxp.Extract(parts[0])
		}
		wr.Validate()
	}
	return wr
}

// nolint:errcheck // Not checking errors on writes to a bytes.Buffer
func (wr WeaponRange) hash(h hash.Hash32) {
	_ = binary.Write(h, binary.LittleEndian, wr.HalfDamageRange)
	_ = binary.Write(h, binary.LittleEndian, wr.MinRange)
	_ = binary.Write(h, binary.LittleEndian, wr.MaxRange)
	_ = binary.Write(h, binary.LittleEndian, wr.MusclePowered)
	_ = binary.Write(h, binary.LittleEndian, wr.RangeInMiles)
}

// Resolve any bonuses that apply.
func (wr WeaponRange) Resolve(w *Weapon, modifiersTooltip *xio.ByteBuffer) WeaponRange {
	result := wr
	result.MusclePowered = w.ResolveBoolFlag(MusclePoweredWeaponSwitchType, result.MusclePowered)
	result.RangeInMiles = w.ResolveBoolFlag(RangeInMilesWeaponSwitchType, result.RangeInMiles)
	if result.MusclePowered {
		var st fxp.Int
		if w.Owner != nil {
			st = w.Owner.RatedStrength()
		}
		if st == 0 {
			if pc := w.PC(); pc != nil {
				st = pc.ThrowingStrength()
			}
		}
		if st > 0 {
			result.HalfDamageRange = result.HalfDamageRange.Mul(st).Trunc().Max(0)
			result.MinRange = result.MinRange.Mul(st).Trunc().Max(0)
			result.MaxRange = result.MaxRange.Mul(st).Trunc().Max(0)
		}
	}
	for _, bonus := range w.collectWeaponBonuses(1, modifiersTooltip, WeaponHalfDamageRangeBonusFeatureType, WeaponMinRangeBonusFeatureType, WeaponMaxRangeBonusFeatureType) {
		switch bonus.Type {
		case WeaponHalfDamageRangeBonusFeatureType:
			result.HalfDamageRange += bonus.AdjustedAmount()
		case WeaponMinRangeBonusFeatureType:
			result.MinRange += bonus.AdjustedAmount()
		case WeaponMaxRangeBonusFeatureType:
			result.MaxRange += bonus.AdjustedAmount()
		default:
		}
	}
	result.Validate()
	return result
}

// String returns a string suitable for presentation, matching the standard GURPS weapon table entry format for this
// data. Call .Resolve() prior to calling this method if you want the resolved values.
func (wr WeaponRange) String(musclePowerIsResolved bool) string {
	var buffer strings.Builder
	if wr.HalfDamageRange != 0 {
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.HalfDamageRange.String())
		buffer.WriteByte('/')
	}
	if wr.MinRange != 0 || wr.MaxRange != 0 {
		if wr.MinRange != 0 && wr.MinRange != wr.MaxRange {
			if wr.MusclePowered && !musclePowerIsResolved {
				buffer.WriteByte('x')
			}
			buffer.WriteString(wr.MinRange.String())
			buffer.WriteByte('-')
		}
		if wr.MusclePowered && !musclePowerIsResolved {
			buffer.WriteByte('x')
		}
		buffer.WriteString(wr.MaxRange.String())
	}
	if wr.RangeInMiles && buffer.Len() != 0 {
		buffer.WriteByte(' ')
		buffer.WriteString(fxp.Mile.String())
	}
	return buffer.String()
}

// Validate ensures that the data is valid.
func (wr *WeaponRange) Validate() {
	wr.HalfDamageRange = wr.HalfDamageRange.Max(0)
	wr.MinRange = wr.MinRange.Max(0)
	wr.MaxRange = wr.MaxRange.Max(0)
	if wr.MinRange > wr.MaxRange {
		wr.MinRange, wr.MaxRange = wr.MaxRange, wr.MinRange
	}
	if wr.HalfDamageRange < wr.MinRange || wr.HalfDamageRange >= wr.MaxRange {
		wr.HalfDamageRange = 0
	}
}
